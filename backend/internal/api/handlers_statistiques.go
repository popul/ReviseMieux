// Package api contient les handlers HTTP
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/services"
)

// HandlersStatistiques gère les endpoints de statistiques
type HandlersStatistiques struct {
	serviceStatistiques *services.ServiceStatistiques
}

// NouveauHandlersStatistiques crée une nouvelle instance des handlers de statistiques
func NouveauHandlersStatistiques(serviceStatistiques *services.ServiceStatistiques) *HandlersStatistiques {
	return &HandlersStatistiques{
		serviceStatistiques: serviceStatistiques,
	}
}

// ReponseStatistiques représente la réponse de l'endpoint /api/statistiques
type ReponseStatistiques struct {
	Succes        bool                        `json:"succes"`
	Statistiques  *StatistiquesReponse        `json:"statistiques,omitempty"`
	Erreur        *ErreurReponse              `json:"erreur,omitempty"`
}

// StatistiquesReponse représente les statistiques dans la réponse API
type StatistiquesReponse struct {
	NombreCours   int      `json:"nombreCours"`
	NombreFiches  int      `json:"nombreFiches"`
	NombreQuiz    int      `json:"nombreQuiz"`
	QuizCompletes int      `json:"quizCompletes"`
	ScoreMoyen    *float64 `json:"scoreMoyen,omitempty"`
}

// ReponseCoursRecents représente la réponse de l'endpoint /api/cours/recents
type ReponseCoursRecents struct {
	Succes bool                `json:"succes"`
	Cours  []CoursResumeReponse `json:"cours"`
	Erreur *ErreurReponse      `json:"erreur,omitempty"`
}

// CoursResumeReponse représente un résumé de cours dans la réponse API
type CoursResumeReponse struct {
	ID               string `json:"id"`
	Titre            string `json:"titre"`
	Matiere          string `json:"matiere,omitempty"`
	NombreFiches     int    `json:"nombreFiches"`
	NombreQuiz       int    `json:"nombreQuiz"`
	DateCreation     string `json:"dateCreation"`
	DateModification string `json:"dateModification"`
}

// ObtenirStatistiquesHandler retourne les statistiques globales
func (h *HandlersStatistiques) ObtenirStatistiquesHandler(c *gin.Context) {
	// Vérifier que le service est disponible
	if h.serviceStatistiques == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseStatistiques{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de statistiques n'est pas configuré",
			},
		})
		return
	}

	// Récupérer les statistiques
	stats, err := h.serviceStatistiques.ObtenirStatistiques(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ReponseStatistiques{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ERREUR_STATISTIQUES",
				Message: "Erreur lors de la récupération des statistiques",
			},
		})
		return
	}

	c.JSON(http.StatusOK, ReponseStatistiques{
		Succes: true,
		Statistiques: &StatistiquesReponse{
			NombreCours:   stats.NombreCours,
			NombreFiches:  stats.NombreFiches,
			NombreQuiz:    stats.NombreQuiz,
			QuizCompletes: stats.QuizCompletes,
			ScoreMoyen:    stats.ScoreMoyen,
		},
	})
}

// ObtenirCoursRecentsHandler retourne les cours récents avec leurs statistiques
func (h *HandlersStatistiques) ObtenirCoursRecentsHandler(c *gin.Context) {
	// Vérifier que le service est disponible
	if h.serviceStatistiques == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseCoursRecents{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de statistiques n'est pas configuré",
			},
		})
		return
	}

	// Récupérer les cours récents (max 10)
	coursRecents, err := h.serviceStatistiques.ListerCoursRecents(c.Request.Context(), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ReponseCoursRecents{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ERREUR_COURS_RECENTS",
				Message: "Erreur lors de la récupération des cours récents",
			},
		})
		return
	}

	// Convertir en format réponse
	coursReponse := make([]CoursResumeReponse, len(coursRecents))
	for i, c := range coursRecents {
		coursReponse[i] = CoursResumeReponse{
			ID:               c.ID,
			Titre:            c.Titre,
			Matiere:          c.Matiere,
			NombreFiches:     c.NombreFiches,
			NombreQuiz:       c.NombreQuiz,
			DateCreation:     c.DateCreation.Format("2006-01-02T15:04:05Z07:00"),
			DateModification: c.DateModification.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, ReponseCoursRecents{
		Succes: true,
		Cours:  coursReponse,
	})
}
