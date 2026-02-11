// Package store contient les acces a la base de donnees
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// QuestionExamen represente une question d'examen blanc
type QuestionExamen struct {
	Numero          int             `json:"numero"`
	Type            string          `json:"type"` // "definition", "comprehension", "application", "synthese"
	Difficulte      string          `json:"difficulte"` // "facile", "moyen", "difficile"
	Enonce          string          `json:"enonce"`
	Bareme          int             `json:"bareme"`
	ReponseAttendue string          `json:"reponse_attendue"`
	Indices         []IndiceExamen  `json:"indices"`
}

// IndiceExamen represente un indice progressif pour une question
type IndiceExamen struct {
	Niveau int    `json:"niveau"` // 1, 2, 3
	Texte  string `json:"texte"`
}

// ExamenBlanc represente un examen blanc genere a partir d'un cours
type ExamenBlanc struct {
	ID            string            `json:"id"`
	CoursID       string            `json:"coursId"`
	Questions     []QuestionExamen  `json:"questions"`
	DureeMinutes  int               `json:"dureeMinutes"`
	DateCreation  time.Time         `json:"dateCreation"`
}

// ReponseExamen represente la reponse d'un eleve a une question
type ReponseExamen struct {
	QuestionNumero int    `json:"questionNumero"`
	Texte          string `json:"texte"`
}

// IndiceUtilise represente un indice utilise par l'eleve
type IndiceUtilise struct {
	QuestionNumero int `json:"questionNumero"`
	NiveauIndice   int `json:"niveauIndice"`
}

// PointFort represente un point fort identifie
type PointFort struct {
	Concept     string `json:"concept"`
	Commentaire string `json:"commentaire"`
}

// PointFaible represente un point faible identifie
type PointFaible struct {
	Concept     string `json:"concept"`
	Commentaire string `json:"commentaire"`
}

// EtapePlanRevision represente une etape du plan de revision
type EtapePlanRevision struct {
	Priorite    int    `json:"priorite"`
	Action      string `json:"action"`
	Concept     string `json:"concept"`
	Ressource   string `json:"ressource,omitempty"`
}

// SessionExamen represente une session d'examen blanc
type SessionExamen struct {
	ID             string              `json:"id"`
	ExamenID       string              `json:"examenId"`
	Reponses       []ReponseExamen     `json:"reponses"`
	IndicesUtilises []IndiceUtilise    `json:"indicesUtilises"`
	NoteEstimee    *float64            `json:"noteEstimee,omitempty"`
	PointsForts    []PointFort         `json:"pointsForts,omitempty"`
	PointsFaibles  []PointFaible       `json:"pointsFaibles,omitempty"`
	PlanRevision   []EtapePlanRevision `json:"planRevision,omitempty"`
	Termine        bool                `json:"termine"`
	DateDebut      time.Time           `json:"dateDebut"`
	DateFin        *time.Time          `json:"dateFin,omitempty"`
}

// ExamenRepository definit les operations pour les examens blancs
type ExamenRepository interface {
	CreerExamen(ctx context.Context, examen *ExamenBlanc) error
	ObtenirExamen(ctx context.Context, id string) (*ExamenBlanc, error)
	ListerExamensParCours(ctx context.Context, coursID string) ([]*ExamenBlanc, error)

	CreerSession(ctx context.Context, session *SessionExamen) error
	ObtenirSession(ctx context.Context, id string) (*SessionExamen, error)
	MettreAJourSession(ctx context.Context, session *SessionExamen) error
}

// ExamenRepo implemente ExamenRepository avec PostgreSQL
type ExamenRepo struct {
	db *sql.DB
}

// NouveauExamenRepo cree un nouveau repository pour les examens
func NouveauExamenRepo(store *Store) *ExamenRepo {
	return &ExamenRepo{db: store.DB()}
}

// Verification que ExamenRepo implemente ExamenRepository
var _ ExamenRepository = (*ExamenRepo)(nil)

