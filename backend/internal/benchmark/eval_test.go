package benchmark

import (
	"context"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Densité", "densite"},
		{"MASSE VOLUMIQUE", "masse volumique"},
		{"  extra   spaces  ", "extra spaces"},
		{"ρ = m / V", "ρ = m / v"},
		{"L'unité SI", "l'unite si"},
	}
	for _, tt := range tests {
		got := normalize(tt.input)
		if got != tt.want {
			t.Errorf("normalize(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestJaccard(t *testing.T) {
	tests := []struct {
		a, b []string
		want float64
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}, 1.0},
		{[]string{"a", "b"}, []string{"b", "c"}, 1.0 / 3.0},
		{[]string{}, []string{}, 1.0},
		{[]string{"x"}, []string{"y"}, 0.0},
	}
	for _, tt := range tests {
		got := jaccard(tt.a, tt.b)
		if diff := got - tt.want; diff > 0.001 || diff < -0.001 {
			t.Errorf("jaccard(%v, %v) = %f, want %f", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestTermSimilarity(t *testing.T) {
	tests := []struct {
		a, b string
		min  float64
	}{
		{"densité", "densité", 1.0},
		{"masse volumique", "Masse Volumique", 1.0},
		{"la densité", "densité d'un corps", 0.2},
		{"xyz", "abc", 0.0},
	}
	for _, tt := range tests {
		got := termSimilarity(tt.a, tt.b)
		if got < tt.min {
			t.Errorf("termSimilarity(%q, %q) = %f, want >= %f", tt.a, tt.b, got, tt.min)
		}
	}
}

func TestIsTraceable(t *testing.T) {
	source := "La masse volumique est le rapport de la masse sur le volume. ρ = m / V. L'unité SI est le kg/m³."

	tests := []struct {
		term string
		want bool
	}{
		{"masse volumique", true},
		{"ρ = m / V", true},
		{"kg/m³", true},
		{"théorème de Pythagore", false}, // hallucination
		{"rapport masse volume", true},   // words present in source
	}
	for _, tt := range tests {
		got := isTraceable(tt.term, source)
		if got != tt.want {
			t.Errorf("isTraceable(%q) = %v, want %v", tt.term, got, tt.want)
		}
	}
}

func TestMatchItems(t *testing.T) {
	golden := []GoldenItem{
		{Type: "KNOWLEDGE", Term: "masse volumique"},
		{Type: "PROCEDURE", Term: "ρ = m / V"},
	}
	parsed := []ParsedItem{
		{Type: "KNOWLEDGE", Term: "La masse volumique"},
		{Type: "PROCEDURE", Term: "formule ρ = m / V"},
		{Type: "KNOWLEDGE", Term: "densité"},
	}
	matches := matchItems(golden, parsed)
	if len(matches) != 2 {
		t.Fatalf("matches = %d, want 2", len(matches))
	}
}

func TestEvaluate(t *testing.T) {
	tc := TestCase{
		ID:      "test_01",
		Subject: "Physique",
		Blocks: []OCRBlock{
			{Text: "La masse volumique est le rapport de la masse sur le volume. ρ = m / V.", BlockType: "TEXT", Confidence: 0.95},
		},
		Golden: GoldenOutput{
			Items: []GoldenItem{
				{Type: "KNOWLEDGE", Term: "masse volumique", Keywords: []string{"masse", "volume", "rapport"}},
				{Type: "PROCEDURE", Term: "ρ = m / V", Keywords: []string{"formule", "masse volumique"}},
			},
			Notions: []string{"Masse volumique"},
		},
	}

	parsed := &ParsedOutput{
		Items: []ParsedItem{
			{Type: "KNOWLEDGE", Term: "La masse volumique", Keywords: []string{"masse", "volume"}, Confidence: 0.9},
			{Type: "PROCEDURE", Term: "ρ = m / V", Keywords: []string{"formule", "masse volumique"}, Confidence: 0.85},
		},
		Notions: []string{"Masse volumique (ρ)"},
	}

	prov := &mockProvider{name: "test", model: "test-v1", priceIn: 3.0, priceOut: 15.0}
	resp := &Response{TokensInput: 1000, TokensOutput: 500, LatencyMs: 2000}

	result := Evaluate(tc, parsed, resp, prov)

	if result.CompletenessScore < 0.9 {
		t.Errorf("completeness = %f, want >= 0.9", result.CompletenessScore)
	}
	if result.ClassificationScore < 0.9 {
		t.Errorf("classification = %f, want >= 0.9", result.ClassificationScore)
	}
	if result.FidelityScore < 0.9 {
		t.Errorf("fidelity = %f, want >= 0.9", result.FidelityScore)
	}
	if result.HallucinationRate > 0.1 {
		t.Errorf("hallucination = %f, want <= 0.1", result.HallucinationRate)
	}
	if result.SchemaCompliance != true {
		t.Error("schema compliance should be true")
	}
	if result.CostUSD <= 0 {
		t.Error("cost should be > 0")
	}
}

type mockProvider struct {
	name     string
	model    string
	priceIn  float64
	priceOut float64
}

func (m *mockProvider) Name() string            { return m.name }
func (m *mockProvider) ModelID() string          { return m.model }
func (m *mockProvider) PricePerMInput() float64  { return m.priceIn }
func (m *mockProvider) PricePerMOutput() float64 { return m.priceOut }
func (m *mockProvider) StructureBlocks(_ context.Context, _, _ string) (*Response, error) {
	return nil, nil
}
