package chapter

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/event"
)

var (
	ErrExamNotFound = errors.New("exam: not found")
	ErrExamPast     = errors.New("exam: date is in the past")
)

// Exam represents a scheduled exam linked to one or more chapters (Z6-AC10).
type Exam struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Title      string
	ExamDate   time.Time
	ChapterIDs []uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewExam creates an exam linked to chapters.
func NewExam(idGen event.IDGenerator, userID uuid.UUID, title string, examDate time.Time, chapterIDs []uuid.UUID, now time.Time) *Exam {
	return &Exam{
		ID:         idGen.New(),
		UserID:     userID,
		Title:      title,
		ExamDate:   examDate,
		ChapterIDs: chapterIDs,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// AddChapter adds a chapter to the exam.
func (e *Exam) AddChapter(chapterID uuid.UUID, now time.Time) {
	for _, id := range e.ChapterIDs {
		if id == chapterID {
			return // already linked
		}
	}
	e.ChapterIDs = append(e.ChapterIDs, chapterID)
	e.UpdatedAt = now
}

// RemoveChapter removes a chapter from the exam.
func (e *Exam) RemoveChapter(chapterID uuid.UUID, now time.Time) {
	for i, id := range e.ChapterIDs {
		if id == chapterID {
			e.ChapterIDs = append(e.ChapterIDs[:i], e.ChapterIDs[i+1:]...)
			e.UpdatedAt = now
			return
		}
	}
}

// DaysUntil returns the number of days until the exam from the given time.
func (e *Exam) DaysUntil(from time.Time) int {
	d := e.ExamDate.Sub(from)
	return int(d.Hours() / 24)
}
