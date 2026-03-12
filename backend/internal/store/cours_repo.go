// Package store contient les accès à la base de données
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Cours représente un cours scanné avec son texte OCR
type Cours struct {
	ID                 string            `json:"id"`
	Titre              string            `json:"titre"`
	Matiere            string            `json:"matiere,omitempty"`
	TexteOCR           string            `json:"texteOCR"`
	TexteCorrige       string            `json:"texteCorrige,omitempty"`
	Confiance          float64           `json:"confiance"`
	ZonesIncertaines   []ZoneIncertaine  `json:"zonesIncertaines"`
	FichiersOriginaux  []string          `json:"fichiersOriginaux"`
	Images             []string          `json:"images"`
	BlocsTexte         json.RawMessage   `json:"blocsTexte,omitempty"`
	Resume             json.RawMessage   `json:"resume,omitempty"`
	StatutOCR          string            `json:"statutOCR"`
	PagesTraitees      int               `json:"pagesTraitees"`
	NombrePages        int               `json:"nombrePages"`
	ErreurOCR          string            `json:"erreurOCR,omitempty"`
	DateCreation       time.Time         `json:"dateCreation"`
	DateModification   time.Time         `json:"dateModification"`
}

// ZoneIncertaine représente une zone du texte OCR avec faible confiance
type ZoneIncertaine struct {
	Debut  int    `json:"debut"`
	Fin    int    `json:"fin"`
	Texte  string `json:"texte"`
	Raison string `json:"raison"`
}

// CoursRepository définit les opérations CRUD pour les cours
type CoursRepository interface {
	Creer(ctx context.Context, cours *Cours) error
	ObtenirParID(ctx context.Context, id string) (*Cours, error)
	Lister(ctx context.Context, limite, offset int) ([]*Cours, error)
	MettreAJour(ctx context.Context, cours *Cours) error
	Supprimer(ctx context.Context, id string) error
	Compter(ctx context.Context) (int, error)
	MettreAJourProgressionOCR(ctx context.Context, id string, pagesTraitees int, texteOCR string, confiance float64, zonesIncertaines []byte, blocsTexte []byte, titre string, matiere string) error
	TerminerOCR(ctx context.Context, id string, texteOCR string, confiance float64, zonesIncertaines []byte, blocsTexte []byte, titre string, matiere string) error
	EchouerOCR(ctx context.Context, id string, erreur string) error
	RecupererOCRBloques(ctx context.Context) (int, error)
}

// CoursRepo implémente CoursRepository avec PostgreSQL
type CoursRepo struct {
	db *sql.DB
}

// NouveauCoursRepo crée un nouveau repository pour les cours
func NouveauCoursRepo(store *Store) *CoursRepo {
	return &CoursRepo{db: store.DB()}
}

// Vérification que CoursRepo implémente CoursRepository
var _ CoursRepository = (*CoursRepo)(nil)

