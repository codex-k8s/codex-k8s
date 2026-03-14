package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// EntityRow mirrors one mission_control_entities row.
type EntityRow struct {
	ID                int64              `db:"id"`
	ProjectID         string             `db:"project_id"`
	EntityKind        string             `db:"entity_kind"`
	EntityExternalKey string             `db:"entity_external_key"`
	ProviderKind      string             `db:"provider_kind"`
	ProviderURL       pgtype.Text        `db:"provider_url"`
	Title             string             `db:"title"`
	ActiveState       string             `db:"active_state"`
	SyncStatus        string             `db:"sync_status"`
	ProjectionVersion int64              `db:"projection_version"`
	CardPayloadJSON   []byte             `db:"card_payload_json"`
	DetailPayloadJSON []byte             `db:"detail_payload_json"`
	LastTimelineAt    pgtype.Timestamptz `db:"last_timeline_at"`
	ProviderUpdatedAt pgtype.Timestamptz `db:"provider_updated_at"`
	ProjectedAt       time.Time          `db:"projected_at"`
	StaleAfter        pgtype.Timestamptz `db:"stale_after"`
	CreatedAt         time.Time          `db:"created_at"`
	UpdatedAt         time.Time          `db:"updated_at"`
}

// RelationRow mirrors one mission_control_relations row.
type RelationRow struct {
	ID             int64     `db:"id"`
	ProjectID      string    `db:"project_id"`
	SourceEntityID int64     `db:"source_entity_id"`
	RelationKind   string    `db:"relation_kind"`
	TargetEntityID int64     `db:"target_entity_id"`
	SourceKind     string    `db:"source_kind"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// TimelineEntryRow mirrors one mission_control_timeline_entries row.
type TimelineEntryRow struct {
	ID               int64       `db:"id"`
	ProjectID        string      `db:"project_id"`
	EntityID         int64       `db:"entity_id"`
	SourceKind       string      `db:"source_kind"`
	EntryExternalKey string      `db:"entry_external_key"`
	CommandID        pgtype.Text `db:"command_id"`
	Summary          string      `db:"summary"`
	BodyMarkdown     pgtype.Text `db:"body_markdown"`
	PayloadJSON      []byte      `db:"payload_json"`
	OccurredAt       time.Time   `db:"occurred_at"`
	ProviderURL      pgtype.Text `db:"provider_url"`
	IsReadOnly       bool        `db:"is_read_only"`
	CreatedAt        time.Time   `db:"created_at"`
}

// CommandRow mirrors one mission_control_commands row.
type CommandRow struct {
	ID                  string             `db:"id"`
	ProjectID           string             `db:"project_id"`
	CommandKind         string             `db:"command_kind"`
	TargetEntityID      pgtype.Int8        `db:"target_entity_id"`
	ActorID             string             `db:"actor_id"`
	BusinessIntentKey   string             `db:"business_intent_key"`
	CorrelationID       string             `db:"correlation_id"`
	Status              string             `db:"status"`
	FailureReason       pgtype.Text        `db:"failure_reason"`
	ApprovalRequestID   pgtype.Text        `db:"approval_request_id"`
	ApprovalState       string             `db:"approval_state"`
	ApprovalRequestedAt pgtype.Timestamptz `db:"approval_requested_at"`
	ApprovalDecidedAt   pgtype.Timestamptz `db:"approval_decided_at"`
	PayloadJSON         []byte             `db:"payload_json"`
	ResultPayloadJSON   []byte             `db:"result_payload_json"`
	ProviderDeliveries  []byte             `db:"provider_deliveries_json"`
	LeaseOwner          pgtype.Text        `db:"lease_owner"`
	LeaseUntil          pgtype.Timestamptz `db:"lease_until"`
	RequestedAt         time.Time          `db:"requested_at"`
	UpdatedAt           time.Time          `db:"updated_at"`
	ReconciledAt        pgtype.Timestamptz `db:"reconciled_at"`
}

// WarmupSummaryRow mirrors aggregate mission-control warmup summary values.
type WarmupSummaryRow struct {
	ProjectID            string `db:"project_id"`
	EntityCount          int64  `db:"entity_count"`
	RelationCount        int64  `db:"relation_count"`
	TimelineEntryCount   int64  `db:"timeline_entry_count"`
	CommandCount         int64  `db:"command_count"`
	MaxProjectionVersion int64  `db:"max_projection_version"`
}
