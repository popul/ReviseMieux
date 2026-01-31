-- Migration 001: Créer la table des cours
-- Extension pour UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Table des cours
CREATE TABLE cours (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    titre VARCHAR(255) NOT NULL,
    matiere VARCHAR(100),
    texte_ocr TEXT NOT NULL,
    texte_corrige TEXT,
    confiance DECIMAL(3,2),
    zones_incertaines JSONB DEFAULT '[]',
    fichiers_originaux JSONB DEFAULT '[]',
    date_creation TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    date_modification TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index pour les recherches par date
CREATE INDEX idx_cours_date_creation ON cours(date_creation DESC);
