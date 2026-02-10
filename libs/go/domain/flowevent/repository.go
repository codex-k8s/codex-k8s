package flowevent

import (
	"context"
	"encoding/json"
	"time"
)

// InsertParams defines a single flow event record.
type InsertParams struct {
	CorrelationID string
	ActorType     string
	ActorID       string
	EventType     string
	Payload       json.RawMessage
	CreatedAt     time.Time
}

// Repository persists flow events.
type Repository interface {
	Insert(ctx context.Context, params InsertParams) error
}
