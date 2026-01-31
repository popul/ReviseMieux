-- Migration 008: Créer la table des quotas journaliers
CREATE TABLE quotas_journaliers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date DATE NOT NULL UNIQUE,
    pages_ocr INT DEFAULT 0,
    generations INT DEFAULT 0
);

-- Index pour les requêtes par date
CREATE INDEX idx_quotas_date ON quotas_journaliers(date);
