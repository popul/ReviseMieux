// Package api contient les handlers OCR pour le traitement d'images et PDF
package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	Succes           bool                        `json:"succes"`
	Texte            string                      `json:"texte,omitempty"`
	Confiance        float64                     `json:"confiance,omitempty"`
	ZonesIncertaines []store.ZoneIncertaine      `json:"zonesIncertaines,omitempty"`
	NombrePages      int                         `json:"nombrePages,omitempty"`
	CoursID          string                      `json:"coursId,omitempty"`
	TitreSuggere     string                      `json:"titreSuggere,omitempty"`
	MatiereSuggeree  string                      `json:"matiereSuggeree,omitempty"`
	BlocsTexte       []services.BlocTexteParPage `json:"blocsTexte,omitempty"`
	Erreur           *ErreurReponse              `json:"erreur,omitempty"`
}

// ReponseUpload représente la réponse JSON rapide de l'upload OCR
type ReponseUpload struct {
	Succes      bool           `json:"succes"`
	CoursID     string         `json:"coursId,omitempty"`
	NombrePages int            `json:"nombrePages,omitempty"`
	Erreur      *ErreurReponse `json:"erreur,omitempty"`
}

// ErreurReponse représente une erreur dans la réponse JSON
type ErreurReponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RequeteOCR représente les paramètres optionnels de la requête OCR
type RequeteOCR struct {
	Titre   string `form:"titre"`
	Matiere string `form:"matiere"`
}

// UploadOCRHandler upload les fichiers, crée le cours, et lance le traitement OCR en arrière-plan.
// POST /api/ocr
// Content-Type: multipart/form-data
// Retourne immédiatement le coursId. Le client poll ensuite GET /api/cours/:id pour la progression.
func (h *HandlersOCR) UploadOCRHandler(c *gin.Context) {
	if h.serviceOCR == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseUpload{
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
		c.JSON(http.StatusBadRequest, ReponseUpload{
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
		c.JSON(http.StatusBadRequest, ReponseUpload{
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
		c.JSON(http.StatusBadRequest, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    codeErreur,
				Message: message,
			},
		})
		return
	}

	// Extraire les images en mémoire (avant que le multipart soit libéré)
	imagesBytes, err := h.serviceOCR.ExtraireImagesMultipart(fichiers)
	if err != nil {
		codeErreur, message := h.convertirErreurOCR(err)
		c.JSON(http.StatusBadRequest, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    codeErreur,
				Message: message,
			},
		})
		return
	}

	nombrePages := len(imagesBytes)

	// Titre initial
	titre := requete.Titre
	if titre == "" {
		titre = "Cours sans titre"
	}

	// Créer le cours en DB avec statut "en_cours"
	cours := &store.Cours{
		Titre:             titre,
		Matiere:           requete.Matiere,
		TexteOCR:          "",
		Confiance:         0,
		ZonesIncertaines:  []store.ZoneIncertaine{},
		FichiersOriginaux: collecterNomsFichiers(fichiers),
		Images:            []string{},
		StatutOCR:         "en_cours",
		NombrePages:       nombrePages,
		PagesTraitees:     0,
	}

	if err := h.coursRepo.Creer(c.Request.Context(), cours); err != nil {
		c.JSON(http.StatusInternalServerError, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "ERREUR_CREATION",
				Message: "Erreur lors de la création du cours",
			},
		})
		return
	}

	// Sauvegarder les images sur disque
	if h.serviceStorage != nil {
		imagesSauvegardees := []string{}
		for _, imgBytes := range imagesBytes {
			nomImage, err := h.serviceStorage.SauvegarderImageBytes(cours.ID, imgBytes, ".jpg")
			if err == nil {
				imagesSauvegardees = append(imagesSauvegardees, nomImage)
			}
		}
		if len(imagesSauvegardees) > 0 {
			cours.Images = imagesSauvegardees
			h.coursRepo.MettreAJour(c.Request.Context(), cours)
		}
	}

	// Lancer le traitement OCR en arrière-plan
	go h.traiterOCRAsynchrone(cours.ID, imagesBytes, cours.Images, requete.Titre, requete.Matiere)

	c.JSON(http.StatusOK, ReponseUpload{
		Succes:      true,
		CoursID:     cours.ID,
		NombrePages: nombrePages,
	})
}

