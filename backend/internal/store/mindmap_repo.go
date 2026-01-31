// Package store contient les accès à la base de données
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TypeNoeud représente le type d'un nœud de mindmap
type TypeNoeud string

const (
	TypeNoeudCentral TypeNoeud = "central"
	TypeNoeudBranche TypeNoeud = "branche"
	TypeNoeudFeuille TypeNoeud = "feuille"
)

// Position représente la position d'un nœud dans la mindmap
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NoeudMindmap représente un nœud dans une mindmap
type NoeudMindmap struct {
	ID       string    `json:"id"`
	Label    string    `json:"label"`
	Type     TypeNoeud `json:"type"`
	Position Position  `json:"position"`
}

// LienMindmap représente un lien entre deux nœuds
type LienMindmap struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// Mindmap représente une carte mentale générée
type Mindmap struct {
	ID           string         `json:"id"`
	CoursID      string         `json:"coursId"`
	Noeuds       []NoeudMindmap `json:"noeuds"`
	Liens        []LienMindmap  `json:"liens"`
	DateCreation time.Time      `json:"dateCreation"`
}

// MindmapRepository définit les opérations sur les mindmaps
type MindmapRepository interface {
	// Creer crée une nouvelle mindmap
	Creer(ctx context.Context, mindmap *Mindmap) error
	// ObtenirParID récupère une mindmap par son ID
	ObtenirParID(ctx context.Context, id string) (*Mindmap, error)
	// ObtenirParCours récupère la mindmap d'un cours (une seule par cours)
	ObtenirParCours(ctx context.Context, coursID string) (*Mindmap, error)
	// Supprimer supprime une mindmap par son ID
	Supprimer(ctx context.Context, id string) error
	// SupprimerParCours supprime toutes les mindmaps d'un cours
	SupprimerParCours(ctx context.Context, coursID string) error
	// Compter compte le nombre total de mindmaps
	Compter(ctx context.Context) (int, error)
	// CompterParCours compte les mindmaps d'un cours
	CompterParCours(ctx context.Context, coursID string) (int, error)
}

// MindmapRepo implémente MindmapRepository avec PostgreSQL
type MindmapRepo struct {
	store *Store
}

// NouveauMindmapRepo crée un nouveau repository mindmap
func NouveauMindmapRepo(s *Store) *MindmapRepo {
	return &MindmapRepo{store: s}
}

// Creer crée une nouvelle mindmap
func (r *MindmapRepo) Creer(ctx context.Context, mindmap *Mindmap) error {
	if mindmap.ID == "" {
		mindmap.ID = uuid.New().String()
	}
	mindmap.DateCreation = time.Now()

	// Sérialiser les nœuds et liens en JSON
	noeudsJSON, err := json.Marshal(mindmap.Noeuds)
	if err != nil {
		return fmt.Errorf("erreur sérialisation noeuds: %w", err)
	}

	liensJSON, err := json.Marshal(mindmap.Liens)
	if err != nil {
		return fmt.Errorf("erreur sérialisation liens: %w", err)
	}

	query := `
		INSERT INTO mindmaps (id, cours_id, noeuds, liens, date_creation)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = r.store.db.ExecContext(ctx, query,
		mindmap.ID,
		mindmap.CoursID,
		noeudsJSON,
		liensJSON,
		mindmap.DateCreation,
	)

	if err != nil {
		return fmt.Errorf("erreur création mindmap: %w", err)
	}

	return nil
}

// ObtenirParID récupère une mindmap par son ID
func (r *MindmapRepo) ObtenirParID(ctx context.Context, id string) (*Mindmap, error) {
	query := `
		SELECT id, cours_id, noeuds, liens, date_creation
		FROM mindmaps
		WHERE id = $1
	`

	var mindmap Mindmap
	var noeudsJSON, liensJSON []byte

	err := r.store.db.QueryRowContext(ctx, query, id).Scan(
		&mindmap.ID,
		&mindmap.CoursID,
		&noeudsJSON,
		&liensJSON,
		&mindmap.DateCreation,
	)

	if err != nil {
		return nil, fmt.Errorf("erreur récupération mindmap: %w", err)
	}

	// Désérialiser les nœuds et liens
	if err := json.Unmarshal(noeudsJSON, &mindmap.Noeuds); err != nil {
		return nil, fmt.Errorf("erreur désérialisation noeuds: %w", err)
	}

	if err := json.Unmarshal(liensJSON, &mindmap.Liens); err != nil {
		return nil, fmt.Errorf("erreur désérialisation liens: %w", err)
	}

	return &mindmap, nil
}

// ObtenirParCours récupère la mindmap d'un cours (la plus récente si plusieurs)
func (r *MindmapRepo) ObtenirParCours(ctx context.Context, coursID string) (*Mindmap, error) {
	query := `
		SELECT id, cours_id, noeuds, liens, date_creation
		FROM mindmaps
		WHERE cours_id = $1
		ORDER BY date_creation DESC
		LIMIT 1
	`

	var mindmap Mindmap
	var noeudsJSON, liensJSON []byte

	err := r.store.db.QueryRowContext(ctx, query, coursID).Scan(
		&mindmap.ID,
		&mindmap.CoursID,
		&noeudsJSON,
		&liensJSON,
		&mindmap.DateCreation,
	)

	if err != nil {
		return nil, fmt.Errorf("erreur récupération mindmap: %w", err)
	}

	// Désérialiser les nœuds et liens
	if err := json.Unmarshal(noeudsJSON, &mindmap.Noeuds); err != nil {
		return nil, fmt.Errorf("erreur désérialisation noeuds: %w", err)
	}

	if err := json.Unmarshal(liensJSON, &mindmap.Liens); err != nil {
		return nil, fmt.Errorf("erreur désérialisation liens: %w", err)
	}

	return &mindmap, nil
}

// Supprimer supprime une mindmap par son ID
func (r *MindmapRepo) Supprimer(ctx context.Context, id string) error {
	query := `DELETE FROM mindmaps WHERE id = $1`
	result, err := r.store.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression mindmap: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("mindmap non trouvée")
	}

	return nil
}

// SupprimerParCours supprime toutes les mindmaps d'un cours
func (r *MindmapRepo) SupprimerParCours(ctx context.Context, coursID string) error {
	query := `DELETE FROM mindmaps WHERE cours_id = $1`
	_, err := r.store.db.ExecContext(ctx, query, coursID)
	if err != nil {
		return fmt.Errorf("erreur suppression mindmaps du cours: %w", err)
	}

	return nil
}

// Compter compte le nombre total de mindmaps
func (r *MindmapRepo) Compter(ctx context.Context) (int, error) {
	var count int
	err := r.store.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mindmaps").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage mindmaps: %w", err)
	}
	return count, nil
}

// CompterParCours compte les mindmaps d'un cours
func (r *MindmapRepo) CompterParCours(ctx context.Context, coursID string) (int, error) {
	var count int
	err := r.store.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM mindmaps WHERE cours_id = $1",
		coursID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage mindmaps du cours: %w", err)
	}
	return count, nil
}
