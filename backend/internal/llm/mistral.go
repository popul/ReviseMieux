// Package llm fournit les adaptateurs pour les services LLM (OpenAI, Mistral)
package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	mistralBaseURL       = "https://api.mistral.ai/v1"
	mistralModeleChat    = "mistral-large-latest"
	mistralModeleVision  = "pixtral-large-latest"
	mistralTimeoutDefaut = 60 * time.Second
)

// ClientMistral implémente AdaptateurLLM pour Mistral AI
type ClientMistral struct {
	cleAPI     string
	httpClient *http.Client
}

// NouveauClientMistral crée un nouveau client Mistral
func NouveauClientMistral(cleAPI string) *ClientMistral {
	return &ClientMistral{
		cleAPI: cleAPI,
		httpClient: &http.Client{
			Timeout: mistralTimeoutDefaut,
		},
	}
}

// Nom retourne le nom du fournisseur
func (c *ClientMistral) Nom() string {
	return "Mistral"
}

// EstDisponible vérifie si le service Mistral est accessible
func (c *ClientMistral) EstDisponible(ctx context.Context) bool {
	if c.cleAPI == "" {
		return false
	}

	// Vérifie avec un appel minimal à l'API models
	req, err := http.NewRequestWithContext(ctx, "GET", mistralBaseURL+"/models", nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+c.cleAPI)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GenererTexte génère du texte à partir d'un prompt
func (c *ClientMistral) GenererTexte(ctx context.Context, prompt string, options OptionsGeneration) (string, error) {
	messages := []messageMistral{
		{Role: "user", Content: prompt},
	}

	if options.SystemPrompt != "" {
		messages = []messageMistral{
			{Role: "system", Content: options.SystemPrompt},
			{Role: "user", Content: prompt},
		}
	}

	modele := mistralModeleChat
	if options.Modele != "" {
		modele = options.Modele
	}

	requete := requeteMistralChat{
		Model:       modele,
		Messages:    messages,
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
	}

	reponse, err := c.appelChatCompletion(ctx, requete)
	if err != nil {
		return "", err
	}

	if len(reponse.Choices) == 0 {
		return "", &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "aucune réponse générée",
			Recuperable: true,
		}
	}

	return reponse.Choices[0].Message.Content, nil
}

// GenererJSON génère une réponse JSON structurée
func (c *ClientMistral) GenererJSON(ctx context.Context, prompt string, schema interface{}, options OptionsGeneration) ([]byte, error) {
	messages := []messageMistral{
		{Role: "user", Content: prompt},
	}

	if options.SystemPrompt != "" {
		messages = []messageMistral{
			{Role: "system", Content: options.SystemPrompt},
			{Role: "user", Content: prompt},
		}
	}

	modele := mistralModeleChat
	if options.Modele != "" {
		modele = options.Modele
	}

	requete := requeteMistralChat{
		Model:          modele,
		Messages:       messages,
		Temperature:    options.Temperature,
		MaxTokens:      options.MaxTokens,
		ResponseFormat: &formatReponseMistral{Type: "json_object"},
	}

	reponse, err := c.appelChatCompletion(ctx, requete)
	if err != nil {
		return nil, err
	}

	if len(reponse.Choices) == 0 {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "aucune réponse générée",
			Recuperable: true,
		}
	}

	contenu := reponse.Choices[0].Message.Content

	// Valide que c'est du JSON valide
	var jsonValide json.RawMessage
	if err := json.Unmarshal([]byte(contenu), &jsonValide); err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "réponse JSON invalide: " + err.Error(),
			Recuperable: true,
		}
	}

	return []byte(contenu), nil
}

// ExtraireTexteImage extrait le texte d'une image via Pixtral (modèle vision de Mistral)
func (c *ClientMistral) ExtraireTexteImage(ctx context.Context, image []byte, options OptionsOCR) (*ResultatOCR, error) {
	// Encode l'image en base64
	imageBase64 := base64.StdEncoding.EncodeToString(image)

	// Détecte le type MIME
	mimeType := detecterMimeType(image)

	// Construit le prompt OCR
	promptOCR := construirePromptOCRMistral(options)

	// Prépare le message avec l'image (format Mistral/Pixtral)
	contenuMulti := []contenuMistral{
		{
			Type: "text",
			Text: promptOCR,
		},
		{
			Type: "image_url",
			ImageURL: &imageURLMistral{
				URL: fmt.Sprintf("data:%s;base64,%s", mimeType, imageBase64),
			},
		},
	}

	requete := requeteMistralVision{
		Model: mistralModeleVision,
		Messages: []messageMistralVision{
			{
				Role:    "user",
				Content: contenuMulti,
			},
		},
		Temperature:    0.1, // Basse température pour l'OCR (plus déterministe)
		MaxTokens:      8000,
		ResponseFormat: &formatReponseMistral{Type: "json_object"},
	}

	reponse, err := c.appelChatCompletionVision(ctx, requete)
	if err != nil {
		return nil, err
	}

	if len(reponse.Choices) == 0 {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "aucune réponse générée pour l'OCR",
			Recuperable: true,
		}
	}

	// Parse la réponse JSON
	return parserReponseOCRMistral(reponse.Choices[0].Message.Content)
}

