package openaicompat

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/popul/revisemieux/internal/domain/session"
)

func TestParseScorerResponse_GoldenFile(t *testing.T) {
	data, err := os.ReadFile(goldenFilePath("scorer_response_golden.json"))
	if err != nil {
		t.Fatalf("impossible de lire le golden file: %v", err)
	}

	// Parse the raw scoring response (same struct used in ScoreAnswer)
	var raw struct {
		Score       float64 `json:"score"`
		IsCorrect   bool    `json:"is_correct"`
		Explanation string  `json:"explanation"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("erreur parsing JSON: %v", err)
	}

	// Map to domain type (same as ScoreAnswer does)
	result := &session.LLMScoreResult{
		Score:       raw.Score,
		IsCorrect:   raw.IsCorrect,
		Explanation: raw.Explanation,
	}

	if result.Score < 0 || result.Score > 1 {
		t.Errorf("score hors bornes [0,1]: %f", result.Score)
	}

	if result.Explanation == "" {
		t.Error("explanation vide")
	}

	// Verify coherence: score >= 0.7 should mean is_correct
	if result.Score >= 0.7 && !result.IsCorrect {
		t.Errorf("incohérence: score %.2f >= 0.7 mais is_correct=false", result.Score)
	}
}

func TestParseScorerResponse_ScoreBounds(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		score   float64
	}{
		{
			"score correct",
			`{"score": 0.5, "is_correct": false, "explanation": "Réponse partielle"}`,
			false,
			0.5,
		},
		{
			"score zero",
			`{"score": 0.0, "is_correct": false, "explanation": "Réponse fausse"}`,
			false,
			0.0,
		},
		{
			"score parfait",
			`{"score": 1.0, "is_correct": true, "explanation": "Parfait"}`,
			false,
			1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw struct {
				Score       float64 `json:"score"`
				IsCorrect   bool    `json:"is_correct"`
				Explanation string  `json:"explanation"`
			}
			err := json.Unmarshal([]byte(tt.input), &raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("erreur inattendue: %v", err)
			}
			if !tt.wantErr && raw.Score != tt.score {
				t.Errorf("attendu score %f, obtenu %f", tt.score, raw.Score)
			}
		})
	}
}

func TestParseScorerResponse_Feedback(t *testing.T) {
	// Test that the explanation field carries useful feedback
	input := `{
		"score": 0.3,
		"is_correct": false,
		"explanation": "L'élève a confondu le numérateur et le dénominateur. La bonne réponse est 3/4, pas 4/3."
	}`

	var raw struct {
		Score       float64 `json:"score"`
		IsCorrect   bool    `json:"is_correct"`
		Explanation string  `json:"explanation"`
	}
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		t.Fatalf("erreur parsing: %v", err)
	}

	if raw.Explanation == "" {
		t.Error("explanation vide alors qu'un feedback est attendu")
	}

	if raw.Score >= 0.7 {
		t.Errorf("score %.2f >= 0.7 pour une réponse incorrecte", raw.Score)
	}

	if raw.IsCorrect {
		t.Error("is_correct devrait être false pour cette réponse")
	}
}

func TestParseScorerResponse_JSONMalformed(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"json invalide", `{broken`},
		{"json vide", ``},
		{"texte brut", `la réponse est correcte`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw struct {
				Score       float64 `json:"score"`
				IsCorrect   bool    `json:"is_correct"`
				Explanation string  `json:"explanation"`
			}
			err := json.Unmarshal([]byte(tt.input), &raw)
			if err == nil {
				t.Error("attendu une erreur pour JSON malformé, obtenu nil")
			}
		})
	}
}

func TestParseScorerResponse_WithMarkdownFences(t *testing.T) {
	input := "```json\n{\"score\": 0.9, \"is_correct\": true, \"explanation\": \"Très bien\"}\n```"

	cleaned := stripMarkdownFences(input)
	var raw struct {
		Score       float64 `json:"score"`
		IsCorrect   bool    `json:"is_correct"`
		Explanation string  `json:"explanation"`
	}
	if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
		t.Fatalf("erreur parsing après strip fences: %v", err)
	}

	if raw.Score != 0.9 {
		t.Errorf("attendu score 0.9, obtenu %f", raw.Score)
	}
}

func TestParseScorerResponse_MissingFields(t *testing.T) {
	// JSON with missing optional fields should still parse
	input := `{"score": 0.5}`
	var raw struct {
		Score       float64 `json:"score"`
		IsCorrect   bool    `json:"is_correct"`
		Explanation string  `json:"explanation"`
	}
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		t.Fatalf("erreur parsing: %v", err)
	}

	if raw.Score != 0.5 {
		t.Errorf("attendu score 0.5, obtenu %f", raw.Score)
	}
	// Missing fields should default to zero values
	if raw.IsCorrect != false {
		t.Error("is_correct devrait être false par défaut")
	}
	if raw.Explanation != "" {
		t.Error("explanation devrait être vide par défaut")
	}
}
