package handler_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/popul/revisemieux/internal/domain/chapter"
	"github.com/popul/revisemieux/internal/domain/event"
)

// stubPublisher is a no-op event publisher for handler tests.
type stubPublisher struct {
	events []event.Event
}

func (p *stubPublisher) Publish(_ context.Context, events ...event.Event) error {
	p.events = append(p.events, events...)
	return nil
}

// mockChapterRepoWithItem is a mock that returns a valid Item for FindItemByID.
// Used by validation tests where Confirm/Correct need to update items.
type mockChapterRepoWithItem struct {
	mockChapterRepo
}

func (m *mockChapterRepoWithItem) FindItemByID(_ context.Context, id uuid.UUID) (*chapter.Item, error) {
	term := "test item"
	return &chapter.Item{
		ID:        id,
		ChapterID: uuid.Must(uuid.NewV7()),
		ItemType:  chapter.ItemKnowledge,
		Term:      &term,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockChapterRepoWithItem) SaveItem(_ context.Context, _ *chapter.Item) error { return nil }
