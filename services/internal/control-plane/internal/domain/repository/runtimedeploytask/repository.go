package runtimedeploytask

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	Task                = entitytypes.RuntimeDeployTask
	UpsertDesiredParams = querytypes.RuntimeDeployTaskUpsertDesiredParams
	ClaimParams         = querytypes.RuntimeDeployTaskClaimParams
	MarkSucceededParams = querytypes.RuntimeDeployTaskMarkSucceededParams
	MarkFailedParams    = querytypes.RuntimeDeployTaskMarkFailedParams
	RenewLeaseParams    = querytypes.RuntimeDeployTaskRenewLeaseParams
	ListFilter          = querytypes.RuntimeDeployTaskListFilter
	AppendLogParams     = querytypes.RuntimeDeployTaskAppendLogParams
)

// Repository persists runtime deployment desired and actual state.
type Repository interface {
	// UpsertDesired creates or updates desired state for one run-bound deploy task.
	UpsertDesired(ctx context.Context, params UpsertDesiredParams) (Task, error)
	// GetByRunID returns one runtime deploy task by run id.
	GetByRunID(ctx context.Context, runID string) (Task, bool, error)
	// ClaimNext acquires one pending/expired-running task lease for reconciler processing.
	ClaimNext(ctx context.Context, params ClaimParams) (Task, bool, error)
	// MarkSucceeded sets terminal success state for one leased task.
	MarkSucceeded(ctx context.Context, params MarkSucceededParams) (bool, error)
	// MarkFailed sets terminal failed state for one leased task.
	MarkFailed(ctx context.Context, params MarkFailedParams) (bool, error)
	// RenewLease extends active lease for one running task.
	RenewLease(ctx context.Context, params RenewLeaseParams) (bool, error)
	// ListRecent returns recent runtime deploy tasks ordered by update time.
	ListRecent(ctx context.Context, filter ListFilter) ([]Task, error)
	// AppendLog appends one task log line and keeps latest MaxLines entries.
	AppendLog(ctx context.Context, params AppendLogParams) error
}
