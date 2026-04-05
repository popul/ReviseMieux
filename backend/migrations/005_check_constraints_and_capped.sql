-- Migration: 005_check_constraints_and_capped.sql
-- Adds missing CHECK constraints for data integrity + capped_at_ok column for Z1-AC13.

-- ============================================================
-- CHECK constraints on float/integer ranges
-- ============================================================

-- Items: confidence must be in [0, 1]
ALTER TABLE items ADD CONSTRAINT chk_items_confidence
    CHECK (confidence >= 0 AND confidence <= 1);

-- Items: fidelity_score must be in [0, 1]
ALTER TABLE items ADD CONSTRAINT chk_items_fidelity_score
    CHECK (fidelity_score IS NULL OR (fidelity_score >= 0 AND fidelity_score <= 1));

-- Blocks: confidence must be in [0, 1]
ALTER TABLE blocks ADD CONSTRAINT chk_blocks_confidence
    CHECK (confidence >= 0 AND confidence <= 1);

-- Blocks: classification_confidence must be in [0, 1]
ALTER TABLE blocks ADD CONSTRAINT chk_blocks_classification_confidence
    CHECK (classification_confidence IS NULL OR (classification_confidence >= 0 AND classification_confidence <= 1));

-- Pages: ocr_confidence must be in [0, 1]
ALTER TABLE pages ADD CONSTRAINT chk_pages_ocr_confidence
    CHECK (ocr_confidence IS NULL OR (ocr_confidence >= 0 AND ocr_confidence <= 1));

-- Attempts: score must be in [0, 1]
ALTER TABLE attempts ADD CONSTRAINT chk_attempts_score
    CHECK (score >= 0 AND score <= 1);

-- Attempts: response_time_ms must be positive
ALTER TABLE attempts ADD CONSTRAINT chk_attempts_response_time
    CHECK (response_time_ms IS NULL OR response_time_ms >= 0);

-- Masteries: consecutive counters must be non-negative
ALTER TABLE masteries ADD CONSTRAINT chk_masteries_consecutive_successes
    CHECK (consecutive_successes >= 0);

ALTER TABLE masteries ADD CONSTRAINT chk_masteries_consecutive_failures
    CHECK (consecutive_failures >= 0);

-- Masteries: current_difficulty must be in [1, 5]
ALTER TABLE masteries ADD CONSTRAINT chk_masteries_difficulty
    CHECK (current_difficulty IS NULL OR (current_difficulty >= 1 AND current_difficulty <= 5));

-- Questions: times_seen must be non-negative
ALTER TABLE questions ADD CONSTRAINT chk_questions_times_seen
    CHECK (times_seen >= 0);

-- Validation tasks: priority must be non-negative
ALTER TABLE validation_tasks ADD CONSTRAINT chk_validation_tasks_priority
    CHECK (priority >= 0);

-- LLM call logs: tokens and duration must be non-negative
ALTER TABLE llm_call_logs ADD CONSTRAINT chk_llm_tokens_input
    CHECK (tokens_input >= 0);

ALTER TABLE llm_call_logs ADD CONSTRAINT chk_llm_tokens_output
    CHECK (tokens_output >= 0);

ALTER TABLE llm_call_logs ADD CONSTRAINT chk_llm_duration
    CHECK (duration_ms >= 0);

ALTER TABLE llm_call_logs ADD CONSTRAINT chk_llm_cost
    CHECK (cost_usd >= 0);

-- ============================================================
-- Z1-AC13: capped_at_ok column on masteries
-- Items with validation_required=true cap mastery at OK state.
-- ============================================================

ALTER TABLE masteries ADD COLUMN IF NOT EXISTS capped_at_ok BOOLEAN NOT NULL DEFAULT false;
