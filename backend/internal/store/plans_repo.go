// Package store contient les accès à la base de données
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PlanRevision représente un plan de révision
type PlanRevision struct {
	ID               string     `json:"id"`
	Titre            string     `json:"titre"`
	Description      string     `json:"description,omitempty"`
	Matiere          string     `json:"matiere,omitempty"`
	IconeMatiere     string     `json:"iconeMatiere"`
	DateEcheance     *time.Time `json:"dateEcheance,omitempty"`
	DateCreation     time.Time  `json:"dateCreation"`
	DateModification time.Time  `json:"dateModification"`
}

// PlanRevisionResume représente un résumé compact d'un plan pour la sidebar
type PlanRevisionResume struct {
	ID           string     `json:"id"`
	Titre        string     `json:"titre"`
	IconeMatiere string     `json:"iconeMatiere"`
	NombreCours  int        `json:"nombreCours"`
	Progression  int        `json:"progression"`
	DateEcheance *time.Time `json:"dateEcheance,omitempty"`
}

// PlansRevisionRepository définit les opérations pour les plans de révision
type PlansRevisionRepository interface {
	Creer(ctx context.Context, plan *PlanRevision) error
	ObtenirParID(ctx context.Context, id string) (*PlanRevision, error)
	Lister(ctx context.Context, limite, offset int) ([]*PlanRevision, error)
	MettreAJour(ctx context.Context, plan *PlanRevision) error
	Supprimer(ctx context.Context, id string) error
	AjouterCours(ctx context.Context, planID, coursID string, ordre int) error
	RetirerCours(ctx context.Context, planID, coursID string) error
	ListerCoursIDs(ctx context.Context, planID string) ([]string, error)
	ListerResumes(ctx context.Context) ([]*PlanRevisionResume, error)
	ListerPlansParCours(ctx context.Context, coursID string) ([]*PlanRevisionResume, error)
}

// PlansRevisionRepo implémente PlansRevisionRepository avec PostgreSQL
type PlansRevisionRepo struct {
	db *sql.DB
}

// NouveauPlansRevisionRepo crée un nouveau repository pour les plans de révision
func NouveauPlansRevisionRepo(store *Store) *PlansRevisionRepo {
	return &PlansRevisionRepo{db: store.DB()}
}

// Vérification que PlansRevisionRepo implémente PlansRevisionRepository
var _ PlansRevisionRepository = (*PlansRevisionRepo)(nil)

// Creer insère un nouveau plan de révision
func (r *PlansRevisionRepo) Creer(ctx context.Context, plan *PlanRevision) error {
	if plan.ID == "" {
		plan.ID = uuid.New().String()
	}
	if plan.IconeMatiere == "" {
		plan.IconeMatiere = "📋"
	}

	now := time.Now()
	plan.DateCreation = now
	plan.DateModification = now

	query := `
		INSERT INTO plans_revision (id, titre, description, matiere, icone_matiere, date_echeance, date_creation, date_modification)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		plan.ID,
		plan.Titre,
		plan.Description,
		plan.Matiere,
		plan.IconeMatiere,
		plan.DateEcheance,
		plan.DateCreation,
		plan.DateModification,
	)
	if err != nil {
		return fmt.Errorf("erreur création plan de révision: %w", err)
	}

	return nil
}

// ObtenirParID récupère un plan de révision par son identifiant
func (r *PlansRevisionRepo) ObtenirParID(ctx context.Context, id string) (*PlanRevision, error) {
	query := `
		SELECT id, titre, COALESCE(description, ''), COALESCE(matiere, ''), COALESCE(icone_matiere, '📋'), date_echeance, date_creation, date_modification
		FROM plans_revision
		WHERE id = $1
	`

	plan := &PlanRevision{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&plan.ID,
		&plan.Titre,
		&plan.Description,
		&plan.Matiere,
		&plan.IconeMatiere,
		&plan.DateEcheance,
		&plan.DateCreation,
		&plan.DateModification,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("plan de révision non trouvé: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("erreur récupération plan de révision: %w", err)
	}

	return plan, nil
}

// Lister récupère les plans de révision avec pagination
func (r *PlansRevisionRepo) Lister(ctx context.Context, limite, offset int) ([]*PlanRevision, error) {
	query := `
		SELECT id, titre, COALESCE(description, ''), COALESCE(matiere, ''), COALESCE(icone_matiere, '📋'), date_echeance, date_creation, date_modification
		FROM plans_revision
		ORDER BY date_creation DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limite, offset)
	if err != nil {
		return nil, fmt.Errorf("erreur listage plans de révision: %w", err)
	}
	defer rows.Close()

	var plans []*PlanRevision
	for rows.Next() {
		plan := &PlanRevision{}
		err := rows.Scan(
			&plan.ID,
			&plan.Titre,
			&plan.Description,
			&plan.Matiere,
			&plan.IconeMatiere,
			&plan.DateEcheance,
			&plan.DateCreation,
			&plan.DateModification,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan plan de révision: %w", err)
		}
		plans = append(plans, plan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur itération plans de révision: %w", err)
	}

	return plans, nil
}

