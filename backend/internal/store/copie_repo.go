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

// CopieExamen représente une copie d'examen scannée avec son analyse
type CopieExamen struct {
	ID                    string           `json:"id"`
	CoursID               string           `json:"coursId,omitempty"`
	Titre                 string           `json:"titre"`
	Matiere               string           `json:"matiere,omitempty"`
	NoteObtenue           *float64         `json:"noteObtenue,omitempty"`
	NoteTotale            *float64         `json:"noteTotale,omitempty"`
	TexteOCR              string           `json:"texteOCR"`
	AnnotationsProfesseur string           `json:"annotationsProfesseur,omitempty"`
	Confiance             float64          `json:"confiance"`
	ZonesIncertaines      []ZoneIncertaine `json:"zonesIncertaines"`
	FichiersOriginaux     []string         `json:"fichiersOriginaux"`
	DateExamen            *time.Time       `json:"dateExamen,omitempty"`
	DateCreation          time.Time        `json:"dateCreation"`
	DateModification      time.Time        `json:"dateModification"`
}

// ErreurAnalyse représente une erreur identifiée dans une copie d'examen
type ErreurAnalyse struct {
	ID            string    `json:"id"`
	CopieID       string    `json:"copieId"`
	TypeErreur    string    `json:"typeErreur"` // 'comprehension', 'methode', 'inattention'
	TexteOriginal string    `json:"texteOriginal,omitempty"`
	Correction    string    `json:"correction,omitempty"`
	Explication   string    `json:"explication"`
	Conseil       string    `json:"conseil,omitempty"`
	Severite      string    `json:"severite"` // 'legere', 'moderate', 'grave'
	PositionDebut *int      `json:"positionDebut,omitempty"`
	PositionFin   *int      `json:"positionFin,omitempty"`
	DateCreation  time.Time `json:"dateCreation"`
}

// CopieExamenRepository définit les opérations CRUD pour les copies d'examens
type CopieExamenRepository interface {
	Creer(ctx context.Context, copie *CopieExamen) error
	ObtenirParID(ctx context.Context, id string) (*CopieExamen, error)
	Lister(ctx context.Context, limite, offset int) ([]*CopieExamen, error)
	ListerParCours(ctx context.Context, coursID string) ([]*CopieExamen, error)
	MettreAJour(ctx context.Context, copie *CopieExamen) error
	Supprimer(ctx context.Context, id string) error
	Compter(ctx context.Context) (int, error)
}

// ErreurAnalyseRepository définit les opérations CRUD pour les erreurs d'analyse
type ErreurAnalyseRepository interface {
	CreerPlusieurs(ctx context.Context, erreurs []*ErreurAnalyse) error
	ListerParCopie(ctx context.Context, copieID string) ([]*ErreurAnalyse, error)
	SupprimerParCopie(ctx context.Context, copieID string) error
	CompterParType(ctx context.Context, copieID string) (map[string]int, error)
}

// CopieExamenRepo implémente CopieExamenRepository avec PostgreSQL
type CopieExamenRepo struct {
	db *sql.DB
}

// ErreurAnalyseRepo implémente ErreurAnalyseRepository avec PostgreSQL
type ErreurAnalyseRepo struct {
	db *sql.DB
}

// NouveauCopieExamenRepo crée un nouveau repository pour les copies d'examens
func NouveauCopieExamenRepo(store *Store) *CopieExamenRepo {
	return &CopieExamenRepo{db: store.DB()}
}

// NouveauErreurAnalyseRepo crée un nouveau repository pour les erreurs d'analyse
func NouveauErreurAnalyseRepo(store *Store) *ErreurAnalyseRepo {
	return &ErreurAnalyseRepo{db: store.DB()}
}

// Vérification que CopieExamenRepo implémente CopieExamenRepository
var _ CopieExamenRepository = (*CopieExamenRepo)(nil)

// Vérification que ErreurAnalyseRepo implémente ErreurAnalyseRepository
var _ ErreurAnalyseRepository = (*ErreurAnalyseRepo)(nil)

