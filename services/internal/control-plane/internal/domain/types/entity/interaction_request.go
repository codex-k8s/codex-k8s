package entity

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionRequest stores canonical user interaction aggregate.
type InteractionRequest struct {
	ID                    string
	ProjectID             string
	RunID                 string
	InteractionKind       enumtypes.InteractionKind
	State                 enumtypes.InteractionState
	ResolutionKind        enumtypes.InteractionResolutionKind
	RecipientProvider     string
	RecipientRef          string
	RequestPayloadJSON    json.RawMessage
	ContextLinksJSON      json.RawMessage
	ResponseDeadlineAt    *time.Time
	EffectiveResponseID   int64
	LastDeliveryAttemptNo int
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
