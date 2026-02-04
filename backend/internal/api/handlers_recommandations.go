// Package api contient les handlers HTTP de l'API
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/services"
)

// HandlersRecommandations gère les endpoints de recommandations
type HandlersRecommandations struct {
	serviceRecommandations *services.ServiceRecommandations
}

// NouveauHandlersRecommandations crée une nouvelle instance des handlers
func NouveauHandlersRecommandations(serviceRecommandations *services.ServiceRecommandations) *HandlersRecommandations {
	return &HandlersRecommandations{
		serviceRecommandations: serviceRecommandations,
	}
}

// ReponseRecommandations représente la réponse de l'API pour les recommandations
type ReponseRecommandations struct {
	Succes          bool                             `json:"succes"`
	Recommandations *services.ResultatRecommandations `json:"recommandations,omitempty"`
	Erreur          *ErreurReponse                   `json:"erreur,omitempty"`
}

// GenererRecommandationsCopieHandler génère des recommandations basées sur une copie analysée
// POST /api/copies/:id/recommandations
func (h *HandlersRecommandations) GenererRecommandationsCopieHandler(c *gin.Context) {
	copieID := c.Param("id")
	if copieID == "" {
		c.JSON(http.StatusBadRequest, ReponseRecommandations{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'identifiant de la copie est requis",
			},
		})
		return
	}

	// Options depuis query params
	inclureQuiz := c.DefaultQuery("inclure_quiz", "true") == "true"

	source := &services.SourceRecommandation{
		CopieID:            copieID,
		IncludeQuizResults: inclureQuiz,
	}

	resultat, err := h.serviceRecommandations.GenererRecommandations(c.Request.Context(), source)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errCode := "ERREUR_GENERATION"
		errMessage := err.Error()

		// Gérer les erreurs spécifiques
		if err == services.ErrDonneesInsuffisantes {
			statusCode = http.StatusBadRequest
			errCode = "DONNEES_INSUFFISANTES"
			errMessage = "Pas assez de données pour générer des recommandations. Analysez d'abord la copie."
		}

		c.JSON(statusCode, ReponseRecommandations{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    errCode,
				Message: errMessage,
			},
		})
		return
	}

	c.JSON(http.StatusOK, ReponseRecommandations{
		Succes:          true,
		Recommandations: resultat,
	})
}

// ObtenirRecommandationsPrioritairesHandler retourne les recommandations prioritaires globales
// GET /api/recommandations/prioritaires
func (h *HandlersRecommandations) ObtenirRecommandationsPrioritairesHandler(c *gin.Context) {
	matiere := c.Query("matiere")

	resultat, err := h.serviceRecommandations.GenererRecommandationsPrioritaires(c.Request.Context(), matiere)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errCode := "ERREUR_GENERATION"
		errMessage := err.Error()

		if err == services.ErrDonneesInsuffisantes {
			statusCode = http.StatusBadRequest
			errCode = "DONNEES_INSUFFISANTES"
			errMessage = "Pas assez de données pour générer des recommandations. Complétez des quiz ou analysez des copies d'abord."
		}

		c.JSON(statusCode, ReponseRecommandations{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    errCode,
				Message: errMessage,
			},
		})
		return
	}

	c.JSON(http.StatusOK, ReponseRecommandations{
		Succes:          true,
		Recommandations: resultat,
	})
}

// ObtenirRecommandationsParMatiereHandler retourne les recommandations pour une matière spécifique
// GET /api/recommandations/par-matiere/:matiere
func (h *HandlersRecommandations) ObtenirRecommandationsParMatiereHandler(c *gin.Context) {
	matiere := c.Param("matiere")
	if matiere == "" {
		c.JSON(http.StatusBadRequest, ReponseRecommandations{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "MATIERE_MANQUANTE",
				Message: "La matière est requise",
			},
		})
		return
	}

	source := &services.SourceRecommandation{
		Matiere:            matiere,
		IncludeQuizResults: true,
	}

	resultat, err := h.serviceRecommandations.GenererRecommandations(c.Request.Context(), source)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errCode := "ERREUR_GENERATION"
		errMessage := err.Error()

		if err == services.ErrDonneesInsuffisantes {
			statusCode = http.StatusBadRequest
			errCode = "DONNEES_INSUFFISANTES"
			errMessage = "Pas assez de données pour cette matière. Complétez des quiz ou analysez des copies."
		}

		c.JSON(statusCode, ReponseRecommandations{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    errCode,
				Message: errMessage,
			},
		})
		return
	}

	c.JSON(http.StatusOK, ReponseRecommandations{
		Succes:          true,
		Recommandations: resultat,
	})
}
