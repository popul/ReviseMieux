package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Port        string
	GinMode     string
	DatabaseURL string
	RedisURL    string
	JWTSecret     string
	JWTExpiry     time.Duration
	MigrationsDir string
	S3Endpoint    string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
	S3Region    string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Port:        envOrDefault("PORT", "8080"),
		GinMode:     envOrDefault("GIN_MODE", "debug"),
		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://revisemieux:revisemieux@localhost:5432/revisemieux?sslmode=disable"),
		RedisURL:    envOrDefault("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:     envOrDefault("JWT_SECRET", ""),
		MigrationsDir: envOrDefault("MIGRATIONS_DIR", "migrations"),
		S3Endpoint:    envOrDefault("S3_ENDPOINT", "http://localhost:9000"),
		S3Bucket:    envOrDefault("S3_BUCKET", "revisemieux"),
		S3AccessKey: envOrDefault("S3_ACCESS_KEY", ""),
		S3SecretKey: envOrDefault("S3_SECRET_KEY", ""),
		S3Region:    envOrDefault("S3_REGION", "us-east-1"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required")
	}

	expiryStr := envOrDefault("JWT_EXPIRY", "24h")
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		return nil, fmt.Errorf("config: invalid JWT_EXPIRY %q: %w", expiryStr, err)
	}
	cfg.JWTExpiry = expiry

	return cfg, nil
}

// MustLoad calls Load and panics on error.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
