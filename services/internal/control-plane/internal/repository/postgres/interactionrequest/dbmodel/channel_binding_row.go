package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// ChannelBindingRow mirrors one interaction_channel_bindings row.
type ChannelBindingRow struct {
	ID                     int64              `db:"id"`
	InteractionID          string             `db:"interaction_id"`
	AdapterKind            string             `db:"adapter_kind"`
	RecipientRef           string             `db:"recipient_ref"`
	ProviderChatRef        pgtype.Text        `db:"provider_chat_ref"`
	ProviderMessageRefJSON []byte             `db:"provider_message_ref_json"`
	CallbackTokenKeyID     pgtype.Text        `db:"callback_token_key_id"`
	CallbackTokenExpiresAt pgtype.Timestamptz `db:"callback_token_expires_at"`
	EditCapability         string             `db:"edit_capability"`
	ContinuationState      string             `db:"continuation_state"`
	LastOperatorSignalCode pgtype.Text        `db:"last_operator_signal_code"`
	LastOperatorSignalAt   pgtype.Timestamptz `db:"last_operator_signal_at"`
	CreatedAt              time.Time          `db:"created_at"`
	UpdatedAt              time.Time          `db:"updated_at"`
}
