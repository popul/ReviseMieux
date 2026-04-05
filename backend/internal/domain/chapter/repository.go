package chapter

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the port for chapter aggregate persistence.
type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Chapter, error)
	FindByUser(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]*Chapter, error)
	Save(ctx context.Context, ch *Chapter) error

	// Items
	FindItemByID(ctx context.Context, id uuid.UUID) (*Item, error)
	FindItemsByChapter(ctx context.Context, chapterID uuid.UUID, includeArchived bool) ([]*Item, error)
	SaveItem(ctx context.Context, item *Item) error
	SaveItems(ctx context.Context, items []*Item) error

	// Notions
	FindNotionsByChapter(ctx context.Context, chapterID uuid.UUID) ([]*Notion, error)
	SaveNotion(ctx context.Context, n *Notion) error

	// Revisions
	FindRevisionByID(ctx context.Context, id uuid.UUID) (*Revision, error)
	FindCurrentRevision(ctx context.Context, chapterID uuid.UUID) (*Revision, error)
	CountRevisionsByChapter(ctx context.Context, chapterID uuid.UUID) (int, error)
	SaveRevision(ctx context.Context, r *Revision) error

	// Pages
	FindPagesByRevision(ctx context.Context, revisionID uuid.UUID) ([]*Page, error)
	SavePage(ctx context.Context, p *Page) error
}

// ExamRepository is the port for exam aggregate persistence.
type ExamRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Exam, error)
	FindByUser(ctx context.Context, userID uuid.UUID) ([]*Exam, error)
	FindActiveByChapter(ctx context.Context, chapterID uuid.UUID) ([]*Exam, error)
	Save(ctx context.Context, exam *Exam) error
}
