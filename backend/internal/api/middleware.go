// Package api contient les handlers et middleware HTTP
package api

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// ConfigurerMiddleware configure les middleware globaux
func ConfigurerMiddleware(r *gin.Engine) {
	// Recovery middleware - récupère des panics
	r.Use(gin.Recovery())

	// Logger middleware personnalisé
	r.Use(loggerMiddleware())

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           5 * time.Minute,
	}))
}

// loggerMiddleware crée un middleware de logging personnalisé
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		debut := time.Now()
		chemin := c.Request.URL.Path
		methode := c.Request.Method

		// Traiter la requête
		c.Next()

		// Logger après traitement
		duree := time.Since(debut)
		statut := c.Writer.Status()

		log.Printf("[%s] %s %s %d %v",
			methode,
			chemin,
			c.ClientIP(),
			statut,
			duree,
		)
	}
}
