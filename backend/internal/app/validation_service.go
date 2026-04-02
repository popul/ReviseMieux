package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
	"github.com/popul/revisemieux/internal/domain/validation"
)

const confidenceThreshold = 0.85

// All templates available for unrestricted items.
var allTemplates = []string{
	"GEN.KNOW.FLASH_MCQ",
	"GEN.KNOW.DEF_SHORT",
	"GEN.KNOW.CLOZE_KEYWORDS",
	"PC.FORMULA.APPLY",
	"PC.UNITS.CONVERT",
	"PC.FORMULA.ISOLATE",
	"GEN.WRITE.SHORT",
}

// Restricted templates for validation_required items (Z3-AC01).
var restrictedTemplates = []string{
	"GEN.KNOW.DEF_SHORT",
	"GEN.KNOW.FLASH_MCQ",
}

// ValidationService handles HITL validation use cases.
type ValidationService struct {
	valRepo     validation.Repository
	chapterRepo chapter.Repository
	publisher   event.Publisher
	clock       event.Clock
	idGen       event.IDGenerator
}

// NewValidationService creates a new ValidationService.
func NewValidationService(
	valRepo validation.Repository,
	chapterRepo chapter.Repository,
	publisher event.Publisher,
	clock event.Clock,
	idGen event.IDGenerator,
) *ValidationService {
	return &ValidationService{
		valRepo:     valRepo,
		chapterRepo: chapterRepo,
		publisher:   publisher,
		clock:       clock,
		idGen:       idGen,
	}
}

// IsChapterUsable returns whether a chapter can be used for diagnostic/sessions,
// even if validation tasks are pending (Z3-AC06).
func (s *ValidationService) IsChapterUsable(ctx context.Context, chapterID uuid.UUID) (bool, int, error) {
	items, err := s.chapterRepo.FindItemsByChapter(ctx, chapterID, false)
	if err != nil {
		return false, 0, fmt.Errorf("validation_service: find items: %w", err)
	}

	pendingCount := 0
	for _, item := range items {
		if item.ValidationRequired {
			pendingCount++
		}
	}

	// Chapter is always usable — pending validations don't block (Z3-AC06)
	return true, pendingCount, nil
}

// Confirm resolves a validation task as confirmed (Z3-AC02).
func (s *ValidationService) Confirm(ctx context.Context, taskID, resolverID uuid.UUID) error {
	now := s.clock.Now()

	task, err := s.valRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("validation_service: find task: %w", err)
	}

	if err := task.Confirm(resolverID, now); err != nil {
		return fmt.Errorf("validation_service: %w", err)
	}
	if err := s.valRepo.Save(ctx, task); err != nil {
		return fmt.Errorf("validation_service: save task: %w", err)
	}

	// Update item: validation_required=false, confidence=max(current, 0.85)
	item, err := s.chapterRepo.FindItemByID(ctx, task.ItemID)
	if err != nil {
		return fmt.Errorf("validation_service: find item: %w", err)
	}
	item.ValidationRequired = false
	if item.Confidence < confidenceThreshold {
		item.Confidence = confidenceThreshold
	}
	item.UpdatedAt = now
	if err := s.chapterRepo.SaveItem(ctx, item); err != nil {
		return fmt.Errorf("validation_service: save item: %w", err)
	}

	// Publish ValidationResolved event → triggers cache invalidation
	s.publisher.Publish(ctx, event.ValidationResolved{
		BaseEvent: event.BaseEvent{OccurredOn: now},
		ItemID:    task.ItemID,
	})

	return nil
}

// Correct resolves a validation task with a correction (Z3-AC03).
func (s *ValidationService) Correct(ctx context.Context, taskID, resolverID uuid.UUID, correctedTerm *string) error {
	now := s.clock.Now()

	task, err := s.valRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("validation_service: find task: %w", err)
	}

	if err := task.Correct(resolverID, now); err != nil {
		return fmt.Errorf("validation_service: %w", err)
	}
	if err := s.valRepo.Save(ctx, task); err != nil {
		return fmt.Errorf("validation_service: save task: %w", err)
	}

	// Update item with correction
	item, err := s.chapterRepo.FindItemByID(ctx, task.ItemID)
	if err != nil {
		return fmt.Errorf("validation_service: find item: %w", err)
	}
	if correctedTerm != nil {
		item.Term = correctedTerm
	}
	item.ValidationRequired = false
	item.Confidence = 1.0 // human correction = max confidence
	item.UpdatedAt = now
	if err := s.chapterRepo.SaveItem(ctx, item); err != nil {
		return fmt.Errorf("validation_service: save item: %w", err)
	}

	s.publisher.Publish(ctx, event.ValidationResolved{
		BaseEvent: event.BaseEvent{OccurredOn: now},
		ItemID:    task.ItemID,
	})

	return nil
}

