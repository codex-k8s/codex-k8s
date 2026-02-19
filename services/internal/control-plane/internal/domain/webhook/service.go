package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/google/uuid"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	agentrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agent"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	projectrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/project"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectmember"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
	runstatusdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runstatus"
)

const githubWebhookActorID = floweventdomain.ActorIDGitHubWebhook

const (
	agentKeyPM = "pm" // Product manager: defines and refines product artifacts.
	agentKeySA = "sa" // Solution architect: drives architecture decisions and constraints.

	agentKeyEM = "em" // Engineering manager: coordinates delivery process and gates.

	defaultRunAgentKey = "dev"      // Developer: implements code and documentation changes.
	agentKeyReviewer   = "reviewer" // Reviewer: performs preliminary technical review.
	agentKeyQA         = "qa"       // QA: validates quality, test scenarios, and regressions.
	agentKeySRE        = "sre"      // SRE/OPS: handles operations, stability, and runtime diagnostics.
	agentKeyKM         = "km"       // Knowledge manager: maintains traceability and self-improve loop.
)

type pushMainDeployTarget struct {
	BuildRef  string
	TargetEnv string
	Namespace string
}

type runStatusService interface {
	CleanupNamespacesByIssue(ctx context.Context, params runstatusdomain.CleanupByIssueParams) (runstatusdomain.CleanupByIssueResult, error)
	CleanupNamespacesByPullRequest(ctx context.Context, params runstatusdomain.CleanupByPullRequestParams) (runstatusdomain.CleanupByIssueResult, error)
	PostTriggerLabelConflictComment(ctx context.Context, params runstatusdomain.TriggerLabelConflictCommentParams) (runstatusdomain.TriggerLabelConflictCommentResult, error)
}

// Service ingests provider webhooks into idempotent run and flow-event records.
type Service struct {
	agentRuns  agentrunrepo.Repository
	agents     agentrepo.Repository
	flowEvents floweventrepo.Repository
	repos      repocfgrepo.Repository
	projects   projectrepo.Repository
	users      userrepo.Repository
	members    projectmemberrepo.Repository
	runStatus  runStatusService

	learningModeDefault bool
	triggerLabels       TriggerLabels
	runtimeModePolicy   RuntimeModePolicy
	platformNamespace   string
}

// Config wires webhook domain dependencies.
type Config struct {
	LearningModeDefault bool
	TriggerLabels       TriggerLabels
	RuntimeModePolicy   RuntimeModePolicy
	PlatformNamespace   string
	RunStatus           runStatusService
	Members             projectmemberrepo.Repository
	Users               userrepo.Repository
	Projects            projectrepo.Repository
	Repos               repocfgrepo.Repository
	FlowEvents          floweventrepo.Repository
	Agents              agentrepo.Repository
	AgentRuns           agentrunrepo.Repository
}

