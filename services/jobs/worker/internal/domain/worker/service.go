package worker

import (
	"context"
	"errors"
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
	// RunLeaseTTL defines maximum duration of running-run ownership by one worker.
	RunLeaseTTL time.Duration
	// RuntimePrepareRetryTimeout limits total retry time for runtime deploy preparation.
	RuntimePrepareRetryTimeout time.Duration
	// RuntimePrepareRetryInterval defines delay between retryable runtime deploy attempts.
	RuntimePrepareRetryInterval time.Duration

	// ProjectLearningModeDefault is applied when the worker auto-creates projects from webhook payloads.
	ProjectLearningModeDefault bool
	// RunNamespacePrefix defines prefix for full-env run namespaces.
	RunNamespacePrefix string
	// DefaultNamespaceTTL applies to full-env namespace retention when role-specific override is absent.
	DefaultNamespaceTTL time.Duration
	// NamespaceTTLByRole contains full-env namespace retention overrides per agent role key.
	NamespaceTTLByRole map[string]time.Duration
	// NamespaceLeaseSweepLimit limits how many expired managed namespaces are cleaned per tick.
	NamespaceLeaseSweepLimit int
	// StateInReviewLabel is applied to PR when run is ready for owner review.
	StateInReviewLabel string
	// ControlPlaneGRPCTarget is control-plane gRPC endpoint used by run jobs for callbacks.
	ControlPlaneGRPCTarget string
	// ControlPlaneMCPBaseURL is MCP endpoint passed to run job environment.
	ControlPlaneMCPBaseURL string
	// OpenAIAPIKey is injected into run pods for codex login.
	OpenAIAPIKey string
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
	// JobImage is primary image for run Jobs.
	JobImage string
	// JobImageFallback is optional fallback image for run Jobs.
	JobImageFallback string
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
	// RuntimePreparer prepares runtime environment stack before run job launch.
	RuntimePreparer RuntimeEnvironmentPreparer
	// MCPTokenIssuer issues short-lived MCP token for run pods.
	MCPTokenIssuer MCPTokenIssuer
	// RunStatus updates one run-bound issue status comment.
	RunStatus RunStatusNotifier
	// Logger records worker diagnostics.
	Logger *slog.Logger
	// JobImageChecker checks whether image references are available before launch.
	JobImageChecker JobImageAvailabilityChecker
}

// Service orchestrates pending runs to Kubernetes Jobs and final statuses.
type Service struct {
	cfg       Config
	runs      runqueuerepo.Repository
	events    floweventrepo.Repository
	feedback  learningfeedbackrepo.Repository
	launcher  Launcher
	deployer  RuntimeEnvironmentPreparer
	mcpTokens MCPTokenIssuer
	runStatus RunStatusNotifier
	logger    *slog.Logger
	labels    runAgentLabelCatalog
	image     JobImageSelectionPolicy
	now       func() time.Time
}

// JobImageAvailabilityChecker checks run Job image existence.
type JobImageAvailabilityChecker interface {
	IsImageAvailable(ctx context.Context, imageRef string) (bool, error)
	ResolvePreviousImage(ctx context.Context, imageRef string) (string, bool, error)
}