// MarkUnknown defers a task when student says "Je ne sais pas" (Z3-AC04).
// Creates an admin task with high priority.
func (s *ValidationService) MarkUnknown(ctx context.Context, taskID, resolverID uuid.UUID) error {
	now := s.clock.Now()

	task, err := s.valRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("validation_service: find task: %w", err)
	}

	if err := task.DeferByStudent(resolverID, now); err != nil {
		return fmt.Errorf("validation_service: %w", err)
	}
	if err := s.valRepo.Save(ctx, task); err != nil {
		return fmt.Errorf("validation_service: save task: %w", err)
	}

	// Create admin task with high priority
	adminTask := validation.NewValidationTask(s.idGen, task.ItemID, validation.SourceStudentDeferred, now)
	adminTask.Priority = 10 // high priority
	if err := s.valRepo.Save(ctx, adminTask); err != nil {
		return fmt.Errorf("validation_service: save admin task: %w", err)
	}

	return nil
}

// Ignore marks a task as ignored by the student (Z3-AC05).
func (s *ValidationService) Ignore(ctx context.Context, taskID, resolverID uuid.UUID) error {
	now := s.clock.Now()

	task, err := s.valRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("validation_service: find task: %w", err)
	}

	if err := task.IgnoreByStudent(resolverID, now); err != nil {
		return fmt.Errorf("validation_service: %w", err)
	}
	if err := s.valRepo.Save(ctx, task); err != nil {
		return fmt.Errorf("validation_service: save task: %w", err)
	}

	return nil
}

// maxValidationTasks is the cap on validation tasks per chapter (Z2-AC09).
const maxValidationTasks = 8

// CreateValidationTasksIfNeeded creates validation tasks for items below
// the confidence threshold (Z3-AC09), capped at 8 highest-priority items (Z2-AC09).
func (s *ValidationService) CreateValidationTasksIfNeeded(ctx context.Context, items []*chapter.Item) []*validation.ValidationTask {
	now := s.clock.Now()

	// Collect eligible items sorted by lowest confidence first (highest priority)
	type candidate struct {
		item       *chapter.Item
		confidence float32
	}
	var candidates []candidate
	for _, item := range items {
		if item.Confidence < confidenceThreshold {
			candidates = append(candidates, candidate{item: item, confidence: item.Confidence})
		}
	}

	// Sort by confidence ascending (lowest confidence = highest priority)
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].confidence < candidates[i].confidence {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Cap at maxValidationTasks (Z2-AC09)
	limit := len(candidates)
	if limit > maxValidationTasks {
		limit = maxValidationTasks
	}

	var tasks []*validation.ValidationTask
	for i := 0; i < limit; i++ {
		task := validation.NewValidationTask(s.idGen, candidates[i].item.ID, validation.SourceUncertainty, now)
		s.valRepo.Save(ctx, task)
		tasks = append(tasks, task)
	}

	return tasks
}

// GetByID returns a single validation task by its ID.
func (s *ValidationService) GetByID(ctx context.Context, taskID uuid.UUID) (*validation.ValidationTask, error) {
	return s.valRepo.FindByID(ctx, taskID)
}

// ListPending returns pending validation tasks.
func (s *ValidationService) ListPending(ctx context.Context, limit int) ([]*validation.ValidationTask, error) {
	return s.valRepo.FindPendingAll(ctx, limit)
}

// EligibleTemplates returns the templates available for an item based on
// its validation_required status (Z3-AC01).
func (s *ValidationService) EligibleTemplates(validationRequired bool) []string {
	if validationRequired {
		return restrictedTemplates
	}
	return allTemplates
}
