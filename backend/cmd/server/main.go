package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/revisemieux/backend/internal/api"
	"github.com/revisemieux/backend/internal/config"
	"github.com/revisemieux/backend/internal/store"
)

func main() {
	// Charger le fichier .env si présent
	godotenv.Load()

	// Charger la configuration
	cfg := config.Charger()

	// Connexion à la base de données
	db, err := store.NouveauStore(cfg.DatabaseURL)
	if err != nil {
		log.Printf("⚠️  Avertissement: connexion DB échouée: %v", err)
		log.Printf("   L'API démarrera sans connexion à la base de données")
	} else {
		defer db.Fermer()
		log.Println("✓ Connexion à la base de données établie")
	}

	// Créer le routeur Gin
	r := gin.Default()

	// Configurer les middleware
	api.ConfigurerMiddleware(r)

	// Créer les handlers avec le store
	handlers := api.NouveauHandlers(db)

	// Configurer les routes
	api.ConfigurerRoutes(r, handlers)

	// Démarrer le serveur
	log.Printf("🚀 Serveur démarré sur le port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Erreur démarrage serveur: %v", err)
	}
}
