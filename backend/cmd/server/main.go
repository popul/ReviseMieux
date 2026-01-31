package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/revisemieux/backend/internal/api"
	"github.com/revisemieux/backend/internal/config"
	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/services"
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

	// Créer le gestionnaire LLM avec fallback automatique
	var gestionnaireLLM *llm.GestionnaireLLM
	if cfg.OpenAIAPIKey != "" || cfg.MistralAPIKey != "" {
		gestionnaireLLM = llm.NouveauGestionnaireAvecCles(cfg.OpenAIAPIKey, cfg.MistralAPIKey)
		if gestionnaireLLM != nil {
			log.Printf("✓ Service LLM configuré (fournisseur: %s)", gestionnaireLLM.Nom())
		}
	}
	if gestionnaireLLM == nil {
		log.Printf("⚠️  Avertissement: aucune clé API LLM configurée (OPENAI_API_KEY ou MISTRAL_API_KEY)")
		log.Printf("   L'OCR et la génération de contenu ne seront pas disponibles")
	}

	// Créer le service OCR
	var serviceOCR *services.ServiceOCR
	if gestionnaireLLM != nil {
		serviceOCR = services.NouveauServiceOCR(gestionnaireLLM)
		log.Println("✓ Service OCR initialisé")
	}

	// Créer le repository des cours
	var coursRepo store.CoursRepository
	if db != nil {
		coursRepo = store.NouveauCoursRepo(db)
		log.Println("✓ Repository cours initialisé")
	}

	// Créer le routeur Gin
	r := gin.Default()

	// Configurer les middleware
	api.ConfigurerMiddleware(r)

	// Créer les handlers avec toutes les dépendances
	handlers := api.NouveauHandlers(db, serviceOCR, coursRepo)

	// Configurer les routes
	api.ConfigurerRoutes(r, handlers)

	// Démarrer le serveur
	log.Printf("🚀 Serveur démarré sur le port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Erreur démarrage serveur: %v", err)
	}
}