// DetecterOrientation détecte si une image est pivotée et retourne l'angle de rotation (0, 90, 180, 270)
func (c *ClientMistral) DetecterOrientation(ctx context.Context, image []byte) (int, error) {
	imageBase64 := base64.StdEncoding.EncodeToString(image)
	mimeType := detecterMimeType(image)

	prompt := `Is the text in this image upright or rotated? Look at the physical orientation of characters relative to the image frame. Reply only in JSON: {"rotation": 0} if upright, {"rotation": 90}, {"rotation": 180} if upside down, {"rotation": 270}.`

	contenuMulti := []contenuMistral{
		{Type: "text", Text: prompt},
		{
			Type: "image_url",
			ImageURL: &imageURLMistral{
				URL: fmt.Sprintf("data:%s;base64,%s", mimeType, imageBase64),
			},
		},
	}

	requete := requeteMistralVision{
		Model: mistralModeleVision,
		Messages: []messageMistralVision{
			{Role: "user", Content: contenuMulti},
		},
		Temperature:    0,
		MaxTokens:      50,
		ResponseFormat: &formatReponseMistral{Type: "json_object"},
	}

	reponse, err := c.appelChatCompletionVision(ctx, requete)
	if err != nil {
		return 0, err
	}

	if len(reponse.Choices) == 0 {
		return 0, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "aucune réponse pour la détection d'orientation",
			Recuperable: true,
		}
	}

	var result struct {
		Rotation int `json:"rotation"`
	}
	if err := json.Unmarshal([]byte(reponse.Choices[0].Message.Content), &result); err != nil {
		return 0, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "réponse orientation invalide: " + err.Error(),
			Recuperable: false,
		}
	}

	switch result.Rotation {
	case 0, 90, 180, 270:
		return result.Rotation, nil
	default:
		return 0, nil
	}
}

// appelChatCompletion effectue un appel à l'API chat/completions de Mistral
func (c *ClientMistral) appelChatCompletion(ctx context.Context, requete requeteMistralChat) (*reponseMistralChat, error) {
	body, err := json.Marshal(requete)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "erreur de sérialisation: " + err.Error(),
			Recuperable: false,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", mistralBaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "erreur de création de requête: " + err.Error(),
			Recuperable: false,
		}
	}

	req.Header.Set("Authorization", "Bearer "+c.cleAPI)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "erreur réseau: " + err.Error(),
			Recuperable: true,
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "erreur de lecture de réponse: " + err.Error(),
			Recuperable: true,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.gererErreurHTTP(resp.StatusCode, respBody)
	}

	var reponse reponseMistralChat
	if err := json.Unmarshal(respBody, &reponse); err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "erreur de parsing de réponse: " + err.Error(),
			Recuperable: false,
		}
	}

	return &reponse, nil
}

// appelChatCompletionVision effectue un appel à l'API chat/completions avec vision (Pixtral)
func (c *ClientMistral) appelChatCompletionVision(ctx context.Context, requete requeteMistralVision) (*reponseMistralChat, error) {
	body, err := json.Marshal(requete)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "erreur de sérialisation: " + err.Error(),
			Recuperable: false,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", mistralBaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "erreur de création de requête: " + err.Error(),
			Recuperable: false,
		}
	}

	req.Header.Set("Authorization", "Bearer "+c.cleAPI)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "erreur réseau: " + err.Error(),
			Recuperable: true,
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "erreur de lecture de réponse: " + err.Error(),
			Recuperable: true,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.gererErreurHTTP(resp.StatusCode, respBody)
	}

	var reponse reponseMistralChat
	if err := json.Unmarshal(respBody, &reponse); err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "erreur de parsing de réponse: " + err.Error(),
			Recuperable: false,
		}
	}

	return &reponse, nil
}

// gererErreurHTTP convertit une erreur HTTP en ErreurLLM
func (c *ClientMistral) gererErreurHTTP(statusCode int, body []byte) *ErreurLLM {
	var errReponse erreurMistral
	json.Unmarshal(body, &errReponse)

	message := errReponse.Message
	if message == "" {
		message = fmt.Sprintf("erreur HTTP %d", statusCode)
	}

	erreur := &ErreurLLM{
		Fournisseur: "Mistral",
		Code:        statusCode,
		Message:     message,
	}

	switch statusCode {
	case http.StatusTooManyRequests:
		erreur.RateLimited = true
		erreur.Recuperable = true
	case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
		erreur.Recuperable = true
	case http.StatusUnauthorized, http.StatusForbidden:
		erreur.Recuperable = false
	default:
		erreur.Recuperable = statusCode >= 500
	}

	return erreur
}

// Types pour les requêtes/réponses Mistral

