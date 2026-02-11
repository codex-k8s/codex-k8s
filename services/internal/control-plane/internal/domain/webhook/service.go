package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	projectrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/project"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectmember"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
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
	triggerLabels       TriggerLabels
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
	triggerLabels TriggerLabels,
) *Service {
	defaults := defaultTriggerLabels()
	if strings.TrimSpace(triggerLabels.RunDev) == "" {
		triggerLabels.RunDev = defaults.RunDev
	}
	if strings.TrimSpace(triggerLabels.RunDevRevise) == "" {
		triggerLabels.RunDevRevise = defaults.RunDevRevise
	}

	return &Service{
		agentRuns:           agentRuns,
		flowEvents:          flowEvents,
		repos:               repos,
		projects:            projects,
		users:               users,
		members:             members,
		learningModeDefault: learningModeDefault,
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

	var envelope githubEnvelope
	if err := json.Unmarshal(cmd.Payload, &envelope); err != nil {
		return IngestResult{}, errs.Validation{Field: "payload", Msg: "must be valid JSON"}
	}

	projectID, repositoryID, servicesYAMLPath, hasBinding, err := s.resolveProjectBinding(ctx, envelope)
	if err != nil {
		return IngestResult{}, fmt.Errorf("resolve project binding: %w", err)
	}

	trigger, hasIssueRunTrigger := s.resolveIssueRunTrigger(cmd.EventType, envelope)
	if strings.EqualFold(strings.TrimSpace(cmd.EventType), "issues") && !hasIssueRunTrigger {
		return s.recordIgnoredWebhook(ctx, cmd, envelope, "issue_event_not_trigger_label", "", hasBinding)
	}
	if hasIssueRunTrigger {
		if !hasBinding || strings.TrimSpace(projectID) == "" {
			return s.recordIgnoredWebhook(ctx, cmd, envelope, "repository_not_bound_for_issue_label", trigger.Kind, hasBinding)
		}

		allowed, reason, err := s.isActorAllowedForIssueTrigger(ctx, projectID, envelope.Sender.Login)
		if err != nil {
			return IngestResult{}, fmt.Errorf("authorize issue label trigger actor: %w", err)
		}
		if !allowed {
			return s.recordIgnoredWebhook(ctx, cmd, envelope, reason, trigger.Kind, hasBinding)
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

	runPayload, err := buildRunPayload(cmd, envelope, payloadProjectID, repositoryID, servicesYAMLPath, hasBinding, learningMode, triggerPtr(trigger, hasIssueRunTrigger))
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

	eventPayload, err := buildEventPayload(cmd, envelope, createResult.Inserted, createResult.RunID, triggerPtr(trigger, hasIssueRunTrigger))
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

func buildRunPayload(
	cmd IngestCommand,
	envelope githubEnvelope,
	projectID string,
	repositoryID string,
	servicesYAMLPath string,
	hasBinding bool,
	learningMode bool,
	trigger *issueRunTrigger,
) (json.RawMessage, error) {
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

	if envelope.Issue.Number > 0 {
		issuePayload := map[string]any{
			"id":       envelope.Issue.ID,
			"number":   envelope.Issue.Number,
			"title":    envelope.Issue.Title,
			"html_url": envelope.Issue.HTMLURL,
			"state":    envelope.Issue.State,
			"user": map[string]any{
				"id":    envelope.Issue.User.ID,
				"login": envelope.Issue.User.Login,
			},
		}
		if envelope.Issue.PullRequest != nil {
			issuePayload["pull_request"] = map[string]any{
				"url":      envelope.Issue.PullRequest.URL,
				"html_url": envelope.Issue.PullRequest.HTMLURL,
			}
		}
		payload["issue"] = issuePayload
	}
	if trigger != nil {
		payload["trigger"] = map[string]any{
			"source": "issue_label",
			"label":  trigger.Label,
			"kind":   trigger.Kind,
		}
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal run payload: %w", err)
	}
	return b, nil
}

func buildEventPayload(
	cmd IngestCommand,
	envelope githubEnvelope,
	inserted bool,
	runID string,
	trigger *issueRunTrigger,
) (json.RawMessage, error) {
	payload := buildBaseFlowEventPayload(cmd, envelope)
	payload["inserted"] = inserted
	payload["run_id"] = runID
	if trigger != nil {
		payload["label"] = trigger.Label
		payload["run_kind"] = trigger.Kind
	}
	if envelope.Issue.Number > 0 {
		payload["issue_number"] = envelope.Issue.Number
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal flow event payload: %w", err)
	}
	return b, nil
}

func buildIgnoredEventPayload(cmd IngestCommand, envelope githubEnvelope, reason string, runKind string, hasBinding bool) (json.RawMessage, error) {
	payload := buildBaseFlowEventPayload(cmd, envelope)
	payload["reason"] = reason
	payload["binding_resolved"] = hasBinding

	if strings.TrimSpace(envelope.Label.Name) != "" {
		payload["label"] = envelope.Label.Name
	}
	if strings.TrimSpace(runKind) != "" {
		payload["run_kind"] = runKind
	}
	if envelope.Issue.Number > 0 {
		payload["issue"] = map[string]any{
			"id":       envelope.Issue.ID,
			"number":   envelope.Issue.Number,
			"title":    envelope.Issue.Title,
			"html_url": envelope.Issue.HTMLURL,
			"state":    envelope.Issue.State,
		}
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal ignored event payload: %w", err)
	}
	return b, nil
}

func buildBaseFlowEventPayload(cmd IngestCommand, envelope githubEnvelope) map[string]any {
	return map[string]any{
		"source":         "github",
		"delivery_id":    cmd.DeliveryID,
		"event_type":     cmd.EventType,
		"action":         envelope.Action,
		"correlation_id": cmd.CorrelationID,
		"sender": map[string]any{
			"id":    envelope.Sender.ID,
			"login": envelope.Sender.Login,
		},
		"repository": map[string]any{
			"id":        envelope.Repository.ID,
			"full_name": envelope.Repository.FullName,
			"name":      envelope.Repository.Name,
		},
	}
}

func (s *Service) recordIgnoredWebhook(
	ctx context.Context,
	cmd IngestCommand,
	envelope githubEnvelope,
	reason string,
	runKind string,
	hasBinding bool,
) (IngestResult, error) {
	payload, err := buildIgnoredEventPayload(cmd, envelope, reason, runKind, hasBinding)
	if err != nil {
		return IngestResult{}, fmt.Errorf("build ignored flow event payload: %w", err)
	}

	if err := s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: cmd.CorrelationID,
		ActorType:     "system",
		ActorID:       "github-webhook",
		EventType:     "webhook.ignored",
		Payload:       payload,
		CreatedAt:     cmd.ReceivedAt,
	}); err != nil {
		return IngestResult{}, fmt.Errorf("insert ignored flow event: %w", err)
	}

	return IngestResult{
		CorrelationID: cmd.CorrelationID,
		Status:        "ignored",
		Duplicate:     false,
	}, nil
}

func (s *Service) resolveIssueRunTrigger(eventType string, envelope githubEnvelope) (issueRunTrigger, bool) {
	if !strings.EqualFold(strings.TrimSpace(eventType), "issues") {
		return issueRunTrigger{}, false
	}
	if !strings.EqualFold(strings.TrimSpace(envelope.Action), "labeled") {
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
			Kind:  "dev",
		}, true
	case strings.EqualFold(label, s.triggerLabels.RunDevRevise):
		return issueRunTrigger{
			Label: label,
			Kind:  "dev_revise",
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
