// Package api contient les handlers HTTP
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/services"
	"github.com/revisemieux/backend/internal/store"
)

// HandlersPlans gère les endpoints liés aux plans de révision
type HandlersPlans struct {
	servicePlans *services.ServicePlans
}

// NouveauHandlersPlans crée une nouvelle instance des handlers de plans
func NouveauHandlersPlans(servicePlans *services.ServicePlans) *HandlersPlans {
	return &HandlersPlans{
		servicePlans: servicePlans,
	}
}

// RequeteCreerPlan représente le corps de la requête de création
type RequeteCreerPlan struct {
	Titre        string   `json:"titre" binding:"required"`
	Description  string   `json:"description"`
	Matiere      string   `json:"matiere"`
	IconeMatiere string   `json:"iconeMatiere"`
	DateEcheance *string  `json:"dateEcheance"`
	CoursIds     []string `json:"coursIds"`
}

// RequeteMettreAJourPlan représente le corps de la requête de mise à jour
type RequeteMettreAJourPlan struct {
	Titre        string  `json:"titre"`
	Description  string  `json:"description"`
	Matiere      string  `json:"matiere"`
	IconeMatiere string  `json:"iconeMatiere"`
	DateEcheance *string `json:"dateEcheance"`
}

// RequeteAjouterCours représente le corps de la requête d'ajout de cours
type RequeteAjouterCours struct {
	CoursId string `json:"coursId" binding:"required"`
}

// ListerPlansHandler liste les plans (résumés ou complets)
func (h *HandlersPlans) ListerPlansHandler(c *gin.Context) {
	if h.servicePlans == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de plans n'est pas configuré",
			},
		})
		return
	}

	// Si ?resume=true, retourner les résumés pour la sidebar
	if c.Query("resume") == "true" {
		resumes, err := h.servicePlans.ListerResumes(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"succes": false,
				"erreur": gin.H{
					"code":    "ERREUR_INTERNE",
					"message": "Erreur lors de la récupération des plans",
				},
			})
			return
		}

		if resumes == nil {
			resumes = make([]*store.PlanRevisionResume, 0)
		}

		c.JSON(http.StatusOK, gin.H{
			"succes": true,
			"plans":  resumes,
		})
		return
	}

	// Sinon, liste paginée
	page := 1
	limite := 20
	if p := c.Query("page"); p != "" {
		if pInt, err := parseInt(p); err == nil && pInt > 0 {
			page = pInt
		}
	}
	if l := c.Query("limite"); l != "" {
		if lInt, err := parseInt(l); err == nil && lInt > 0 && lInt <= 100 {
			limite = lInt
		}
	}
	offset := (page - 1) * limite

	plans, err := h.servicePlans.ListerPlans(c.Request.Context(), limite, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_INTERNE",
				"message": "Erreur lors de la récupération des plans",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes": true,
		"plans":  plans,
		"page":   page,
		"limite": limite,
	})
}

// CreerPlanHandler crée un nouveau plan de révision
func (h *HandlersPlans) CreerPlanHandler(c *gin.Context) {
	if h.servicePlans == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de plans n'est pas configuré",
			},
		})
		return
	}

	var req RequeteCreerPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "REQUETE_INVALIDE",
				"message": "Données de requête invalides: " + err.Error(),
			},
		})
		return
	}

	icone := req.IconeMatiere
	if icone == "" {
		icone = "📋"
	}

	plan, err := h.servicePlans.CreerPlan(
		c.Request.Context(),
		req.Titre,
		req.Description,
		req.Matiere,
		icone,
		req.DateEcheance,
		req.CoursIds,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_INTERNE",
				"message": "Erreur lors de la création du plan: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"succes": true,
		"plan":   plan,
	})
}

// ObtenirPlanHandler récupère un plan par son ID
func (h *HandlersPlans) ObtenirPlanHandler(c *gin.Context) {
	if h.servicePlans == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de plans n'est pas configuré",
			},
		})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ID_MANQUANT",
				"message": "L'ID du plan est requis",
			},
		})
		return
	}

	plan, err := h.servicePlans.ObtenirPlan(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "PLAN_NON_TROUVE",
				"message": "Plan de révision non trouvé",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes": true,
		"plan":   plan,
	})
}