// CreerExamen insere un nouvel examen blanc
func (r *ExamenRepo) CreerExamen(ctx context.Context, examen *ExamenBlanc) error {
	if examen.ID == "" {
		examen.ID = uuid.New().String()
	}
	examen.DateCreation = time.Now()

	questionsJSON, err := json.Marshal(examen.Questions)
	if err != nil {
		return fmt.Errorf("erreur serialisation questions: %w", err)
	}

	query := `
		INSERT INTO examens_blancs (id, cours_id, questions, duree_minutes, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = r.db.ExecContext(ctx, query,
		examen.ID,
		examen.CoursID,
		questionsJSON,
		examen.DureeMinutes,
		examen.DateCreation,
	)
	if err != nil {
		return fmt.Errorf("erreur insertion examen: %w", err)
	}

	return nil
}

// ObtenirExamen recupere un examen par son ID
func (r *ExamenRepo) ObtenirExamen(ctx context.Context, id string) (*ExamenBlanc, error) {
	query := `
		SELECT id, cours_id, questions, duree_minutes, created_at
		FROM examens_blancs
		WHERE id = $1
	`

	examen := &ExamenBlanc{}
	var questionsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&examen.ID,
		&examen.CoursID,
		&questionsJSON,
		&examen.DureeMinutes,
		&examen.DateCreation,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur recuperation examen: %w", err)
	}

	if err := json.Unmarshal(questionsJSON, &examen.Questions); err != nil {
		return nil, fmt.Errorf("erreur deserialisation questions: %w", err)
	}

	return examen, nil
}

// ListerExamensParCours recupere tous les examens d'un cours
func (r *ExamenRepo) ListerExamensParCours(ctx context.Context, coursID string) ([]*ExamenBlanc, error) {
	query := `
		SELECT id, cours_id, questions, duree_minutes, created_at
		FROM examens_blancs
		WHERE cours_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, coursID)
	if err != nil {
		return nil, fmt.Errorf("erreur requete examens: %w", err)
	}
	defer rows.Close()

	var examens []*ExamenBlanc
	for rows.Next() {
		examen := &ExamenBlanc{}
		var questionsJSON []byte

		err := rows.Scan(
			&examen.ID,
			&examen.CoursID,
			&questionsJSON,
			&examen.DureeMinutes,
			&examen.DateCreation,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan examen: %w", err)
		}

		if err := json.Unmarshal(questionsJSON, &examen.Questions); err != nil {
			return nil, fmt.Errorf("erreur deserialisation questions: %w", err)
		}

		examens = append(examens, examen)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur iteration examens: %w", err)
	}

	return examens, nil
}

// --- Sessions ---

