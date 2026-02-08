// Package services contient les services métier de l'application
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/revisemieux/backend/internal/llm"
)

// Constantes de configuration OCR
const (
	// TailleMaxFichier est la taille maximale d'un fichier (10MB)
	TailleMaxFichier = 10 * 1024 * 1024
	// NombreMaxPages est le nombre maximum de pages/images par upload
	NombreMaxPages = 10
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
	ErrTropDePages           = &ErreurOCR{Code: "TROP_DE_PAGES", Message: "Trop de pages (max 10)"}
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

// ResultatOCRCours contient le résultat complet de l'OCR d'un cours
type ResultatOCRCours struct {
	Texte            string               `json:"texte"`
	Confiance        float64              `json:"confiance"`
	ZonesIncertaines []llm.ZoneIncertaine `json:"zones_incertaines"`
	NombrePages      int                  `json:"nombre_pages"`
	TitreSuggere     string               `json:"titre_suggere,omitempty"`
	MatiereSuggeree  string               `json:"matiere_suggeree,omitempty"`
}

// ServiceOCR gère l'extraction de texte des images et PDF
type ServiceOCR struct {
	gestionnaireLLM *llm.GestionnaireLLM
}

// NouveauServiceOCR crée une nouvelle instance du service OCR
func NouveauServiceOCR(gestionnaireLLM *llm.GestionnaireLLM) *ServiceOCR {
	return &ServiceOCR{
		gestionnaireLLM: gestionnaireLLM,
	}
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
	if len(images) > NombreMaxPages {
		return nil, ErrTropDePages
	}

	if len(images) == 0 {
		return nil, ErrFichierVide
	}

	// Traiter chaque image et combiner les résultats
	var textesExtraits []string
	var toutesZonesIncertaines []llm.ZoneIncertaine
	var confianceTotale float64
	offsetTexte := 0

	options := llm.OptionsOCRDefaut()

	for _, img := range images {
		resultat, err := s.gestionnaireLLM.ExtraireTexteImage(ctx, img, options)
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
	}, nil
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

	// Sinon, retourner l'image telle quelle
	return [][]byte{contenu}, nil
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

	if len(fichiers) > NombreMaxPages {
		return ErrTropDePages
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
