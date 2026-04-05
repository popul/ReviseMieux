package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
)

// ExamService handles exam-related use cases (Z6-AC10).
type ExamService struct {
	examRepo  chapter.ExamRepository
	publisher event.Publisher
	clock     event.Clock
	idGen     event.IDGenerator
}

// NewExamService creates a new ExamService.
func NewExamService(
	examRepo chapter.ExamRepository,
	publisher event.Publisher,
	clock event.Clock,
	idGen event.IDGenerator,
) *ExamService {
	return &ExamService{
		examRepo:  examRepo,
		publisher: publisher,
		clock:     clock,
		idGen:     idGen,
	}
}

// CreateExam creates an exam linked to chapters, persists it, and triggers
// mastery compression (Z6-AC10).
func (s *ExamService) CreateExam(ctx context.Context, userID uuid.UUID, title string, examDate time.Time, chapterIDs []uuid.UUID) (*chapter.Exam, error) {
	now := s.clock.Now()

	exam := chapter.NewExam(s.idGen, userID, title, examDate, chapterIDs, now)

	if err := s.examRepo.Save(ctx, exam); err != nil {
		return nil, fmt.Errorf("exam_service: save exam: %w", err)
	}

	// Publish ExamCreated -> triggers mastery compression for linked chapters
	if err := s.publisher.Publish(ctx, event.ExamCreated{
		BaseEvent:  event.BaseEvent{OccurredOn: now},
		ExamID:     exam.ID,
		ChapterIDs: chapterIDs,
	}); err != nil {
		return nil, fmt.Errorf("exam_service: publish event: %w", err)
	}

	return exam, nil
}

// UpdateExam allows adding/removing chapters and changing the date, then persists.
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

	if err := s.examRepo.Save(ctx, exam); err != nil {
		return fmt.Errorf("exam_service: save exam: %w", err)
	}

	return nil
}

// GetByID returns an exam by ID.
func (s *ExamService) GetByID(ctx context.Context, examID uuid.UUID) (*chapter.Exam, error) {
	return s.examRepo.FindByID(ctx, examID)
}

// ListByUser returns all exams for a user.
func (s *ExamService) ListByUser(ctx context.Context, userID uuid.UUID) ([]*chapter.Exam, error) {
	return s.examRepo.FindByUser(ctx, userID)
}

// FindActiveByChapter returns active exams linked to a chapter (for tightening).
func (s *ExamService) FindActiveByChapter(ctx context.Context, chapterID uuid.UUID) ([]*chapter.Exam, error) {
	return s.examRepo.FindActiveByChapter(ctx, chapterID)
}
