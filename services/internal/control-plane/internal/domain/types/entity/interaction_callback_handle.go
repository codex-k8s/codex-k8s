package entity

import (
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionCallbackHandle stores one hashed Telegram callback or free-text session handle.
type InteractionCallbackHandle struct {
	ID                  int64
	InteractionID       string
	ChannelBindingID    int64
	HandleHash          []byte
	HandleKind          enumtypes.InteractionCallbackHandleKind
	OptionID            string
	State               enumtypes.InteractionCallbackHandleState
	ResponseDeadlineAt  time.Time
	GraceExpiresAt      time.Time
	UsedCallbackEventID int64
	UsedAt              *time.Time
	CreatedAt           time.Time
}
