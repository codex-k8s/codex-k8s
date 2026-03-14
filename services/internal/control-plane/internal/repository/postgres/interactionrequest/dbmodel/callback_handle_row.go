package dbmodel

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// CallbackHandleRow mirrors one interaction_callback_handles row.
type CallbackHandleRow struct {
	ID                  int64              `db:"id"`
	InteractionID       string             `db:"interaction_id"`
	ChannelBindingID    int64              `db:"channel_binding_id"`
	HandleHash          []byte             `db:"handle_hash"`
	HandleKind          string             `db:"handle_kind"`
	OptionID            pgtype.Text        `db:"option_id"`
	State               string             `db:"state"`
	ResponseDeadlineAt  time.Time          `db:"response_deadline_at"`
	GraceExpiresAt      time.Time          `db:"grace_expires_at"`
	UsedCallbackEventID pgtype.Int8        `db:"used_callback_event_id"`
	UsedAt              pgtype.Timestamptz `db:"used_at"`
	CreatedAt           time.Time          `db:"created_at"`
}