// MettreAJour met à jour un plan de révision existant
func (r *PlansRevisionRepo) MettreAJour(ctx context.Context, plan *PlanRevision) error {
	plan.DateModification = time.Now()

	query := `
		UPDATE plans_revision
		SET titre = $2, description = $3, matiere = $4, icone_matiere = $5, date_echeance = $6, date_modification = $7
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		plan.ID,
		plan.Titre,
		plan.Description,
		plan.Matiere,
		plan.IconeMatiere,
		plan.DateEcheance,
		plan.DateModification,
	)
	if err != nil {
		return fmt.Errorf("erreur mise à jour plan de révision: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification mise à jour: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("plan de révision non trouvé: %s", plan.ID)
	}

	return nil
}

// Supprimer supprime un plan de révision par son identifiant
func (r *PlansRevisionRepo) Supprimer(ctx context.Context, id string) error {
	query := `DELETE FROM plans_revision WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression plan de révision: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("plan de révision non trouvé: %s", id)
	}

	return nil
}

// AjouterCours ajoute un cours à un plan de révision
func (r *PlansRevisionRepo) AjouterCours(ctx context.Context, planID, coursID string, ordre int) error {
	query := `
		INSERT INTO plans_revision_cours (plan_id, cours_id, ordre)
		VALUES ($1, $2, $3)
		ON CONFLICT (plan_id, cours_id) DO UPDATE SET ordre = $3
	`

	_, err := r.db.ExecContext(ctx, query, planID, coursID, ordre)
	if err != nil {
		return fmt.Errorf("erreur ajout cours au plan: %w", err)
	}

	return nil
}

// RetirerCours retire un cours d'un plan de révision
func (r *PlansRevisionRepo) RetirerCours(ctx context.Context, planID, coursID string) error {
	query := `DELETE FROM plans_revision_cours WHERE plan_id = $1 AND cours_id = $2`

	result, err := r.db.ExecContext(ctx, query, planID, coursID)
	if err != nil {
		return fmt.Errorf("erreur retrait cours du plan: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification retrait: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cours non trouvé dans le plan")
	}

	return nil
}

// ListerCoursIDs récupère les IDs des cours d'un plan
func (r *PlansRevisionRepo) ListerCoursIDs(ctx context.Context, planID string) ([]string, error) {
	query := `
		SELECT cours_id FROM plans_revision_cours
		WHERE plan_id = $1
		ORDER BY ordre ASC, date_ajout ASC
	`

	rows, err := r.db.QueryContext(ctx, query, planID)
	if err != nil {
		return nil, fmt.Errorf("erreur listage cours du plan: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("erreur scan cours ID: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// ListerResumes récupère les résumés de tous les plans pour la sidebar
func (r *PlansRevisionRepo) ListerResumes(ctx context.Context) ([]*PlanRevisionResume, error) {
	query := `
		SELECT
			p.id,
			p.titre,
			COALESCE(p.icone_matiere, '📋'),
			COUNT(prc.cours_id)::int AS nombre_cours,
			COALESCE(
				CASE WHEN COUNT(prc.cours_id) = 0 THEN 0
				ELSE (COUNT(DISTINCT CASE WHEN f.cours_id IS NOT NULL THEN prc.cours_id END) * 100 / COUNT(DISTINCT prc.cours_id))
				END,
			0)::int AS progression,
			p.date_echeance
		FROM plans_revision p
		LEFT JOIN plans_revision_cours prc ON prc.plan_id = p.id
		LEFT JOIN fiches f ON f.cours_id = prc.cours_id
		GROUP BY p.id, p.titre, p.icone_matiere, p.date_echeance, p.date_creation
		ORDER BY p.date_creation DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("erreur listage résumés plans: %w", err)
	}
	defer rows.Close()

	var resumes []*PlanRevisionResume
	for rows.Next() {
		resume := &PlanRevisionResume{}
		err := rows.Scan(
			&resume.ID,
			&resume.Titre,
			&resume.IconeMatiere,
			&resume.NombreCours,
			&resume.Progression,
			&resume.DateEcheance,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan résumé plan: %w", err)
		}
		resumes = append(resumes, resume)
	}

	return resumes, rows.Err()
}

// ListerPlansParCours récupère les plans contenant un cours donné
func (r *PlansRevisionRepo) ListerPlansParCours(ctx context.Context, coursID string) ([]*PlanRevisionResume, error) {
	query := `
		SELECT
			p.id,
			p.titre,
			COALESCE(p.icone_matiere, '📋'),
			COUNT(prc2.cours_id)::int AS nombre_cours,
			0 AS progression,
			p.date_echeance
		FROM plans_revision p
		INNER JOIN plans_revision_cours prc ON prc.plan_id = p.id AND prc.cours_id = $1
		LEFT JOIN plans_revision_cours prc2 ON prc2.plan_id = p.id
		GROUP BY p.id, p.titre, p.icone_matiere, p.date_echeance, p.date_creation
		ORDER BY p.date_creation DESC
	`

	rows, err := r.db.QueryContext(ctx, query, coursID)
	if err != nil {
		return nil, fmt.Errorf("erreur listage plans par cours: %w", err)
	}
	defer rows.Close()

	var resumes []*PlanRevisionResume
	for rows.Next() {
		resume := &PlanRevisionResume{}
		err := rows.Scan(
			&resume.ID,
			&resume.Titre,
			&resume.IconeMatiere,
			&resume.NombreCours,
			&resume.Progression,
			&resume.DateEcheance,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan résumé plan: %w", err)
		}
		resumes = append(resumes, resume)
	}

	return resumes, rows.Err()
}
