package missioncontrol

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

type (
	Entity                    = entitytypes.MissionControlEntity
	Relation                  = entitytypes.MissionControlRelation
	TimelineEntry             = entitytypes.MissionControlTimelineEntry
	Command                   = entitytypes.MissionControlCommand
	UpsertEntityParams        = querytypes.MissionControlEntityUpsertParams
	UpdateEntityParams        = querytypes.MissionControlEntityProjectionUpdateParams
	EntityListFilter          = querytypes.MissionControlEntityListFilter
	RelationSeed              = querytypes.MissionControlRelationSeed
	ReplaceRelationsParams    = querytypes.MissionControlRelationReplaceParams
	UpsertTimelineEntryParams = querytypes.MissionControlTimelineEntryUpsertParams
	TimelineListFilter        = querytypes.MissionControlTimelineListFilter
	OptionalStringPatch       = querytypes.MissionControlOptionalStringPatch
	OptionalTimePatch         = querytypes.MissionControlOptionalTimePatch
	OptionalJSONPatch         = querytypes.MissionControlOptionalJSONPatch
	CommandFailureReasonPatch = querytypes.MissionControlCommandFailureReasonPatch
	CommandApprovalStatePatch = querytypes.MissionControlCommandApprovalStatePatch
	CreateCommandParams       = querytypes.MissionControlCommandCreateParams
	UpdateCommandStatusParams = querytypes.MissionControlCommandStatusUpdateParams
	CommandListFilter         = querytypes.MissionControlCommandListFilter
	WarmupSummary             = valuetypes.MissionControlWarmupSummary
)

// Repository persists Mission Control projection foundation under control-plane ownership.
type Repository interface {
	// UpsertEntity stores one projection row without optimistic concurrency checks.
	UpsertEntity(ctx context.Context, params UpsertEntityParams) (Entity, error)
	// UpdateEntityProjection stores one projection row guarded by expected projection_version.
	UpdateEntityProjection(ctx context.Context, params UpdateEntityParams) (Entity, error)
	// GetEntityByPublicID loads one entity by public identity tuple.
	GetEntityByPublicID(ctx context.Context, projectID string, entityKind enumtypes.MissionControlEntityKind, entityExternalKey string) (Entity, bool, error)
	// GetEntityByID loads one entity by internal persistence id without leaking it outside the domain.
	GetEntityByID(ctx context.Context, projectID string, entityID int64) (Entity, bool, error)
	// ListEntities returns active-set rows for one project with optional filters.
	ListEntities(ctx context.Context, filter EntityListFilter) ([]Entity, error)
	// ReplaceRelationsForSource rewrites relation edges for one source entity.
	ReplaceRelationsForSource(ctx context.Context, params ReplaceRelationsParams) error
	// ListRelationsForEntity returns edges where one entity is source or target.
	ListRelationsForEntity(ctx context.Context, projectID string, entityID int64) ([]Relation, error)
	// UpsertTimelineEntry stores one timeline projection row keyed by source external id.
	UpsertTimelineEntry(ctx context.Context, params UpsertTimelineEntryParams) (TimelineEntry, error)
	// ListTimelineEntries returns timeline entries for one entity ordered newest first.
	ListTimelineEntries(ctx context.Context, filter TimelineListFilter) ([]TimelineEntry, error)
	// CreateCommand inserts one command-ledger row.
	CreateCommand(ctx context.Context, params CreateCommandParams) (Command, error)
	// GetCommandByID loads one command row by id scoped to one project.
	GetCommandByID(ctx context.Context, projectID string, commandID string) (Command, bool, error)
	// GetCommandByBusinessIntent loads one command row by semantic dedupe key.
	GetCommandByBusinessIntent(ctx context.Context, projectID string, businessIntentKey string) (Command, bool, error)
	// ListCommands returns command rows for one project with optional status filter.
	ListCommands(ctx context.Context, filter CommandListFilter) ([]Command, error)
	// UpdateCommandStatus persists one command status transition.
	UpdateCommandStatus(ctx context.Context, params UpdateCommandStatusParams) (Command, bool, error)
	// GetWarmupSummary returns aggregate counts used by worker warmup verification.
	GetWarmupSummary(ctx context.Context, projectID string) (WarmupSummary, error)
}
