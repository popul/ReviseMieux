// Package api contient les handlers HTTP
package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/config"
	"github.com/revisemieux/backend/internal/services"
	"github.com/revisemieux/backend/internal/store"
)

// ConfigFrontend contient la configuration exposée au frontend
type ConfigFrontend struct {
	NombreMaxPages int `json:"nombreMaxPages"`
}

// Handlers contient les dépendances des handlers
type Handlers struct {
	store                   *store.Store
	coursRepo               store.CoursRepository
	configFrontend          *ConfigFrontend
	handlersOCR             *HandlersOCR
	handlersGeneration      *HandlersGeneration
	handlersStatistiques    *HandlersStatistiques
	handlersQuotas          *HandlersQuotas
	handlersCopies          *HandlersCopies
	handlersRecommandations *HandlersRecommandations
	handlersImages          *HandlersImages
	handlersConcepts        *HandlersConcepts
	handlersExamen          *HandlersExamen
	handlersLexique         *HandlersLexique
	handlersPlans           *HandlersPlans
}

// NouveauHandlers crée une nouvelle instance de Handlers
func NouveauHandlers(
	cfg *config.Config,
	s *store.Store,
	serviceOCR *services.ServiceOCR,
	serviceGeneration *services.ServiceGeneration,
	serviceStatistiques *services.ServiceStatistiques,
	serviceQuotas *services.ServiceQuotas,
	serviceAnalyseErreurs *services.ServiceAnalyseErreurs,
	serviceRecommandations *services.ServiceRecommandations,
	serviceStorage *services.ServiceStorage,
	coursRepo store.CoursRepository,
	copieRepo store.CopieExamenRepository,
	erreurRepo store.ErreurAnalyseRepository,
	serviceConcepts *services.ServiceConcepts,
	serviceExamen *services.ServiceExamen,
	serviceLexique *services.ServiceLexique,
	servicePlans *services.ServicePlans,
) *Handlers {
	return &Handlers{
		store:                   s,
		coursRepo:               coursRepo,
		configFrontend:          &ConfigFrontend{NombreMaxPages: cfg.NombreMaxPages},
		handlersOCR:             NouveauHandlersOCR(serviceOCR, serviceStorage, coursRepo, serviceGeneration, serviceConcepts),
		handlersGeneration:      NouveauHandlersGeneration(serviceGeneration),
		handlersStatistiques:    NouveauHandlersStatistiques(serviceStatistiques),
		handlersQuotas:          NouveauHandlersQuotas(serviceQuotas),
		handlersCopies:          NouveauHandlersCopies(serviceOCR, serviceAnalyseErreurs, copieRepo, erreurRepo),
		handlersRecommandations: NouveauHandlersRecommandations(serviceRecommandations),
		handlersImages:          NouveauHandlersImages(serviceStorage, coursRepo),
		handlersConcepts:        NouveauHandlersConcepts(serviceConcepts),
		handlersExamen:          NouveauHandlersExamen(serviceExamen),
		handlersLexique:         NouveauHandlersLexique(serviceLexique),
		handlersPlans:           NouveauHandlersPlans(servicePlans),
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

// ConfigFrontendHandler retourne la configuration pour le frontend
func (h *Handlers) ConfigFrontendHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"succes": true,
		"config": h.configFrontend,
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

// MettreAJourCoursHandler met à jour un cours
func (h *Handlers) MettreAJourCoursHandler(c *gin.Context) {
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

	// Récupérer le cours existant
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

	// Parser les données de mise à jour
	var req struct {
		Titre            string                 `json:"titre"`
		Matiere          string                 `json:"matiere"`
		TexteOCR         string                 `json:"texteOCR"`
		ZonesIncertaines []store.ZoneIncertaine `json:"zonesIncertaines"`
		Images           []string               `json:"images"`
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

	// Mettre à jour les champs si fournis
	if req.Titre != "" {
		cours.Titre = req.Titre
	}
	if req.Matiere != "" {
		cours.Matiere = req.Matiere
	}
	if req.TexteOCR != "" {
		cours.TexteOCR = req.TexteOCR
	}
	if req.ZonesIncertaines != nil {
		cours.ZonesIncertaines = req.ZonesIncertaines
	}
	if req.Images != nil {
		cours.Images = req.Images
	}

	// Sauvegarder les modifications
	if err := h.coursRepo.MettreAJour(c.Request.Context(), cours); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_INTERNE",
				"message": "Erreur lors de la mise à jour du cours",
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
		h.handlersOCR.TraiterOCRStreamHandler(c)
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

// GenererResumeHandler génère un résumé pour un cours
func (h *Handlers) GenererResumeHandler(c *gin.Context) {
	if h.handlersGeneration != nil {
		h.handlersGeneration.GenererResumeHandler(c)
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

// RetraiterOCRCoursHandler relance l'OCR sur un cours existant
func (h *Handlers) RetraiterOCRCoursHandler(c *gin.Context) {
	if h.handlersOCR != nil {
		h.handlersOCR.RetraiterOCRCoursHandler(c)
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

// ObtenirProgressionHandler retourne l'historique des quiz et stats par matière
func (h *Handlers) ObtenirProgressionHandler(c *gin.Context) {
	if h.handlersStatistiques != nil {
		h.handlersStatistiques.ObtenirProgressionHandler(c)
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

// --- Handlers Copies d'Examens ---

// TraiterOCRCopieHandler traite l'OCR d'une copie d'examen
func (h *Handlers) TraiterOCRCopieHandler(c *gin.Context) {
	if h.handlersCopies != nil {
		h.handlersCopies.TraiterOCRCopieHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de copies n'est pas configuré",
		},
	})
}

// ListerCopiesHandler liste les copies d'examens
func (h *Handlers) ListerCopiesHandler(c *gin.Context) {
	if h.handlersCopies != nil {
		h.handlersCopies.ListerCopiesHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de copies n'est pas configuré",
		},
	})
}

// ObtenirCopieHandler récupère une copie d'examen par son ID
func (h *Handlers) ObtenirCopieHandler(c *gin.Context) {
	if h.handlersCopies != nil {
		h.handlersCopies.ObtenirCopieHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de copies n'est pas configuré",
		},
	})
}

// SupprimerCopieHandler supprime une copie d'examen
func (h *Handlers) SupprimerCopieHandler(c *gin.Context) {
	if h.handlersCopies != nil {
		h.handlersCopies.SupprimerCopieHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de copies n'est pas configuré",
		},
	})
}

// AnalyserCopieHandler lance l'analyse des erreurs d'une copie
func (h *Handlers) AnalyserCopieHandler(c *gin.Context) {
	if h.handlersCopies != nil {
		h.handlersCopies.AnalyserCopieHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'analyse n'est pas configuré",
		},
	})
}

