package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/flowevent"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/learningfeedback"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

const defaultWorkerID = "worker"
const defaultStateInReviewLabel = "state:in-review"

// Config defines worker run-loop behavior.
type Config struct {
	// WorkerID uniquely identifies current worker instance.
	WorkerID string
	// ClaimLimit limits number of pending runs claimed per tick.
	ClaimLimit int
	// RunningCheckLimit limits running runs reconciled per tick.
	RunningCheckLimit int
	// SlotsPerProject defines slot pool size per project scope.
	SlotsPerProject int
	// SlotLeaseTTL defines maximum duration of slot ownership.
	SlotLeaseTTL time.Duration

	// ProjectLearningModeDefault is applied when the worker auto-creates projects from webhook payloads.
	ProjectLearningModeDefault bool
	// RunNamespacePrefix defines prefix for full-env run namespaces.
	RunNamespacePrefix string
	// CleanupFullEnvNamespace enables namespace cleanup after run completion.
	CleanupFullEnvNamespace bool
	// RunDebugLabel keeps namespace for post-run debugging when present on the issue.
	RunDebugLabel string
	// StateInReviewLabel is applied to PR when run is ready for owner review.
	StateInReviewLabel string
	// ControlPlaneGRPCTarget is control-plane gRPC endpoint used by run jobs for callbacks.
	ControlPlaneGRPCTarget string
	// ControlPlaneMCPBaseURL is MCP endpoint passed to run job environment.
	ControlPlaneMCPBaseURL string
	// OpenAIAPIKey is injected into run pods for codex login.
	OpenAIAPIKey string
	// OpenAIAuthFile is optional Codex auth.json payload injected into run pods.
	OpenAIAuthFile string
	// Context7APIKey enables Context7 documentation calls from run pods when set.
	Context7APIKey string
	// GitBotToken is injected into run pods for git transport only.
	GitBotToken string
	// GitBotUsername is GitHub username used with bot token for git transport auth.
	GitBotUsername string
	// GitBotMail is git author email configured in run pods.
	GitBotMail string
	// AgentDefaultModel is fallback model when run config labels do not override model.
	AgentDefaultModel string
	// AgentDefaultReasoningEffort is fallback reasoning profile when run config labels do not override reasoning.
	AgentDefaultReasoningEffort string
	// AgentDefaultLocale is fallback prompt locale.
	AgentDefaultLocale string
	// AgentBaseBranch is default base branch for PR flow.
	AgentBaseBranch string
	// AIModelGPT53CodexLabel maps GitHub label to gpt-5.3-codex model.
	AIModelGPT53CodexLabel string
	// AIModelGPT53CodexSparkLabel maps GitHub label to gpt-5.3-codex-spark model.
	AIModelGPT53CodexSparkLabel string
	// AIModelGPT52CodexLabel maps GitHub label to gpt-5.2-codex model.
	AIModelGPT52CodexLabel string
	// AIModelGPT52Label maps GitHub label to gpt-5.2 model.
	AIModelGPT52Label string
	// AIModelGPT51CodexMaxLabel maps GitHub label to gpt-5.1-codex-max model.
	AIModelGPT51CodexMaxLabel string
	// AIModelGPT51CodexMiniLabel maps GitHub label to gpt-5.1-codex-mini model.
	AIModelGPT51CodexMiniLabel string
	// AIReasoningLowLabel maps GitHub label to low reasoning profile.
	AIReasoningLowLabel string
	// AIReasoningMediumLabel maps GitHub label to medium reasoning profile.
	AIReasoningMediumLabel string
	// AIReasoningHighLabel maps GitHub label to high reasoning profile.
	AIReasoningHighLabel string
	// AIReasoningExtraHighLabel maps GitHub label to extra-high reasoning profile.
	AIReasoningExtraHighLabel string
}

// Dependencies groups service collaborators to keep constructor signatures compact.
type Dependencies struct {
	// Runs provides queue and lifecycle operations over agent runs.
	Runs runqueuerepo.Repository
	// Events persists flow lifecycle events.
	Events floweventrepo.Repository
	// Feedback persists optional learning-mode explanations.
	Feedback learningfeedbackrepo.Repository
	// Launcher starts and reconciles Kubernetes jobs.
	Launcher Launcher
	// MCPTokenIssuer issues short-lived MCP token for run pods.
	MCPTokenIssuer MCPTokenIssuer
	// RunStatus updates one run-bound issue status comment.
	RunStatus RunStatusNotifier
	// Logger records worker diagnostics.
	Logger *slog.Logger
}