// NewService wires webhook domain dependencies.
func NewService(cfg Config) *Service {
	triggerLabels := cfg.TriggerLabels.withDefaults()

	return &Service{
		agentRuns:           cfg.AgentRuns,
		agents:              cfg.Agents,
		flowEvents:          cfg.FlowEvents,
		repos:               cfg.Repos,
		projects:            cfg.Projects,
		users:               cfg.Users,
		members:             cfg.Members,
		runStatus:           cfg.RunStatus,
		learningModeDefault: cfg.LearningModeDefault,
		triggerLabels:       triggerLabels,
		runtimeModePolicy:   cfg.RuntimeModePolicy.withDefaults(),
		platformNamespace:   strings.TrimSpace(cfg.PlatformNamespace),
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
	if err := s.maybeCleanupRunNamespaces(ctx, cmd, envelope, hasBinding); err != nil {
		return IngestResult{}, fmt.Errorf("cleanup run namespaces on close event: %w", err)
	}

	trigger, hasIssueRunTrigger, conflict := s.resolveIssueRunTrigger(cmd.EventType, envelope)
	pushTarget, hasPushMainDeploy := s.resolvePushMainDeploy(cmd.EventType, envelope)
	if strings.EqualFold(strings.TrimSpace(cmd.EventType), string(webhookdomain.GitHubEventIssues)) && !hasIssueRunTrigger {
		return s.recordIgnoredWebhook(ctx, cmd, envelope, ignoredWebhookParams{
			Reason:     "issue_event_not_trigger_label",
			RunKind:    "",
			HasBinding: hasBinding,
		})
	}
	if hasIssueRunTrigger && len(conflict.ConflictingLabels) > 1 {
		if err := s.postTriggerConflictComment(ctx, cmd, envelope, trigger, conflict.ConflictingLabels); err != nil {
			return IngestResult{}, fmt.Errorf("post trigger conflict comment: %w", err)
		}
		return s.recordIgnoredWebhook(ctx, cmd, envelope, ignoredWebhookParams{
			Reason:            "issue_trigger_label_conflict",
			RunKind:           trigger.Kind,
			HasBinding:        hasBinding,
			ConflictingLabels: conflict.ConflictingLabels,
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
	if hasPushMainDeploy {
		if !hasBinding || strings.TrimSpace(projectID) == "" {
			return s.recordIgnoredWebhook(ctx, cmd, envelope, ignoredWebhookParams{
				Reason:     "repository_not_bound_for_push_main",
				RunKind:    "",
				HasBinding: hasBinding,
			})
		}
	}
	if !hasIssueRunTrigger && !hasPushMainDeploy {
		return s.recordReceivedWebhookWithoutRun(ctx, cmd, envelope)
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

	learningMode := false
	agent := runAgentProfile{}
	runtimeMode := agentdomain.RuntimeModeFullEnv
	runtimeModeSource := runtimeModeSourcePushMain
	runtimeTargetEnv := pushTarget.TargetEnv
	runtimeNamespace := pushTarget.Namespace
	runtimeBuildRef := pushTarget.BuildRef
	runtimeDeployOnly := true

	if hasIssueRunTrigger {
		learningMode, err = s.resolveLearningMode(ctx, learningProjectID, envelope.Sender.Login)
		if err != nil {
			return IngestResult{}, fmt.Errorf("resolve learning mode: %w", err)
		}
		agent, err = s.resolveRunAgent(ctx, payloadProjectID, triggerPtr(trigger, hasIssueRunTrigger))
		if err != nil {
			return IngestResult{}, fmt.Errorf("resolve run agent: %w", err)
		}
		runtimeMode, runtimeModeSource = s.resolveRunRuntimeMode(triggerPtr(trigger, hasIssueRunTrigger))
		runtimeTargetEnv = ""
		runtimeNamespace = ""
		runtimeBuildRef = ""
		runtimeDeployOnly = false
	}

	runPayload, err := buildRunPayload(runPayloadInput{
		Command:           cmd,
		Envelope:          envelope,
		ProjectID:         payloadProjectID,
		RepositoryID:      repositoryID,
		ServicesYAMLPath:  servicesYAMLPath,
		HasBinding:        hasBinding,
		LearningMode:      learningMode,
		Trigger:           triggerPtr(trigger, hasIssueRunTrigger),
		Agent:             agent,
		RuntimeMode:       runtimeMode,
		RuntimeSource:     runtimeModeSource,
		RuntimeTargetEnv:  runtimeTargetEnv,
		RuntimeNamespace:  runtimeNamespace,
		RuntimeBuildRef:   runtimeBuildRef,
		RuntimeDeployOnly: runtimeDeployOnly,
	})
	if err != nil {
		return IngestResult{}, fmt.Errorf("build run payload: %w", err)
	}

	createResult, err := s.agentRuns.CreatePendingIfAbsent(ctx, agentrunrepo.CreateParams{
		CorrelationID: cmd.CorrelationID,
		ProjectID:     projectID,
		AgentID:       agent.ID,
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
		Command:           cmd,
		Envelope:          envelope,
		Reason:            params.Reason,
		RunKind:           params.RunKind,
		HasBinding:        params.HasBinding,
		ConflictingLabels: params.ConflictingLabels,
	})
	if err != nil {
		return IngestResult{}, fmt.Errorf("build ignored flow event payload: %w", err)
	}

	if err := s.insertSystemFlowEvent(ctx, cmd.CorrelationID, floweventdomain.EventTypeWebhookIgnored, payload, cmd.ReceivedAt); err != nil {
		return IngestResult{}, fmt.Errorf("insert ignored flow event: %w", err)
	}

	return IngestResult{
		CorrelationID: cmd.CorrelationID,
		Status:        webhookdomain.IngestStatusIgnored,
		Duplicate:     false,
	}, nil
}

func (s *Service) recordReceivedWebhookWithoutRun(ctx context.Context, cmd IngestCommand, envelope githubWebhookEnvelope) (IngestResult, error) {
	payload, err := buildReceivedEventPayload(cmd, envelope)
	if err != nil {
		return IngestResult{}, fmt.Errorf("build flow event payload: %w", err)
	}

	if err := s.insertSystemFlowEvent(ctx, cmd.CorrelationID, floweventdomain.EventTypeWebhookReceived, payload, cmd.ReceivedAt); err != nil {
		return IngestResult{}, fmt.Errorf("insert flow event: %w", err)
	}

	return IngestResult{
		CorrelationID: cmd.CorrelationID,
		Status:        webhookdomain.IngestStatusAccepted,
		Duplicate:     false,
	}, nil
}

func (s *Service) insertSystemFlowEvent(ctx context.Context, correlationID string, eventType floweventdomain.EventType, payload json.RawMessage, createdAt time.Time) error {
	return s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: correlationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       githubWebhookActorID,
		EventType:     eventType,
		Payload:       payload,
		CreatedAt:     createdAt,
	})
}

func (s *Service) maybeCleanupRunNamespaces(ctx context.Context, cmd IngestCommand, envelope githubWebhookEnvelope, hasBinding bool) error {
	if s.runStatus == nil || !hasBinding {
		return nil
	}

	repositoryFullName := strings.TrimSpace(envelope.Repository.FullName)
	if repositoryFullName == "" {
		return nil
	}

	eventType := strings.ToLower(strings.TrimSpace(cmd.EventType))
	action := strings.ToLower(strings.TrimSpace(envelope.Action))
	requestedByID := strings.TrimSpace(envelope.Sender.Login)
	if requestedByID == "" {
		requestedByID = string(githubWebhookActorID)
	}

	var cleanupErr error
	switch {
	case eventType == string(webhookdomain.GitHubEventIssues) && action == "closed" && envelope.Issue.Number > 0:
		_, cleanupErr = s.runStatus.CleanupNamespacesByIssue(ctx, runstatusdomain.CleanupByIssueParams{
			RepositoryFullName: repositoryFullName,
			IssueNumber:        envelope.Issue.Number,
			RequestedByID:      requestedByID,
		})
	case eventType == string(webhookdomain.GitHubEventPullRequest) && action == "closed" && envelope.PullRequest.Number > 0:
		_, cleanupErr = s.runStatus.CleanupNamespacesByPullRequest(ctx, runstatusdomain.CleanupByPullRequestParams{
			RepositoryFullName: repositoryFullName,
			PRNumber:           envelope.PullRequest.Number,
			RequestedByID:      requestedByID,
		})
	}
	if cleanupErr != nil {
		return cleanupErr
	}

	return nil
}

func (s *Service) resolveIssueRunTrigger(eventType string, envelope githubWebhookEnvelope) (issueRunTrigger, bool, triggerConflictResult) {
	switch strings.TrimSpace(strings.ToLower(eventType)) {
	case string(webhookdomain.GitHubEventIssues):
		if !strings.EqualFold(strings.TrimSpace(envelope.Action), string(webhookdomain.GitHubActionLabeled)) {
			return issueRunTrigger{}, false, triggerConflictResult{}
		}

		label := strings.TrimSpace(envelope.Label.Name)
		if label == "" {
			return issueRunTrigger{}, false, triggerConflictResult{}
		}
		kind, ok := s.triggerLabels.resolveKind(label)
		if !ok {
			return issueRunTrigger{}, false, triggerConflictResult{}
		}
		conflictingLabels := s.triggerLabels.collectIssueTriggerLabels(envelope.Issue.Labels)
		return issueRunTrigger{
				Source: webhookdomain.TriggerSourceIssueLabel,
				Label:  label,
				Kind:   kind,
			}, true, triggerConflictResult{
				ConflictingLabels: conflictingLabels,
			}
	case string(webhookdomain.GitHubEventPullRequestReview):
		if !strings.EqualFold(strings.TrimSpace(envelope.Action), string(webhookdomain.GitHubActionSubmitted)) {
			return issueRunTrigger{}, false, triggerConflictResult{}
		}
		if !strings.EqualFold(strings.TrimSpace(envelope.Review.State), webhookdomain.GitHubReviewStateChangesRequested) {
			return issueRunTrigger{}, false, triggerConflictResult{}
		}
		hasDevLabel := containsLabel(envelope.PullRequest.Labels, s.triggerLabels.RunDev)
		hasDevReviseLabel := containsLabel(envelope.PullRequest.Labels, s.triggerLabels.RunDevRevise)
		if !hasDevLabel && !hasDevReviseLabel {
			return issueRunTrigger{}, false, triggerConflictResult{}
		}
		if hasDevLabel && hasDevReviseLabel {
			return issueRunTrigger{}, false, triggerConflictResult{}
		}
		return issueRunTrigger{
			Source: webhookdomain.TriggerSourcePullRequestReview,
			Label:  s.triggerLabels.RunDevRevise,
			Kind:   webhookdomain.TriggerKindDevRevise,
		}, true, triggerConflictResult{}
	default:
		return issueRunTrigger{}, false, triggerConflictResult{}
	}
}

func (s *Service) resolvePushMainDeploy(eventType string, envelope githubWebhookEnvelope) (pushMainDeployTarget, bool) {
	if !strings.EqualFold(strings.TrimSpace(eventType), string(webhookdomain.GitHubEventPush)) {
		return pushMainDeployTarget{}, false
	}
	if !isMainBranchRef(envelope.Ref) {
		return pushMainDeployTarget{}, false
	}
	if envelope.Deleted || isDeletedGitCommitSHA(envelope.After) {
		return pushMainDeployTarget{}, false
	}

	buildRef := strings.TrimSpace(envelope.After)
	if buildRef == "" {
		return pushMainDeployTarget{}, false
	}

	target := pushMainDeployTarget{
		BuildRef:  buildRef,
		TargetEnv: "production",
		Namespace: s.platformNamespace,
	}
	return target, true
}

func isMainBranchRef(ref string) bool {
	normalized := strings.ToLower(strings.TrimSpace(ref))
	return normalized == "refs/heads/main" || normalized == "refs/heads/master"
}

func isDeletedGitCommitSHA(sha string) bool {
	normalized := strings.ToLower(strings.TrimSpace(sha))
	if normalized == "" {
		return false
	}
	for _, ch := range normalized {
		if ch != '0' {
			return false
		}
	}
	return true
}

func containsLabel(labels []githubLabelRecord, expected string) bool {
	target := normalizeLabelToken(expected)
	if target == "" {
		return false
	}
	for _, item := range labels {
		if normalizeLabelToken(item.Name) == target {
			return true
		}
	}
	return false
}

func (s *Service) postTriggerConflictComment(ctx context.Context, cmd IngestCommand, envelope githubWebhookEnvelope, trigger issueRunTrigger, conflictingLabels []string) error {
	if s.runStatus == nil || envelope.Issue.Number <= 0 {
		return nil
	}
	repositoryFullName := strings.TrimSpace(envelope.Repository.FullName)
	if repositoryFullName == "" {
		return nil
	}
	_, err := s.runStatus.PostTriggerLabelConflictComment(ctx, runstatusdomain.TriggerLabelConflictCommentParams{
		CorrelationID:      cmd.CorrelationID,
		RepositoryFullName: repositoryFullName,
		IssueNumber:        int(envelope.Issue.Number),
		Locale:             localeFromEventType(cmd.EventType),
		TriggerLabel:       strings.TrimSpace(trigger.Label),
		ConflictingLabels:  conflictingLabels,
	})
	return err
}

func localeFromEventType(_ string) string {
	return "ru"
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

func (s *Service) resolveRunAgent(ctx context.Context, projectID string, trigger *issueRunTrigger) (runAgentProfile, error) {
	agentKey := resolveRunAgentKey(trigger)
	if s.agents == nil {
		return runAgentProfile{}, fmt.Errorf("agent repository is not configured")
	}

	agent, ok, err := s.agents.FindEffectiveByKey(ctx, projectID, agentKey)
	if err != nil {
		return runAgentProfile{}, err
	}
	if !ok {
		return runAgentProfile{}, fmt.Errorf("failed_precondition: no active agent configured for key %q", agentKey)
	}

	return runAgentProfile{
		ID:   agent.ID,
		Key:  agent.AgentKey,
		Name: agent.Name,
	}, nil
}

func resolveRunAgentKey(trigger *issueRunTrigger) string {
	if trigger == nil {
		return defaultRunAgentKey
	}
	switch trigger.Kind {
	case webhookdomain.TriggerKindDev, webhookdomain.TriggerKindDevRevise:
		return defaultRunAgentKey
	case webhookdomain.TriggerKindIntake,
		webhookdomain.TriggerKindIntakeRevise,
		webhookdomain.TriggerKindVision,
		webhookdomain.TriggerKindVisionRevise,
		webhookdomain.TriggerKindPRD,
		webhookdomain.TriggerKindPRDRevise:
		return agentKeyPM
	case webhookdomain.TriggerKindArch,
		webhookdomain.TriggerKindArchRevise,
		webhookdomain.TriggerKindDesign,
		webhookdomain.TriggerKindDesignRevise:
		return agentKeySA
	case webhookdomain.TriggerKindPlan,
		webhookdomain.TriggerKindPlanRevise,
		webhookdomain.TriggerKindRelease,
		webhookdomain.TriggerKindRethink:
		return agentKeyEM
	case webhookdomain.TriggerKindDocAudit:
		return agentKeyReviewer
	case webhookdomain.TriggerKindQA:
		return agentKeyQA
	case webhookdomain.TriggerKindPostDeploy, webhookdomain.TriggerKindOps:
		return agentKeySRE
	case webhookdomain.TriggerKindSelfImprove:
		return agentKeyKM
	default:
		return defaultRunAgentKey
	}
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

func (s *Service) resolveRunRuntimeMode(trigger *issueRunTrigger) (agentdomain.RuntimeMode, string) {
	return s.runtimeModePolicy.resolve(trigger)
}

func deriveProjectID(correlationID string, envelope githubWebhookEnvelope) string {
	fullName := strings.TrimSpace(envelope.Repository.FullName)
	if fullName != "" {
		return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("repo:"+strings.ToLower(fullName))).String()
	}
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte("correlation:"+correlationID)).String()
}
