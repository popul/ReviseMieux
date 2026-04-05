package eventbus

import (
	"context"
	"fmt"
	"log"

	"github.com/popul/revisemieux/internal/domain/event"
)

// Handler processes a single domain event.
type Handler func(ctx context.Context, evt event.Event) error

// SyncDispatcher is a synchronous event dispatcher that routes domain events
// to registered handlers. Suitable for Lot 0 local usage where eventual
// consistency is not needed.
//
// Each event is dispatched to all handlers registered for its EventName().
// If a handler returns an error, it is logged but dispatch continues to the
// remaining handlers (at-least-once delivery semantics within the same process).
type SyncDispatcher struct {
	handlers map[string][]Handler
}

// NewSyncDispatcher creates a new SyncDispatcher.
func NewSyncDispatcher() *SyncDispatcher {
	return &SyncDispatcher{
		handlers: make(map[string][]Handler),
	}
}

// Compile-time check.
var _ event.Publisher = (*SyncDispatcher)(nil)

// On registers a handler for events with the given name.
func (d *SyncDispatcher) On(eventName string, h Handler) {
	d.handlers[eventName] = append(d.handlers[eventName], h)
}

// Publish dispatches events to all registered handlers synchronously.
// Logs and continues on handler errors to avoid breaking the caller.
func (d *SyncDispatcher) Publish(ctx context.Context, events ...event.Event) error {
	var firstErr error
	for _, evt := range events {
		name := evt.EventName()
		log.Printf("[event] %s at %s", name, evt.OccurredAt())

		handlers, ok := d.handlers[name]
		if !ok {
			continue
		}
		for _, h := range handlers {
			if err := h(ctx, evt); err != nil {
				log.Printf("[event] handler error for %s: %v", name, err)
				if firstErr == nil {
					firstErr = fmt.Errorf("event handler %s: %w", name, err)
				}
			}
		}
	}
	return firstErr
}
