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
	openAIBaseURL       = "https://api.openai.com/v1"
	openAIModeleChat    = "gpt-4o"
	openAIModeleVision  = "gpt-4o"
	openAITimeoutDefaut = 60 * time.Second
)

// ClientOpenAI implémente AdaptateurLLM pour OpenAI
type ClientOpenAI struct {
	cleAPI     string
	httpClient *http.Client
}

// NouveauClientOpenAI crée un nouveau client OpenAI
func NouveauClientOpenAI(cleAPI string) *ClientOpenAI {
	return &ClientOpenAI{
		cleAPI: cleAPI,
		httpClient: &http.Client{
			Timeout: openAITimeoutDefaut,
		},
	}
}

// Nom retourne le nom du fournisseur
func (c *ClientOpenAI) Nom() string {
	return "OpenAI"
}

// EstDisponible vérifie si le service OpenAI est accessible
func (c *ClientOpenAI) EstDisponible(ctx context.Context) bool {
	if c.cleAPI == "" {
		return false
	}

	// Vérifie avec un appel minimal à l'API models
	req, err := http.NewRequestWithContext(ctx, "GET", openAIBaseURL+"/models", nil)
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
func (c *ClientOpenAI) GenererTexte(ctx context.Context, prompt string, options OptionsGeneration) (string, error) {
	messages := []messageChat{
		{Role: "user", Content: prompt},
	}

	if options.SystemPrompt != "" {
		messages = []messageChat{
			{Role: "system", Content: options.SystemPrompt},
			{Role: "user", Content: prompt},
		}
	}

	requete := requeteChatCompletion{
		Model:       openAIModeleChat,
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
			Fournisseur: "OpenAI",
			Code:        500,
			Message:     "aucune réponse générée",
			Recuperable: true,
		}
	}

	return reponse.Choices[0].Message.Content, nil
}

// GenererJSON génère une réponse JSON structurée
func (c *ClientOpenAI) GenererJSON(ctx context.Context, prompt string, schema interface{}, options OptionsGeneration) ([]byte, error) {
	messages := []messageChat{
		{Role: "user", Content: prompt},
	}

	if options.SystemPrompt != "" {
		messages = []messageChat{
			{Role: "system", Content: options.SystemPrompt},
			{Role: "user", Content: prompt},
		}
	}

	requete := requeteChatCompletion{
		Model:          openAIModeleChat,
		Messages:       messages,
		Temperature:    options.Temperature,
		MaxTokens:      options.MaxTokens,
		ResponseFormat: &formatReponse{Type: "json_object"},
	}

	reponse, err := c.appelChatCompletion(ctx, requete)
	if err != nil {
		return nil, err
	}

	if len(reponse.Choices) == 0 {
		return nil, &ErreurLLM{
			Fournisseur: "OpenAI",
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
			Fournisseur: "OpenAI",
			Code:        500,
			Message:     "réponse JSON invalide: " + err.Error(),
			Recuperable: true,
		}
	}

	return []byte(contenu), nil
}

// ExtraireTexteImage extrait le texte d'une image via l'API Vision
func (c *ClientOpenAI) ExtraireTexteImage(ctx context.Context, image []byte, options OptionsOCR) (*ResultatOCR, error) {
	// Encode l'image en base64
	imageBase64 := base64.StdEncoding.EncodeToString(image)

	// Détecte le type MIME (simplifié)
	mimeType := detecterMimeType(image)

	// Construit le prompt OCR
	promptOCR := construirePromptOCR(options)

	// Prépare le message avec l'image
	contenuMulti := []contenuMessage{
		{
			Type: "text",
			Text: promptOCR,
		},
		{
			Type: "image_url",
			ImageURL: &imageURL{
				URL:    fmt.Sprintf("data:%s;base64,%s", mimeType, imageBase64),
				Detail: "high",
			},
		},
	}

	requete := requeteChatCompletionVision{
		Model: openAIModeleVision,
		Messages: []messageChatVision{
			{
				Role:    "user",
				Content: contenuMulti,
			},
		},
		Temperature:    0.1, // Basse température pour l'OCR (plus déterministe)
		MaxTokens:      8000,
		ResponseFormat: &formatReponse{Type: "json_object"},
	}

	reponse, err := c.appelChatCompletionVision(ctx, requete)
	if err != nil {
		return nil, err
	}

	if len(reponse.Choices) == 0 {
		return nil, &ErreurLLM{
			Fournisseur: "OpenAI",
			Code:        500,
			Message:     "aucune réponse générée pour l'OCR",
			Recuperable: true,
		}
	}

	// Parse la réponse JSON
	return parserReponseOCR(reponse.Choices[0].Message.Content)
}

// appelChatCompletion effectue un appel à l'API chat/completions
func (c *ClientOpenAI) appelChatCompletion(ctx context.Context, requete requeteChatCompletion) (*reponseChatCompletion, error) {
	body, err := json.Marshal(requete)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "OpenAI",
			Code:        500,
			Message:     "erreur de sérialisation: " + err.Error(),
			Recuperable: false,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openAIBaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "OpenAI",
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
			Fournisseur: "OpenAI",
			Code:        500,
			Message:     "erreur réseau: " + err.Error(),
			Recuperable: true,
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "OpenAI",
			Code:        500,
			Message:     "erreur de lecture de réponse: " + err.Error(),
			Recuperable: true,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.gererErreurHTTP(resp.StatusCode, respBody)
	}

	var reponse reponseChatCompletion
	if err := json.Unmarshal(respBody, &reponse); err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "OpenAI",
			Code:        500,
			Message:     "erreur de parsing de réponse: " + err.Error(),
			Recuperable: false,
		}
	}

	return &reponse, nil
}

