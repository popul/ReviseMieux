// Package ocr fournit l'extraction de texte hybride Tesseract + LLM
package ocr

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"sync"

	"github.com/otiai10/gosseract/v2"
)

// tesseractMu sérialise les appels gosseract (CGo) car la librairie C Tesseract n'est pas thread-safe.
var tesseractMu sync.Mutex

// BlocTesseract représente un bloc de texte détecté par Tesseract avec sa position en pixels
type BlocTesseract struct {
	Texte     string
	X, Y      int     // pixels coin supérieur gauche
	Largeur   int     // pixels
	Hauteur   int     // pixels
	Confiance float64 // 0-100
}

// ExtraireBlocsTesseract extrait les blocs de texte d'une image avec Tesseract.
// Retourne les blocs, la largeur et hauteur de l'image en pixels.
func ExtraireBlocsTesseract(imageBytes []byte, lang string) ([]BlocTesseract, int, int, error) {
	tesseractMu.Lock()
	defer tesseractMu.Unlock()

	client := gosseract.NewClient()
	defer client.Close()

	if lang != "" {
		if err := client.SetLanguage(lang); err != nil {
			return nil, 0, 0, fmt.Errorf("erreur langue tesseract: %w", err)
		}
	}

	if err := client.SetImageFromBytes(imageBytes); err != nil {
		return nil, 0, 0, fmt.Errorf("erreur chargement image tesseract: %w", err)
	}

	// Extraire les bounding boxes au niveau paragraphe (RIL_BLOCK)
	// RIL_TEXTLINE produit trop de faux positifs sur le manuscrit
	boxes, err := client.GetBoundingBoxes(gosseract.RIL_BLOCK)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("erreur extraction blocs tesseract: %w", err)
	}

	// Obtenir les vraies dimensions de l'image
	config, _, err := image.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("erreur lecture dimensions image: %w", err)
	}
	largeurImage := config.Width
	hauteurImage := config.Height

	// Convertir en BlocTesseract
	var blocs []BlocTesseract
	for _, box := range boxes {
		texte := box.Word
		if texte == "" {
			continue
		}
		blocs = append(blocs, BlocTesseract{
			Texte:     texte,
			X:         box.Box.Min.X,
			Y:         box.Box.Min.Y,
			Largeur:   box.Box.Max.X - box.Box.Min.X,
			Hauteur:   box.Box.Max.Y - box.Box.Min.Y,
			Confiance: float64(box.Confidence),
		})
	}

	return blocs, largeurImage, hauteurImage, nil
}

// TesseractDisponible vérifie si Tesseract est installé et fonctionnel
func TesseractDisponible() bool {
	tesseractMu.Lock()
	defer tesseractMu.Unlock()

	client := gosseract.NewClient()
	defer client.Close()
	// Si on peut créer un client sans erreur, Tesseract est disponible
	return client != nil
}
