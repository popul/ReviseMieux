-- Rollback 004: Supprimer la table des sessions de quiz
DROP INDEX IF EXISTS idx_quiz_sessions_quiz_id;
DROP TABLE IF EXISTS quiz_sessions;