// ObtenirPlanCompletHandler récupère un plan avec ses cours et compteurs d'artefacts
func (h *HandlersPlans) ObtenirPlanCompletHandler(c *gin.Context) {
	if h.servicePlans == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de plans n'est pas configuré",
			},
		})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ID_MANQUANT",
				"message": "L'ID du plan est requis",
			},
		})
		return
	}

	planComplet, err := h.servicePlans.ObtenirPlanComplet(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "PLAN_NON_TROUVE",
				"message": "Plan de révision non trouvé",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes": true,
		"plan":   planComplet.Plan,
		"cours":  planComplet.Cours,
	})
}

// MettreAJourPlanHandler met à jour un plan existant
func (h *HandlersPlans) MettreAJourPlanHandler(c *gin.Context) {
	if h.servicePlans == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de plans n'est pas configuré",
			},
		})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ID_MANQUANT",
				"message": "L'ID du plan est requis",
			},
		})
		return
	}

	var req RequeteMettreAJourPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "REQUETE_INVALIDE",
				"message": "Données de requête invalides",
			},
		})
		return
	}

	plan, err := h.servicePlans.MettreAJourPlan(
		c.Request.Context(),
		id,
		req.Titre,
		req.Description,
		req.Matiere,
		req.IconeMatiere,
		req.DateEcheance,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "PLAN_NON_TROUVE",
				"message": "Plan de révision non trouvé",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes": true,
		"plan":   plan,
	})
}

// SupprimerPlanHandler supprime un plan
func (h *HandlersPlans) SupprimerPlanHandler(c *gin.Context) {
	if h.servicePlans == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de plans n'est pas configuré",
			},
		})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ID_MANQUANT",
				"message": "L'ID du plan est requis",
			},
		})
		return
	}

	if err := h.servicePlans.SupprimerPlan(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "PLAN_NON_TROUVE",
				"message": "Plan de révision non trouvé",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes":  true,
		"message": "Plan supprimé avec succès",
	})
}

// AjouterCoursHandler ajoute un cours à un plan
func (h *HandlersPlans) AjouterCoursHandler(c *gin.Context) {
	if h.servicePlans == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de plans n'est pas configuré",
			},
		})
		return
	}

	planID := c.Param("id")
	if planID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ID_MANQUANT",
				"message": "L'ID du plan est requis",
			},
		})
		return
	}

	var req RequeteAjouterCours
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "REQUETE_INVALIDE",
				"message": "L'ID du cours est requis",
			},
		})
		return
	}

	if err := h.servicePlans.AjouterCours(c.Request.Context(), planID, req.CoursId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_INTERNE",
				"message": "Erreur lors de l'ajout du cours au plan",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes":  true,
		"message": "Cours ajouté au plan",
	})
}

// RetirerCoursHandler retire un cours d'un plan
func (h *HandlersPlans) RetirerCoursHandler(c *gin.Context) {
	if h.servicePlans == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de plans n'est pas configuré",
			},
		})
		return
	}

	planID := c.Param("id")
	coursID := c.Param("coursId")
	if planID == "" || coursID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ID_MANQUANT",
				"message": "L'ID du plan et du cours sont requis",
			},
		})
		return
	}

	if err := h.servicePlans.RetirerCours(c.Request.Context(), planID, coursID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "COURS_NON_TROUVE_DANS_PLAN",
				"message": "Ce cours n'est pas dans le plan",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes":  true,
		"message": "Cours retiré du plan",
	})
}

// ListerPlansParCoursHandler liste les plans contenant un cours donné
func (h *HandlersPlans) ListerPlansParCoursHandler(c *gin.Context) {
	if h.servicePlans == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de plans n'est pas configuré",
			},
		})
		return
	}

	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ID_MANQUANT",
				"message": "L'ID du cours est requis",
			},
		})
		return
	}

	plans, err := h.servicePlans.ListerPlansParCours(c.Request.Context(), coursID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_INTERNE",
				"message": "Erreur lors de la récupération des plans",
			},
		})
		return
	}

	if plans == nil {
		plans = make([]*store.PlanRevisionResume, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"succes": true,
		"plans":  plans,
	})
}
