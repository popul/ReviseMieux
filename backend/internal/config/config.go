// Package config gère la configuration de l'application
package config

import (
	"os"
)

// Config contient toute la configuration de l'application
type Config struct {
	// Serveur
	Port string

	// Base de données
	DatabaseURL string

	// LLM
	OpenAIAPIKey  string
	MistralAPIKey string

	// Quotas journaliers
	QuotaOCRJour        int
	QuotaGenerationJour int

	// Stockage
	StoragePath string
}

// Charger charge la configuration depuis les variables d'environnement
func Charger() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://revisemieux:revisemieux@localhost:5432/revisemieux?sslmode=disable"),
		OpenAIAPIKey:        getEnv("OPENAI_API_KEY", ""),
		MistralAPIKey:       getEnv("MISTRAL_API_KEY", ""),
		QuotaOCRJour:        getEnvInt("QUOTA_OCR_JOUR", 50),
		QuotaGenerationJour: getEnvInt("QUOTA_GENERATION_JOUR", 100),
		StoragePath:         getEnv("STORAGE_PATH", "./storage"),
	}
}

// getEnv retourne la valeur de la variable d'environnement ou la valeur par défaut
func getEnv(cle, defaut string) string {
	if valeur := os.Getenv(cle); valeur != "" {
		return valeur
	}
	return defaut
}

// getEnvInt retourne la valeur entière de la variable d'environnement ou la valeur par défaut
func getEnvInt(cle string, defaut int) int {
	valeur := os.Getenv(cle)
	if valeur == "" {
		return defaut
	}
	var resultat int
	for _, c := range valeur {
		if c >= '0' && c <= '9' {
			resultat = resultat*10 + int(c-'0')
		} else {
			return defaut
		}
	}
	return resultat
}