// Service orchestrates pending runs to Kubernetes Jobs and final statuses.
type Service struct {
	cfg       Config
	runs      runqueuerepo.Repository
	events    floweventrepo.Repository
	feedback  learningfeedbackrepo.Repository
	launcher  Launcher
	mcpTokens MCPTokenIssuer
	runStatus RunStatusNotifier
	logger    *slog.Logger
	labels    runAgentLabelCatalog
	now       func() time.Time
}

// NewService creates worker orchestrator instance.
func NewService(cfg Config, deps Dependencies) *Service {
	if cfg.ClaimLimit <= 0 {
		cfg.ClaimLimit = 1
	}
	if cfg.RunningCheckLimit <= 0 {
		cfg.RunningCheckLimit = 100
	}
	if cfg.SlotsPerProject <= 0 {
		cfg.SlotsPerProject = 1
	}
	if cfg.SlotLeaseTTL <= 0 {
		cfg.SlotLeaseTTL = 5 * time.Minute
	}
	if cfg.WorkerID == "" {
		cfg.WorkerID = defaultWorkerID
	}
	if cfg.RunNamespacePrefix == "" {
		cfg.RunNamespacePrefix = defaultRunNamespacePrefix
	}
	cfg.RunDebugLabel = strings.TrimSpace(cfg.RunDebugLabel)
	if cfg.RunDebugLabel == "" {
		cfg.RunDebugLabel = defaultRunDebugLabel
	}
	cfg.StateInReviewLabel = strings.TrimSpace(cfg.StateInReviewLabel)
	if cfg.StateInReviewLabel == "" {
		cfg.StateInReviewLabel = defaultStateInReviewLabel
	}
	cfg.ControlPlaneGRPCTarget = strings.TrimSpace(cfg.ControlPlaneGRPCTarget)
	if cfg.ControlPlaneGRPCTarget == "" {
		cfg.ControlPlaneGRPCTarget = "codex-k8s-control-plane:9090"
	}
	cfg.ControlPlaneMCPBaseURL = resolveControlPlaneMCPBaseURL(cfg.ControlPlaneMCPBaseURL, cfg.ControlPlaneGRPCTarget)
	cfg.OpenAIAPIKey = strings.TrimSpace(cfg.OpenAIAPIKey)
	cfg.OpenAIAuthFile = strings.TrimSpace(cfg.OpenAIAuthFile)
	cfg.Context7APIKey = strings.TrimSpace(cfg.Context7APIKey)
	cfg.GitBotToken = strings.TrimSpace(cfg.GitBotToken)
	cfg.GitBotUsername = strings.TrimSpace(cfg.GitBotUsername)
	if cfg.GitBotUsername == "" {
		cfg.GitBotUsername = "codex-bot"
	}
	cfg.GitBotMail = strings.TrimSpace(cfg.GitBotMail)
	if cfg.GitBotMail == "" {
		cfg.GitBotMail = "codex-bot@codex-k8s.local"
	}
	hasOpenAIAuthFile := cfg.OpenAIAuthFile != ""
	cfg.AgentDefaultModel = strings.TrimSpace(cfg.AgentDefaultModel)
	if cfg.AgentDefaultModel == "" {
		if hasOpenAIAuthFile {
			cfg.AgentDefaultModel = modelGPT53Codex
		} else {
			cfg.AgentDefaultModel = modelGPT52Codex
		}
	}
	if hasOpenAIAuthFile && strings.EqualFold(cfg.AgentDefaultModel, modelGPT52Codex) {
		cfg.AgentDefaultModel = modelGPT53Codex
	}
	if !hasOpenAIAuthFile && isGPT53Model(cfg.AgentDefaultModel) {
		cfg.AgentDefaultModel = modelGPT52Codex
	}
	cfg.AgentDefaultReasoningEffort = strings.TrimSpace(cfg.AgentDefaultReasoningEffort)
	if cfg.AgentDefaultReasoningEffort == "" {
		cfg.AgentDefaultReasoningEffort = "high"
	}
	cfg.AgentDefaultLocale = strings.TrimSpace(cfg.AgentDefaultLocale)
	if cfg.AgentDefaultLocale == "" {
		cfg.AgentDefaultLocale = "ru"
	}
	cfg.AgentBaseBranch = strings.TrimSpace(cfg.AgentBaseBranch)
	if cfg.AgentBaseBranch == "" {
		cfg.AgentBaseBranch = "main"
	}
	labelCatalog := runAgentLabelCatalogFromConfig(cfg)
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	if deps.MCPTokenIssuer == nil {
		deps.MCPTokenIssuer = noopMCPTokenIssuer{}
	}
	if deps.RunStatus == nil {
		deps.RunStatus = noopRunStatusNotifier{}
	}

	return &Service{
		cfg:       cfg,
		runs:      deps.Runs,
		events:    deps.Events,
		feedback:  deps.Feedback,
		launcher:  deps.Launcher,
		mcpTokens: deps.MCPTokenIssuer,
		runStatus: deps.RunStatus,
		logger:    deps.Logger,
		labels:    labelCatalog,
		now:       time.Now,
	}
}

