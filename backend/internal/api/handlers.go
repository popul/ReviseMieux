// Package api contient les handlers HTTP
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handlers contient les dépendances des handlers
type Handlers struct {
	// Les dépendances seront ajoutées ici (store, services, etc.)
}

// NouveauHandlers crée une nouvelle instance de Handlers
func NouveauHandlers() *Handlers {
	return &Handlers{}
}

// HealthHandler retourne l'état de santé du serveur
func (h *Handlers) HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RootHandler retourne les informations de l'API
func (h *Handlers) RootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"nom":     "Révise mieux API",
		"version": "0.1.0",
	})
}

// StatutHandler retourne le statut complet de l'API
func (h *Handlers) StatutHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"statut":        "ok",
		"baseDeDonnees": "connectee", // TODO: vérification réelle
		"version":       "0.1.0",
	})
}

// --- Handlers Cours ---

// ListerCoursHandler liste tous les cours
func (h *Handlers) ListerCoursHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"cours": []interface{}{},
		"total": 0,
	})
}

// CreerCoursHandler crée un nouveau cours
func (h *Handlers) CreerCoursHandler(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"message": "Création de cours pas encore implémentée",
	})
}

// ObtenirCoursHandler retourne un cours par son ID
func (h *Handlers) ObtenirCoursHandler(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"message": "Récupération de cours pas encore implémentée",
	})
}

// --- Handlers OCR ---

// OCRHandler traite une image/PDF pour l'OCR
func (h *Handlers) OCRHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Traitement OCR pas encore implémenté",
	})
}

// --- Handlers Génération ---

// GenererFichesHandler génère des fiches de révision
func (h *Handlers) GenererFichesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Génération de fiches pas encore implémentée",
	})
}

// GenererQuizHandler génère un quiz
func (h *Handlers) GenererQuizHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Génération de quiz pas encore implémentée",
	})
}

// GenererMindmapHandler génère une mindmap
func (h *Handlers) GenererMindmapHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Génération de mindmap pas encore implémentée",
	})
}
