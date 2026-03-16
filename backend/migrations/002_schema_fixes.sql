-- Migration: 002_schema_fixes.sql
-- Fix schema mismatches between Go domain code and DB schema

-- ============================================================
-- Enum additions
-- ============================================================

-- Z2-AC03: BLURRY page status
ALTER TYPE page_status ADD VALUE IF NOT EXISTS 'BLURRY';

-- Z3-AC04/05: Student-side validation task resolutions
ALTER TYPE validation_task_status ADD VALUE IF NOT EXISTS 'DEFERRED_BY_STUDENT';
ALTER TYPE validation_task_status ADD VALUE IF NOT EXISTS 'IGNORED_BY_STUDENT';

-- Z3-AC04: Student deferred source
ALTER TYPE validation_task_source ADD VALUE IF NOT EXISTS 'student_deferred';

-- ============================================================
-- Table alterations
-- ============================================================

-- Z2-AC02: OCR timeout tracking on pages
ALTER TABLE pages ADD COLUMN IF NOT EXISTS fail_reason TEXT;
ALTER TABLE pages ADD COLUMN IF NOT EXISTS ocr_confidence REAL;

-- Session ID on questions (repo expects direct FK, not just join table)
ALTER TABLE questions ADD COLUMN IF NOT EXISTS session_id UUID REFERENCES sessions(id);
CREATE INDEX IF NOT EXISTS idx_questions_session ON questions(session_id);

-- Z2-AC08: Visual block source tracking on items
ALTER TABLE items ADD COLUMN IF NOT EXISTS source_image_url TEXT;
