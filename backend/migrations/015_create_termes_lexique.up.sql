-- Migration 015: Creer la table des termes du lexique
CREATE TABLE termes_lexique (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cours_id UUID NOT NULL REFERENCES cours(id) ON DELETE CASCADE,
    terme VARCHAR(255) NOT NULL,
    definition TEXT NOT NULL,
    contexte TEXT,
    exemple TEXT,
    categorie VARCHAR(100),
    maitrise INT DEFAULT 0 CHECK (maitrise >= 0 AND maitrise <= 5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_termes_lexique_cours_id ON termes_lexique(cours_id);
CREATE INDEX idx_termes_lexique_maitrise ON termes_lexique(maitrise);
