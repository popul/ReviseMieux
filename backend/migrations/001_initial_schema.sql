-- Migration: 001_initial_schema.sql
-- Schéma initial Lot 0 — Révise Mieux
-- Choix : UUID v7, CREATE TYPE enums, tables de jointure, JSONB pour champs structurés

-- ============================================================
-- Extensions
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- Enums
-- ============================================================

CREATE TYPE user_role AS ENUM ('student', 'parent');

CREATE TYPE block_type AS ENUM (
    'TEXT', 'PHOTO', 'SCHEMA', 'MAP', 'GRAPH', 'TABLE', 'CIRCUIT', 'DECORATIVE'
);

CREATE TYPE pedagogical_classification AS ENUM ('pedagogical', 'decorative');

CREATE TYPE item_type AS ENUM ('KNOWLEDGE', 'PROCEDURE', 'DOCUMENT', 'WRITING');

CREATE TYPE mastery_state AS ENUM ('UNKNOWN', 'FRAGILE', 'OK', 'SOLID');

CREATE TYPE session_type AS ENUM (
    'daily', 'diagnostic', 'mock_exam', 'evening_first', 'pre_class'
);

CREATE TYPE session_status AS ENUM (
    'COMPOSING', 'IN_PROGRESS', 'COMPLETED', 'EXPIRED', 'ABANDONED'
);

CREATE TYPE session_trigger AS ENUM ('manual', 'scheduled', 'notification');

CREATE TYPE page_status AS ENUM (
    'UPLOADING', 'OCR_PENDING', 'OCR_PROCESSING', 'PROCESSED',
    'ITEMS_GENERATING', 'DONE', 'NO_ITEMS', 'ITEMS_FAILED', 'FAILED'
);

CREATE TYPE revision_status AS ENUM ('PROCESSING', 'READY', 'PARTIAL', 'FAILED');

CREATE TYPE validation_task_status AS ENUM (
    'PENDING', 'CONFIRMED', 'CORRECTED', 'UNKNOWN_ANSWER', 'IGNORED'
);

CREATE TYPE validation_task_source AS ENUM (
    'uncertainty_detection', 'student_report', 'anomaly_detection',
    'coherence_check', 'fidelity_check'
);

CREATE TYPE exam_status AS ENUM ('active', 'past');

CREATE TYPE question_type AS ENUM ('MCQ', 'SHORT_ANSWER', 'NUMERIC', 'CLOZE', 'RUBRIC');

CREATE TYPE fidelity_flag AS ENUM ('low', 'medium');

CREATE TYPE coherence_flag AS ENUM ('contradiction', 'orphan_reference');

CREATE TYPE anomaly_flag AS ENUM ('high_failure_rate');

CREATE TYPE visual_block_type AS ENUM (
    'diagram', 'graph', 'table', 'figure', 'map', 'circuit', 'photo'
);

-- ============================================================
-- Tables
-- ============================================================