// traiterOCRAsynchrone traite les images OCR en arrière-plan page par page
func (h *HandlersOCR) traiterOCRAsynchrone(coursID string, images [][]byte, nomImages []string, titreRequete string, matiereRequete string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC dans traiterOCRAsynchrone cours %s: %v", coursID, r)
		}
	}()

	log.Printf("Démarrage OCR asynchrone cours %s (%d pages)", coursID, len(images))
	ctx := context.Background()

	onPageTraitee := func(pagesTerminees int, resultatPartiel *services.ResultatOCRCours, pageIdx int, imageCorrigee []byte) {
		// Si l'image a été pivotée, mettre à jour le fichier sur disque
		if imageCorrigee != nil && h.serviceStorage != nil && pageIdx < len(nomImages) {
			if err := h.serviceStorage.EcraserImage(coursID, nomImages[pageIdx], imageCorrigee); err != nil {
				log.Printf("Erreur écriture image pivotée page %d cours %s: %v", pageIdx+1, coursID, err)
			} else {
				log.Printf("Image pivotée sauvegardée page %d cours %s", pageIdx+1, coursID)
			}
		}

		// Déterminer le titre et la matière
		titre := titreRequete
		if titre == "" && resultatPartiel.TitreSuggere != "" {
			titre = resultatPartiel.TitreSuggere
		}
		if titre == "" {
			titre = "Cours sans titre"
		}

		matiere := matiereRequete
		if matiere == "" && resultatPartiel.MatiereSuggeree != "" {
			matiere = resultatPartiel.MatiereSuggeree
		}

		zonesJSON, _ := json.Marshal(convertirZonesIncertaines(resultatPartiel.ZonesIncertaines))
		blocsJSON, _ := json.Marshal(resultatPartiel.BlocsTexte)

		if err := h.coursRepo.MettreAJourProgressionOCR(ctx, coursID, pagesTerminees,
			resultatPartiel.Texte, resultatPartiel.Confiance,
			zonesJSON, blocsJSON, titre, matiere); err != nil {
			log.Printf("Erreur mise à jour progression OCR cours %s: %v", coursID, err)
		}
	}

	resultat, err := h.serviceOCR.TraiterImagesProgressif(ctx, images, onPageTraitee)
	if err != nil {
		log.Printf("Erreur traitement OCR asynchrone cours %s: %v", coursID, err)
		if errEchouer := h.coursRepo.EchouerOCR(ctx, coursID, err.Error()); errEchouer != nil {
			log.Printf("Erreur marquage échec OCR cours %s: %v", coursID, errEchouer)
		}
		return
	}

	// Déterminer le titre et la matière finaux
	titre := titreRequete
	if titre == "" && resultat.TitreSuggere != "" {
		titre = resultat.TitreSuggere
	}
	if titre == "" {
		titre = "Cours sans titre"
	}

	matiere := matiereRequete
	if matiere == "" && resultat.MatiereSuggeree != "" {
		matiere = resultat.MatiereSuggeree
	}

	zonesJSON, _ := json.Marshal(convertirZonesIncertaines(resultat.ZonesIncertaines))
	blocsJSON, _ := json.Marshal(resultat.BlocsTexte)

	if err := h.coursRepo.TerminerOCR(ctx, coursID,
		resultat.Texte, resultat.Confiance,
		zonesJSON, blocsJSON, titre, matiere); err != nil {
		log.Printf("Erreur finalisation OCR cours %s: %v", coursID, err)
		return
	}

	log.Printf("OCR terminé pour cours %s (%d pages)", coursID, resultat.NombrePages)

	// Pas d'auto-génération : l'utilisateur vérifie d'abord le texte OCR
	// puis clique sur "Générer le résumé" manuellement
}

