package models

// MCPApprovalCallbackRequest describes decision callback from external approver/executor adapters.
type MCPApprovalCallbackRequest struct {
	ApprovalRequestID int64  `json:"approval_request_id"`
	Decision          string `json:"decision"`
	Reason            string `json:"reason"`
	ActorID           string `json:"actor_id"`
}

// InteractionCallbackEnvelope describes one user interaction callback from an external adapter.
type InteractionCallbackEnvelope struct {
	SchemaVersion     string                      `json:"schema_version"`
	InteractionID     string                      `json:"interaction_id"`
	DeliveryID        string                      `json:"delivery_id,omitempty"`
	AdapterEventID    string                      `json:"adapter_event_id"`
	CallbackKind      string                      `json:"callback_kind"`
	OccurredAt        string                      `json:"occurred_at"`
	AdapterDeliveryID string                      `json:"adapter_delivery_id,omitempty"`
	DeliveryStatus    string                      `json:"delivery_status,omitempty"`
	Response          *InteractionResponsePayload `json:"response,omitempty"`
	Error             *InteractionCallbackError   `json:"error,omitempty"`
}

// InteractionResponsePayload stores one typed user response inside adapter callback payload.
type InteractionResponsePayload struct {
	ResponseKind     string `json:"response_kind"`
	SelectedOptionID string `json:"selected_option_id,omitempty"`
	FreeText         string `json:"free_text,omitempty"`
	ResponderRef     string `json:"responder_ref,omitempty"`
}

// InteractionCallbackError stores adapter-side failure details that do not alter classification directly.
type InteractionCallbackError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// InteractionCallbackOutcome is the typed HTTP response for interaction callbacks.
type InteractionCallbackOutcome struct {
	Accepted         bool   `json:"accepted"`
	Classification   string `json:"classification"`
	InteractionState string `json:"interaction_state"`
	ResumeRequired   bool   `json:"resume_required"`
	Message          string `json:"message,omitempty"`
}
