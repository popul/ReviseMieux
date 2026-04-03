// Package config gère la configuration de l'application
package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigYAML représente la structure du fichier config.yaml
type ConfigYAML struct {
	NombreMaxPages int `yaml:"nombre_max_pages"`
}

// Config contient toute la configuration de l'application
type Config struct {
	// Serveur
	Port string

	// Base de données
	DatabaseURL string

	// LLM
	OpenAIAPIKey  string
	MistralAPIKey string
	GeminiAPIKey  string

	// Quotas journaliers
	QuotaOCRJour        int
	QuotaGenerationJour int

	// Stockage local (filesystem)
	StoragePath string

	// Stockage S3/MinIO
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3UseSSL    bool

	// Tesseract OCR
	TesseractEnabled bool

	// Limite de pages
	NombreMaxPages int
}

// Charger charge la configuration depuis le fichier YAML et les variables d'environnement
func Charger() *Config {
	return ChargerAvecFichier("config.yaml")
}

// ChargerAvecFichier charge la configuration depuis un fichier YAML puis applique les env vars
func ChargerAvecFichier(chemin string) *Config {
	// 1. Lire le fichier YAML (ignoré si absent ou invalide)
	var configYAML ConfigYAML
	if data, err := os.ReadFile(chemin); err == nil {
		_ = yaml.Unmarshal(data, &configYAML)
	}

	// 2. Construire la config avec les env vars en priorité
	nombreMaxPages := configYAML.NombreMaxPages
	if v := getEnvInt("NOMBRE_MAX_PAGES", 0); v > 0 {
		nombreMaxPages = v
	}

	// 3. Appliquer le défaut si valeur <= 0
	if nombreMaxPages <= 0 {
		nombreMaxPages = 30
	}

	return &Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://revisemieux:revisemieux@localhost:5432/revisemieux?sslmode=disable"),
		OpenAIAPIKey:        getEnv("OPENAI_API_KEY", ""),
		MistralAPIKey:       getEnv("MISTRAL_API_KEY", ""),
		GeminiAPIKey:        getEnv("GEMINI_API_KEY", ""),
		QuotaOCRJour:        getEnvInt("QUOTA_OCR_JOUR", 50),
		QuotaGenerationJour: getEnvInt("QUOTA_GENERATION_JOUR", 100),
		StoragePath:         getEnv("STORAGE_PATH", "./storage"),
		S3Endpoint:          getEnv("S3_ENDPOINT", "localhost:9000"),
		S3AccessKey:         getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:         getEnv("S3_SECRET_KEY", ""),
		S3Bucket:            getEnv("S3_BUCKET", "revisemieux"),
		S3UseSSL:            getEnvBool("S3_USE_SSL", false),
		TesseractEnabled:    getEnvBool("TESSERACT_ENABLED", true),
		NombreMaxPages:      nombreMaxPages,
	}
}

// getEnv retourne la valeur de la variable d'environnement ou la valeur par défaut
func getEnv(cle, defaut string) string {
	if valeur := os.Getenv(cle); valeur != "" {
		return valeur
	}
	return defaut
}

// getEnvBool retourne la valeur booléenne de la variable d'environnement ou la valeur par défaut
func getEnvBool(cle string, defaut bool) bool {
	valeur := os.Getenv(cle)
	if valeur == "" {
		return defaut
	}
	switch strings.ToLower(valeur) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return defaut
	}
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
