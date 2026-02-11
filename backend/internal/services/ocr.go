// Package services contient les services métier de l'application
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/image/draw"

	"github.com/revisemieux/backend/internal/llm"
	"github.com/revisemieux/backend/internal/ocr"
)

// Constantes de configuration OCR
const (
	// TailleMaxFichier est la taille maximale d'un fichier (10MB)
	TailleMaxFichier = 10 * 1024 * 1024

	// dimensionMaxImage est la dimension maximale (largeur ou hauteur) pour les images
	// envoyées au LLM. OpenAI redimensionne en interne à 2048px max, donc envoyer
	// plus grand ne fait que gaspiller de la bande passante.
	dimensionMaxImage = 2048

	// qualiteJPEG est la qualité de compression JPEG pour les images redimensionnées
	qualiteJPEG = 85
)

// Types MIME acceptés pour l'OCR
var typesMIMEAcceptes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"application/pdf": true,
}

// Extensions acceptées (fallback si MIME non détecté)
var extensionsAcceptees = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".pdf":  true,
}

// ErreurOCR représente une erreur du service OCR
type ErreurOCR struct {
	Code    string
	Message string
}

func (e *ErreurOCR) Error() string {
	return e.Message
}

// Codes d'erreur OCR
var (
	ErrTypeFichierInvalide   = &ErreurOCR{Code: "TYPE_INVALIDE", Message: "Type de fichier non supporté"}
	ErrFichierTropGrand      = &ErreurOCR{Code: "FICHIER_TROP_GRAND", Message: "Fichier trop volumineux (max 10MB)"}
	ErrFichierVide           = &ErreurOCR{Code: "FICHIER_VIDE", Message: "Fichier vide"}
	ErrLLMNonDisponible      = &ErreurOCR{Code: "LLM_NON_DISPONIBLE", Message: "Service LLM non disponible"}
	ErrExtractionPDFEchouee  = &ErreurOCR{Code: "EXTRACTION_PDF_ECHOUEE", Message: "Échec de l'extraction des pages PDF"}
)

// FichierUpload représente un fichier uploadé pour l'OCR
type FichierUpload struct {
	Nom      string
	Contenu  []byte
	TypeMIME string
}

// BlocTexteParPage regroupe les blocs de texte d'une page avec un index de page
type BlocTexteParPage struct {
	Page       int             `json:"page"`
	BlocsTexte []llm.BlocTexte `json:"blocs_texte"`
}

// ResultatOCRCours contient le résultat complet de l'OCR d'un cours
type ResultatOCRCours struct {
	Texte            string               `json:"texte"`
	Confiance        float64              `json:"confiance"`
	ZonesIncertaines []llm.ZoneIncertaine `json:"zones_incertaines"`
	NombrePages      int                  `json:"nombre_pages"`
	TitreSuggere     string               `json:"titre_suggere,omitempty"`
	MatiereSuggeree  string               `json:"matiere_suggeree,omitempty"`
	BlocsTexte       []BlocTexteParPage   `json:"blocs_texte,omitempty"`
}

// ServiceOCR gère l'extraction de texte des images et PDF
type ServiceOCR struct {
	gestionnaireLLM *llm.GestionnaireLLM
	tesseractActif  bool
	tesseractLang   string
	nombreMaxPages  int
}

// NouveauServiceOCR crée une nouvelle instance du service OCR
func NouveauServiceOCR(gestionnaireLLM *llm.GestionnaireLLM, tesseractEnabled bool, nombreMaxPages int) *ServiceOCR {
	tesseractActif := tesseractEnabled && ocr.TesseractDisponible()
	if nombreMaxPages <= 0 {
		nombreMaxPages = 30
	}
	return &ServiceOCR{
		gestionnaireLLM: gestionnaireLLM,
		tesseractActif:  tesseractActif,
		tesseractLang:   "fra",
		nombreMaxPages:  nombreMaxPages,
	}
}

// NombreMaxPages retourne la limite configurée de pages par upload
func (s *ServiceOCR) NombreMaxPages() int {
	return s.nombreMaxPages
}

// erreurTropDePages retourne une erreur avec le nombre max dynamique
func (s *ServiceOCR) erreurTropDePages() *ErreurOCR {
	return &ErreurOCR{
		Code:    "TROP_DE_PAGES",
		Message: fmt.Sprintf("Trop de pages (max %d)", s.nombreMaxPages),
	}
}

