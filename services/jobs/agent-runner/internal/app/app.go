package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	cpclient "github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/controlplane"
	"github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/runner"
)

// ExitError keeps process exit code for top-level main.
type ExitError = runner.ExitError

// AsExitError checks if error wraps runner exit error.
func AsExitError(err error) (ExitError, bool) {
	var target ExitError
	if errors.As(err, &target) {
		return target, true
	}
	return ExitError{}, false
}

// Run starts and executes one runner job lifecycle.
func Run() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	appCtx := context.Background()

	dialCtx, cancel := context.WithTimeout(appCtx, 30*time.Second)
	defer cancel()
	cp, err := cpclient.Dial(dialCtx, cfg.ControlPlaneGRPCTarget, cfg.MCPBearerToken)
	if err != nil {
		return fmt.Errorf("dial control-plane callback client: %w", err)
	}
	defer func() { _ = cp.Close() }()

	runnerService := runner.NewService(runner.Config{
		RunID:                  cfg.RunID,
		CorrelationID:          cfg.CorrelationID,
		ProjectID:              cfg.ProjectID,
		RepositoryFullName:     cfg.RepositoryFullName,
		AgentKey:               cfg.AgentKey,
		IssueNumber:            cfg.IssueNumber,
		TriggerKind:            cfg.TriggerKind,
		PromptTemplateKind:     cfg.PromptTemplateKind,
		PromptTemplateSource:   cfg.PromptTemplateSource,
		PromptTemplateLocale:   cfg.PromptTemplateLocale,
		AgentModel:             cfg.AgentModel,
		AgentReasoningEffort:   cfg.AgentReasoningEffort,
		AgentBaseBranch:        cfg.AgentBaseBranch,
		AgentDisplayName:       cfg.AgentDisplayName,
		ControlPlaneGRPCTarget: cfg.ControlPlaneGRPCTarget,
		MCPBaseURL:             cfg.MCPBaseURL,
		MCPBearerToken:         cfg.MCPBearerToken,
		GitBotToken:            cfg.GitBotToken,
		GitBotUsername:         cfg.GitBotUsername,
		GitBotMail:             cfg.GitBotMail,
		OpenAIAPIKey:           cfg.OpenAIAPIKey,
	}, cp, logger)

	if err := runnerService.Run(appCtx); err != nil {
		return err
	}
	return nil
}
