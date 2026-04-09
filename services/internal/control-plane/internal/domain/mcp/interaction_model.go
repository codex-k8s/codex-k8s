package mcp

import (
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

const (
	ToolMCPUserNotify          ToolName = "user.notify"
	ToolMCPUserDecisionRequest ToolName = "user.decision.request"
)

const (
	interactionToolStatusAccepted            = "accepted"
	interactionToolStatusPendingUserResponse = "pending_user_response"
	interactionDeliveryStateQueued           = "queued"
)

// UserNotificationKind is an audit-safe notification semantic.
type UserNotificationKind string

const (
	UserNotificationKindCompletion   UserNotificationKind = "completion"
	UserNotificationKindNextStep     UserNotificationKind = "next_step"
	UserNotificationKindStatusUpdate UserNotificationKind = "status_update"
	UserNotificationKindWarning      UserNotificationKind = "warning"
)

// UserNotifyInput describes one async built-in notification request.
type UserNotifyInput struct {
	NotificationKind UserNotificationKind `json:"notification_kind"`
	Summary          string               `json:"summary"`
	DetailsMarkdown  string               `json:"details_markdown,omitempty"`
	ActionLabel      string               `json:"action_label,omitempty"`
	ActionURL        string               `json:"action_url,omitempty"`
}

// UserNotifyResult is output for user.notify tool.
type UserNotifyResult struct {
	Status        string `json:"status"`
	InteractionID string `json:"interaction_id"`
	DeliveryState string `json:"delivery_state"`
	Message       string `json:"message,omitempty"`
}

// UserDecisionOption describes one machine-readable decision alternative.
type UserDecisionOption struct {
	OptionID    string `json:"option_id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// UserDecisionRequestInput describes one blocking built-in decision request.
type UserDecisionRequestInput struct {
	Question            string               `json:"question"`
	DetailsMarkdown     string               `json:"details_markdown,omitempty"`
	Options             []UserDecisionOption `json:"options"`
	AllowFreeText       bool                 `json:"allow_free_text,omitempty"`
	FreeTextPlaceholder string               `json:"free_text_placeholder,omitempty"`
	ResponseTTLSeconds  int32                `json:"response_ttl_seconds"`
}

// UserDecisionRequestResult is output for user.decision.request tool.
type UserDecisionRequestResult struct {
	Status        string `json:"status"`
	InteractionID string `json:"interaction_id"`
	WaitState     string `json:"wait_state"`
	WaitReason    string `json:"wait_reason"`
	ExpiresAt     string `json:"expires_at"`
}

// SubmitInteractionCallbackParams describes normalized callback processing request for future transport adapters.
type SubmitInteractionCallbackParams = querytypes.InteractionCallbackApplyParams

// SubmitInteractionCallbackResult describes callback classification outcome.
type SubmitInteractionCallbackResult struct {
	Accepted            bool                                              `json:"accepted"`
	Classification      enumtypes.InteractionCallbackResultClassification `json:"classification"`
	InteractionState    string                                            `json:"interaction_state"`
	ResumeRequired      bool                                              `json:"resume_required"`
	ContinuationAction  enumtypes.InteractionContinuationAction           `json:"continuation_action"`
	EffectiveResponseID int64                                             `json:"effective_response_id,omitempty"`
	ResumePayload       *valuetypes.InteractionResumePayload              `json:"resume_payload,omitempty"`
}
