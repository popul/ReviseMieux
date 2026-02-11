// Package api contient les handlers OCR pour le traitement d'images et PDF
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/services"
	"github.com/revisemieux/backend/internal/store"
)

// HandlersOCR contient les handlers OCR avec leurs dépendances
type HandlersOCR struct {
	serviceOCR        *services.ServiceOCR
	serviceStorage    *services.ServiceStorage
	coursRepo         store.CoursRepository
	serviceGeneration *services.ServiceGeneration
	serviceConcepts   *services.ServiceConcepts
}

// NouveauHandlersOCR crée une nouvelle instance des handlers OCR
func NouveauHandlersOCR(serviceOCR *services.ServiceOCR, serviceStorage *services.ServiceStorage, coursRepo store.CoursRepository, serviceGeneration *services.ServiceGeneration, serviceConcepts *services.ServiceConcepts) *HandlersOCR {
	return &HandlersOCR{
		serviceOCR:        serviceOCR,
		serviceStorage:    serviceStorage,
		coursRepo:         coursRepo,
		serviceGeneration: serviceGeneration,
		serviceConcepts:   serviceConcepts,
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
	TitreSuggere     string                 `json:"titreSuggere,omitempty"`
	MatiereSuggeree  string                 `json:"matiereSuggeree,omitempty"`
	BlocsTexte       []services.BlocTexteParPage `json:"blocsTexte,omitempty"`
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
		TitreSuggere:     resultat.TitreSuggere,
		MatiereSuggeree:  resultat.MatiereSuggeree,
		BlocsTexte:       resultat.BlocsTexte,
	}

	// Sauvegarder le cours si demandé
	if requete.Sauvegarder && h.coursRepo != nil {
		// Utiliser le titre fourni, sinon le titre suggéré, sinon un titre par défaut
		titre := requete.Titre
		if titre == "" && resultat.TitreSuggere != "" {
			titre = resultat.TitreSuggere
		}
		if titre == "" {
			titre = "Cours sans titre"
		}

		// Utiliser la matière fournie, sinon la matière suggérée
		matiere := requete.Matiere
		if matiere == "" && resultat.MatiereSuggeree != "" {
			matiere = resultat.MatiereSuggeree
		}

		// Générer un ID pour le cours avant de sauvegarder les images
		coursID := ""
		// Sérialiser les blocs de texte en JSON pour le stockage
		var blocsTexteJSON json.RawMessage
		if len(resultat.BlocsTexte) > 0 {
			blocsJSON, errJSON := json.Marshal(resultat.BlocsTexte)
			if errJSON == nil {
				blocsTexteJSON = json.RawMessage(blocsJSON)
			}
		}

		cours := &store.Cours{
			Titre:             titre,
			Matiere:           matiere,
			TexteOCR:          resultat.Texte,
			Confiance:         resultat.Confiance,
			ZonesIncertaines:  zonesIncertaines,
			FichiersOriginaux: nomsFichiers,
			Images:            []string{},
			BlocsTexte:        blocsTexteJSON,
		}

		// Créer le cours d'abord pour obtenir l'ID
		if err := h.coursRepo.Creer(c.Request.Context(), cours); err != nil {
			// Log l'erreur mais ne pas faire échouer la requête
			// Le texte OCR a été extrait avec succès
			c.JSON(http.StatusOK, reponse)
			return
		}
		coursID = cours.ID

		// Sauvegarder les images si le service de stockage est disponible
		if h.serviceStorage != nil {
			imagesSauvegardees := []string{}
			for _, fichier := range fichiers {
				// Vérifier si c'est une image (pas un PDF)
				contentType := fichier.Header.Get("Content-Type")
				if contentType == "image/jpeg" || contentType == "image/png" || contentType == "image/gif" || contentType == "image/webp" {
					nomImage, err := h.serviceStorage.SauvegarderImage(coursID, fichier)
					if err == nil {
						imagesSauvegardees = append(imagesSauvegardees, nomImage)
					}
				}
			}

			// Mettre à jour le cours avec les images
			if len(imagesSauvegardees) > 0 {
				cours.Images = imagesSauvegardees
				h.coursRepo.MettreAJour(c.Request.Context(), cours)
			}
		}

		reponse.CoursID = coursID

		// Lancer la génération automatique du résumé et des concepts en arrière-plan
		if coursID != "" && h.serviceGeneration != nil {
			go func(id string) {
				ctx := context.Background()
				if _, err := h.serviceGeneration.GenererResume(ctx, id); err != nil {
					log.Printf("Auto-génération résumé échouée pour cours %s: %v", id, err)
				} else {
					log.Printf("Résumé généré automatiquement pour cours %s", id)
				}
				if h.serviceConcepts != nil {
					if _, err := h.serviceConcepts.ExtraireConcepts(ctx, id); err != nil {
						log.Printf("Auto-extraction concepts échouée pour cours %s: %v", id, err)
					} else {
						log.Printf("Concepts extraits automatiquement pour cours %s", id)
					}
				}
			}(coursID)
		}
	}

	c.JSON(http.StatusOK, reponse)
}

