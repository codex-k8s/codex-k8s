package githubratelimitwait

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

type (
	Wait                    = entitytypes.GitHubRateLimitWait
	Evidence                = entitytypes.GitHubRateLimitWaitEvidence
	CreateWaitParams        = querytypes.GitHubRateLimitWaitCreateParams
	UpdateWaitParams        = querytypes.GitHubRateLimitWaitUpdateParams
	CreateEvidenceParams    = querytypes.GitHubRateLimitWaitEvidenceCreateParams
	RefreshProjectionResult = valuetypes.GitHubRateLimitProjectionRefreshResult
)

// Repository persists rate-limit waits, evidence, and dominant run linkage for GitHub resilience.
type Repository interface {
	// Create inserts one new wait aggregate.
	Create(ctx context.Context, params CreateWaitParams) (Wait, error)
	// Update mutates one existing wait aggregate by id.
	Update(ctx context.Context, params UpdateWaitParams) (Wait, bool, error)
	// GetByID returns one wait aggregate by id.
	GetByID(ctx context.Context, waitID string) (Wait, bool, error)
	// GetBySignalID returns one wait aggregate by latest signal id.
	GetBySignalID(ctx context.Context, signalID string) (Wait, bool, error)
	// GetOpenByRunAndContour returns one open wait for run+contour when present.
	GetOpenByRunAndContour(ctx context.Context, runID string, contourKind enumtypes.GitHubRateLimitContourKind) (Wait, bool, error)
	// ListByRunID returns all waits for one run ordered by newest update first.
	ListByRunID(ctx context.Context, runID string) ([]Wait, error)
	// AppendEvidence inserts one append-only evidence row.
	AppendEvidence(ctx context.Context, params CreateEvidenceParams) (Evidence, error)
	// RefreshRunProjection elects dominant open wait and synchronizes typed run/session wait linkage.
	RefreshRunProjection(ctx context.Context, runID string) (RefreshProjectionResult, error)
}
