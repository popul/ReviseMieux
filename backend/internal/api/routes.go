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
			cours.DELETE("/:id", h.SupprimerCoursHandler)
			cours.GET("/:id/fiches", h.ObtenirFichesHandler)
			cours.GET("/:id/ressources", h.ObtenirRessourcesHandler)
			cours.GET("/:id/mindmap", h.ObtenirMindmapHandler)
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
			generer.POST("/ressources", h.GenererRessourcesHandler)
		}

		// Routes Quiz
		quiz := api.Group("/quiz")
		{
			quiz.GET("/:id", h.ObtenirQuizHandler)
			quiz.POST("/:id/demarrer", h.DemarrerSessionHandler)
			quiz.POST("/:id/session/:sessionId/repondre", h.RepondreHandler)
			quiz.POST("/:id/session/:sessionId/terminer", h.TerminerSessionHandler)
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

		// Routes Recommandations
		recommandations := api.Group("/recommandations")
		{
			recommandations.GET("/prioritaires", h.ObtenirRecommandationsPrioritairesHandler)
			recommandations.GET("/par-matiere/:matiere", h.ObtenirRecommandationsParMatiereHandler)
		}
	}
}
