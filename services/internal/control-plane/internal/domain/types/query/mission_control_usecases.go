package query

import (
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

// MissionControlActiveSetQuery defines one active-set read-model lookup.
type MissionControlActiveSetQuery struct {
	ProjectID    string
	ActiveStates []enumtypes.MissionControlActiveState
	SyncStatuses []enumtypes.MissionControlSyncStatus
	Limit        int
}

// MissionControlEntityDetailsQuery defines one entity details lookup.
type MissionControlEntityDetailsQuery struct {
	ProjectID      string
	EntityKind     enumtypes.MissionControlEntityKind
	EntityPublicID string
	TimelineLimit  int
}

// MissionControlSubmitCommandParams defines typed admission input for Mission Control commands.
type MissionControlSubmitCommandParams struct {
	ProjectID                 string
	ActorID                   string
	CorrelationID             string
	CommandKind               enumtypes.MissionControlCommandKind
	TargetEntityRef           *valuetypes.MissionControlEntityRef
	BusinessIntentKey         string
	ExpectedProjectionVersion int64
	Payload                   valuetypes.MissionControlCommandPayload
	RequestedAt               time.Time
}

// MissionControlApprovalDecisionParams defines one owner approval decision over a pending command.
type MissionControlApprovalDecisionParams struct {
	ProjectID       string
	CommandID       string
	Decision        enumtypes.MissionControlApprovalState
	ApproverActorID string
	StatusMessage   string
	UpdatedAt       time.Time
}

// MissionControlCommandQueueParams defines one transition into queued state.
type MissionControlCommandQueueParams struct {
	ProjectID     string
	CommandID     string
	StatusMessage string
	UpdatedAt     time.Time
}

// MissionControlCommandSyncProgressParams defines one transition into pending_sync state.
type MissionControlCommandSyncProgressParams struct {
	ProjectID           string
	CommandID           string
	StatusMessage       string
	ProviderDeliveryIDs []string
	UpdatedAt           time.Time
}

// MissionControlCommandReconcileParams defines one terminal successful reconcile transition.
type MissionControlCommandReconcileParams struct {
	ProjectID           string
	CommandID           string
	StatusMessage       string
	ProviderDeliveryIDs []string
	ReconciledAt        time.Time
	UpdatedAt           time.Time
}

// MissionControlCommandFailureParams defines one failed command transition.
type MissionControlCommandFailureParams struct {
	ProjectID           string
	CommandID           string
	FailureReason       enumtypes.MissionControlCommandFailureReason
	StatusMessage       string
	ProviderDeliveryIDs []string
	UpdatedAt           time.Time
}

// MissionControlCommandCancelParams defines one cancelled command transition.
type MissionControlCommandCancelParams struct {
	ProjectID     string
	CommandID     string
	StatusMessage string
	UpdatedAt     time.Time
}
