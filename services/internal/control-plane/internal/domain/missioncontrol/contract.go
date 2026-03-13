package missioncontrol

import (
	"context"

	missioncontrolrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/missioncontrol"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

type (
	WarmupRequest             = querytypes.MissionControlWarmupRequest
	WarmupSummary             = valuetypes.MissionControlWarmupSummary
	ActiveSetQuery            = querytypes.MissionControlActiveSetQuery
	EntityDetailsQuery        = querytypes.MissionControlEntityDetailsQuery
	SubmitCommandParams       = querytypes.MissionControlSubmitCommandParams
	ApprovalDecisionParams    = querytypes.MissionControlApprovalDecisionParams
	CommandQueueParams        = querytypes.MissionControlCommandQueueParams
	CommandSyncProgressParams = querytypes.MissionControlCommandSyncProgressParams
	CommandReconcileParams    = querytypes.MissionControlCommandReconcileParams
	CommandFailureParams      = querytypes.MissionControlCommandFailureParams
	CommandCancelParams       = querytypes.MissionControlCommandCancelParams
	UpsertEntityParams        = missioncontrolrepo.UpsertEntityParams
	UpdateEntityParams        = missioncontrolrepo.UpdateEntityParams
	ReplaceRelationsParams    = missioncontrolrepo.ReplaceRelationsParams
	UpsertTimelineEntryParams = missioncontrolrepo.UpsertTimelineEntryParams
	Entity                    = missioncontrolrepo.Entity
	Relation                  = missioncontrolrepo.Relation
	TimelineEntry             = missioncontrolrepo.TimelineEntry
	Command                   = missioncontrolrepo.Command
	ActiveSet                 = valuetypes.MissionControlActiveSet
	EntityDetails             = valuetypes.MissionControlEntityDetails
	CommandAdmission          = valuetypes.MissionControlCommandAdmission
	CommandStatusView         = valuetypes.MissionControlCommandStatusView
)

// WarmupExecutor defines the owner-controlled warmup/backfill entry-point that worker wave #371 must execute.
type WarmupExecutor interface {
	RunWarmup(ctx context.Context, params WarmupRequest) (WarmupSummary, error)
}

// DomainService defines Mission Control owner-owned use-cases consumed by future worker and transport waves.
type DomainService interface {
	WarmupExecutor
	ListActiveSet(ctx context.Context, params ActiveSetQuery) (ActiveSet, error)
	GetEntityDetails(ctx context.Context, params EntityDetailsQuery) (EntityDetails, error)
	UpsertEntity(ctx context.Context, params UpsertEntityParams, correlationID string) (Entity, error)
	UpdateEntityProjection(ctx context.Context, params UpdateEntityParams, correlationID string) (Entity, error)
	ReplaceRelationsForSource(ctx context.Context, params ReplaceRelationsParams, correlationID string) error
	UpsertTimelineEntry(ctx context.Context, params UpsertTimelineEntryParams, correlationID string) (TimelineEntry, error)
	SubmitCommand(ctx context.Context, params SubmitCommandParams) (CommandAdmission, error)
	GetCommandStatus(ctx context.Context, projectID string, commandID string) (CommandStatusView, error)
	QueueCommand(ctx context.Context, params CommandQueueParams) (Command, error)
	MarkCommandPendingSync(ctx context.Context, params CommandSyncProgressParams) (Command, error)
	MarkCommandReconciled(ctx context.Context, params CommandReconcileParams) (Command, error)
	MarkCommandFailed(ctx context.Context, params CommandFailureParams) (Command, error)
	CancelCommand(ctx context.Context, params CommandCancelParams) (Command, error)
	ApplyApprovalDecision(ctx context.Context, params ApprovalDecisionParams) (Command, error)
}
