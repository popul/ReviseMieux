package eventbus

import (
	"context"
	"log"

	"github.com/popul/revisemieux/internal/domain/event"
)

var _ event.Publisher = (*LogPublisher)(nil)

// LogPublisher is a simple event publisher that logs events.
// Suitable for Lot 0 local usage. Replace with a proper message broker later.
type LogPublisher struct{}

// NewLogPublisher creates a new LogPublisher.
func NewLogPublisher() *LogPublisher {
	return &LogPublisher{}
}

// Publish logs the events for Lot 0 — no async processing needed yet.
func (p *LogPublisher) Publish(_ context.Context, events ...event.Event) error {
	for _, e := range events {
		log.Printf("[event] %s at %s", e.EventName(), e.OccurredAt())
	}
	return nil
}