// Creer insère une nouvelle copie d'examen dans la base de données
func (r *CopieExamenRepo) Creer(ctx context.Context, copie *CopieExamen) error {
	if copie.ID == "" {
		copie.ID = uuid.New().String()
	}

	now := time.Now()
	copie.DateCreation = now
	copie.DateModification = now

	zonesJSON, err := json.Marshal(copie.ZonesIncertaines)
	if err != nil {
		return fmt.Errorf("erreur sérialisation zones incertaines: %w", err)
	}

	fichiersJSON, err := json.Marshal(copie.FichiersOriginaux)
	if err != nil {
		return fmt.Errorf("erreur sérialisation fichiers originaux: %w", err)
	}

	query := `
		INSERT INTO copies_examens (id, cours_id, titre, matiere, note_obtenue, note_totale, texte_ocr, annotations_professeur, confiance, zones_incertaines, fichiers_originaux, date_examen, date_creation, date_modification)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err = r.db.ExecContext(ctx, query,
		copie.ID,
		nullString(copie.CoursID),
		copie.Titre,
		nullString(copie.Matiere),
		nullFloat64(copie.NoteObtenue),
		nullFloat64(copie.NoteTotale),
		copie.TexteOCR,
		nullString(copie.AnnotationsProfesseur),
		copie.Confiance,
		zonesJSON,
		fichiersJSON,
		nullTime(copie.DateExamen),
		copie.DateCreation,
		copie.DateModification,
	)
	if err != nil {
		return fmt.Errorf("erreur création copie examen: %w", err)
	}

	return nil
}

// ObtenirParID récupère une copie d'examen par son identifiant
func (r *CopieExamenRepo) ObtenirParID(ctx context.Context, id string) (*CopieExamen, error) {
	query := `
		SELECT id, cours_id, titre, matiere, note_obtenue, note_totale, texte_ocr, annotations_professeur, confiance, zones_incertaines, fichiers_originaux, date_examen, date_creation, date_modification
		FROM copies_examens
		WHERE id = $1
	`

	copie := &CopieExamen{}
	var coursID, matiere, annotations sql.NullString
	var noteObtenue, noteTotale sql.NullFloat64
	var dateExamen sql.NullTime
	var zonesJSON, fichiersJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&copie.ID,
		&coursID,
		&copie.Titre,
		&matiere,
		&noteObtenue,
		&noteTotale,
		&copie.TexteOCR,
		&annotations,
		&copie.Confiance,
		&zonesJSON,
		&fichiersJSON,
		&dateExamen,
		&copie.DateCreation,
		&copie.DateModification,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("copie examen non trouvée: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("erreur récupération copie examen: %w", err)
	}

	copie.CoursID = coursID.String
	copie.Matiere = matiere.String
	copie.AnnotationsProfesseur = annotations.String
	if noteObtenue.Valid {
		copie.NoteObtenue = &noteObtenue.Float64
	}
	if noteTotale.Valid {
		copie.NoteTotale = &noteTotale.Float64
	}
	if dateExamen.Valid {
		copie.DateExamen = &dateExamen.Time
	}

	if err := json.Unmarshal(zonesJSON, &copie.ZonesIncertaines); err != nil {
		return nil, fmt.Errorf("erreur désérialisation zones incertaines: %w", err)
	}

	if err := json.Unmarshal(fichiersJSON, &copie.FichiersOriginaux); err != nil {
		return nil, fmt.Errorf("erreur désérialisation fichiers originaux: %w", err)
	}

	return copie, nil
}

// Lister récupère une liste de copies d'examens avec pagination
func (r *CopieExamenRepo) Lister(ctx context.Context, limite, offset int) ([]*CopieExamen, error) {
	query := `
		SELECT id, cours_id, titre, matiere, note_obtenue, note_totale, texte_ocr, annotations_professeur, confiance, zones_incertaines, fichiers_originaux, date_examen, date_creation, date_modification
		FROM copies_examens
		ORDER BY date_creation DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limite, offset)
	if err != nil {
		return nil, fmt.Errorf("erreur liste copies examens: %w", err)
	}
	defer rows.Close()

	return scanCopies(rows)
}

// ListerParCours récupère les copies d'examens associées à un cours
func (r *CopieExamenRepo) ListerParCours(ctx context.Context, coursID string) ([]*CopieExamen, error) {
	query := `
		SELECT id, cours_id, titre, matiere, note_obtenue, note_totale, texte_ocr, annotations_professeur, confiance, zones_incertaines, fichiers_originaux, date_examen, date_creation, date_modification
		FROM copies_examens
		WHERE cours_id = $1
		ORDER BY date_examen DESC NULLS LAST, date_creation DESC
	`

	rows, err := r.db.QueryContext(ctx, query, coursID)
	if err != nil {
		return nil, fmt.Errorf("erreur liste copies par cours: %w", err)
	}
	defer rows.Close()

	return scanCopies(rows)
}

