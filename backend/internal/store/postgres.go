// Package store contient les accès à la base de données
package store

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Store représente la connexion à la base de données
type Store struct {
	db *sql.DB
}

// NouveauStore crée une nouvelle connexion à la base de données
func NouveauStore(databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erreur connexion base de données: %w", err)
	}

	// Vérifier la connexion
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erreur ping base de données: %w", err)
	}

	return &Store{db: db}, nil
}

// Fermer ferme la connexion à la base de données
func (s *Store) Fermer() error {
	return s.db.Close()
}

// Ping vérifie que la connexion est active
func (s *Store) Ping() error {
	return s.db.Ping()
}

// DB retourne la connexion brute pour les transactions complexes
func (s *Store) DB() *sql.DB {
	return s.db
}