// appelChatCompletionVision effectue un appel à l'API chat/completions avec vision
func (c *ClientOpenAI) appelChatCompletionVision(ctx context.Context, requete requeteChatCompletionVision) (*reponseChatCompletion, error) {
	body, err := json.Marshal(requete)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "OpenAI",
			Code:        500,
			Message:     "erreur de sérialisation: " + err.Error(),
			Recuperable: false,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openAIBaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "OpenAI",
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
			Fournisseur: "OpenAI",
			Code:        500,
			Message:     "erreur réseau: " + err.Error(),
			Recuperable: true,
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "OpenAI",
			Code:        500,
			Message:     "erreur de lecture de réponse: " + err.Error(),
			Recuperable: true,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.gererErreurHTTP(resp.StatusCode, respBody)
	}

	var reponse reponseChatCompletion
	if err := json.Unmarshal(respBody, &reponse); err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "OpenAI",
			Code:        500,
			Message:     "erreur de parsing de réponse: " + err.Error(),
			Recuperable: false,
		}
	}

	return &reponse, nil
}

// gererErreurHTTP convertit une erreur HTTP en ErreurLLM
func (c *ClientOpenAI) gererErreurHTTP(statusCode int, body []byte) *ErreurLLM {
	var errReponse erreurOpenAI
	json.Unmarshal(body, &errReponse)

	message := errReponse.Error.Message
	if message == "" {
		message = fmt.Sprintf("erreur HTTP %d", statusCode)
	}

	erreur := &ErreurLLM{
		Fournisseur: "OpenAI",
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

// Types pour les requêtes/réponses OpenAI

type messageChat struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type requeteChatCompletion struct {
	Model          string         `json:"model"`
	Messages       []messageChat  `json:"messages"`
	Temperature    float64        `json:"temperature,omitempty"`
	MaxTokens      int            `json:"max_tokens,omitempty"`
	ResponseFormat *formatReponse `json:"response_format,omitempty"`
}

type formatReponse struct {
	Type string `json:"type"`
}

type reponseChatCompletion struct {
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

// Types pour Vision API

type contenuMessage struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type messageChatVision struct {
	Role    string           `json:"role"`
	Content []contenuMessage `json:"content"`
}

type requeteChatCompletionVision struct {
	Model          string              `json:"model"`
	Messages       []messageChatVision `json:"messages"`
	Temperature    float64             `json:"temperature,omitempty"`
	MaxTokens      int                 `json:"max_tokens,omitempty"`
	ResponseFormat *formatReponse      `json:"response_format,omitempty"`
}

type erreurOpenAI struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Fonctions utilitaires

// detecterMimeType détecte le type MIME d'une image à partir de ses premiers octets
func detecterMimeType(data []byte) string {
	if len(data) < 4 {
		return "application/octet-stream"
	}

	// PNG: 89 50 4E 47
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}

	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}

	// WebP: RIFF....WEBP
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "image/webp"
	}

	// GIF: GIF87a ou GIF89a
	if len(data) >= 6 && string(data[0:3]) == "GIF" {
		return "image/gif"
	}

	return "application/octet-stream"
}

// construirePromptOCR construit le prompt pour l'extraction OCR
func construirePromptOCR(options OptionsOCR) string {
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

	if options.DetailConfiance {
		promptBase += `
Pour chaque zone où tu n'es pas certain du texte (écriture illisible, mots flous, taches, etc.),
indique-le dans le champ "zones_incertaines".

Réponds uniquement en JSON avec ce format exact:
{
  "texte": "Le texte extrait complet",
  "confiance": 0.95,
  "zones_incertaines": [
    {
      "debut": 45,
      "fin": 52,
      "texte": "mot probable",
      "raison": "écriture peu lisible"
    }
  ],
  "blocs_texte": [
    {
      "texte": "Contenu du bloc",
      "position": {"x": 5, "y": 10, "largeur": 90, "hauteur": 8},
      "confiance": 0.95
    }
  ]
}

- "confiance" est un score entre 0 et 1 représentant ta confiance globale dans l'extraction
- "debut" et "fin" sont les indices de caractères dans le texte extrait
- Si tout est clair, retourne une liste vide pour "zones_incertaines"
- "blocs_texte" contient chaque bloc de texte distinct avec sa position estimée et sa confiance
- "position": x, y = coin supérieur gauche en % de l'image (0-100), largeur et hauteur en %
`
	} else {
		promptBase += `
Réponds uniquement en JSON avec ce format exact:
{
  "texte": "Le texte extrait complet",
  "confiance": 0.95,
  "zones_incertaines": [],
  "blocs_texte": [
    {
      "texte": "Contenu du bloc",
      "position": {"x": 5, "y": 10, "largeur": 90, "hauteur": 8},
      "confiance": 0.95
    }
  ]
}

- "confiance" est un score entre 0 et 1 représentant ta confiance globale dans l'extraction
- "blocs_texte" contient chaque bloc de texte distinct avec sa position estimée et sa confiance
- "position": x, y = coin supérieur gauche en % de l'image (0-100), largeur et hauteur en %
`
	}

	return promptBase
}

// parserReponseOCR parse la réponse JSON de l'OCR
func parserReponseOCR(contenu string) (*ResultatOCR, error) {
	var resultat ResultatOCR
	if err := json.Unmarshal([]byte(contenu), &resultat); err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "OpenAI",
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

// Vérifie que ClientOpenAI implémente AdaptateurLLM
var _ AdaptateurLLM = (*ClientOpenAI)(nil)
