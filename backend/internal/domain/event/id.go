package event

import "github.com/google/uuid"

// IDGenerator is a port for generating unique identifiers (UUIDv7).
type IDGenerator interface {
	New() uuid.UUID
}

// UUIDv7Generator is the production implementation using google/uuid.
type UUIDv7Generator struct{}

func (UUIDv7Generator) New() uuid.UUID { return uuid.Must(uuid.NewV7()) }
