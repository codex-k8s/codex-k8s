package entity

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionChannelBinding stores one channel-specific delivery binding for an interaction.
type InteractionChannelBinding struct {
	ID                     int64
	InteractionID          string
	AdapterKind            string
	RecipientRef           string
	ProviderChatRef        string
	ProviderMessageRefJSON json.RawMessage
	CallbackTokenKeyID     string
	CallbackTokenExpiresAt *time.Time
	EditCapability         enumtypes.InteractionEditCapability
	ContinuationState      enumtypes.InteractionContinuationState
	LastOperatorSignalCode enumtypes.InteractionOperatorSignalCode
	LastOperatorSignalAt   *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