// Tick executes one reconciliation iteration.
func (s *Service) Tick(ctx context.Context) error {
	if err := s.reconcileRunning(ctx); err != nil {
		return fmt.Errorf("reconcile running runs: %w", err)
	}
	if err := s.launchPending(ctx); err != nil {
		return fmt.Errorf("launch pending runs: %w", err)
	}
	return nil
}

// reconcileRunning polls active runs and finalizes those with terminal Kubernetes job states.
func (s *Service) reconcileRunning(ctx context.Context) error {
	running, err := s.runs.ListRunning(ctx, s.cfg.RunningCheckLimit)
	if err != nil {
		return fmt.Errorf("list running runs: %w", err)
	}

	for _, run := range running {
		execution := resolveRunExecutionContext(run.RunID, run.ProjectID, run.RunPayload, s.cfg.RunNamespacePrefix)
		ref := s.launcher.JobRef(run.RunID, execution.Namespace)
		state, err := s.launcher.Status(ctx, ref)
		if err != nil {
			s.logger.Error("check run job status failed", "run_id", run.RunID, "job_name", ref.Name, "err", err)
			continue
		}

		switch state {
		case JobStateSucceeded:
			if err := s.finishRun(ctx, finishRunParams{
				Run:       run,
				Execution: execution,
				Status:    rundomain.StatusSucceeded,
				EventType: floweventdomain.EventTypeRunSucceeded,
				Ref:       ref,
			}); err != nil {
				return err
			}
		case JobStateFailed:
			if err := s.finishRun(ctx, finishRunParams{
				Run:       run,
				Execution: execution,
				Status:    rundomain.StatusFailed,
				EventType: floweventdomain.EventTypeRunFailed,
				Ref:       ref,
				Extra: runFinishedEventExtra{
					Reason: runFailureReasonKubernetesJobFailed,
				},
			}); err != nil {
				return err
			}
		case JobStateNotFound:
			if err := s.finishRun(ctx, finishRunParams{
				Run:       run,
				Execution: execution,
				Status:    rundomain.StatusFailed,
				EventType: floweventdomain.EventTypeRunFailedJobNotFound,
				Ref:       ref,
				Extra: runFinishedEventExtra{
					Reason: runFailureReasonKubernetesJobNotFound,
				},
			}); err != nil {
				return err
			}
		case JobStatePending, JobStateRunning:
			continue
		default:
			s.logger.Warn("unknown job state", "run_id", run.RunID, "state", state)
		}
	}

	return nil
}

