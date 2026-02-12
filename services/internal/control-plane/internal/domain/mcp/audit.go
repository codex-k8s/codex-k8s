package mcp

import (
	"context"
	"encoding/json"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
)

const payloadMarshalFailedMessage = "payload_marshal_failed"

type runTokenIssuedEventPayload struct {
	RunID       string                  `json:"run_id"`
	ProjectID   string                  `json:"project_id,omitempty"`
	Namespace   string                  `json:"namespace,omitempty"`
	RuntimeMode agentdomain.RuntimeMode `json:"runtime_mode"`
	ExpiresAt   string                  `json:"expires_at"`
}

type promptContextAssembledEventPayload struct {
	RunID           string                  `json:"run_id"`
	ProjectID       string                  `json:"project_id,omitempty"`
	Namespace       string                  `json:"namespace,omitempty"`
	RuntimeMode     agentdomain.RuntimeMode `json:"runtime_mode"`
	RepositoryID    string                  `json:"repository_id,omitempty"`
	RepositoryOwner string                  `json:"repository_owner,omitempty"`
	RepositoryName  string                  `json:"repository_name,omitempty"`
}

type mcpToolEventPayload struct {
	Server      string                  `json:"server"`
	Tool        ToolName                `json:"tool"`
	Category    ToolCategory            `json:"category"`
	Approval    ToolApprovalPolicy      `json:"approval_state"`
	RunID       string                  `json:"run_id"`
	ProjectID   string                  `json:"project_id,omitempty"`
	Namespace   string                  `json:"namespace,omitempty"`
	RuntimeMode agentdomain.RuntimeMode `json:"runtime_mode"`
	Status      ToolExecutionStatus     `json:"status,omitempty"`
	Error       string                  `json:"error,omitempty"`
	Message     string                  `json:"message,omitempty"`
}

type marshalErrorPayload struct {
	Error string `json:"error"`
}

func encodeRunTokenIssuedEventPayload(payload runTokenIssuedEventPayload) json.RawMessage {
	return marshalEventPayload(payload)
}

func encodePromptContextAssembledEventPayload(payload promptContextAssembledEventPayload) json.RawMessage {
	return marshalEventPayload(payload)
}

func encodeMCPToolEventPayload(payload mcpToolEventPayload) json.RawMessage {
	return marshalEventPayload(payload)
}

func marshalEventPayload(payload any) json.RawMessage {
	raw, err := json.Marshal(payload)
	if err == nil {
		return raw
	}
	fallback, fallbackErr := json.Marshal(marshalErrorPayload{Error: payloadMarshalFailedMessage})
	if fallbackErr != nil {
		return json.RawMessage(`{"error":"payload_marshal_failed"}`)
	}
	return fallback
}

func (s *Service) auditPromptContextAssembled(ctx context.Context, runCtx resolvedRunContext) {
	if s.flowEvents == nil {
		return
	}
	_ = s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: runCtx.Session.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorIDControlPlaneMCP,
		EventType:     floweventdomain.EventTypePromptContextAssembled,
		Payload: encodePromptContextAssembledEventPayload(promptContextAssembledEventPayload{
			RunID:           runCtx.Session.RunID,
			ProjectID:       runCtx.Session.ProjectID,
			Namespace:       runCtx.Session.Namespace,
			RuntimeMode:     runCtx.Session.RuntimeMode,
			RepositoryID:    runCtx.Repository.ID,
			RepositoryOwner: runCtx.Repository.Owner,
			RepositoryName:  runCtx.Repository.Name,
		}),
		CreatedAt: s.now().UTC(),
	})
}

func (s *Service) auditToolCalled(ctx context.Context, session SessionContext, tool ToolCapability) {
	s.insertToolEvent(ctx, session, tool, floweventdomain.EventTypeMCPToolCalled, ToolExecutionStatusOK, "", "")
}

func (s *Service) auditToolSucceeded(ctx context.Context, session SessionContext, tool ToolCapability) {
	s.insertToolEvent(ctx, session, tool, floweventdomain.EventTypeMCPToolSucceeded, ToolExecutionStatusOK, "", "")
}

func (s *Service) auditToolFailed(ctx context.Context, session SessionContext, tool ToolCapability, err error) {
	s.insertToolEvent(ctx, session, tool, floweventdomain.EventTypeMCPToolFailed, "", errString(err), "")
}

func (s *Service) auditToolApprovalPending(ctx context.Context, session SessionContext, tool ToolCapability, message string) {
	s.insertToolEvent(
		ctx,
		session,
		tool,
		floweventdomain.EventTypeMCPToolApprovalPending,
		ToolExecutionStatusApprovalRequired,
		"",
		message,
	)
}

func (s *Service) insertToolEvent(ctx context.Context, session SessionContext, tool ToolCapability, eventType floweventdomain.EventType, status ToolExecutionStatus, errText string, message string) {
	if s.flowEvents == nil {
		return
	}

	_ = s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: session.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorIDControlPlaneMCP,
		EventType:     eventType,
		Payload: encodeMCPToolEventPayload(mcpToolEventPayload{
			Server:      s.cfg.ServerName,
			Tool:        tool.Name,
			Category:    tool.Category,
			Approval:    tool.Approval,
			RunID:       session.RunID,
			ProjectID:   session.ProjectID,
			Namespace:   session.Namespace,
			RuntimeMode: session.RuntimeMode,
			Status:      status,
			Error:       errText,
			Message:     message,
		}),
		CreatedAt: s.now().UTC(),
	})
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
