package ocr

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/popul/revisemieux/internal/domain/chapter"
)

func goldenFilePath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "testdata", "ocr", name)
}

func TestParseGeminiResponse_Golden(t *testing.T) {
	data, err := os.ReadFile(goldenFilePath("gemini_response_golden.json"))
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	result, err := parseOCRResponse(string(data))
	if err != nil {
		t.Fatalf("parseOCRResponse() error: %v", err)
	}

	if len(result.Blocks) != 5 {
		t.Fatalf("expected 5 blocks, got %d", len(result.Blocks))
	}

	// Verify block types
	expectedTypes := []chapter.BlockType{
		chapter.BlockText,   // TEXT
		chapter.BlockText,   // TEXT
		chapter.BlockText,   // TEXT
		chapter.BlockSchema, // DIAGRAM -> SCHEMA
		chapter.BlockTable,  // TABLE
	}
	for i, bt := range expectedTypes {
		if result.Blocks[i].BlockType != bt {
			t.Errorf("block[%d].BlockType = %q, want %q", i, result.Blocks[i].BlockType, bt)
		}
	}

	// Verify first block content
	if result.Blocks[0].Text != "Chapitre 3 : Le theoreme de Pythagore" {
		t.Errorf("block[0].Text = %q, want title text", result.Blocks[0].Text)
	}
	if result.Blocks[0].Confidence < 0.9 {
		t.Errorf("block[0].Confidence = %f, want >= 0.9", result.Blocks[0].Confidence)
	}

	// Verify DIAGRAM -> SCHEMA mapping
	if result.Blocks[3].BlockType != chapter.BlockSchema {
		t.Errorf("block[3] DIAGRAM should map to BlockSchema, got %q", result.Blocks[3].BlockType)
	}
	if result.Blocks[3].Text != "Triangle rectangle avec cotes a, b, c" {
		t.Errorf("block[3].Text = %q, want diagram text", result.Blocks[3].Text)
	}

	// Verify confidence values are preserved
	if result.Blocks[4].Confidence < 0.87 || result.Blocks[4].Confidence > 0.89 {
		t.Errorf("block[4].Confidence = %f, want ~0.88", result.Blocks[4].Confidence)
	}
}

func TestParseGeminiResponse_EmptyBlocks(t *testing.T) {
	result, err := parseOCRResponse(`{"blocks": []}`)
	if err != nil {
		t.Fatalf("parseOCRResponse() error: %v", err)
	}
	if len(result.Blocks) != 0 {
		t.Errorf("expected 0 blocks, got %d", len(result.Blocks))
	}
}

func TestParseGeminiResponse_InvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"completely invalid", "not json at all"},
		{"empty string", ""},
		{"partial JSON", `{"blocks": [`},
		{"wrong structure", `{"items": [{"text": "hello"}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseOCRResponse(tt.raw)
			// Either an error or empty result is acceptable for malformed input
			if err == nil && result != nil && len(result.Blocks) > 0 {
				t.Error("expected error or empty result for invalid JSON")
			}
		})
	}
}

func TestMapBlockType(t *testing.T) {
	// Test the mapping logic used in parseOCRResponse:
	// Valid BlockTypes pass through, "DIAGRAM" maps to SCHEMA, unknown maps to TEXT
	tests := []struct {
		input  string
		expect chapter.BlockType
	}{
		{"TEXT", chapter.BlockText},
		{"PHOTO", chapter.BlockPhoto},
		{"SCHEMA", chapter.BlockSchema},
		{"TABLE", chapter.BlockTable},
		{"MAP", chapter.BlockMap},
		{"GRAPH", chapter.BlockGraph},
		{"CIRCUIT", chapter.BlockCircuit},
		{"DECORATIVE", chapter.BlockDecorative},
		// Non-standard types
		{"DIAGRAM", chapter.BlockSchema},  // Explicit mapping in parseOCRResponse
		{"FORMULA", chapter.BlockText},    // Unknown -> TEXT fallback
		{"TITLE", chapter.BlockText},      // Unknown -> TEXT fallback
		{"unknown", chapter.BlockText},    // Unknown -> TEXT fallback
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Build a minimal JSON with this block type and parse it
			input := `{"blocks": [{"text": "test", "block_type": "` + tt.input + `", "confidence": 0.9}]}`
			result, err := parseOCRResponse(input)
			if err != nil {
				t.Fatalf("parseOCRResponse() error: %v", err)
			}
			if len(result.Blocks) != 1 {
				t.Fatalf("expected 1 block, got %d", len(result.Blocks))
			}
			if result.Blocks[0].BlockType != tt.expect {
				t.Errorf("MapBlockType(%q) = %q, want %q", tt.input, result.Blocks[0].BlockType, tt.expect)
			}
		})
	}
}

func TestEncodeBase64(t *testing.T) {
	input := []byte("hello world")
	got := encodeBase64(input)
	if got != "aGVsbG8gd29ybGQ=" {
		t.Errorf("encodeBase64() = %q, want %q", got, "aGVsbG8gd29ybGQ=")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		max    int
		expect string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is a ..."},
		{"", 5, ""},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.expect {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.expect)
		}
	}
}