type messageMistral struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type requeteMistralChat struct {
	Model          string                `json:"model"`
	Messages       []messageMistral      `json:"messages"`
	Temperature    float64               `json:"temperature,omitempty"`
	MaxTokens      int                   `json:"max_tokens,omitempty"`
	ResponseFormat *formatReponseMistral `json:"response_format,omitempty"`
}

type formatReponseMistral struct {
	Type string `json:"type"`
}

type reponseMistralChat struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Types pour Pixtral (Vision API)

type contenuMistral struct {
	Type     string           `json:"type"`
	Text     string           `json:"text,omitempty"`
	ImageURL *imageURLMistral `json:"image_url,omitempty"`
}

type imageURLMistral struct {
	URL string `json:"url"`
}

type messageMistralVision struct {
	Role    string           `json:"role"`
	Content []contenuMistral `json:"content"`
}

type requeteMistralVision struct {
	Model          string                 `json:"model"`
	Messages       []messageMistralVision `json:"messages"`
	Temperature    float64                `json:"temperature,omitempty"`
	MaxTokens      int                    `json:"max_tokens,omitempty"`
	ResponseFormat *formatReponseMistral  `json:"response_format,omitempty"`
}

type erreurMistral struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Fonctions utilitaires spécifiques à Mistral

// construirePromptOCRMistral construit le prompt pour l'extraction OCR avec Pixtral
func construirePromptOCRMistral(options OptionsOCR) string {
	langue := options.Langue
	if langue == "" {
		langue = "français"
	}

	typeDoc := "mixte (manuscrit et imprimé)"
	switch options.TypeDocument {
	case "manuscrit":
		typeDoc = "manuscrit (écriture à la main)"
	case "imprime":
		typeDoc = "imprimé (texte typographié)"
	}

	promptBase := fmt.Sprintf(`Tu es un expert en OCR (reconnaissance optique de caractères).
Analyse cette image qui contient du texte %s en %s.

Extrais tout le texte visible dans l'image en respectant la mise en page originale autant que possible.

Identifie les blocs de texte distincts (paragraphes, titres, lignes séparées) et retourne-les
dans le champ "blocs_texte". Pour chaque bloc, estime sa position dans l'image en pourcentages (0-100)
avec x, y pour le coin supérieur gauche et largeur, hauteur pour les dimensions.
`, typeDoc, langue)

	matieres := "mathematiques, francais, histoire, geographie, sciences, anglais, physique, chimie, svt, ses, philosophie, espagnol, allemand, italien, economie, informatique"

	if options.DetailConfiance {
		promptBase += fmt.Sprintf(`
Pour chaque zone où tu n'es pas certain du texte, indique-le dans "zones_incertaines".

Déduis aussi un titre court (max 50 caractères) et la matière scolaire parmi : %s.

Réponds uniquement en JSON :
{
  "texte": "Le texte extrait complet",
  "confiance": 0.95,
  "zones_incertaines": [
    {"debut": 45, "fin": 52, "texte": "mot probable", "raison": "écriture peu lisible"}
  ],
  "blocs_texte": [
    {"texte": "Contenu du bloc", "position": {"x": 5, "y": 10, "largeur": 90, "hauteur": 8}, "confiance": 0.95}
  ],
  "titre_suggere": "Titre court du cours",
  "matiere_suggeree": "mathematiques"
}
`, matieres)
	} else {
		promptBase += fmt.Sprintf(`
Déduis aussi un titre court (max 50 caractères) et la matière scolaire parmi : %s.

Réponds uniquement en JSON :
{
  "texte": "Le texte extrait complet",
  "confiance": 0.95,
  "zones_incertaines": [],
  "blocs_texte": [
    {"texte": "Contenu du bloc", "position": {"x": 5, "y": 10, "largeur": 90, "hauteur": 8}, "confiance": 0.95}
  ],
  "titre_suggere": "Titre court du cours",
  "matiere_suggeree": "mathematiques"
}
`, matieres)
	}

	return promptBase
}

// parserReponseOCRMistral parse la réponse JSON de l'OCR Mistral
func parserReponseOCRMistral(contenu string) (*ResultatOCR, error) {
	var resultat ResultatOCR
	if err := json.Unmarshal([]byte(contenu), &resultat); err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Mistral",
			Code:        500,
			Message:     "erreur de parsing de la réponse OCR: " + err.Error(),
			Recuperable: false,
		}
	}

	// Valide les indices des zones incertaines
	for i, zone := range resultat.ZonesIncertaines {
		if zone.Debut < 0 || zone.Fin < zone.Debut || zone.Fin > len(resultat.Texte) {
			// Corrige les indices invalides
			resultat.ZonesIncertaines[i].Debut = max(0, zone.Debut)
			resultat.ZonesIncertaines[i].Fin = min(len(resultat.Texte), zone.Fin)
		}
	}

	return &resultat, nil
}

// Vérifie que ClientMistral implémente AdaptateurLLM
var _ AdaptateurLLM = (*ClientMistral)(nil)
