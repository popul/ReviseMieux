-- Migration: 006_table_categories.sql
-- Tag tables as 'reference' or 'user_data' via PostgreSQL comments.
-- Used by the E2E seed endpoint to know which tables to truncate.
--
-- Convention:
--   - 'reference'  : données seedées par les migrations (templates, config). Jamais truncatées.
--   - 'user_data'  : données générées par les utilisateurs. Truncatées lors du reset E2E.
--   - schema_migrations n'a pas de tag (toujours préservée).

-- Reference tables
COMMENT ON TABLE templates IS 'reference';
COMMENT ON TABLE template_variables IS 'reference';

-- User data tables
COMMENT ON TABLE users IS 'user_data';
COMMENT ON TABLE chapters IS 'user_data';
COMMENT ON TABLE chapter_revisions IS 'user_data';
COMMENT ON TABLE chapter_exams IS 'user_data';
COMMENT ON TABLE exams IS 'user_data';
COMMENT ON TABLE pages IS 'user_data';
COMMENT ON TABLE blocks IS 'user_data';
COMMENT ON TABLE visual_blocks IS 'user_data';
COMMENT ON TABLE notions IS 'user_data';
COMMENT ON TABLE notion_concept_tags IS 'user_data';
COMMENT ON TABLE exam_notions IS 'user_data';
COMMENT ON TABLE items IS 'user_data';
COMMENT ON TABLE item_keywords IS 'user_data';
COMMENT ON TABLE item_steps IS 'user_data';
COMMENT ON TABLE item_tags IS 'user_data';
COMMENT ON TABLE item_visual_blocks IS 'user_data';
COMMENT ON TABLE questions IS 'user_data';
COMMENT ON TABLE masteries IS 'user_data';
COMMENT ON TABLE sessions IS 'user_data';
COMMENT ON TABLE session_chapters IS 'user_data';
COMMENT ON TABLE session_questions IS 'user_data';
COMMENT ON TABLE attempts IS 'user_data';
COMMENT ON TABLE validation_tasks IS 'user_data';
COMMENT ON TABLE llm_call_logs IS 'user_data';
