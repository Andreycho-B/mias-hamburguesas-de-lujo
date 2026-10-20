package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"mias-backend/internal/config"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Ciclo de reintentos para permitir que la base de datos en la nube termine de iniciar
	var pool *pgxpool.Pool
	var err error
	maxRetries := 6
	for i := 1; i <= maxRetries; i++ {
		log.Printf("[MIGRATE] Conectando a PostgreSQL (intento %d/%d)...", i, maxRetries)
		connectCtx, cancelConnect := context.WithTimeout(context.Background(), 8*time.Second)
		pool, err = pgxpool.New(connectCtx, cfg.DatabaseURL)
		if err == nil {
			if pingErr := pool.Ping(connectCtx); pingErr == nil {
				cancelConnect()
				log.Printf("[MIGRATE] Conexión establecida exitosamente.")
				break
			} else {
				pool.Close()
				err = pingErr
			}
		}
		cancelConnect()
		if i < maxRetries {
			log.Printf("[MIGRATE] Base de datos no lista aún (%v). Reintentando en 3s...", err)
			time.Sleep(3 * time.Second)
		}
	}
	if err != nil {
		log.Fatalf("[MIGRATE] Error crítico al conectar a PostgreSQL tras %d intentos: %v", maxRetries, err)
	}
	defer pool.Close()

	// Control de versiones con tabla schema_migrations
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Fatalf("[MIGRATE] Error asegurando tabla schema_migrations: %v", err)
	}

	applied := make(map[string]bool)
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err == nil {
		for rows.Next() {
			var v string
			if scanErr := rows.Scan(&v); scanErr == nil {
				applied[v] = true
			}
		}
		rows.Close()
	}

	// Compatibilidad retroactiva: si 'users' ya existe pero schema_migrations estaba vacía
	if len(applied) == 0 {
		var usersExists bool
		_ = pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users')").Scan(&usersExists)
		if usersExists {
			log.Printf("[MIGRATE] Esquema inicial existente detectado. Marcando migraciones base previas...")
			_, _ = pool.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ('00001_init.sql'), ('00002_seed.sql') ON CONFLICT DO NOTHING")
			applied["00001_init.sql"] = true
			applied["00002_seed.sql"] = true
		}
	}

	// Determine migrations folder path
	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = filepath.Join("backend", "migrations")
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("[MIGRATE] Cannot read migrations dir: %v", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, filepath.Join(migrationsDir, f.Name()))
		}
	}
	sort.Strings(sqlFiles)

	for _, file := range sqlFiles {
		base := filepath.Base(file)
		if applied[base] {
			log.Printf("[MIGRATE] %s ya aplicada previamente, omitiendo.", base)
			continue
		}

		log.Printf("[MIGRATE] Applying %s...", base)
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("[MIGRATE] Failed to read %s: %v", file, err)
		}

		rawSQL := string(content)
		upSQL := rawSQL
		if idx := strings.Index(rawSQL, "-- +goose Down"); idx != -1 {
			upSQL = rawSQL[:idx]
		}
		upSQL = strings.ReplaceAll(upSQL, "-- +goose Up", "")

		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Fatalf("[MIGRATE] Error iniciando transacción para %s: %v", base, err)
		}

		if _, err := tx.Exec(ctx, upSQL); err != nil {
			tx.Rollback(ctx)
			log.Fatalf("[MIGRATE] Error aplicando %s: %v", base, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", base); err != nil {
			tx.Rollback(ctx)
			log.Fatalf("[MIGRATE] Error registrando migración %s: %v", base, err)
		}

		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("[MIGRATE] Error confirmando transacción para %s: %v", base, err)
		}

		applied[base] = true
		log.Printf("[MIGRATE] Successfully applied %s", base)
	}

	// Verify tables
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&count)
	if err != nil {
		log.Fatalf("[MIGRATE] Failed to count tables: %v", err)
	}
	fmt.Printf("\n[MIGRATE SUCCESS] Database initialized with %d public tables.\n", count)
}