-- Users
CREATE TABLE users (
    id          UUID PRIMARY KEY,
    role        user_role NOT NULL,
    email       TEXT UNIQUE,
    display_name TEXT,
    consent_parent_at TIMESTAMPTZ,
    linked_student_id UUID REFERENCES users(id),
    timezone    TEXT NOT NULL DEFAULT 'Europe/Paris',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Exams
CREATE TABLE exams (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id),
    name        TEXT,
    exam_date   DATE NOT NULL,
    status      exam_status NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Chapters
CREATE TABLE chapters (
    id                  UUID PRIMARY KEY,
    user_id             UUID NOT NULL REFERENCES users(id),
    subject             TEXT NOT NULL,
    class_level         TEXT NOT NULL,
    name                TEXT NOT NULL,
    pack_id             TEXT,
    current_revision_id UUID, -- FK ajoutée après création de chapter_revisions
    archived            BOOLEAN NOT NULL DEFAULT false,
    is_demo             BOOLEAN NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Chapter <-> Exam (N:N)
CREATE TABLE chapter_exams (
    chapter_id  UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    exam_id     UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    PRIMARY KEY (chapter_id, exam_id)
);

-- Exam <-> Notion (N:N, pour scope exam par notion)
-- Créée après la table notions

-- Chapter Revisions
CREATE TABLE chapter_revisions (
    id              UUID PRIMARY KEY,
    chapter_id      UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    revision_number INTEGER NOT NULL,
    status          revision_status NOT NULL DEFAULT 'PROCESSING',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (chapter_id, revision_number)
);

-- FK différée : chapters.current_revision_id -> chapter_revisions
ALTER TABLE chapters
    ADD CONSTRAINT fk_chapters_current_revision
    FOREIGN KEY (current_revision_id) REFERENCES chapter_revisions(id);

-- Pages
CREATE TABLE pages (
    id          UUID PRIMARY KEY,
    revision_id UUID NOT NULL REFERENCES chapter_revisions(id) ON DELETE CASCADE,
    photo_url   TEXT NOT NULL,
    page_order  INTEGER NOT NULL,
    ocr_status  page_status NOT NULL DEFAULT 'UPLOADING',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Blocks
CREATE TABLE blocks (
    id              UUID PRIMARY KEY,
    page_id         UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    block_type      block_type NOT NULL,
    crop_url        TEXT,
    crop_bbox       JSONB, -- {x, y, w, h}
    confidence      REAL NOT NULL DEFAULT 0,
    ocr_text        TEXT,
    pedagogical_classification pedagogical_classification,
    classification_confidence  REAL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Visual Blocks (dual coding — schémas, graphiques, tableaux, cartes, circuits, photos)
CREATE TABLE visual_blocks (
    id              UUID PRIMARY KEY,
    block_id        UUID NOT NULL REFERENCES blocks(id) ON DELETE CASCADE,
    chapter_id      UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    visual_type     visual_block_type NOT NULL,
    image_url       TEXT NOT NULL,
    thumbnail_url   TEXT,
    width_px        INTEGER NOT NULL,
    height_px       INTEGER NOT NULL,
    labels          JSONB NOT NULL DEFAULT '[]', -- [{text, position: {x, y}}]
    axis_labels     JSONB, -- {x_label, y_label, x_unit?, y_unit?}
    table_structure JSONB, -- {rows, cols, headers[]?, cells[][]}
    caption         TEXT,
    alt_text        TEXT,
    retention_expires_at TIMESTAMPTZ, -- RGPD J+30
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_visual_blocks_chapter ON visual_blocks(chapter_id);
CREATE INDEX idx_visual_blocks_block ON visual_blocks(block_id);

-- Notions
CREATE TABLE notions (
    id          UUID PRIMARY KEY,
    chapter_id  UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Notion concept_tags
CREATE TABLE notion_concept_tags (
    notion_id   UUID NOT NULL REFERENCES notions(id) ON DELETE CASCADE,
    concept_tag TEXT NOT NULL,
    PRIMARY KEY (notion_id, concept_tag)
);

-- Exam <-> Notion (N:N)
CREATE TABLE exam_notions (
    exam_id     UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    notion_id   UUID NOT NULL REFERENCES notions(id) ON DELETE CASCADE,
    PRIMARY KEY (exam_id, notion_id)
);

-- Items
CREATE TABLE items (
    id                  UUID PRIMARY KEY,
    chapter_id          UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    notion_id           UUID REFERENCES notions(id) ON DELETE SET NULL,
    revision_id         UUID NOT NULL REFERENCES chapter_revisions(id),
    item_type           item_type NOT NULL,
    term                TEXT,
    linked_doc_id       UUID, -- FK vers documents (post Lot 0)
    confidence          REAL NOT NULL DEFAULT 0,
    validation_required BOOLEAN NOT NULL DEFAULT false,
    archived            BOOLEAN NOT NULL DEFAULT false,
    fidelity_score      REAL,
    fidelity_flag       fidelity_flag,
    coherence_flag      coherence_flag,
    anomaly_flag        anomaly_flag,
    llm_model_version   TEXT,
    prompt_template_version TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Item keywords
CREATE TABLE item_keywords (
    item_id     UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    keyword     TEXT NOT NULL,
    position    INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (item_id, keyword)
);

-- Item steps (pour PROCEDURE)
CREATE TABLE item_steps (
    id          UUID PRIMARY KEY,
    item_id     UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    step_order  INTEGER NOT NULL,
    content     TEXT NOT NULL,
    UNIQUE (item_id, step_order)
);

-- Item tags
CREATE TABLE item_tags (
    item_id     UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    tag         TEXT NOT NULL,
    PRIMARY KEY (item_id, tag)
);

-- Item <-> VisualBlock (N:N)
CREATE TABLE item_visual_blocks (
    item_id         UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    visual_block_id UUID NOT NULL REFERENCES visual_blocks(id) ON DELETE CASCADE,
    PRIMARY KEY (item_id, visual_block_id)
);

-- Notion <-> Item (via items.notion_id, pas besoin de table jointure)

-- Templates
CREATE TABLE templates (
    id              TEXT PRIMARY KEY, -- ex: GEN.KNOW.DEF_SHORT
    name            TEXT NOT NULL,
    version         INTEGER NOT NULL DEFAULT 1,
    question_type   question_type NOT NULL,
    difficulty      INTEGER NOT NULL DEFAULT 1,
    eligibility     JSONB NOT NULL DEFAULT '{}',
    prompt_template TEXT NOT NULL,
    grading         JSONB NOT NULL DEFAULT '{}',
    uses_visual     BOOLEAN NOT NULL DEFAULT false,
    visual_interaction_type TEXT, -- label_completion, describe, matching, read_value, identify_zone
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Template variables
CREATE TABLE template_variables (
    template_id TEXT NOT NULL REFERENCES templates(id) ON DELETE CASCADE,
    variable    TEXT NOT NULL,
    PRIMARY KEY (template_id, variable)
);

-- Questions
CREATE TABLE questions (
    id                  UUID PRIMARY KEY,
    template_id         TEXT NOT NULL REFERENCES templates(id),
    item_id             UUID NOT NULL REFERENCES items(id),
    visual_block_id     UUID REFERENCES visual_blocks(id),
    rendered_prompt     TEXT NOT NULL,
    rendered_visual_url TEXT, -- URL du visuel transformé (légendes masquées, zones floutées)
    expected_answer     JSONB NOT NULL,
    grading_policy      TEXT NOT NULL,
    clarification       JSONB, -- {intent, starter_hint}
    llm_model_version   TEXT,
    prompt_template_version TEXT,
    times_seen          INTEGER NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Mastery
CREATE TABLE masteries (
    id                      UUID PRIMARY KEY,
    user_id                 UUID NOT NULL REFERENCES users(id),
    item_id                 UUID NOT NULL REFERENCES items(id),
    state                   mastery_state NOT NULL DEFAULT 'UNKNOWN',
    next_due_at             TIMESTAMPTZ,
    last_review_at          TIMESTAMPTZ,
    last_success_at         TIMESTAMPTZ,
    consecutive_successes   INTEGER NOT NULL DEFAULT 0,
    consecutive_failures    INTEGER NOT NULL DEFAULT 0,
    current_difficulty      INTEGER, -- override Z1-AC19, null = template standard
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, item_id)
);

-- Sessions
CREATE TABLE sessions (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id),
    session_type    session_type NOT NULL,
    status          session_status NOT NULL DEFAULT 'COMPOSING',
    trigger         session_trigger NOT NULL DEFAULT 'manual',
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    current_question_index INTEGER NOT NULL DEFAULT 0,
    includes_pre_class     BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Session <-> Chapter (N:N)
CREATE TABLE session_chapters (
    session_id  UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    chapter_id  UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    PRIMARY KEY (session_id, chapter_id)
);

-- Session <-> Question (ordonnée)
CREATE TABLE session_questions (
    session_id      UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    question_id     UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    question_order  INTEGER NOT NULL,
    PRIMARY KEY (session_id, question_id)
);

-- Attempts
CREATE TABLE attempts (
    id              UUID PRIMARY KEY,
    question_id     UUID NOT NULL REFERENCES questions(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    answer          JSONB NOT NULL,
    score           REAL NOT NULL,
    feedback        TEXT,
    source          TEXT NOT NULL DEFAULT 'interactive',
    rapid_response  BOOLEAN NOT NULL DEFAULT false,
    response_time_ms INTEGER,
    hint_used       BOOLEAN NOT NULL DEFAULT false,
    clarification_used BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Validation Tasks
CREATE TABLE validation_tasks (
    id          UUID PRIMARY KEY,
    item_id     UUID NOT NULL REFERENCES items(id),
    crop_url    TEXT,
    suggestion  TEXT,
    priority    INTEGER NOT NULL DEFAULT 0,
    status      validation_task_status NOT NULL DEFAULT 'PENDING',
    resolved_by UUID REFERENCES users(id),
    source      validation_task_source NOT NULL,
    student_note TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- Index
-- ============================================================

-- Mastery : requêtes fréquentes par user + état + due date
CREATE INDEX idx_masteries_user_state ON masteries(user_id, state);
CREATE INDEX idx_masteries_user_due ON masteries(user_id, next_due_at) WHERE next_due_at IS NOT NULL;
CREATE INDEX idx_masteries_item ON masteries(item_id);

-- Items par chapitre (composition de session, carte leçon)
CREATE INDEX idx_items_chapter ON items(chapter_id) WHERE NOT archived;
CREATE INDEX idx_items_revision ON items(revision_id);

-- Sessions actives par user
CREATE INDEX idx_sessions_user_status ON sessions(user_id, status);

-- Attempts par question et par user
CREATE INDEX idx_attempts_question ON attempts(question_id);
CREATE INDEX idx_attempts_user ON attempts(user_id, created_at);

-- Validation tasks non résolues
CREATE INDEX idx_validation_tasks_pending ON validation_tasks(status) WHERE status = 'PENDING';

-- Pages par revision
CREATE INDEX idx_pages_revision ON pages(revision_id, page_order);

-- Blocks par page
CREATE INDEX idx_blocks_page ON blocks(page_id);

-- Chapter revisions
CREATE INDEX idx_chapter_revisions_chapter ON chapter_revisions(chapter_id);

-- Chapters par user
CREATE INDEX idx_chapters_user ON chapters(user_id) WHERE NOT archived;
