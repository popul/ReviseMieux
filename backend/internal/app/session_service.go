package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/domain/mastery"
	"github.com/popul/revisemieux/internal/domain/session"
)

// Difficulty-1 templates eligible for evening_first sessions.
var difficulty1Templates = []string{
	"GEN.KNOW.FLASH_MCQ",
	"GEN.KNOW.DEF_SHORT",
	"GEN.KNOW.CLOZE_KEYWORDS",
}

// SessionService handles session-related use cases.
type SessionService struct {
	sessionRepo session.Repository
	chapterRepo chapter.Repository
	masteryRepo mastery.Repository
	publisher   event.Publisher
	clock       event.Clock
	idGen       event.IDGenerator
}

// NewSessionService creates a new SessionService.
func NewSessionService(
	sessionRepo session.Repository,
	chapterRepo chapter.Repository,
	masteryRepo mastery.Repository,
	publisher event.Publisher,
	clock event.Clock,
	idGen event.IDGenerator,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		chapterRepo: chapterRepo,
		masteryRepo: masteryRepo,
		publisher:   publisher,
		clock:       clock,
		idGen:       idGen,
	}
}

// ProposeEveningFirst creates an evening_first session after pipeline completion.
// Z6-AC04: triggered immediately after pipeline produces items.
// Z6-AC05: 100% UNKNOWN items, difficulty 1 templates, 6-10 questions.
func (s *SessionService) ProposeEveningFirst(ctx context.Context, userID, chapterID uuid.UUID, itemIDs []uuid.UUID) (*session.Session, error) {
	now := s.clock.Now()

	if len(itemIDs) == 0 {
		return nil, session.ErrEmptyPool
	}

	sess := session.NewSession(s.idGen, userID, session.TypeEveningFirst, session.TriggerScheduled, now)
	sess.ChapterIDs = []uuid.UUID{chapterID}

	if err := s.sessionRepo.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("session_service: save session: %w", err)
	}

	// Select items for the session: max 10 questions, min 6
	maxQuestions := 10
	if len(itemIDs) < maxQuestions {
		maxQuestions = len(itemIDs)
	}
	if maxQuestions < 6 {
		maxQuestions = len(itemIDs) // use all if fewer than 6
	}

	selectedItems := itemIDs[:maxQuestions]

	// Generate questions using difficulty 1 templates (round-robin)
	for i, itemID := range selectedItems {
		templateID := difficulty1Templates[i%len(difficulty1Templates)]
		q := &session.Question{
			ID:         s.idGen.New(),
			SessionID:  sess.ID,
			TemplateID: templateID,
			ItemID:     itemID,
			CreatedAt:  now,
		}
		if err := s.sessionRepo.SaveQuestion(ctx, q); err != nil {
			return nil, fmt.Errorf("session_service: save question: %w", err)
		}
	}

	return sess, nil
}

// ComposeDaily creates a daily session for a chapter.
// Z4-AC06: returns error if no items available (no empty session created).
func (s *SessionService) ComposeDaily(ctx context.Context, userID, chapterID uuid.UUID) (*session.Session, error) {
	now := s.clock.Now()

	// Get due items for this user
	dueMasteries, err := s.masteryRepo.FindDueByUser(ctx, userID, now)
	if err != nil {
		return nil, fmt.Errorf("session_service: find due: %w", err)
	}

	// Filter to items from this chapter
	chapterItems, err := s.chapterRepo.FindItemsByChapter(ctx, chapterID, false)
	if err != nil {
		return nil, fmt.Errorf("session_service: find items: %w", err)
	}

	chapterItemSet := make(map[uuid.UUID]bool)
	for _, item := range chapterItems {
		chapterItemSet[item.ID] = true
	}

	var eligibleMasteries []*mastery.Mastery
	for _, m := range dueMasteries {
		if chapterItemSet[m.ItemID] {
			eligibleMasteries = append(eligibleMasteries, m)
		}
	}

	if len(eligibleMasteries) == 0 {
		return nil, session.ErrEmptyPool
	}

	sess := session.NewSession(s.idGen, userID, session.TypeDaily, session.TriggerManual, now)
	sess.ChapterIDs = []uuid.UUID{chapterID}

	if err := s.sessionRepo.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("session_service: save session: %w", err)
	}

	// Select up to 10 items
	maxQuestions := 10
	if len(eligibleMasteries) < maxQuestions {
		maxQuestions = len(eligibleMasteries)
	}

	for i := 0; i < maxQuestions; i++ {
		templateID := difficulty1Templates[i%len(difficulty1Templates)]
		q := &session.Question{
			ID:         s.idGen.New(),
			SessionID:  sess.ID,
			TemplateID: templateID,
			ItemID:     eligibleMasteries[i].ItemID,
			CreatedAt:  now,
		}
		if err := s.sessionRepo.SaveQuestion(ctx, q); err != nil {
			return nil, fmt.Errorf("session_service: save question: %w", err)
		}
	}

	return sess, nil
}

// ResumeSession allows resuming an in-progress session.
// Z4-AC04: resumes from last known question index.
func (s *SessionService) ResumeSession(ctx context.Context, sessionID uuid.UUID) (*session.Session, error) {
	sess, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session_service: find session: %w", err)
	}

	if err := sess.Resume(); err != nil {
		return nil, fmt.Errorf("session_service: %w", err)
	}

	return sess, nil
}

// AvailableSessionTypes returns session types available based on schedule status.
// Z6-AC11: without schedule, daily/evening_first/mock_exam still available;
// pre_class requires schedule.
func (s *SessionService) AvailableSessionTypes(hasSchedule bool) []session.SessionType {
	types := []session.SessionType{
		session.TypeDaily,
		session.TypeEveningFirst,
		session.TypeMockExam,
		session.TypeDiagnostic,
	}
	if hasSchedule {
		types = append(types, session.TypePreClass)
	}
	return types
}
