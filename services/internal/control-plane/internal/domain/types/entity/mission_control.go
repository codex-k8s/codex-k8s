package entity

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// MissionControlEntity stores one persisted active-set projection row.
type MissionControlEntity struct {
	ID                int64
	ProjectID         string
	EntityKind        enumtypes.MissionControlEntityKind
	EntityExternalKey string
	ProviderKind      enumtypes.MissionControlProviderKind
	ProviderURL       string
	Title             string
	ActiveState       enumtypes.MissionControlActiveState
	SyncStatus        enumtypes.MissionControlSyncStatus
	ProjectionVersion int64
	CardPayloadJSON   json.RawMessage
	DetailPayloadJSON json.RawMessage
	LastTimelineAt    *time.Time
	ProviderUpdatedAt *time.Time
	ProjectedAt       time.Time
	StaleAfter        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// MissionControlRelation stores one typed graph edge between projection entities.
type MissionControlRelation struct {
	ID             int64
	ProjectID      string
	SourceEntityID int64
	RelationKind   enumtypes.MissionControlRelationKind
	TargetEntityID int64
	SourceKind     enumtypes.MissionControlRelationSourceKind
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// MissionControlTimelineEntry stores one append-only timeline projection item.
type MissionControlTimelineEntry struct {
	ID               int64
	ProjectID        string
	EntityID         int64
	SourceKind       enumtypes.MissionControlTimelineSourceKind
	EntryExternalKey string
	CommandID        string
	Summary          string
	BodyMarkdown     string
	PayloadJSON      json.RawMessage
	OccurredAt       time.Time
	ProviderURL      string
	IsReadOnly       bool
	CreatedAt        time.Time
}

// MissionControlCommand stores one command-ledger row.
type MissionControlCommand struct {
	ID                  string
	ProjectID           string
	CommandKind         enumtypes.MissionControlCommandKind
	TargetEntityID      *int64
	ActorID             string
	BusinessIntentKey   string
	CorrelationID       string
	Status              enumtypes.MissionControlCommandStatus
	FailureReason       enumtypes.MissionControlCommandFailureReason
	ApprovalRequestID   string
	ApprovalState       enumtypes.MissionControlApprovalState
	ApprovalRequestedAt *time.Time
	ApprovalDecidedAt   *time.Time
	PayloadJSON         json.RawMessage
	ResultPayloadJSON   json.RawMessage
	ProviderDeliveries  json.RawMessage
	LeaseOwner          string
	LeaseUntil          *time.Time
	RequestedAt         time.Time
	UpdatedAt           time.Time
	ReconciledAt        *time.Time
}