// JobImageSelectionPolicy defines primary/fallback image configuration for run job launches.
type JobImageSelectionPolicy struct {
	Primary  string
	Fallback string
	Checker  JobImageAvailabilityChecker
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
	if cfg.RuntimePrepareRetryTimeout <= 0 {
		cfg.RuntimePrepareRetryTimeout = 30 * time.Minute
	}
	if cfg.RunLeaseTTL <= 0 {
		cfg.RunLeaseTTL = cfg.RuntimePrepareRetryTimeout + 5*time.Minute
	}
	if cfg.RunLeaseTTL <= 0 {
		cfg.RunLeaseTTL = 45 * time.Minute
	}
	if cfg.RuntimePrepareRetryInterval <= 0 {
		cfg.RuntimePrepareRetryInterval = 3 * time.Second
	}
	if cfg.WorkerID == "" {
		cfg.WorkerID = defaultWorkerID
	}
	if cfg.RunNamespacePrefix == "" {
		cfg.RunNamespacePrefix = defaultRunNamespacePrefix
	}
	if cfg.DefaultNamespaceTTL <= 0 {
		cfg.DefaultNamespaceTTL = 24 * time.Hour
	}
	if cfg.NamespaceLeaseSweepLimit <= 0 {
		cfg.NamespaceLeaseSweepLimit = 200
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
	cfg.AgentDefaultModel = strings.TrimSpace(cfg.AgentDefaultModel)
	if cfg.AgentDefaultModel == "" {
		cfg.AgentDefaultModel = modelGPT53Codex
	}
	cfg.AgentDefaultReasoningEffort = strings.TrimSpace(strings.ToLower(cfg.AgentDefaultReasoningEffort))
	switch cfg.AgentDefaultReasoningEffort {
	case "extra-high", "extra_high", "extra high", "x-high":
		cfg.AgentDefaultReasoningEffort = "xhigh"
	}
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
	cfg.JobImage = strings.TrimSpace(cfg.JobImage)
	cfg.JobImageFallback = strings.TrimSpace(cfg.JobImageFallback)
	labelCatalog := runAgentLabelCatalogFromConfig(cfg)
	cfg.NamespaceTTLByRole = normalizeNamespaceTTLByRole(cfg.NamespaceTTLByRole)
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	if deps.MCPTokenIssuer == nil {
		deps.MCPTokenIssuer = noopMCPTokenIssuer{}
	}
	if deps.RuntimePreparer == nil {
		deps.RuntimePreparer = noopRuntimeEnvironmentPreparer{}
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
		deployer:  deps.RuntimePreparer,
		mcpTokens: deps.MCPTokenIssuer,
		runStatus: deps.RunStatus,
		logger:    deps.Logger,
		labels:    labelCatalog,
		image: JobImageSelectionPolicy{
			Primary:  cfg.JobImage,
			Fallback: cfg.JobImageFallback,
			Checker:  deps.JobImageChecker,
		},
		now: time.Now,
	}
}

// Tick executes one reconciliation iteration.
func (s *Service) Tick(ctx context.Context) error {
	if err := s.cleanupExpiredNamespaces(ctx); err != nil {
		return fmt.Errorf("cleanup expired namespaces: %w", err)
	}
	if err := s.reconcileRunning(ctx); err != nil {
		return fmt.Errorf("reconcile running runs: %w", err)
	}
	if err := s.launchPending(ctx); err != nil {
		return fmt.Errorf("launch pending runs: %w", err)
	}
	return nil
}

func (s *Service) cleanupExpiredNamespaces(ctx context.Context) error {
	cleaned, err := s.launcher.CleanupExpiredNamespaces(ctx, NamespaceCleanupParams{
		Now:   s.now().UTC(),
		Limit: s.cfg.NamespaceLeaseSweepLimit,
	})
	if err != nil {
		return err
	}
	for _, item := range cleaned {
		s.logger.Info(
			"cleaned expired run namespace",
			"namespace", item.Namespace,
			"run_id", item.RunID,
			"expires_at", item.ExpiresAt.Format(time.RFC3339),
		)
		runID := strings.TrimSpace(item.RunID)
		if runID == "" {
			continue
		}
		if _, upsertErr := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
			RunID:       runID,
			Phase:       RunStatusPhaseNamespaceDeleted,
			RuntimeMode: string(agentdomain.RuntimeModeFullEnv),
			Namespace:   item.Namespace,
			Deleted:     true,
		}); upsertErr != nil {
			s.logger.Warn("upsert run status comment (namespace ttl cleanup) failed", "run_id", runID, "namespace", item.Namespace, "err", upsertErr)
		}
	}
	return nil
}

