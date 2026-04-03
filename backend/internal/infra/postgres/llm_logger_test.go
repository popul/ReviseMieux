//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/infra/postgres"
)

func TestLLMCallLogger_LogCall(t *testing.T) {
	tdb := setupTestDB(t)
	logger := postgres.NewLLMCallLogger(tdb.pool)

	itemsCount := 10
	avgConfidence := 0.95
	userID := uuid.Must(uuid.NewV7())
	chapterID := uuid.Must(uuid.NewV7())

	entry := event.LLMCallEntry{
		Timestamp:     time.Now().UTC().Truncate(time.Microsecond),
		CallType:      "structuration",
		Model:         "claude-sonnet-4-20250514",
		PromptVersion: "v1.0",
		PromptHash:    "sha256-abc123",
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

	err := logger.LogCall(context.Background(), entry)
	if err != nil {
		t.Fatalf("LogCall() error = %v", err)
	}

	// Verify the entry was inserted
	var count int
	err = tdb.pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM llm_call_logs WHERE call_type = 'structuration'").Scan(&count)
	if err != nil {
		t.Fatalf("query error = %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestLLMCallLogger_LogCall_NilOptionalFields(t *testing.T) {
	tdb := setupTestDB(t)
	logger := postgres.NewLLMCallLogger(tdb.pool)

	entry := event.LLMCallEntry{
		Timestamp:     time.Now().UTC().Truncate(time.Microsecond),
		CallType:      "ocr",
		Model:         "gpt-4o",
		PromptVersion: "v2.1",
		PromptHash:    "sha256-def456",
		TokensInput:   500,
		TokensOutput:  200,
		DurationMs:    1200,
		CostUSD:       0.001,
		Status:        "error",
		ItemsCount:    nil,
		AvgConfidence: nil,
		CacheHit:      false,
		Batch:         false,
		UserID:        nil,
		ChapterID:     nil,
		Error:         "timeout after 30s",
	}

	err := logger.LogCall(context.Background(), entry)
	if err != nil {
		t.Fatalf("LogCall() with nil optional fields error = %v", err)
	}

	// Verify NULL optional fields
	var hasItemsCount, hasAvgConfidence, hasUserID, hasChapterID bool
	err = tdb.pool.QueryRow(context.Background(), `
		SELECT items_count IS NOT NULL, avg_confidence IS NOT NULL,
		       user_id IS NOT NULL, chapter_id IS NOT NULL
		FROM llm_call_logs WHERE call_type = 'ocr'
	`).Scan(&hasItemsCount, &hasAvgConfidence, &hasUserID, &hasChapterID)
	if err != nil {
		t.Fatalf("query error = %v", err)
	}
	if hasItemsCount {
		t.Error("expected items_count to be NULL")
	}
	if hasAvgConfidence {
		t.Error("expected avg_confidence to be NULL")
	}
	if hasUserID {
		t.Error("expected user_id to be NULL")
	}
	if hasChapterID {
		t.Error("expected chapter_id to be NULL")
	}
}
