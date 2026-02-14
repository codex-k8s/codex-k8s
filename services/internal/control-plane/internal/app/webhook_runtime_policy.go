package app

import (
	"log/slog"
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/webhook"
)

func loadWebhookRuntimeModePolicy(cfg Config, logger *slog.Logger) webhook.RuntimeModePolicy {
	path := strings.TrimSpace(cfg.ServicesConfigPath)
	if path == "" {
		return webhook.DefaultRuntimeModePolicy()
	}

	loaded, err := servicescfg.Load(path, servicescfg.LoadOptions{Env: cfg.ServicesConfigEnv})
	if err != nil {
		logger.Warn("skip services.yaml runtime policy: load failed", "path", path, "env", cfg.ServicesConfigEnv, "err", err)
		return webhook.DefaultRuntimeModePolicy()
	}

	policy := webhook.DefaultRuntimeModePolicy()
	policy.Configured = true
	policy.Source = path
	policy.DefaultMode = toAgentRuntimeMode(loaded.Stack.Spec.WebhookRuntime.DefaultMode)
	policy.TriggerModes = make(map[webhookdomain.TriggerKind]agentdomain.RuntimeMode, len(loaded.Stack.Spec.WebhookRuntime.TriggerModes))
	for rawTrigger, mode := range loaded.Stack.Spec.WebhookRuntime.TriggerModes {
		trigger := webhookdomain.NormalizeTriggerKind(rawTrigger)
		policy.TriggerModes[trigger] = toAgentRuntimeMode(mode)
	}
	return policy
}

func toAgentRuntimeMode(mode servicescfg.RuntimeMode) agentdomain.RuntimeMode {
	return agentdomain.ParseRuntimeMode(string(mode))
}
