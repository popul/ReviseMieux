-- Rollback 008: Supprimer la table des quotas
DROP INDEX IF EXISTS idx_quotas_date;
DROP TABLE IF EXISTS quotas_journaliers;