// PivoterImageHandler pivote une image de 90° à gauche ou à droite, écrase le fichier
// et relance l'OCR de la page.
// POST /api/cours/:id/images/:filename/pivoter?sens=gauche|droite
func (h *HandlersOCR) PivoterImageHandler(c *gin.Context) {
	if h.serviceOCR == nil || h.serviceStorage == nil || h.coursRepo == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{
				Code:    "SERVICE_NON_DISPONIBLE",
				Message: "Le service OCR ou de stockage n'est pas configuré",
			},
		})
		return
	}

	coursID := c.Param("id")
	filename := c.Param("filename")
	sens := c.Query("sens")

	if coursID == "" || filename == "" {
		c.JSON(http.StatusBadRequest, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "PARAMS_MANQUANTS", Message: "L'identifiant du cours et le nom de fichier sont requis"},
		})
		return
	}

	var angle int
	switch sens {
	case "gauche":
		angle = 270
	case "droite":
		angle = 90
	default:
		c.JSON(http.StatusBadRequest, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "SENS_INVALIDE", Message: "Le paramètre 'sens' doit être 'gauche' ou 'droite'"},
		})
		return
	}

	cours, err := h.coursRepo.ObtenirParID(c.Request.Context(), coursID)
	if err != nil {
		c.JSON(http.StatusNotFound, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "COURS_NON_TROUVE", Message: "Cours non trouvé"},
		})
		return
	}

	// Lire l'image
	cheminImage, err := h.serviceStorage.ObtenirImage(coursID, filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "IMAGE_ILLISIBLE", Message: "Impossible de trouver l'image"},
		})
		return
	}
	imageData, err := os.ReadFile(cheminImage)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "IMAGE_ILLISIBLE", Message: "Impossible de lire l'image"},
		})
		return
	}

	// Pivoter l'image
	pivoted, err := services.PivoterImage(imageData, angle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "ERREUR_ROTATION", Message: "Erreur lors de la rotation de l'image"},
		})
		return
	}

	// Écraser le fichier sur disque
	if err := h.serviceStorage.EcraserImage(coursID, filename, pivoted); err != nil {
		c.JSON(http.StatusInternalServerError, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "ERREUR_ECRITURE", Message: "Erreur lors de la sauvegarde de l'image pivotée"},
		})
		return
	}

	// Trouver l'index de la page
	pageIdx := -1
	for i, img := range cours.Images {
		if img == filename {
			pageIdx = i
			break
		}
	}
	if pageIdx == -1 {
		c.JSON(http.StatusBadRequest, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "IMAGE_NON_TROUVEE", Message: "L'image ne fait pas partie de ce cours"},
		})
		return
	}

	// Passer le cours en statut "en_cours" (1 page)
	cours.StatutOCR = "en_cours"
	cours.NombrePages = 1
	cours.PagesTraitees = 0
	cours.ErreurOCR = ""
	if err := h.coursRepo.MettreAJour(c.Request.Context(), cours); err != nil {
		c.JSON(http.StatusInternalServerError, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "ERREUR_MAJ", Message: "Erreur lors de la mise à jour du cours"},
		})
		return
	}

	go h.traiterOCRPageAsynchrone(cours, pageIdx, pivoted, filename, false)

	c.JSON(http.StatusOK, ReponseUpload{
		Succes:      true,
		CoursID:     coursID,
		NombrePages: 1,
	})
}

// convertirZonesIncertaines convertit les zones du format llm vers store
func convertirZonesIncertaines(zones []llm.ZoneIncertaine) []store.ZoneIncertaine {
	result := make([]store.ZoneIncertaine, len(zones))
	for i, zone := range zones {
		result[i] = store.ZoneIncertaine{
			Debut:  zone.Debut,
			Fin:    zone.Fin,
			Texte:  zone.Texte,
			Raison: zone.Raison,
		}
	}
	return result
}

