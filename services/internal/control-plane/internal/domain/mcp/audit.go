package mcp

import (
	"context"
	"encoding/json"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
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

type approvalEventPayload struct {
	RequestID     int64           `json:"request_id"`
	RunID         string          `json:"run_id,omitempty"`
	ProjectID     string          `json:"project_id,omitempty"`
	ToolName      string          `json:"tool_name"`
	Action        string          `json:"action"`
	ApprovalMode  string          `json:"approval_mode"`
	ApprovalState string          `json:"approval_state"`
	RequestedBy   string          `json:"requested_by,omitempty"`
	AppliedBy     string          `json:"applied_by,omitempty"`
	ActorID       string          `json:"actor_id,omitempty"`
	Reason        string          `json:"reason,omitempty"`
	TargetRef     json.RawMessage `json:"target_ref,omitempty"`
	ToolCategory  ToolCategory    `json:"tool_category,omitempty"`
}

type runWaitEventPayload struct {
	RunID                string `json:"run_id"`
	WaitState            string `json:"wait_state,omitempty"`
	TimeoutGuardDisabled bool   `json:"timeout_guard_disabled"`
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

func encodeApprovalEventPayload(payload approvalEventPayload) json.RawMessage {
	return marshalEventPayload(payload)
}

func encodeRunWaitEventPayload(payload runWaitEventPayload) json.RawMessage {
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

func (s *Service) auditApprovalRequested(ctx context.Context, session SessionContext, request entitytypes.MCPActionRequest, tool ToolCapability) {
	s.insertApprovalEvent(ctx, session, floweventdomain.ActorTypeSystem, floweventdomain.ActorIDControlPlaneMCP, floweventdomain.EventTypeApprovalRequested, approvalEventPayload{
		RequestID:     request.ID,
		RunID:         request.RunID,
		ProjectID:     request.ProjectID,
		ToolName:      request.ToolName,
		Action:        request.Action,
		ApprovalMode:  string(request.ApprovalMode),
		ApprovalState: string(request.ApprovalState),
		RequestedBy:   request.RequestedBy,
		TargetRef:     request.TargetRef,
		ToolCategory:  tool.Category,
	})
}

func (s *Service) auditApprovalApproved(ctx context.Context, session SessionContext, request entitytypes.MCPActionRequest, actorID string, reason string) {
	s.auditApprovalStateChange(
		ctx,
		session,
		request,
		floweventdomain.ActorTypeHuman,
		floweventdomain.EventTypeApprovalApproved,
		actorID,
		reason,
	)
}

func (s *Service) auditApprovalDenied(ctx context.Context, session SessionContext, request entitytypes.MCPActionRequest, actorID string, reason string) {
	s.auditApprovalStateChange(
		ctx,
		session,
		request,
		floweventdomain.ActorTypeHuman,
		floweventdomain.EventTypeApprovalDenied,
		actorID,
		reason,
	)
}

func (s *Service) auditApprovalExpired(ctx context.Context, session SessionContext, request entitytypes.MCPActionRequest, actorID string, reason string) {
	s.auditApprovalStateChange(
		ctx,
		session,
		request,
		floweventdomain.ActorTypeHuman,
		floweventdomain.EventTypeApprovalExpired,
		actorID,
		reason,
	)
}

func (s *Service) auditApprovalFailed(ctx context.Context, session SessionContext, request entitytypes.MCPActionRequest, actorID string, reason string) {
	s.auditApprovalStateChange(
		ctx,
		session,
		request,
		floweventdomain.ActorTypeSystem,
		floweventdomain.EventTypeApprovalFailed,
		actorID,
		reason,
	)
}

func (s *Service) auditApprovalApplied(ctx context.Context, session SessionContext, request entitytypes.MCPActionRequest, actorID string) {
	s.auditApprovalStateChange(
		ctx,
		session,
		request,
		floweventdomain.ActorTypeSystem,
		floweventdomain.EventTypeApprovalApplied,
		actorID,
		"",
	)
}

func (s *Service) auditRunWaitPaused(ctx context.Context, session SessionContext, payload runWaitPayload) {
	s.insertRunWaitEvent(ctx, session, floweventdomain.EventTypeRunWaitPaused, runWaitEventPayload(payload))
}

func (s *Service) auditRunWaitResumed(ctx context.Context, session SessionContext, payload runWaitPayload) {
	eventPayload := runWaitEventPayload(payload)
	eventPayload.WaitState = ""
	s.insertRunWaitEvent(ctx, session, floweventdomain.EventTypeRunWaitResumed, eventPayload)
}

func (s *Service) auditApprovalStateChange(
	ctx context.Context,
	session SessionContext,
	request entitytypes.MCPActionRequest,
	actorType floweventdomain.ActorType,
	eventType floweventdomain.EventType,
	actorID string,
	reason string,
) {
	s.insertApprovalEvent(ctx, session, actorType, "", eventType, approvalEventPayload{
		RequestID:     request.ID,
		RunID:         request.RunID,
		ProjectID:     request.ProjectID,
		ToolName:      request.ToolName,
		Action:        request.Action,
		ApprovalMode:  string(request.ApprovalMode),
		ApprovalState: string(request.ApprovalState),
		RequestedBy:   request.RequestedBy,
		AppliedBy:     request.AppliedBy,
		ActorID:       actorID,
		Reason:        reason,
		TargetRef:     request.TargetRef,
	})
}

func (s *Service) insertApprovalEvent(
	ctx context.Context,
	session SessionContext,
	actorType floweventdomain.ActorType,
	actorID floweventdomain.ActorID,
	eventType floweventdomain.EventType,
	payload approvalEventPayload,
) {
	if s.flowEvents == nil {
		return
	}
	_ = s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: session.CorrelationID,
		ActorType:     actorType,
		ActorID:       actorID,
		EventType:     eventType,
		Payload:       encodeApprovalEventPayload(payload),
		CreatedAt:     s.now().UTC(),
	})
}

func (s *Service) insertRunWaitEvent(ctx context.Context, session SessionContext, eventType floweventdomain.EventType, payload runWaitEventPayload) {
	if s.flowEvents == nil {
		return
	}
	_ = s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: session.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorIDControlPlaneMCP,
		EventType:     eventType,
		Payload:       encodeRunWaitEventPayload(payload),
		CreatedAt:     s.now().UTC(),
	})
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