// launchPending claims pending runs, prepares runtime namespace (for full-env), and launches Kubernetes jobs.
func (s *Service) launchPending(ctx context.Context) error {
	for range s.cfg.ClaimLimit {
		claimed, ok, err := s.runs.ClaimNextPending(ctx, runqueuerepo.ClaimParams{
			WorkerID:                   s.cfg.WorkerID,
			SlotsPerProject:            s.cfg.SlotsPerProject,
			LeaseTTL:                   s.cfg.SlotLeaseTTL,
			ProjectLearningModeDefault: s.cfg.ProjectLearningModeDefault,
		})
		if err != nil {
			return fmt.Errorf("claim pending run: %w", err)
		}
		if !ok {
			return nil
		}

		execution := resolveRunExecutionContext(claimed.RunID, claimed.ProjectID, claimed.RunPayload, s.cfg.RunNamespacePrefix)
		runningRun := runningRunFromClaimed(claimed)
		if execution.RuntimeMode != agentdomain.RuntimeModeFullEnv {
			if err := s.finishRun(ctx, finishRunParams{
				Run:       runningRun,
				Execution: execution,
				Status:    rundomain.StatusSucceeded,
				EventType: floweventdomain.EventTypeRunSucceeded,
				Ref:       s.launcher.JobRef(claimed.RunID, execution.Namespace),
			}); err != nil {
				return fmt.Errorf("finish code-only run: %w", err)
			}
			continue
		}

		agentCtx, err := resolveRunAgentContext(claimed.RunPayload, runAgentDefaults{
			DefaultModel:           s.cfg.AgentDefaultModel,
			DefaultReasoningEffort: s.cfg.AgentDefaultReasoningEffort,
			DefaultLocale:          s.cfg.AgentDefaultLocale,
			AllowGPT53:             strings.TrimSpace(s.cfg.OpenAIAuthFile) != "",
			LabelCatalog:           s.labels,
		})
		if err != nil {
			s.logger.Error("resolve run agent context failed", "run_id", claimed.RunID, "err", err)
			eventType := floweventdomain.EventTypeRunFailedLaunchError
			reason := runFailureReasonAgentContextResolve
			if isFailedPreconditionError(err) {
				eventType = floweventdomain.EventTypeRunFailedPrecondition
				reason = runFailureReasonPreconditionFailed
			}
			if finishErr := s.finishRun(ctx, finishRunParams{
				Run:       runningRunFromClaimed(claimed),
				Execution: execution,
				Status:    rundomain.StatusFailed,
				EventType: eventType,
				Ref:       s.launcher.JobRef(claimed.RunID, execution.Namespace),
				Extra: runFinishedEventExtra{
					Error:  err.Error(),
					Reason: reason,
				},
			}); finishErr != nil {
				return fmt.Errorf("mark run failed after context resolve error: %w", finishErr)
			}
			continue
		}

		namespaceSpec := NamespaceSpec{
			RunID:         claimed.RunID,
			ProjectID:     claimed.ProjectID,
			CorrelationID: claimed.CorrelationID,
			RuntimeMode:   execution.RuntimeMode,
			Namespace:     execution.Namespace,
		}
		if execution.RuntimeMode == agentdomain.RuntimeModeFullEnv {
			if err := s.launcher.EnsureNamespace(ctx, namespaceSpec); err != nil {
				s.logger.Error(
					"prepare run namespace failed",
					"run_id", claimed.RunID,
					"namespace", execution.Namespace,
					"runtime_mode", execution.RuntimeMode,
					"err", err,
				)
				if finishErr := s.finishLaunchFailedRun(ctx, runningRun, execution, err, runFailureReasonNamespacePrepareFailed); finishErr != nil {
					return fmt.Errorf("mark run failed after namespace prepare error: %w", finishErr)
				}
				continue
			}

			if err := s.insertNamespaceLifecycleEvent(ctx, namespaceLifecycleEventParams{
				CorrelationID: claimed.CorrelationID,
				EventType:     floweventdomain.EventTypeRunNamespacePrepared,
				RunID:         claimed.RunID,
				ProjectID:     claimed.ProjectID,
				Execution:     execution,
			}); err != nil {
				return fmt.Errorf("insert run.namespace.prepared event: %w", err)
			}
		}

		issuedMCPToken, err := s.mcpTokens.IssueRunMCPToken(ctx, IssueMCPTokenParams{
			RunID:       claimed.RunID,
			Namespace:   execution.Namespace,
			RuntimeMode: execution.RuntimeMode,
		})
		if err != nil {
			s.logger.Error("issue run mcp token failed", "run_id", claimed.RunID, "err", err)
			if finishErr := s.finishLaunchFailedRun(ctx, runningRun, execution, err, runFailureReasonMCPTokenIssueFailed); finishErr != nil {
				return fmt.Errorf("mark run failed after mcp token issue error: %w", finishErr)
			}
			continue
		}

		ref, err := s.launcher.Launch(ctx, JobSpec{
			RunID:                  claimed.RunID,
			CorrelationID:          claimed.CorrelationID,
			ProjectID:              claimed.ProjectID,
			SlotNo:                 claimed.SlotNo,
			RuntimeMode:            execution.RuntimeMode,
			Namespace:              execution.Namespace,
			ControlPlaneGRPCTarget: s.cfg.ControlPlaneGRPCTarget,
			MCPBaseURL:             s.cfg.ControlPlaneMCPBaseURL,
			MCPBearerToken:         issuedMCPToken.Token,
			RepositoryFullName:     agentCtx.RepositoryFullName,
			IssueNumber:            agentCtx.IssueNumber,
			TriggerKind:            agentCtx.TriggerKind,
			TriggerLabel:           agentCtx.TriggerLabel,
			TargetBranch:           agentCtx.TargetBranch,
			ExistingPRNumber:       agentCtx.ExistingPRNumber,
			AgentKey:               agentCtx.AgentKey,
			AgentModel:             agentCtx.Model,
			AgentReasoningEffort:   agentCtx.ReasoningEffort,
			PromptTemplateKind:     agentCtx.PromptTemplateKind,
			PromptTemplateSource:   agentCtx.PromptTemplateSource,
			PromptTemplateLocale:   agentCtx.PromptTemplateLocale,
			StateInReviewLabel:     s.cfg.StateInReviewLabel,
			BaseBranch:             s.cfg.AgentBaseBranch,
			OpenAIAPIKey:           s.cfg.OpenAIAPIKey,
			OpenAIAuthFile:         s.cfg.OpenAIAuthFile,
			Context7APIKey:         s.cfg.Context7APIKey,
			GitBotToken:            s.cfg.GitBotToken,
			AgentDisplayName:       agentCtx.AgentDisplayName,
			GitBotUsername:         s.cfg.GitBotUsername,
			GitBotMail:             s.cfg.GitBotMail,
		})
		if err != nil {
			s.logger.Error("launch run job failed", "run_id", claimed.RunID, "err", err)
			if finishErr := s.finishRun(ctx, finishRunParams{
				Run:       runningRun,
				Execution: execution,
				Status:    rundomain.StatusFailed,
				EventType: floweventdomain.EventTypeRunFailedLaunchError,
				Ref:       ref,
				Extra: runFinishedEventExtra{
					Error: err.Error(),
				},
			}); finishErr != nil {
				return fmt.Errorf("mark run failed after launch error: %w", finishErr)
			}
			continue
		}

		if err := s.insertEvent(ctx, floweventrepo.InsertParams{
			CorrelationID: claimed.CorrelationID,
			ActorType:     floweventdomain.ActorTypeSystem,
			ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
			EventType:     floweventdomain.EventTypeRunStarted,
			Payload: encodeRunStartedEventPayload(runStartedEventPayload{
				RunID:                claimed.RunID,
				ProjectID:            claimed.ProjectID,
				SlotNo:               claimed.SlotNo,
				JobName:              ref.Name,
				JobNamespace:         ref.Namespace,
				RuntimeMode:          execution.RuntimeMode,
				RepositoryFullName:   agentCtx.RepositoryFullName,
				AgentKey:             agentCtx.AgentKey,
				IssueNumber:          agentCtx.IssueNumber,
				TriggerKind:          agentCtx.TriggerKind,
				TriggerLabel:         agentCtx.TriggerLabel,
				Model:                agentCtx.Model,
				ModelSource:          agentCtx.ModelSource,
				ReasoningEffort:      agentCtx.ReasoningEffort,
				ReasoningSource:      agentCtx.ReasoningSource,
				PromptTemplateKind:   agentCtx.PromptTemplateKind,
				PromptTemplateSource: agentCtx.PromptTemplateSource,
				PromptTemplateLocale: agentCtx.PromptTemplateLocale,
				BaseBranch:           s.cfg.AgentBaseBranch,
			}),
			CreatedAt: s.now().UTC(),
		}); err != nil {
			return fmt.Errorf("insert run.started event: %w", err)
		}

		if _, err := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
			RunID:           claimed.RunID,
			Phase:           RunStatusPhaseStarted,
			JobName:         ref.Name,
			JobNamespace:    ref.Namespace,
			RuntimeMode:     string(execution.RuntimeMode),
			Namespace:       execution.Namespace,
			TriggerKind:     agentCtx.TriggerKind,
			PromptLocale:    agentCtx.PromptTemplateLocale,
			Model:           agentCtx.Model,
			ReasoningEffort: agentCtx.ReasoningEffort,
			RunStatus:       string(rundomain.StatusRunning),
		}); err != nil {
			s.logger.Warn("upsert run status comment (started) failed", "run_id", claimed.RunID, "err", err)
		}
	}

	return nil
}

