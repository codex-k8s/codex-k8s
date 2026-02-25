package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
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
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
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

	triggerSourcePullRequestLabel = "pull_request_label"
)

type pushMainDeployTarget struct {
	BuildRef  string
	TargetEnv string
	Namespace string
}

type runStatusService interface {
	UpsertRunStatusComment(ctx context.Context, params runstatusdomain.UpsertCommentParams) (runstatusdomain.UpsertCommentResult, error)
	CleanupNamespacesByIssue(ctx context.Context, params runstatusdomain.CleanupByIssueParams) (runstatusdomain.CleanupByIssueResult, error)
	CleanupNamespacesByPullRequest(ctx context.Context, params runstatusdomain.CleanupByPullRequestParams) (runstatusdomain.CleanupByIssueResult, error)
	PostTriggerLabelConflictComment(ctx context.Context, params runstatusdomain.TriggerLabelConflictCommentParams) (runstatusdomain.TriggerLabelConflictCommentResult, error)
	PostTriggerWarningComment(ctx context.Context, params runstatusdomain.TriggerWarningCommentParams) (runstatusdomain.TriggerWarningCommentResult, error)
}

type runtimeErrorRecorder interface {
	RecordBestEffort(ctx context.Context, params querytypes.RuntimeErrorRecordParams)
}

type pushMainVersionBumpClient interface {
	GetFile(ctx context.Context, token string, owner string, repo string, filePath string, ref string) ([]byte, bool, error)
	ListChangedFilesBetweenCommits(ctx context.Context, token string, owner string, repo string, beforeSHA string, afterSHA string) ([]string, error)
	CommitFilesOnBranch(ctx context.Context, token string, owner string, repo string, branch string, baseSHA string, message string, files map[string][]byte) (string, error)
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
	runtimeErr runtimeErrorRecorder

	learningModeDefault bool
	triggerLabels       TriggerLabels
	runtimeModePolicy   RuntimeModePolicy
	platformNamespace   string
	githubToken         string
	githubMgmt          pushMainVersionBumpClient
	autoVersionBump     bool
}

