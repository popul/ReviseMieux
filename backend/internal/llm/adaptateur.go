// Package llm fournit les adaptateurs pour les services LLM (OpenAI, Mistral)
package llm

import (
	"context"
)

// AdaptateurLLM définit l'interface commune pour les fournisseurs LLM
type AdaptateurLLM interface {
	// GenererTexte génère du texte à partir d'un prompt
	GenererTexte(ctx context.Context, prompt string, options OptionsGeneration) (string, error)

	// GenererJSON génère une réponse JSON structurée
	GenererJSON(ctx context.Context, prompt string, schema interface{}, options OptionsGeneration) ([]byte, error)

	// ExtraireTexteImage extrait le texte d'une image (OCR)
	ExtraireTexteImage(ctx context.Context, image []byte, options OptionsOCR) (*ResultatOCR, error)

	// DetecterOrientation détecte si une image est pivotée et retourne
	// l'angle de rotation nécessaire (0, 90, 180, 270)
	DetecterOrientation(ctx context.Context, image []byte) (int, error)

	// EstDisponible vérifie si le service est accessible
	EstDisponible(ctx context.Context) bool

	// Nom retourne le nom du fournisseur (pour les logs)
	Nom() string
}

// OptionsGeneration contient les options pour la génération de texte
type OptionsGeneration struct {
	// Modele permet de surcharger le modèle par défaut du fournisseur (ex: "gpt-4o-mini")
	Modele string

	// Temperature contrôle la créativité (0.0 = déterministe, 1.0 = créatif)
	Temperature float64

	// MaxTokens limite le nombre de tokens dans la réponse
	MaxTokens int

	// FormatReponse indique le format souhaité ("text" ou "json")
	FormatReponse string

	// SystemPrompt est le prompt système optionnel
	SystemPrompt string
}

// OptionsOCR contient les options pour l'extraction de texte d'image
type OptionsOCR struct {
	// Langue indique la langue attendue du texte (ex: "fr", "en")
	Langue string

	// DetailConfiance active le calcul détaillé de la confiance par zone
	DetailConfiance bool

	// TypeDocument indique le type de document ("manuscrit", "imprime", "mixte")
	TypeDocument string
}

// ResultatOCR contient le résultat de l'extraction OCR
type ResultatOCR struct {
	// Texte est le texte extrait de l'image
	Texte string `json:"texte"`

	// Confiance est le score de confiance global (0.0 à 1.0)
	Confiance float64 `json:"confiance"`

	// ZonesIncertaines liste les zones où le texte est incertain
	ZonesIncertaines []ZoneIncertaine `json:"zones_incertaines"`

	// BlocsTexte contient les blocs de texte avec leurs positions approximatives dans l'image
	BlocsTexte []BlocTexte `json:"blocs_texte,omitempty"`

	// TitreSuggere est le titre déduit par le LLM lors de l'OCR
	TitreSuggere string `json:"titre_suggere,omitempty"`

	// MatiereSuggeree est la matière scolaire déduite par le LLM lors de l'OCR
	MatiereSuggeree string `json:"matiere_suggeree,omitempty"`

	// Rotation est le nombre de degrés de rotation horaire nécessaire pour remettre l'image à l'endroit (0, 90, 180, 270)
	Rotation int `json:"rotation"`
}

// BlocTexte représente un bloc de texte avec sa position dans l'image
type BlocTexte struct {
	// Texte est le contenu textuel du bloc
	Texte string `json:"texte"`

	// Position est la position approximative du bloc dans l'image (en pourcentages 0-100)
	Position PositionBloc `json:"position"`

	// Confiance est le score de confiance pour ce bloc spécifique (0.0 à 1.0)
	Confiance float64 `json:"confiance"`
}

// PositionBloc représente la position d'un bloc de texte en pourcentages de l'image
type PositionBloc struct {
	// X est la position horizontale du coin supérieur gauche (0-100)
	X float64 `json:"x"`

	// Y est la position verticale du coin supérieur gauche (0-100)
	Y float64 `json:"y"`

	// Largeur est la largeur du bloc (0-100)
	Largeur float64 `json:"largeur"`

	// Hauteur est la hauteur du bloc (0-100)
	Hauteur float64 `json:"hauteur"`
}

// ZoneIncertaine représente une zone de texte où l'OCR est incertain
type ZoneIncertaine struct {
	// Debut est l'index de début dans le texte extrait
	Debut int `json:"debut"`

	// Fin est l'index de fin dans le texte extrait
	Fin int `json:"fin"`

	// Texte est le texte probable dans cette zone
	Texte string `json:"texte"`

	// Raison explique pourquoi cette zone est incertaine
	Raison string `json:"raison"`
}

// ErreurLLM représente une erreur retournée par un fournisseur LLM
type ErreurLLM struct {
	// Fournisseur est le nom du fournisseur (OpenAI, Mistral)
	Fournisseur string

	// Code est le code d'erreur HTTP ou interne
	Code int

	// Message est le message d'erreur
	Message string

	// Recuperable indique si l'erreur peut être récupérée par un retry ou fallback
	Recuperable bool

	// RateLimited indique si c'est une erreur de rate limiting
	RateLimited bool
}

// Error implémente l'interface error
func (e *ErreurLLM) Error() string {
	return e.Fournisseur + ": " + e.Message
}

// OptionsGenerationDefaut retourne les options par défaut pour la génération
func OptionsGenerationDefaut() OptionsGeneration {
	return OptionsGeneration{
		Temperature:   0.7,
		MaxTokens:     2000,
		FormatReponse: "text",
	}
}

// OptionsOCRDefaut retourne les options par défaut pour l'OCR
func OptionsOCRDefaut() OptionsOCR {
	return OptionsOCR{
		Langue:          "fr",
		DetailConfiance: true,
		TypeDocument:    "mixte",
	}
}
