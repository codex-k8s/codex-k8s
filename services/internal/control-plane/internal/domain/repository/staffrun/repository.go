package staffrun

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

type (
	Run       = entitytypes.StaffRun
	FlowEvent = entitytypes.StaffFlowEvent
)

// Repository loads staff run state from PostgreSQL.
type Repository interface {
	// ListAll returns recent runs for platform admins.
	ListAll(ctx context.Context, limit int) ([]Run, error)
	// ListForUser returns recent runs for user's projects.
	ListForUser(ctx context.Context, userID string, limit int) ([]Run, error)
	// GetByID returns a run by id.
	GetByID(ctx context.Context, runID string) (Run, bool, error)
	// ListEventsByCorrelation returns flow events for a correlation id.
	ListEventsByCorrelation(ctx context.Context, correlationID string, limit int) ([]FlowEvent, error)
	// DeleteFlowEventsByProjectID removes flow_events linked to runs of a project.
	DeleteFlowEventsByProjectID(ctx context.Context, projectID string) error
	// GetCorrelationByRunID returns correlation id for a run id.
	GetCorrelationByRunID(ctx context.Context, runID string) (correlationID string, projectID string, ok bool, err error)
}