// TesseractActif retourne true si Tesseract est activé et disponible
func (s *ServiceOCR) TesseractActif() bool {
	return s.tesseractActif
}

// TraiterFichiers traite plusieurs fichiers uploadés et retourne le texte OCR combiné
func (s *ServiceOCR) TraiterFichiers(ctx context.Context, fichiers []*multipart.FileHeader) (*ResultatOCRCours, error) {
	if s.gestionnaireLLM == nil {
		return nil, ErrLLMNonDisponible
	}

	// Collecter toutes les images à traiter
	var images [][]byte
	for _, fh := range fichiers {
		imgs, err := s.extraireImages(fh)
		if err != nil {
			return nil, err
		}
		images = append(images, imgs...)
	}

	// Vérifier le nombre total de pages
	if len(images) > s.nombreMaxPages {
		return nil, s.erreurTropDePages()
	}

	if len(images) == 0 {
		return nil, ErrFichierVide
	}

	// Traiter chaque image et combiner les résultats
	var textesExtraits []string
	var toutesZonesIncertaines []llm.ZoneIncertaine
	var tousBlocsTexte []BlocTexteParPage
	var confianceTotale float64
	offsetTexte := 0

	options := llm.OptionsOCRDefaut()

	for i, img := range images {
		resultat, err := s.traiterImageHybride(ctx, img, options)
		if err != nil {
			return nil, fmt.Errorf("erreur OCR: %w", err)
		}

		textesExtraits = append(textesExtraits, resultat.Texte)
		confianceTotale += resultat.Confiance

		// Ajuster les indices des zones incertaines avec l'offset
		for _, zone := range resultat.ZonesIncertaines {
			zoneAjustee := llm.ZoneIncertaine{
				Debut:  zone.Debut + offsetTexte,
				Fin:    zone.Fin + offsetTexte,
				Texte:  zone.Texte,
				Raison: zone.Raison,
			}
			toutesZonesIncertaines = append(toutesZonesIncertaines, zoneAjustee)
		}

		// Collecter les blocs de texte avec positions pour cette page
		if len(resultat.BlocsTexte) > 0 {
			tousBlocsTexte = append(tousBlocsTexte, BlocTexteParPage{
				Page:       i,
				BlocsTexte: resultat.BlocsTexte,
			})
		}

		// Mettre à jour l'offset pour la prochaine page
		// +2 pour le séparateur "\n\n" entre les pages
		offsetTexte += len(resultat.Texte) + 2
	}

	// Combiner les textes avec des séparateurs
	texteCombine := strings.Join(textesExtraits, "\n\n")
	confianceMoyenne := confianceTotale / float64(len(images))

	// Extraire le titre et la matière suggérés
	titreSuggere, matiereSuggeree := s.extraireMetadonnees(ctx, texteCombine)

	return &ResultatOCRCours{
		Texte:            texteCombine,
		Confiance:        confianceMoyenne,
		ZonesIncertaines: toutesZonesIncertaines,
		NombrePages:      len(images),
		TitreSuggere:     titreSuggere,
		MatiereSuggeree:  matiereSuggeree,
		BlocsTexte:       tousBlocsTexte,
	}, nil
}

// CallbackProgression est appelé après chaque page traitée avec succès.
// Si le callback retourne une erreur, le traitement s'arrête.
type CallbackProgression func(page, total int) error

