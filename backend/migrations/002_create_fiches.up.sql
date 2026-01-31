-- Migration 002: Créer la table des fiches de révision
CREATE TABLE fiches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cours_id UUID NOT NULL REFERENCES cours(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    reponse TEXT NOT NULL,
    difficulte VARCHAR(20) NOT NULL CHECK (difficulte IN ('facile', 'moyen', 'difficile')),
    ordre INT DEFAULT 0,
    date_creation TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index pour les requêtes par cours
CREATE INDEX idx_fiches_cours_id ON fiches(cours_id);
