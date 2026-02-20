package app

import (
	"log/slog"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
)

const defaultNamespaceLeaseTTL = 24 * time.Hour

func loadWebhookRuntimeNamespaceTTLPolicy(cfg Config, logger *slog.Logger) (time.Duration, map[string]time.Duration) {
	if logger == nil {
		logger = slog.Default()
	}

	path := strings.TrimSpace(cfg.ServicesConfigPath)
	if path == "" {
		return defaultNamespaceLeaseTTL, map[string]time.Duration{}
	}

	loaded, err := servicescfg.Load(path, servicescfg.LoadOptions{Env: cfg.ServicesConfigEnv})
	if err != nil {
		logger.Warn("skip services.yaml namespace ttl policy: load failed", "path", path, "env", cfg.ServicesConfigEnv, "err", err)
		return defaultNamespaceLeaseTTL, map[string]time.Duration{}
	}

	defaultTTL := defaultNamespaceLeaseTTL
	if raw := strings.TrimSpace(loaded.Stack.Spec.WebhookRuntime.DefaultNamespaceTTL); raw != "" {
		parsed, parseErr := time.ParseDuration(raw)
		if parseErr != nil || parsed <= 0 {
			logger.Warn("ignore invalid default namespace ttl from services.yaml", "value", raw, "err", parseErr)
		} else {
			defaultTTL = parsed
		}
	}

	ttlByRole := make(map[string]time.Duration, len(loaded.Stack.Spec.WebhookRuntime.NamespaceTTLByRole))
	for rawRole, rawTTL := range loaded.Stack.Spec.WebhookRuntime.NamespaceTTLByRole {
		role := strings.ToLower(strings.TrimSpace(rawRole))
		if role == "" {
			continue
		}

		ttlValue := strings.TrimSpace(rawTTL)
		if ttlValue == "" {
			continue
		}

		ttl, parseErr := time.ParseDuration(ttlValue)
		if parseErr != nil || ttl <= 0 {
			logger.Warn("ignore invalid namespace ttl by role from services.yaml", "role", role, "value", ttlValue, "err", parseErr)
			continue
		}
		ttlByRole[role] = ttl
	}

	return defaultTTL, ttlByRole
}