// CreerSession cree une nouvelle session d'examen
func (r *ExamenRepo) CreerSession(ctx context.Context, session *SessionExamen) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.DateDebut = time.Now()
	session.Reponses = []ReponseExamen{}
	session.IndicesUtilises = []IndiceUtilise{}
	session.Termine = false

	reponsesJSON, err := json.Marshal(session.Reponses)
	if err != nil {
		return fmt.Errorf("erreur serialisation reponses: %w", err)
	}

	indicesJSON, err := json.Marshal(session.IndicesUtilises)
	if err != nil {
		return fmt.Errorf("erreur serialisation indices: %w", err)
	}

	query := `
		INSERT INTO sessions_examen (id, examen_id, reponses, indices_utilises, termine, started_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = r.db.ExecContext(ctx, query,
		session.ID,
		session.ExamenID,
		reponsesJSON,
		indicesJSON,
		session.Termine,
		session.DateDebut,
	)
	if err != nil {
		return fmt.Errorf("erreur insertion session examen: %w", err)
	}

	return nil
}

// ObtenirSession recupere une session par son ID
func (r *ExamenRepo) ObtenirSession(ctx context.Context, id string) (*SessionExamen, error) {
	query := `
		SELECT id, examen_id, reponses, indices_utilises, note_estimee,
			   points_forts, points_faibles, plan_revision, termine, started_at, completed_at
		FROM sessions_examen
		WHERE id = $1
	`

	session := &SessionExamen{}
	var reponsesJSON, indicesJSON []byte
	var pointsFortsJSON, pointsFaiblesJSON, planRevisionJSON sql.NullString
	var noteEstimee sql.NullFloat64
	var dateFin sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.ExamenID,
		&reponsesJSON,
		&indicesJSON,
		&noteEstimee,
		&pointsFortsJSON,
		&pointsFaiblesJSON,
		&planRevisionJSON,
		&session.Termine,
		&session.DateDebut,
		&dateFin,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur recuperation session examen: %w", err)
	}

	if reponsesJSON != nil {
		if err := json.Unmarshal(reponsesJSON, &session.Reponses); err != nil {
			return nil, fmt.Errorf("erreur deserialisation reponses: %w", err)
		}
	}
	if session.Reponses == nil {
		session.Reponses = []ReponseExamen{}
	}

	if indicesJSON != nil {
		if err := json.Unmarshal(indicesJSON, &session.IndicesUtilises); err != nil {
			return nil, fmt.Errorf("erreur deserialisation indices: %w", err)
		}
	}
	if session.IndicesUtilises == nil {
		session.IndicesUtilises = []IndiceUtilise{}
	}

	if noteEstimee.Valid {
		session.NoteEstimee = &noteEstimee.Float64
	}
	if dateFin.Valid {
		session.DateFin = &dateFin.Time
	}

	if pointsFortsJSON.Valid {
		if err := json.Unmarshal([]byte(pointsFortsJSON.String), &session.PointsForts); err != nil {
			return nil, fmt.Errorf("erreur deserialisation points forts: %w", err)
		}
	}
	if pointsFaiblesJSON.Valid {
		if err := json.Unmarshal([]byte(pointsFaiblesJSON.String), &session.PointsFaibles); err != nil {
			return nil, fmt.Errorf("erreur deserialisation points faibles: %w", err)
		}
	}
	if planRevisionJSON.Valid {
		if err := json.Unmarshal([]byte(planRevisionJSON.String), &session.PlanRevision); err != nil {
			return nil, fmt.Errorf("erreur deserialisation plan revision: %w", err)
		}
	}

	return session, nil
}

// MettreAJourSession met a jour une session existante
func (r *ExamenRepo) MettreAJourSession(ctx context.Context, session *SessionExamen) error {
	reponsesJSON, err := json.Marshal(session.Reponses)
	if err != nil {
		return fmt.Errorf("erreur serialisation reponses: %w", err)
	}

	indicesJSON, err := json.Marshal(session.IndicesUtilises)
	if err != nil {
		return fmt.Errorf("erreur serialisation indices: %w", err)
	}

	var pointsFortsJSON, pointsFaiblesJSON, planRevisionJSON []byte

	if session.PointsForts != nil {
		pointsFortsJSON, err = json.Marshal(session.PointsForts)
		if err != nil {
			return fmt.Errorf("erreur serialisation points forts: %w", err)
		}
	}
	if session.PointsFaibles != nil {
		pointsFaiblesJSON, err = json.Marshal(session.PointsFaibles)
		if err != nil {
			return fmt.Errorf("erreur serialisation points faibles: %w", err)
		}
	}
	if session.PlanRevision != nil {
		planRevisionJSON, err = json.Marshal(session.PlanRevision)
		if err != nil {
			return fmt.Errorf("erreur serialisation plan revision: %w", err)
		}
	}

	query := `
		UPDATE sessions_examen
		SET reponses = $1, indices_utilises = $2, note_estimee = $3,
			points_forts = $4, points_faibles = $5, plan_revision = $6,
			termine = $7, completed_at = $8
		WHERE id = $9
	`

	var notePtr *float64
	if session.NoteEstimee != nil {
		notePtr = session.NoteEstimee
	}

	_, err = r.db.ExecContext(ctx, query,
		reponsesJSON,
		indicesJSON,
		notePtr,
		nullBytes(pointsFortsJSON),
		nullBytes(pointsFaiblesJSON),
		nullBytes(planRevisionJSON),
		session.Termine,
		session.DateFin,
		session.ID,
	)
	if err != nil {
		return fmt.Errorf("erreur mise a jour session examen: %w", err)
	}

	return nil
}

// nullBytes convertit un slice de bytes en sql.NullString pour les colonnes JSONB
func nullBytes(b []byte) sql.NullString {
	if b == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: string(b), Valid: true}
}
