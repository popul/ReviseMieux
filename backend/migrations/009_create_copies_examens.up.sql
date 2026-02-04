-- Migration 009: Créer la table des copies d'examens corrigées
CREATE TABLE copies_examens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cours_id UUID REFERENCES cours(id) ON DELETE SET NULL,
    titre VARCHAR(255) NOT NULL,
    matiere VARCHAR(100),
    note_obtenue DECIMAL(4,2),
    note_totale DECIMAL(4,2),
    texte_ocr TEXT NOT NULL,
    annotations_professeur TEXT,
    confiance DECIMAL(3,2),
    zones_incertaines JSONB DEFAULT '[]',
    fichiers_originaux JSONB DEFAULT '[]',
    date_examen DATE,
    date_creation TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    date_modification TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index pour les requêtes par cours
CREATE INDEX idx_copies_examens_cours_id ON copies_examens(cours_id);

-- Index pour les recherches par date
CREATE INDEX idx_copies_examens_date_creation ON copies_examens(date_creation DESC);

-- Index pour les recherches par matière
CREATE INDEX idx_copies_examens_matiere ON copies_examens(matiere);
