package flowevent

import (
	"context"
	"encoding/json"
	"time"
)

// InsertParams defines a single audit-like flow event record.
type InsertParams struct {
	// CorrelationID links the event to a webhook processing flow.
	CorrelationID string
	// ActorType describes source category, for example "system".
	ActorType string
	// ActorID identifies the concrete actor inside ActorType.
	ActorID string
	// EventType stores canonical event name.
	EventType string
	// Payload contains structured event details.
	Payload json.RawMessage
	// CreatedAt defines event timestamp in UTC.
	CreatedAt time.Time
}

// Repository persists flow events.
type Repository interface {
	// Insert appends a flow event entry.
	Insert(ctx context.Context, params InsertParams) error
}
