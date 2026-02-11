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

// HandlersConcepts gère les endpoints liés aux concepts
type HandlersConcepts struct {
	serviceConcepts *services.ServiceConcepts
}

// NouveauHandlersConcepts crée une nouvelle instance des handlers de concepts
func NouveauHandlersConcepts(serviceConcepts *services.ServiceConcepts) *HandlersConcepts {
	return &HandlersConcepts{
		serviceConcepts: serviceConcepts,
	}
}

// ReponseConcepts représente la réponse API pour les concepts
type ReponseConcepts struct {
	Succes         bool              `json:"succes"`
	Concepts       []ConceptReponse  `json:"concepts,omitempty"`
	NombreExtraits int               `json:"nombreExtraits,omitempty"`
	Erreur         *ErreurReponse    `json:"erreur,omitempty"`
}

// ConceptReponse représente un concept dans la réponse API
type ConceptReponse struct {
	ID                string              `json:"id"`
	CoursID           string              `json:"coursId"`
	Nom               string              `json:"nom"`
	Definition        string              `json:"definition"`
	Importance        string              `json:"importance"`
	PositionDansCours *PositionConceptAPI `json:"positionDansCours,omitempty"`
	CreatedAt         string              `json:"createdAt"`
	UpdatedAt         string              `json:"updatedAt"`
}

// PositionConceptAPI représente la position d'un concept dans la réponse API
type PositionConceptAPI struct {
	Debut int `json:"debut"`
	Fin   int `json:"fin"`
}

// ReponseConceptUnique représente la réponse API pour un seul concept
type ReponseConceptUnique struct {
	Succes  bool             `json:"succes"`
	Concept *ConceptReponse  `json:"concept,omitempty"`
	Erreur  *ErreurReponse   `json:"erreur,omitempty"`
}

// RequeteMettreAJourConcept représente la requête de mise à jour d'un concept
type RequeteMettreAJourConcept struct {
	Nom               string              `json:"nom"`
	Definition        string              `json:"definition"`
	Importance        string              `json:"importance"`
	PositionDansCours *PositionConceptAPI `json:"positionDansCours,omitempty"`
}

// ExtraireConceptsHandler extrait les concepts d'un cours via le LLM
func (h *HandlersConcepts) ExtraireConceptsHandler(c *gin.Context) {
	if h.serviceConcepts == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseConcepts{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de concepts n'est pas configuré",
			},
		})
		return
	}

	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, ReponseConcepts{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'ID du cours est requis",
			},
		})
		return
	}

	resultat, err := h.serviceConcepts.ExtraireConcepts(c.Request.Context(), coursID)
	if err != nil {
		h.gererErreurConcepts(c, err)
		return
	}

	// Convertir les concepts pour la réponse
	conceptsReponse := make([]ConceptReponse, 0, len(resultat.Concepts))
	for _, concept := range resultat.Concepts {
		conceptsReponse = append(conceptsReponse, convertirConceptEnReponse(concept))
	}

	c.JSON(http.StatusOK, ReponseConcepts{
		Succes:         true,
		Concepts:       conceptsReponse,
		NombreExtraits: resultat.NombreExtraits,
	})
}

// ListerConceptsHandler récupère les concepts existants d'un cours
func (h *HandlersConcepts) ListerConceptsHandler(c *gin.Context) {
	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, ReponseConcepts{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'ID du cours est requis",
			},
		})
		return
	}

	if h.serviceConcepts == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseConcepts{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de concepts n'est pas configuré",
			},
		})
		return
	}

	concepts, err := h.serviceConcepts.GetConceptsByCours(c.Request.Context(), coursID)
	if err != nil {
		h.gererErreurConcepts(c, err)
		return
	}

	// Convertir les concepts pour la réponse
	conceptsReponse := make([]ConceptReponse, 0, len(concepts))
	for _, concept := range concepts {
		conceptsReponse = append(conceptsReponse, convertirConceptEnReponse(concept))
	}

	c.JSON(http.StatusOK, ReponseConcepts{
		Succes:         true,
		Concepts:       conceptsReponse,
		NombreExtraits: len(conceptsReponse),
	})
}

