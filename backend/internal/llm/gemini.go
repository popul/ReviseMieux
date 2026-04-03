// Package llm fournit les adaptateurs pour les services LLM
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
	// Gemini utilise l'endpoint compatible OpenAI
	geminiBaseURL       = "https://generativelanguage.googleapis.com/v1beta/openai"
	geminiModeleChat    = "gemini-2.5-flash-preview-05-20"
	geminiModeleVision  = "gemini-2.5-flash-preview-05-20"
	geminiTimeoutDefaut = 90 * time.Second
)

// ClientGemini implemente AdaptateurLLM pour Google Gemini via l'endpoint compatible OpenAI
type ClientGemini struct {
	cleAPI     string
	baseURL    string
	httpClient *http.Client
}

// NouveauClientGemini cree un nouveau client Gemini
func NouveauClientGemini(cleAPI string) *ClientGemini {
	return &ClientGemini{
		cleAPI:  cleAPI,
		baseURL: geminiBaseURL,
		httpClient: &http.Client{
			Timeout: geminiTimeoutDefaut,
		},
	}
}

// NouveauClientGeminiAvecURL cree un client Gemini avec une URL de base personnalisee
func NouveauClientGeminiAvecURL(cleAPI, baseURL string) *ClientGemini {
	c := NouveauClientGemini(cleAPI)
	if baseURL != "" {
		c.baseURL = baseURL
	}
	return c
}

// Nom retourne le nom du fournisseur
func (c *ClientGemini) Nom() string {
	return "Gemini"
}

// EstDisponible verifie si le service Gemini est accessible
func (c *ClientGemini) EstDisponible(ctx context.Context) bool {
	if c.cleAPI == "" {
		return false
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/models", nil)
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

// GenererTexte genere du texte a partir d'un prompt
func (c *ClientGemini) GenererTexte(ctx context.Context, prompt string, options OptionsGeneration) (string, error) {
	messages := []messageChat{
		{Role: "user", Content: prompt},
	}

	if options.SystemPrompt != "" {
		messages = []messageChat{
			{Role: "system", Content: options.SystemPrompt},
			{Role: "user", Content: prompt},
		}
	}

	modele := geminiModeleChat
	if options.Modele != "" {
		modele = options.Modele
	}

	requete := requeteChatCompletion{
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
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "aucune reponse generee",
			Recuperable: true,
		}
	}

	return reponse.Choices[0].Message.Content, nil
}

// GenererJSON genere une reponse JSON structuree
func (c *ClientGemini) GenererJSON(ctx context.Context, prompt string, schema interface{}, options OptionsGeneration) ([]byte, error) {
	messages := []messageChat{
		{Role: "user", Content: prompt},
	}

	if options.SystemPrompt != "" {
		messages = []messageChat{
			{Role: "system", Content: options.SystemPrompt},
			{Role: "user", Content: prompt},
		}
	}

	modele := geminiModeleChat
	if options.Modele != "" {
		modele = options.Modele
	}

	requete := requeteChatCompletion{
		Model:          modele,
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
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "aucune reponse generee",
			Recuperable: true,
		}
	}

	contenu := reponse.Choices[0].Message.Content

	var jsonValide json.RawMessage
	if err := json.Unmarshal([]byte(contenu), &jsonValide); err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "reponse JSON invalide: " + err.Error(),
			Recuperable: true,
		}
	}

	return []byte(contenu), nil
}

// ExtraireTexteImage extrait le texte d'une image via Gemini Vision (VLM direct)
func (c *ClientGemini) ExtraireTexteImage(ctx context.Context, image []byte, options OptionsOCR) (*ResultatOCR, error) {
	imageBase64 := base64.StdEncoding.EncodeToString(image)
	mimeType := detecterMimeType(image)

	promptOCR := construirePromptOCR(options)

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

	modele := geminiModeleVision
	requete := requeteChatCompletionVision{
		Model: modele,
		Messages: []messageChatVision{
			{
				Role:    "user",
				Content: contenuMulti,
			},
		},
		Temperature:    0.1,
		MaxTokens:      8000,
		ResponseFormat: &formatReponse{Type: "json_object"},
	}

	reponse, err := c.appelChatCompletionVision(ctx, requete)
	if err != nil {
		return nil, err
	}

	if len(reponse.Choices) == 0 {
		return nil, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "aucune reponse generee pour l'OCR",
			Recuperable: true,
		}
	}

	return parserReponseOCR(reponse.Choices[0].Message.Content)
}