// collecterNomsFichiers extrait les noms de fichiers des multipart FileHeaders
func collecterNomsFichiers(fichiers []*multipart.FileHeader) []string {
	noms := make([]string, len(fichiers))
	for i, f := range fichiers {
		noms[i] = f.Filename
	}
	return noms
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

// RetraiterOCRCoursHandler relance l'OCR sur un cours existant (asynchrone).
// POST /api/cours/:id/reocr
// POST /api/cours/:id/reocr?page=N  (re-OCR d'une seule page, 0-indexed)
// Retourne immédiatement. Le client poll GET /api/cours/:id pour la progression.
func (h *HandlersOCR) RetraiterOCRCoursHandler(c *gin.Context) {
	if h.serviceOCR == nil || h.serviceStorage == nil || h.coursRepo == nil {
		c.JSON(http.StatusServiceUnavailable, ReponseUpload{
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
		c.JSON(http.StatusBadRequest, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "ID_MANQUANT", Message: "L'identifiant du cours est requis"},
		})
		return
	}

	cours, err := h.coursRepo.ObtenirParID(c.Request.Context(), coursID)
	if err != nil {
		c.JSON(http.StatusNotFound, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "COURS_NON_TROUVE", Message: "Cours non trouvé"},
		})
		return
	}

	if len(cours.Images) == 0 {
		c.JSON(http.StatusBadRequest, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "PAS_D_IMAGES", Message: "Ce cours n'a pas d'images à re-traiter"},
		})
		return
	}

	// Détection d'orientation : activée par défaut, désactivée si detecterOrientation=false
	detecterOrientation := c.Query("detecterOrientation") != "false"

	// Vérifier si une seule page est demandée
	pageParam := c.Query("page")
	if pageParam != "" {
		pageIdx, err := strconv.Atoi(pageParam)
		if err != nil || pageIdx < 0 || pageIdx >= len(cours.Images) {
			c.JSON(http.StatusBadRequest, ReponseUpload{
				Succes: false,
				Erreur: &ErreurReponse{Code: "PAGE_INVALIDE", Message: "Index de page invalide"},
			})
			return
		}
		h.retraiterPageUnique(c, cours, pageIdx, detecterOrientation)
		return
	}

	// Re-OCR de toutes les pages
	var images [][]byte
	for _, nomImage := range cours.Images {
		cheminImage, err := h.serviceStorage.ObtenirImage(coursID, nomImage)
		if err != nil {
			continue
		}
		data, err := os.ReadFile(cheminImage)
		if err != nil {
			continue
		}
		images = append(images, data)
	}

	if len(images) == 0 {
		c.JSON(http.StatusBadRequest, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "IMAGES_ILLISIBLES", Message: "Impossible de lire les images du cours"},
		})
		return
	}

	// Passer le cours en statut "en_cours"
	cours.StatutOCR = "en_cours"
	cours.NombrePages = len(images)
	cours.PagesTraitees = 0
	cours.ErreurOCR = ""
	if err := h.coursRepo.MettreAJour(c.Request.Context(), cours); err != nil {
		c.JSON(http.StatusInternalServerError, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "ERREUR_MAJ", Message: "Erreur lors de la mise à jour du cours"},
		})
		return
	}

	go h.traiterOCRAsynchrone(coursID, images, cours.Images, cours.Titre, cours.Matiere)

	c.JSON(http.StatusOK, ReponseUpload{
		Succes:      true,
		CoursID:     coursID,
		NombrePages: len(images),
	})
}

// retraiterPageUnique relance l'OCR sur une seule page d'un cours (asynchrone).
func (h *HandlersOCR) retraiterPageUnique(c *gin.Context, cours *store.Cours, pageIdx int, detecterOrientation bool) {
	nomImage := cours.Images[pageIdx]
	cheminImage, err := h.serviceStorage.ObtenirImage(cours.ID, nomImage)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "IMAGE_ILLISIBLE", Message: "Impossible de lire l'image"},
		})
		return
	}
	imageData, err := os.ReadFile(cheminImage)
	if err != nil {
		c.JSON(http.StatusBadRequest, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "IMAGE_ILLISIBLE", Message: "Impossible de lire l'image"},
		})
		return
	}

	// Passer le cours en statut "en_cours" (1 page)
	cours.StatutOCR = "en_cours"
	cours.NombrePages = 1
	cours.PagesTraitees = 0
	cours.ErreurOCR = ""
	if err := h.coursRepo.MettreAJour(c.Request.Context(), cours); err != nil {
		c.JSON(http.StatusInternalServerError, ReponseUpload{
			Succes: false,
			Erreur: &ErreurReponse{Code: "ERREUR_MAJ", Message: "Erreur lors de la mise à jour du cours"},
		})
		return
	}

	go h.traiterOCRPageAsynchrone(cours, pageIdx, imageData, nomImage, detecterOrientation)

	c.JSON(http.StatusOK, ReponseUpload{
		Succes:      true,
		CoursID:     cours.ID,
		NombrePages: 1,
	})
}

