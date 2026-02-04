// Package api contient les handlers HTTP
package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/services"
	"github.com/revisemieux/backend/internal/store"
)

// Handlers contient les dépendances des handlers
type Handlers struct {
	store                *store.Store
	coursRepo            store.CoursRepository
	handlersOCR          *HandlersOCR
	handlersGeneration   *HandlersGeneration
	handlersStatistiques *HandlersStatistiques
	handlersQuotas       *HandlersQuotas
}

// NouveauHandlers crée une nouvelle instance de Handlers
func NouveauHandlers(
	s *store.Store,
	serviceOCR *services.ServiceOCR,
	serviceGeneration *services.ServiceGeneration,
	serviceStatistiques *services.ServiceStatistiques,
	serviceQuotas *services.ServiceQuotas,
	coursRepo store.CoursRepository,
) *Handlers {
	return &Handlers{
		store:                s,
		coursRepo:            coursRepo,
		handlersOCR:          NouveauHandlersOCR(serviceOCR, coursRepo),
		handlersGeneration:   NouveauHandlersGeneration(serviceGeneration),
		handlersStatistiques: NouveauHandlersStatistiques(serviceStatistiques),
		handlersQuotas:       NouveauHandlersQuotas(serviceQuotas),
	}
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
	// Vérifier la connexion à la base de données
	statutDB := "connectee"
	if h.store == nil {
		statutDB = "non configuree"
	} else if err := h.store.Ping(); err != nil {
		statutDB = "erreur"
	}

	c.JSON(http.StatusOK, gin.H{
		"statut":        "ok",
		"baseDeDonnees": statutDB,
		"version":       "0.1.0",
	})
}

// --- Handlers Cours ---

// ListerCoursHandler liste tous les cours
func (h *Handlers) ListerCoursHandler(c *gin.Context) {
	if h.coursRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de cours n'est pas configuré",
			},
		})
		return
	}

	// Pagination
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

	ctx := c.Request.Context()

	// Récupérer les cours
	cours, err := h.coursRepo.Lister(ctx, limite, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_INTERNE",
				"message": "Erreur lors de la récupération des cours",
			},
		})
		return
	}

	// Compter le total
	total, err := h.coursRepo.Compter(ctx)
	if err != nil {
		total = len(cours)
	}

	c.JSON(http.StatusOK, gin.H{
		"succes": true,
		"cours":  cours,
		"total":  total,
		"page":   page,
		"limite": limite,
	})
}

// CreerCoursHandler crée un nouveau cours
func (h *Handlers) CreerCoursHandler(c *gin.Context) {
	if h.coursRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de cours n'est pas configuré",
			},
		})
		return
	}

	var req struct {
		Titre   string `json:"titre" binding:"required"`
		Matiere string `json:"matiere"`
		Texte   string `json:"texte"`
	}

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

	cours := &store.Cours{
		Titre:            req.Titre,
		Matiere:          req.Matiere,
		TexteOCR:         req.Texte,
		Confiance:        1.0,
		ZonesIncertaines: []store.ZoneIncertaine{},
	}

	if err := h.coursRepo.Creer(c.Request.Context(), cours); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_INTERNE",
				"message": "Erreur lors de la création du cours",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"succes": true,
		"cours":  cours,
	})
}

// ObtenirCoursHandler retourne un cours par son ID
func (h *Handlers) ObtenirCoursHandler(c *gin.Context) {
	if h.coursRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de cours n'est pas configuré",
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
				"message": "L'identifiant du cours est requis",
			},
		})
		return
	}

	cours, err := h.coursRepo.ObtenirParID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "cours non trouvé: "+id {
			c.JSON(http.StatusNotFound, gin.H{
				"succes": false,
				"erreur": gin.H{
					"code":    "COURS_NON_TROUVE",
					"message": "Aucun cours trouvé avec cet identifiant",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_INTERNE",
				"message": "Erreur lors de la récupération du cours",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes": true,
		"cours":  cours,
	})
}

// SupprimerCoursHandler supprime un cours par son ID
func (h *Handlers) SupprimerCoursHandler(c *gin.Context) {
	if h.coursRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service de cours n'est pas configuré",
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
				"message": "L'identifiant du cours est requis",
			},
		})
		return
	}

	if err := h.coursRepo.Supprimer(c.Request.Context(), id); err != nil {
		if err.Error() == "cours non trouvé: "+id {
			c.JSON(http.StatusNotFound, gin.H{
				"succes": false,
				"erreur": gin.H{
					"code":    "COURS_NON_TROUVE",
					"message": "Aucun cours trouvé avec cet identifiant",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_INTERNE",
				"message": "Erreur lors de la suppression du cours",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes":  true,
		"message": "Cours supprimé avec succès",
	})
}

// parseInt convertit une chaîne en entier
func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("caractère invalide")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// --- Handlers OCR ---

// OCRHandler traite une image/PDF pour l'OCR
func (h *Handlers) OCRHandler(c *gin.Context) {
	if h.handlersOCR != nil {
		h.handlersOCR.TraiterOCRHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service OCR n'est pas configuré",
		},
	})
}

