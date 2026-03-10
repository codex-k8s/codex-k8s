package runtimedeploy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	runtimedeploytaskrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimedeploytask"
)

type taskActionAuditPayload struct {
	RunID            string `json:"run_id"`
	Action           string `json:"action"`
	PreviousStatus   string `json:"previous_status"`
	CurrentStatus    string `json:"current_status"`
	AlreadyTerminal  bool   `json:"already_terminal"`
	RequestedByType  string `json:"requested_by_type"`
	RequestedByID    string `json:"requested_by_id,omitempty"`
	RequestedByEmail string `json:"requested_by_email,omitempty"`
	RequestedByLogin string `json:"requested_by_github_login,omitempty"`
	Reason           string `json:"reason,omitempty"`
}

// RequestTaskAction records one cancel/stop action with idempotent terminal semantics.
func (s *Service) RequestTaskAction(ctx context.Context, params TaskActionParams) (TaskActionResult, error) {
	normalized, err := normalizeTaskActionParams(params)
	if err != nil {
		return TaskActionResult{}, err
	}
	if s.tasks == nil {
		return TaskActionResult{}, errs.Validation{Field: "runtime_deploy", Msg: "task repository is not configured"}
	}

	result, err := s.tasks.RequestAction(ctx, runtimedeploytaskrepo.RequestActionParams{
		RunID:       normalized.RunID,
		Action:      normalized.Action,
		RequestedAt: normalized.RequestedAt.UTC(),
		RequestedBy: requestedByLabel(normalized.Actor),
		Reason:      buildTaskActionReason(normalized),
	})
	if err != nil {
		return TaskActionResult{}, err
	}

	out := TaskActionResult{
		RunID:           normalized.RunID,
		Action:          normalized.Action,
		PreviousStatus:  result.PreviousStatus,
		CurrentStatus:   result.CurrentStatus,
		AlreadyTerminal: result.AlreadyTerminal,
	}
	s.appendTaskActionLogBestEffort(ctx, normalized, out)
	s.insertTaskActionAuditBestEffort(ctx, normalized, out)
	return out, nil
}

func normalizeTaskActionParams(params TaskActionParams) (TaskActionParams, error) {
	params.RunID = strings.TrimSpace(params.RunID)
	if params.RunID == "" {
		return TaskActionParams{}, errs.Validation{Field: "run_id", Msg: "is required"}
	}
	switch params.Action {
	case TaskActionCancel, TaskActionStop:
	default:
		return TaskActionParams{}, errs.Validation{Field: "action", Msg: "must be cancel or stop"}
	}

	params.Actor.UserID = strings.TrimSpace(params.Actor.UserID)
	params.Actor.Email = strings.TrimSpace(params.Actor.Email)
	params.Actor.GitHubLogin = strings.TrimSpace(params.Actor.GitHubLogin)
	if requestedByLabel(params.Actor) == "" {
		return TaskActionParams{}, errs.Validation{Field: "requested_by", Msg: "actor identity is required"}
	}

	params.Reason = strings.TrimSpace(params.Reason)
	if params.RequestedAt.IsZero() {
		params.RequestedAt = time.Now().UTC()
	} else {
		params.RequestedAt = params.RequestedAt.UTC()
	}
	return params, nil
}

func requestedByLabel(actor TaskActionActor) string {
	if actor.Email != "" {
		return actor.Email
	}
	if actor.GitHubLogin != "" {
		return actor.GitHubLogin
	}
	return actor.UserID
}

func buildTaskActionReason(params TaskActionParams) string {
	base := fmt.Sprintf("%s requested by %s", actionLabel(params.Action), requestedByLabel(params.Actor))
	if params.Reason == "" {
		return base
	}
	return base + ": " + params.Reason
}

func actionLabel(action TaskAction) string {
	switch action {
	case TaskActionStop:
		return "stop"
	default:
		return "cancel"
	}
}

func actionLogMessage(result TaskActionResult) string {
	if result.AlreadyTerminal {
		return fmt.Sprintf("Ignored %s request because task is already %s", actionLabel(result.Action), result.CurrentStatus)
	}
	return fmt.Sprintf("Operator requested %s: %s -> %s", actionLabel(result.Action), result.PreviousStatus, result.CurrentStatus)
}

func (s *Service) appendTaskActionLogBestEffort(ctx context.Context, params TaskActionParams, result TaskActionResult) {
	message := actionLogMessage(result)
	if params.Reason != "" && !result.AlreadyTerminal {
		message += " (" + params.Reason + ")"
	}
	if err := s.tasks.AppendLog(ctx, runtimedeploytaskrepo.AppendLogParams{
		RunID:    params.RunID,
		Stage:    "control",
		Level:    "info",
		Message:  message,
		MaxLines: 200,
	}); err != nil {
		s.logger.Warn("append runtime deploy action log failed", "run_id", params.RunID, "action", params.Action, "err", err)
	}
}

func (s *Service) insertTaskActionAuditBestEffort(ctx context.Context, params TaskActionParams, result TaskActionResult) {
	if s.flowEvents == nil || s.runs == nil {
		return
	}
	run, found, err := s.runs.GetByID(ctx, params.RunID)
	if err != nil {
		s.logger.Warn("load run for runtime deploy action audit failed", "run_id", params.RunID, "action", params.Action, "err", err)
		return
	}
	if !found || strings.TrimSpace(run.CorrelationID) == "" {
		return
	}

	rawPayload, err := json.Marshal(taskActionAuditPayload{
		RunID:            params.RunID,
		Action:           string(params.Action),
		PreviousStatus:   string(result.PreviousStatus),
		CurrentStatus:    string(result.CurrentStatus),
		AlreadyTerminal:  result.AlreadyTerminal,
		RequestedByType:  string(floweventdomain.ActorTypeHuman),
		RequestedByID:    params.Actor.UserID,
		RequestedByEmail: params.Actor.Email,
		RequestedByLogin: params.Actor.GitHubLogin,
		Reason:           params.Reason,
	})
	if err != nil {
		s.logger.Warn("marshal runtime deploy action audit payload failed", "run_id", params.RunID, "action", params.Action, "err", err)
		return
	}

	if err := s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: run.CorrelationID,
		ActorType:     floweventdomain.ActorTypeHuman,
		ActorID:       floweventdomain.ActorID(requestedByLabel(params.Actor)),
		EventType:     actionEventType(params.Action),
		Payload:       rawPayload,
		CreatedAt:     params.RequestedAt.UTC(),
	}); err != nil {
		s.logger.Warn("insert runtime deploy action audit failed", "run_id", params.RunID, "action", params.Action, "err", err)
	}
}

func actionEventType(action TaskAction) floweventdomain.EventType {
	if action == TaskActionStop {
		return floweventdomain.EventTypeRuntimeDeployStopRequested
	}
	return floweventdomain.EventTypeRuntimeDeployCancelRequested
}
