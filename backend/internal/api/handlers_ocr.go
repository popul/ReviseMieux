// Package api contient les handlers OCR pour le traitement d'images et PDF
package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/services"
	"github.com/revisemieux/backend/internal/store"
)

// HandlersOCR contient les handlers OCR avec leurs dépendances
type HandlersOCR struct {
	serviceOCR *services.ServiceOCR
	coursRepo  store.CoursRepository
}

// NouveauHandlersOCR crée une nouvelle instance des handlers OCR
func NouveauHandlersOCR(serviceOCR *services.ServiceOCR, coursRepo store.CoursRepository) *HandlersOCR {
	return &HandlersOCR{
		serviceOCR: serviceOCR,
		coursRepo:  coursRepo,
	}
}

// ReponseOCR représente la réponse JSON de l'endpoint OCR
type ReponseOCR struct {
	Succes           bool                   `json:"succes"`
	Texte            string                 `json:"texte,omitempty"`
	Confiance        float64                `json:"confiance,omitempty"`
	ZonesIncertaines []store.ZoneIncertaine `json:"zonesIncertaines,omitempty"`
	NombrePages      int                    `json:"nombrePages,omitempty"`
	CoursID          string                 `json:"coursId,omitempty"`
	Erreur           *ErreurReponse         `json:"erreur,omitempty"`
}

// ErreurReponse représente une erreur dans la réponse JSON
type ErreurReponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RequeteOCR représente les paramètres optionnels de la requête OCR
type RequeteOCR struct {
	Titre        string `form:"titre"`
	Matiere      string `form:"matiere"`
	Sauvegarder  bool   `form:"sauvegarder"`
}

// TraiterOCRHandler traite une image ou PDF pour l'OCR
// POST /api/ocr
// Content-Type: multipart/form-data
// Champs:
//   - fichiers[]: fichiers à traiter (required, max 10)
//   - titre: titre du cours (optional)
//   - matiere: matière du cours (optional)
//   - sauvegarder: si true, sauvegarde le cours en base (optional, default false)
func (h *HandlersOCR) TraiterOCRHandler(c *gin.Context) {
	// Vérifier que le service OCR est disponible
	if h.serviceOCR == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseOCR{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service OCR n'est pas configuré",
			},
		})
		return
	}

	// Parser les paramètres de la requête
	var requete RequeteOCR
	if err := c.ShouldBind(&requete); err != nil {
		// Ignorer l'erreur de binding, les champs sont optionnels
	}

	// Récupérer les fichiers uploadés
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, ReponseOCR{
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
		// Essayer aussi avec "fichiers" sans les crochets
		fichiers = form.File["fichiers"]
	}

	if len(fichiers) == 0 {
		c.JSON(http.StatusBadRequest, ReponseOCR{
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
		c.JSON(http.StatusBadRequest, ReponseOCR{
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
		c.JSON(statusCode, ReponseOCR{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    codeErreur,
				Message: message,
			},
		})
		return
	}

	// Convertir les zones incertaines du format llm vers store
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

	// Construire la réponse
	reponse := ReponseOCR{
		Succes:           true,
		Texte:            resultat.Texte,
		Confiance:        resultat.Confiance,
		ZonesIncertaines: zonesIncertaines,
		NombrePages:      resultat.NombrePages,
	}

	// Sauvegarder le cours si demandé
	if requete.Sauvegarder && h.coursRepo != nil {
		titre := requete.Titre
		if titre == "" {
			titre = "Cours sans titre"
		}

		cours := &store.Cours{
			Titre:             titre,
			Matiere:           requete.Matiere,
			TexteOCR:          resultat.Texte,
			Confiance:         resultat.Confiance,
			ZonesIncertaines:  zonesIncertaines,
			FichiersOriginaux: nomsFichiers,
		}

		if err := h.coursRepo.Creer(c.Request.Context(), cours); err != nil {
			// Log l'erreur mais ne pas faire échouer la requête
			// Le texte OCR a été extrait avec succès
			c.JSON(http.StatusOK, reponse)
			return
		}

		reponse.CoursID = cours.ID
	}

	c.JSON(http.StatusOK, reponse)
}

// convertirErreurOCR convertit une erreur de validation OCR en code et message
func (h *HandlersOCR) convertirErreurOCR(err error) (string, string) {
	var errOCR *services.ErreurOCR
	if errors.As(err, &errOCR) {
		return errOCR.Code, errOCR.Message
	}
	return "ERREUR_VALIDATION", err.Error()
}

// convertirErreurTraitement convertit une erreur de traitement en status HTTP, code et message
func (h *HandlersOCR) convertirErreurTraitement(err error) (int, string, string) {
	// Erreurs OCR service
	var errOCR *services.ErreurOCR
	if errors.As(err, &errOCR) {
		switch errOCR.Code {
		case "LLM_NON_DISPONIBLE":
			return http.StatusServiceUnavailable, errOCR.Code, errOCR.Message
		default:
			return http.StatusBadRequest, errOCR.Code, errOCR.Message
		}
	}

	// Erreurs LLM
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

	// Erreur générique
	return http.StatusInternalServerError, "ERREUR_INTERNE", "Une erreur inattendue s'est produite"
}

// Vérification que HandlersOCR implémente les méthodes nécessaires
var _ interface {
	TraiterOCRHandler(*gin.Context)
} = (*HandlersOCR)(nil)
