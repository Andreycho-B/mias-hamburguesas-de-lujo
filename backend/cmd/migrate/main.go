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

	log.Printf("[MIGRATE] Connecting to PostgreSQL at %s...", cfg.DatabaseURL)
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[MIGRATE] Failed to connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("[MIGRATE] Ping failed: %v", err)
	}
	log.Printf("[MIGRATE] Connected successfully.")

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
