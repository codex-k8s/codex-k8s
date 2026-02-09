package flowevent

import (
	"context"
	"encoding/json"
	"time"
)

// InsertParams defines a single flow event record.
type InsertParams struct {
	// CorrelationID links event to run flow.
	CorrelationID string
	// ActorType describes event source category.
	ActorType string
	// ActorID identifies a concrete source.
	ActorID string
	// EventType stores canonical event name.
	EventType string
	// Payload contains structured event details.
	Payload json.RawMessage
	// CreatedAt is a stable event timestamp.
	CreatedAt time.Time
}

// Repository persists flow events.
type Repository interface {
	// Insert appends a flow event to the log.
	Insert(ctx context.Context, params InsertParams) error
}
