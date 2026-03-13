package value

import (
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// InteractionResumePayload is deterministic machine-readable result passed to future resume path.
type InteractionResumePayload struct {
	InteractionID    string                         `json:"interaction_id"`
	ToolName         string                         `json:"tool_name"`
	RequestStatus    enumtypes.InteractionRequestStatus `json:"request_status"`
	ResponseKind     enumtypes.InteractionResponseKind  `json:"response_kind"`
	SelectedOptionID string                         `json:"selected_option_id,omitempty"`
	FreeText         string                         `json:"free_text,omitempty"`
	ResolvedAt       string                         `json:"resolved_at"`
	ResolutionReason string                         `json:"resolution_reason"`
}