// TraiterFichiersAvecProgression traite les fichiers en envoyant un callback de progression après chaque page.
func (s *ServiceOCR) TraiterFichiersAvecProgression(ctx context.Context, fichiers []*multipart.FileHeader, onProgression CallbackProgression) (*ResultatOCRCours, error) {
	if s.gestionnaireLLM == nil {
		return nil, ErrLLMNonDisponible
	}

	// Collecter toutes les images à traiter
	var images [][]byte
	for _, fh := range fichiers {
		imgs, err := s.extraireImages(fh)
		if err != nil {
			return nil, err
		}
		images = append(images, imgs...)
	}

	if len(images) > s.nombreMaxPages {
		return nil, s.erreurTropDePages()
	}
	if len(images) == 0 {
		return nil, ErrFichierVide
	}

	total := len(images)
	var textesExtraits []string
	var toutesZonesIncertaines []llm.ZoneIncertaine
	var tousBlocsTexte []BlocTexteParPage
	var confianceTotale float64
	offsetTexte := 0

	options := llm.OptionsOCRDefaut()

	for i, img := range images {
		resultat, err := s.traiterImageHybride(ctx, img, options)
		if err != nil {
			return nil, fmt.Errorf("erreur OCR page %d: %w", i+1, err)
		}

		textesExtraits = append(textesExtraits, resultat.Texte)
		confianceTotale += resultat.Confiance

		for _, zone := range resultat.ZonesIncertaines {
			zoneAjustee := llm.ZoneIncertaine{
				Debut:  zone.Debut + offsetTexte,
				Fin:    zone.Fin + offsetTexte,
				Texte:  zone.Texte,
				Raison: zone.Raison,
			}
			toutesZonesIncertaines = append(toutesZonesIncertaines, zoneAjustee)
		}

		if len(resultat.BlocsTexte) > 0 {
			tousBlocsTexte = append(tousBlocsTexte, BlocTexteParPage{
				Page:       i,
				BlocsTexte: resultat.BlocsTexte,
			})
		}

		offsetTexte += len(resultat.Texte) + 2

		// Notifier la progression
		if onProgression != nil {
			if err := onProgression(i+1, total); err != nil {
				return nil, fmt.Errorf("client déconnecté: %w", err)
			}
		}
	}

	texteCombine := strings.Join(textesExtraits, "\n\n")
	confianceMoyenne := confianceTotale / float64(total)

	titreSuggere, matiereSuggeree := s.extraireMetadonnees(ctx, texteCombine)

	return &ResultatOCRCours{
		Texte:            texteCombine,
		Confiance:        confianceMoyenne,
		ZonesIncertaines: toutesZonesIncertaines,
		NombrePages:      total,
		TitreSuggere:     titreSuggere,
		MatiereSuggeree:  matiereSuggeree,
		BlocsTexte:       tousBlocsTexte,
	}, nil
}

// traiterImageHybride traite une image avec LLM + Tesseract en parallèle.
// Si Tesseract a une confiance suffisante → positions Tesseract (pixel-perfect) + texte LLM.
// Si Tesseract échoue ou est indisponible → fallback sur les blocs LLM (positions approximatives).
func (s *ServiceOCR) traiterImageHybride(ctx context.Context, img []byte, options llm.OptionsOCR) (*llm.ResultatOCR, error) {
	if !s.tesseractActif {
		// Sans Tesseract : LLM avec ses positions approximatives
		return s.gestionnaireLLM.ExtraireTexteImage(ctx, img, options)
	}

	type resultatLLM struct {
		resultat *llm.ResultatOCR
		err      error
	}
	type resultatTess struct {
		blocs   []ocr.BlocTesseract
		largeur int
		hauteur int
		err     error
	}

	var wg sync.WaitGroup
	chLLM := make(chan resultatLLM, 1)
	chTess := make(chan resultatTess, 1)

	// Lancer LLM en parallèle
	wg.Add(1)
	go func() {
		defer wg.Done()
		res, err := s.gestionnaireLLM.ExtraireTexteImage(ctx, img, options)
		chLLM <- resultatLLM{resultat: res, err: err}
	}()

	// Lancer Tesseract en parallèle
	wg.Add(1)
	go func() {
		defer wg.Done()
		blocs, largeur, hauteur, err := ocr.ExtraireBlocsTesseract(img, s.tesseractLang)
		chTess <- resultatTess{blocs: blocs, largeur: largeur, hauteur: hauteur, err: err}
	}()

	// Attendre les résultats
	resLLM := <-chLLM
	resTess := <-chTess
	wg.Wait()

	// Si le LLM a échoué, on ne peut rien faire
	if resLLM.err != nil {
		return nil, resLLM.err
	}

	// Si Tesseract a échoué, fallback sur les blocs LLM (positions approximatives)
	if resTess.err != nil {
		log.Printf("⚠️ Tesseract échoué, fallback sur positions LLM: %v", resTess.err)
		return resLLM.resultat, nil
	}

	// Fusionner : positions Tesseract (si confiance OK) ou fallback LLM
	resLLM.resultat.BlocsTexte = ocr.FusionnerBlocsHybride(
		resLLM.resultat.BlocsTexte,
		resTess.blocs,
		resTess.largeur,
		resTess.hauteur,
	)

	return resLLM.resultat, nil
}

