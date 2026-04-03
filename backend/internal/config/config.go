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

	// LLM provider selection: "gemini" (default), "anthropic", "mistral"
	LLMProvider string

	// Anthropic LLM
	AnthropicAPIKey        string
	AnthropicStructModel   string // Model for structuration (default: claude-sonnet-4-6)
	AnthropicFidelityModel string // Model for fidelity check (default: claude-haiku-4-5)

	// Google Gemini LLM
	GoogleAIAPIKey   string
	GeminiStructModel string // Model for structuration (default: gemini-2.5-flash)

	// Mistral LLM
	MistralAPIKey     string
	MistralStructModel string // Model for structuration (default: mistral-small-latest)
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

		LLMProvider: envOrDefault("LLM_PROVIDER", "gemini"),

		AnthropicAPIKey:        envOrDefault("ANTHROPIC_API_KEY", ""),
		AnthropicStructModel:   envOrDefault("ANTHROPIC_STRUCT_MODEL", "claude-sonnet-4-6"),
		AnthropicFidelityModel: envOrDefault("ANTHROPIC_FIDELITY_MODEL", "claude-haiku-4-5"),

		GoogleAIAPIKey:    envOrDefault("GOOGLE_AI_API_KEY", ""),
		GeminiStructModel: envOrDefault("GEMINI_STRUCT_MODEL", "gemini-2.5-flash"),

		MistralAPIKey:      envOrDefault("MISTRAL_API_KEY", ""),
		MistralStructModel: envOrDefault("MISTRAL_STRUCT_MODEL", "mistral-small-latest"),
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
