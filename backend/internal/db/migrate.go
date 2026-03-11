package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

	for _, f := range files {
		version := filepath.Base(f)

		var exists bool
		err := pool.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)",
			version,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("db.Migrate: check %s: %w", version, err)
		}
		if exists {
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
			// Extract a short error context
			errMsg := err.Error()
			if idx := strings.Index(errMsg, "\n"); idx > 0 {
				errMsg = errMsg[:idx]
			}
			return fmt.Errorf("db.Migrate: apply %s: %s", version, errMsg)
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
