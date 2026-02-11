// Package services contient les services métier de l'application
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/revisemieux/backend/internal/store"
)

// ServicePlans gère les plans de révision
type ServicePlans struct {
	plansRepo   store.PlansRevisionRepository
	coursRepo   store.CoursRepository
	fichesRepo  store.FichesRepository
	quizRepo    store.QuizRepository
	mindmapRepo store.MindmapRepository
}

// NouveauServicePlans crée une nouvelle instance du service de plans
func NouveauServicePlans(
	plansRepo store.PlansRevisionRepository,
	coursRepo store.CoursRepository,
	fichesRepo store.FichesRepository,
	quizRepo store.QuizRepository,
	mindmapRepo store.MindmapRepository,
) *ServicePlans {
	return &ServicePlans{
		plansRepo:   plansRepo,
		coursRepo:   coursRepo,
		fichesRepo:  fichesRepo,
		quizRepo:    quizRepo,
		mindmapRepo: mindmapRepo,
	}
}

// CoursAvecArtifacts représente un cours avec le compteur de ses artefacts
type CoursAvecArtifacts struct {
	Cours        *store.Cours `json:"cours"`
	NombreFiches int          `json:"nombreFiches"`
	NombreQuiz   int          `json:"nombreQuiz"`
	AMindmap     bool         `json:"aMindmap"`
	AResume      bool         `json:"aResume"`
}

// PlanComplet représente un plan avec ses cours et compteurs d'artefacts
type PlanComplet struct {
	Plan  *store.PlanRevision `json:"plan"`
	Cours []CoursAvecArtifacts `json:"cours"`
}

// CreerPlan crée un nouveau plan de révision et y ajoute les cours
func (s *ServicePlans) CreerPlan(ctx context.Context, titre, description, matiere, icone string, dateEcheance *string, coursIDs []string) (*store.PlanRevision, error) {
	plan := &store.PlanRevision{
		Titre:        titre,
		Description:  description,
		Matiere:      matiere,
		IconeMatiere: icone,
	}

	if dateEcheance != nil && *dateEcheance != "" {
		t, err := parseDate(*dateEcheance)
		if err != nil {
			return nil, fmt.Errorf("date d'échéance invalide: %w", err)
		}
		plan.DateEcheance = &t
	}

	if err := s.plansRepo.Creer(ctx, plan); err != nil {
		return nil, fmt.Errorf("erreur création plan: %w", err)
	}

	// Ajouter les cours au plan
	for i, coursID := range coursIDs {
		if err := s.plansRepo.AjouterCours(ctx, plan.ID, coursID, i); err != nil {
			return nil, fmt.Errorf("erreur ajout cours %s au plan: %w", coursID, err)
		}
	}

	return plan, nil
}

// ObtenirPlan récupère un plan par son ID
func (s *ServicePlans) ObtenirPlan(ctx context.Context, id string) (*store.PlanRevision, error) {
	return s.plansRepo.ObtenirParID(ctx, id)
}

// ListerPlans liste les plans avec pagination
func (s *ServicePlans) ListerPlans(ctx context.Context, limite, offset int) ([]*store.PlanRevision, error) {
	return s.plansRepo.Lister(ctx, limite, offset)
}

// MettreAJourPlan met à jour un plan existant
func (s *ServicePlans) MettreAJourPlan(ctx context.Context, id, titre, description, matiere, icone string, dateEcheance *string) (*store.PlanRevision, error) {
	plan, err := s.plansRepo.ObtenirParID(ctx, id)
	if err != nil {
		return nil, err
	}

	if titre != "" {
		plan.Titre = titre
	}
	if description != "" {
		plan.Description = description
	}
	if matiere != "" {
		plan.Matiere = matiere
	}
	if icone != "" {
		plan.IconeMatiere = icone
	}
	if dateEcheance != nil {
		if *dateEcheance == "" {
			plan.DateEcheance = nil
		} else {
			t, err := parseDate(*dateEcheance)
			if err != nil {
				return nil, fmt.Errorf("date d'échéance invalide: %w", err)
			}
			plan.DateEcheance = &t
		}
	}

	if err := s.plansRepo.MettreAJour(ctx, plan); err != nil {
		return nil, err
	}

	return plan, nil
}

// SupprimerPlan supprime un plan
func (s *ServicePlans) SupprimerPlan(ctx context.Context, id string) error {
	return s.plansRepo.Supprimer(ctx, id)
}

// AjouterCours ajoute un cours à un plan
func (s *ServicePlans) AjouterCours(ctx context.Context, planID, coursID string) error {
	// Vérifier que le plan existe
	if _, err := s.plansRepo.ObtenirParID(ctx, planID); err != nil {
		return err
	}

	// Récupérer le nombre de cours existants pour l'ordre
	ids, err := s.plansRepo.ListerCoursIDs(ctx, planID)
	if err != nil {
		return err
	}

	return s.plansRepo.AjouterCours(ctx, planID, coursID, len(ids))
}

// RetirerCours retire un cours d'un plan
func (s *ServicePlans) RetirerCours(ctx context.Context, planID, coursID string) error {
	return s.plansRepo.RetirerCours(ctx, planID, coursID)
}

// ListerResumes liste les résumés des plans pour la sidebar
func (s *ServicePlans) ListerResumes(ctx context.Context) ([]*store.PlanRevisionResume, error) {
	return s.plansRepo.ListerResumes(ctx)
}

// ListerPlansParCours récupère les plans contenant un cours donné
func (s *ServicePlans) ListerPlansParCours(ctx context.Context, coursID string) ([]*store.PlanRevisionResume, error) {
	return s.plansRepo.ListerPlansParCours(ctx, coursID)
}

// ObtenirPlanComplet récupère un plan avec ses cours et compteurs d'artefacts
func (s *ServicePlans) ObtenirPlanComplet(ctx context.Context, planID string) (*PlanComplet, error) {
	plan, err := s.plansRepo.ObtenirParID(ctx, planID)
	if err != nil {
		return nil, err
	}

	coursIDs, err := s.plansRepo.ListerCoursIDs(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("erreur listage cours du plan: %w", err)
	}

	coursAvecArtifacts := make([]CoursAvecArtifacts, 0, len(coursIDs))
	for _, coursID := range coursIDs {
		cours, err := s.coursRepo.ObtenirParID(ctx, coursID)
		if err != nil {
			continue // cours peut avoir été supprimé
		}

		ca := CoursAvecArtifacts{
			Cours:   cours,
			AResume: len(cours.Resume) > 0 && string(cours.Resume) != "null",
		}

		// Compter les fiches
		if s.fichesRepo != nil {
			fiches, err := s.fichesRepo.ListerParCours(ctx, coursID)
			if err == nil {
				ca.NombreFiches = len(fiches)
			}
		}

		// Compter les quiz
		if s.quizRepo != nil {
			quiz, err := s.quizRepo.ListerParCours(ctx, coursID)
			if err == nil {
				ca.NombreQuiz = len(quiz)
			}
		}

		// Vérifier la mindmap
		if s.mindmapRepo != nil {
			mindmap, err := s.mindmapRepo.ObtenirParCours(ctx, coursID)
			if err == nil && mindmap != nil {
				ca.AMindmap = true
			}
		}

		coursAvecArtifacts = append(coursAvecArtifacts, ca)
	}

	return &PlanComplet{
		Plan:  plan,
		Cours: coursAvecArtifacts,
	}, nil
}

// parseDate parse une date ISO 8601
func parseDate(s string) (time.Time, error) {
	// Essayer plusieurs formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("format de date non reconnu: %s", s)
}
