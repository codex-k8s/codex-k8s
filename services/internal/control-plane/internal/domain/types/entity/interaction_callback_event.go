package entity

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionCallbackEvent stores one normalized callback evidence record.
type InteractionCallbackEvent struct {
	ID                      int64
	InteractionID           string
	ChannelBindingID        int64
	DeliveryID              string
	AdapterEventID          string
	CallbackKind            enumtypes.InteractionCallbackKind
	Classification          enumtypes.InteractionCallbackRecordClassification
	CallbackHandleHash      []byte
	NormalizedPayloadJSON   json.RawMessage
	RawPayloadJSON          json.RawMessage
	ProviderMessageRefJSON  json.RawMessage
	ProviderUpdateID        string
	ProviderCallbackQueryID string
	ReceivedAt              time.Time
	ProcessedAt             *time.Time
}
