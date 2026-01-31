-- Rollback 007: Supprimer la table d'activité
DROP INDEX IF EXISTS idx_activites_date;
DROP TABLE IF EXISTS activites;
