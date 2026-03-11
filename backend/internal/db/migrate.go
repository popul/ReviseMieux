package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`

// Migrate applies all pending SQL migrations from the given directory.
// Migrations are sorted by filename and tracked in the schema_migrations table.
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	if _, err := pool.Exec(ctx, migrationTable); err != nil {
		return fmt.Errorf("db.Migrate: create tracking table: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("db.Migrate: glob: %w", err)
	}
	sort.Strings(files)

	// Load all applied versions in one query to avoid N+1.
	applied := make(map[string]bool)
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("db.Migrate: load applied versions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("db.Migrate: scan version: %w", err)
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("db.Migrate: iterate versions: %w", err)
	}

	for _, f := range files {
		version := filepath.Base(f)
		if applied[version] {
			continue
		}

		sql, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("db.Migrate: read %s: %w", version, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("db.Migrate: begin tx for %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("db.Migrate: apply %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("db.Migrate: record %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("db.Migrate: commit %s: %w", version, err)
		}

		log.Printf("Applied migration: %s", version)
	}

	return nil
}