// Config wires webhook domain dependencies.
type Config struct {
	LearningModeDefault bool
	TriggerLabels       TriggerLabels
	RuntimeModePolicy   RuntimeModePolicy
	PlatformNamespace   string
	GitHubToken         string
	GitHubMgmt          pushMainVersionBumpClient
	PushMainAutoBump    bool
	RunStatus           runStatusService
	RuntimeErrors       runtimeErrorRecorder
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
		runtimeErr:          cfg.RuntimeErrors,
		learningModeDefault: cfg.LearningModeDefault,
		triggerLabels:       triggerLabels,
		runtimeModePolicy:   cfg.RuntimeModePolicy.withDefaults(),
		platformNamespace:   strings.TrimSpace(cfg.PlatformNamespace),
		githubToken:         strings.TrimSpace(cfg.GitHubToken),
		githubMgmt:          cfg.GitHubMgmt,
		autoVersionBump:     cfg.PushMainAutoBump,
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

	projectID, repositoryID, servicesYAMLPath, repositoryDefaultRef, hasBinding, err := s.resolveProjectBinding(ctx, envelope)
	if err != nil {
		return IngestResult{}, fmt.Errorf("resolve project binding: %w", err)
	}
	if err := s.maybeCleanupRunNamespaces(ctx, cmd, envelope, hasBinding); err != nil {
		return IngestResult{}, fmt.Errorf("cleanup run namespaces on close event: %w", err)
	}

	trigger, hasIssueRunTrigger, conflict, reviewMeta, err := s.resolveIssueRunTrigger(ctx, projectID, cmd.EventType, envelope)
	if err != nil {
		return IngestResult{}, fmt.Errorf("resolve issue run trigger: %w", err)
	}
	effectiveCmd := cmd
	effectiveCmd.CorrelationID = s.resolveCorrelationID(cmd, envelope, trigger, hasIssueRunTrigger)
	if reviewMeta.ReceivedChangesRequested {
		s.recordPullRequestReviewChangesRequestedEvent(ctx, effectiveCmd, envelope)
		if hasIssueRunTrigger {
			s.recordPullRequestReviewStageResolvedEvent(ctx, effectiveCmd, envelope, trigger, reviewMeta)
		} else if strings.TrimSpace(conflict.IgnoreReason) != "" {
			s.recordPullRequestReviewStageAmbiguousEvent(ctx, effectiveCmd, envelope, conflict, reviewMeta)
		}
	}
	pushTarget, hasPushMainDeploy := s.resolvePushMainDeploy(effectiveCmd.EventType, envelope)
	if strings.EqualFold(strings.TrimSpace(effectiveCmd.EventType), string(webhookdomain.GitHubEventIssues)) && !hasIssueRunTrigger {
		return s.recordIgnoredWebhook(ctx, effectiveCmd, envelope, ignoredWebhookParams{
			Reason:     "issue_event_not_trigger_label",
			RunKind:    "",
			HasBinding: hasBinding,
		})
	}
	if hasIssueRunTrigger && len(conflict.ConflictingLabels) > 1 {
		if err := s.postTriggerConflictComment(ctx, effectiveCmd, envelope, trigger, conflict.ConflictingLabels); err != nil {
			return IngestResult{}, fmt.Errorf("post trigger conflict comment: %w", err)
		}
		return s.recordIgnoredWebhook(ctx, effectiveCmd, envelope, ignoredWebhookParams{
			Reason:            "issue_trigger_label_conflict",
			RunKind:           trigger.Kind,
			HasBinding:        hasBinding,
			ConflictingLabels: conflict.ConflictingLabels,
		})
	}
	if !hasIssueRunTrigger && conflict.IgnoreReason != "" {
		return s.recordIgnoredWebhook(ctx, effectiveCmd, envelope, ignoredWebhookParams{
			Reason:            conflict.IgnoreReason,
			RunKind:           "",
			HasBinding:        hasBinding,
			ConflictingLabels: conflict.ConflictingLabels,
			SuggestedLabels:   conflict.SuggestedLabels,
		})
	}
	if hasIssueRunTrigger {
		if !hasBinding || strings.TrimSpace(projectID) == "" {
			return s.recordIgnoredWebhook(ctx, effectiveCmd, envelope, ignoredWebhookParams{
				Reason:     string(runstatusdomain.TriggerWarningReasonRepositoryNotBoundForIssueLabel),
				RunKind:    trigger.Kind,
				HasBinding: hasBinding,
			})
		}

		allowed, reason, err := s.isActorAllowedForIssueTrigger(ctx, projectID, envelope.Sender.Login)
		if err != nil {
			return IngestResult{}, fmt.Errorf("authorize issue label trigger actor: %w", err)
		}
		if !allowed {
			return s.recordIgnoredWebhook(ctx, effectiveCmd, envelope, ignoredWebhookParams{
				Reason:     reason,
				RunKind:    trigger.Kind,
				HasBinding: hasBinding,
			})
		}
	}
	if hasPushMainDeploy {
		if !hasBinding || strings.TrimSpace(projectID) == "" {
			return s.recordIgnoredWebhook(ctx, effectiveCmd, envelope, ignoredWebhookParams{
				Reason:     "repository_not_bound_for_push_main",
				RunKind:    "",
				HasBinding: hasBinding,
			})
		}
		if strings.TrimSpace(servicesYAMLPath) == "" {
			servicesYAMLPath = "services.yaml"
		}
		bumped, err := s.maybeAutoBumpMainVersions(ctx, envelope, servicesYAMLPath, pushTarget.BuildRef)
		if err != nil {
			return IngestResult{}, fmt.Errorf("auto bump services versions for push main: %w", err)
		}
		if bumped {
			return s.recordIgnoredWebhook(ctx, effectiveCmd, envelope, ignoredWebhookParams{
				Reason:     "push_main_versions_autobumped",
				RunKind:    "",
				HasBinding: hasBinding,
			})
		}
	}
	if !hasIssueRunTrigger && !hasPushMainDeploy {
		return s.recordReceivedWebhookWithoutRun(ctx, effectiveCmd, envelope)
	}

	fallbackProjectID := deriveProjectID(effectiveCmd.CorrelationID, envelope)

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
	var profileHints *githubRunProfileHints
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
		if strings.EqualFold(strings.TrimSpace(trigger.Source), webhookdomain.TriggerSourcePullRequestReview) {
			profileHints, err = s.loadProfileHintsFromRunHistory(ctx, payloadProjectID, strings.TrimSpace(envelope.Repository.FullName), reviewMeta.ResolvedIssueNumber, envelope.PullRequest.Number)
			if err != nil {
				return IngestResult{}, fmt.Errorf("resolve profile hints from run history: %w", err)
			}
		}
		runtimeMode, runtimeModeSource = s.resolveRunRuntimeMode(triggerPtr(trigger, hasIssueRunTrigger))
		runtimeTargetEnv = ""
		runtimeNamespace = ""
		runtimeBuildRef = strings.TrimSpace(repositoryDefaultRef)
		runtimeDeployOnly = false
		if trigger.Kind == webhookdomain.TriggerKindAIRepair {
			runtimeTargetEnv = "production"
			runtimeNamespace = strings.TrimSpace(s.platformNamespace)
		}
	}

	runPayload, err := buildRunPayload(runPayloadInput{
		Command:           effectiveCmd,
		Envelope:          envelope,
		ProjectID:         payloadProjectID,
		RepositoryID:      repositoryID,
		ServicesYAMLPath:  servicesYAMLPath,
		HasBinding:        hasBinding,
		LearningMode:      learningMode,
		Trigger:           triggerPtr(trigger, hasIssueRunTrigger),
		Agent:             agent,
		ProfileHints:      profileHints,
		ResolvedIssueNo:   reviewMeta.ResolvedIssueNumber,
		ResolvedIssueURL:  buildGitHubIssueURL(strings.TrimSpace(envelope.Repository.FullName), reviewMeta.ResolvedIssueNumber),
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
		CorrelationID: effectiveCmd.CorrelationID,
		ProjectID:     projectID,
		AgentID:       agent.ID,
		RunPayload:    runPayload,
		LearningMode:  learningMode,
	})
	if err != nil {
		return IngestResult{}, fmt.Errorf("create pending agent run: %w", err)
	}
	if hasIssueRunTrigger && createResult.Inserted {
		s.postRunLaunchPlannedFeedback(ctx, createResult.RunID, trigger, runtimeMode, runtimeNamespace, effectiveCmd.EventType)
	}

	eventPayload, err := buildEventPayload(eventPayloadInput{
		Command:  effectiveCmd,
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
		CorrelationID: effectiveCmd.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       githubWebhookActorID,
		EventType:     eventType,
		Payload:       eventPayload,
		CreatedAt:     effectiveCmd.ReceivedAt,
	}); err != nil {
		return IngestResult{}, fmt.Errorf("insert flow event: %w", err)
	}

	return IngestResult{
		CorrelationID: effectiveCmd.CorrelationID,
		RunID:         createResult.RunID,
		Status:        status,
		Duplicate:     !createResult.Inserted,
	}, nil
}

