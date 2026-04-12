package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/popul/revisemieux/internal/db"
)

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://revisemieux:revisemieux@localhost:5432/revisemieux?sslmode=disable"
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool, migrationsDir); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations OK")
}
