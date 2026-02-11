-- Migration 013: Ajouter la colonne blocs_texte pour les positions OCR
-- Stocke les blocs de texte avec leurs positions approximatives dans l'image
ALTER TABLE cours ADD COLUMN IF NOT EXISTS blocs_texte JSONB DEFAULT '[]';