// reconcileRunning polls active runs and finalizes those with terminal Kubernetes job states.
func (s *Service) reconcileRunning(ctx context.Context) error {
	running, err := s.runs.ClaimRunning(ctx, runqueuerepo.ClaimRunningParams{
		WorkerID: s.cfg.WorkerID,
		LeaseTTL: s.cfg.RunLeaseTTL,
		Limit:    s.cfg.RunningCheckLimit,
	})
	if err != nil {
		return fmt.Errorf("claim running runs: %w", err)
	}

	for _, run := range running {
		s.keepRunSlotLeaseAlive(ctx, run)

		execution := resolveRunExecutionContext(run.RunID, run.ProjectID, run.RunPayload, s.cfg.RunNamespacePrefix)
		runtimePayload := parseRunRuntimePayload(run.RunPayload)
		deployOnlyRun := runtimePayload.Runtime != nil && runtimePayload.Runtime.DeployOnly

		if execution.RuntimeMode != agentdomain.RuntimeModeFullEnv && !deployOnlyRun {
			if err := s.finishRun(ctx, finishRunParams{
				Run:       run,
				Execution: execution,
				Status:    rundomain.StatusSucceeded,
				EventType: floweventdomain.EventTypeRunSucceeded,
				Ref:       s.launcher.JobRef(run.RunID, execution.Namespace),
			}); err != nil {
				return err
			}
			continue
		}

		if deployOnlyRun {
			prepareParams := buildPrepareRunEnvironmentParamsFromRunning(run, execution)
			prepared, ready, err := s.prepareRuntimeEnvironmentPoll(ctx, prepareParams)
			if err != nil {
				if errors.Is(err, errRuntimeDeployTaskCanceled) {
					if cancelErr := s.finishRuntimePrepareCanceledRun(ctx, run, execution, true); cancelErr != nil {
						return cancelErr
					}
					continue
				}
				s.logger.Error("prepare runtime environment for running deploy-only run failed", "run_id", run.RunID, "err", err)
				if finishErr := s.finishLaunchFailedRun(ctx, run, execution, err, runFailureReasonRuntimeDeployFailed); finishErr != nil {
					return finishErr
				}
				continue
			}
			if !ready {
				continue
			}

			finishExecution := execution
			if resolvedNamespace := sanitizeDNSLabelValue(prepared.Namespace); resolvedNamespace != "" {
				finishExecution.Namespace = resolvedNamespace
			}

			if err := s.finishRun(ctx, finishRunParams{
				Run:                  run,
				Execution:            finishExecution,
				Status:               rundomain.StatusSucceeded,
				EventType:            floweventdomain.EventTypeRunSucceeded,
				SkipNamespaceCleanup: true,
			}); err != nil {
				return err
			}
			continue
		}

		ref := s.launcher.JobRef(run.RunID, execution.Namespace)
		state, err := s.launcher.Status(ctx, ref)
		if err != nil {
			s.logger.Error("check run job status failed", "run_id", run.RunID, "job_name", ref.Name, "err", err)
			continue
		}

		if state == JobStateNotFound {
			// Full-env runs may be launched into persistent slot namespaces, while run payload keeps
			// the default namespace strategy (`codex-issue-*`). Resolve the actual namespace by label
			// to avoid failing runs with "job not found" after preparation succeeded.
			resolved, ok, err := s.launcher.FindRunJobRefByRunID(ctx, run.RunID)
			if err != nil {
				s.logger.Warn("resolve run job ref by run id failed", "run_id", run.RunID, "err", err)
			} else if ok {
				ref = resolved
				if resolvedNamespace := sanitizeDNSLabelValue(resolved.Namespace); resolvedNamespace != "" {
					execution.Namespace = resolvedNamespace
					ref.Namespace = resolvedNamespace
				}

				state, err = s.launcher.Status(ctx, ref)
				if err != nil {
					s.logger.Error("check run job status failed", "run_id", run.RunID, "job_name", ref.Name, "err", err)
					continue
				}
			}
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
			recovered, err := s.tryRecoverMissingRunJob(ctx, run, execution)
			if err != nil {
				return err
			}
			if recovered {
				continue
			}

			// We mark runs as "running" when they are claimed, but full-env runs may spend
			// significant time in runtime preparation before the actual job exists.
			// With multiple worker replicas this prevents another worker from failing the run
			// while the claiming worker is still preparing the environment.
			if s.shouldIgnoreJobNotFound(run) {
				continue
			}
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

func (s *Service) shouldIgnoreJobNotFound(run runqueuerepo.RunningRun) bool {
	startedAt := run.StartedAt
	if startedAt.IsZero() {
		return true
	}

	grace := s.cfg.RuntimePrepareRetryTimeout
	if grace <= 0 {
		grace = 30 * time.Second
	}
	grace += 5 * time.Second

	now := s.now().UTC()
	if now.Before(startedAt) {
		return true
	}
	return now.Sub(startedAt) < grace
}

// launchPending claims pending runs, prepares runtime namespace (for full-env), and launches Kubernetes jobs.
func (s *Service) launchPending(ctx context.Context) error {
	for range s.cfg.ClaimLimit {
		claimed, ok, err := s.runs.ClaimNextPending(ctx, runqueuerepo.ClaimParams{
			WorkerID:                   s.cfg.WorkerID,
			SlotsPerProject:            s.cfg.SlotsPerProject,
			LeaseTTL:                   s.cfg.SlotLeaseTTL,
			RunLeaseTTL:                s.cfg.RunLeaseTTL,
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
		prepareParams := buildPrepareRunEnvironmentParams(claimed, execution)
		deployOnlyRun := prepareParams.DeployOnly

		if execution.RuntimeMode != agentdomain.RuntimeModeFullEnv && !deployOnlyRun {
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

		runPayload := parseRunRuntimePayload(claimed.RunPayload)
		leaseCtx := resolveNamespaceLeaseContext(claimed.RunPayload)
		leaseTTL := s.cfg.DefaultNamespaceTTL
		triggerKind := ""
		var agentCtx runAgentContext

		if deployOnlyRun {
			if runPayload.Trigger != nil {
				triggerKind = string(runPayload.Trigger.Kind)
			}
		} else {
			agentCtx, err = resolveRunAgentContext(claimed.RunPayload, runAgentDefaults{
				DefaultModel:           s.cfg.AgentDefaultModel,
				DefaultReasoningEffort: s.cfg.AgentDefaultReasoningEffort,
				DefaultLocale:          s.cfg.AgentDefaultLocale,
				AllowGPT53:             true,
				LabelCatalog:           s.labels,
			})
			if err != nil {
				s.logger.Error("resolve run agent context failed", "run_id", claimed.RunID, "err", err)
				if finishErr := s.failRunAfterAgentContextResolve(ctx, runningRun, execution, err); finishErr != nil {
					return finishErr
				}
				continue
			}
			triggerKind = agentCtx.TriggerKind
			if leaseCtx.AgentKey == "" {
				leaseCtx.AgentKey = strings.ToLower(strings.TrimSpace(agentCtx.AgentKey))
			}
			if leaseCtx.IssueNumber <= 0 {
				leaseCtx.IssueNumber = agentCtx.IssueNumber
			}
			if !leaseCtx.IsRevise {
				leaseCtx.IsRevise = resolvePromptTemplateKindForTrigger(agentCtx.TriggerKind) == promptTemplateKindRevise
			}
			leaseTTL = s.resolveNamespaceTTL(leaseCtx.AgentKey)

			if execution.RuntimeMode == agentdomain.RuntimeModeFullEnv &&
				leaseCtx.IsRevise &&
				prepareParams.Namespace == "" &&
				leaseCtx.IssueNumber > 0 &&
				leaseCtx.AgentKey != "" {
				reusableNamespace, found, reuseErr := s.launcher.FindReusableNamespace(ctx, NamespaceReuseLookup{
					ProjectID:   runningRun.ProjectID,
					IssueNumber: leaseCtx.IssueNumber,
					AgentKey:    leaseCtx.AgentKey,
					Now:         s.now().UTC(),
				})
				if reuseErr != nil {
					s.logger.Warn(
						"resolve reusable namespace for revise run failed",
						"run_id", runningRun.RunID,
						"project_id", runningRun.ProjectID,
						"issue_number", leaseCtx.IssueNumber,
						"agent_key", leaseCtx.AgentKey,
						"err", reuseErr,
					)
				} else if found {
					prepareParams.Namespace = reusableNamespace.Namespace
					execution.Namespace = reusableNamespace.Namespace
				}
			}
		}

		if _, err := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
			RunID:       runningRun.RunID,
			Phase:       RunStatusPhasePreparingRuntime,
			RuntimeMode: string(execution.RuntimeMode),
			Namespace:   execution.Namespace,
			TriggerKind: triggerKind,
			RunStatus:   string(rundomain.StatusRunning),
		}); err != nil {
			s.logger.Warn("upsert run status comment (preparing runtime) failed", "run_id", runningRun.RunID, "err", err)
		}

		prepared, ready, err := s.prepareRuntimeEnvironmentPoll(ctx, prepareParams)
		if err != nil {
			if errors.Is(err, errRuntimeDeployTaskCanceled) {
				if cancelErr := s.finishRuntimePrepareCanceledRun(ctx, runningRun, execution, deployOnlyRun); cancelErr != nil {
					return cancelErr
				}
				continue
			}
			s.logger.Error("prepare runtime environment failed", "run_id", claimed.RunID, "err", err)
			if finishErr := s.finishLaunchFailedRun(ctx, runningRun, execution, err, runFailureReasonRuntimeDeployFailed); finishErr != nil {
				return fmt.Errorf("mark run failed after runtime deploy error: %w", finishErr)
			}
			continue
		}
		if !ready {
			continue
		}

		launchExecution := execution
		if resolvedNamespace := sanitizeDNSLabelValue(prepared.Namespace); resolvedNamespace != "" {
			launchExecution.Namespace = resolvedNamespace
		}

		if deployOnlyRun {
			if err := s.finishRun(ctx, finishRunParams{
				Run:                  runningRun,
				Execution:            launchExecution,
				Status:               rundomain.StatusSucceeded,
				EventType:            floweventdomain.EventTypeRunSucceeded,
				SkipNamespaceCleanup: true,
			}); err != nil {
				return fmt.Errorf("finish deploy-only run: %w", err)
			}
			continue
		}

		if err := s.launchPreparedFullEnvRunJob(ctx, runningRun, launchExecution, agentCtx, namespaceLeaseSpec{
			AgentKey:    leaseCtx.AgentKey,
			IssueNumber: leaseCtx.IssueNumber,
			TTL:         leaseTTL,
		}); err != nil {
			return err
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
		LeaseOwner: s.cfg.WorkerID,
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
		params.Execution.Namespace != "" &&
		!params.SkipNamespaceCleanup {
		s.upsertNamespaceStatusComment(ctx, params, false, "upsert run status comment (namespace retained by ttl policy) failed")
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

func (s *Service) upsertNamespaceStatusComment(ctx context.Context, params finishRunParams, deleted bool, warnMessage string) {
	if _, err := s.runStatus.UpsertRunStatusComment(ctx, RunStatusCommentParams{
		RunID:        params.Run.RunID,
		Phase:        RunStatusPhaseNamespaceDeleted,
		JobName:      params.Ref.Name,
		JobNamespace: params.Ref.Namespace,
		RuntimeMode:  string(params.Execution.RuntimeMode),
		Namespace:    params.Execution.Namespace,
		RunStatus:    string(params.Status),
		Deleted:      deleted,
	}); err != nil {
		s.logger.Warn(warnMessage, "run_id", params.Run.RunID, "err", err)
	}
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
			NamespaceLeaseTTL: func() string {
				if params.Extra.NamespaceLeaseTTL <= 0 {
					return ""
				}
				return params.Extra.NamespaceLeaseTTL.String()
			}(),
			NamespaceLeaseExpiresAt: func() string {
				if params.Extra.NamespaceLeaseExpiresAt.IsZero() {
					return ""
				}
				return params.Extra.NamespaceLeaseExpiresAt.UTC().Format(time.RFC3339)
			}(),
			NamespaceReused: params.Extra.NamespaceReused,
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
		SlotID:        claimed.SlotID,
		SlotNo:        claimed.SlotNo,
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
