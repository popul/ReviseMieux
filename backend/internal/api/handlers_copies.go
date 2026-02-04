// Package api contient les handlers pour les copies d'examens et l'analyse d'erreurs
package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/services"
	"github.com/revisemieux/backend/internal/store"
)

// HandlersCopies contient les handlers pour les copies d'examens
type HandlersCopies struct {
	serviceOCR            *services.ServiceOCR
	serviceAnalyseErreurs *services.ServiceAnalyseErreurs
	copieRepo             store.CopieExamenRepository
	erreurRepo            store.ErreurAnalyseRepository
}

// NouveauHandlersCopies crée une nouvelle instance des handlers pour les copies
func NouveauHandlersCopies(
	serviceOCR *services.ServiceOCR,
	serviceAnalyseErreurs *services.ServiceAnalyseErreurs,
	copieRepo store.CopieExamenRepository,
	erreurRepo store.ErreurAnalyseRepository,
) *HandlersCopies {
	return &HandlersCopies{
		serviceOCR:            serviceOCR,
		serviceAnalyseErreurs: serviceAnalyseErreurs,
		copieRepo:             copieRepo,
		erreurRepo:            erreurRepo,
	}
}

// ReponseCopie représente la réponse JSON pour une copie d'examen
type ReponseCopie struct {
	Succes bool                `json:"succes"`
	Copie  *store.CopieExamen  `json:"copie,omitempty"`
	Erreur *ErreurReponse      `json:"erreur,omitempty"`
}

// ReponseCopies représente la réponse JSON pour une liste de copies
type ReponseCopies struct {
	Succes bool                 `json:"succes"`
	Copies []*store.CopieExamen `json:"copies,omitempty"`
	Total  int                  `json:"total,omitempty"`
	Page   int                  `json:"page,omitempty"`
	Limite int                  `json:"limite,omitempty"`
	Erreur *ErreurReponse       `json:"erreur,omitempty"`
}

// ReponseAnalyse représente la réponse JSON pour une analyse d'erreurs
type ReponseAnalyse struct {
	Succes           bool                           `json:"succes"`
	Resultat         *services.ResultatAnalyseErreurs `json:"resultat,omitempty"`
	Erreur           *ErreurReponse                 `json:"erreur,omitempty"`
}

// RequeteOCRCopie représente les paramètres de la requête OCR pour une copie
type RequeteOCRCopie struct {
	Titre                 string   `form:"titre"`
	Matiere               string   `form:"matiere"`
	CoursID               string   `form:"coursId"`
	NoteObtenue           *float64 `form:"noteObtenue"`
	NoteTotale            *float64 `form:"noteTotale"`
	AnnotationsProfesseur string   `form:"annotationsProfesseur"`
}

// RequeteAnalyse représente les paramètres de la requête d'analyse
type RequeteAnalyse struct {
	InclusAnnotations bool   `json:"inclusAnnotations"`
	CoursID           string `json:"coursId,omitempty"`
}

// TraiterOCRCopieHandler traite l'OCR d'une copie d'examen
// POST /api/copies/ocr
// Content-Type: multipart/form-data
// Champs:
//   - fichiers[]: fichiers à traiter (required)
//   - titre: titre de la copie (required)
//   - matiere: matière de l'examen (optional)
//   - coursId: ID du cours associé (optional)
//   - noteObtenue: note obtenue (optional)
//   - noteTotale: note totale (optional)
//   - annotationsProfesseur: annotations du professeur (optional)
func (h *HandlersCopies) TraiterOCRCopieHandler(c *gin.Context) {
	if h.serviceOCR == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseCopie{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service OCR n'est pas configuré",
			},
		})
		return
	}

	// Parser les paramètres de la requête
	var requete RequeteOCRCopie
	if err := c.ShouldBind(&requete); err != nil {
		// Ignorer l'erreur de binding, certains champs sont optionnels
	}

	// Récupérer les fichiers uploadés
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, ReponseCopie{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "REQUETE_INVALIDE",
				Message: "La requête doit être de type multipart/form-data",
			},
		})
		return
	}

	fichiers := form.File["fichiers[]"]
	if len(fichiers) == 0 {
		fichiers = form.File["fichiers"]
	}

	if len(fichiers) == 0 {
		c.JSON(http.StatusBadRequest, ReponseCopie{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "FICHIERS_MANQUANTS",
				Message: "Aucun fichier fourni. Utilisez le champ 'fichiers[]' ou 'fichiers'",
			},
		})
		return
	}

	// Valider les fichiers
	if err := h.serviceOCR.ValiderFichiers(fichiers); err != nil {
		codeErreur, message := h.convertirErreurOCR(err)
		c.JSON(http.StatusBadRequest, ReponseCopie{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    codeErreur,
				Message: message,
			},
		})
		return
	}

	// Traiter les fichiers avec le service OCR
	resultat, err := h.serviceOCR.TraiterFichiers(c.Request.Context(), fichiers)
	if err != nil {
		statusCode, codeErreur, message := h.convertirErreurTraitement(err)
		c.JSON(statusCode, ReponseCopie{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    codeErreur,
				Message: message,
			},
		})
		return
	}

	// Convertir les zones incertaines
	zonesIncertaines := make([]store.ZoneIncertaine, len(resultat.ZonesIncertaines))
	for i, zone := range resultat.ZonesIncertaines {
		zonesIncertaines[i] = store.ZoneIncertaine{
			Debut:  zone.Debut,
			Fin:    zone.Fin,
			Texte:  zone.Texte,
			Raison: zone.Raison,
		}
	}

	// Collecter les noms de fichiers
	nomsFichiers := make([]string, len(fichiers))
	for i, f := range fichiers {
		nomsFichiers[i] = f.Filename
	}

	// Créer la copie d'examen
	titre := requete.Titre
	if titre == "" {
		titre = "Copie sans titre"
	}

	copie := &store.CopieExamen{
		Titre:                 titre,
		Matiere:               requete.Matiere,
		CoursID:               requete.CoursID,
		NoteObtenue:           requete.NoteObtenue,
		NoteTotale:            requete.NoteTotale,
		TexteOCR:              resultat.Texte,
		AnnotationsProfesseur: requete.AnnotationsProfesseur,
		Confiance:             resultat.Confiance,
		ZonesIncertaines:      zonesIncertaines,
		FichiersOriginaux:     nomsFichiers,
	}

	// Sauvegarder la copie
	if h.copieRepo != nil {
		if err := h.copieRepo.Creer(c.Request.Context(), copie); err != nil {
			c.JSON(http.StatusInternalServerError, ReponseCopie{
				Succes: false,
				Erreur: &ErreurReponse{
					Code:    "ERREUR_SAUVEGARDE",
					Message: "Erreur lors de la sauvegarde de la copie",
				},
			})
			return
		}
	}

	c.JSON(http.StatusCreated, ReponseCopie{
		Succes: true,
		Copie:  copie,
	})
}

