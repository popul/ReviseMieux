-- Migration 004: Créer la table des sessions de quiz
CREATE TABLE quiz_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id UUID NOT NULL REFERENCES quiz(id) ON DELETE CASCADE,
    reponses JSONB DEFAULT '[]',
    score DECIMAL(5,2),
    termine BOOLEAN DEFAULT FALSE,
    date_debut TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    date_fin TIMESTAMP WITH TIME ZONE
);

-- Index pour les requêtes par quiz
CREATE INDEX idx_quiz_sessions_quiz_id ON quiz_sessions(quiz_id);
