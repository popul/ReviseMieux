// Package api contient les handlers HTTP
package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/services"
	"github.com/revisemieux/backend/internal/store"
)

// HandlersLexique gere les endpoints lies au lexique
type HandlersLexique struct {
	serviceLexique *services.ServiceLexique
}

// NouveauHandlersLexique cree une nouvelle instance des handlers de lexique
func NouveauHandlersLexique(serviceLexique *services.ServiceLexique) *HandlersLexique {
	return &HandlersLexique{
		serviceLexique: serviceLexique,
	}
}

// ReponseLexique represente la reponse API pour le lexique
type ReponseLexique struct {
	Succes         bool              `json:"succes"`
	Termes         []TermeLexiqueAPI `json:"termes,omitempty"`
	NombreExtraits int               `json:"nombreExtraits,omitempty"`
	Erreur         *ErreurReponse    `json:"erreur,omitempty"`
}

// TermeLexiqueAPI represente un terme dans la reponse API
type TermeLexiqueAPI struct {
	ID         string `json:"id"`
	CoursID    string `json:"coursId"`
	Terme      string `json:"terme"`
	Definition string `json:"definition"`
	Contexte   string `json:"contexte,omitempty"`
	Exemple    string `json:"exemple,omitempty"`
	Categorie  string `json:"categorie,omitempty"`
	Maitrise   int    `json:"maitrise"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// RequeteMaitrise represente la requete de mise a jour de maitrise
type RequeteMaitrise struct {
	Maitrise int `json:"maitrise"`
}

// RequeteQuizVocabulaire represente la requete pour generer un quiz
type RequeteQuizVocabulaire struct {
	Mode string `json:"mode" binding:"required"`
}

// ReponseQuizVocabulaire represente la reponse API pour un quiz vocabulaire
type ReponseQuizVocabulaire struct {
	Succes bool                       `json:"succes"`
	Quiz   *services.QuizVocabulaire  `json:"quiz,omitempty"`
	Erreur *ErreurReponse             `json:"erreur,omitempty"`
}

// ExtraireTermesHandler extrait les termes d un cours via le LLM
func (h *HandlersLexique) ExtraireTermesHandler(c *gin.Context) {
	if h.serviceLexique == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseLexique{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de lexique n est pas configure",
			},
		})
		return
	}

	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, ReponseLexique{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L ID du cours est requis",
			},
		})
		return
	}

	resultat, err := h.serviceLexique.ExtraireTermes(c.Request.Context(), coursID)
	if err != nil {
		h.gererErreurLexique(c, err)
		return
	}

	termesReponse := make([]TermeLexiqueAPI, 0, len(resultat.Termes))
	for _, terme := range resultat.Termes {
		termesReponse = append(termesReponse, convertirTermeEnReponse(terme))
	}

	c.JSON(http.StatusOK, ReponseLexique{
		Succes:         true,
		Termes:         termesReponse,
		NombreExtraits: resultat.NombreExtraits,
	})
}

// ListerTermesHandler recupere les termes existants d un cours
func (h *HandlersLexique) ListerTermesHandler(c *gin.Context) {
	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, ReponseLexique{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L ID du cours est requis",
			},
		})
		return
	}

	if h.serviceLexique == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseLexique{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de lexique n est pas configure",
			},
		})
		return
	}

	termes, err := h.serviceLexique.GetTermesByCours(c.Request.Context(), coursID)
	if err != nil {
		h.gererErreurLexique(c, err)
		return
	}

	termesReponse := make([]TermeLexiqueAPI, 0, len(termes))
	for _, terme := range termes {
		termesReponse = append(termesReponse, convertirTermeEnReponse(terme))
	}

	c.JSON(http.StatusOK, ReponseLexique{
		Succes:         true,
		Termes:         termesReponse,
		NombreExtraits: len(termesReponse),
	})
}

// MettreAJourMaitriseHandler met a jour le niveau de maitrise d un terme
func (h *HandlersLexique) MettreAJourMaitriseHandler(c *gin.Context) {
	if h.serviceLexique == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de lexique n est pas configure",
			},
		})
		return
	}

	termeID := c.Param("id")
	if termeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ID_MANQUANT",
				"message": "L ID du terme est requis",
			},
		})
		return
	}

	var req RequeteMaitrise
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "REQUETE_INVALIDE",
				"message": "Donnees de requete invalides",
			},
		})
		return
	}

	if err := h.serviceLexique.MettreAJourMaitrise(c.Request.Context(), termeID, req.Maitrise); err != nil {
		h.gererErreurLexique(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes":   true,
		"message":  "Maitrise mise a jour",
		"maitrise": req.Maitrise,
	})
}

// GenererQuizVocabulaireHandler genere un quiz de vocabulaire
func (h *HandlersLexique) GenererQuizVocabulaireHandler(c *gin.Context) {
	if h.serviceLexique == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseQuizVocabulaire{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de lexique n est pas configure",
			},
		})
		return
	}

	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, ReponseQuizVocabulaire{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L ID du cours est requis",
			},
		})
		return
	}

	var req RequeteQuizVocabulaire
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ReponseQuizVocabulaire{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "REQUETE_INVALIDE",
				Message: "Le mode de quiz est requis (terme_vers_definition ou definition_vers_terme)",
			},
		})
		return
	}

	quiz, err := h.serviceLexique.GenererQuizVocabulaire(c.Request.Context(), coursID, req.Mode)
	if err != nil {
		h.gererErreurLexique(c, err)
		return
	}

	c.JSON(http.StatusOK, ReponseQuizVocabulaire{
		Succes: true,
		Quiz:   quiz,
	})
}

// convertirTermeEnReponse convertit un terme store en reponse API
func convertirTermeEnReponse(terme *store.TermeLexique) TermeLexiqueAPI {
	return TermeLexiqueAPI{
		ID:         terme.ID,
		CoursID:    terme.CoursID,
		Terme:      terme.Terme,
		Definition: terme.Definition,
		Contexte:   terme.Contexte,
		Exemple:    terme.Exemple,
		Categorie:  terme.Categorie,
		Maitrise:   terme.Maitrise,
		CreatedAt:  terme.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  terme.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// gererErreurLexique gere les erreurs specifiques au lexique
func (h *HandlersLexique) gererErreurLexique(c *gin.Context, err error) {
	var errGen *services.ErreurGeneration
	if errors.As(err, &errGen) {
		statusCode := http.StatusInternalServerError
		switch errGen.Code {
		case "COURS_NON_TROUVE", "TERME_NON_TROUVE":
			statusCode = http.StatusNotFound
		case "COURS_VIDE":
			statusCode = http.StatusBadRequest
		case "MAITRISE_INVALIDE", "MODE_QUIZ_INVALIDE":
			statusCode = http.StatusBadRequest
		case "SERVICE_NON_DISPONIBLE":
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, ReponseLexique{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    errGen.Code,
				Message: errGen.Message,
			},
		})
		return
	}

	var errLLM *llm.ErreurLLM
	if errors.As(err, &errLLM) {
		if errLLM.RateLimited {
			c.JSON(http.StatusTooManyRequests, ReponseLexique{
				Succes: false,
				Erreur: &ErreurReponse{
					Code:    "QUOTA_DEPASSE",
					Message: "Limite d appels API atteinte, reessayez plus tard",
				},
			})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, ReponseLexique{
		Succes: false,
		Erreur: &ErreurReponse{
			Code:    "ERREUR_INTERNE",
			Message: "Une erreur est survenue lors du traitement du lexique",
		},
	})
}
