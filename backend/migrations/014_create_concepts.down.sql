-- Rollback 014: Supprimer la table des concepts
DROP INDEX IF EXISTS idx_concepts_importance;
DROP INDEX IF EXISTS idx_concepts_cours_id;
DROP TABLE IF EXISTS concepts;
