package entity

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionDeliveryAttempt stores one outbound dispatch attempt snapshot.
type InteractionDeliveryAttempt struct {
	ID                  int64
	InteractionID       string
	AttemptNo           int
	DeliveryID          string
	AdapterKind         string
	Status              enumtypes.InteractionDeliveryAttemptStatus
	RequestEnvelopeJSON json.RawMessage
	AckPayloadJSON      json.RawMessage
	AdapterDeliveryID   string
	Retryable           bool
	NextRetryAt         *time.Time
	LastErrorCode       string
	StartedAt           time.Time
	FinishedAt          *time.Time
}
