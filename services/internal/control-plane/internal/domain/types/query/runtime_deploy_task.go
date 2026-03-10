package query

import (
	"time"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

// RuntimeDeployTaskUpsertDesiredParams describes one desired runtime deployment state update.
type RuntimeDeployTaskUpsertDesiredParams struct {
	RunID              string
	RuntimeMode        string
	Namespace          string
	TargetEnv          string
	SlotNo             int
	RepositoryFullName string
	ServicesYAMLPath   string
	BuildRef           string
	DeployOnly         bool
}

// RuntimeDeployTaskClaimParams describes one runtime deploy lease claim request.
type RuntimeDeployTaskClaimParams struct {
	LeaseOwner          string
	LeaseTTL            string
	StaleRunningTimeout string
}

// RuntimeDeployTaskMarkSucceededParams describes one successful deployment completion.
type RuntimeDeployTaskMarkSucceededParams struct {
	RunID           string
	LeaseOwner      string
	ResultNamespace string
	ResultTargetEnv string
}

// RuntimeDeployTaskMarkFailedParams describes one failed deployment completion.
type RuntimeDeployTaskMarkFailedParams struct {
	RunID      string
	LeaseOwner string
	LastError  string
}

// RuntimeDeployTaskRenewLeaseParams describes one active lease extension.
type RuntimeDeployTaskRenewLeaseParams struct {
	RunID      string
	LeaseOwner string
	LeaseTTL   string
}

// RuntimeDeployTaskRequeueParams describes one running task requeue request for graceful handover.
type RuntimeDeployTaskRequeueParams struct {
	RunID      string
	LeaseOwner string
	LastError  string
}

// RuntimeDeployTaskAction identifies one operator-requested terminal action.
type RuntimeDeployTaskAction string

const (
	RuntimeDeployTaskActionCancel RuntimeDeployTaskAction = "cancel"
	RuntimeDeployTaskActionStop   RuntimeDeployTaskAction = "stop"
)

// RuntimeDeployTaskRequestActionParams describes one cancel/stop request.
type RuntimeDeployTaskRequestActionParams struct {
	RunID       string
	Action      RuntimeDeployTaskAction
	RequestedAt time.Time
	RequestedBy string
	Reason      string
}

// RuntimeDeployTaskRequestActionResult describes one action mutation result.
type RuntimeDeployTaskRequestActionResult struct {
	Task            entitytypes.RuntimeDeployTask
	PreviousStatus  entitytypes.RuntimeDeployTaskStatus
	CurrentStatus   entitytypes.RuntimeDeployTaskStatus
	AlreadyTerminal bool
}

// RuntimeDeployTaskListFilter describes optional task list filters.
type RuntimeDeployTaskListFilter struct {
	Limit     int
	Status    string
	TargetEnv string
}

// RuntimeDeployTaskAppendLogParams appends one log line for a task.
type RuntimeDeployTaskAppendLogParams struct {
	RunID    string
	Stage    string
	Level    string
	Message  string
	MaxLines int
}
