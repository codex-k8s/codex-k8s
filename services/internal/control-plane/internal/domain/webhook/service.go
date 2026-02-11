package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/google/uuid"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	projectrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/project"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectmember"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
)

const githubWebhookActorID = floweventdomain.ActorIDGitHubWebhook

// Service ingests provider webhooks into idempotent run and flow-event records.
type Service struct {
	agentRuns  agentrunrepo.Repository
	flowEvents floweventrepo.Repository
	repos      repocfgrepo.Repository
	projects   projectrepo.Repository
	users      userrepo.Repository
	members    projectmemberrepo.Repository

	learningModeDefault bool
	triggerLabels       TriggerLabels
}

// Config wires webhook domain dependencies.
type Config struct {
	AgentRuns           agentrunrepo.Repository
	FlowEvents          floweventrepo.Repository
	Repos               repocfgrepo.Repository
	Projects            projectrepo.Repository
	Users               userrepo.Repository
	Members             projectmemberrepo.Repository
	LearningModeDefault bool
	TriggerLabels       TriggerLabels
}

// NewService wires webhook domain dependencies.
func NewService(cfg Config) *Service {
	defaults := defaultTriggerLabels()
	triggerLabels := cfg.TriggerLabels
	if strings.TrimSpace(triggerLabels.RunDev) == "" {
		triggerLabels.RunDev = defaults.RunDev
	}
	if strings.TrimSpace(triggerLabels.RunDevRevise) == "" {
		triggerLabels.RunDevRevise = defaults.RunDevRevise
	}

	return &Service{
		agentRuns:           cfg.AgentRuns,
		flowEvents:          cfg.FlowEvents,
		repos:               cfg.Repos,
		projects:            cfg.Projects,
		users:               cfg.Users,
		members:             cfg.Members,
		learningModeDefault: cfg.LearningModeDefault,
		triggerLabels:       triggerLabels,
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

	var envelope githubWebhookEnvelope
	if err := json.Unmarshal(cmd.Payload, &envelope); err != nil {
		return IngestResult{}, errs.Validation{Field: "payload", Msg: "must be valid JSON"}
	}

	projectID, repositoryID, servicesYAMLPath, hasBinding, err := s.resolveProjectBinding(ctx, envelope)
	if err != nil {
		return IngestResult{}, fmt.Errorf("resolve project binding: %w", err)
	}

	trigger, hasIssueRunTrigger := s.resolveIssueRunTrigger(cmd.EventType, envelope)
	if strings.EqualFold(strings.TrimSpace(cmd.EventType), string(webhookdomain.GitHubEventIssues)) && !hasIssueRunTrigger {
		return s.recordIgnoredWebhook(ctx, cmd, envelope, ignoredWebhookParams{
			Reason:     "issue_event_not_trigger_label",
			RunKind:    "",
			HasBinding: hasBinding,
		})
	}
	if hasIssueRunTrigger {
		if !hasBinding || strings.TrimSpace(projectID) == "" {
			return s.recordIgnoredWebhook(ctx, cmd, envelope, ignoredWebhookParams{
				Reason:     "repository_not_bound_for_issue_label",
				RunKind:    trigger.Kind,
				HasBinding: hasBinding,
			})
		}

		allowed, reason, err := s.isActorAllowedForIssueTrigger(ctx, projectID, envelope.Sender.Login)
		if err != nil {
			return IngestResult{}, fmt.Errorf("authorize issue label trigger actor: %w", err)
		}
		if !allowed {
			return s.recordIgnoredWebhook(ctx, cmd, envelope, ignoredWebhookParams{
				Reason:     reason,
				RunKind:    trigger.Kind,
				HasBinding: hasBinding,
			})
		}
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

	runPayload, err := buildRunPayload(runPayloadInput{
		Command:          cmd,
		Envelope:         envelope,
		ProjectID:        payloadProjectID,
		RepositoryID:     repositoryID,
		ServicesYAMLPath: servicesYAMLPath,
		HasBinding:       hasBinding,
		LearningMode:     learningMode,
		Trigger:          triggerPtr(trigger, hasIssueRunTrigger),
	})
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

	eventPayload, err := buildEventPayload(eventPayloadInput{
		Command:  cmd,
		Envelope: envelope,
		Inserted: createResult.Inserted,
		RunID:    createResult.RunID,
		Trigger:  triggerPtr(trigger, hasIssueRunTrigger),
	})
	if err != nil {
		return IngestResult{}, fmt.Errorf("build event payload: %w", err)
	}

	eventType := floweventdomain.EventTypeWebhookReceived
	status := webhookdomain.IngestStatusAccepted
	if !createResult.Inserted {
		eventType = floweventdomain.EventTypeWebhookDuplicate
		status = webhookdomain.IngestStatusDuplicate
	}

	if err := s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: cmd.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       githubWebhookActorID,
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

func (s *Service) recordIgnoredWebhook(ctx context.Context, cmd IngestCommand, envelope githubWebhookEnvelope, params ignoredWebhookParams) (IngestResult, error) {
	payload, err := buildIgnoredEventPayload(ignoredEventPayloadInput{
		Command:    cmd,
		Envelope:   envelope,
		Reason:     params.Reason,
		RunKind:    params.RunKind,
		HasBinding: params.HasBinding,
	})
	if err != nil {
		return IngestResult{}, fmt.Errorf("build ignored flow event payload: %w", err)
	}

	if err := s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: cmd.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       githubWebhookActorID,
		EventType:     floweventdomain.EventTypeWebhookIgnored,
		Payload:       payload,
		CreatedAt:     cmd.ReceivedAt,
	}); err != nil {
		return IngestResult{}, fmt.Errorf("insert ignored flow event: %w", err)
	}

	return IngestResult{
		CorrelationID: cmd.CorrelationID,
		Status:        webhookdomain.IngestStatusIgnored,
		Duplicate:     false,
	}, nil
}

func (s *Service) resolveIssueRunTrigger(eventType string, envelope githubWebhookEnvelope) (issueRunTrigger, bool) {
	if !strings.EqualFold(strings.TrimSpace(eventType), string(webhookdomain.GitHubEventIssues)) {
		return issueRunTrigger{}, false
	}
	if !strings.EqualFold(strings.TrimSpace(envelope.Action), string(webhookdomain.GitHubActionLabeled)) {
		return issueRunTrigger{}, false
	}

	label := strings.TrimSpace(envelope.Label.Name)
	if label == "" {
		return issueRunTrigger{}, false
	}
	switch {
	case strings.EqualFold(label, s.triggerLabels.RunDev):
		return issueRunTrigger{
			Label: label,
			Kind:  webhookdomain.TriggerKindDev,
		}, true
	case strings.EqualFold(label, s.triggerLabels.RunDevRevise):
		return issueRunTrigger{
			Label: label,
			Kind:  webhookdomain.TriggerKindDevRevise,
		}, true
	default:
		return issueRunTrigger{}, false
	}
}

func (s *Service) isActorAllowedForIssueTrigger(ctx context.Context, projectID string, senderLogin string) (bool, string, error) {
	login := strings.TrimSpace(senderLogin)
	if login == "" {
		return false, "sender_login_missing", nil
	}
	if s.users == nil {
		return false, "users_repository_unavailable", nil
	}

	u, ok, err := s.users.GetByGitHubLogin(ctx, login)
	if err != nil {
		return false, "", err
	}
	if !ok || u.ID == "" {
		return false, "sender_not_allowed", nil
	}
	if u.IsPlatformOwner {
		return true, "platform_owner", nil
	}
	if u.IsPlatformAdmin {
		return true, "platform_admin", nil
	}
	if s.members == nil {
		return false, "project_membership_repository_unavailable", nil
	}

	role, ok, err := s.members.GetRole(ctx, projectID, u.ID)
	if err != nil {
		return false, "", err
	}
	if !ok {
		return false, "sender_not_project_member", nil
	}

	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "read_write":
		return true, "project_member_" + strings.ToLower(strings.TrimSpace(role)), nil
	default:
		if strings.TrimSpace(role) == "" {
			return false, "sender_role_not_permitted", nil
		}
		return false, "sender_role_" + strings.ToLower(strings.TrimSpace(role)) + "_not_permitted", nil
	}
}

func triggerPtr(trigger issueRunTrigger, ok bool) *issueRunTrigger {
	if !ok {
		return nil
	}
	t := trigger
	return &t
}

func (s *Service) resolveProjectBinding(ctx context.Context, envelope githubWebhookEnvelope) (projectID string, repositoryID string, servicesYAMLPath string, ok bool, err error) {
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

func deriveProjectID(correlationID string, envelope githubWebhookEnvelope) string {
	fullName := strings.TrimSpace(envelope.Repository.FullName)
	if fullName != "" {
		return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("repo:"+strings.ToLower(fullName))).String()
	}
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("correlation:"+correlationID)).String()
}
