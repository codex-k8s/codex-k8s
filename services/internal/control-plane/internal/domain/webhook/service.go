package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"

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
	gitHubSenderTypeUser = "user"

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
	EnsureNeedInputLabel(ctx context.Context, params runstatusdomain.EnsureNeedInputLabelParams) (runstatusdomain.EnsureNeedInputLabelResult, error)
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

		allowed, reason, err := s.isActorAllowedForIssueTrigger(ctx, projectID, envelope.Sender.Login, envelope.Sender.Type)
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
		triggerSource := strings.TrimSpace(trigger.Source)
		if strings.EqualFold(triggerSource, webhookdomain.TriggerSourcePullRequestReview) || strings.EqualFold(triggerSource, triggerSourcePullRequestLabel) {
			prHeadRef := strings.TrimSpace(envelope.PullRequest.Head.Ref)
			if prHeadRef != "" {
				runtimeBuildRef = prHeadRef
			}
		} else if strings.EqualFold(triggerSource, webhookdomain.TriggerSourceIssueLabel) {
			runtimeBuildRef = s.resolveRuntimeBuildRefForIssueTrigger(ctx, payloadProjectID, envelope, runtimeBuildRef, runtimeMode)
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
