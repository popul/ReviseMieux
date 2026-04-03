package anthropic

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/popul/revisemieux/internal/domain/chapter"
)

// goldenFilePath returns the absolute path to a golden file in testdata/llm/.
func goldenFilePath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "testdata", "llm", name)
}

func TestParseStructurationResponse_GoldenFile(t *testing.T) {
	data, err := os.ReadFile(goldenFilePath("structuration_response_golden.json"))
	if err != nil {
		t.Fatalf("impossible de lire le golden file: %v", err)
	}

	var raw rawStructurationResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("erreur parsing JSON: %v", err)
	}

	if len(raw.Items) == 0 {
		t.Fatal("aucun item parsé")
	}

	if len(raw.Items) < 1 || len(raw.Items) > 50 {
		t.Errorf("nombre d'items hors fourchette [1,50]: %d", len(raw.Items))
	}

	if len(raw.Notions) == 0 {
		t.Error("aucune notion parsée")
	}

	// Validate all item types are valid domain types
	validTypes := map[string]bool{
		"KNOWLEDGE": true,
		"PROCEDURE": true,
		"DOCUMENT":  true,
		"WRITING":   true,
	}

	for i, item := range raw.Items {
		if !validTypes[item.Type] {
			t.Errorf("item %d: type invalide '%s' (attendu KNOWLEDGE|PROCEDURE|DOCUMENT|WRITING)", i, item.Type)
		}

		if item.Confidence < 0 || item.Confidence > 1 {
			t.Errorf("item %d: confiance hors bornes [0,1]: %f", i, item.Confidence)
		}

		if item.Term == "" {
			t.Errorf("item %d: term vide", i)
		}

		if item.NotionName == "" {
			t.Errorf("item %d: notion_name vide", i)
		}
	}
}

func TestParseStructurationResponse_MapToDomain(t *testing.T) {
	data, err := os.ReadFile(goldenFilePath("structuration_response_golden.json"))
	if err != nil {
		t.Fatalf("impossible de lire le golden file: %v", err)
	}

	var raw rawStructurationResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("erreur parsing JSON: %v", err)
	}

	// Simulate the mapping logic from StructureBlocks
	result := &chapter.StructurationResult{
		Notions: raw.Notions,
	}
	for _, item := range raw.Items {
		itemType, err := chapter.ParseItemType(item.Type)
		if err != nil {
			continue
		}
		result.Items = append(result.Items, chapter.StructuredItem{
			Type:       itemType,
			Term:       item.Term,
			Keywords:   item.Keywords,
			Steps:      item.Steps,
			NotionName: item.NotionName,
			Confidence: item.Confidence,
		})
	}

	if len(result.Items) == 0 {
		t.Fatal("aucun item domaine après mapping")
	}

	// Check that PROCEDURE items have steps
	for i, item := range result.Items {
		if item.Type == chapter.ItemProcedure && len(item.Steps) == 0 {
			t.Errorf("item %d: PROCEDURE sans steps", i)
		}
	}

	// Check that all four types are represented in golden file
	foundTypes := map[chapter.ItemType]bool{}
	for _, item := range result.Items {
		foundTypes[item.Type] = true
	}
	for _, expectedType := range []chapter.ItemType{chapter.ItemKnowledge, chapter.ItemProcedure, chapter.ItemDocument, chapter.ItemWriting} {
		if !foundTypes[expectedType] {
			t.Errorf("type %s absent du golden file", expectedType)
		}
	}
}

func TestParseStructurationResponse_JSONMalformed(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"json invalide", `{invalid json`},
		{"json vide", ``},
		{"tableau au lieu d'objet", `[1, 2, 3]`},
		{"texte brut", `ceci n'est pas du JSON`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw rawStructurationResponse
			err := json.Unmarshal([]byte(tt.input), &raw)
			if err == nil {
				t.Error("attendu une erreur pour JSON malformé, obtenu nil")
			}
		})
	}
}

func TestParseStructurationResponse_UnknownTypeSkipped(t *testing.T) {
	input := `{
		"items": [
			{"type": "KNOWLEDGE", "term": "Test", "keywords": [], "steps": [], "notion_name": "N", "confidence": 0.9},
			{"type": "INVALID_TYPE", "term": "Bad", "keywords": [], "steps": [], "notion_name": "N", "confidence": 0.5}
		],
		"notions": ["N"]
	}`

	var raw rawStructurationResponse
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		t.Fatalf("erreur parsing: %v", err)
	}

	// Simulate the mapping logic (unknown types should be skipped)
	var items []chapter.StructuredItem
	for _, item := range raw.Items {
		itemType, err := chapter.ParseItemType(item.Type)
		if err != nil {
			continue // skip unknown
		}
		items = append(items, chapter.StructuredItem{Type: itemType, Term: item.Term})
	}

	if len(items) != 1 {
		t.Errorf("attendu 1 item (type invalide ignoré), obtenu %d", len(items))
	}
	if items[0].Type != chapter.ItemKnowledge {
		t.Errorf("attendu KNOWLEDGE, obtenu %s", items[0].Type)
	}
}

func TestParseStructurationResponse_EmptyItems(t *testing.T) {
	input := `{"items": [], "notions": []}`

	var raw rawStructurationResponse
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		t.Fatalf("erreur parsing: %v", err)
	}

	if len(raw.Items) != 0 {
		t.Errorf("attendu 0 items, obtenu %d", len(raw.Items))
	}
	if len(raw.Notions) != 0 {
		t.Errorf("attendu 0 notions, obtenu %d", len(raw.Notions))
	}
}

func TestStripMarkdownFences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"pas de fences",
			`{"items": []}`,
			`{"items": []}`,
		},
		{
			"fences json",
			"```json\n{\"items\": []}\n```",
			`{"items": []}`,
		},
		{
			"fences sans type",
			"```\n{\"items\": []}\n```",
			`{"items": []}`,
		},
		{
			"espaces autour",
			"  \n```json\n{\"items\": []}\n```\n  ",
			`{"items": []}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripMarkdownFences(tt.input)
			if result != tt.expected {
				t.Errorf("attendu %q, obtenu %q", tt.expected, result)
			}
		})
	}
}

func TestParseStructurationResponse_WithMarkdownFences(t *testing.T) {
	input := "```json\n" + `{
		"items": [
			{"type": "KNOWLEDGE", "term": "Test", "keywords": ["a"], "steps": [], "notion_name": "N", "confidence": 0.9}
		],
		"notions": ["N"]
	}` + "\n```"

	cleaned := stripMarkdownFences(input)
	var raw rawStructurationResponse
	if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
		t.Fatalf("erreur parsing après strip fences: %v", err)
	}

	if len(raw.Items) != 1 {
		t.Errorf("attendu 1 item, obtenu %d", len(raw.Items))
	}
}