// envoyerEvenementSSE écrit un événement SSE formaté dans le writer
func envoyerEvenementSSE(w http.ResponseWriter, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// TraiterOCRStreamHandler traite l'OCR avec streaming SSE page par page
// POST /api/ocr
func (h *HandlersOCR) TraiterOCRStreamHandler(c *gin.Context) {
	// Phase de validation (avant SSE headers) — erreurs en JSON classique
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

	var requete RequeteOCR
	if err := c.ShouldBind(&requete); err != nil {
		// Ignorer l'erreur de binding, les champs sont optionnels
	}

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

	// Validation OK — basculer en mode SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	// Callback de progression
	onProgression := func(page, total int) error {
		select {
		case <-c.Request.Context().Done():
			return fmt.Errorf("client déconnecté")
		default:
		}
		envoyerEvenementSSE(c.Writer, "progress", map[string]int{
			"page":  page,
			"total": total,
		})
		return nil
	}

	// Traiter les fichiers avec progression
	resultat, err := h.serviceOCR.TraiterFichiersAvecProgression(c.Request.Context(), fichiers, onProgression)
	if err != nil {
		_, codeErreur, message := h.convertirErreurTraitement(err)
		envoyerEvenementSSE(c.Writer, "error", map[string]string{
			"code":    codeErreur,
			"message": message,
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

	nomsFichiers := make([]string, len(fichiers))
	for i, f := range fichiers {
		nomsFichiers[i] = f.Filename
	}

	reponse := ReponseOCR{
		Succes:           true,
		Texte:            resultat.Texte,
		Confiance:        resultat.Confiance,
		ZonesIncertaines: zonesIncertaines,
		NombrePages:      resultat.NombrePages,
		TitreSuggere:     resultat.TitreSuggere,
		MatiereSuggeree:  resultat.MatiereSuggeree,
		BlocsTexte:       resultat.BlocsTexte,
	}

	// Sauvegarder le cours si demandé
	if requete.Sauvegarder && h.coursRepo != nil {
		titre := requete.Titre
		if titre == "" && resultat.TitreSuggere != "" {
			titre = resultat.TitreSuggere
		}
		if titre == "" {
			titre = "Cours sans titre"
		}

		matiere := requete.Matiere
		if matiere == "" && resultat.MatiereSuggeree != "" {
			matiere = resultat.MatiereSuggeree
		}

		var blocsTexteJSON json.RawMessage
		if len(resultat.BlocsTexte) > 0 {
			blocsJSON, errJSON := json.Marshal(resultat.BlocsTexte)
			if errJSON == nil {
				blocsTexteJSON = json.RawMessage(blocsJSON)
			}
		}

		cours := &store.Cours{
			Titre:             titre,
			Matiere:           matiere,
			TexteOCR:          resultat.Texte,
			Confiance:         resultat.Confiance,
			ZonesIncertaines:  zonesIncertaines,
			FichiersOriginaux: nomsFichiers,
			Images:            []string{},
			BlocsTexte:        blocsTexteJSON,
		}

		if err := h.coursRepo.Creer(c.Request.Context(), cours); err != nil {
			envoyerEvenementSSE(c.Writer, "complete", reponse)
			return
		}
		reponse.CoursID = cours.ID

		if h.serviceStorage != nil {
			imagesSauvegardees := []string{}
			for _, fichier := range fichiers {
				contentType := fichier.Header.Get("Content-Type")
				if contentType == "image/jpeg" || contentType == "image/png" || contentType == "image/gif" || contentType == "image/webp" {
					nomImage, err := h.serviceStorage.SauvegarderImage(cours.ID, fichier)
					if err == nil {
						imagesSauvegardees = append(imagesSauvegardees, nomImage)
					}
				}
			}
			if len(imagesSauvegardees) > 0 {
				cours.Images = imagesSauvegardees
				h.coursRepo.MettreAJour(c.Request.Context(), cours)
			}
		}

		// Lancer génération automatique en arrière-plan
		if cours.ID != "" && h.serviceGeneration != nil {
			go func(id string) {
				ctx := context.Background()
				if _, err := h.serviceGeneration.GenererResume(ctx, id); err != nil {
					log.Printf("Auto-génération résumé échouée pour cours %s: %v", id, err)
				} else {
					log.Printf("Résumé généré automatiquement pour cours %s", id)
				}
				if h.serviceConcepts != nil {
					if _, err := h.serviceConcepts.ExtraireConcepts(ctx, id); err != nil {
						log.Printf("Auto-extraction concepts échouée pour cours %s: %v", id, err)
					} else {
						log.Printf("Concepts extraits automatiquement pour cours %s", id)
					}
				}
			}(cours.ID)
		}
	}

	envoyerEvenementSSE(c.Writer, "complete", reponse)
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

// RetraiterOCRCoursHandler relance l'OCR sur un cours existant pour générer les blocsTexte
// POST /api/cours/:id/reocr
func (h *HandlersOCR) RetraiterOCRCoursHandler(c *gin.Context) {
	if h.serviceOCR == nil || h.serviceStorage == nil || h.coursRepo == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseOCR{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service OCR ou de stockage n'est pas configuré",
			},
		})
		return
	}

	coursID := c.Param("id")
	if coursID == "" {
		c.JSON(http.StatusBadRequest, ReponseOCR{
			Succes: false,
			Erreur: &ErreurReponse{Code: "ID_MANQUANT", Message: "L'identifiant du cours est requis"},
		})
		return
	}

	// Récupérer le cours
	cours, err := h.coursRepo.ObtenirParID(c.Request.Context(), coursID)
	if err != nil {
		c.JSON(http.StatusNotFound, ReponseOCR{
			Succes: false,
			Erreur: &ErreurReponse{Code: "COURS_NON_TROUVE", Message: "Cours non trouvé"},
		})
		return
	}

	if len(cours.Images) == 0 {
		c.JSON(http.StatusBadRequest, ReponseOCR{
			Succes: false,
			Erreur: &ErreurReponse{Code: "PAS_D_IMAGES", Message: "Ce cours n'a pas d'images à re-traiter"},
		})
		return
	}

	// Lire toutes les images depuis le stockage
	var images [][]byte
	for _, nomImage := range cours.Images {
		cheminImage, err := h.serviceStorage.ObtenirImage(coursID, nomImage)
		if err != nil {
			continue // Ignorer les images manquantes
		}
		data, err := os.ReadFile(cheminImage)
		if err != nil {
			continue
		}
		images = append(images, data)
	}

	if len(images) == 0 {
		c.JSON(http.StatusBadRequest, ReponseOCR{
			Succes: false,
			Erreur: &ErreurReponse{Code: "IMAGES_ILLISIBLES", Message: "Impossible de lire les images du cours"},
		})
		return
	}

	// Lancer l'OCR sur les images
	resultat, err := h.serviceOCR.RetraiterOCRImages(c.Request.Context(), images)
	if err != nil {
		statusCode, codeErreur, message := h.convertirErreurTraitement(err)
		c.JSON(statusCode, ReponseOCR{
			Succes: false,
			Erreur: &ErreurReponse{Code: codeErreur, Message: message},
		})
		return
	}

	// Mettre à jour le cours avec les nouveaux résultats OCR
	cours.TexteOCR = resultat.Texte
	cours.Confiance = resultat.Confiance

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
	cours.ZonesIncertaines = zonesIncertaines

	// Mettre à jour les blocs de texte (toujours, même si nil pour effacer les anciens)
	if len(resultat.BlocsTexte) > 0 {
		blocsJSON, errJSON := json.Marshal(resultat.BlocsTexte)
		if errJSON == nil {
			cours.BlocsTexte = json.RawMessage(blocsJSON)
		}
	} else {
		cours.BlocsTexte = nil
	}

	if err := h.coursRepo.MettreAJour(c.Request.Context(), cours); err != nil {
		c.JSON(http.StatusInternalServerError, ReponseOCR{
			Succes: false,
			Erreur: &ErreurReponse{Code: "ERREUR_MAJ", Message: "Erreur lors de la mise à jour du cours"},
		})
		return
	}

	// Lancer la génération automatique du résumé et des concepts en arrière-plan
	if coursID != "" && h.serviceGeneration != nil {
		go func(id string) {
			ctx := context.Background()
			if _, err := h.serviceGeneration.GenererResume(ctx, id); err != nil {
				log.Printf("Auto-génération résumé échouée pour cours %s: %v", id, err)
			} else {
				log.Printf("Résumé généré automatiquement pour cours %s", id)
			}
			if h.serviceConcepts != nil {
				if _, err := h.serviceConcepts.ExtraireConcepts(ctx, id); err != nil {
					log.Printf("Auto-extraction concepts échouée pour cours %s: %v", id, err)
				} else {
					log.Printf("Concepts extraits automatiquement pour cours %s", id)
				}
			}
		}(coursID)
	}

	c.JSON(http.StatusOK, ReponseOCR{
		Succes:           true,
		Texte:            resultat.Texte,
		Confiance:        resultat.Confiance,
		ZonesIncertaines: zonesIncertaines,
		NombrePages:      resultat.NombrePages,
		CoursID:          coursID,
		BlocsTexte:       resultat.BlocsTexte,
	})
}

// Vérification que HandlersOCR implémente les méthodes nécessaires
var _ interface {
	TraiterOCRHandler(*gin.Context)
} = (*HandlersOCR)(nil)
