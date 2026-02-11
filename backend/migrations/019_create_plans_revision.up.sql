CREATE TABLE plans_revision (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    titre VARCHAR(255) NOT NULL,
    description TEXT,
    matiere VARCHAR(100),
    icone_matiere VARCHAR(10) DEFAULT '📋',
    date_echeance TIMESTAMPTZ,
    date_creation TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    date_modification TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE plans_revision_cours (
    plan_id UUID NOT NULL REFERENCES plans_revision(id) ON DELETE CASCADE,
    cours_id UUID NOT NULL REFERENCES cours(id) ON DELETE CASCADE,
    ordre INT NOT NULL DEFAULT 0,
    date_ajout TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (plan_id, cours_id)
);

CREATE INDEX idx_plans_revision_date_echeance ON plans_revision(date_echeance);
CREATE INDEX idx_plans_revision_cours_plan_id ON plans_revision_cours(plan_id);
CREATE INDEX idx_plans_revision_cours_cours_id ON plans_revision_cours(cours_id);