// scanCopies lit les lignes et retourne une liste de copies d'examens
func scanCopies(rows *sql.Rows) ([]*CopieExamen, error) {
	var copies []*CopieExamen
	for rows.Next() {
		copie := &CopieExamen{}
		var coursID, matiere, annotations sql.NullString
		var noteObtenue, noteTotale sql.NullFloat64
		var dateExamen sql.NullTime
		var zonesJSON, fichiersJSON []byte

		err := rows.Scan(
			&copie.ID,
			&coursID,
			&copie.Titre,
			&matiere,
			&noteObtenue,
			&noteTotale,
			&copie.TexteOCR,
			&annotations,
			&copie.Confiance,
			&zonesJSON,
			&fichiersJSON,
			&dateExamen,
			&copie.DateCreation,
			&copie.DateModification,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan copie examen: %w", err)
		}

		copie.CoursID = coursID.String
		copie.Matiere = matiere.String
		copie.AnnotationsProfesseur = annotations.String
		if noteObtenue.Valid {
			copie.NoteObtenue = &noteObtenue.Float64
		}
		if noteTotale.Valid {
			copie.NoteTotale = &noteTotale.Float64
		}
		if dateExamen.Valid {
			copie.DateExamen = &dateExamen.Time
		}

		if err := json.Unmarshal(zonesJSON, &copie.ZonesIncertaines); err != nil {
			return nil, fmt.Errorf("erreur désérialisation zones incertaines: %w", err)
		}

		if err := json.Unmarshal(fichiersJSON, &copie.FichiersOriginaux); err != nil {
			return nil, fmt.Errorf("erreur désérialisation fichiers originaux: %w", err)
		}

		copies = append(copies, copie)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur itération copies examens: %w", err)
	}

	return copies, nil
}

// MettreAJour met à jour une copie d'examen existante
func (r *CopieExamenRepo) MettreAJour(ctx context.Context, copie *CopieExamen) error {
	copie.DateModification = time.Now()

	zonesJSON, err := json.Marshal(copie.ZonesIncertaines)
	if err != nil {
		return fmt.Errorf("erreur sérialisation zones incertaines: %w", err)
	}

	fichiersJSON, err := json.Marshal(copie.FichiersOriginaux)
	if err != nil {
		return fmt.Errorf("erreur sérialisation fichiers originaux: %w", err)
	}

	query := `
		UPDATE copies_examens
		SET cours_id = $2, titre = $3, matiere = $4, note_obtenue = $5, note_totale = $6, texte_ocr = $7, annotations_professeur = $8, confiance = $9, zones_incertaines = $10, fichiers_originaux = $11, date_examen = $12, date_modification = $13
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		copie.ID,
		nullString(copie.CoursID),
		copie.Titre,
		nullString(copie.Matiere),
		nullFloat64(copie.NoteObtenue),
		nullFloat64(copie.NoteTotale),
		copie.TexteOCR,
		nullString(copie.AnnotationsProfesseur),
		copie.Confiance,
		zonesJSON,
		fichiersJSON,
		nullTime(copie.DateExamen),
		copie.DateModification,
	)
	if err != nil {
		return fmt.Errorf("erreur mise à jour copie examen: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification mise à jour: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("copie examen non trouvée: %s", copie.ID)
	}

	return nil
}

// Supprimer supprime une copie d'examen par son identifiant
func (r *CopieExamenRepo) Supprimer(ctx context.Context, id string) error {
	query := `DELETE FROM copies_examens WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression copie examen: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("copie examen non trouvée: %s", id)
	}

	return nil
}

