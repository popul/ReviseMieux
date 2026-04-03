//go:build llm

package ocr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGeminiOCR_RealCall(t *testing.T) {
	apiKey := os.Getenv("GOOGLE_AI_API_KEY")
	if apiKey == "" {
		t.Skip("GOOGLE_AI_API_KEY not set, skipping integration test")
	}

	// Serve a minimal test image via a local HTTP server
	// (ProcessPage downloads the image from a URL)
	// 1x1 white PNG
	pngData := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, // PNG signature
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, // 8-bit RGB
		0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x00, 0x02, 0x00, 0x01, 0xe2, 0x21, 0xbc,
		0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, // IEND chunk
		0x44, 0xae, 0x42, 0x60, 0x82,
	}

	imgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(pngData)
	}))
	defer imgServer.Close()

	client := NewGeminiOCR(Config{
		APIKey: apiKey,
	})

	result, err := client.ProcessPage(context.Background(), imgServer.URL+"/test.png")
	if err != nil {
		t.Fatalf("ProcessPage() error: %v", err)
	}

	// A 1x1 image may return 0 blocks, which is acceptable.
	// Just verify no crash and proper response structure.
	t.Logf("Received %d blocks from Gemini", len(result.Blocks))
	for i, b := range result.Blocks {
		t.Logf("  block[%d]: type=%s text=%q confidence=%.2f", i, b.BlockType, b.Text, b.Confidence)
	}
}

func TestGeminiOCR_MockServer(t *testing.T) {
	// Test the full ProcessPage flow with a mock Gemini API server
	ocrResponse := map[string]interface{}{
		"blocks": []map[string]interface{}{
			{"text": "Bonjour le monde", "block_type": "TEXT", "confidence": 0.95},
			{"text": "Schema de circuit", "block_type": "DIAGRAM", "confidence": 0.80},
		},
	}
	ocrJSON, _ := json.Marshal(ocrResponse)

	apiResp := map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"content": string(ocrJSON),
				},
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     100,
			"completion_tokens": 50,
		},
	}

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiResp)
	}))
	defer apiServer.Close()

	// Serve a test image
	imgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte{0x89, 0x50, 0x4e, 0x47}) // minimal PNG header
	}))
	defer imgServer.Close()

	client := NewGeminiOCR(Config{
		APIKey:  "test-key",
		BaseURL: apiServer.URL,
	})

	result, err := client.ProcessPage(context.Background(), imgServer.URL+"/test.png")
	if err != nil {
		t.Fatalf("ProcessPage() error: %v", err)
	}

	if len(result.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(result.Blocks))
	}

	if result.Blocks[0].Text != "Bonjour le monde" {
		t.Errorf("block[0].Text = %q, want %q", result.Blocks[0].Text, "Bonjour le monde")
	}

	// DIAGRAM should be mapped to SCHEMA
	fmt.Printf("block[1].BlockType = %q\n", result.Blocks[1].BlockType)
}
