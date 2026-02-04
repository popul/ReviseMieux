-- Rollback 010: Supprimer la table d'analyse des erreurs
DROP INDEX IF EXISTS idx_erreurs_analyse_type;
DROP INDEX IF EXISTS idx_erreurs_analyse_copie_id;
DROP TABLE IF EXISTS erreurs_analyse;
