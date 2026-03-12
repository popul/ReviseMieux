package openaicompat

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/popul/revisemieux/internal/benchmark"
)

// OCRBenchmarkProvider implements benchmark.OCRProvider for OpenAI-compatible vision APIs.
type OCRBenchmarkProvider struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	name       string
	model      string
	priceIn    float64
	priceOut   float64
}

// NewOCRBenchmarkProvider creates an OCR benchmark provider for an OpenAI-compatible vision API.
func NewOCRBenchmarkProvider(cfg Config) *OCRBenchmarkProvider {
	return &OCRBenchmarkProvider{
		httpClient: &http.Client{Timeout: 180 * time.Second},
		baseURL:    cfg.BaseURL,
		apiKey:     cfg.APIKey,
		name:       cfg.Name,
		model:      cfg.Model,
		priceIn:    cfg.PriceIn,
		priceOut:   cfg.PriceOut,
	}
}

func (p *OCRBenchmarkProvider) Name() string            { return p.name }
func (p *OCRBenchmarkProvider) ModelID() string          { return p.model }
func (p *OCRBenchmarkProvider) PricePerMInput() float64  { return p.priceIn }
func (p *OCRBenchmarkProvider) PricePerMOutput() float64 { return p.priceOut }

// visionContent is a content part in a vision message (text or image_url).
type visionContent struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *visionImageURL `json:"image_url,omitempty"`
}

type visionImageURL struct {
	URL string `json:"url"`
}

// visionMessage is a chat message with mixed content (text + images).
type visionMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type visionRequest struct {
	Model       string          `json:"model"`
	Messages    []visionMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
}

// ExtractBlocks sends images to the vision API and returns OCR blocks.
func (p *OCRBenchmarkProvider) ExtractBlocks(ctx context.Context, imagePaths []string, subject string) (*benchmark.Response, error) {
	// Build user message content: images + text instruction
	var contentParts []visionContent
	for _, imgPath := range imagePaths {
		data, err := os.ReadFile(imgPath)
		if err != nil {
			return nil, fmt.Errorf("%s ocr: read image %s: %w", p.name, imgPath, err)
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		mediaType := detectMediaType(imgPath)
		contentParts = append(contentParts, visionContent{
			Type: "image_url",
			ImageURL: &visionImageURL{
				URL: fmt.Sprintf("data:%s;base64,%s", mediaType, encoded),
			},
		})
	}
	contentParts = append(contentParts, visionContent{
		Type: "text",
		Text: fmt.Sprintf("Matière : %s\n\nExtrait tous les blocs de texte visibles sur ces photos de cahier.", subject),
	})

	userContent, _ := json.Marshal(contentParts)

	systemContent, _ := json.Marshal(ocrSystemPrompt())

	temp := 0.0
	reqBody := visionRequest{
		Model: p.model,
		Messages: []visionMessage{
			{Role: "system", Content: systemContent},
			{Role: "user", Content: userContent},
		},
		MaxTokens:   8192,
		Temperature: &temp,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%s ocr marshal: %w", p.name, err)
	}

	url := p.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("%s ocr request: %w", p.name, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	start := time.Now()
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s ocr call: %w", p.name, err)
	}
	defer resp.Body.Close()
	latency := time.Since(start).Milliseconds()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s ocr read body: %w", p.name, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s ocr HTTP %d: %s", p.name, resp.StatusCode, truncate(string(respBody), 500))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("%s ocr parse response: %w", p.name, err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("%s ocr: no choices in response", p.name)
	}

	text := chatResp.Choices[0].Message.Content

	return &benchmark.Response{
		RawJSON:      []byte(text),
		TokensInput:  chatResp.Usage.PromptTokens,
		TokensOutput: chatResp.Usage.CompletionTokens,
		LatencyMs:    latency,
		ModelVersion: chatResp.Model,
	}, nil
}

func detectMediaType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

func ocrSystemPrompt() string {
	return `Tu es un système OCR spécialisé dans l'extraction de texte à partir de photos de cahiers de collégiens français.

Ta tâche : à partir de photos de cahier, extraire TOUS les blocs de texte visibles, qu'ils soient manuscrits ou imprimés.

## Règles

1. Extraire le texte FIDÈLEMENT tel qu'il apparaît, avec l'orthographe de l'élève (y compris les fautes).
2. Chaque bloc correspond à une section logique (titre, paragraphe, question, réponse, légende de schéma).
3. Décrire les schémas, graphiques et illustrations entre crochets [Schéma : description].
4. Attribuer un block_type : "TEXT" pour le texte, "DIAGRAM" pour les schémas/graphiques/photos.
5. Attribuer un score de confidence (0-1) reflétant la lisibilité du texte.
6. Confidence < 0.7 si le texte manuscrit est difficile à lire ou ambigu.

## Format de sortie

Réponds UNIQUEMENT avec un JSON valide, sans markdown, sans commentaire :

{
  "blocks": [
    {
      "text": "texte extrait fidèlement",
      "block_type": "TEXT",
      "confidence": 0.85
    }
  ]
}`
}