// MettreAJourConceptHandler met à jour un concept
func (h *HandlersConcepts) MettreAJourConceptHandler(c *gin.Context) {
	if h.serviceConcepts == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseConceptUnique{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de concepts n'est pas configuré",
			},
		})
		return
	}

	conceptID := c.Param("id")
	if conceptID == "" {
		c.JSON(http.StatusBadRequest, ReponseConceptUnique{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'ID du concept est requis",
			},
		})
		return
	}

	var req RequeteMettreAJourConcept
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ReponseConceptUnique{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "REQUETE_INVALIDE",
				Message: "Données de requête invalides",
			},
		})
		return
	}

	// Valider l'importance si fournie
	if req.Importance != "" && req.Importance != "essentiel" && req.Importance != "important" && req.Importance != "secondaire" {
		c.JSON(http.StatusBadRequest, ReponseConceptUnique{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "IMPORTANCE_INVALIDE",
				Message: "L'importance doit être: essentiel, important ou secondaire",
			},
		})
		return
	}

	// Mettre à jour le concept via le service
	concept, err := h.serviceConcepts.MettreAJourConcept(c.Request.Context(), conceptID, req.Nom, req.Definition, req.Importance, req.PositionDansCours)
	if err != nil {
		h.gererErreurConcepts(c, err)
		return
	}

	c.JSON(http.StatusOK, ReponseConceptUnique{
		Succes:  true,
		Concept: ptrConceptReponse(convertirConceptEnReponse(concept)),
	})
}

// SupprimerConceptHandler supprime un concept
func (h *HandlersConcepts) SupprimerConceptHandler(c *gin.Context) {
	if h.serviceConcepts == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de concepts n'est pas configuré",
			},
		})
		return
	}

	conceptID := c.Param("id")
	if conceptID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ID_MANQUANT",
				"message": "L'ID du concept est requis",
			},
		})
		return
	}

	if err := h.serviceConcepts.SupprimerConcept(c.Request.Context(), conceptID); err != nil {
		h.gererErreurConcepts(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes":  true,
		"message": "Concept supprimé avec succès",
	})
}

// convertirConceptEnReponse convertit un concept store en réponse API
func convertirConceptEnReponse(concept *store.Concept) ConceptReponse {
	reponse := ConceptReponse{
		ID:         concept.ID,
		CoursID:    concept.CoursID,
		Nom:        concept.Nom,
		Definition: concept.Definition,
		Importance: concept.Importance,
		CreatedAt:  concept.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  concept.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if concept.PositionDansCours != nil {
		reponse.PositionDansCours = &PositionConceptAPI{
			Debut: concept.PositionDansCours.Debut,
			Fin:   concept.PositionDansCours.Fin,
		}
	}

	return reponse
}

// ptrConceptReponse retourne un pointeur vers ConceptReponse
func ptrConceptReponse(c ConceptReponse) *ConceptReponse {
	return &c
}

// gererErreurConcepts gère les erreurs spécifiques aux concepts
func (h *HandlersConcepts) gererErreurConcepts(c *gin.Context, err error) {
	var errGen *services.ErreurGeneration
	if errors.As(err, &errGen) {
		statusCode := http.StatusInternalServerError
		switch errGen.Code {
		case "COURS_NON_TROUVE", "CONCEPT_NON_TROUVE":
			statusCode = http.StatusNotFound
		case "COURS_VIDE":
			statusCode = http.StatusBadRequest
		case "SERVICE_NON_DISPONIBLE":
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, ReponseConcepts{
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
			c.JSON(http.StatusTooManyRequests, ReponseConcepts{
				Succes: false,
				Erreur: &ErreurReponse{
					Code:    "QUOTA_DEPASSE",
					Message: "Limite d'appels API atteinte, réessayez plus tard",
				},
			})
			return
		}
	}

	c.JSON(http.StatusInternalServerError, ReponseConcepts{
		Succes: false,
		Erreur: &ErreurReponse{
			Code:    "ERREUR_INTERNE",
			Message: "Une erreur est survenue lors du traitement des concepts",
		},
	})
}