// ListerCopiesHandler liste toutes les copies d'examens
// GET /api/copies
// Query params: page (default 1), limite (default 20, max 100)
func (h *HandlersCopies) ListerCopiesHandler(c *gin.Context) {
	if h.copieRepo == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseCopies{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de copies n'est pas configuré",
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

	// Récupérer les copies
	copies, err := h.copieRepo.Lister(ctx, limite, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ReponseCopies{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ERREUR_INTERNE",
				Message: "Erreur lors de la récupération des copies",
			},
		})
		return
	}

	// Compter le total
	total, err := h.copieRepo.Compter(ctx)
	if err != nil {
		total = len(copies)
	}

	c.JSON(http.StatusOK, ReponseCopies{
		Succes: true,
		Copies: copies,
		Total:  total,
		Page:   page,
		Limite: limite,
	})
}

// ObtenirCopieHandler récupère une copie d'examen par son ID
// GET /api/copies/:id
func (h *HandlersCopies) ObtenirCopieHandler(c *gin.Context) {
	if h.copieRepo == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseCopie{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de copies n'est pas configuré",
			},
		})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ReponseCopie{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'identifiant de la copie est requis",
			},
		})
		return
	}

	copie, err := h.copieRepo.ObtenirParID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "copie examen non trouvée: "+id {
			c.JSON(http.StatusNotFound, ReponseCopie{
				Succes: false,
				Erreur: &ErreurReponse{
					Code:    "COPIE_NON_TROUVEE",
					Message: "Aucune copie trouvée avec cet identifiant",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ReponseCopie{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ERREUR_INTERNE",
				Message: "Erreur lors de la récupération de la copie",
			},
		})
		return
	}

	c.JSON(http.StatusOK, ReponseCopie{
		Succes: true,
		Copie:  copie,
	})
}

// SupprimerCopieHandler supprime une copie d'examen par son ID
// DELETE /api/copies/:id
func (h *HandlersCopies) SupprimerCopieHandler(c *gin.Context) {
	if h.copieRepo == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseCopie{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service de copies n'est pas configuré",
			},
		})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ReponseCopie{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'identifiant de la copie est requis",
			},
		})
		return
	}

	if err := h.copieRepo.Supprimer(c.Request.Context(), id); err != nil {
		if err.Error() == "copie examen non trouvée: "+id {
			c.JSON(http.StatusNotFound, ReponseCopie{
				Succes: false,
				Erreur: &ErreurReponse{
					Code:    "COPIE_NON_TROUVEE",
					Message: "Aucune copie trouvée avec cet identifiant",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ReponseCopie{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ERREUR_INTERNE",
				Message: "Erreur lors de la suppression de la copie",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"succes":  true,
		"message": "Copie supprimée avec succès",
	})
}

// AnalyserCopieHandler lance l'analyse des erreurs d'une copie
// POST /api/copies/:id/analyser
// Body: { "inclusAnnotations": true, "coursId": "..." }
func (h *HandlersCopies) AnalyserCopieHandler(c *gin.Context) {
	if h.serviceAnalyseErreurs == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseAnalyse{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service d'analyse n'est pas configuré",
			},
		})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ReponseAnalyse{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ID_MANQUANT",
				Message: "L'identifiant de la copie est requis",
			},
		})
		return
	}

	// Parser les options d'analyse
	var requete RequeteAnalyse
	if err := c.ShouldBindJSON(&requete); err != nil {
		// Utiliser les valeurs par défaut si pas de body JSON
		requete = RequeteAnalyse{
			InclusAnnotations: true,
		}
	}

	options := &services.OptionsAnalyseErreurs{
		InclusAnnotations: requete.InclusAnnotations,
		CoursID:           requete.CoursID,
	}

	resultat, err := h.serviceAnalyseErreurs.AnalyserCopie(c.Request.Context(), id, options)
	if err != nil {
		statusCode, codeErreur, message := h.convertirErreurAnalyse(err)
		c.JSON(statusCode, ReponseAnalyse{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    codeErreur,
				Message: message,
			},
		})
		return
	}

	c.JSON(http.StatusOK, ReponseAnalyse{
		Succes:   true,
		Resultat: resultat,
	})
}

