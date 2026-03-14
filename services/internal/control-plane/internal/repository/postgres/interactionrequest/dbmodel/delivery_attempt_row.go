package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// DeliveryAttemptRow mirrors one interaction_delivery_attempts row.
type DeliveryAttemptRow struct {
	ID                     int64              `db:"id"`
	InteractionID          string             `db:"interaction_id"`
	ChannelBindingID       pgtype.Int8        `db:"channel_binding_id"`
	AttemptNo              int32              `db:"attempt_no"`
	DeliveryID             string             `db:"delivery_id"`
	AdapterKind            string             `db:"adapter_kind"`
	DeliveryRole           string             `db:"delivery_role"`
	Status                 string             `db:"status"`
	RequestEnvelopeJSON    []byte             `db:"request_envelope_json"`
	AckPayloadJSON         []byte             `db:"ack_payload_json"`
	AdapterDeliveryID      pgtype.Text        `db:"adapter_delivery_id"`
	ProviderMessageRefJSON []byte             `db:"provider_message_ref_json"`
	Retryable              bool               `db:"retryable"`
	NextRetryAt            pgtype.Timestamptz `db:"next_retry_at"`
	LastErrorCode          pgtype.Text        `db:"last_error_code"`
	ContinuationReason     pgtype.Text        `db:"continuation_reason"`
	StartedAt              time.Time          `db:"started_at"`
	FinishedAt             pgtype.Timestamptz `db:"finished_at"`
}
