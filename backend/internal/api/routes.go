// Package api contient la configuration des routes
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/services"
)

// ConfigurerRoutes configure toutes les routes de l'API
func ConfigurerRoutes(r *gin.Engine, h *Handlers, serviceQuotas *services.ServiceQuotas) {
	// Routes racine
	r.GET("/health", h.HealthHandler)
	r.GET("/", h.RootHandler)

	// Groupe API
	api := r.Group("/api")
	{
		api.GET("/config", h.ConfigFrontendHandler)
		api.GET("/statut", h.StatutHandler)
		api.GET("/statistiques", h.ObtenirStatistiquesHandler)
		api.GET("/progression", h.ObtenirProgressionHandler)
		api.GET("/quotas", h.ObtenirQuotasHandler)

		// Routes Cours
		cours := api.Group("/cours")
		{
			cours.GET("", h.ListerCoursHandler)
			cours.GET("/recents", h.ObtenirCoursRecentsHandler)
			cours.POST("", h.CreerCoursHandler)
			cours.GET("/:id", h.ObtenirCoursHandler)
			cours.PUT("/:id", h.MettreAJourCoursHandler)
			cours.DELETE("/:id", h.SupprimerCoursHandler)
			cours.GET("/:id/fiches", h.ObtenirFichesHandler)
			cours.GET("/:id/mindmap", h.ObtenirMindmapHandler)

			// Routes Résumé
			cours.POST("/:id/resume/generer", MiddlewareVerificationQuotaGeneration(serviceQuotas), h.GenererResumeHandler)

			// Routes Re-OCR
			cours.POST("/:id/reocr", MiddlewareVerificationQuotaOCR(serviceQuotas), h.RetraiterOCRCoursHandler)

			// Routes Concepts
			cours.POST("/:id/concepts/extraire", MiddlewareVerificationQuotaGeneration(serviceQuotas), h.ExtraireConceptsHandler)
			cours.GET("/:id/concepts", h.ListerConceptsHandler)

			// Routes Lexique
			cours.POST("/:id/lexique/extraire", MiddlewareVerificationQuotaGeneration(serviceQuotas), h.ExtraireTermesLexiqueHandler)
			cours.GET("/:id/lexique", h.ListerTermesLexiqueHandler)
			cours.POST("/:id/lexique/quiz", h.GenererQuizVocabulaireHandler)

			// Routes Examen Blanc
			cours.POST("/:id/examen/generer", MiddlewareVerificationQuotaGeneration(serviceQuotas), h.GenererExamenHandler)

			// Routes Plans par cours
			cours.GET("/:id/plans", h.ListerPlansParCoursHandler)

			// Routes Images
			cours.GET("/:id/images", h.ListerImagesHandler)
			cours.GET("/:id/images/:filename", h.ServirImageHandler)
			cours.POST("/:id/images", h.AjouterImageHandler)
			cours.DELETE("/:id/images/:filename", h.SupprimerImageHandler)
			cours.PUT("/:id/images/ordre", h.ReordonnerImagesHandler)
			cours.POST("/:id/images/:filename/deplacer", h.DeplacerImageHandler)
		}

		// Routes OCR (avec middleware quota)
		api.POST("/ocr", MiddlewareVerificationQuotaOCR(serviceQuotas), h.OCRHandler)

		// Routes Génération (avec middleware quota)
		generer := api.Group("/generer")
		generer.Use(MiddlewareVerificationQuotaGeneration(serviceQuotas))
		{
			generer.POST("/fiches", h.GenererFichesHandler)
			generer.POST("/quiz", h.GenererQuizHandler)
			generer.POST("/mindmap", h.GenererMindmapHandler)
		}

		// Routes Quiz
		quiz := api.Group("/quiz")
		{
			quiz.GET("/:id", h.ObtenirQuizHandler)
			quiz.POST("/:id/demarrer", h.DemarrerSessionHandler)
			quiz.POST("/:id/session/:sessionId/repondre", h.RepondreHandler)
			quiz.POST("/:id/session/:sessionId/terminer", h.TerminerSessionHandler)
		}

		// Routes Concepts (modification/suppression)
		concepts := api.Group("/concepts")
		{
			concepts.PUT("/:id", h.MettreAJourConceptHandler)
			concepts.DELETE("/:id", h.SupprimerConceptHandler)
		}

		// Routes Lexique (modification maitrise)
		lexique := api.Group("/lexique")
		{
			lexique.PUT("/:id/maitrise", h.MettreAJourMaitriseHandler)
		}


		// Routes Examens Blancs
		examens := api.Group("/examens")
		{
			examens.GET("/:id", h.ObtenirExamenHandler)
			examens.POST("/:id/sessions", h.DemarrerSessionExamenHandler)
		}

		// Routes Sessions Examen
		sessionsExamen := api.Group("/sessions-examen")
		{
			sessionsExamen.GET("/:id", h.ObtenirSessionExamenHandler)
			sessionsExamen.POST("/:id/indice", h.DemanderIndiceHandler)
			sessionsExamen.POST("/:id/corriger", MiddlewareVerificationQuotaGeneration(serviceQuotas), h.CorrigerExamenHandler)
		}

		// Routes Copies d'examens
		copies := api.Group("/copies")
		{
			copies.GET("", h.ListerCopiesHandler)
			copies.POST("/ocr", MiddlewareVerificationQuotaOCR(serviceQuotas), h.TraiterOCRCopieHandler)
			copies.GET("/:id", h.ObtenirCopieHandler)
			copies.DELETE("/:id", h.SupprimerCopieHandler)
			copies.POST("/:id/analyser", MiddlewareVerificationQuotaGeneration(serviceQuotas), h.AnalyserCopieHandler)
			copies.GET("/:id/erreurs", h.ObtenirErreursHandler)
			copies.POST("/:id/recommandations", MiddlewareVerificationQuotaGeneration(serviceQuotas), h.GenererRecommandationsCopieHandler)
		}

		// Routes Plans de Révision
		plans := api.Group("/plans")
		{
			plans.GET("", h.ListerPlansHandler)
			plans.POST("", h.CreerPlanHandler)
			plans.GET("/:id", h.ObtenirPlanHandler)
			plans.GET("/:id/complet", h.ObtenirPlanCompletHandler)
			plans.PUT("/:id", h.MettreAJourPlanHandler)
			plans.DELETE("/:id", h.SupprimerPlanHandler)
			plans.POST("/:id/cours", h.AjouterCoursAuPlanHandler)
			plans.DELETE("/:id/cours/:coursId", h.RetirerCoursDuPlanHandler)
		}

		// Routes Recommandations
		recommandations := api.Group("/recommandations")
		{
			recommandations.GET("/prioritaires", h.ObtenirRecommandationsPrioritairesHandler)
			recommandations.GET("/par-matiere/:matiere", h.ObtenirRecommandationsParMatiereHandler)
		}
	}
}
