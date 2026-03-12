package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
)

// ExamService handles exam-related use cases (Z6-AC10).
type ExamService struct {
	chapterRepo chapter.Repository
	publisher   event.Publisher
	clock       event.Clock
	idGen       event.IDGenerator
}

// NewExamService creates a new ExamService.
func NewExamService(
	chapterRepo chapter.Repository,
	publisher event.Publisher,
	clock event.Clock,
	idGen event.IDGenerator,
) *ExamService {
	return &ExamService{
		chapterRepo: chapterRepo,
		publisher:   publisher,
		clock:       clock,
		idGen:       idGen,
	}
}

// CreateExam creates an exam linked to chapters and triggers compression (Z6-AC10).
func (s *ExamService) CreateExam(ctx context.Context, userID uuid.UUID, title string, examDate time.Time, chapterIDs []uuid.UUID) (*chapter.Exam, error) {
	now := s.clock.Now()

	exam := chapter.NewExam(s.idGen, userID, title, examDate, chapterIDs, now)

	// Publish ExamCreated → triggers mastery compression for linked chapters
	s.publisher.Publish(ctx, event.ExamCreated{
		BaseEvent:  event.BaseEvent{OccurredOn: now},
		ExamID:     exam.ID,
		ChapterIDs: chapterIDs,
	})

	return exam, nil
}

// UpdateExam allows adding/removing chapters and changing the date.
func (s *ExamService) UpdateExam(ctx context.Context, exam *chapter.Exam, title *string, examDate *time.Time, addChapterIDs, removeChapterIDs []uuid.UUID) error {
	now := s.clock.Now()

	if title != nil {
		exam.Title = *title
	}
	if examDate != nil {
		exam.ExamDate = *examDate
	}
	for _, id := range addChapterIDs {
		exam.AddChapter(id, now)
	}
	for _, id := range removeChapterIDs {
		exam.RemoveChapter(id, now)
	}
	exam.UpdatedAt = now

	return nil
}
