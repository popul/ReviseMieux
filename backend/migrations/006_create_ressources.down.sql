-- Rollback 006: Supprimer la table des ressources
DROP INDEX IF EXISTS idx_ressources_cours_id;
DROP TABLE IF EXISTS ressources;
