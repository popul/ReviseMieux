-- Rollback 002: Supprimer la table des fiches
DROP INDEX IF EXISTS idx_fiches_cours_id;
DROP TABLE IF EXISTS fiches;
