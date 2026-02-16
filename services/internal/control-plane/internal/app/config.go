package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config defines environment-backed runtime settings for control-plane.
type Config struct {
	// GRPCAddr is the bind address for the gRPC server.
	GRPCAddr string `env:"CODEXK8S_CONTROL_PLANE_GRPC_ADDR" envDefault:":9090"`
	// HTTPAddr is the bind address for the HTTP health/metrics server.
	HTTPAddr string `env:"CODEXK8S_CONTROL_PLANE_HTTP_ADDR" envDefault:":8081"`
	// KubeconfigPath is optional kubeconfig path for local development.
	KubeconfigPath string `env:"CODEXK8S_KUBECONFIG"`

	// PublicBaseURL is used to build default webhook URL when CODEXK8S_GITHUB_WEBHOOK_URL is empty.
	PublicBaseURL string `env:"CODEXK8S_PUBLIC_BASE_URL,required,notEmpty"`

	// BootstrapOwnerEmail is the first allowed email for staff access (platform admin).
	BootstrapOwnerEmail          string   `env:"CODEXK8S_BOOTSTRAP_OWNER_EMAIL,required,notEmpty"`
	BootstrapAllowedEmails       []string `env:"CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS"`
	BootstrapPlatformAdminEmails []string `env:"CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS"`

	// LearningModeDefault controls the default for newly created projects.
	// Empty string means "false".
	LearningModeDefault string `env:"CODEXK8S_LEARNING_MODE_DEFAULT" envDefault:"false"`

	// GitHubWebhookSecret is used when attaching repository hooks (staff operations).
	GitHubWebhookSecret  string   `env:"CODEXK8S_GITHUB_WEBHOOK_SECRET,required,notEmpty"`
	GitHubWebhookURL     string   `env:"CODEXK8S_GITHUB_WEBHOOK_URL"`
	GitHubWebhookEvents  []string `env:"CODEXK8S_GITHUB_WEBHOOK_EVENTS" envDefault:"push,pull_request,issues,issue_comment,pull_request_review,pull_request_review_comment"`
	RunIntakeLabel       string   `env:"CODEXK8S_RUN_INTAKE_LABEL" envDefault:"run:intake"`
	RunIntakeReviseLabel string   `env:"CODEXK8S_RUN_INTAKE_REVISE_LABEL" envDefault:"run:intake:revise"`
	RunVisionLabel       string   `env:"CODEXK8S_RUN_VISION_LABEL" envDefault:"run:vision"`
	RunVisionReviseLabel string   `env:"CODEXK8S_RUN_VISION_REVISE_LABEL" envDefault:"run:vision:revise"`
	RunPRDLabel          string   `env:"CODEXK8S_RUN_PRD_LABEL" envDefault:"run:prd"`
	RunPRDReviseLabel    string   `env:"CODEXK8S_RUN_PRD_REVISE_LABEL" envDefault:"run:prd:revise"`
	RunArchLabel         string   `env:"CODEXK8S_RUN_ARCH_LABEL" envDefault:"run:arch"`
	RunArchReviseLabel   string   `env:"CODEXK8S_RUN_ARCH_REVISE_LABEL" envDefault:"run:arch:revise"`
	RunDesignLabel       string   `env:"CODEXK8S_RUN_DESIGN_LABEL" envDefault:"run:design"`
	RunDesignReviseLabel string   `env:"CODEXK8S_RUN_DESIGN_REVISE_LABEL" envDefault:"run:design:revise"`
	RunPlanLabel         string   `env:"CODEXK8S_RUN_PLAN_LABEL" envDefault:"run:plan"`
	RunPlanReviseLabel   string   `env:"CODEXK8S_RUN_PLAN_REVISE_LABEL" envDefault:"run:plan:revise"`
	RunDevLabel          string   `env:"CODEXK8S_RUN_DEV_LABEL" envDefault:"run:dev"`
	RunDevReviseLabel    string   `env:"CODEXK8S_RUN_DEV_REVISE_LABEL" envDefault:"run:dev:revise"`
	RunDocAuditLabel     string   `env:"CODEXK8S_RUN_DOC_AUDIT_LABEL" envDefault:"run:doc-audit"`
	RunQALabel           string   `env:"CODEXK8S_RUN_QA_LABEL" envDefault:"run:qa"`
	RunReleaseLabel      string   `env:"CODEXK8S_RUN_RELEASE_LABEL" envDefault:"run:release"`
	RunPostDeployLabel   string   `env:"CODEXK8S_RUN_POSTDEPLOY_LABEL" envDefault:"run:postdeploy"`
	RunOpsLabel          string   `env:"CODEXK8S_RUN_OPS_LABEL" envDefault:"run:ops"`
	RunSelfImproveLabel  string   `env:"CODEXK8S_RUN_SELF_IMPROVE_LABEL" envDefault:"run:self-improve"`
	RunRethinkLabel      string   `env:"CODEXK8S_RUN_RETHINK_LABEL" envDefault:"run:rethink"`
	// ServicesConfigPath points to services.yaml used for webhook runtime policy.
	ServicesConfigPath string `env:"CODEXK8S_SERVICES_CONFIG_PATH" envDefault:"services.yaml"`
	// ServicesConfigEnv selects environment context when rendering services.yaml.
	ServicesConfigEnv string `env:"CODEXK8S_SERVICES_CONFIG_ENV" envDefault:"production"`
	// RepositoryRoot points to repository root used for services.yaml manifests and build templates.
	RepositoryRoot string `env:"CODEXK8S_REPOSITORY_ROOT" envDefault:"."`
	// RuntimeDeployRolloutTimeout controls readiness wait timeout for applied workloads.
	RuntimeDeployRolloutTimeout string `env:"CODEXK8S_RUNTIME_DEPLOY_ROLLOUT_TIMEOUT" envDefault:"20m"`
	// RuntimeDeployKanikoTimeout controls timeout for kaniko build jobs.
	RuntimeDeployKanikoTimeout string `env:"CODEXK8S_RUNTIME_DEPLOY_KANIKO_TIMEOUT" envDefault:"30m"`
	// RuntimeDeployWaitPollInterval controls polling interval for waiting on deploy task completion.
	RuntimeDeployWaitPollInterval string `env:"CODEXK8S_RUNTIME_DEPLOY_WAIT_POLL_INTERVAL" envDefault:"2s"`
	// RuntimeDeployReconcileInterval controls background deploy reconciler tick interval.
	RuntimeDeployReconcileInterval string `env:"CODEXK8S_RUNTIME_DEPLOY_RECONCILE_INTERVAL" envDefault:"3s"`
	// RuntimeDeployLeaseTTL controls deploy task lease duration for reconciler lock.
	RuntimeDeployLeaseTTL string `env:"CODEXK8S_RUNTIME_DEPLOY_LEASE_TTL" envDefault:"10m"`
	// RuntimeDeployWorkerID identifies current deploy reconciler instance.
	RuntimeDeployWorkerID string `env:"CODEXK8S_RUNTIME_DEPLOY_WORKER_ID"`
	// RuntimeDeployFieldManager is a server-side apply field manager name.
	RuntimeDeployFieldManager string `env:"CODEXK8S_RUNTIME_DEPLOY_FIELD_MANAGER" envDefault:"codex-k8s-control-plane"`
	// InternalRegistryHost points to internal registry host:port for image management APIs.
	InternalRegistryHost string `env:"CODEXK8S_INTERNAL_REGISTRY_HOST" envDefault:"codex-k8s-registry:5000"`
	// InternalRegistryScheme sets registry URL scheme.
	InternalRegistryScheme string `env:"CODEXK8S_INTERNAL_REGISTRY_SCHEME" envDefault:"http"`
	// RegistryHTTPTimeout controls timeout for internal registry API calls.
	RegistryHTTPTimeout string `env:"CODEXK8S_REGISTRY_HTTP_TIMEOUT" envDefault:"15s"`
	// RegistryCleanupKeepTags controls default keep policy for registry cleanup.
	RegistryCleanupKeepTags int `env:"CODEXK8S_REGISTRY_CLEANUP_KEEP_TAGS" envDefault:"5"`
	// GitHubPAT is platform-scoped GitHub token used for repository/project management paths.
	GitHubPAT string `env:"CODEXK8S_GITHUB_PAT"`
	// GitHubRepo is the platform repository (owner/name) used for bootstrap seeding and webhook-driven dogfooding.
	GitHubRepo string `env:"CODEXK8S_GITHUB_REPO,required,notEmpty"`
	// FirstProjectGitHubRepo is an optional initial project repository (owner/name) to seed into DB.
	FirstProjectGitHubRepo string `env:"CODEXK8S_FIRST_PROJECT_GITHUB_REPO"`
	// GitBotToken is runtime GitHub bot token used for comments/labels and run messaging paths.
	GitBotToken string `env:"CODEXK8S_GIT_BOT_TOKEN"`

	// TokenEncryptionKey is used to encrypt/decrypt repository tokens stored in DB.
	TokenEncryptionKey string `env:"CODEXK8S_TOKEN_ENCRYPTION_KEY,required,notEmpty"`
	// MCPTokenSigningKey is used to sign short-lived MCP bearer tokens.
	// If empty, TokenEncryptionKey is used as fallback.
	MCPTokenSigningKey string `env:"CODEXK8S_MCP_TOKEN_SIGNING_KEY"`
	// MCPTokenTTL defines default TTL for run-bound MCP tokens.
	MCPTokenTTL string `env:"CODEXK8S_MCP_TOKEN_TTL" envDefault:"24h"`
	// ControlPlaneMCPBaseURL is effective MCP endpoint included in prompt context and run env.
	ControlPlaneMCPBaseURL string `env:"CODEXK8S_CONTROL_PLANE_MCP_BASE_URL" envDefault:"http://codex-k8s-control-plane:8081/mcp"`
	// RunAgentLogsRetentionDays controls how long run-scoped agent logs are retained in agent_runs.
	RunAgentLogsRetentionDays int `env:"CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS" envDefault:"14"`

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

	// ProjectDBAdminHost is PostgreSQL admin host used by MCP database lifecycle tool.
	ProjectDBAdminHost string `env:"CODEXK8S_PROJECT_DB_ADMIN_HOST,required,notEmpty"`
	// ProjectDBAdminPort is PostgreSQL admin port used by MCP database lifecycle tool.
	ProjectDBAdminPort int `env:"CODEXK8S_PROJECT_DB_ADMIN_PORT" envDefault:"5432"`
	// ProjectDBAdminUser is PostgreSQL superuser/login used by MCP database lifecycle tool.
	ProjectDBAdminUser string `env:"CODEXK8S_PROJECT_DB_ADMIN_USER,required,notEmpty"`
	// ProjectDBAdminPassword is PostgreSQL superuser password used by MCP database lifecycle tool.
	ProjectDBAdminPassword string `env:"CODEXK8S_PROJECT_DB_ADMIN_PASSWORD,required,notEmpty"`
	// ProjectDBAdminSSLMode is PostgreSQL SSL mode for admin connection.
	ProjectDBAdminSSLMode string `env:"CODEXK8S_PROJECT_DB_ADMIN_SSLMODE" envDefault:"disable"`
	// ProjectDBAdminDatabase is admin database name for lifecycle connection.
	ProjectDBAdminDatabase string `env:"CODEXK8S_PROJECT_DB_ADMIN_DATABASE" envDefault:"postgres"`
	// ProjectDBLifecycleAllowedEnvs contains allowed environment names for MCP database lifecycle tool.
	ProjectDBLifecycleAllowedEnvs []string `env:"CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS" envDefault:"dev,production,prod"`
}

func (c Config) LearningModeDefaultBool() (bool, error) {
	if strings.TrimSpace(c.LearningModeDefault) == "" {
		return false, nil
	}
	v, err := strconv.ParseBool(c.LearningModeDefault)
	if err != nil {
		return false, fmt.Errorf("parse CODEXK8S_LEARNING_MODE_DEFAULT=%q: %w", c.LearningModeDefault, err)
	}
	return v, nil
}

// LoadConfig parses and validates configuration from environment variables.
func LoadConfig() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse app config from environment: %w", err)
	}
	return cfg, nil
}