// finishRun persists terminal run state, emits flow events, and finalizes runtime namespace lifecycle.
func (s *Service) finishRun(ctx context.Context, params finishRunParams) error {
	finishedAt := s.now().UTC()
	updated, err := s.runs.FinishRun(ctx, runqueuerepo.FinishParams{
		RunID:      params.Run.RunID,
		ProjectID:  params.Run.ProjectID,
		Status:     params.Status,
		FinishedAt: finishedAt,
	})
	if err != nil {
		return fmt.Errorf("finish run %s as %s: %w", params.Run.RunID, params.Status, err)
	}
	if !updated {
		return nil
	}

	payload := runFinishedEventPayload{
		RunID:        params.Run.RunID,
		ProjectID:    params.Run.ProjectID,
		Status:       params.Status,
		JobName:      params.Ref.Name,
		JobNamespace: params.Ref.Namespace,
		RuntimeMode:  params.Execution.RuntimeMode,
		Namespace:    params.Execution.Namespace,
		Error:        params.Extra.Error,
		Reason:       params.Extra.Reason,
	}

	if err := s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: params.Run.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
		EventType:     params.EventType,
		Payload:       encodeRunFinishedEventPayload(payload),
		CreatedAt:     finishedAt,
	}); err != nil {
		return fmt.Errorf("insert finish event: %w", err)
	}

	if params.Execution.RuntimeMode == agentdomain.RuntimeModeFullEnv {
		if _, err := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
			RunID:        params.Run.RunID,
			Phase:        RunStatusPhaseFinished,
			JobName:      params.Ref.Name,
			JobNamespace: params.Ref.Namespace,
			RuntimeMode:  string(params.Execution.RuntimeMode),
			Namespace:    params.Execution.Namespace,
			RunStatus:    string(params.Status),
		}); err != nil {
			s.logger.Warn("upsert run status comment (finished) failed", "run_id", params.Run.RunID, "err", err)
		}
	}

	if params.Run.LearningMode && s.feedback != nil {
		namespace := params.Ref.Namespace
		if params.Execution.Namespace != "" {
			namespace = params.Execution.Namespace
		}
		explanation := fmt.Sprintf(
			"Learning mode is enabled for this run.\n\n"+
				"Why this is executed as a Kubernetes Job: it provides isolation, reproducibility and clear lifecycle states.\n"+
				"Why we use DB-backed slots: it prevents concurrent workers from overloading a project and makes multi-pod behavior deterministic.\n"+
				"Tradeoffs: Jobs are heavier than in-process execution; DB locking requires careful indexing and timeouts.\n\n"+
				"Result: status=%s, job=%s/%s.",
			params.Status,
			namespace,
			params.Ref.Name,
		)
		if err := s.feedback.Insert(ctx, learningfeedbackrepo.InsertParams{
			RunID:       params.Run.RunID,
			Kind:        learningfeedbackrepo.KindInline,
			Explanation: explanation,
		}); err != nil {
			s.logger.Error("insert learning feedback failed", "run_id", params.Run.RunID, "err", err)
		}
	}

	if params.Execution.RuntimeMode == agentdomain.RuntimeModeFullEnv &&
		params.Execution.Namespace != "" {
		debugPolicy := s.resolveRunDebugPolicy(params.Run.RunPayload)
		if debugPolicy.SkipCleanup {
			if err := s.insertNamespaceLifecycleEvent(ctx, namespaceLifecycleEventParams{
				CorrelationID: params.Run.CorrelationID,
				EventType:     floweventdomain.EventTypeRunNamespaceCleanupSkipped,
				RunID:         params.Run.RunID,
				ProjectID:     params.Run.ProjectID,
				Execution:     params.Execution,
				Extra: namespaceLifecycleEventExtra{
					Reason:         debugPolicy.Reason,
					CleanupCommand: buildNamespaceCleanupCommand(params.Execution.Namespace),
				},
			}); err != nil {
				s.logger.Error("insert run.namespace.cleanup_skipped event failed", "run_id", params.Run.RunID, "err", err)
			}
			return nil
		}

		cleanupSpec := NamespaceSpec{
			RunID:         params.Run.RunID,
			ProjectID:     params.Run.ProjectID,
			CorrelationID: params.Run.CorrelationID,
			RuntimeMode:   params.Execution.RuntimeMode,
			Namespace:     params.Execution.Namespace,
		}
		cleanupErr := s.launcher.CleanupNamespace(ctx, cleanupSpec)
		if cleanupErr != nil {
			s.logger.Error(
				"cleanup run namespace failed",
				"run_id", params.Run.RunID,
				"namespace", params.Execution.Namespace,
				"err", cleanupErr,
			)
			if err := s.insertNamespaceLifecycleEvent(ctx, namespaceLifecycleEventParams{
				CorrelationID: params.Run.CorrelationID,
				EventType:     floweventdomain.EventTypeRunNamespaceCleanupFailed,
				RunID:         params.Run.RunID,
				ProjectID:     params.Run.ProjectID,
				Execution:     params.Execution,
				Extra: namespaceLifecycleEventExtra{
					Error: cleanupErr.Error(),
				},
			}); err != nil {
				s.logger.Error("insert run.namespace.cleanup_failed event failed", "run_id", params.Run.RunID, "err", err)
			}
		} else {
			if err := s.insertNamespaceLifecycleEvent(ctx, namespaceLifecycleEventParams{
				CorrelationID: params.Run.CorrelationID,
				EventType:     floweventdomain.EventTypeRunNamespaceCleaned,
				RunID:         params.Run.RunID,
				ProjectID:     params.Run.ProjectID,
				Execution:     params.Execution,
			}); err != nil {
				s.logger.Error("insert run.namespace.cleaned event failed", "run_id", params.Run.RunID, "err", err)
			}
			if _, err := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
				RunID:        params.Run.RunID,
				Phase:        RunStatusPhaseNamespaceDeleted,
				JobName:      params.Ref.Name,
				JobNamespace: params.Ref.Namespace,
				RuntimeMode:  string(params.Execution.RuntimeMode),
				Namespace:    params.Execution.Namespace,
				RunStatus:    string(params.Status),
				Deleted:      true,
			}); err != nil {
				s.logger.Warn("upsert run status comment (namespace deleted) failed", "run_id", params.Run.RunID, "err", err)
			}
		}
	}

	return nil
}