// extraireImages extrait les images d'un fichier uploadé
// Pour les images, retourne le contenu tel quel
// Pour les PDF, extrait chaque page comme image
func (s *ServiceOCR) extraireImages(fh *multipart.FileHeader) ([][]byte, error) {
	// Vérifier la taille
	if fh.Size > TailleMaxFichier {
		return nil, ErrFichierTropGrand
	}

	if fh.Size == 0 {
		return nil, ErrFichierVide
	}

	// Vérifier le type MIME
	typeMIME := fh.Header.Get("Content-Type")
	ext := strings.ToLower(filepath.Ext(fh.Filename))

	// Si le type MIME n'est pas reconnu, essayer avec l'extension
	if !typesMIMEAcceptes[typeMIME] {
		if !extensionsAcceptees[ext] {
			return nil, ErrTypeFichierInvalide
		}
		// Déduire le type MIME de l'extension
		typeMIME = s.typeMIMEDepuisExtension(ext)
	}

	// Ouvrir le fichier
	file, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("erreur ouverture fichier: %w", err)
	}
	defer file.Close()

	// Lire le contenu
	contenu, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture fichier: %w", err)
	}

	// Détecter le type réel basé sur les magic bytes
	typeMIMEReel := s.detecterTypeMIME(contenu)
	if typeMIMEReel != "" {
		typeMIME = typeMIMEReel
	}

	// Si c'est un PDF, extraire les pages
	if typeMIME == "application/pdf" {
		return s.extrairePagesPDF(contenu)
	}

	// Redimensionner si nécessaire (économise bande passante, pas de perte de qualité OCR)
	contenuRedim, err := redimensionnerImage(contenu, typeMIME)
	if err != nil {
		log.Printf("Redimensionnement échoué, utilisation de l'image originale: %v", err)
		return [][]byte{contenu}, nil
	}

	return [][]byte{contenuRedim}, nil
}

// redimensionnerImage redimensionne une image si elle dépasse dimensionMaxImage pixels
// sur son plus grand côté. Retourne l'image originale si déjà assez petite.
func redimensionnerImage(data []byte, typeMIME string) ([]byte, error) {
	// Décoder l'image pour obtenir les dimensions
	reader := bytes.NewReader(data)
	cfg, _, err := image.DecodeConfig(reader)
	if err != nil {
		return data, nil // format non reconnu, on garde l'original
	}

	// Pas besoin de redimensionner si déjà assez petit
	if cfg.Width <= dimensionMaxImage && cfg.Height <= dimensionMaxImage {
		return data, nil
	}

	// Calculer les nouvelles dimensions en gardant le ratio
	newW, newH := cfg.Width, cfg.Height
	if newW > newH {
		newH = newH * dimensionMaxImage / newW
		newW = dimensionMaxImage
	} else {
		newW = newW * dimensionMaxImage / newH
		newH = dimensionMaxImage
	}

	// Décoder l'image complète
	reader.Reset(data)
	src, _, err := image.Decode(reader)
	if err != nil {
		return data, nil
	}

	// Redimensionner avec interpolation de qualité (CatmullRom)
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	// Ré-encoder en JPEG (meilleur ratio taille/qualité pour l'OCR)
	var buf bytes.Buffer
	if typeMIME == "image/png" {
		err = png.Encode(&buf, dst)
	} else {
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: qualiteJPEG})
	}
	if err != nil {
		return data, nil
	}

	log.Printf("Image redimensionnée: %dx%d → %dx%d (%d Ko → %d Ko)",
		cfg.Width, cfg.Height, newW, newH, len(data)/1024, buf.Len()/1024)

	return buf.Bytes(), nil
}