// ObtenirErreursHandler récupère les erreurs d'analyse d'une copie
// GET /api/copies/:id/erreurs
func (h *HandlersCopies) ObtenirErreursHandler(c *gin.Context) {
	if h.erreurRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "SERVICE_NON_DISPONIBLE",
				"message": "Le service d'erreurs n'est pas configuré",
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
				"message": "L'identifiant de la copie est requis",
			},
		})
		return
	}

	erreurs, err := h.erreurRepo.ListerParCopie(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"succes": false,
			"erreur": gin.H{
				"code":    "ERREUR_INTERNE",
				"message": "Erreur lors de la récupération des erreurs",
			},
		})
		return
	}

	// Compter par type
	comptesParType, _ := h.erreurRepo.CompterParType(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{
		"succes":          true,
		"erreurs":         erreurs,
		"nombreErreurs":   len(erreurs),
		"comptesParType":  comptesParType,
	})
}

// convertirErreurOCR convertit une erreur OCR en code et message
func (h *HandlersCopies) convertirErreurOCR(err error) (string, string) {
	var errOCR *services.ErreurOCR
	if errors.As(err, &errOCR) {
		return errOCR.Code, errOCR.Message
	}
	return "ERREUR_VALIDATION", err.Error()
}

// convertirErreurTraitement convertit une erreur de traitement OCR
func (h *HandlersCopies) convertirErreurTraitement(err error) (int, string, string) {
	var errOCR *services.ErreurOCR
	if errors.As(err, &errOCR) {
		switch errOCR.Code {
		case "LLM_NON_DISPONIBLE":
			return http.StatusServiceUnavailable, errOCR.Code, errOCR.Message
		default:
			return http.StatusBadRequest, errOCR.Code, errOCR.Message
		}
	}

	var errLLM *llm.ErreurLLM
	if errors.As(err, &errLLM) {
		if errLLM.RateLimited {
			return http.StatusTooManyRequests, "RATE_LIMIT", "Trop de requêtes, veuillez réessayer plus tard"
		}
		if !errLLM.Recuperable {
			return http.StatusServiceUnavailable, "SERVICE_LLM_ERREUR", errLLM.Message
		}
		return http.StatusInternalServerError, "ERREUR_OCR", errLLM.Message
	}

	return http.StatusInternalServerError, "ERREUR_INTERNE", "Une erreur inattendue s'est produite"
}

// convertirErreurAnalyse convertit une erreur d'analyse en code HTTP et message
func (h *HandlersCopies) convertirErreurAnalyse(err error) (int, string, string) {
	var errGen *services.ErreurGeneration
	if errors.As(err, &errGen) {
		switch errGen.Code {
		case "COPIE_NON_TROUVEE":
			return http.StatusNotFound, errGen.Code, errGen.Message
		case "COPIE_VIDE":
			return http.StatusBadRequest, errGen.Code, errGen.Message
		case "SERVICE_NON_DISPONIBLE":
			return http.StatusServiceUnavailable, errGen.Code, errGen.Message
		default:
			return http.StatusInternalServerError, errGen.Code, errGen.Message
		}
	}

	var errLLM *llm.ErreurLLM
	if errors.As(err, &errLLM) {
		if errLLM.RateLimited {
			return http.StatusTooManyRequests, "RATE_LIMIT", "Trop de requêtes, veuillez réessayer plus tard"
		}
		if !errLLM.Recuperable {
			return http.StatusServiceUnavailable, "SERVICE_LLM_ERREUR", errLLM.Message
		}
		return http.StatusInternalServerError, "ERREUR_ANALYSE", errLLM.Message
	}

	return http.StatusInternalServerError, "ERREUR_INTERNE", "Une erreur inattendue s'est produite"
}

// Vérification que HandlersCopies implémente les méthodes nécessaires
var _ interface {
	TraiterOCRCopieHandler(*gin.Context)
	ListerCopiesHandler(*gin.Context)
	ObtenirCopieHandler(*gin.Context)
	SupprimerCopieHandler(*gin.Context)
	AnalyserCopieHandler(*gin.Context)
	ObtenirErreursHandler(*gin.Context)
} = (*HandlersCopies)(nil)
