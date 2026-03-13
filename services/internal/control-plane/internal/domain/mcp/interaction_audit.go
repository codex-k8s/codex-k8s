package mcp

import (
	"context"
	"encoding/json"
	"strings"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

type interactionRequestEventPayload struct {
	InteractionID string `json:"interaction_id"`
	ToolName      string `json:"tool_name"`
	RunID         string `json:"run_id"`
}

type interactionCallbackEventPayload struct {
	InteractionID  string `json:"interaction_id"`
	AdapterEventID string `json:"adapter_event_id"`
	CallbackKind   string `json:"callback_kind"`
}

type interactionResponseEventPayload struct {
	InteractionID    string `json:"interaction_id"`
	ResponseRecordID int64  `json:"response_record_id,omitempty"`
	ResponseKind     string `json:"response_kind,omitempty"`
	Classification   string `json:"classification,omitempty"`
}

type interactionWaitEventPayload struct {
	RunID          string `json:"run_id"`
	InteractionID  string `json:"interaction_id"`
	WaitDeadlineAt string `json:"wait_deadline_at,omitempty"`
	RequestStatus  string `json:"request_status,omitempty"`
}

func (s *Service) auditInteractionRequestCreated(ctx context.Context, session SessionContext, interaction entitytypes.InteractionRequest, toolName ToolName) {
	s.insertInteractionEvent(ctx, session, floweventdomain.EventTypeInteractionRequestCreated, interactionRequestEventPayload{
		InteractionID: interaction.ID,
		ToolName:      string(toolName),
		RunID:         interaction.RunID,
	})
}

func (s *Service) auditInteractionCallbackReceived(ctx context.Context, session SessionContext, interactionID string, adapterEventID string, kind enumtypes.InteractionCallbackKind) {
	s.insertInteractionEvent(ctx, session, floweventdomain.EventTypeInteractionCallbackReceived, interactionCallbackEventPayload{
		InteractionID:  interactionID,
		AdapterEventID: strings.TrimSpace(adapterEventID),
		CallbackKind:   string(kind),
	})
}

func (s *Service) auditInteractionResponseAccepted(ctx context.Context, session SessionContext, interactionID string, response *entitytypes.InteractionResponseRecord) {
	payload := interactionResponseEventPayload{InteractionID: interactionID}
	if response != nil {
		payload.ResponseRecordID = response.ID
		payload.ResponseKind = string(response.ResponseKind)
	}
	s.insertInteractionEvent(ctx, session, floweventdomain.EventTypeInteractionResponseAccepted, payload)
}

func (s *Service) auditInteractionResponseRejected(ctx context.Context, session SessionContext, interactionID string, classification enumtypes.InteractionCallbackResultClassification) {
	s.insertInteractionEvent(ctx, session, floweventdomain.EventTypeInteractionResponseRejected, interactionResponseEventPayload{
		InteractionID:  interactionID,
		Classification: string(classification),
	})
}

func (s *Service) auditInteractionWaitEntered(ctx context.Context, session SessionContext, interactionID string, deadlineAt string) {
	s.auditInteractionWaitLifecycle(ctx, session, interactionID, floweventdomain.EventTypeInteractionWaitEntered, strings.TrimSpace(deadlineAt), true)
}

func (s *Service) auditInteractionWaitResumed(ctx context.Context, session SessionContext, interactionID string, requestStatus string) {
	s.auditInteractionWaitLifecycle(ctx, session, interactionID, floweventdomain.EventTypeInteractionWaitResumed, strings.TrimSpace(requestStatus), false)
}

func (s *Service) auditInteractionWaitLifecycle(ctx context.Context, session SessionContext, interactionID string, eventType floweventdomain.EventType, value string, deadline bool) {
	payload := interactionWaitEventPayload{
		RunID:         session.RunID,
		InteractionID: interactionID,
	}
	if deadline {
		payload.WaitDeadlineAt = value
	} else {
		payload.RequestStatus = value
	}
	s.insertInteractionEvent(ctx, session, eventType, payload)
}

func (s *Service) insertInteractionEvent(ctx context.Context, session SessionContext, eventType floweventdomain.EventType, payload any) {
	if s.flowEvents == nil {
		return
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		encoded = json.RawMessage(`{"error":"payload_marshal_failed"}`)
	}

	_ = s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: session.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorIDControlPlaneMCP,
		EventType:     eventType,
		Payload:       encoded,
		CreatedAt:     s.now().UTC(),
	})
}
