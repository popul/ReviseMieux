-- Rollback Migration 013: Supprimer la colonne blocs_texte
ALTER TABLE cours DROP COLUMN IF EXISTS blocs_texte;
