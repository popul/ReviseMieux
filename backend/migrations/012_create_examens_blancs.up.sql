CREATE TABLE examens_blancs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cours_id UUID NOT NULL REFERENCES cours(id) ON DELETE CASCADE,
    questions JSONB NOT NULL,
    duree_minutes INT DEFAULT 30,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sessions_examen (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    examen_id UUID NOT NULL REFERENCES examens_blancs(id) ON DELETE CASCADE,
    reponses JSONB,
    indices_utilises JSONB,
    note_estimee DECIMAL(5,2),
    points_forts JSONB,
    points_faibles JSONB,
    plan_revision JSONB,
    termine BOOLEAN DEFAULT FALSE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_examens_blancs_cours_id ON examens_blancs(cours_id);
CREATE INDEX idx_sessions_examen_examen_id ON sessions_examen(examen_id);
