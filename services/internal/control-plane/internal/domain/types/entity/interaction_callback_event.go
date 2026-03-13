package entity

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionCallbackEvent stores one normalized callback evidence record.
type InteractionCallbackEvent struct {
	ID                     int64
	InteractionID          string
	DeliveryID             string
	AdapterEventID         string
	CallbackKind           enumtypes.InteractionCallbackKind
	Classification         enumtypes.InteractionCallbackRecordClassification
	NormalizedPayloadJSON  json.RawMessage
	RawPayloadJSON         json.RawMessage
	ReceivedAt             time.Time
	ProcessedAt            *time.Time
}
