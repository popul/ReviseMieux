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

// Question représente une question de quiz QCM
type Question struct {
	ID              string   `json:"id"`
	Enonce          string   `json:"enonce"`
	Choix           []string `json:"choix"`
	ReponseCorrecte int      `json:"reponseCorrecte"` // Index du choix correct (0-3)
	Explication     string   `json:"explication"`
}

// Quiz représente un quiz généré à partir d'un cours
type Quiz struct {
	ID               string     `json:"id"`
	CoursID          string     `json:"coursId"`
	Titre            string     `json:"titre"`
	Difficulte       string     `json:"difficulte"` // "facile", "moyen", "difficile"
	NombreQuestions  int        `json:"nombreQuestions"`
	Questions        []Question `json:"questions"`
	DateCreation     time.Time  `json:"dateCreation"`
}

// ReponseSession représente une réponse à une question dans une session
type ReponseSession struct {
	QuestionID  string `json:"questionId"`
	ChoixIndex  int    `json:"choixIndex"`
	EstCorrecte bool   `json:"estCorrecte"`
}

// QuizSession représente une session de quiz (une tentative)
type QuizSession struct {
	ID        string           `json:"id"`
	QuizID    string           `json:"quizId"`
	Reponses  []ReponseSession `json:"reponses"`
	Score     *float64         `json:"score,omitempty"` // Pourcentage (0-100)
	Termine   bool             `json:"termine"`
	DateDebut time.Time        `json:"dateDebut"`
	DateFin   *time.Time       `json:"dateFin,omitempty"`
}

// QuizRepository définit les opérations pour les quiz
type QuizRepository interface {
	Creer(ctx context.Context, quiz *Quiz) error
	ObtenirParID(ctx context.Context, id string) (*Quiz, error)
	ListerParCours(ctx context.Context, coursID string) ([]*Quiz, error)
	Compter(ctx context.Context) (int, error)
	Supprimer(ctx context.Context, id string) error

	// Sessions
	CreerSession(ctx context.Context, session *QuizSession) error
	ObtenirSession(ctx context.Context, id string) (*QuizSession, error)
	MettreAJourSession(ctx context.Context, session *QuizSession) error

	// Statistiques
	CompterQuizCompletes(ctx context.Context) (int, error)
	ScoreMoyen(ctx context.Context) (float64, error)
}

// QuizRepo implémente QuizRepository avec PostgreSQL
type QuizRepo struct {
	db *sql.DB
}

// NouveauQuizRepo crée un nouveau repository pour les quiz
func NouveauQuizRepo(store *Store) *QuizRepo {
	return &QuizRepo{db: store.DB()}
}

// Vérification que QuizRepo implémente QuizRepository
var _ QuizRepository = (*QuizRepo)(nil)

// Creer insère un nouveau quiz
func (r *QuizRepo) Creer(ctx context.Context, quiz *Quiz) error {
	if quiz.ID == "" {
		quiz.ID = uuid.New().String()
	}
	quiz.DateCreation = time.Now()

	questionsJSON, err := json.Marshal(quiz.Questions)
	if err != nil {
		return fmt.Errorf("erreur sérialisation questions: %w", err)
	}

	query := `
		INSERT INTO quiz (id, cours_id, titre, difficulte, nombre_questions, questions, date_creation)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.ExecContext(ctx, query,
		quiz.ID,
		quiz.CoursID,
		quiz.Titre,
		quiz.Difficulte,
		quiz.NombreQuestions,
		questionsJSON,
		quiz.DateCreation,
	)
	if err != nil {
		return fmt.Errorf("erreur insertion quiz: %w", err)
	}

	return nil
}

// ObtenirParID récupère un quiz par son ID
func (r *QuizRepo) ObtenirParID(ctx context.Context, id string) (*Quiz, error) {
	query := `
		SELECT id, cours_id, titre, difficulte, nombre_questions, questions, date_creation
		FROM quiz
		WHERE id = $1
	`

	quiz := &Quiz{}
	var questionsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&quiz.ID,
		&quiz.CoursID,
		&quiz.Titre,
		&quiz.Difficulte,
		&quiz.NombreQuestions,
		&questionsJSON,
		&quiz.DateCreation,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur récupération quiz: %w", err)
	}

	if err := json.Unmarshal(questionsJSON, &quiz.Questions); err != nil {
		return nil, fmt.Errorf("erreur désérialisation questions: %w", err)
	}

	return quiz, nil
}

// ListerParCours récupère tous les quiz d'un cours
func (r *QuizRepo) ListerParCours(ctx context.Context, coursID string) ([]*Quiz, error) {
	query := `
		SELECT id, cours_id, titre, difficulte, nombre_questions, questions, date_creation
		FROM quiz
		WHERE cours_id = $1
		ORDER BY date_creation DESC
	`

	rows, err := r.db.QueryContext(ctx, query, coursID)
	if err != nil {
		return nil, fmt.Errorf("erreur requête quiz: %w", err)
	}
	defer rows.Close()

	var quizzes []*Quiz
	for rows.Next() {
		quiz := &Quiz{}
		var questionsJSON []byte

		err := rows.Scan(
			&quiz.ID,
			&quiz.CoursID,
			&quiz.Titre,
			&quiz.Difficulte,
			&quiz.NombreQuestions,
			&questionsJSON,
			&quiz.DateCreation,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan quiz: %w", err)
		}

		if err := json.Unmarshal(questionsJSON, &quiz.Questions); err != nil {
			return nil, fmt.Errorf("erreur désérialisation questions: %w", err)
		}

		quizzes = append(quizzes, quiz)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur itération quiz: %w", err)
	}

	return quizzes, nil
}

// Compter retourne le nombre total de quiz
func (r *QuizRepo) Compter(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM quiz`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage quiz: %w", err)
	}

	return count, nil
}

