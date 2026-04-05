package benchmark

import (
	"context"
	"testing"
)

func TestWordOverlap(t *testing.T) {
	tests := []struct {
		name     string
		golden   string
		produced string
		min      float64
		max      float64
	}{
		{
			name:     "identical texts",
			golden:   "La masse volumique est le rapport de la masse sur le volume",
			produced: "La masse volumique est le rapport de la masse sur le volume",
			min:      1.0,
			max:      1.0,
		},
		{
			name:     "partial match",
			golden:   "Les poils absorbants permettent l'absorption de l'eau et des sels minéraux",
			produced: "Les poils absorbants absorbent l'eau",
			min:      0.4,
			max:      0.9,
		},
		{
			name:     "no match",
			golden:   "Le théorème de Pythagore",
			produced: "Les cellules végétales contiennent des chloroplastes",
			min:      0.0,
			max:      0.1,
		},
		{
			name:     "empty golden",
			golden:   "",
			produced: "anything",
			min:      1.0,
			max:      1.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wordOverlap(tt.golden, tt.produced)
			if got < tt.min || got > tt.max {
				t.Errorf("wordOverlap() = %f, want in [%f, %f]", got, tt.min, tt.max)
			}
		})
	}
}

func TestMatchBlocks(t *testing.T) {
	golden := []OCRBlock{
		{Text: "Les poils absorbants absorbent l'eau et les sels minéraux", BlockType: "TEXT"},
		{Text: "Les stomates permettent les échanges gazeux", BlockType: "TEXT"},
	}
	produced := []OCRBlock{
		{Text: "Les stomates échanges gazeux atmosphère feuille", BlockType: "TEXT"},
		{Text: "Poils absorbants absorption eau sels minéraux racine", BlockType: "TEXT"},
		{Text: "Schéma de la racine", BlockType: "DIAGRAM"},
	}

	matches := matchBlocks(golden, produced)
	if len(matches) != 2 {
		t.Fatalf("matches = %d, want 2", len(matches))
	}
}

func TestEvaluateOCR(t *testing.T) {
	tc := TestCase{
		ID:      "test_ocr_01",
		Subject: "SVT",
		Blocks: []OCRBlock{
			{Text: "Les poils absorbants absorbent l'eau et les sels minéraux par la racine", BlockType: "TEXT", Confidence: 0.85},
			{Text: "Schéma de la racine avec poils absorbants", BlockType: "DIAGRAM", Confidence: 0.80},
		},
	}

	produced := []OCRBlock{
		{Text: "Les poils absorbants absorbent l'eau et les sels minéraux par la racine", BlockType: "TEXT", Confidence: 0.90},
		{Text: "Schéma de la racine avec poils absorbants", BlockType: "DIAGRAM", Confidence: 0.75},
	}

	prov := &mockOCRProvider{name: "test", model: "test-v1", priceIn: 3.0, priceOut: 15.0}
	resp := &Response{TokensInput: 5000, TokensOutput: 500, LatencyMs: 3000}

	result := EvaluateOCR(tc, produced, resp, prov)

	if result.DetectionScore < 0.9 {
		t.Errorf("detection = %f, want >= 0.9", result.DetectionScore)
	}
	if result.TextAccuracy < 0.9 {
		t.Errorf("text accuracy = %f, want >= 0.9", result.TextAccuracy)
	}
	if result.TypeAccuracy < 0.9 {
		t.Errorf("type accuracy = %f, want >= 0.9", result.TypeAccuracy)
	}
	if result.QualityScore < 0.9 {
		t.Errorf("quality = %f, want >= 0.9", result.QualityScore)
	}
	if result.CostUSD <= 0 {
		t.Error("cost should be > 0")
	}
}

func TestEvaluateOCR_NilProduced(t *testing.T) {
	tc := TestCase{
		ID:      "test_ocr_nil",
		Subject: "SVT",
		Blocks: []OCRBlock{
			{Text: "Some text", BlockType: "TEXT"},
		},
	}
	prov := &mockOCRProvider{name: "test", model: "test-v1", priceIn: 3.0, priceOut: 15.0}
	result := EvaluateOCR(tc, nil, nil, prov)

	if result.QualityScore != 0 {
		t.Errorf("quality should be 0 for nil produced, got %f", result.QualityScore)
	}
}

func TestSummarizeOCR(t *testing.T) {
	results := []OCREvalResult{
		{CaseID: "c1", DetectionScore: 1.0, TextAccuracy: 0.9, TypeAccuracy: 1.0, LatencyMs: 2000, CostUSD: 0.01, QualityScore: 0.95},
		{CaseID: "c2", DetectionScore: 0.8, TextAccuracy: 0.7, TypeAccuracy: 0.8, LatencyMs: 3000, CostUSD: 0.02, QualityScore: 0.75},
	}
	s := SummarizeOCR("test", "test-v1", results)

	if s.AvgDetectionScore != 0.9 {
		t.Errorf("avg detection = %f, want 0.9", s.AvgDetectionScore)
	}
	if s.TotalCostUSD != 0.03 {
		t.Errorf("total cost = %f, want 0.03", s.TotalCostUSD)
	}
}

func TestParseOCROutput(t *testing.T) {
	t.Run("wrapped format", func(t *testing.T) {
		raw := `{"blocks": [{"text": "hello", "block_type": "TEXT", "confidence": 0.9}]}`
		blocks, err := ParseOCROutput([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		if len(blocks) != 1 {
			t.Fatalf("got %d blocks, want 1", len(blocks))
		}
		if blocks[0].Text != "hello" {
			t.Errorf("text = %q, want %q", blocks[0].Text, "hello")
		}
	})

	t.Run("array format", func(t *testing.T) {
		raw := `[{"text": "hello", "block_type": "TEXT", "confidence": 0.9}]`
		blocks, err := ParseOCROutput([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		if len(blocks) != 1 {
			t.Fatalf("got %d blocks, want 1", len(blocks))
		}
	})
}

type mockOCRProvider struct {
	name     string
	model    string
	priceIn  float64
	priceOut float64
}

func (m *mockOCRProvider) Name() string             { return m.name }
func (m *mockOCRProvider) ModelID() string          { return m.model }
func (m *mockOCRProvider) PricePerMInput() float64  { return m.priceIn }
func (m *mockOCRProvider) PricePerMOutput() float64 { return m.priceOut }
func (m *mockOCRProvider) ExtractBlocks(_ context.Context, _ []string, _ string) (*Response, error) {
	return nil, nil
}
