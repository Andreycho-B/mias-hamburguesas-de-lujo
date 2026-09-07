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

	// Idempotencia: Verificar si la tabla 'users' ya existe para evitar errores en reinicios
	var usersTableExists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users')").Scan(&usersTableExists)
	if err == nil && usersTableExists {
		var count int
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&count)
		log.Printf("[MIGRATE SUCCESS] Base de datos ya inicializada previamente (%d tablas públicas). Omitiendo migración.", count)
		return
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
		log.Printf("[MIGRATE] Applying %s...", filepath.Base(file))
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("[MIGRATE] Failed to read %s: %v", file, err)
		}

		rawSQL := string(content)
		// Extract Up section
		upSQL := rawSQL
		if idx := strings.Index(rawSQL, "-- +goose Down"); idx != -1 {
			upSQL = rawSQL[:idx]
		}
		// Remove +goose Up annotations
		upSQL = strings.ReplaceAll(upSQL, "-- +goose Up", "")

		_, err = pool.Exec(ctx, upSQL)
		if err != nil {
			log.Fatalf("[MIGRATE] Error applying %s: %v", filepath.Base(file), err)
		}
		log.Printf("[MIGRATE] Successfully applied %s", filepath.Base(file))
	}

	// Verify tables
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&count)
	if err != nil {
		log.Fatalf("[MIGRATE] Failed to count tables: %v", err)
	}
	fmt.Printf("\n[MIGRATE SUCCESS] Database initialized with %d public tables.\n", count)
}
