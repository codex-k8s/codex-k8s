package staffrun

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	Run        = entitytypes.StaffRun
	FlowEvent  = entitytypes.StaffFlowEvent
	RunLogs    = entitytypes.StaffRunLogs
	ListFilter = querytypes.StaffRunListFilter
)

// Repository loads staff run state from PostgreSQL.
type Repository interface {
	// ListAll returns recent runs for platform admins.
	ListAll(ctx context.Context, limit int) ([]Run, error)
	// ListForUser returns recent runs for user's projects.
	ListForUser(ctx context.Context, userID string, limit int) ([]Run, error)
	// ListJobsAll returns runtime jobs list for platform admins.
	ListJobsAll(ctx context.Context, filter ListFilter) ([]Run, error)
	// ListJobsForUser returns runtime jobs list scoped to user projects.
	ListJobsForUser(ctx context.Context, userID string, filter ListFilter) ([]Run, error)
	// ListWaitsAll returns wait queue list for platform admins.
	ListWaitsAll(ctx context.Context, filter ListFilter) ([]Run, error)
	// ListWaitsForUser returns wait queue list scoped to user projects.
	ListWaitsForUser(ctx context.Context, userID string, filter ListFilter) ([]Run, error)
	// GetByID returns a run by id.
	GetByID(ctx context.Context, runID string) (Run, bool, error)
	// GetLogsByRunID returns one run logs snapshot by run id.
	GetLogsByRunID(ctx context.Context, runID string) (RunLogs, bool, error)
	// ListEventsByCorrelation returns flow events for a correlation id.
	ListEventsByCorrelation(ctx context.Context, correlationID string, limit int) ([]FlowEvent, error)
	// DeleteFlowEventsByProjectID removes flow_events linked to runs of a project.
	DeleteFlowEventsByProjectID(ctx context.Context, projectID string) error
	// GetCorrelationByRunID returns correlation id for a run id.
	GetCorrelationByRunID(ctx context.Context, runID string) (correlationID string, projectID string, ok bool, err error)
}