func (s *Service) finishLaunchFailedRun(ctx context.Context, run runqueuerepo.RunningRun, execution valuetypes.RunExecutionContext, failure error, reason runFailureReason) error {
	return s.finishRun(ctx, finishRunParams{
		Run:       run,
		Execution: execution,
		Status:    rundomain.StatusFailed,
		EventType: floweventdomain.EventTypeRunFailedLaunchError,
		Extra: runFinishedEventExtra{
			Error:  failure.Error(),
			Reason: reason,
		},
	})
}

// insertEvent persists one flow event with contextual error wrapping.
func (s *Service) insertEvent(ctx context.Context, params floweventrepo.InsertParams) error {
	if err := s.events.Insert(ctx, params); err != nil {
		return fmt.Errorf("insert flow event %s for correlation %s: %w", params.EventType, params.CorrelationID, err)
	}
	return nil
}

// insertNamespaceLifecycleEvent records namespace lifecycle transitions in flow_events.
func (s *Service) insertNamespaceLifecycleEvent(ctx context.Context, params namespaceLifecycleEventParams) error {
	return s.insertEvent(ctx, floweventrepo.InsertParams{
		CorrelationID: params.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorID(s.cfg.WorkerID),
		EventType:     params.EventType,
		Payload: encodeNamespaceLifecycleEventPayload(namespaceLifecycleEventPayload{
			RunID:          params.RunID,
			ProjectID:      params.ProjectID,
			RuntimeMode:    params.Execution.RuntimeMode,
			Namespace:      params.Execution.Namespace,
			Error:          params.Extra.Error,
			Reason:         params.Extra.Reason,
			CleanupCommand: params.Extra.CleanupCommand,
		}),
		CreatedAt: s.now().UTC(),
	})
}

// runningRunFromClaimed reuses claimed fields for failure finalization paths before the next reconcile tick.
func runningRunFromClaimed(claimed runqueuerepo.ClaimedRun) runqueuerepo.RunningRun {
	return runqueuerepo.RunningRun{
		RunID:         claimed.RunID,
		CorrelationID: claimed.CorrelationID,
		ProjectID:     claimed.ProjectID,
		LearningMode:  claimed.LearningMode,
		RunPayload:    claimed.RunPayload,
	}
}

func isFailedPreconditionError(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(strings.TrimSpace(err.Error()), "failed_precondition:")
}
