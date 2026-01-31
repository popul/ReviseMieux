// Package services contient la logique métier
package services

import (
	"context"
	"errors"
	"time"

	"github.com/revisemieux/backend/internal/store"
)

// Erreurs de quota
var (
	ErrQuotaOCRDepasse        = errors.New("quota OCR journalier dépassé")
	ErrQuotaGenerationDepasse = errors.New("quota génération journalier dépassé")
)

// TypeOperation représente le type d'opération soumise aux quotas
type TypeOperation string

const (
	TypeOperationOCR        TypeOperation = "ocr"
	TypeOperationGeneration TypeOperation = "generation"
)

// StatutQuota représente l'état actuel des quotas
type StatutQuota struct {
	PagesOCRUtilisees     int `json:"pagesOcrUtilisees"`
	PagesOCRMax           int `json:"pagesOcrMax"`
	GenerationsUtilisees  int `json:"generationsUtilisees"`
	GenerationsMax        int `json:"generationsMax"`
	PagesOCRRestantes     int `json:"pagesOcrRestantes"`
	GenerationsRestantes  int `json:"generationsRestantes"`
}

// ServiceQuotas gère les quotas journaliers
type ServiceQuotas struct {
	quotasRepo      store.QuotasRepository
	limiteOCR       int
	limiteGeneration int
}

// NouveauServiceQuotas crée un nouveau service de quotas
func NouveauServiceQuotas(quotasRepo store.QuotasRepository, limiteOCR, limiteGeneration int) *ServiceQuotas {
	return &ServiceQuotas{
		quotasRepo:       quotasRepo,
		limiteOCR:        limiteOCR,
		limiteGeneration: limiteGeneration,
	}
}

// VerifierQuotaOCR vérifie si le quota OCR permet d'ajouter des pages
func (s *ServiceQuotas) VerifierQuotaOCR(ctx context.Context, nombrePages int) error {
	quotas, err := s.quotasRepo.ObtenirOuCreer(ctx, time.Now())
	if err != nil {
		return err
	}

	if quotas.PagesOCR+nombrePages > s.limiteOCR {
		return ErrQuotaOCRDepasse
	}

	return nil
}

// VerifierQuotaGeneration vérifie si le quota de génération permet une nouvelle génération
func (s *ServiceQuotas) VerifierQuotaGeneration(ctx context.Context) error {
	quotas, err := s.quotasRepo.ObtenirOuCreer(ctx, time.Now())
	if err != nil {
		return err
	}

	if quotas.Generations >= s.limiteGeneration {
		return ErrQuotaGenerationDepasse
	}

	return nil
}

// IncrementerOCR incrémente le compteur OCR après succès
func (s *ServiceQuotas) IncrementerOCR(ctx context.Context, nombrePages int) error {
	return s.quotasRepo.IncrementerOCR(ctx, time.Now(), nombrePages)
}

// IncrementerGeneration incrémente le compteur de génération après succès
func (s *ServiceQuotas) IncrementerGeneration(ctx context.Context) error {
	return s.quotasRepo.IncrementerGeneration(ctx, time.Now())
}

// ObtenirStatut retourne l'état actuel des quotas
func (s *ServiceQuotas) ObtenirStatut(ctx context.Context) (*StatutQuota, error) {
	quotas, err := s.quotasRepo.ObtenirOuCreer(ctx, time.Now())
	if err != nil {
		return nil, err
	}

	pagesRestantes := s.limiteOCR - quotas.PagesOCR
	if pagesRestantes < 0 {
		pagesRestantes = 0
	}

	genRestantes := s.limiteGeneration - quotas.Generations
	if genRestantes < 0 {
		genRestantes = 0
	}

	return &StatutQuota{
		PagesOCRUtilisees:     quotas.PagesOCR,
		PagesOCRMax:           s.limiteOCR,
		GenerationsUtilisees:  quotas.Generations,
		GenerationsMax:        s.limiteGeneration,
		PagesOCRRestantes:     pagesRestantes,
		GenerationsRestantes:  genRestantes,
	}, nil
}
