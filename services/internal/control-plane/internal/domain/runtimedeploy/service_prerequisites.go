package runtimedeploy

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/codex-k8s/kodex/libs/go/servicescfg"
)

const (
	defaultLetsEncryptDirectoryURL           = "https://acme-v02.api.letsencrypt.org/directory"
	defaultTelegramInteractionAdapterBaseURL = "http://kodex-telegram-interaction-adapter:8080"
	defaultTelegramInteractionAdapterTimeout = "10s"
	telegramInteractionAdapterSecretBytes    = 32
)

type telegramRuntimeSecretValues struct {
	BaseURL               string
	BearerToken           string
	WebhookSecret         string
	Timeout               string
	BotToken              string
	ChatID                string
	RecipientBindingsJSON string
}

func (s *Service) ensureCodexK8sPrerequisites(ctx context.Context, repositoryRoot string, namespace string, vars map[string]string, stack *servicescfg.Stack, runID string) error {
	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace == "" {
		return fmt.Errorf("namespace is required")
	}

	if err := s.k8s.EnsureNamespace(ctx, targetNamespace); err != nil {
		return fmt.Errorf("ensure namespace %s: %w", targetNamespace, err)
	}

	secretResolver := servicescfg.NewSecretResolver(stack)
	targetEnv := secretResolver.CanonicalEnvironment(valueOr(vars, "KODEX_ENV", ""))
	if targetEnv == "" {
		targetEnv = "production"
	}

	platformNamespace := strings.TrimSpace(valueOr(vars, "KODEX_PLATFORM_NAMESPACE", ""))
	if platformNamespace == "" {
		platformNamespace = strings.TrimSpace(valueOr(vars, "KODEX_PRODUCTION_NAMESPACE", ""))
	}
	if platformNamespace == "" {
		platformNamespace = targetNamespace
	}

	existingPostgres, _, err := s.k8s.GetSecretData(ctx, targetNamespace, "kodex-postgres")
	if err != nil {
		return fmt.Errorf("load kodex-postgres secret: %w", err)
	}
	existingRuntime, _, err := s.k8s.GetSecretData(ctx, targetNamespace, "kodex-runtime")
	if err != nil {
		return fmt.Errorf("load kodex-runtime secret: %w", err)
	}
	existingOAuth, _, err := s.k8s.GetSecretData(ctx, targetNamespace, "kodex-oauth2-proxy")
	if err != nil {
		return fmt.Errorf("load oauth2-proxy secret: %w", err)
	}

	platformRuntime, _, err := s.k8s.GetSecretData(ctx, platformNamespace, "kodex-runtime")
	if err != nil {
		return fmt.Errorf("load platform kodex-runtime secret: %w", err)
	}
	platformOAuth, _, err := s.k8s.GetSecretData(ctx, platformNamespace, "kodex-oauth2-proxy")
	if err != nil {
		return fmt.Errorf("load platform kodex-oauth2-proxy secret: %w", err)
	}

	sharedRuntime := platformRuntime
	sharedOAuth := platformOAuth
	if targetEnv == "ai" {
		aiRuntime, found, err := s.k8s.GetSecretData(ctx, platformNamespace, "kodex-runtime-ai")
		if err != nil {
			return fmt.Errorf("load platform kodex-runtime-ai secret: %w", err)
		}
		if found {
			sharedRuntime = aiRuntime
		} else if len(platformRuntime) > 0 {
			if err := s.k8s.UpsertSecret(ctx, platformNamespace, "kodex-runtime-ai", cloneSecretData(platformRuntime)); err != nil {
				return fmt.Errorf("seed platform kodex-runtime-ai secret: %w", err)
			}
		}

		aiOAuth, found, err := s.k8s.GetSecretData(ctx, platformNamespace, "kodex-oauth2-proxy-ai")
		if err != nil {
			return fmt.Errorf("load platform kodex-oauth2-proxy-ai secret: %w", err)
		}
		if found {
			sharedOAuth = aiOAuth
		} else if len(platformOAuth) > 0 {
			if err := s.k8s.UpsertSecret(ctx, platformNamespace, "kodex-oauth2-proxy-ai", cloneSecretData(platformOAuth)); err != nil {
				return fmt.Errorf("seed platform kodex-oauth2-proxy-ai secret: %w", err)
			}
		}
	}

	oauthClientID, oauthClientIDSource, err := requiredValueFromSecrets(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_GITHUB_OAUTH_CLIENT_ID")
	if err != nil {
		return err
	}
	oauthClientSecret, oauthClientSecretSource, err := requiredValueFromSecrets(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_GITHUB_OAUTH_CLIENT_SECRET")
	if err != nil {
		return err
	}
	if strings.TrimSpace(runID) != "" {
		s.appendTaskLogBestEffort(ctx, runID, "prerequisites", "info", "Resolved KODEX_GITHUB_OAUTH_CLIENT_ID via "+oauthClientIDSource)
		s.appendTaskLogBestEffort(ctx, runID, "prerequisites", "info", "Resolved KODEX_GITHUB_OAUTH_CLIENT_SECRET via "+oauthClientSecretSource)
	}

	internalRegistryService := valueOrExisting(vars, existingRuntime, "KODEX_INTERNAL_REGISTRY_SERVICE", "kodex-registry")
	internalRegistryPort := valueOrExisting(vars, existingRuntime, "KODEX_INTERNAL_REGISTRY_PORT", "5000")
	internalRegistryHost := valueOrExisting(vars, existingRuntime, "KODEX_INTERNAL_REGISTRY_HOST", "127.0.0.1:"+internalRegistryPort)
	internalRegistryStorageSize := valueOrExisting(vars, existingRuntime, "KODEX_INTERNAL_REGISTRY_STORAGE_SIZE", "20Gi")

	k8sAPICIDR := strings.TrimSpace(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_K8S_API_CIDR", ""))
	k8sAPIPort := strings.TrimSpace(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_K8S_API_PORT", "6443"))
	if k8sAPIPort == "" {
		k8sAPIPort = "6443"
	}
	// Network policies are applied from templates using vars; carry resolved values from secrets back into vars.
	vars["KODEX_K8S_API_CIDR"] = k8sAPICIDR
	vars["KODEX_K8S_API_PORT"] = k8sAPIPort

	networkPolicyBaseline := strings.TrimSpace(valueOr(vars, "KODEX_NETWORK_POLICY_BASELINE", "true"))
	if targetEnv != "ai" && strings.EqualFold(networkPolicyBaseline, "true") && k8sAPICIDR == "" {
		return fmt.Errorf("KODEX_K8S_API_CIDR is required when network policy baseline is enabled; set it to node IP /32 or disable KODEX_NETWORK_POLICY_BASELINE")
	}

	postgresDB := valueOrExisting(vars, existingPostgres, "KODEX_POSTGRES_DB", "kodex")
	postgresUser := valueOrExisting(vars, existingPostgres, "KODEX_POSTGRES_USER", "kodex")
	postgresPassword, err := valueOrExistingOrRandomHex(vars, existingPostgres, "KODEX_POSTGRES_PASSWORD", 24)
	if err != nil {
		return fmt.Errorf("resolve KODEX_POSTGRES_PASSWORD: %w", err)
	}

	appSecretKey, err := valueOrExistingOrSharedOrGenerated(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_APP_SECRET_KEY", func() (string, error) {
		return randomHex(32)
	})
	if err != nil {
		return fmt.Errorf("resolve KODEX_APP_SECRET_KEY: %w", err)
	}
	tokenEncryptionKey, err := valueOrExistingOrSharedOrGenerated(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_TOKEN_ENCRYPTION_KEY", func() (string, error) {
		return randomHex(32)
	})
	if err != nil {
		return fmt.Errorf("resolve KODEX_TOKEN_ENCRYPTION_KEY: %w", err)
	}
	mcpTokenSigningKey := strings.TrimSpace(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_MCP_TOKEN_SIGNING_KEY", ""))
	if mcpTokenSigningKey == "" {
		mcpTokenSigningKey = tokenEncryptionKey
	}
	githubWebhookSecret, err := valueOrExistingOrSharedOrGenerated(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_GITHUB_WEBHOOK_SECRET", func() (string, error) {
		return randomHex(32)
	})
	if err != nil {
		return fmt.Errorf("resolve KODEX_GITHUB_WEBHOOK_SECRET: %w", err)
	}
	jwtSigningKey, err := valueOrExistingOrSharedOrGenerated(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_JWT_SIGNING_KEY", func() (string, error) {
		return randomHex(32)
	})
	if err != nil {
		return fmt.Errorf("resolve KODEX_JWT_SIGNING_KEY: %w", err)
	}
	bootstrapOwnerEmail := valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_BOOTSTRAP_OWNER_EMAIL", "owner@example.invalid")
	letsEncryptEmailFallback := strings.TrimSpace(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_BOOTSTRAP_OWNER_EMAIL", ""))
	letsEncryptEmail := strings.TrimSpace(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_LETSENCRYPT_EMAIL", letsEncryptEmailFallback))
	letsEncryptServer := strings.TrimSpace(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_LETSENCRYPT_SERVER", defaultLetsEncryptDirectoryURL))
	telegramRuntime, err := resolveTelegramRuntimeSecretValues(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime)
	if err != nil {
		return err
	}

	// TLS preparation runs in the same reconcile cycle; keep resolved values in template vars.
	vars["KODEX_LETSENCRYPT_EMAIL"] = letsEncryptEmail
	vars["KODEX_LETSENCRYPT_SERVER"] = letsEncryptServer

	postgresData := map[string][]byte{
		"KODEX_POSTGRES_DB":       []byte(postgresDB),
		"KODEX_POSTGRES_USER":     []byte(postgresUser),
		"KODEX_POSTGRES_PASSWORD": []byte(postgresPassword),
	}
	if err := s.k8s.UpsertSecret(ctx, targetNamespace, "kodex-postgres", postgresData); err != nil {
		return fmt.Errorf("upsert kodex-postgres secret: %w", err)
	}

	runtimeSecret := map[string][]byte{
		"KODEX_INTERNAL_REGISTRY_SERVICE":      []byte(internalRegistryService),
		"KODEX_INTERNAL_REGISTRY_PORT":         []byte(internalRegistryPort),
		"KODEX_INTERNAL_REGISTRY_HOST":         []byte(internalRegistryHost),
		"KODEX_INTERNAL_REGISTRY_STORAGE_SIZE": []byte(internalRegistryStorageSize),
		"KODEX_KANIKO_MAX_PARALLEL":            []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_KANIKO_MAX_PARALLEL", "8")),
		"KODEX_IMAGE_MIRROR_MAX_PARALLEL":      []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_IMAGE_MIRROR_MAX_PARALLEL", "8")),
		"KODEX_RUNTIME_DEPLOY_WORKERS_PER_POD": []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_RUNTIME_DEPLOY_WORKERS_PER_POD", "4")),
		"KODEX_K8S_API_CIDR":                   []byte(k8sAPICIDR),
		"KODEX_K8S_API_PORT":                   []byte(k8sAPIPort),
		"KODEX_GITHUB_PAT":                     []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_GITHUB_PAT", "")),
		"KODEX_GITHUB_REPO":                    []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_GITHUB_REPO", "")),
		"KODEX_FIRST_PROJECT_GITHUB_REPO":      []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_FIRST_PROJECT_GITHUB_REPO", "")),
		"KODEX_OPENAI_API_KEY":                 []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_OPENAI_API_KEY", "")),
		// Project DB admin settings are tied to the Postgres instance deployed into the target namespace.
		// Do not inherit them from platform (production) env/secrets, otherwise ai slots break with auth mismatch.
		"KODEX_PROJECT_DB_ADMIN_HOST":                                []byte("postgres"),
		"KODEX_PROJECT_DB_ADMIN_PORT":                                []byte("5432"),
		"KODEX_PROJECT_DB_ADMIN_USER":                                []byte(postgresUser),
		"KODEX_PROJECT_DB_ADMIN_PASSWORD":                            []byte(postgresPassword),
		"KODEX_PROJECT_DB_ADMIN_SSLMODE":                             []byte("disable"),
		"KODEX_PROJECT_DB_ADMIN_DATABASE":                            []byte("postgres"),
		"KODEX_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS":                    []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS", "dev,production,prod")),
		"KODEX_GIT_BOT_TOKEN":                                        []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_GIT_BOT_TOKEN", "")),
		"KODEX_GIT_BOT_USERNAME":                                     []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_GIT_BOT_USERNAME", "codex-bot")),
		"KODEX_GIT_BOT_MAIL":                                         []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_GIT_BOT_MAIL", "codex-bot@kodex.local")),
		"KODEX_TELEGRAM_BOT_TOKEN":                                   []byte(telegramRuntime.BotToken),
		"KODEX_TELEGRAM_CHAT_ID":                                     []byte(telegramRuntime.ChatID),
		"KODEX_TELEGRAM_INTERACTION_ADAPTER_BASE_URL":                []byte(telegramRuntime.BaseURL),
		"KODEX_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN":            []byte(telegramRuntime.BearerToken),
		"KODEX_TELEGRAM_INTERACTION_ADAPTER_WEBHOOK_SECRET":          []byte(telegramRuntime.WebhookSecret),
		"KODEX_TELEGRAM_INTERACTION_ADAPTER_TIMEOUT":                 []byte(telegramRuntime.Timeout),
		"KODEX_TELEGRAM_INTERACTION_ADAPTER_RECIPIENT_BINDINGS_JSON": []byte(telegramRuntime.RecipientBindingsJSON),
		"KODEX_CONTEXT7_API_KEY":                                     []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_CONTEXT7_API_KEY", "")),
		"KODEX_APP_SECRET_KEY":                                       []byte(appSecretKey),
		"KODEX_TOKEN_ENCRYPTION_KEY":                                 []byte(tokenEncryptionKey),
		"KODEX_MCP_TOKEN_SIGNING_KEY":                                []byte(mcpTokenSigningKey),
		"KODEX_MCP_TOKEN_TTL":                                        []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_MCP_TOKEN_TTL", "24h")),
		"KODEX_RUN_AGENT_LOGS_RETENTION_DAYS":                        []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_RUN_AGENT_LOGS_RETENTION_DAYS", "14")),
		"KODEX_LEARNING_MODE_DEFAULT":                                []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_LEARNING_MODE_DEFAULT", "true")),
		"KODEX_GITHUB_WEBHOOK_SECRET":                                []byte(githubWebhookSecret),
		"KODEX_GITHUB_WEBHOOK_URL":                                   []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_GITHUB_WEBHOOK_URL", "")),
		"KODEX_GITHUB_WEBHOOK_EVENTS":                                []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_GITHUB_WEBHOOK_EVENTS", "push,pull_request,issues,issue_comment,pull_request_review,pull_request_review_comment")),
		"KODEX_PRODUCTION_DOMAIN":                                    []byte(valueOr(vars, "KODEX_PRODUCTION_DOMAIN", "")),
		"KODEX_AI_DOMAIN":                                            []byte(valueOr(vars, "KODEX_AI_DOMAIN", "")),
		"KODEX_PUBLIC_BASE_URL":                                      []byte(valueOr(vars, "KODEX_PUBLIC_BASE_URL", "https://example.invalid")),
		"KODEX_INTERACTION_CALLBACK_BASE_URL":                        []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_INTERACTION_CALLBACK_BASE_URL", "http://kodex")),
		"KODEX_BOOTSTRAP_OWNER_EMAIL":                                []byte(bootstrapOwnerEmail),
		"KODEX_LETSENCRYPT_EMAIL":                                    []byte(letsEncryptEmail),
		"KODEX_LETSENCRYPT_SERVER":                                   []byte(letsEncryptServer),
		"KODEX_BOOTSTRAP_ALLOWED_EMAILS":                             []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_BOOTSTRAP_ALLOWED_EMAILS", "")),
		"KODEX_BOOTSTRAP_PLATFORM_ADMIN_EMAILS":                      []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_BOOTSTRAP_PLATFORM_ADMIN_EMAILS", "")),
		"KODEX_GITHUB_OAUTH_CLIENT_ID":                               []byte(oauthClientID),
		"KODEX_GITHUB_OAUTH_CLIENT_SECRET":                           []byte(oauthClientSecret),
		"KODEX_JWT_SIGNING_KEY":                                      []byte(jwtSigningKey),
		"KODEX_JWT_TTL":                                              []byte(valueOrExistingOrShared(secretResolver, targetEnv, vars, existingRuntime, sharedRuntime, "KODEX_JWT_TTL", "15m")),
		// UI hot-reload is only supported in ai slots. For production/production we
		// intentionally keep the value empty so api-gateway serves embedded static UI.
		"KODEX_VITE_DEV_UPSTREAM": []byte(resolveViteDevUpstream(vars)),
	}
	if err := s.k8s.UpsertSecret(ctx, targetNamespace, "kodex-runtime", runtimeSecret); err != nil {
		return fmt.Errorf("upsert kodex-runtime secret: %w", err)
	}

	oauthCookie, _, _ := valueForEnvironment(secretResolver, targetEnv, vars, "OAUTH2_PROXY_COOKIE_SECRET")
	oauthCookie = strings.TrimSpace(oauthCookie)
	if oauthCookie == "" {
		if value, ok := existingOAuth["OAUTH2_PROXY_COOKIE_SECRET"]; ok {
			existingCookie := string(value)
			if isValidOAuthCookieSecret(existingCookie) {
				oauthCookie = existingCookie
			}
		}
	}
	if oauthCookie == "" && sharedOAuth != nil {
		if value, ok := sharedOAuth["OAUTH2_PROXY_COOKIE_SECRET"]; ok {
			sharedCookie := string(value)
			if isValidOAuthCookieSecret(sharedCookie) {
				oauthCookie = sharedCookie
			}
		}
	}
	if oauthCookie == "" {
		oauthCookie = strings.TrimSpace(valueOr(vars, "OAUTH2_PROXY_COOKIE_SECRET", ""))
	}
	if oauthCookie == "" {
		oauthCookie, err = randomHex(16)
		if err != nil {
			return fmt.Errorf("generate oauth2-proxy cookie secret: %w", err)
		}
	}
	oauthSecret := map[string][]byte{
		"OAUTH2_PROXY_CLIENT_ID":     []byte(oauthClientID),
		"OAUTH2_PROXY_CLIENT_SECRET": []byte(oauthClientSecret),
		"OAUTH2_PROXY_COOKIE_SECRET": []byte(oauthCookie),
	}
	if err := s.k8s.UpsertSecret(ctx, targetNamespace, "kodex-oauth2-proxy", oauthSecret); err != nil {
		return fmt.Errorf("upsert kodex-oauth2-proxy secret: %w", err)
	}

	// Seed separate environment-scoped platform secrets (ai) once so operators can diverge them later.
	if targetNamespace == platformNamespace && targetEnv != "ai" {
		if err := s.seedPlatformEnvSecrets(ctx, platformNamespace, vars, runtimeSecret, oauthSecret); err != nil {
			return err
		}
	}

	labels := make(map[string]string, len(labelCatalogDefaults))
	for key, fallback := range labelCatalogDefaults {
		labels[key] = valueOr(vars, key, fallback)
	}
	if err := s.k8s.UpsertConfigMap(ctx, targetNamespace, "kodex-label-catalog", labels); err != nil {
		return fmt.Errorf("upsert kodex-label-catalog configmap: %w", err)
	}

	repoRoot := strings.TrimSpace(repositoryRoot)
	if repoRoot == "" {
		repoRoot = s.cfg.RepositoryRoot
	}
	migrationsData, err := readMigrationFiles(filepath.Join(repoRoot, "services/internal/control-plane/cmd/cli/migrations"))
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	if len(migrationsData) > 0 {
		if err := s.k8s.UpsertConfigMap(ctx, targetNamespace, "kodex-migrations", migrationsData); err != nil {
			return fmt.Errorf("upsert kodex-migrations configmap: %w", err)
		}
	}

	return nil
}

func resolveViteDevUpstream(vars map[string]string) string {
	// Prefer explicit config env when present (control-plane sets it).
	targetEnv := strings.ToLower(strings.TrimSpace(valueOr(vars, "KODEX_SERVICES_CONFIG_ENV", "")))
	if targetEnv == "" {
		targetEnv = strings.ToLower(strings.TrimSpace(valueOr(vars, "KODEX_ENV", "")))
	}

	// Keep production/prod prod-like by default. Hot-reload is only for ai slots.
	if targetEnv != "ai" {
		return ""
	}

	return strings.TrimSpace(valueOr(vars, "KODEX_VITE_DEV_UPSTREAM", "http://kodex-web-console:5173"))
}

func requiredValueFromSecrets(resolver servicescfg.SecretResolver, envName string, values map[string]string, existing map[string][]byte, shared map[string][]byte, key string) (string, string, error) {
	value, source := valueOrExistingOrSharedWithSource(resolver, envName, values, existing, shared, key, "")
	if strings.TrimSpace(value) == "" {
		return "", "", fmt.Errorf("%s is required", key)
	}
	return value, source, nil
}

func valueOrExisting(values map[string]string, existing map[string][]byte, key string, fallback string) string {
	if value := strings.TrimSpace(valueOr(values, key, "")); value != "" {
		return value
	}
	if existing != nil {
		if raw, ok := existing[key]; ok {
			if value := strings.TrimSpace(string(raw)); value != "" {
				return value
			}
		}
	}
	return fallback
}

func valueOrExistingOrShared(resolver servicescfg.SecretResolver, envName string, values map[string]string, existing map[string][]byte, shared map[string][]byte, key string, fallback string) string {
	value, _ := valueOrExistingOrSharedWithSource(resolver, envName, values, existing, shared, key, fallback)
	return value
}

func valueOrExistingOrSharedWithSource(resolver servicescfg.SecretResolver, envName string, values map[string]string, existing map[string][]byte, shared map[string][]byte, key string, fallback string) (string, string) {
	if overrideValue, source, ok := valueForEnvironment(resolver, envName, values, key); ok {
		return overrideValue, "env_override:" + source
	}
	if existing != nil {
		if raw, ok := existing[key]; ok {
			if value := strings.TrimSpace(string(raw)); value != "" {
				return value, "secret:" + key
			}
		}
	}
	if shared != nil {
		if raw, ok := shared[key]; ok {
			if value := strings.TrimSpace(string(raw)); value != "" {
				return value, "shared_secret:" + key
			}
		}
	}
	if value := strings.TrimSpace(valueOr(values, key, "")); value != "" {
		return value, "env:" + key
	}
	if strings.TrimSpace(fallback) == "" {
		return "", ""
	}
	return fallback, "fallback"
}

func valueOrExistingOrRandomHex(values map[string]string, existing map[string][]byte, key string, numBytes int) (string, error) {
	if value := strings.TrimSpace(valueOr(values, key, "")); value != "" {
		if existing != nil {
			if raw, ok := existing[key]; ok {
				existingValue := strings.TrimSpace(string(raw))
				if existingValue != "" && value != existingValue {
					return "", fmt.Errorf("%s differs from existing secret value; refusing to rotate automatically", key)
				}
			}
		}
		return value, nil
	}
	if existing != nil {
		if raw, ok := existing[key]; ok {
			if value := strings.TrimSpace(string(raw)); value != "" {
				return value, nil
			}
		}
	}
	return randomHex(numBytes)
}

func valueOrExistingOrSharedOrGenerated(resolver servicescfg.SecretResolver, envName string, values map[string]string, existing map[string][]byte, shared map[string][]byte, key string, generate func() (string, error)) (string, error) {
	if overrideValue, _, ok := valueForEnvironment(resolver, envName, values, key); ok {
		if err := ensureNoSecretConflict(existing, key, overrideValue, "existing"); err != nil {
			return "", err
		}
		if err := ensureNoSecretConflict(shared, key, overrideValue, "shared"); err != nil {
			return "", err
		}
		return overrideValue, nil
	}
	if existing != nil {
		if raw, ok := existing[key]; ok {
			if value := strings.TrimSpace(string(raw)); value != "" {
				return value, nil
			}
		}
	}
	if shared != nil {
		if raw, ok := shared[key]; ok {
			if value := strings.TrimSpace(string(raw)); value != "" {
				return value, nil
			}
		}
	}
	if value := strings.TrimSpace(valueOr(values, key, "")); value != "" {
		return value, nil
	}
	return generate()
}

func resolveTelegramRuntimeSecretValues(resolver servicescfg.SecretResolver, envName string, values map[string]string, existing map[string][]byte, shared map[string][]byte) (telegramRuntimeSecretValues, error) {
	bearerToken, err := valueOrExistingOrSharedOrGenerated(resolver, envName, values, existing, shared, "KODEX_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN", func() (string, error) {
		return randomBase64URL(telegramInteractionAdapterSecretBytes)
	})
	if err != nil {
		return telegramRuntimeSecretValues{}, fmt.Errorf("resolve KODEX_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN: %w", err)
	}
	webhookSecret, err := valueOrExistingOrSharedOrGenerated(resolver, envName, values, existing, shared, "KODEX_TELEGRAM_INTERACTION_ADAPTER_WEBHOOK_SECRET", func() (string, error) {
		return randomBase64URL(telegramInteractionAdapterSecretBytes)
	})
	if err != nil {
		return telegramRuntimeSecretValues{}, fmt.Errorf("resolve KODEX_TELEGRAM_INTERACTION_ADAPTER_WEBHOOK_SECRET: %w", err)
	}

	botToken := strings.TrimSpace(valueOrExistingOrShared(resolver, envName, values, existing, shared, "KODEX_TELEGRAM_BOT_TOKEN", ""))
	chatID := strings.TrimSpace(valueOrExistingOrShared(resolver, envName, values, existing, shared, "KODEX_TELEGRAM_CHAT_ID", ""))
	recipientBindingsJSON := strings.TrimSpace(valueOrExistingOrShared(resolver, envName, values, existing, shared, "KODEX_TELEGRAM_INTERACTION_ADAPTER_RECIPIENT_BINDINGS_JSON", ""))
	baseURL := strings.TrimSpace(valueOrExistingOrShared(resolver, envName, values, existing, shared, "KODEX_TELEGRAM_INTERACTION_ADAPTER_BASE_URL", ""))
	if baseURL == "" && (botToken != "" || recipientBindingsJSON != "") {
		baseURL = defaultTelegramInteractionAdapterBaseURL
	}

	timeout := strings.TrimSpace(valueOrExistingOrShared(resolver, envName, values, existing, shared, "KODEX_TELEGRAM_INTERACTION_ADAPTER_TIMEOUT", defaultTelegramInteractionAdapterTimeout))
	if timeout == "" {
		timeout = defaultTelegramInteractionAdapterTimeout
	}

	return telegramRuntimeSecretValues{
		BaseURL:               baseURL,
		BearerToken:           bearerToken,
		WebhookSecret:         webhookSecret,
		Timeout:               timeout,
		BotToken:              botToken,
		ChatID:                chatID,
		RecipientBindingsJSON: recipientBindingsJSON,
	}, nil
}

func valueForEnvironment(resolver servicescfg.SecretResolver, envName string, values map[string]string, key string) (string, string, bool) {
	overrideKey, ok := resolver.ResolveOverrideKey(envName, key)
	if !ok {
		return "", "", false
	}
	overrideValue := strings.TrimSpace(values[overrideKey])
	if overrideValue == "" {
		return "", "", false
	}
	return overrideValue, overrideKey, true
}

func ensureNoSecretConflict(secret map[string][]byte, key string, value string, source string) error {
	if secret == nil {
		return nil
	}
	raw, ok := secret[key]
	if !ok {
		return nil
	}
	existing := strings.TrimSpace(string(raw))
	if existing == "" || existing == value {
		return nil
	}
	return fmt.Errorf("%s differs from %s secret value; refusing to rotate automatically", key, source)
}

func cloneSecretData(input map[string][]byte) map[string][]byte {
	if len(input) == 0 {
		return map[string][]byte{}
	}
	out := make(map[string][]byte, len(input))
	for key, value := range input {
		out[key] = append([]byte(nil), value...)
	}
	return out
}

func (s *Service) seedPlatformEnvSecrets(ctx context.Context, platformNamespace string, vars map[string]string, runtimeSecret map[string][]byte, oauthSecret map[string][]byte) error {
	platformNamespace = strings.TrimSpace(platformNamespace)
	if platformNamespace == "" {
		return nil
	}

	if _, found, err := s.k8s.GetSecretData(ctx, platformNamespace, "kodex-runtime-ai"); err != nil {
		return fmt.Errorf("check kodex-runtime-ai exists: %w", err)
	} else if !found {
		aiRuntime := cloneSecretData(runtimeSecret)
		aiVars := cloneStringMap(vars)
		aiVars["KODEX_SERVICES_CONFIG_ENV"] = "ai"
		aiVars["KODEX_ENV"] = "ai"
		aiRuntime["KODEX_VITE_DEV_UPSTREAM"] = []byte(resolveViteDevUpstream(aiVars))
		if err := s.k8s.UpsertSecret(ctx, platformNamespace, "kodex-runtime-ai", aiRuntime); err != nil {
			return fmt.Errorf("seed kodex-runtime-ai secret: %w", err)
		}
	}

	if _, found, err := s.k8s.GetSecretData(ctx, platformNamespace, "kodex-oauth2-proxy-ai"); err != nil {
		return fmt.Errorf("check kodex-oauth2-proxy-ai exists: %w", err)
	} else if !found {
		if err := s.k8s.UpsertSecret(ctx, platformNamespace, "kodex-oauth2-proxy-ai", cloneSecretData(oauthSecret)); err != nil {
			return fmt.Errorf("seed kodex-oauth2-proxy-ai secret: %w", err)
		}
	}

	return nil
}
func randomHex(numBytes int) (string, error) {
	if numBytes <= 0 {
		numBytes = 16
	}
	raw := make([]byte, numBytes)
	if _, err := cryptorand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func randomBase64URL(numBytes int) (string, error) {
	if numBytes <= 0 {
		numBytes = 32
	}
	raw := make([]byte, numBytes)
	if _, err := cryptorand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func isValidOAuthCookieSecret(value string) bool {
	size := len(strings.TrimSpace(value))
	return size == 16 || size == 24 || size == 32
}

func readMigrationFiles(dir string) (map[string]string, error) {
	items, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	files := make([]string, 0, len(items))
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		name := strings.TrimSpace(item.Name())
		if !strings.HasSuffix(strings.ToLower(name), ".sql") {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)

	data := make(map[string]string, len(files))
	for _, name := range files {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		data[name] = string(raw)
	}
	return data, nil
}
