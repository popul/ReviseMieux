// Package services contient la logique métier
package services

import (
	"context"
	"time"

	"github.com/revisemieux/backend/internal/store"
)

// Statistiques représente les statistiques globales de l'application
type Statistiques struct {
	NombreCours      int      `json:"nombreCours"`
	NombreFiches     int      `json:"nombreFiches"`
	NombreQuiz       int      `json:"nombreQuiz"`
	QuizCompletes    int      `json:"quizCompletes"`
	ScoreMoyen       *float64 `json:"scoreMoyen,omitempty"`
	DateMiseAJour    time.Time `json:"dateMiseAJour"`
}

// CoursResume représente un résumé d'un cours pour le dashboard
type CoursResume struct {
	ID               string    `json:"id"`
	Titre            string    `json:"titre"`
	Matiere          string    `json:"matiere,omitempty"`
	NombreFiches     int       `json:"nombreFiches"`
	NombreQuiz       int       `json:"nombreQuiz"`
	DateCreation     time.Time `json:"dateCreation"`
	DateModification time.Time `json:"dateModification"`
}

// ServiceStatistiques fournit les statistiques de l'application
type ServiceStatistiques struct {
	coursRepo   store.CoursRepository
	fichesRepo  store.FichesRepository
	quizRepo    store.QuizRepository
}

// NouveauServiceStatistiques crée un nouveau service de statistiques
func NouveauServiceStatistiques(
	coursRepo store.CoursRepository,
	fichesRepo store.FichesRepository,
	quizRepo store.QuizRepository,
) *ServiceStatistiques {
	return &ServiceStatistiques{
		coursRepo:  coursRepo,
		fichesRepo: fichesRepo,
		quizRepo:   quizRepo,
	}
}

// ObtenirStatistiques récupère les statistiques globales
func (s *ServiceStatistiques) ObtenirStatistiques(ctx context.Context) (*Statistiques, error) {
	stats := &Statistiques{
		DateMiseAJour: time.Now(),
	}

	// Compter les cours
	if s.coursRepo != nil {
		count, err := s.coursRepo.Compter(ctx)
		if err == nil {
			stats.NombreCours = count
		}
	}

	// Compter les fiches
	if s.fichesRepo != nil {
		count, err := s.fichesRepo.Compter(ctx)
		if err == nil {
			stats.NombreFiches = count
		}
	}

	// Compter les quiz
	if s.quizRepo != nil {
		count, err := s.quizRepo.Compter(ctx)
		if err == nil {
			stats.NombreQuiz = count
		}

		// Quiz complétés
		completed, err := s.quizRepo.CompterQuizCompletes(ctx)
		if err == nil {
			stats.QuizCompletes = completed
		}

		// Score moyen (seulement s'il y a des quiz complétés)
		if completed > 0 {
			avg, err := s.quizRepo.ScoreMoyen(ctx)
			if err == nil && avg > 0 {
				stats.ScoreMoyen = &avg
			}
		}
	}

	return stats, nil
}

// ListerCoursRecents récupère les cours récents avec leurs stats
func (s *ServiceStatistiques) ListerCoursRecents(ctx context.Context, limite int) ([]*CoursResume, error) {
	if s.coursRepo == nil {
		return []*CoursResume{}, nil
	}

	// Récupérer les cours récents
	cours, err := s.coursRepo.Lister(ctx, limite, 0)
	if err != nil {
		return nil, err
	}

	resumes := make([]*CoursResume, 0, len(cours))
	for _, c := range cours {
		resume := &CoursResume{
			ID:               c.ID,
			Titre:            c.Titre,
			Matiere:          c.Matiere,
			DateCreation:     c.DateCreation,
			DateModification: c.DateModification,
		}

		// Compter les fiches du cours
		if s.fichesRepo != nil {
			count, err := s.fichesRepo.CompterParCours(ctx, c.ID)
			if err == nil {
				resume.NombreFiches = count
			}
		}

		// Compter les quiz du cours (via ListerParCours)
		if s.quizRepo != nil {
			quizzes, err := s.quizRepo.ListerParCours(ctx, c.ID)
			if err == nil {
				resume.NombreQuiz = len(quizzes)
			}
		}

		resumes = append(resumes, resume)
	}

	return resumes, nil
}