// ObtenirErreursHandler récupère les erreurs d'analyse d'une copie
func (h *Handlers) ObtenirErreursHandler(c *gin.Context) {
	if h.handlersCopies != nil {
		h.handlersCopies.ObtenirErreursHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'erreurs n'est pas configuré",
		},
	})
}

// --- Handlers Recommandations ---

// GenererRecommandationsCopieHandler génère des recommandations pour une copie
func (h *Handlers) GenererRecommandationsCopieHandler(c *gin.Context) {
	if h.handlersRecommandations != nil {
		h.handlersRecommandations.GenererRecommandationsCopieHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de recommandations n'est pas configuré",
		},
	})
}

// ObtenirRecommandationsPrioritairesHandler retourne les recommandations prioritaires
func (h *Handlers) ObtenirRecommandationsPrioritairesHandler(c *gin.Context) {
	if h.handlersRecommandations != nil {
		h.handlersRecommandations.ObtenirRecommandationsPrioritairesHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de recommandations n'est pas configuré",
		},
	})
}

// ObtenirRecommandationsParMatiereHandler retourne les recommandations pour une matière
func (h *Handlers) ObtenirRecommandationsParMatiereHandler(c *gin.Context) {
	if h.handlersRecommandations != nil {
		h.handlersRecommandations.ObtenirRecommandationsParMatiereHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de recommandations n'est pas configuré",
		},
	})
}

// --- Handlers Images ---

