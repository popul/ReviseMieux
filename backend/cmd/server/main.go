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

		// Exécuter les migrations automatiquement au démarrage
		if err := db.Migrer("migrations"); err != nil {
			log.Printf("⚠️  Avertissement: erreur lors des migrations: %v", err)
		} else {
			log.Println("✓ Migrations appliquées")
		}
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

	// Créer les repositories
	var coursRepo store.CoursRepository
	var fichesRepo store.FichesRepository
	var quizRepo store.QuizRepository
	var ressourcesRepo store.RessourcesRepository
	var mindmapRepo store.MindmapRepository
	var quotasRepo store.QuotasRepository
	var copieRepo store.CopieExamenRepository
	var erreurRepo store.ErreurAnalyseRepository
	if db != nil {
		coursRepo = store.NouveauCoursRepo(db)
		fichesRepo = store.NouveauFichesRepo(db)
		quizRepo = store.NouveauQuizRepo(db)
		ressourcesRepo = store.NouveauRessourcesRepo(db)
		mindmapRepo = store.NouveauMindmapRepo(db)
		quotasRepo = store.NouveauQuotasRepo(db)
		copieRepo = store.NouveauCopieExamenRepo(db)
		erreurRepo = store.NouveauErreurAnalyseRepo(db)
		log.Println("✓ Repositories initialisés (cours, fiches, quiz, ressources, mindmaps, quotas, copies, erreurs)")
	}

	// Créer le service de génération
	var serviceGeneration *services.ServiceGeneration
	if gestionnaireLLM != nil && coursRepo != nil {
		serviceGeneration = services.NouveauServiceGeneration(gestionnaireLLM, coursRepo, fichesRepo, quizRepo, ressourcesRepo, mindmapRepo)
		log.Println("✓ Service génération initialisé")
	}

	// Créer le service de statistiques
	var serviceStatistiques *services.ServiceStatistiques
	if coursRepo != nil {
		serviceStatistiques = services.NouveauServiceStatistiques(coursRepo, fichesRepo, quizRepo)
		log.Println("✓ Service statistiques initialisé")
	}

	// Créer le service de quotas
	var serviceQuotas *services.ServiceQuotas
	if quotasRepo != nil {
		serviceQuotas = services.NouveauServiceQuotas(quotasRepo, cfg.QuotaOCRJour, cfg.QuotaGenerationJour)
		log.Printf("✓ Service quotas initialisé (OCR: %d/jour, Génération: %d/jour)", cfg.QuotaOCRJour, cfg.QuotaGenerationJour)
	}

	// Créer le service d'analyse des erreurs
	var serviceAnalyseErreurs *services.ServiceAnalyseErreurs
	if gestionnaireLLM != nil && copieRepo != nil {
		serviceAnalyseErreurs = services.NouveauServiceAnalyseErreurs(gestionnaireLLM, copieRepo, erreurRepo, coursRepo)
		log.Println("✓ Service analyse d'erreurs initialisé")
	}

	// Créer le service de recommandations
	var serviceRecommandations *services.ServiceRecommandations
	if gestionnaireLLM != nil {
		serviceRecommandations = services.NouveauServiceRecommandations(gestionnaireLLM, copieRepo, erreurRepo, coursRepo, quizRepo)
		log.Println("✓ Service recommandations initialisé")
	}

	// Créer le service de stockage
	var serviceStorage *services.ServiceStorage
	serviceStorage, err = services.NouveauServiceStorage(cfg.StoragePath)
	if err != nil {
		log.Printf("⚠️  Avertissement: erreur lors de la création du service de stockage: %v", err)
	} else {
		log.Printf("✓ Service stockage initialisé (chemin: %s)", cfg.StoragePath)
	}

	// Créer le routeur Gin
	r := gin.Default()

	// Configurer les middleware
	api.ConfigurerMiddleware(r)

	// Créer les handlers avec toutes les dépendances
	handlers := api.NouveauHandlers(db, serviceOCR, serviceGeneration, serviceStatistiques, serviceQuotas, serviceAnalyseErreurs, serviceRecommandations, serviceStorage, coursRepo, copieRepo, erreurRepo)

	// Configurer les routes
	api.ConfigurerRoutes(r, handlers, serviceQuotas)

	// Démarrer le serveur
	log.Printf("🚀 Serveur démarré sur le port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Erreur démarrage serveur: %v", err)
	}
}