// Supprimer supprime un quiz par son ID
func (r *QuizRepo) Supprimer(ctx context.Context, id string) error {
	query := `DELETE FROM quiz WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression quiz: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("quiz non trouvé: %s", id)
	}

	return nil
}

// --- Sessions ---

// CreerSession crée une nouvelle session de quiz
func (r *QuizRepo) CreerSession(ctx context.Context, session *QuizSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.DateDebut = time.Now()
	session.Reponses = []ReponseSession{}
	session.Termine = false

	reponsesJSON, err := json.Marshal(session.Reponses)
	if err != nil {
		return fmt.Errorf("erreur sérialisation réponses: %w", err)
	}

	query := `
		INSERT INTO quiz_sessions (id, quiz_id, reponses, termine, date_debut)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = r.db.ExecContext(ctx, query,
		session.ID,
		session.QuizID,
		reponsesJSON,
		session.Termine,
		session.DateDebut,
	)
	if err != nil {
		return fmt.Errorf("erreur insertion session: %w", err)
	}

	return nil
}

// ObtenirSession récupère une session par son ID
func (r *QuizRepo) ObtenirSession(ctx context.Context, id string) (*QuizSession, error) {
	query := `
		SELECT id, quiz_id, reponses, score, termine, date_debut, date_fin
		FROM quiz_sessions
		WHERE id = $1
	`

	session := &QuizSession{}
	var reponsesJSON []byte
	var score sql.NullFloat64
	var dateFin sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.QuizID,
		&reponsesJSON,
		&score,
		&session.Termine,
		&session.DateDebut,
		&dateFin,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erreur récupération session: %w", err)
	}

	if err := json.Unmarshal(reponsesJSON, &session.Reponses); err != nil {
		return nil, fmt.Errorf("erreur désérialisation réponses: %w", err)
	}

	if score.Valid {
		session.Score = &score.Float64
	}
	if dateFin.Valid {
		session.DateFin = &dateFin.Time
	}

	return session, nil
}

// MettreAJourSession met à jour une session existante
func (r *QuizRepo) MettreAJourSession(ctx context.Context, session *QuizSession) error {
	reponsesJSON, err := json.Marshal(session.Reponses)
	if err != nil {
		return fmt.Errorf("erreur sérialisation réponses: %w", err)
	}

	query := `
		UPDATE quiz_sessions
		SET reponses = $1, score = $2, termine = $3, date_fin = $4
		WHERE id = $5
	`

	var scorePtr *float64
	if session.Score != nil {
		scorePtr = session.Score
	}

	_, err = r.db.ExecContext(ctx, query,
		reponsesJSON,
		scorePtr,
		session.Termine,
		session.DateFin,
		session.ID,
	)
	if err != nil {
		return fmt.Errorf("erreur mise à jour session: %w", err)
	}

	return nil
}

// --- Statistiques ---

// CompterQuizCompletes retourne le nombre de sessions terminées
func (r *QuizRepo) CompterQuizCompletes(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM quiz_sessions WHERE termine = true`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage quiz complétés: %w", err)
	}

	return count, nil
}

// ScoreMoyen retourne le score moyen de toutes les sessions terminées
func (r *QuizRepo) ScoreMoyen(ctx context.Context) (float64, error) {
	query := `SELECT COALESCE(AVG(score), 0) FROM quiz_sessions WHERE termine = true AND score IS NOT NULL`

	var avg float64
	err := r.db.QueryRowContext(ctx, query).Scan(&avg)
	if err != nil {
		return 0, fmt.Errorf("erreur calcul score moyen: %w", err)
	}

	return avg, nil
}
