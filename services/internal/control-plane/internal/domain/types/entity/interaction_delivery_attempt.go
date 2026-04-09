package entity

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionDeliveryAttempt stores one outbound dispatch attempt snapshot.
type InteractionDeliveryAttempt struct {
	ID                     int64
	InteractionID          string
	ChannelBindingID       int64
	AttemptNo              int
	DeliveryID             string
	AdapterKind            string
	DeliveryRole           enumtypes.InteractionDeliveryRole
	Status                 enumtypes.InteractionDeliveryAttemptStatus
	RequestEnvelopeJSON    json.RawMessage
	AckPayloadJSON         json.RawMessage
	AdapterDeliveryID      string
	ProviderMessageRefJSON json.RawMessage
	Retryable              bool
	NextRetryAt            *time.Time
	LastErrorCode          string
	ContinuationReason     string
	StartedAt              time.Time
	FinishedAt             *time.Time
}
