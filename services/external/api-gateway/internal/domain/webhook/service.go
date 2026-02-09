package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/errs"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/flowevent"
)

// Service ingests provider webhooks into idempotent run and flow-event records.
type Service struct {
	agentRuns  agentrunrepo.Repository
	flowEvents floweventrepo.Repository
}

// NewService wires webhook domain dependencies.
func NewService(agentRuns agentrunrepo.Repository, flowEvents floweventrepo.Repository) *Service {
	return &Service{
		agentRuns:  agentRuns,
		flowEvents: flowEvents,
	}
}

// IngestGitHubWebhook validates payload and records idempotent webhook processing state.
func (s *Service) IngestGitHubWebhook(ctx context.Context, cmd IngestCommand) (IngestResult, error) {
	if cmd.CorrelationID == "" {
		return IngestResult{}, errs.Validation{Field: "correlation_id", Msg: "is required"}
	}
	if cmd.DeliveryID == "" {
		return IngestResult{}, errs.Validation{Field: "delivery_id", Msg: "is required"}
	}
	if cmd.EventType == "" {
		return IngestResult{}, errs.Validation{Field: "event_type", Msg: "is required"}
	}
	if len(cmd.Payload) == 0 {
		return IngestResult{}, errs.Validation{Field: "payload", Msg: "is required"}
	}

	if cmd.ReceivedAt.IsZero() {
		cmd.ReceivedAt = time.Now().UTC()
	}

	runPayload, err := buildRunPayload(cmd)
	if err != nil {
		return IngestResult{}, fmt.Errorf("build run payload: %w", err)
	}

	createResult, err := s.agentRuns.CreatePendingIfAbsent(ctx, agentrunrepo.CreateParams{
		CorrelationID: cmd.CorrelationID,
		RunPayload:    runPayload,
	})
	if err != nil {
		return IngestResult{}, fmt.Errorf("create pending agent run: %w", err)
	}

	eventPayload, err := buildEventPayload(cmd, createResult.Inserted, createResult.RunID)
	if err != nil {
		return IngestResult{}, fmt.Errorf("build event payload: %w", err)
	}

	eventType := "webhook.received"
	status := "accepted"
	if !createResult.Inserted {
		eventType = "webhook.duplicate"
		status = "duplicate"
	}

	if err := s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: cmd.CorrelationID,
		ActorType:     "system",
		ActorID:       "github-webhook",
		EventType:     eventType,
		Payload:       eventPayload,
		CreatedAt:     cmd.ReceivedAt,
	}); err != nil {
		return IngestResult{}, fmt.Errorf("insert flow event: %w", err)
	}

	return IngestResult{
		CorrelationID: cmd.CorrelationID,
		RunID:         createResult.RunID,
		Status:        status,
		Duplicate:     !createResult.Inserted,
	}, nil
}

func buildRunPayload(cmd IngestCommand) (json.RawMessage, error) {
	var envelope githubEnvelope
	if err := json.Unmarshal(cmd.Payload, &envelope); err != nil {
		return nil, errs.Validation{Field: "payload", Msg: "must be valid JSON"}
	}

	payload := map[string]any{
		"source":         "github",
		"delivery_id":    cmd.DeliveryID,
		"event_type":     cmd.EventType,
		"received_at":    cmd.ReceivedAt.UTC().Format(time.RFC3339Nano),
		"repository":     map[string]any{"id": envelope.Repository.ID, "full_name": envelope.Repository.FullName, "name": envelope.Repository.Name, "private": envelope.Repository.Private},
		"installation":   map[string]any{"id": envelope.Installation.ID},
		"sender":         map[string]any{"id": envelope.Sender.ID, "login": envelope.Sender.Login},
		"action":         envelope.Action,
		"raw_payload":    json.RawMessage(cmd.Payload),
		"correlation_id": cmd.CorrelationID,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal run payload: %w", err)
	}
	return b, nil
}

func buildEventPayload(cmd IngestCommand, inserted bool, runID string) (json.RawMessage, error) {
	payload := map[string]any{
		"source":         "github",
		"delivery_id":    cmd.DeliveryID,
		"event_type":     cmd.EventType,
		"correlation_id": cmd.CorrelationID,
		"inserted":       inserted,
		"run_id":         runID,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal flow event payload: %w", err)
	}
	return b, nil
}
