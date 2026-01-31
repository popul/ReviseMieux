package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/revisemieux/backend/internal/api"
	"github.com/revisemieux/backend/internal/config"
)

func main() {
	// Charger le fichier .env si présent
	godotenv.Load()

	// Charger la configuration
	cfg := config.Charger()

	// Créer le routeur Gin
	r := gin.Default()

	// Configurer les middleware
	api.ConfigurerMiddleware(r)

	// Créer les handlers
	handlers := api.NouveauHandlers()

	// Configurer les routes
	api.ConfigurerRoutes(r, handlers)

	// Démarrer le serveur
	log.Printf("🚀 Serveur démarré sur le port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Erreur démarrage serveur: %v", err)
	}
}
