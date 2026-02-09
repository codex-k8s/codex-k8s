package staffrun

import "context"

// Run is a staff-visible run record.
type Run struct {
	ID            string
	CorrelationID string
	ProjectID     string
	Status        string
	CreatedAt     string
	StartedAt     string
	FinishedAt    string
}

// FlowEvent is a staff-visible flow event.
type FlowEvent struct {
	CorrelationID string
	EventType     string
	CreatedAt     string
	PayloadJSON   []byte
}

// Repository loads staff run state from PostgreSQL.
type Repository interface {
	// ListAll returns recent runs for platform admins.
	ListAll(ctx context.Context, limit int) ([]Run, error)
	// ListForUser returns recent runs for user's projects.
	ListForUser(ctx context.Context, userID string, limit int) ([]Run, error)
	// ListEventsByCorrelation returns flow events for a correlation id.
	ListEventsByCorrelation(ctx context.Context, correlationID string, limit int) ([]FlowEvent, error)
	// GetCorrelationByRunID returns correlation id for a run id.
	GetCorrelationByRunID(ctx context.Context, runID string) (correlationID string, projectID string, ok bool, err error)
}