func (s *Service) recordIgnoredWebhook(ctx context.Context, cmd IngestCommand, envelope githubWebhookEnvelope, params ignoredWebhookParams) (IngestResult, error) {
	s.recordIgnoredWebhookWarning(ctx, cmd, envelope, params)
	s.postIgnoredWebhookDiagnosticComment(ctx, cmd, envelope, params)

	payload, err := buildIgnoredEventPayload(ignoredEventPayloadInput{
		Command:           cmd,
		Envelope:          envelope,
		Reason:            params.Reason,
		RunKind:           params.RunKind,
		HasBinding:        params.HasBinding,
		ConflictingLabels: params.ConflictingLabels,
		SuggestedLabels:   params.SuggestedLabels,
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

type pullRequestReviewFlowEventPayload struct {
	RepositoryFullName string                    `json:"repository_full_name"`
	PullRequestNumber  int64                     `json:"pull_request_number,omitempty"`
	IssueNumber        int64                     `json:"issue_number,omitempty"`
	ResolverSource     string                    `json:"resolver_source,omitempty"`
	TriggerLabel       string                    `json:"trigger_label,omitempty"`
	TriggerKind        webhookdomain.TriggerKind `json:"trigger_kind,omitempty"`
	Reason             string                    `json:"reason,omitempty"`
	ConflictingLabels  []string                  `json:"conflicting_labels,omitempty"`
	SuggestedLabels    []string                  `json:"suggested_labels,omitempty"`
}

func (s *Service) recordPullRequestReviewChangesRequestedEvent(ctx context.Context, cmd IngestCommand, envelope githubWebhookEnvelope) {
	payload, err := json.Marshal(pullRequestReviewFlowEventPayload{
		RepositoryFullName: strings.TrimSpace(envelope.Repository.FullName),
		PullRequestNumber:  envelope.PullRequest.Number,
		IssueNumber:        envelope.Issue.Number,
	})
	if err != nil {
		return
	}
	_ = s.insertSystemFlowEvent(ctx, cmd.CorrelationID, floweventdomain.EventTypeRunReviewChangesRequested, payload, cmd.ReceivedAt)
}

func (s *Service) recordPullRequestReviewStageResolvedEvent(ctx context.Context, cmd IngestCommand, envelope githubWebhookEnvelope, trigger issueRunTrigger, meta pullRequestReviewResolutionMeta) {
	payload, err := json.Marshal(pullRequestReviewFlowEventPayload{
		RepositoryFullName: strings.TrimSpace(envelope.Repository.FullName),
		PullRequestNumber:  envelope.PullRequest.Number,
		IssueNumber:        meta.ResolvedIssueNumber,
		ResolverSource:     strings.TrimSpace(meta.ResolverSource),
		TriggerLabel:       strings.TrimSpace(trigger.Label),
		TriggerKind:        trigger.Kind,
		SuggestedLabels:    normalizeWebhookLabels(meta.SuggestedLabels),
	})
	if err != nil {
		return
	}
	_ = s.insertSystemFlowEvent(ctx, cmd.CorrelationID, floweventdomain.EventTypeRunReviseStageResolved, payload, cmd.ReceivedAt)
}

func (s *Service) recordPullRequestReviewStageAmbiguousEvent(ctx context.Context, cmd IngestCommand, envelope githubWebhookEnvelope, conflict triggerConflictResult, meta pullRequestReviewResolutionMeta) {
	payload, err := json.Marshal(pullRequestReviewFlowEventPayload{
		RepositoryFullName: strings.TrimSpace(envelope.Repository.FullName),
		PullRequestNumber:  envelope.PullRequest.Number,
		IssueNumber:        meta.ResolvedIssueNumber,
		ResolverSource:     strings.TrimSpace(meta.ResolverSource),
		Reason:             strings.TrimSpace(conflict.IgnoreReason),
		ConflictingLabels:  normalizeWebhookLabels(conflict.ConflictingLabels),
		SuggestedLabels:    normalizeWebhookLabels(meta.SuggestedLabels),
	})
	if err != nil {
		return
	}
	_ = s.insertSystemFlowEvent(ctx, cmd.CorrelationID, floweventdomain.EventTypeRunReviseStageAmbiguous, payload, cmd.ReceivedAt)
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

func (s *Service) resolveIssueRunTrigger(ctx context.Context, projectID string, eventType string, envelope githubWebhookEnvelope) (issueRunTrigger, bool, triggerConflictResult, pullRequestReviewResolutionMeta, error) {
	switch strings.TrimSpace(strings.ToLower(eventType)) {
	case string(webhookdomain.GitHubEventIssues):
		if !strings.EqualFold(strings.TrimSpace(envelope.Action), string(webhookdomain.GitHubActionLabeled)) {
			return issueRunTrigger{}, false, triggerConflictResult{}, pullRequestReviewResolutionMeta{}, nil
		}

		label := strings.TrimSpace(envelope.Label.Name)
		if label == "" {
			return issueRunTrigger{}, false, triggerConflictResult{}, pullRequestReviewResolutionMeta{}, nil
		}
		kind, ok := s.triggerLabels.resolveKind(label)
		if !ok {
			return issueRunTrigger{}, false, triggerConflictResult{}, pullRequestReviewResolutionMeta{}, nil
		}
		conflictingLabels := s.triggerLabels.collectIssueTriggerLabels(envelope.Issue.Labels)
		return issueRunTrigger{
				Source: webhookdomain.TriggerSourceIssueLabel,
				Label:  label,
				Kind:   kind,
			}, true, triggerConflictResult{
				ConflictingLabels: conflictingLabels,
			}, pullRequestReviewResolutionMeta{}, nil
	case string(webhookdomain.GitHubEventPullRequest):
		if !strings.EqualFold(strings.TrimSpace(envelope.Action), string(webhookdomain.GitHubActionLabeled)) {
			return issueRunTrigger{}, false, triggerConflictResult{}, pullRequestReviewResolutionMeta{}, nil
		}
		if envelope.PullRequest.Number <= 0 {
			return issueRunTrigger{}, false, triggerConflictResult{}, pullRequestReviewResolutionMeta{}, nil
		}

		label := strings.TrimSpace(envelope.Label.Name)
		if label == "" || !s.triggerLabels.isNeedReviewerLabel(label) {
			return issueRunTrigger{}, false, triggerConflictResult{}, pullRequestReviewResolutionMeta{}, nil
		}
		return issueRunTrigger{
			Source: triggerSourcePullRequestLabel,
			Label:  label,
			Kind:   webhookdomain.TriggerKindDev,
		}, true, triggerConflictResult{}, pullRequestReviewResolutionMeta{}, nil
	case string(webhookdomain.GitHubEventPullRequestReview):
		if !strings.EqualFold(strings.TrimSpace(envelope.Action), string(webhookdomain.GitHubActionSubmitted)) {
			return issueRunTrigger{}, false, triggerConflictResult{}, pullRequestReviewResolutionMeta{}, nil
		}
		if !strings.EqualFold(strings.TrimSpace(envelope.Review.State), webhookdomain.GitHubReviewStateChangesRequested) {
			return issueRunTrigger{}, false, triggerConflictResult{}, pullRequestReviewResolutionMeta{}, nil
		}
		resolution, err := s.resolvePullRequestReviewReviseTrigger(ctx, projectID, envelope)
		if err != nil {
			return issueRunTrigger{}, false, triggerConflictResult{}, pullRequestReviewResolutionMeta{}, err
		}
		meta := pullRequestReviewResolutionMeta{
			ReceivedChangesRequested: true,
			ResolverSource:           resolution.resolverSource,
			ResolvedIssueNumber:      resolution.resolvedIssue,
			SuggestedLabels:          resolution.suggestedLabels,
		}
		if !resolution.ok {
			return issueRunTrigger{}, false, triggerConflictResult{
				ConflictingLabels: resolution.conflicting,
				IgnoreReason:      resolution.ignoreReason,
				SuggestedLabels:   resolution.suggestedLabels,
				ResolverSource:    resolution.resolverSource,
				ResolvedIssue:     resolution.resolvedIssue,
			}, meta, nil
		}
		return resolution.trigger, true, triggerConflictResult{
			ResolverSource: resolution.resolverSource,
			ResolvedIssue:  resolution.resolvedIssue,
		}, meta, nil
	default:
		return issueRunTrigger{}, false, triggerConflictResult{}, pullRequestReviewResolutionMeta{}, nil
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

func (s *Service) resolveCorrelationID(cmd IngestCommand, envelope githubWebhookEnvelope, trigger issueRunTrigger, hasIssueRunTrigger bool) string {
	correlationID := strings.TrimSpace(cmd.CorrelationID)
	if correlationID == "" {
		return correlationID
	}
	if !hasIssueRunTrigger {
		return correlationID
	}
	if !strings.EqualFold(strings.TrimSpace(trigger.Source), triggerSourcePullRequestLabel) {
		return correlationID
	}
	if !s.triggerLabels.isNeedReviewerLabel(trigger.Label) {
		return correlationID
	}

	repositoryFullName := strings.ToLower(strings.TrimSpace(envelope.Repository.FullName))
	if repositoryFullName == "" || envelope.PullRequest.Number <= 0 {
		return correlationID
	}

	action := strings.ToLower(strings.TrimSpace(envelope.Action))
	triggerLabel := strings.ToLower(strings.TrimSpace(trigger.Label))
	updatedAt := normalizeReviewerTriggerTimestamp(envelope.PullRequest.UpdatedAt)
	if action == "" || triggerLabel == "" || updatedAt == "" {
		return correlationID
	}

	signature := fmt.Sprintf("pull_request_label|%s|%d|%s|%s|%s", repositoryFullName, envelope.PullRequest.Number, action, triggerLabel, updatedAt)
	return "pr-label-reviewer-" + uuid.NewSHA1(uuid.NameSpaceURL, []byte(signature)).String()
}

func normalizeReviewerTriggerTimestamp(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339Nano, trimmed)
	if err != nil {
		return trimmed
	}
	return parsed.UTC().Format(time.RFC3339Nano)
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

func hasAnyLabel(labels []githubLabelRecord, candidates ...string) bool {
	for _, label := range candidates {
		if containsLabel(labels, label) {
			return true
		}
	}
	return false
}

func normalizeWebhookLabels(labels []string) []string {
	result := make([]string, 0, len(labels))
	for _, raw := range labels {
		label := strings.TrimSpace(raw)
		if label == "" {
			continue
		}
		if !slices.Contains(result, label) {
			result = append(result, label)
		}
	}
	slices.Sort(result)
	return result
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

func (s *Service) recordIgnoredWebhookWarning(ctx context.Context, cmd IngestCommand, envelope githubWebhookEnvelope, params ignoredWebhookParams) {
	if s.runtimeErr == nil || !isRunCreationWarningReason(params.Reason) {
		return
	}
	details, _ := json.Marshal(map[string]any{
		"reason":              strings.TrimSpace(params.Reason),
		"event_type":          strings.TrimSpace(cmd.EventType),
		"repository_fullname": strings.TrimSpace(envelope.Repository.FullName),
		"issue_number":        envelope.Issue.Number,
		"pull_request_number": envelope.PullRequest.Number,
		"run_kind":            strings.TrimSpace(string(params.RunKind)),
		"conflicting_labels":  params.ConflictingLabels,
		"suggested_labels":    params.SuggestedLabels,
	})
	message := "Run was not created: " + strings.TrimSpace(params.Reason)
	s.runtimeErr.RecordBestEffort(ctx, querytypes.RuntimeErrorRecordParams{
		Source:        "webhook.trigger",
		Level:         "warning",
		Message:       message,
		CorrelationID: strings.TrimSpace(cmd.CorrelationID),
		ProjectID:     "",
		DetailsJSON:   details,
	})
}

func (s *Service) postIgnoredWebhookDiagnosticComment(ctx context.Context, cmd IngestCommand, envelope githubWebhookEnvelope, params ignoredWebhookParams) {
	if s.runStatus == nil || !isRunCreationWarningReason(params.Reason) {
		return
	}
	repositoryFullName := strings.TrimSpace(envelope.Repository.FullName)
	if repositoryFullName == "" {
		return
	}

	threadKind := ""
	threadNumber := 0
	if strings.EqualFold(strings.TrimSpace(cmd.EventType), string(webhookdomain.GitHubEventPullRequestReview)) && envelope.PullRequest.Number > 0 {
		threadKind = "pull_request"
		threadNumber = int(envelope.PullRequest.Number)
	} else if envelope.Issue.Number > 0 {
		threadKind = "issue"
		threadNumber = int(envelope.Issue.Number)
	}
	if threadKind == "" || threadNumber <= 0 {
		return
	}

	_, _ = s.runStatus.PostTriggerWarningComment(ctx, runstatusdomain.TriggerWarningCommentParams{
		CorrelationID:      cmd.CorrelationID,
		RepositoryFullName: repositoryFullName,
		ThreadKind:         threadKind,
		ThreadNumber:       threadNumber,
		Locale:             localeFromEventType(cmd.EventType),
		ReasonCode:         runstatusdomain.TriggerWarningReasonCode(strings.TrimSpace(params.Reason)),
		ConflictingLabels:  params.ConflictingLabels,
		SuggestedLabels:    params.SuggestedLabels,
	})
}

func isRunCreationWarningReason(reason string) bool {
	normalized := strings.TrimSpace(reason)
	if normalized == "" {
		return false
	}
	if normalized == string(runstatusdomain.TriggerWarningReasonPullRequestReviewMissingStageLabel) ||
		normalized == string(runstatusdomain.TriggerWarningReasonPullRequestReviewStageLabelConflict) ||
		normalized == string(runstatusdomain.TriggerWarningReasonPullRequestReviewStageNotResolved) ||
		normalized == string(runstatusdomain.TriggerWarningReasonPullRequestReviewStageAmbiguous) {
		return true
	}
	if normalized == string(runstatusdomain.TriggerWarningReasonRepositoryNotBoundForIssueLabel) {
		return true
	}
	return strings.HasPrefix(normalized, "sender_")
}

func (s *Service) postRunLaunchPlannedFeedback(ctx context.Context, runID string, trigger issueRunTrigger, runtimeMode agentdomain.RuntimeMode, runtimeNamespace string, eventType string) {
	if s.runStatus == nil {
		return
	}
	trimmedRunID := strings.TrimSpace(runID)
	if trimmedRunID == "" {
		return
	}

	_, _ = s.runStatus.UpsertRunStatusComment(ctx, runstatusdomain.UpsertCommentParams{
		RunID:        trimmedRunID,
		Phase:        runstatusdomain.PhaseCreated,
		RuntimeMode:  string(runtimeMode),
		Namespace:    strings.TrimSpace(runtimeNamespace),
		TriggerKind:  string(trigger.Kind),
		PromptLocale: localeFromEventType(eventType),
		RunStatus:    string(rundomain.StatusPending),
	})
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
	if strings.EqualFold(strings.TrimSpace(trigger.Source), triggerSourcePullRequestLabel) {
		return agentKeyReviewer
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
		return agentKeyKM
	case webhookdomain.TriggerKindQA:
		return agentKeyQA
	case webhookdomain.TriggerKindAIRepair, webhookdomain.TriggerKindPostDeploy, webhookdomain.TriggerKindOps:
		return agentKeySRE
	case webhookdomain.TriggerKindSelfImprove:
		return agentKeyKM
	default:
		return defaultRunAgentKey
	}
}

func (s *Service) resolveProjectBinding(ctx context.Context, envelope githubWebhookEnvelope) (projectID string, repositoryID string, servicesYAMLPath string, defaultRef string, ok bool, err error) {
	if s.repos == nil || envelope.Repository.ID == 0 {
		return "", "", "", "", false, nil
	}
	res, ok, err := s.repos.FindByProviderExternalID(ctx, "github", envelope.Repository.ID)
	if err != nil {
		return "", "", "", "", false, err
	}
	if !ok {
		return "", "", "", "", false, nil
	}
	return res.ProjectID, res.RepositoryID, res.ServicesYAMLPath, strings.TrimSpace(res.DefaultRef), true, nil
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
