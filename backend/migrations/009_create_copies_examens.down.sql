-- Rollback 009: Supprimer la table des copies d'examens
DROP INDEX IF EXISTS idx_copies_examens_matiere;
DROP INDEX IF EXISTS idx_copies_examens_date_creation;
DROP INDEX IF EXISTS idx_copies_examens_cours_id;
DROP TABLE IF EXISTS copies_examens;
