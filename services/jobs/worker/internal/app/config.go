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
	ClaimLimit int `env:"CODEXK8S_WORKER_CLAIM_LIMIT" envDefault:"2"`
	// RunningCheckLimit controls how many running runs are reconciled per tick.
	RunningCheckLimit int `env:"CODEXK8S_WORKER_RUNNING_CHECK_LIMIT" envDefault:"200"`
	// SlotsPerProject defines initial slot pool size per project.
	SlotsPerProject int `env:"CODEXK8S_WORKER_SLOTS_PER_PROJECT" envDefault:"2"`
	// SlotLeaseTTL controls for how long slot is leased before expiration.
	SlotLeaseTTL string `env:"CODEXK8S_WORKER_SLOT_LEASE_TTL" envDefault:"10m"`

	// LearningModeDefault controls default project learning-mode when worker auto-creates projects.
	// Keep empty value to disable by default; set to "true" to enable by default.
	LearningModeDefault string `env:"CODEXK8S_LEARNING_MODE_DEFAULT"`

	// ControlPlaneGRPCTarget is control-plane gRPC address used for internal worker calls.
	ControlPlaneGRPCTarget string `env:"CODEXK8S_CONTROL_PLANE_GRPC_TARGET,required,notEmpty"`
	// ControlPlaneMCPBaseURL is MCP HTTP endpoint passed into spawned run pods.
	ControlPlaneMCPBaseURL string `env:"CODEXK8S_CONTROL_PLANE_MCP_BASE_URL" envDefault:"http://codex-k8s-control-plane:8081/mcp"`
	// OpenAIAPIKey is injected into run pods for codex login.
	OpenAIAPIKey string `env:"CODEXK8S_OPENAI_API_KEY"`
	// OpenAIAuthFile stores optional Codex auth.json content for run pods.
	OpenAIAuthFile string `env:"CODEXK8S_OPENAI_AUTH_FILE"`
	// Context7APIKey enables Context7 documentation calls from run pods when set.
	Context7APIKey string `env:"CODEXK8S_CONTEXT7_API_KEY"`
	// GitBotToken is injected into run pods for git transport (fetch/push only).
	GitBotToken string `env:"CODEXK8S_GIT_BOT_TOKEN"`
	// GitBotUsername is GitHub username used with bot token for git transport auth.
	GitBotUsername string `env:"CODEXK8S_GIT_BOT_USERNAME" envDefault:"codex-bot"`
	// GitBotMail is git author email configured in run pods.
	GitBotMail string `env:"CODEXK8S_GIT_BOT_MAIL" envDefault:"codex-bot@codex-k8s.local"`
	// AgentDefaultModel is fallback model when issue labels do not override model.
	AgentDefaultModel string `env:"CODEXK8S_AGENT_DEFAULT_MODEL" envDefault:"gpt-5.3-codex"`
	// AgentDefaultReasoningEffort is fallback reasoning profile when issue labels do not override reasoning.
	AgentDefaultReasoningEffort string `env:"CODEXK8S_AGENT_DEFAULT_REASONING_EFFORT" envDefault:"high"`
	// AgentDefaultLocale is fallback prompt locale.
	AgentDefaultLocale string `env:"CODEXK8S_AGENT_DEFAULT_LOCALE" envDefault:"ru"`
	// AgentBaseBranch is default base branch for PR flow.
	AgentBaseBranch string `env:"CODEXK8S_AGENT_BASE_BRANCH" envDefault:"main"`

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
	K8sNamespace string `env:"CODEXK8S_WORKER_K8S_NAMESPACE" envDefault:"codex-k8s-ai-staging"`
	// JobImage is a container image used for spawned run Jobs.
	JobImage string `env:"CODEXK8S_WORKER_JOB_IMAGE" envDefault:"busybox:1.36"`
	// JobCommand is a shell command executed by run Jobs.
	JobCommand string `env:"CODEXK8S_WORKER_JOB_COMMAND" envDefault:"/usr/local/bin/codex-k8s-agent-runner"`
	// JobTTLSeconds controls ttlSecondsAfterFinished for run Jobs.
	JobTTLSeconds int32 `env:"CODEXK8S_WORKER_JOB_TTL_SECONDS" envDefault:"600"`
	// JobBackoffLimit controls Job retry attempts.
	JobBackoffLimit int32 `env:"CODEXK8S_WORKER_JOB_BACKOFF_LIMIT" envDefault:"0"`
	// JobActiveDeadlineSeconds controls max run duration before termination.
	JobActiveDeadlineSeconds int64 `env:"CODEXK8S_WORKER_JOB_ACTIVE_DEADLINE_SECONDS" envDefault:"900"`
	// RunNamespacePrefix defines prefix for full-env runtime namespaces.
	RunNamespacePrefix string `env:"CODEXK8S_WORKER_RUN_NAMESPACE_PREFIX" envDefault:"codex-issue"`
	// RunNamespaceCleanup enables namespace cleanup after run completion.
	RunNamespaceCleanup bool `env:"CODEXK8S_WORKER_RUN_NAMESPACE_CLEANUP" envDefault:"true"`
	// RunDebugLabel keeps full-env namespace for post-run debugging when present on issue labels.
	RunDebugLabel string `env:"CODEXK8S_RUN_DEBUG_LABEL" envDefault:"run:debug"`
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
	// RunResourceRequestsCPU controls requests.cpu hard quota.
	RunResourceRequestsCPU string `env:"CODEXK8S_WORKER_RUN_QUOTA_REQUESTS_CPU" envDefault:"4"`
	// RunResourceRequestsMemory controls requests.memory hard quota.
	RunResourceRequestsMemory string `env:"CODEXK8S_WORKER_RUN_QUOTA_REQUESTS_MEMORY" envDefault:"8Gi"`
	// RunResourceLimitsCPU controls limits.cpu hard quota.
	RunResourceLimitsCPU string `env:"CODEXK8S_WORKER_RUN_QUOTA_LIMITS_CPU" envDefault:"8"`
	// RunResourceLimitsMemory controls limits.memory hard quota.
	RunResourceLimitsMemory string `env:"CODEXK8S_WORKER_RUN_QUOTA_LIMITS_MEMORY" envDefault:"16Gi"`
	// RunDefaultRequestCPU controls default CPU request via LimitRange.
	RunDefaultRequestCPU string `env:"CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_REQUEST_CPU" envDefault:"250m"`
	// RunDefaultRequestMemory controls default memory request via LimitRange.
	RunDefaultRequestMemory string `env:"CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_REQUEST_MEMORY" envDefault:"256Mi"`
	// RunDefaultLimitCPU controls default CPU limit via LimitRange.
	RunDefaultLimitCPU string `env:"CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_CPU" envDefault:"1"`
	// RunDefaultLimitMemory controls default memory limit via LimitRange.
	RunDefaultLimitMemory string `env:"CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_MEMORY" envDefault:"1Gi"`
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
