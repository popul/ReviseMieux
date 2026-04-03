-- Migration: 004_llm_call_logs.sql
-- Table de monitoring des appels LLM pour suivi coûts, latence et qualité.

CREATE TABLE llm_call_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT now(),
    call_type       TEXT NOT NULL,  -- 'structuration', 'fidelity', 'question_gen', 'coherence', 'ocr', 'scoring'
    model           TEXT NOT NULL,
    prompt_version  TEXT NOT NULL,
    prompt_hash     TEXT NOT NULL,
    tokens_input    INTEGER NOT NULL DEFAULT 0,
    tokens_output   INTEGER NOT NULL DEFAULT 0,
    duration_ms     INTEGER NOT NULL DEFAULT 0,
    cost_usd        DOUBLE PRECISION NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'success',  -- 'success', 'error', 'timeout'
    items_count     INTEGER,
    avg_confidence  DOUBLE PRECISION,
    cache_hit       BOOLEAN NOT NULL DEFAULT false,
    batch           BOOLEAN NOT NULL DEFAULT false,
    user_id         UUID,
    chapter_id      UUID,
    error           TEXT
);

CREATE INDEX idx_llm_call_logs_type_ts ON llm_call_logs (call_type, timestamp);
CREATE INDEX idx_llm_call_logs_model_ts ON llm_call_logs (model, timestamp);