// Creer insère un nouveau cours dans la base de données
func (r *CoursRepo) Creer(ctx context.Context, cours *Cours) error {
	if cours.ID == "" {
		cours.ID = uuid.New().String()
	}

	now := time.Now()
	cours.DateCreation = now
	cours.DateModification = now

	zonesJSON, err := json.Marshal(cours.ZonesIncertaines)
	if err != nil {
		return fmt.Errorf("erreur sérialisation zones incertaines: %w", err)
	}

	fichiersJSON, err := json.Marshal(cours.FichiersOriginaux)
	if err != nil {
		return fmt.Errorf("erreur sérialisation fichiers originaux: %w", err)
	}

	imagesJSON, err := json.Marshal(cours.Images)
	if err != nil {
		return fmt.Errorf("erreur sérialisation images: %w", err)
	}

	blocsTexteJSON := cours.BlocsTexte
	if blocsTexteJSON == nil {
		blocsTexteJSON = json.RawMessage("[]")
	}

	// Préparer le resume (nil si non défini)
	var resumeParam interface{}
	if cours.Resume != nil {
		resumeParam = cours.Resume
	}

	// Valeurs par défaut pour le statut OCR
	statutOCR := cours.StatutOCR
	if statutOCR == "" {
		statutOCR = "termine"
	}

	query := `
		INSERT INTO cours (id, titre, matiere, texte_ocr, texte_corrige, confiance, zones_incertaines, fichiers_originaux, images, blocs_texte, date_creation, date_modification, resume, statut_ocr, pages_traitees, nombre_pages, erreur_ocr)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err = r.db.ExecContext(ctx, query,
		cours.ID,
		cours.Titre,
		nullString(cours.Matiere),
		cours.TexteOCR,
		nullString(cours.TexteCorrige),
		cours.Confiance,
		zonesJSON,
		fichiersJSON,
		imagesJSON,
		blocsTexteJSON,
		cours.DateCreation,
		cours.DateModification,
		resumeParam,
		statutOCR,
		cours.PagesTraitees,
		cours.NombrePages,
		nullString(cours.ErreurOCR),
	)
	if err != nil {
		return fmt.Errorf("erreur création cours: %w", err)
	}

	return nil
}

// ObtenirParID récupère un cours par son identifiant
func (r *CoursRepo) ObtenirParID(ctx context.Context, id string) (*Cours, error) {
	query := `
		SELECT id, titre, matiere, texte_ocr, texte_corrige, confiance, zones_incertaines, fichiers_originaux, COALESCE(images, '[]'::jsonb), COALESCE(blocs_texte, '[]'::jsonb), date_creation, date_modification, COALESCE(resume, 'null'::jsonb), statut_ocr, pages_traitees, nombre_pages, erreur_ocr
		FROM cours
		WHERE id = $1
	`

	cours := &Cours{}
	var matiere, texteCorrige, erreurOCR sql.NullString
	var zonesJSON, fichiersJSON, imagesJSON, blocsTexteJSON, resumeJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cours.ID,
		&cours.Titre,
		&matiere,
		&cours.TexteOCR,
		&texteCorrige,
		&cours.Confiance,
		&zonesJSON,
		&fichiersJSON,
		&imagesJSON,
		&blocsTexteJSON,
		&cours.DateCreation,
		&cours.DateModification,
		&resumeJSON,
		&cours.StatutOCR,
		&cours.PagesTraitees,
		&cours.NombrePages,
		&erreurOCR,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("cours non trouvé: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("erreur récupération cours: %w", err)
	}

	cours.Matiere = matiere.String
	cours.TexteCorrige = texteCorrige.String
	cours.ErreurOCR = erreurOCR.String

	if err := json.Unmarshal(zonesJSON, &cours.ZonesIncertaines); err != nil {
		return nil, fmt.Errorf("erreur désérialisation zones incertaines: %w", err)
	}

	if err := json.Unmarshal(fichiersJSON, &cours.FichiersOriginaux); err != nil {
		return nil, fmt.Errorf("erreur désérialisation fichiers originaux: %w", err)
	}

	if err := json.Unmarshal(imagesJSON, &cours.Images); err != nil {
		return nil, fmt.Errorf("erreur désérialisation images: %w", err)
	}

	// BlocsTexte est stocké comme RawMessage JSON, pas besoin de désérialiser
	if len(blocsTexteJSON) > 0 && string(blocsTexteJSON) != "[]" {
		cours.BlocsTexte = json.RawMessage(blocsTexteJSON)
	}

	// Resume est stocké comme RawMessage JSON
	if len(resumeJSON) > 0 && string(resumeJSON) != "null" {
		cours.Resume = json.RawMessage(resumeJSON)
	}

	return cours, nil
}

// Lister récupère une liste de cours avec pagination
func (r *CoursRepo) Lister(ctx context.Context, limite, offset int) ([]*Cours, error) {
	query := `
		SELECT id, titre, matiere, texte_ocr, texte_corrige, confiance, zones_incertaines, fichiers_originaux, COALESCE(images, '[]'::jsonb), COALESCE(blocs_texte, '[]'::jsonb), date_creation, date_modification, COALESCE(resume, 'null'::jsonb), statut_ocr, pages_traitees, nombre_pages, erreur_ocr
		FROM cours
		ORDER BY date_creation DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limite, offset)
	if err != nil {
		return nil, fmt.Errorf("erreur liste cours: %w", err)
	}
	defer rows.Close()

	var coursList []*Cours
	for rows.Next() {
		cours := &Cours{}
		var matiere, texteCorrige, erreurOCR sql.NullString
		var zonesJSON, fichiersJSON, imagesJSON, blocsTexteJSON, resumeJSON []byte

		err := rows.Scan(
			&cours.ID,
			&cours.Titre,
			&matiere,
			&cours.TexteOCR,
			&texteCorrige,
			&cours.Confiance,
			&zonesJSON,
			&fichiersJSON,
			&imagesJSON,
			&blocsTexteJSON,
			&cours.DateCreation,
			&cours.DateModification,
			&resumeJSON,
			&cours.StatutOCR,
			&cours.PagesTraitees,
			&cours.NombrePages,
			&erreurOCR,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan cours: %w", err)
		}

		cours.Matiere = matiere.String
		cours.TexteCorrige = texteCorrige.String
		cours.ErreurOCR = erreurOCR.String

		if err := json.Unmarshal(zonesJSON, &cours.ZonesIncertaines); err != nil {
			return nil, fmt.Errorf("erreur désérialisation zones incertaines: %w", err)
		}

		if err := json.Unmarshal(fichiersJSON, &cours.FichiersOriginaux); err != nil {
			return nil, fmt.Errorf("erreur désérialisation fichiers originaux: %w", err)
		}

		if err := json.Unmarshal(imagesJSON, &cours.Images); err != nil {
			return nil, fmt.Errorf("erreur désérialisation images: %w", err)
		}

		// BlocsTexte est stocké comme RawMessage JSON
		if len(blocsTexteJSON) > 0 && string(blocsTexteJSON) != "[]" {
			cours.BlocsTexte = json.RawMessage(blocsTexteJSON)
		}

		// Resume est stocké comme RawMessage JSON
		if len(resumeJSON) > 0 && string(resumeJSON) != "null" {
			cours.Resume = json.RawMessage(resumeJSON)
		}

		coursList = append(coursList, cours)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur itération cours: %w", err)
	}

	return coursList, nil
}

