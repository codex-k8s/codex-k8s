package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// RequestRow mirrors one interaction_requests row.
type RequestRow struct {
	ID                    string             `db:"id"`
	ProjectID             string             `db:"project_id"`
	RunID                 string             `db:"run_id"`
	InteractionKind       string             `db:"interaction_kind"`
	State                 string             `db:"state"`
	ResolutionKind        string             `db:"resolution_kind"`
	RecipientProvider     string             `db:"recipient_provider"`
	RecipientRef          string             `db:"recipient_ref"`
	RequestPayloadJSON    []byte             `db:"request_payload_json"`
	ContextLinksJSON      []byte             `db:"context_links_json"`
	ResponseDeadlineAt    pgtype.Timestamptz `db:"response_deadline_at"`
	EffectiveResponseID   pgtype.Int8        `db:"effective_response_id"`
	LastDeliveryAttemptNo int32              `db:"last_delivery_attempt_no"`
	CreatedAt             time.Time          `db:"created_at"`
	UpdatedAt             time.Time          `db:"updated_at"`
}