// --- Handlers Génération ---

// GenererFichesHandler génère des fiches de révision
func (h *Handlers) GenererFichesHandler(c *gin.Context) {
	if h.handlersGeneration != nil {
		h.handlersGeneration.GenererFichesHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de génération n'est pas configuré",
		},
	})
}

// ObtenirFichesHandler récupère les fiches d'un cours
func (h *Handlers) ObtenirFichesHandler(c *gin.Context) {
	if h.handlersGeneration != nil {
		h.handlersGeneration.ObtenirFichesHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service n'est pas configuré",
		},
	})
}

// GenererQuizHandler génère un quiz
func (h *Handlers) GenererQuizHandler(c *gin.Context) {
	if h.handlersGeneration != nil {
		h.handlersGeneration.GenererQuizHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de génération n'est pas configuré",
		},
	})
}

// ObtenirQuizHandler récupère un quiz par son ID
func (h *Handlers) ObtenirQuizHandler(c *gin.Context) {
	if h.handlersGeneration != nil {
		h.handlersGeneration.ObtenirQuizHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service n'est pas configuré",
		},
	})
}

// DemarrerSessionHandler démarre une session de quiz
func (h *Handlers) DemarrerSessionHandler(c *gin.Context) {
	if h.handlersGeneration != nil {
		h.handlersGeneration.DemarrerSessionHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service n'est pas configuré",
		},
	})
}

// RepondreHandler enregistre une réponse à une question
func (h *Handlers) RepondreHandler(c *gin.Context) {
	if h.handlersGeneration != nil {
		h.handlersGeneration.RepondreHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service n'est pas configuré",
		},
	})
}

// TerminerSessionHandler termine une session de quiz
func (h *Handlers) TerminerSessionHandler(c *gin.Context) {
	if h.handlersGeneration != nil {
		h.handlersGeneration.TerminerSessionHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service n'est pas configuré",
		},
	})
}

// GenererMindmapHandler génère une mindmap
func (h *Handlers) GenererMindmapHandler(c *gin.Context) {
	if h.handlersGeneration != nil {
		h.handlersGeneration.GenererMindmapHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de génération n'est pas configuré",
		},
	})
}

// ObtenirMindmapHandler récupère la mindmap d'un cours
func (h *Handlers) ObtenirMindmapHandler(c *gin.Context) {
	if h.handlersGeneration != nil {
		h.handlersGeneration.ObtenirMindmapHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service n'est pas configuré",
		},
	})
}

// --- Handlers Ressources ---

// GenererRessourcesHandler génère des ressources complémentaires
func (h *Handlers) GenererRessourcesHandler(c *gin.Context) {
	if h.handlersGeneration != nil {
		h.handlersGeneration.GenererRessourcesHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de génération n'est pas configuré",
		},
	})
}

// ObtenirRessourcesHandler récupère les ressources d'un cours
func (h *Handlers) ObtenirRessourcesHandler(c *gin.Context) {
	if h.handlersGeneration != nil {
		h.handlersGeneration.ObtenirRessourcesHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service n'est pas configuré",
		},
	})
}

// --- Handlers Statistiques ---

// ObtenirStatistiquesHandler retourne les statistiques globales
func (h *Handlers) ObtenirStatistiquesHandler(c *gin.Context) {
	if h.handlersStatistiques != nil {
		h.handlersStatistiques.ObtenirStatistiquesHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de statistiques n'est pas configuré",
		},
	})
}

// ObtenirCoursRecentsHandler retourne les cours récents avec leurs statistiques
func (h *Handlers) ObtenirCoursRecentsHandler(c *gin.Context) {
	if h.handlersStatistiques != nil {
		h.handlersStatistiques.ObtenirCoursRecentsHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de statistiques n'est pas configuré",
		},
	})
}

// --- Handlers Quotas ---

// ObtenirQuotasHandler retourne l'état actuel des quotas
func (h *Handlers) ObtenirQuotasHandler(c *gin.Context) {
	if h.handlersQuotas != nil {
		h.handlersQuotas.ObtenirQuotasHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de quotas n'est pas configuré",
		},
	})
}