// traiterOCRPageAsynchrone traite une seule page en arrière-plan et fusionne le résultat.
// Si detecterOrientation est true, le LLM détecte l'orientation avant l'OCR.
func (h *HandlersOCR) traiterOCRPageAsynchrone(cours *store.Cours, pageIdx int, imageData []byte, nomImage string, detecterOrientation bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC dans traiterOCRPageAsynchrone cours %s page %d: %v", cours.ID, pageIdx, r)
		}
	}()

	log.Printf("Démarrage OCR page %d cours %s", pageIdx+1, cours.ID)
	ctx := context.Background()

	resultat, err := h.serviceOCR.RetraiterOCRImages(ctx, [][]byte{imageData}, detecterOrientation)
	if err != nil {
		log.Printf("Erreur OCR page %d cours %s: %v", pageIdx+1, cours.ID, err)
		if errEchouer := h.coursRepo.EchouerOCR(ctx, cours.ID, err.Error()); errEchouer != nil {
			log.Printf("Erreur marquage échec OCR cours %s: %v", cours.ID, errEchouer)
		}
		return
	}

	// Sauvegarder l'image pivotée si nécessaire
	if h.serviceStorage != nil && len(resultat.ImagesCorrigees) > 0 {
		if err := h.serviceStorage.EcraserImage(cours.ID, nomImage, resultat.ImagesCorrigees[0]); err != nil {
			log.Printf("Erreur écriture image pivotée page %d cours %s: %v", pageIdx+1, cours.ID, err)
		}
	}

	// Recharger le cours pour avoir les données les plus récentes
	coursFrais, err := h.coursRepo.ObtenirParID(ctx, cours.ID)
	if err != nil {
		log.Printf("Erreur rechargement cours %s: %v", cours.ID, err)
		return
	}

	// Fusionner les blocs : parser les blocs existants, remplacer ceux de la page
	var blocsExistants []services.BlocTexteParPage
	if coursFrais.BlocsTexte != nil {
		json.Unmarshal(coursFrais.BlocsTexte, &blocsExistants)
	}

	// Retirer les blocs de cette page
	var blocsFiltres []services.BlocTexteParPage
	for _, b := range blocsExistants {
		if b.Page != pageIdx {
			blocsFiltres = append(blocsFiltres, b)
		}
	}

	// Ajouter les nouveaux blocs de cette page (résultat a page=0, on remappe)
	for _, b := range resultat.BlocsTexte {
		blocsFiltres = append(blocsFiltres, services.BlocTexteParPage{
			Page:       pageIdx,
			BlocsTexte: b.BlocsTexte,
		})
	}

	// Reconstruire le texte à partir de tous les blocs triés par page
	var textesParPage []string
	pageMap := make(map[int][]string)
	for _, b := range blocsFiltres {
		for _, bloc := range b.BlocsTexte {
			pageMap[b.Page] = append(pageMap[b.Page], bloc.Texte)
		}
	}
	for i := 0; i < len(coursFrais.Images); i++ {
		if textes, ok := pageMap[i]; ok {
			textesParPage = append(textesParPage, strings.Join(textes, "\n"))
		}
	}
	texteComplet := strings.Join(textesParPage, "\n\n")

	// Recalculer la confiance moyenne
	var confianceTotale float64
	var nbBlocs int
	for _, b := range blocsFiltres {
		for _, bloc := range b.BlocsTexte {
			confianceTotale += bloc.Confiance
			nbBlocs++
		}
	}
	confiance := float64(0)
	if nbBlocs > 0 {
		confiance = confianceTotale / float64(nbBlocs)
	}

	blocsJSON, _ := json.Marshal(blocsFiltres)
	zonesJSON, _ := json.Marshal(coursFrais.ZonesIncertaines)

	if err := h.coursRepo.TerminerOCR(ctx, cours.ID, texteComplet, confiance, zonesJSON, blocsJSON, coursFrais.Titre, coursFrais.Matiere); err != nil {
		log.Printf("Erreur finalisation OCR page %d cours %s: %v", pageIdx+1, cours.ID, err)
		return
	}

	log.Printf("OCR page %d terminé pour cours %s", pageIdx+1, cours.ID)
}
