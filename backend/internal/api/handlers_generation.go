// Package api contient les handlers HTTP
package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/services"
)

// HandlersGeneration gère les endpoints de génération de contenu
type HandlersGeneration struct {
	serviceGeneration *services.ServiceGeneration
}

// NouveauHandlersGeneration crée une nouvelle instance des handlers de génération
func NouveauHandlersGeneration(serviceGeneration *services.ServiceGeneration) *HandlersGeneration {
	return &HandlersGeneration{
		serviceGeneration: serviceGeneration,
	}
}

// RequeteGenererFiches représente la requête pour générer des fiches
type RequeteGenererFiches struct {
	CoursID      string `json:"coursId" binding:"required"`
	NombreFiches int    `json:"nombreFiches,omitempty"`
	Difficulte   string `json:"difficulte,omitempty"`
}

// ReponseFiches représente la réponse de génération de fiches
type ReponseFiches struct {
	Succes       bool                 `json:"succes"`
	Fiches       []FicheReponse       `json:"fiches,omitempty"`
	NombreGenere int                  `json:"nombreGenere,omitempty"`
	Erreur       *ErreurReponse       `json:"erreur,omitempty"`
}

// FicheReponse représente une fiche dans la réponse API
type FicheReponse struct {
	ID         string `json:"id"`
	Question   string `json:"question"`
	Reponse    string `json:"reponse"`
	Difficulte string `json:"difficulte"`
	Ordre      int    `json:"ordre"`
}

// GenererFichesHandler génère des fiches de révision pour un cours
func (h *HandlersGeneration) GenererFichesHandler(c *gin.Context) {
	// Vérifier que le service est disponible
	if h.serviceGeneration == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseFiches{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de génération n'est pas configuré",
			},
		})
		return
	}

	// Parser la requête
	var req RequeteGenererFiches
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ReponseFiches{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "REQUETE_INVALIDE",
				Message: "Le champ coursId est requis",
			},
		})
		return
	}

	// Valider la difficulté si fournie
	if req.Difficulte != "" && req.Difficulte != "facile" && req.Difficulte != "moyen" && req.Difficulte != "difficile" {
		c.JSON(http.StatusBadRequest, ReponseFiches{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "DIFFICULTE_INVALIDE",
				Message: "La difficulté doit être: facile, moyen ou difficile",
			},
		})
		return
	}

	// Préparer les options
	options := &services.OptionsGenerationFiches{
		NombreFiches: req.NombreFiches,
		Difficulte:   req.Difficulte,
	}

	// Générer les fiches
	resultat, err := h.serviceGeneration.GenererFiches(c.Request.Context(), req.CoursID, options)
	if err != nil {
		h.gererErreurGeneration(c, err)
		return
	}

	// Convertir les fiches pour la réponse
	fichesReponse := make([]FicheReponse, 0, len(resultat.Fiches))
	for _, f := range resultat.Fiches {
		fichesReponse = append(fichesReponse, FicheReponse{
			ID:         f.ID,
			Question:   f.Question,
			Reponse:    f.Reponse,
			Difficulte: f.Difficulte,
			Ordre:      f.Ordre,
		})
	}

	c.JSON(http.StatusOK, ReponseFiches{
		Succes:       true,
		Fiches:       fichesReponse,
		NombreGenere: resultat.NombreGenere,
	})
}

// ObtenirFichesHandler récupère les fiches existantes d'un cours
func (h *HandlersGeneration) ObtenirFichesHandler(c *gin.Context) {
	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, ReponseFiches{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'ID du cours est requis",
			},
		})
		return
	}

	if h.serviceGeneration == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseFiches{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service n'est pas configuré",
			},
		})
		return
	}

	fiches, err := h.serviceGeneration.ObtenirFichesParCours(c.Request.Context(), coursID)
	if err != nil {
		h.gererErreurGeneration(c, err)
		return
	}

	// Convertir les fiches pour la réponse
	fichesReponse := make([]FicheReponse, 0, len(fiches))
	for _, f := range fiches {
		fichesReponse = append(fichesReponse, FicheReponse{
			ID:         f.ID,
			Question:   f.Question,
			Reponse:    f.Reponse,
			Difficulte: f.Difficulte,
			Ordre:      f.Ordre,
		})
	}

	c.JSON(http.StatusOK, ReponseFiches{
		Succes:       true,
		Fiches:       fichesReponse,
		NombreGenere: len(fichesReponse),
	})
}

// gererErreurGeneration gère les erreurs du service de génération
func (h *HandlersGeneration) gererErreurGeneration(c *gin.Context, err error) {
	// Vérifier si c'est une erreur LLM
	var errLLM *llm.ErreurLLM
	if errors.As(err, &errLLM) {
		if errLLM.RateLimited {
			c.JSON(http.StatusTooManyRequests, ReponseFiches{
				Succes: false,
				Erreur: &ErreurReponse{
					Code:    "QUOTA_DEPASSE",
					Message: "Limite d'appels API atteinte, réessayez plus tard",
				},
			})
			return
		}
	}

	// Vérifier les erreurs spécifiques du service
	var errGen *services.ErreurGeneration
	if errors.As(err, &errGen) {
		statusCode := http.StatusInternalServerError
		switch errGen.Code {
		case "COURS_NON_TROUVE":
			statusCode = http.StatusNotFound
		case "COURS_VIDE":
			statusCode = http.StatusBadRequest
		case "SERVICE_NON_DISPONIBLE":
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, ReponseFiches{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    errGen.Code,
				Message: errGen.Message,
			},
		})
		return
	}

	// Erreur générique
	c.JSON(http.StatusInternalServerError, ReponseFiches{
		Succes: false,
		Erreur: &ErreurReponse{
			Code:    "ERREUR_INTERNE",
			Message: "Une erreur est survenue lors de la génération",
		},
	})
}
