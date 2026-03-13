package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// CallbackEventRow mirrors one interaction_callback_events row.
type CallbackEventRow struct {
	ID                    int64              `db:"id"`
	InteractionID         string             `db:"interaction_id"`
	DeliveryID            pgtype.Text        `db:"delivery_id"`
	AdapterEventID        string             `db:"adapter_event_id"`
	CallbackKind          string             `db:"callback_kind"`
	Classification        string             `db:"classification"`
	NormalizedPayloadJSON []byte             `db:"normalized_payload_json"`
	RawPayloadJSON        []byte             `db:"raw_payload_json"`
	ReceivedAt            time.Time          `db:"received_at"`
	ProcessedAt           pgtype.Timestamptz `db:"processed_at"`
}