// Compter retourne le nombre total de copies d'examens
func (r *CopieExamenRepo) Compter(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM copies_examens`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage copies examens: %w", err)
	}

	return count, nil
}

// CreerPlusieurs insère plusieurs erreurs d'analyse en une transaction
func (r *ErreurAnalyseRepo) CreerPlusieurs(ctx context.Context, erreurs []*ErreurAnalyse) error {
	if len(erreurs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erreur début transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO erreurs_analyse (id, copie_id, type_erreur, texte_original, correction, explication, conseil, severite, position_debut, position_fin, date_creation)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("erreur préparation requête: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, erreur := range erreurs {
		if erreur.ID == "" {
			erreur.ID = uuid.New().String()
		}
		erreur.DateCreation = now

		if erreur.Severite == "" {
			erreur.Severite = "moderate"
		}

		_, err = stmt.ExecContext(ctx,
			erreur.ID,
			erreur.CopieID,
			erreur.TypeErreur,
			nullString(erreur.TexteOriginal),
			nullString(erreur.Correction),
			erreur.Explication,
			nullString(erreur.Conseil),
			erreur.Severite,
			nullInt(erreur.PositionDebut),
			nullInt(erreur.PositionFin),
			erreur.DateCreation,
		)
		if err != nil {
			return fmt.Errorf("erreur insertion erreur analyse: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erreur commit transaction: %w", err)
	}

	return nil
}

// ListerParCopie récupère toutes les erreurs d'analyse d'une copie
func (r *ErreurAnalyseRepo) ListerParCopie(ctx context.Context, copieID string) ([]*ErreurAnalyse, error) {
	query := `
		SELECT id, copie_id, type_erreur, texte_original, correction, explication, conseil, severite, position_debut, position_fin, date_creation
		FROM erreurs_analyse
		WHERE copie_id = $1
		ORDER BY position_debut NULLS LAST, date_creation
	`

	rows, err := r.db.QueryContext(ctx, query, copieID)
	if err != nil {
		return nil, fmt.Errorf("erreur liste erreurs analyse: %w", err)
	}
	defer rows.Close()

	var erreurs []*ErreurAnalyse
	for rows.Next() {
		erreur := &ErreurAnalyse{}
		var texteOriginal, correction, conseil sql.NullString
		var positionDebut, positionFin sql.NullInt32

		err := rows.Scan(
			&erreur.ID,
			&erreur.CopieID,
			&erreur.TypeErreur,
			&texteOriginal,
			&correction,
			&erreur.Explication,
			&conseil,
			&erreur.Severite,
			&positionDebut,
			&positionFin,
			&erreur.DateCreation,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan erreur analyse: %w", err)
		}

		erreur.TexteOriginal = texteOriginal.String
		erreur.Correction = correction.String
		erreur.Conseil = conseil.String
		if positionDebut.Valid {
			val := int(positionDebut.Int32)
			erreur.PositionDebut = &val
		}
		if positionFin.Valid {
			val := int(positionFin.Int32)
			erreur.PositionFin = &val
		}

		erreurs = append(erreurs, erreur)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur itération erreurs analyse: %w", err)
	}

	return erreurs, nil
}

// SupprimerParCopie supprime toutes les erreurs d'analyse d'une copie
func (r *ErreurAnalyseRepo) SupprimerParCopie(ctx context.Context, copieID string) error {
	query := `DELETE FROM erreurs_analyse WHERE copie_id = $1`

	_, err := r.db.ExecContext(ctx, query, copieID)
	if err != nil {
		return fmt.Errorf("erreur suppression erreurs analyse: %w", err)
	}

	return nil
}

// CompterParType retourne le nombre d'erreurs par type pour une copie
func (r *ErreurAnalyseRepo) CompterParType(ctx context.Context, copieID string) (map[string]int, error) {
	query := `
		SELECT type_erreur, COUNT(*) as count
		FROM erreurs_analyse
		WHERE copie_id = $1
		GROUP BY type_erreur
	`

	rows, err := r.db.QueryContext(ctx, query, copieID)
	if err != nil {
		return nil, fmt.Errorf("erreur comptage erreurs par type: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var typeErreur string
		var count int
		if err := rows.Scan(&typeErreur, &count); err != nil {
			return nil, fmt.Errorf("erreur scan comptage: %w", err)
		}
		counts[typeErreur] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur itération comptage: %w", err)
	}

	return counts, nil
}

// nullFloat64 convertit un pointeur float64 en sql.NullFloat64
func nullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

// nullTime convertit un pointeur time.Time en sql.NullTime
func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// nullInt convertit un pointeur int en sql.NullInt32
func nullInt(i *int) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(*i), Valid: true}
}
