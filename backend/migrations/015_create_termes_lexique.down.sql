-- Rollback 015: Supprimer la table des termes du lexique
DROP INDEX IF EXISTS idx_termes_lexique_maitrise;
DROP INDEX IF EXISTS idx_termes_lexique_cours_id;
DROP TABLE IF EXISTS termes_lexique;