// ListerImagesHandler retourne la liste des images d'un cours
func (h *Handlers) ListerImagesHandler(c *gin.Context) {
	if h.handlersImages != nil {
		h.handlersImages.ListerImagesHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'images n'est pas configuré",
		},
	})
}

// ServirImageHandler sert une image d'un cours
func (h *Handlers) ServirImageHandler(c *gin.Context) {
	if h.handlersImages != nil {
		h.handlersImages.ServirImageHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'images n'est pas configuré",
		},
	})
}

// AjouterImageHandler ajoute une image à un cours
func (h *Handlers) AjouterImageHandler(c *gin.Context) {
	if h.handlersImages != nil {
		h.handlersImages.AjouterImageHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'images n'est pas configuré",
		},
	})
}

// SupprimerImageHandler supprime une image d'un cours
func (h *Handlers) SupprimerImageHandler(c *gin.Context) {
	if h.handlersImages != nil {
		h.handlersImages.SupprimerImageHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'images n'est pas configuré",
		},
	})
}

// ReordonnerImagesHandler réordonne les images d'un cours
func (h *Handlers) ReordonnerImagesHandler(c *gin.Context) {
	if h.handlersImages != nil {
		h.handlersImages.ReordonnerImagesHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'images n'est pas configuré",
		},
	})
}

// DeplacerImageHandler déplace une image vers le haut ou vers le bas
func (h *Handlers) DeplacerImageHandler(c *gin.Context) {
	if h.handlersImages != nil {
		h.handlersImages.DeplacerImageHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'images n'est pas configuré",
		},
	})
}

// --- Handlers Concepts ---

// ExtraireConceptsHandler extrait les concepts d'un cours
func (h *Handlers) ExtraireConceptsHandler(c *gin.Context) {
	if h.handlersConcepts != nil {
		h.handlersConcepts.ExtraireConceptsHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de concepts n'est pas configuré",
		},
	})
}

// ListerConceptsHandler récupère les concepts d'un cours
func (h *Handlers) ListerConceptsHandler(c *gin.Context) {
	if h.handlersConcepts != nil {
		h.handlersConcepts.ListerConceptsHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de concepts n'est pas configuré",
		},
	})
}

// MettreAJourConceptHandler met à jour un concept
func (h *Handlers) MettreAJourConceptHandler(c *gin.Context) {
	if h.handlersConcepts != nil {
		h.handlersConcepts.MettreAJourConceptHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de concepts n'est pas configuré",
		},
	})
}

// SupprimerConceptHandler supprime un concept
func (h *Handlers) SupprimerConceptHandler(c *gin.Context) {
	if h.handlersConcepts != nil {
		h.handlersConcepts.SupprimerConceptHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de concepts n'est pas configuré",
		},
	})
}

// --- Handlers Lexique ---

// ExtraireTermesLexiqueHandler extrait les termes d un cours
func (h *Handlers) ExtraireTermesLexiqueHandler(c *gin.Context) {
	if h.handlersLexique != nil {
		h.handlersLexique.ExtraireTermesHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de lexique n est pas configure",
		},
	})
}

// ListerTermesLexiqueHandler recupere les termes d un cours
func (h *Handlers) ListerTermesLexiqueHandler(c *gin.Context) {
	if h.handlersLexique != nil {
		h.handlersLexique.ListerTermesHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de lexique n est pas configure",
		},
	})
}

// MettreAJourMaitriseHandler met a jour la maitrise d un terme
func (h *Handlers) MettreAJourMaitriseHandler(c *gin.Context) {
	if h.handlersLexique != nil {
		h.handlersLexique.MettreAJourMaitriseHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de lexique n est pas configure",
		},
	})
}

// GenererQuizVocabulaireHandler genere un quiz de vocabulaire
func (h *Handlers) GenererQuizVocabulaireHandler(c *gin.Context) {
	if h.handlersLexique != nil {
		h.handlersLexique.GenererQuizVocabulaireHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de lexique n est pas configure",
		},
	})
}

// --- Handlers Examen Blanc ---

// GenererExamenHandler genere un examen blanc pour un cours
func (h *Handlers) GenererExamenHandler(c *gin.Context) {
	if h.handlersExamen != nil {
		h.handlersExamen.GenererExamenHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'examen n'est pas configure",
		},
	})
}

