package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/errs"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/flowevent"
	projectrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/project"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/projectmember"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/repocfg"
	userrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/user"
)

// Service ingests provider webhooks into idempotent run and flow-event records.
type Service struct {
	agentRuns  agentrunrepo.Repository
	flowEvents floweventrepo.Repository
	repos      repocfgrepo.Repository
	projects   projectrepo.Repository
	users      userrepo.Repository
	members    projectmemberrepo.Repository

	learningModeDefault bool
}

// NewService wires webhook domain dependencies.
func NewService(
	agentRuns agentrunrepo.Repository,
	flowEvents floweventrepo.Repository,
	repos repocfgrepo.Repository,
	projects projectrepo.Repository,
	users userrepo.Repository,
	members projectmemberrepo.Repository,
	learningModeDefault bool,
) *Service {
	return &Service{
		agentRuns:  agentRuns,
		flowEvents: flowEvents,
		repos:      repos,
		projects:   projects,
		users:      users,
		members:    members,
		learningModeDefault: learningModeDefault,
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

	var envelope githubEnvelope
	if err := json.Unmarshal(cmd.Payload, &envelope); err != nil {
		return IngestResult{}, errs.Validation{Field: "payload", Msg: "must be valid JSON"}
	}

	projectID, repositoryID, servicesYAMLPath, hasBinding, err := s.resolveProjectBinding(ctx, envelope)
	if err != nil {
		return IngestResult{}, fmt.Errorf("resolve project binding: %w", err)
	}

	fallbackProjectID := deriveProjectID(cmd.CorrelationID, envelope)

	learningProjectID := projectID
	if learningProjectID == "" {
		learningProjectID = fallbackProjectID
	}
	payloadProjectID := projectID
	if payloadProjectID == "" {
		payloadProjectID = fallbackProjectID
	}
	if strings.TrimSpace(servicesYAMLPath) == "" {
		servicesYAMLPath = "services.yaml"
	}

	learningMode, err := s.resolveLearningMode(ctx, learningProjectID, envelope.Sender.Login)
	if err != nil {
		return IngestResult{}, fmt.Errorf("resolve learning mode: %w", err)
	}

	runPayload, err := buildRunPayload(cmd, envelope, payloadProjectID, repositoryID, servicesYAMLPath, hasBinding, learningMode)
	if err != nil {
		return IngestResult{}, fmt.Errorf("build run payload: %w", err)
	}

	createResult, err := s.agentRuns.CreatePendingIfAbsent(ctx, agentrunrepo.CreateParams{
		CorrelationID: cmd.CorrelationID,
		ProjectID:     projectID,
		RunPayload:    runPayload,
		LearningMode:  learningMode,
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

func buildRunPayload(cmd IngestCommand, envelope githubEnvelope, projectID string, repositoryID string, servicesYAMLPath string, hasBinding bool, learningMode bool) (json.RawMessage, error) {
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
		"project": map[string]any{
			"id":               projectID,
			"repository_id":    repositoryID,
			"services_yaml":    servicesYAMLPath,
			"binding_resolved": hasBinding,
		},
		"learning_mode": learningMode,
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

func (s *Service) resolveProjectBinding(ctx context.Context, envelope githubEnvelope) (projectID string, repositoryID string, servicesYAMLPath string, ok bool, err error) {
	if s.repos == nil || envelope.Repository.ID == 0 {
		return "", "", "", false, nil
	}
	res, ok, err := s.repos.FindByProviderExternalID(ctx, "github", envelope.Repository.ID)
	if err != nil {
		return "", "", "", false, err
	}
	if !ok {
		return "", "", "", false, nil
	}
	return res.ProjectID, res.RepositoryID, res.ServicesYAMLPath, true, nil
}

func (s *Service) resolveLearningMode(ctx context.Context, projectID string, senderLogin string) (bool, error) {
	if projectID == "" || s.projects == nil {
		return s.learningModeDefault, nil
	}
	projectDefault, ok, err := s.projects.GetLearningModeDefault(ctx, projectID)
	if err != nil {
		return false, err
	}
	if !ok {
		projectDefault = s.learningModeDefault
	}

	// Member override is optional and best-effort: if we can't map sender->user->member,
	// we fall back to project default.
	if s.users == nil || s.members == nil || senderLogin == "" {
		return projectDefault, nil
	}

	u, ok, err := s.users.GetByGitHubLogin(ctx, senderLogin)
	if err != nil {
		return false, err
	}
	if !ok || u.ID == "" {
		return projectDefault, nil
	}

	override, isMember, err := s.members.GetLearningModeOverride(ctx, projectID, u.ID)
	if err != nil {
		return false, err
	}
	if !isMember || override == nil {
		return projectDefault, nil
	}
	return *override, nil
}

func deriveProjectID(correlationID string, envelope githubEnvelope) string {
	fullName := strings.TrimSpace(envelope.Repository.FullName)
	if fullName != "" {
		return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("repo:"+strings.ToLower(fullName))).String()
	}
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("correlation:"+correlationID)).String()
}
