package worker

import (
	"context"
	"log/slog"
	"strings"
	"time"

	floweventrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/flowevent"
	learningfeedbackrepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/learningfeedback"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
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
	// KubernetesNamespace is default worker namespace for run workloads.
	KubernetesNamespace string
	// AIRepairNamespace is namespace used by ai-repair workload runs.
	AIRepairNamespace string
	// AIRepairServiceAccount is service account used by ai-repair workload pod.
	AIRepairServiceAccount string
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
	cfg.KubernetesNamespace = strings.TrimSpace(cfg.KubernetesNamespace)
	if cfg.KubernetesNamespace == "" {
		cfg.KubernetesNamespace = "default"
	}
	cfg.AIRepairNamespace = sanitizeDNSLabelValue(cfg.AIRepairNamespace)
	if cfg.AIRepairNamespace == "" {
		cfg.AIRepairNamespace = cfg.KubernetesNamespace
	}
	cfg.AIRepairServiceAccount = strings.TrimSpace(cfg.AIRepairServiceAccount)
	if cfg.AIRepairServiceAccount == "" {
		cfg.AIRepairServiceAccount = "codex-k8s-control-plane"
	}
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
