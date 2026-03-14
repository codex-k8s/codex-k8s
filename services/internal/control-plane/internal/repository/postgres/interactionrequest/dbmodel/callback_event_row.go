package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// CallbackEventRow mirrors one interaction_callback_events row.
type CallbackEventRow struct {
	ID                      int64              `db:"id"`
	InteractionID           string             `db:"interaction_id"`
	ChannelBindingID        pgtype.Int8        `db:"channel_binding_id"`
	DeliveryID              pgtype.Text        `db:"delivery_id"`
	AdapterEventID          string             `db:"adapter_event_id"`
	CallbackKind            string             `db:"callback_kind"`
	Classification          string             `db:"classification"`
	CallbackHandleHash      []byte             `db:"callback_handle_hash"`
	NormalizedPayloadJSON   []byte             `db:"normalized_payload_json"`
	RawPayloadJSON          []byte             `db:"raw_payload_json"`
	ProviderMessageRefJSON  []byte             `db:"provider_message_ref_json"`
	ProviderUpdateID        pgtype.Text        `db:"provider_update_id"`
	ProviderCallbackQueryID pgtype.Text        `db:"provider_callback_query_id"`
	ReceivedAt              time.Time          `db:"received_at"`
	ProcessedAt             pgtype.Timestamptz `db:"processed_at"`
}
