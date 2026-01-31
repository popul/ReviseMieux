-- Rollback 003: Supprimer la table des quiz
DROP INDEX IF EXISTS idx_quiz_cours_id;
DROP TABLE IF EXISTS quiz;
