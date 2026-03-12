package anthropic

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdkanthro "github.com/anthropics/anthropic-sdk-go"
	"github.com/popul/revisemieux/internal/benchmark"
)

// OCRBenchmarkProvider implements benchmark.OCRProvider using the Anthropic vision API.
type OCRBenchmarkProvider struct {
	client   sdkanthro.Client
	name     string
	model    string
	priceIn  float64
	priceOut float64
}

// NewOCRBenchmarkProvider creates an OCR benchmark provider for a specific Anthropic model.
func NewOCRBenchmarkProvider(apiKey, model string, priceIn, priceOut float64) *OCRBenchmarkProvider {
	return &OCRBenchmarkProvider{
		client:   newClient(apiKey),
		name:     "Anthropic",
		model:    model,
		priceIn:  priceIn,
		priceOut: priceOut,
	}
}

func (p *OCRBenchmarkProvider) Name() string            { return p.name }
func (p *OCRBenchmarkProvider) ModelID() string          { return p.model }
func (p *OCRBenchmarkProvider) PricePerMInput() float64  { return p.priceIn }
func (p *OCRBenchmarkProvider) PricePerMOutput() float64 { return p.priceOut }

// ExtractBlocks sends images to the Anthropic vision API and returns OCR blocks.
func (p *OCRBenchmarkProvider) ExtractBlocks(ctx context.Context, imagePaths []string, subject string) (*benchmark.Response, error) {
	// Build content blocks: one image block per image + text instruction
	var contentBlocks []sdkanthro.ContentBlockParamUnion
	for _, imgPath := range imagePaths {
		data, err := os.ReadFile(imgPath)
		if err != nil {
			return nil, fmt.Errorf("anthropic ocr: read image %s: %w", imgPath, err)
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		mediaType := detectMediaType(imgPath)
		contentBlocks = append(contentBlocks, sdkanthro.NewImageBlockBase64(mediaType, encoded))
	}
	contentBlocks = append(contentBlocks, sdkanthro.NewTextBlock(
		fmt.Sprintf("Matière : %s\n\nExtrait tous les blocs de texte visibles sur ces photos de cahier.", subject),
	))

	start := time.Now()

	msg, err := p.client.Messages.New(ctx, sdkanthro.MessageNewParams{
		Model:     sdkanthro.Model(p.model),
		MaxTokens: 8192,
		System: []sdkanthro.TextBlockParam{
			{Text: ocrSystemPrompt()},
		},
		Messages: []sdkanthro.MessageParam{
			sdkanthro.NewUserMessage(contentBlocks...),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic ocr: %w", err)
	}

	latency := time.Since(start).Milliseconds()

	text := extractText(msg)
	return &benchmark.Response{
		RawJSON:      []byte(text),
		TokensInput:  int(msg.Usage.InputTokens),
		TokensOutput: int(msg.Usage.OutputTokens),
		LatencyMs:    latency,
		ModelVersion: p.model,
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
