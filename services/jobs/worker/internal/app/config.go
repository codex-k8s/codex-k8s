package app

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

// Config defines environment-backed runtime settings for worker service.
type Config struct {
	// WorkerID identifies current worker instance in logs and events.
	WorkerID string `env:"CODEXK8S_WORKER_ID"`
	// PollInterval controls tick interval for run-loop.
	PollInterval string `env:"CODEXK8S_WORKER_POLL_INTERVAL" envDefault:"5s"`
	// ClaimLimit controls how many pending runs worker claims per tick.
	ClaimLimit int `env:"CODEXK8S_WORKER_CLAIM_LIMIT" envDefault:"10"`
	// RunningCheckLimit controls how many running runs are reconciled per tick.
	RunningCheckLimit int `env:"CODEXK8S_WORKER_RUNNING_CHECK_LIMIT" envDefault:"200"`
	// SlotsPerProject defines initial slot pool size per project.
	SlotsPerProject int `env:"CODEXK8S_WORKER_SLOTS_PER_PROJECT" envDefault:"2"`
	// SlotLeaseTTL controls for how long slot is leased before expiration.
	SlotLeaseTTL string `env:"CODEXK8S_WORKER_SLOT_LEASE_TTL" envDefault:"10m"`
	// RunLeaseTTL controls for how long one worker owns a running run reconciliation lease.
	RunLeaseTTL string `env:"CODEXK8S_WORKER_RUN_LEASE_TTL" envDefault:"45m"`
	// TickTimeout limits one worker Tick execution duration.
	TickTimeout string `env:"CODEXK8S_WORKER_TICK_TIMEOUT" envDefault:"45m"`
	// RuntimePrepareRetryTimeout limits total retry time for runtime deploy preparation RPC.
	RuntimePrepareRetryTimeout string `env:"CODEXK8S_WORKER_RUNTIME_PREPARE_RETRY_TIMEOUT" envDefault:"30m"`
	// RuntimePrepareRetryInterval defines delay between retryable runtime deploy preparation attempts.
	RuntimePrepareRetryInterval string `env:"CODEXK8S_WORKER_RUNTIME_PREPARE_RETRY_INTERVAL" envDefault:"3s"`
	// ServicesConfigPath points to services.yaml for runtime policy (mode/namespace TTL).
	ServicesConfigPath string `env:"CODEXK8S_SERVICES_CONFIG_PATH" envDefault:"services.yaml"`
	// ServicesConfigEnv selects render environment for services.yaml policy.
	ServicesConfigEnv string `env:"CODEXK8S_SERVICES_CONFIG_ENV" envDefault:"production"`

	// LearningModeDefault controls default project learning-mode when worker auto-creates projects.
	// Keep empty value to disable by default; set to "true" to enable by default.
	LearningModeDefault string `env:"CODEXK8S_LEARNING_MODE_DEFAULT"`

	// ControlPlaneGRPCTarget is control-plane gRPC address used for internal worker calls.
	ControlPlaneGRPCTarget string `env:"CODEXK8S_CONTROL_PLANE_GRPC_TARGET,required,notEmpty"`
	// ControlPlaneMCPBaseURL is optional MCP HTTP endpoint passed into spawned run pods.
	// When empty, worker derives it from ControlPlaneGRPCTarget.
	ControlPlaneMCPBaseURL string `env:"CODEXK8S_CONTROL_PLANE_MCP_BASE_URL"`
	// OpenAIAPIKey is injected into run pods for codex login.
	OpenAIAPIKey string `env:"CODEXK8S_OPENAI_API_KEY"`
	// Context7APIKey enables Context7 documentation calls from run pods when set.
	Context7APIKey string `env:"CODEXK8S_CONTEXT7_API_KEY"`
	// GitBotToken is injected into run pods for git transport (fetch/push only).
	GitBotToken string `env:"CODEXK8S_GIT_BOT_TOKEN"`
	// GitBotUsername is GitHub username used with bot token for git transport auth.
	GitBotUsername string `env:"CODEXK8S_GIT_BOT_USERNAME" envDefault:"codex-bot"`
	// GitBotMail is git author email configured in run pods.
	GitBotMail string `env:"CODEXK8S_GIT_BOT_MAIL" envDefault:"codex-bot@codex-k8s.local"`
	// AgentDefaultModel is fallback model when run config labels do not override model.
	AgentDefaultModel string `env:"CODEXK8S_AGENT_DEFAULT_MODEL" envDefault:"gpt-5.2-codex"`
	// AgentDefaultReasoningEffort is fallback reasoning profile when run config labels do not override reasoning.
	AgentDefaultReasoningEffort string `env:"CODEXK8S_AGENT_DEFAULT_REASONING_EFFORT" envDefault:"xhigh"`
	// AgentDefaultLocale is fallback prompt locale.
	AgentDefaultLocale string `env:"CODEXK8S_AGENT_DEFAULT_LOCALE" envDefault:"ru"`
	// AgentBaseBranch is default base branch for PR flow.
	AgentBaseBranch string `env:"CODEXK8S_AGENT_BASE_BRANCH" envDefault:"main"`
	// AIModelGPT53CodexLabel configures label -> model mapping for gpt-5.3-codex.
	AIModelGPT53CodexLabel string `env:"CODEXK8S_AI_MODEL_GPT_5_3_CODEX_LABEL" envDefault:"[ai-model-gpt-5.3-codex]"`
	// AIModelGPT53CodexSparkLabel configures label -> model mapping for gpt-5.3-codex-spark.
	AIModelGPT53CodexSparkLabel string `env:"CODEXK8S_AI_MODEL_GPT_5_3_CODEX_SPARK_LABEL" envDefault:"[ai-model-gpt-5.3-codex-spark]"`
	// AIModelGPT52CodexLabel configures label -> model mapping for gpt-5.2-codex.
	AIModelGPT52CodexLabel string `env:"CODEXK8S_AI_MODEL_GPT_5_2_CODEX_LABEL" envDefault:"[ai-model-gpt-5.2-codex]"`
	// AIModelGPT52Label configures label -> model mapping for gpt-5.2.
	AIModelGPT52Label string `env:"CODEXK8S_AI_MODEL_GPT_5_2_LABEL" envDefault:"[ai-model-gpt-5.2]"`
	// AIModelGPT51CodexMaxLabel configures label -> model mapping for gpt-5.1-codex-max.
	AIModelGPT51CodexMaxLabel string `env:"CODEXK8S_AI_MODEL_GPT_5_1_CODEX_MAX_LABEL" envDefault:"[ai-model-gpt-5.1-codex-max]"`
	// AIModelGPT51CodexMiniLabel configures label -> model mapping for gpt-5.1-codex-mini.
	AIModelGPT51CodexMiniLabel string `env:"CODEXK8S_AI_MODEL_GPT_5_1_CODEX_MINI_LABEL" envDefault:"[ai-model-gpt-5.1-codex-mini]"`
	// AIReasoningLowLabel configures label -> reasoning mapping for low profile.
	AIReasoningLowLabel string `env:"CODEXK8S_AI_REASONING_LOW_LABEL" envDefault:"[ai-reasoning-low]"`
	// AIReasoningMediumLabel configures label -> reasoning mapping for medium profile.
	AIReasoningMediumLabel string `env:"CODEXK8S_AI_REASONING_MEDIUM_LABEL" envDefault:"[ai-reasoning-medium]"`
	// AIReasoningHighLabel configures label -> reasoning mapping for high profile.
	AIReasoningHighLabel string `env:"CODEXK8S_AI_REASONING_HIGH_LABEL" envDefault:"[ai-reasoning-high]"`
	// AIReasoningExtraHighLabel configures label -> reasoning mapping for extra-high profile.
	AIReasoningExtraHighLabel string `env:"CODEXK8S_AI_REASONING_EXTRA_HIGH_LABEL" envDefault:"[ai-reasoning-extra-high]"`

	// DBHost is the PostgreSQL host.
	DBHost string `env:"CODEXK8S_DB_HOST,required,notEmpty"`
	// DBPort is the PostgreSQL port.
	DBPort int `env:"CODEXK8S_DB_PORT" envDefault:"5432"`
	// DBName is the PostgreSQL database name.
	DBName string `env:"CODEXK8S_DB_NAME,required,notEmpty"`
	// DBUser is the PostgreSQL username.
	DBUser string `env:"CODEXK8S_DB_USER,required,notEmpty"`
	// DBPassword is the PostgreSQL password.
	DBPassword string `env:"CODEXK8S_DB_PASSWORD,required,notEmpty"`
	// DBSSLMode is the PostgreSQL SSL mode.
	DBSSLMode string `env:"CODEXK8S_DB_SSLMODE" envDefault:"disable"`

	// KubeconfigPath is optional kubeconfig path for local development.
	KubeconfigPath string `env:"CODEXK8S_KUBECONFIG"`
	// K8sNamespace is a namespace for worker-created Jobs.
	K8sNamespace string `env:"CODEXK8S_WORKER_K8S_NAMESPACE" envDefault:"codex-k8s-prod"`
	// ProductionNamespace is platform production namespace used by ai-repair pod runs.
	ProductionNamespace string `env:"CODEXK8S_PRODUCTION_NAMESPACE" envDefault:"codex-k8s-prod"`
	// JobImage is a container image used for spawned run Jobs.
	JobImage string `env:"CODEXK8S_WORKER_JOB_IMAGE" envDefault:"busybox:1.36"`
	// JobImageFallback is optional fallback image used when primary run image is missing in registry.
	JobImageFallback string `env:"CODEXK8S_WORKER_JOB_IMAGE_FALLBACK"`
	// JobCommand is a shell command executed by run Jobs.
	JobCommand string `env:"CODEXK8S_WORKER_JOB_COMMAND" envDefault:"/usr/local/bin/codex-k8s-agent-runner"`
	// JobTTLSeconds controls ttlSecondsAfterFinished for run Jobs.
	JobTTLSeconds int32 `env:"CODEXK8S_WORKER_JOB_TTL_SECONDS" envDefault:"600"`
	// JobBackoffLimit controls Job retry attempts.
	JobBackoffLimit int32 `env:"CODEXK8S_WORKER_JOB_BACKOFF_LIMIT" envDefault:"0"`
	// JobActiveDeadlineSeconds controls max run duration before termination.
	JobActiveDeadlineSeconds int64 `env:"CODEXK8S_WORKER_JOB_ACTIVE_DEADLINE_SECONDS" envDefault:"18000"`
	// RunNamespacePrefix defines prefix for full-env runtime namespaces.
	RunNamespacePrefix string `env:"CODEXK8S_WORKER_RUN_NAMESPACE_PREFIX" envDefault:"codex-issue"`
	// NamespaceLeaseSweepLimit limits managed namespaces inspected per tick for ttl-based cleanup.
	NamespaceLeaseSweepLimit int `env:"CODEXK8S_WORKER_NAMESPACE_LEASE_SWEEP_LIMIT" envDefault:"200"`
	// StateInReviewLabel is applied to PR when agent run is ready for owner review.
	StateInReviewLabel string `env:"CODEXK8S_STATE_IN_REVIEW_LABEL" envDefault:"state:in-review"`
	// RunServiceAccountName is service account for full-env run jobs.
	RunServiceAccountName string `env:"CODEXK8S_WORKER_RUN_SERVICE_ACCOUNT" envDefault:"codex-runner"`
	// RunRoleName is RBAC role name for full-env run jobs.
	RunRoleName string `env:"CODEXK8S_WORKER_RUN_ROLE_NAME" envDefault:"codex-runner"`
	// RunRoleBindingName is RBAC role binding name for full-env run jobs.
	RunRoleBindingName string `env:"CODEXK8S_WORKER_RUN_ROLE_BINDING_NAME" envDefault:"codex-runner"`
	// RunResourceQuotaName is ResourceQuota name in runtime namespaces.
	RunResourceQuotaName string `env:"CODEXK8S_WORKER_RUN_RESOURCE_QUOTA_NAME" envDefault:"codex-run-quota"`
	// RunLimitRangeName is LimitRange name in runtime namespaces.
	RunLimitRangeName string `env:"CODEXK8S_WORKER_RUN_LIMIT_RANGE_NAME" envDefault:"codex-run-limits"`
	// RunCredentialsSecretName is Secret name used for run pod credentials in runtime namespaces.
	RunCredentialsSecretName string `env:"CODEXK8S_WORKER_RUN_CREDENTIALS_SECRET_NAME" envDefault:"codex-run-credentials"`
	// RunResourceQuotaPods controls max pods per run namespace.
	RunResourceQuotaPods int64 `env:"CODEXK8S_WORKER_RUN_QUOTA_PODS" envDefault:"20"`
	// AIRepairNamespace overrides namespace for ai-repair runs; defaults to production namespace.
	AIRepairNamespace string `env:"CODEXK8S_WORKER_AI_REPAIR_NAMESPACE"`
	// AIRepairServiceAccount is service account used by ai-repair pod runs.
	AIRepairServiceAccount string `env:"CODEXK8S_WORKER_AI_REPAIR_SERVICE_ACCOUNT" envDefault:"codex-k8s-control-plane"`
	// InternalRegistryHost points to internal registry host:port used for deterministic image checks.
	InternalRegistryHost string `env:"CODEXK8S_INTERNAL_REGISTRY_HOST" envDefault:"codex-k8s-registry:5000"`
	// InternalRegistryScheme sets internal registry URL scheme.
	InternalRegistryScheme string `env:"CODEXK8S_INTERNAL_REGISTRY_SCHEME" envDefault:"http"`
	// JobImageCheckTimeout controls timeout for checking image availability in internal registry.
	JobImageCheckTimeout string `env:"CODEXK8S_WORKER_JOB_IMAGE_CHECK_TIMEOUT" envDefault:"10s"`
}

// LoadConfig parses and validates worker configuration from environment.
func LoadConfig() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse worker config from environment: %w", err)
	}

	if cfg.WorkerID == "" {
		hostname, hostErr := os.Hostname()
		if hostErr != nil || hostname == "" {
			cfg.WorkerID = "worker"
		} else {
			cfg.WorkerID = hostname
		}
	}

	return cfg, nil
}
