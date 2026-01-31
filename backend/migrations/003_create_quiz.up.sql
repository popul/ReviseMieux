-- Migration 003: Créer la table des quiz
CREATE TABLE quiz (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cours_id UUID NOT NULL REFERENCES cours(id) ON DELETE CASCADE,
    titre VARCHAR(255) NOT NULL,
    difficulte VARCHAR(20) NOT NULL CHECK (difficulte IN ('facile', 'moyen', 'difficile')),
    nombre_questions INT NOT NULL,
    questions JSONB NOT NULL,
    date_creation TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index pour les requêtes par cours
CREATE INDEX idx_quiz_cours_id ON quiz(cours_id);
