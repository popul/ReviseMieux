-- Migration 006: Créer la table des ressources suggérées
CREATE TABLE ressources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cours_id UUID NOT NULL REFERENCES cours(id) ON DELETE CASCADE,
    titre VARCHAR(255) NOT NULL,
    url VARCHAR(500),
    type VARCHAR(50) NOT NULL CHECK (type IN ('video', 'article', 'site')),
    description TEXT,
    date_creation TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index pour les requêtes par cours
CREATE INDEX idx_ressources_cours_id ON ressources(cours_id);
