package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// RequestRow mirrors one interaction_requests row.
type RequestRow struct {
	ID                     string             `db:"id"`
	ProjectID              string             `db:"project_id"`
	RunID                  string             `db:"run_id"`
	InteractionKind        string             `db:"interaction_kind"`
	ChannelFamily          string             `db:"channel_family"`
	State                  string             `db:"state"`
	ResolutionKind         string             `db:"resolution_kind"`
	RecipientProvider      string             `db:"recipient_provider"`
	RecipientRef           string             `db:"recipient_ref"`
	RequestPayloadJSON     []byte             `db:"request_payload_json"`
	ContextLinksJSON       []byte             `db:"context_links_json"`
	ResponseDeadlineAt     pgtype.Timestamptz `db:"response_deadline_at"`
	EffectiveResponseID    pgtype.Int8        `db:"effective_response_id"`
	ActiveChannelBindingID pgtype.Int8        `db:"active_channel_binding_id"`
	OperatorState          string             `db:"operator_state"`
	OperatorSignalCode     pgtype.Text        `db:"operator_signal_code"`
	OperatorSignalAt       pgtype.Timestamptz `db:"operator_signal_at"`
	LastDeliveryAttemptNo  int32              `db:"last_delivery_attempt_no"`
	CreatedAt              time.Time          `db:"created_at"`
	UpdatedAt              time.Time          `db:"updated_at"`
}
