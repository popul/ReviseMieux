// Package api contient les handlers HTTP
package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/services"
)

// HandlersQuotas gère les endpoints liés aux quotas
type HandlersQuotas struct {
	serviceQuotas *services.ServiceQuotas
}

// NouveauHandlersQuotas crée un nouveau handler de quotas
func NouveauHandlersQuotas(serviceQuotas *services.ServiceQuotas) *HandlersQuotas {
	return &HandlersQuotas{
		serviceQuotas: serviceQuotas,
	}
}

// ReponseQuotas représente la réponse de l'endpoint quotas
type ReponseQuotas struct {
	Succes bool                   `json:"succes"`
	Quotas *StatutQuotaReponse    `json:"quotas,omitempty"`
	Erreur *ErreurQuotaReponse    `json:"erreur,omitempty"`
}

// StatutQuotaReponse représente les quotas dans la réponse JSON
type StatutQuotaReponse struct {
	PagesOCRUtilisees     int `json:"pagesOcrUtilisees"`
	PagesOCRMax           int `json:"pagesOcrMax"`
	GenerationsUtilisees  int `json:"generationsUtilisees"`
	GenerationsMax        int `json:"generationsMax"`
	PagesOCRRestantes     int `json:"pagesOcrRestantes"`
	GenerationsRestantes  int `json:"generationsRestantes"`
}

// ErreurQuotaReponse représente une erreur dans la réponse
type ErreurQuotaReponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ObtenirQuotasHandler retourne l'état actuel des quotas
func (h *HandlersQuotas) ObtenirQuotasHandler(c *gin.Context) {
	if h.serviceQuotas == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseQuotas{
			Succes: false,
			Erreur: &ErreurQuotaReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de quotas n'est pas configuré",
			},
		})
		return
	}

	statut, err := h.serviceQuotas.ObtenirStatut(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ReponseQuotas{
			Succes: false,
			Erreur: &ErreurQuotaReponse{
				Code:    "ERREUR_INTERNE",
				Message: "Erreur lors de la récupération des quotas",
			},
		})
		return
	}

	c.JSON(http.StatusOK, ReponseQuotas{
		Succes: true,
		Quotas: &StatutQuotaReponse{
			PagesOCRUtilisees:     statut.PagesOCRUtilisees,
			PagesOCRMax:           statut.PagesOCRMax,
			GenerationsUtilisees:  statut.GenerationsUtilisees,
			GenerationsMax:        statut.GenerationsMax,
			PagesOCRRestantes:     statut.PagesOCRRestantes,
			GenerationsRestantes:  statut.GenerationsRestantes,
		},
	})
}

// MiddlewareVerificationQuotaOCR vérifie le quota OCR avant traitement
func MiddlewareVerificationQuotaOCR(serviceQuotas *services.ServiceQuotas) gin.HandlerFunc {
	return func(c *gin.Context) {
		if serviceQuotas == nil {
			c.Next()
			return
		}

		// Estimer le nombre de pages (1 par défaut, sera ajusté si multipart)
		nombrePages := 1

		err := serviceQuotas.VerifierQuotaOCR(c.Request.Context(), nombrePages)
		if err != nil {
			if errors.Is(err, services.ErrQuotaOCRDepasse) {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"succes": false,
					"erreur": gin.H{
						"code":    "QUOTA_DEPASSE",
						"message": "Quota OCR journalier atteint. Réessayez demain.",
					},
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"succes": false,
				"erreur": gin.H{
					"code":    "ERREUR_INTERNE",
					"message": "Erreur lors de la vérification des quotas",
				},
			})
			return
		}

		c.Next()
	}
}

// MiddlewareVerificationQuotaGeneration vérifie le quota de génération avant traitement
func MiddlewareVerificationQuotaGeneration(serviceQuotas *services.ServiceQuotas) gin.HandlerFunc {
	return func(c *gin.Context) {
		if serviceQuotas == nil {
			c.Next()
			return
		}

		err := serviceQuotas.VerifierQuotaGeneration(c.Request.Context())
		if err != nil {
			if errors.Is(err, services.ErrQuotaGenerationDepasse) {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"succes": false,
					"erreur": gin.H{
						"code":    "QUOTA_DEPASSE",
						"message": "Quota de génération journalier atteint. Réessayez demain.",
					},
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"succes": false,
				"erreur": gin.H{
					"code":    "ERREUR_INTERNE",
					"message": "Erreur lors de la vérification des quotas",
				},
			})
			return
		}

		c.Next()
	}
}
