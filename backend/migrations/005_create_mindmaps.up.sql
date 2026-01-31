-- Migration 005: Créer la table des mindmaps
CREATE TABLE mindmaps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cours_id UUID NOT NULL REFERENCES cours(id) ON DELETE CASCADE,
    noeuds JSONB NOT NULL,
    liens JSONB NOT NULL,
    date_creation TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index pour les requêtes par cours
CREATE INDEX idx_mindmaps_cours_id ON mindmaps(cours_id);