// extrairePagesPDF extrait les pages d'un PDF sous forme d'images
// Pour le MVP, on utilise une approche simplifiée: on envoie le PDF directement au LLM
// qui peut traiter les PDF multipage via l'API Vision
func (s *ServiceOCR) extrairePagesPDF(contenu []byte) ([][]byte, error) {
	// Note: Pour une implémentation plus robuste, on pourrait utiliser pdfcpu
	// pour extraire les pages et les convertir en images.
	// Pour le MVP, on envoie le PDF directement au LLM Vision API
	// qui supporte les PDF multipage.

	// Vérifier que c'est bien un PDF valide (magic bytes)
	if len(contenu) < 4 || string(contenu[:4]) != "%PDF" {
		return nil, ErrTypeFichierInvalide
	}

	// Pour simplifier le MVP, on traite le PDF comme une seule "image"
	// L'API OpenAI Vision accepte les PDF
	return [][]byte{contenu}, nil
}

// detecterTypeMIME détecte le type MIME basé sur les magic bytes
func (s *ServiceOCR) detecterTypeMIME(contenu []byte) string {
	if len(contenu) < 4 {
		return ""
	}

	// JPEG: FF D8 FF
	if len(contenu) >= 3 && contenu[0] == 0xFF && contenu[1] == 0xD8 && contenu[2] == 0xFF {
		return "image/jpeg"
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(contenu) >= 8 && bytes.Equal(contenu[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "image/png"
	}

	// WebP: RIFF....WEBP
	if len(contenu) >= 12 && string(contenu[:4]) == "RIFF" && string(contenu[8:12]) == "WEBP" {
		return "image/webp"
	}

	// PDF: %PDF
	if string(contenu[:4]) == "%PDF" {
		return "application/pdf"
	}

	return ""
}

// typeMIMEDepuisExtension retourne le type MIME pour une extension donnée
func (s *ServiceOCR) typeMIMEDepuisExtension(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	default:
		return ""
	}
}

// ValiderFichier valide un fichier avant upload
func (s *ServiceOCR) ValiderFichier(fh *multipart.FileHeader) error {
	// Vérifier la taille
	if fh.Size > TailleMaxFichier {
		return ErrFichierTropGrand
	}

	if fh.Size == 0 {
		return ErrFichierVide
	}

	// Vérifier le type MIME ou l'extension
	typeMIME := fh.Header.Get("Content-Type")
	ext := strings.ToLower(filepath.Ext(fh.Filename))

	if !typesMIMEAcceptes[typeMIME] && !extensionsAcceptees[ext] {
		return ErrTypeFichierInvalide
	}

	return nil
}

// ValiderFichiers valide tous les fichiers avant upload
func (s *ServiceOCR) ValiderFichiers(fichiers []*multipart.FileHeader) error {
	if len(fichiers) == 0 {
		return ErrFichierVide
	}

	if len(fichiers) > s.nombreMaxPages {
		return s.erreurTropDePages()
	}

	for _, fh := range fichiers {
		if err := s.ValiderFichier(fh); err != nil {
			return fmt.Errorf("%s: %w", fh.Filename, err)
		}
	}

	return nil
}

// EstErreurOCR vérifie si une erreur est une ErreurOCR
func EstErreurOCR(err error) bool {
	var errOCR *ErreurOCR
	return errors.As(err, &errOCR)
}

// MetadonneesCours représente les métadonnées extraites du texte
type MetadonneesCours struct {
	Titre   string `json:"titre"`
	Matiere string `json:"matiere"`
}

// Liste des matières valides
var matiereValides = []string{
	"mathematiques", "francais", "histoire", "geographie", "sciences",
	"anglais", "physique", "chimie", "svt", "ses", "philosophie",
	"espagnol", "allemand", "italien", "economie", "informatique",
}

// extraireMetadonnees extrait le titre et la matière suggérés du texte OCR
func (s *ServiceOCR) extraireMetadonnees(ctx context.Context, texte string) (titre, matiere string) {
	if s.gestionnaireLLM == nil || texte == "" {
		return "", ""
	}

	// Limiter le texte pour le prompt (les 2000 premiers caractères suffisent)
	texteAnalyse := texte
	if len(texteAnalyse) > 2000 {
		texteAnalyse = texteAnalyse[:2000]
	}

	prompt := fmt.Sprintf(`Analyse ce texte extrait d'un cours scolaire et déduis:
1. Un titre court et descriptif pour ce cours (max 50 caractères)
2. La matière scolaire parmi: %s

Texte du cours:
---
%s
---

Réponds uniquement au format JSON:
{"titre": "...", "matiere": "..."}

Si tu ne peux pas déterminer le titre, utilise les premiers mots significatifs.
Si tu ne peux pas déterminer la matière, utilise une chaîne vide.`,
		strings.Join(matiereValides, ", "), texteAnalyse)

	options := llm.OptionsGeneration{
		Temperature:   0.3, // Basse température pour des réponses cohérentes
		MaxTokens:     100,
		FormatReponse: "json",
	}

	// Utiliser GenererJSON serait idéal mais GenererTexte fonctionne aussi
	reponse, err := s.gestionnaireLLM.GenererTexte(ctx, prompt, options)
	if err != nil {
		// En cas d'erreur, retourner des valeurs vides (pas critique)
		return "", ""
	}

	// Parser la réponse JSON
	var metadonnees MetadonneesCours
	// Nettoyer la réponse (enlever les éventuels backticks markdown)
	reponse = strings.TrimSpace(reponse)
	reponse = strings.TrimPrefix(reponse, "```json")
	reponse = strings.TrimPrefix(reponse, "```")
	reponse = strings.TrimSuffix(reponse, "```")
	reponse = strings.TrimSpace(reponse)

	if err := parseJSON([]byte(reponse), &metadonnees); err != nil {
		return "", ""
	}

	// Valider la matière
	matiereNormalisee := strings.ToLower(strings.TrimSpace(metadonnees.Matiere))
	matiereValide := ""
	for _, m := range matiereValides {
		if m == matiereNormalisee {
			matiereValide = m
			break
		}
	}

	return strings.TrimSpace(metadonnees.Titre), matiereValide
}

// parseJSON est une fonction helper pour parser du JSON
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// RetraiterOCRImages re-traite des images brutes pour re-générer les blocs de texte OCR
func (s *ServiceOCR) RetraiterOCRImages(ctx context.Context, images [][]byte) (*ResultatOCRCours, error) {
	if s.gestionnaireLLM == nil {
		return nil, ErrLLMNonDisponible
	}
	if len(images) == 0 {
		return nil, ErrFichierVide
	}
	if len(images) > s.nombreMaxPages {
		return nil, s.erreurTropDePages()
	}

	// Même logique que TraiterFichiers mais avec des bytes bruts au lieu de FileHeaders
	var textesExtraits []string
	var toutesZonesIncertaines []llm.ZoneIncertaine
	var tousBlocsTexte []BlocTexteParPage
	var confianceTotale float64
	offsetTexte := 0

	options := llm.OptionsOCRDefaut()

	for i, img := range images {
		resultat, err := s.traiterImageHybride(ctx, img, options)
		if err != nil {
			return nil, fmt.Errorf("erreur OCR page %d: %w", i, err)
		}
		textesExtraits = append(textesExtraits, resultat.Texte)
		confianceTotale += resultat.Confiance
		for _, zone := range resultat.ZonesIncertaines {
			zoneAjustee := llm.ZoneIncertaine{
				Debut:  zone.Debut + offsetTexte,
				Fin:    zone.Fin + offsetTexte,
				Texte:  zone.Texte,
				Raison: zone.Raison,
			}
			toutesZonesIncertaines = append(toutesZonesIncertaines, zoneAjustee)
		}
		if len(resultat.BlocsTexte) > 0 {
			tousBlocsTexte = append(tousBlocsTexte, BlocTexteParPage{
				Page:       i,
				BlocsTexte: resultat.BlocsTexte,
			})
		}
		offsetTexte += len(resultat.Texte) + 2
	}

	texteCombine := strings.Join(textesExtraits, "\n\n")
	confianceMoyenne := confianceTotale / float64(len(images))
	titreSuggere, matiereSuggeree := s.extraireMetadonnees(ctx, texteCombine)

	return &ResultatOCRCours{
		Texte:            texteCombine,
		Confiance:        confianceMoyenne,
		ZonesIncertaines: toutesZonesIncertaines,
		NombrePages:      len(images),
		TitreSuggere:     titreSuggere,
		MatiereSuggeree:  matiereSuggeree,
		BlocsTexte:       tousBlocsTexte,
	}, nil
}
