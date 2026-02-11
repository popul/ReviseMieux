-- Migration 014: Créer la table des concepts extraits d'un cours
CREATE TABLE concepts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cours_id UUID NOT NULL REFERENCES cours(id) ON DELETE CASCADE,
    nom VARCHAR(255) NOT NULL,
    definition TEXT NOT NULL,
    importance VARCHAR(20) NOT NULL DEFAULT 'important' CHECK (importance IN ('essentiel', 'important', 'secondaire')),
    position_dans_cours JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index pour les requêtes par cours
CREATE INDEX idx_concepts_cours_id ON concepts(cours_id);

-- Index pour filtrer par importance
CREATE INDEX idx_concepts_importance ON concepts(importance);
