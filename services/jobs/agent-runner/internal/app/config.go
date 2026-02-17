package app

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

// Config defines environment-backed runtime settings for agent-runner job.
type Config struct {
	RunID              string `env:"CODEXK8S_RUN_ID,required,notEmpty"`
	CorrelationID      string `env:"CODEXK8S_CORRELATION_ID,required,notEmpty"`
	ProjectID          string `env:"CODEXK8S_PROJECT_ID"`
	RepositoryFullName string `env:"CODEXK8S_REPOSITORY_FULL_NAME,required,notEmpty"`
	AgentKey           string `env:"CODEXK8S_AGENT_KEY,required,notEmpty"`
	IssueNumber        int64  `env:"CODEXK8S_ISSUE_NUMBER"`
	RunTargetBranch    string `env:"CODEXK8S_RUN_TARGET_BRANCH"`
	ExistingPRNumber   int    `env:"CODEXK8S_EXISTING_PR_NUMBER"`
	RuntimeMode        string `env:"CODEXK8S_RUNTIME_MODE" envDefault:"code-only"`

	ControlPlaneGRPCTarget string `env:"CODEXK8S_CONTROL_PLANE_GRPC_TARGET,required,notEmpty"`
	MCPBaseURL             string `env:"CODEXK8S_MCP_BASE_URL,required,notEmpty"`
	MCPBearerToken         string `env:"CODEXK8S_MCP_BEARER_TOKEN,required,notEmpty"`

	TriggerKind          string `env:"CODEXK8S_RUN_TRIGGER_KIND" envDefault:"dev"`
	TriggerLabel         string `env:"CODEXK8S_RUN_TRIGGER_LABEL"`
	PromptTemplateKind   string `env:"CODEXK8S_PROMPT_TEMPLATE_KIND" envDefault:"work"`
	PromptTemplateSource string `env:"CODEXK8S_PROMPT_TEMPLATE_SOURCE" envDefault:"repo_seed"`
	PromptTemplateLocale string `env:"CODEXK8S_PROMPT_TEMPLATE_LOCALE" envDefault:"ru"`
	StateInReviewLabel   string `env:"CODEXK8S_STATE_IN_REVIEW_LABEL" envDefault:"state:in-review"`
	AgentModel           string `env:"CODEXK8S_AGENT_MODEL"`
	AgentReasoningEffort string `env:"CODEXK8S_AGENT_REASONING_EFFORT" envDefault:"xhigh"`
	AgentBaseBranch      string `env:"CODEXK8S_AGENT_BASE_BRANCH" envDefault:"main"`
	AgentDisplayName     string `env:"CODEXK8S_AGENT_DISPLAY_NAME,required,notEmpty"`

	GitBotToken    string `env:"CODEXK8S_GIT_BOT_TOKEN,required,notEmpty"`
	GitBotUsername string `env:"CODEXK8S_GIT_BOT_USERNAME,required,notEmpty"`
	GitBotMail     string `env:"CODEXK8S_GIT_BOT_MAIL,required,notEmpty"`
	OpenAIAPIKey   string `env:"CODEXK8S_OPENAI_API_KEY"`
}

// LoadConfig parses and validates configuration from environment.
func LoadConfig() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse agent-runner config from environment: %w", err)
	}

	cfg.TriggerKind = normalizeTriggerKind(cfg.TriggerKind)
	cfg.TriggerLabel = strings.TrimSpace(cfg.TriggerLabel)
	if cfg.TriggerLabel == "" {
		cfg.TriggerLabel = webhookdomain.DefaultTriggerLabel(webhookdomain.NormalizeTriggerKind(cfg.TriggerKind))
	}
	cfg.PromptTemplateKind = strings.TrimSpace(strings.ToLower(cfg.PromptTemplateKind))
	if isReviseTriggerKind(cfg.TriggerKind) {
		cfg.PromptTemplateKind = promptTemplateKindReview
	}
	if cfg.PromptTemplateKind != promptTemplateKindReview {
		cfg.PromptTemplateKind = promptTemplateKindWork
	}

	cfg.PromptTemplateSource = strings.TrimSpace(cfg.PromptTemplateSource)
	if cfg.PromptTemplateSource == "" {
		cfg.PromptTemplateSource = promptTemplateSourceSeed
	}
	cfg.PromptTemplateLocale = strings.TrimSpace(cfg.PromptTemplateLocale)
	if cfg.PromptTemplateLocale == "" {
		cfg.PromptTemplateLocale = "ru"
	}
	cfg.StateInReviewLabel = strings.TrimSpace(cfg.StateInReviewLabel)
	if cfg.StateInReviewLabel == "" {
		cfg.StateInReviewLabel = stateInReviewLabelDefault
	}

	cfg.AgentModel = strings.TrimSpace(cfg.AgentModel)
	if cfg.AgentModel == "" {
		cfg.AgentModel = modelGPT52Codex
	}
	cfg.AgentReasoningEffort = strings.TrimSpace(strings.ToLower(cfg.AgentReasoningEffort))
	// Codex CLI expects "xhigh" for the highest reasoning effort.
	switch cfg.AgentReasoningEffort {
	case "extra-high", "extra_high", "extra high", "x-high":
		cfg.AgentReasoningEffort = "xhigh"
	}
	if cfg.AgentReasoningEffort == "" {
		cfg.AgentReasoningEffort = "xhigh"
	}
	cfg.AgentBaseBranch = strings.TrimSpace(cfg.AgentBaseBranch)
	if cfg.AgentBaseBranch == "" {
		cfg.AgentBaseBranch = "main"
	}
	cfg.RuntimeMode = strings.TrimSpace(strings.ToLower(cfg.RuntimeMode))
	if cfg.RuntimeMode != runtimeModeFullEnv {
		cfg.RuntimeMode = runtimeModeCodeOnly
	}

	cfg.ProjectID = strings.TrimSpace(cfg.ProjectID)
	cfg.ControlPlaneGRPCTarget = strings.TrimSpace(cfg.ControlPlaneGRPCTarget)
	cfg.MCPBaseURL = strings.TrimRight(strings.TrimSpace(cfg.MCPBaseURL), "/")
	cfg.MCPBearerToken = strings.TrimSpace(cfg.MCPBearerToken)
	cfg.RepositoryFullName = strings.TrimSpace(cfg.RepositoryFullName)
	cfg.AgentKey = strings.TrimSpace(cfg.AgentKey)
	cfg.RunTargetBranch = strings.TrimSpace(cfg.RunTargetBranch)
	if cfg.ExistingPRNumber < 0 {
		cfg.ExistingPRNumber = 0
	}
	cfg.AgentDisplayName = strings.TrimSpace(cfg.AgentDisplayName)
	cfg.GitBotUsername = strings.TrimSpace(cfg.GitBotUsername)
	cfg.GitBotMail = strings.TrimSpace(cfg.GitBotMail)
	cfg.OpenAIAPIKey = strings.TrimSpace(cfg.OpenAIAPIKey)

	return cfg, nil
}
