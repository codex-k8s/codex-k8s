package app

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config defines environment-backed runtime settings for agent-runner job.
type Config struct {
	RunID              string `env:"CODEXK8S_RUN_ID,required,notEmpty"`
	CorrelationID      string `env:"CODEXK8S_CORRELATION_ID,required,notEmpty"`
	ProjectID          string `env:"CODEXK8S_PROJECT_ID"`
	RepositoryFullName string `env:"CODEXK8S_REPOSITORY_FULL_NAME,required,notEmpty"`
	AgentKey           string `env:"CODEXK8S_AGENT_KEY,required,notEmpty"`
	IssueNumber        int64  `env:"CODEXK8S_ISSUE_NUMBER"`

	ControlPlaneGRPCTarget string `env:"CODEXK8S_CONTROL_PLANE_GRPC_TARGET,required,notEmpty"`
	MCPBaseURL             string `env:"CODEXK8S_MCP_BASE_URL,required,notEmpty"`
	MCPBearerToken         string `env:"CODEXK8S_MCP_BEARER_TOKEN,required,notEmpty"`

	TriggerKind        string `env:"CODEXK8S_RUN_TRIGGER_KIND" envDefault:"dev"`
	PromptTemplateKind string `env:"CODEXK8S_PROMPT_TEMPLATE_KIND" envDefault:"work"`
	PromptTemplateSource string `env:"CODEXK8S_PROMPT_TEMPLATE_SOURCE" envDefault:"repo_seed"`
	PromptTemplateLocale string `env:"CODEXK8S_PROMPT_TEMPLATE_LOCALE" envDefault:"ru"`
	AgentModel           string `env:"CODEXK8S_AGENT_MODEL" envDefault:"gpt-5-codex"`
	AgentReasoningEffort string `env:"CODEXK8S_AGENT_REASONING_EFFORT" envDefault:"high"`
	AgentBaseBranch      string `env:"CODEXK8S_AGENT_BASE_BRANCH" envDefault:"main"`
	AgentDisplayName     string `env:"CODEXK8S_AGENT_DISPLAY_NAME,required,notEmpty"`

	GitBotToken    string `env:"CODEXK8S_GIT_BOT_TOKEN,required,notEmpty"`
	GitBotUsername string `env:"CODEXK8S_GIT_BOT_USERNAME,required,notEmpty"`
	GitBotMail     string `env:"CODEXK8S_GIT_BOT_MAIL,required,notEmpty"`
	OpenAIAPIKey   string `env:"CODEXK8S_OPENAI_API_KEY,required,notEmpty"`
}

// LoadConfig parses and validates configuration from environment.
func LoadConfig() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse agent-runner config from environment: %w", err)
	}

	cfg.TriggerKind = normalizeTriggerKind(cfg.TriggerKind)
	cfg.PromptTemplateKind = strings.TrimSpace(strings.ToLower(cfg.PromptTemplateKind))
	if cfg.TriggerKind == triggerKindDevRevise {
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
	cfg.AgentModel = strings.TrimSpace(cfg.AgentModel)
	if cfg.AgentModel == "" {
		cfg.AgentModel = "gpt-5-codex"
	}
	cfg.AgentReasoningEffort = strings.TrimSpace(strings.ToLower(cfg.AgentReasoningEffort))
	if cfg.AgentReasoningEffort == "" {
		cfg.AgentReasoningEffort = "high"
	}
	cfg.AgentBaseBranch = strings.TrimSpace(cfg.AgentBaseBranch)
	if cfg.AgentBaseBranch == "" {
		cfg.AgentBaseBranch = "main"
	}

	cfg.ProjectID = strings.TrimSpace(cfg.ProjectID)
	cfg.ControlPlaneGRPCTarget = strings.TrimSpace(cfg.ControlPlaneGRPCTarget)
	cfg.MCPBaseURL = strings.TrimRight(strings.TrimSpace(cfg.MCPBaseURL), "/")
	cfg.MCPBearerToken = strings.TrimSpace(cfg.MCPBearerToken)
	cfg.RepositoryFullName = strings.TrimSpace(cfg.RepositoryFullName)
	cfg.AgentKey = strings.TrimSpace(cfg.AgentKey)
	cfg.AgentDisplayName = strings.TrimSpace(cfg.AgentDisplayName)
	cfg.GitBotUsername = strings.TrimSpace(cfg.GitBotUsername)
	cfg.GitBotMail = strings.TrimSpace(cfg.GitBotMail)

	return cfg, nil
}
