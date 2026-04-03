package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/popul/revisemieux/internal/domain/event"
)

// Compile-time check that LLMCallLogger implements event.LLMCallLogger.
var _ event.LLMCallLogger = (*LLMCallLogger)(nil)

// LLMCallLogger implements event.LLMCallLogger using PostgreSQL.
type LLMCallLogger struct {
	pool *pgxpool.Pool
}

// NewLLMCallLogger creates a new LLMCallLogger.
func NewLLMCallLogger(pool *pgxpool.Pool) *LLMCallLogger {
	return &LLMCallLogger{pool: pool}
}

// LogCall inserts an LLM call log entry into the database.
func (l *LLMCallLogger) LogCall(ctx context.Context, entry event.LLMCallEntry) error {
	_, err := l.pool.Exec(ctx, `
		INSERT INTO llm_call_logs
		(timestamp, call_type, model, prompt_version, prompt_hash,
		 tokens_input, tokens_output, duration_ms, cost_usd, status,
		 items_count, avg_confidence, cache_hit, batch, user_id, chapter_id, error)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		entry.Timestamp, entry.CallType, entry.Model, entry.PromptVersion, entry.PromptHash,
		entry.TokensInput, entry.TokensOutput, entry.DurationMs, entry.CostUSD, entry.Status,
		entry.ItemsCount, entry.AvgConfidence, entry.CacheHit, entry.Batch,
		entry.UserID, entry.ChapterID, entry.Error)
	if err != nil {
		return fmt.Errorf("llm_logger.LogCall: %w", err)
	}
	return nil
}
