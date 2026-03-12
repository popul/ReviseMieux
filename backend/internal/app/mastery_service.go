package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/domain/mastery"
)

// MasteryService handles mastery-related use cases.
type MasteryService struct {
	masteryRepo mastery.Repository
	publisher   event.Publisher
	clock       event.Clock
}

// NewMasteryService creates a new MasteryService.
func NewMasteryService(
	masteryRepo mastery.Repository,
	publisher event.Publisher,
	clock event.Clock,
) *MasteryService {
	return &MasteryService{
		masteryRepo: masteryRepo,
		publisher:   publisher,
		clock:       clock,
	}
}

// GetByUser returns all masteries for a user in a given state.
func (s *MasteryService) GetByUser(ctx context.Context, userID uuid.UUID, state mastery.State) ([]*mastery.Mastery, error) {
	return s.masteryRepo.FindByUserAndState(ctx, userID, state)
}

// GetByUserAndItem returns the mastery for a specific user+item pair.
func (s *MasteryService) GetByUserAndItem(ctx context.Context, userID, itemID uuid.UUID) (*mastery.Mastery, error) {
	return s.masteryRepo.FindByUserAndItem(ctx, userID, itemID)
}

// RecordAttempt records a score for a user+item and transitions the mastery state.
func (s *MasteryService) RecordAttempt(ctx context.Context, userID, itemID uuid.UUID, score float64) (*mastery.Mastery, error) {
	now := s.clock.Now()

	m, err := s.masteryRepo.FindByUserAndItem(ctx, userID, itemID)
	if err != nil {
		return nil, fmt.Errorf("mastery_service: find mastery: %w", err)
	}

	if err := m.RecordAttempt(score, now); err != nil {
		return nil, fmt.Errorf("mastery_service: record attempt: %w", err)
	}

	if err := s.masteryRepo.Save(ctx, m); err != nil {
		return nil, fmt.Errorf("mastery_service: save mastery: %w", err)
	}

	s.publisher.Publish(ctx, event.AttemptRecorded{
		BaseEvent: event.BaseEvent{OccurredOn: now},
		UserID:    userID,
		ItemID:    itemID,
		Score:     score,
	})

	return m, nil
}
