// Package api contient la configuration des routes
package api

import (
	"github.com/gin-gonic/gin"
)

// ConfigurerRoutes configure toutes les routes de l'API
func ConfigurerRoutes(r *gin.Engine, h *Handlers) {
	// Routes racine
	r.GET("/health", h.HealthHandler)
	r.GET("/", h.RootHandler)

	// Groupe API
	api := r.Group("/api")
	{
		api.GET("/statut", h.StatutHandler)

		// Routes Cours
		cours := api.Group("/cours")
		{
			cours.GET("", h.ListerCoursHandler)
			cours.POST("", h.CreerCoursHandler)
			cours.GET("/:id", h.ObtenirCoursHandler)
		}

		// Routes OCR
		api.POST("/ocr", h.OCRHandler)

		// Routes Génération
		generer := api.Group("/generer")
		{
			generer.POST("/fiches", h.GenererFichesHandler)
			generer.POST("/quiz", h.GenererQuizHandler)
			generer.POST("/mindmap", h.GenererMindmapHandler)
		}
	}
}
