package event

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestLLMCallEntry_AllFields(t *testing.T) {
	userID := uuid.New()
	chapterID := uuid.New()
	itemsCount := 5
	avgConfidence := 0.92

	entry := LLMCallEntry{
		Timestamp:     time.Now(),
		CallType:      "structuration",
		Model:         "claude-sonnet-4-20250514",
		PromptVersion: "v1.0",
		PromptHash:    "abc123",
		TokensInput:   1500,
		TokensOutput:  800,
		DurationMs:    2345,
		CostUSD:       0.0042,
		Status:        "success",
		ItemsCount:    &itemsCount,
		AvgConfidence: &avgConfidence,
		CacheHit:      false,
		Batch:         false,
		UserID:        &userID,
		ChapterID:     &chapterID,
		Error:         "",
	}

	if entry.CallType != "structuration" {
		t.Errorf("CallType = %q, want %q", entry.CallType, "structuration")
	}
	if entry.TokensInput != 1500 {
		t.Errorf("TokensInput = %d, want %d", entry.TokensInput, 1500)
	}
	if entry.TokensOutput != 800 {
		t.Errorf("TokensOutput = %d, want %d", entry.TokensOutput, 800)
	}
	if *entry.ItemsCount != 5 {
		t.Errorf("ItemsCount = %d, want %d", *entry.ItemsCount, 5)
	}
	if *entry.AvgConfidence != 0.92 {
		t.Errorf("AvgConfidence = %f, want %f", *entry.AvgConfidence, 0.92)
	}
	if *entry.UserID != userID {
		t.Errorf("UserID mismatch")
	}
	if *entry.ChapterID != chapterID {
		t.Errorf("ChapterID mismatch")
	}
}

func TestLLMCallEntry_DefaultValues(t *testing.T) {
	entry := LLMCallEntry{}

	if entry.TokensInput != 0 {
		t.Errorf("default TokensInput = %d, want 0", entry.TokensInput)
	}
	if entry.TokensOutput != 0 {
		t.Errorf("default TokensOutput = %d, want 0", entry.TokensOutput)
	}
	if entry.DurationMs != 0 {
		t.Errorf("default DurationMs = %d, want 0", entry.DurationMs)
	}
	if entry.CostUSD != 0 {
		t.Errorf("default CostUSD = %f, want 0", entry.CostUSD)
	}
	if entry.CacheHit != false {
		t.Errorf("default CacheHit = %v, want false", entry.CacheHit)
	}
	if entry.Batch != false {
		t.Errorf("default Batch = %v, want false", entry.Batch)
	}
	if entry.ItemsCount != nil {
		t.Errorf("default ItemsCount = %v, want nil", entry.ItemsCount)
	}
	if entry.AvgConfidence != nil {
		t.Errorf("default AvgConfidence = %v, want nil", entry.AvgConfidence)
	}
	if entry.UserID != nil {
		t.Errorf("default UserID = %v, want nil", entry.UserID)
	}
	if entry.ChapterID != nil {
		t.Errorf("default ChapterID = %v, want nil", entry.ChapterID)
	}
}
