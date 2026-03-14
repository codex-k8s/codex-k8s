package query

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionChannelBindingEnsureParams upserts the active Telegram binding for one interaction.
type InteractionChannelBindingEnsureParams struct {
	InteractionID          string
	AdapterKind            string
	RecipientRef           string
	CallbackTokenKeyID     string
	CallbackTokenExpiresAt *time.Time
}

// InteractionCallbackHandleUpsertItem describes one deterministic callback handle hash.
type InteractionCallbackHandleUpsertItem struct {
	HandleHash         []byte
	HandleKind         enumtypes.InteractionCallbackHandleKind
	OptionID           string
	ResponseDeadlineAt time.Time
	GraceExpiresAt     time.Time
}

// InteractionCallbackHandleUpsertParams inserts missing callback handles for the active binding.
type InteractionCallbackHandleUpsertParams struct {
	InteractionID    string
	ChannelBindingID int64
	Items            []InteractionCallbackHandleUpsertItem
}

// InteractionDispatchBindingUpdateParams stores adapter ack data for one binding.
type InteractionDispatchBindingUpdateParams struct {
	InteractionID          string
	DeliveryID             string
	AdapterKind            string
	ProviderMessageRefJSON json.RawMessage
	EditCapability         enumtypes.InteractionEditCapability
	CallbackTokenExpiresAt *time.Time
}
