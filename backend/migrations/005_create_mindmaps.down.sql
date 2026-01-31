-- Rollback 005: Supprimer la table des mindmaps
DROP INDEX IF EXISTS idx_mindmaps_cours_id;
DROP TABLE IF EXISTS mindmaps;