// ObtenirExamenHandler recupere un examen par son ID
func (h *Handlers) ObtenirExamenHandler(c *gin.Context) {
	if h.handlersExamen != nil {
		h.handlersExamen.ObtenirExamenHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'examen n'est pas configure",
		},
	})
}

// DemarrerSessionExamenHandler demarre une session d'examen
func (h *Handlers) DemarrerSessionExamenHandler(c *gin.Context) {
	if h.handlersExamen != nil {
		h.handlersExamen.DemarrerSessionExamenHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'examen n'est pas configure",
		},
	})
}

// ObtenirSessionExamenHandler recupere une session d'examen
func (h *Handlers) ObtenirSessionExamenHandler(c *gin.Context) {
	if h.handlersExamen != nil {
		h.handlersExamen.ObtenirSessionExamenHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'examen n'est pas configure",
		},
	})
}

// DemanderIndiceHandler retourne un indice pour une question
func (h *Handlers) DemanderIndiceHandler(c *gin.Context) {
	if h.handlersExamen != nil {
		h.handlersExamen.DemanderIndiceHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'examen n'est pas configure",
		},
	})
}

// CorrigerExamenHandler corrige un examen
func (h *Handlers) CorrigerExamenHandler(c *gin.Context) {
	if h.handlersExamen != nil {
		h.handlersExamen.CorrigerExamenHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service d'examen n'est pas configure",
		},
	})
}

// --- Handlers Plans de Révision ---

// ListerPlansHandler liste les plans de révision
func (h *Handlers) ListerPlansHandler(c *gin.Context) {
	if h.handlersPlans != nil {
		h.handlersPlans.ListerPlansHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de plans n'est pas configuré",
		},
	})
}

// CreerPlanHandler crée un nouveau plan de révision
func (h *Handlers) CreerPlanHandler(c *gin.Context) {
	if h.handlersPlans != nil {
		h.handlersPlans.CreerPlanHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de plans n'est pas configuré",
		},
	})
}

// ObtenirPlanHandler récupère un plan par son ID
func (h *Handlers) ObtenirPlanHandler(c *gin.Context) {
	if h.handlersPlans != nil {
		h.handlersPlans.ObtenirPlanHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de plans n'est pas configuré",
		},
	})
}

// ObtenirPlanCompletHandler récupère un plan avec ses cours et artefacts
func (h *Handlers) ObtenirPlanCompletHandler(c *gin.Context) {
	if h.handlersPlans != nil {
		h.handlersPlans.ObtenirPlanCompletHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de plans n'est pas configuré",
		},
	})
}

// MettreAJourPlanHandler met à jour un plan
func (h *Handlers) MettreAJourPlanHandler(c *gin.Context) {
	if h.handlersPlans != nil {
		h.handlersPlans.MettreAJourPlanHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de plans n'est pas configuré",
		},
	})
}

// SupprimerPlanHandler supprime un plan
func (h *Handlers) SupprimerPlanHandler(c *gin.Context) {
	if h.handlersPlans != nil {
		h.handlersPlans.SupprimerPlanHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de plans n'est pas configuré",
		},
	})
}

// AjouterCoursAuPlanHandler ajoute un cours à un plan
func (h *Handlers) AjouterCoursAuPlanHandler(c *gin.Context) {
	if h.handlersPlans != nil {
		h.handlersPlans.AjouterCoursHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de plans n'est pas configuré",
		},
	})
}

// RetirerCoursDuPlanHandler retire un cours d'un plan
func (h *Handlers) RetirerCoursDuPlanHandler(c *gin.Context) {
	if h.handlersPlans != nil {
		h.handlersPlans.RetirerCoursHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de plans n'est pas configuré",
		},
	})
}

// ListerPlansParCoursHandler liste les plans contenant un cours
func (h *Handlers) ListerPlansParCoursHandler(c *gin.Context) {
	if h.handlersPlans != nil {
		h.handlersPlans.ListerPlansParCoursHandler(c)
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{
		"succes": false,
		"erreur": gin.H{
			"code":    "SERVICE_NON_DISPONIBLE",
			"message": "Le service de plans n'est pas configuré",
		},
	})
}
