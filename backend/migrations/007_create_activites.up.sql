-- Migration 007: Créer la table d'activité pour le dashboard
CREATE TABLE activites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL CHECK (type IN ('ocr', 'fiches', 'quiz', 'mindmap', 'ressources')),
    description TEXT NOT NULL,
    reference_id UUID,
    reference_type VARCHAR(50),
    metadata JSONB DEFAULT '{}',
    date_creation TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index pour les requêtes par date (dashboard activité récente)
CREATE INDEX idx_activites_date ON activites(date_creation DESC);