// DetecterOrientation detecte si une image est pivotee et retourne l'angle de rotation
func (c *ClientGemini) DetecterOrientation(ctx context.Context, image []byte) (int, error) {
	imageBase64 := base64.StdEncoding.EncodeToString(image)
	mimeType := detecterMimeType(image)

	prompt := `Is the text in this image upright or rotated? Look at the physical orientation of characters relative to the image frame. Reply only in JSON: {"rotation": 0} if upright, {"rotation": 90}, {"rotation": 180} if upside down, {"rotation": 270}.`

	contenuMulti := []contenuMessage{
		{Type: "text", Text: prompt},
		{
			Type: "image_url",
			ImageURL: &imageURL{
				URL:    fmt.Sprintf("data:%s;base64,%s", mimeType, imageBase64),
				Detail: "auto",
			},
		},
	}

	requete := requeteChatCompletionVision{
		Model: geminiModeleVision,
		Messages: []messageChatVision{
			{Role: "user", Content: contenuMulti},
		},
		Temperature:    0,
		MaxTokens:      50,
		ResponseFormat: &formatReponse{Type: "json_object"},
	}

	reponse, err := c.appelChatCompletionVision(ctx, requete)
	if err != nil {
		return 0, err
	}

	if len(reponse.Choices) == 0 {
		return 0, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "aucune reponse pour la detection d'orientation",
			Recuperable: true,
		}
	}

	var result struct {
		Rotation int `json:"rotation"`
	}
	if err := json.Unmarshal([]byte(reponse.Choices[0].Message.Content), &result); err != nil {
		return 0, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "reponse orientation invalide: " + err.Error(),
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

// appelChatCompletion effectue un appel a l'API chat/completions compatible OpenAI
func (c *ClientGemini) appelChatCompletion(ctx context.Context, requete requeteChatCompletion) (*reponseChatCompletion, error) {
	body, err := json.Marshal(requete)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "erreur de serialisation: " + err.Error(),
			Recuperable: false,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "erreur de creation de requete: " + err.Error(),
			Recuperable: false,
		}
	}

	req.Header.Set("Authorization", "Bearer "+c.cleAPI)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "erreur reseau: " + err.Error(),
			Recuperable: true,
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "erreur de lecture de reponse: " + err.Error(),
			Recuperable: true,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.gererErreurHTTP(resp.StatusCode, respBody)
	}

	var reponse reponseChatCompletion
	if err := json.Unmarshal(respBody, &reponse); err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "erreur de parsing de reponse: " + err.Error(),
			Recuperable: false,
		}
	}

	return &reponse, nil
}

// appelChatCompletionVision effectue un appel vision via l'endpoint compatible OpenAI
func (c *ClientGemini) appelChatCompletionVision(ctx context.Context, requete requeteChatCompletionVision) (*reponseChatCompletion, error) {
	body, err := json.Marshal(requete)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "erreur de serialisation: " + err.Error(),
			Recuperable: false,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "erreur de creation de requete: " + err.Error(),
			Recuperable: false,
		}
	}

	req.Header.Set("Authorization", "Bearer "+c.cleAPI)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "erreur reseau: " + err.Error(),
			Recuperable: true,
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "erreur de lecture de reponse: " + err.Error(),
			Recuperable: true,
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.gererErreurHTTP(resp.StatusCode, respBody)
	}

	var reponse reponseChatCompletion
	if err := json.Unmarshal(respBody, &reponse); err != nil {
		return nil, &ErreurLLM{
			Fournisseur: "Gemini",
			Code:        500,
			Message:     "erreur de parsing de reponse: " + err.Error(),
			Recuperable: false,
		}
	}

	return &reponse, nil
}

// gererErreurHTTP convertit une erreur HTTP en ErreurLLM
func (c *ClientGemini) gererErreurHTTP(statusCode int, body []byte) *ErreurLLM {
	// Gemini utilise le format d'erreur compatible OpenAI
	var errReponse erreurOpenAI
	json.Unmarshal(body, &errReponse)

	message := errReponse.Error.Message
	if message == "" {
		message = fmt.Sprintf("erreur HTTP %d: %s", statusCode, string(body))
	}

	erreur := &ErreurLLM{
		Fournisseur: "Gemini",
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

// Verifie que ClientGemini implemente AdaptateurLLM
var _ AdaptateurLLM = (*ClientGemini)(nil)
