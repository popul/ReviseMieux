-- Rollback 001: Supprimer la table des cours
DROP INDEX IF EXISTS idx_cours_date_creation;
DROP TABLE IF EXISTS cours;