// MettreAJour met à jour un cours existant
func (r *CoursRepo) MettreAJour(ctx context.Context, cours *Cours) error {
	cours.DateModification = time.Now()

	zonesJSON, err := json.Marshal(cours.ZonesIncertaines)
	if err != nil {
		return fmt.Errorf("erreur sérialisation zones incertaines: %w", err)
	}

	fichiersJSON, err := json.Marshal(cours.FichiersOriginaux)
	if err != nil {
		return fmt.Errorf("erreur sérialisation fichiers originaux: %w", err)
	}

	imagesJSON, err := json.Marshal(cours.Images)
	if err != nil {
		return fmt.Errorf("erreur sérialisation images: %w", err)
	}

	blocsTexteJSON := cours.BlocsTexte
	if blocsTexteJSON == nil {
		blocsTexteJSON = json.RawMessage("[]")
	}

	// Préparer le resume (nil si non défini)
	var resumeParam interface{}
	if cours.Resume != nil {
		resumeParam = cours.Resume
	}

	statutOCR := cours.StatutOCR
	if statutOCR == "" {
		statutOCR = "termine"
	}

	query := `
		UPDATE cours
		SET titre = $2, matiere = $3, texte_ocr = $4, texte_corrige = $5, confiance = $6, zones_incertaines = $7, fichiers_originaux = $8, images = $9, blocs_texte = $10, resume = $11, date_modification = $12, statut_ocr = $13, pages_traitees = $14, nombre_pages = $15, erreur_ocr = $16
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		cours.ID,
		cours.Titre,
		nullString(cours.Matiere),
		cours.TexteOCR,
		nullString(cours.TexteCorrige),
		cours.Confiance,
		zonesJSON,
		fichiersJSON,
		imagesJSON,
		blocsTexteJSON,
		resumeParam,
		cours.DateModification,
		statutOCR,
		cours.PagesTraitees,
		cours.NombrePages,
		nullString(cours.ErreurOCR),
	)
	if err != nil {
		return fmt.Errorf("erreur mise à jour cours: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification mise à jour: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cours non trouvé: %s", cours.ID)
	}

	return nil
}

// Supprimer supprime un cours par son identifiant
func (r *CoursRepo) Supprimer(ctx context.Context, id string) error {
	query := `DELETE FROM cours WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression cours: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cours non trouvé: %s", id)
	}

	return nil
}

// Compter retourne le nombre total de cours
func (r *CoursRepo) Compter(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM cours`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage cours: %w", err)
	}

	return count, nil
}

// MettreAJourProgressionOCR met à jour la progression OCR (page par page)
func (r *CoursRepo) MettreAJourProgressionOCR(ctx context.Context, id string, pagesTraitees int, texteOCR string, confiance float64, zonesIncertaines []byte, blocsTexte []byte, titre string, matiere string) error {
	query := `
		UPDATE cours
		SET pages_traitees = $2, texte_ocr = $3, confiance = $4, zones_incertaines = $5, blocs_texte = $6, titre = $7, matiere = $8, date_modification = $9
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, pagesTraitees, texteOCR, confiance, zonesIncertaines, blocsTexte, titre, nullString(matiere), time.Now())
	if err != nil {
		return fmt.Errorf("erreur mise à jour progression OCR: %w", err)
	}
	return nil
}

// TerminerOCR marque l'OCR comme terminé et met à jour les données finales
func (r *CoursRepo) TerminerOCR(ctx context.Context, id string, texteOCR string, confiance float64, zonesIncertaines []byte, blocsTexte []byte, titre string, matiere string) error {
	query := `
		UPDATE cours
		SET statut_ocr = 'termine', texte_ocr = $2, confiance = $3, zones_incertaines = $4, blocs_texte = $5, titre = $6, matiere = $7, pages_traitees = nombre_pages, date_modification = $8
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, texteOCR, confiance, zonesIncertaines, blocsTexte, titre, nullString(matiere), time.Now())
	if err != nil {
		return fmt.Errorf("erreur finalisation OCR: %w", err)
	}
	return nil
}

// EchouerOCR marque l'OCR comme échoué avec un message d'erreur
func (r *CoursRepo) EchouerOCR(ctx context.Context, id string, erreur string) error {
	query := `
		UPDATE cours
		SET statut_ocr = 'erreur', erreur_ocr = $2, date_modification = $3
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, erreur, time.Now())
	if err != nil {
		return fmt.Errorf("erreur marquage échec OCR: %w", err)
	}
	return nil
}

// RecupererOCRBloques remet les cours bloqués en statut 'termine' au redémarrage du serveur.
// Les goroutines OCR en cours ont été tuées, on conserve les données partielles.
func (r *CoursRepo) RecupererOCRBloques(ctx context.Context) (int, error) {
	query := `
		UPDATE cours
		SET statut_ocr = 'termine', pages_traitees = nombre_pages, date_modification = $1
		WHERE statut_ocr = 'en_cours'
	`
	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("erreur récupération OCR bloqués: %w", err)
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

// nullString convertit une chaîne en sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
