-- Migration 010: Créer la table d'analyse des erreurs
CREATE TABLE erreurs_analyse (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    copie_id UUID NOT NULL REFERENCES copies_examens(id) ON DELETE CASCADE,
    type_erreur VARCHAR(50) NOT NULL, -- 'comprehension', 'methode', 'inattention'
    texte_original TEXT,
    correction TEXT,
    explication TEXT NOT NULL,
    conseil TEXT,
    severite VARCHAR(20) DEFAULT 'moderate', -- 'legere', 'moderate', 'grave'
    position_debut INT,
    position_fin INT,
    date_creation TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index pour les requêtes par copie
CREATE INDEX idx_erreurs_analyse_copie_id ON erreurs_analyse(copie_id);

-- Index pour les statistiques par type d'erreur
CREATE INDEX idx_erreurs_analyse_type ON erreurs_analyse(type_erreur);
