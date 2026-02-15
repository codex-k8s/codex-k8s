package runtimedeploy

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (s *Service) ensureCodexK8sPrerequisites(ctx context.Context, namespace string, vars map[string]string) error {
	existingPostgres, _, err := s.k8s.GetSecretData(ctx, namespace, "codex-k8s-postgres")
	if err != nil {
		return fmt.Errorf("load codex-k8s-postgres secret: %w", err)
	}
	existingRuntime, _, err := s.k8s.GetSecretData(ctx, namespace, "codex-k8s-runtime")
	if err != nil {
		return fmt.Errorf("load codex-k8s-runtime secret: %w", err)
	}

	oauthClientID, err := requiredNonEmptyValue(vars, "CODEXK8S_GITHUB_OAUTH_CLIENT_ID")
	if err != nil {
		return err
	}
	oauthClientSecret, err := requiredNonEmptyValue(vars, "CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET")
	if err != nil {
		return err
	}

	internalRegistryService := valueOrExisting(vars, existingRuntime, "CODEXK8S_INTERNAL_REGISTRY_SERVICE", "codex-k8s-registry")
	internalRegistryPort := valueOrExisting(vars, existingRuntime, "CODEXK8S_INTERNAL_REGISTRY_PORT", "5000")
	internalRegistryHost := valueOrExisting(vars, existingRuntime, "CODEXK8S_INTERNAL_REGISTRY_HOST", "127.0.0.1:"+internalRegistryPort)
	internalRegistryStorageSize := valueOrExisting(vars, existingRuntime, "CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE", "20Gi")

	postgresDB := valueOrExisting(vars, existingPostgres, "CODEXK8S_POSTGRES_DB", "codex_k8s")
	postgresUser := valueOrExisting(vars, existingPostgres, "CODEXK8S_POSTGRES_USER", "codex_k8s")
	postgresPassword, err := valueOrExistingOrRandomHex(vars, existingPostgres, "CODEXK8S_POSTGRES_PASSWORD", 24)
	if err != nil {
		return fmt.Errorf("resolve CODEXK8S_POSTGRES_PASSWORD: %w", err)
	}

	appSecretKey, err := valueOrExistingOrRandomHex(vars, existingRuntime, "CODEXK8S_APP_SECRET_KEY", 32)
	if err != nil {
		return fmt.Errorf("resolve CODEXK8S_APP_SECRET_KEY: %w", err)
	}
	tokenEncryptionKey, err := valueOrExistingOrRandomHex(vars, existingRuntime, "CODEXK8S_TOKEN_ENCRYPTION_KEY", 32)
	if err != nil {
		return fmt.Errorf("resolve CODEXK8S_TOKEN_ENCRYPTION_KEY: %w", err)
	}
	mcpTokenSigningKey := strings.TrimSpace(valueOrExisting(vars, existingRuntime, "CODEXK8S_MCP_TOKEN_SIGNING_KEY", ""))
	if mcpTokenSigningKey == "" {
		mcpTokenSigningKey = tokenEncryptionKey
	}
	githubWebhookSecret, err := valueOrExistingOrRandomHex(vars, existingRuntime, "CODEXK8S_GITHUB_WEBHOOK_SECRET", 32)
	if err != nil {
		return fmt.Errorf("resolve CODEXK8S_GITHUB_WEBHOOK_SECRET: %w", err)
	}
	jwtSigningKey, err := valueOrExistingOrRandomHex(vars, existingRuntime, "CODEXK8S_JWT_SIGNING_KEY", 32)
	if err != nil {
		return fmt.Errorf("resolve CODEXK8S_JWT_SIGNING_KEY: %w", err)
	}

	postgresData := map[string][]byte{
		"CODEXK8S_POSTGRES_DB":       []byte(postgresDB),
		"CODEXK8S_POSTGRES_USER":     []byte(postgresUser),
		"CODEXK8S_POSTGRES_PASSWORD": []byte(postgresPassword),
	}
	if err := s.k8s.UpsertSecret(ctx, namespace, "codex-k8s-postgres", postgresData); err != nil {
		return fmt.Errorf("upsert codex-k8s-postgres secret: %w", err)
	}

	runtimeSecret := map[string][]byte{
		"CODEXK8S_INTERNAL_REGISTRY_SERVICE":         []byte(internalRegistryService),
		"CODEXK8S_INTERNAL_REGISTRY_PORT":            []byte(internalRegistryPort),
		"CODEXK8S_INTERNAL_REGISTRY_HOST":            []byte(internalRegistryHost),
		"CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE":    []byte(internalRegistryStorageSize),
		"CODEXK8S_GITHUB_PAT":                        []byte(valueOr(vars, "CODEXK8S_GITHUB_PAT", "")),
		"CODEXK8S_GITHUB_REPO":                       []byte(valueOr(vars, "CODEXK8S_GITHUB_REPO", "")),
		"CODEXK8S_FIRST_PROJECT_GITHUB_REPO":         []byte(valueOr(vars, "CODEXK8S_FIRST_PROJECT_GITHUB_REPO", "")),
		"CODEXK8S_OPENAI_API_KEY":                    []byte(valueOr(vars, "CODEXK8S_OPENAI_API_KEY", "")),
		"CODEXK8S_OPENAI_AUTH_FILE":                  []byte(valueOr(vars, "CODEXK8S_OPENAI_AUTH_FILE", "")),
		"CODEXK8S_PROJECT_DB_ADMIN_HOST":             []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_ADMIN_HOST", "postgres")),
		"CODEXK8S_PROJECT_DB_ADMIN_PORT":             []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_ADMIN_PORT", "5432")),
		"CODEXK8S_PROJECT_DB_ADMIN_USER":             []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_ADMIN_USER", valueOr(vars, "CODEXK8S_POSTGRES_USER", "codex_k8s"))),
		"CODEXK8S_PROJECT_DB_ADMIN_PASSWORD":         []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_ADMIN_PASSWORD", string(postgresData["CODEXK8S_POSTGRES_PASSWORD"]))),
		"CODEXK8S_PROJECT_DB_ADMIN_SSLMODE":          []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_ADMIN_SSLMODE", "disable")),
		"CODEXK8S_PROJECT_DB_ADMIN_DATABASE":         []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_ADMIN_DATABASE", "postgres")),
		"CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS": []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS", "dev,staging,ai-staging,prod")),
		"CODEXK8S_GIT_BOT_TOKEN":                     []byte(valueOr(vars, "CODEXK8S_GIT_BOT_TOKEN", "")),
		"CODEXK8S_GIT_BOT_USERNAME":                  []byte(valueOr(vars, "CODEXK8S_GIT_BOT_USERNAME", "codex-bot")),
		"CODEXK8S_GIT_BOT_MAIL":                      []byte(valueOr(vars, "CODEXK8S_GIT_BOT_MAIL", "codex-bot@codex-k8s.local")),
		"CODEXK8S_CONTEXT7_API_KEY":                  []byte(valueOr(vars, "CODEXK8S_CONTEXT7_API_KEY", "")),
		"CODEXK8S_APP_SECRET_KEY":                    []byte(appSecretKey),
		"CODEXK8S_TOKEN_ENCRYPTION_KEY":              []byte(tokenEncryptionKey),
		"CODEXK8S_MCP_TOKEN_SIGNING_KEY":             []byte(mcpTokenSigningKey),
		"CODEXK8S_MCP_TOKEN_TTL":                     []byte(valueOr(vars, "CODEXK8S_MCP_TOKEN_TTL", "24h")),
		"CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS":     []byte(valueOr(vars, "CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS", "14")),
		"CODEXK8S_LEARNING_MODE_DEFAULT":             []byte(valueOr(vars, "CODEXK8S_LEARNING_MODE_DEFAULT", "true")),
		"CODEXK8S_GITHUB_WEBHOOK_SECRET":             []byte(githubWebhookSecret),
		"CODEXK8S_GITHUB_WEBHOOK_URL":                []byte(valueOr(vars, "CODEXK8S_GITHUB_WEBHOOK_URL", "")),
		"CODEXK8S_GITHUB_WEBHOOK_EVENTS":             []byte(valueOr(vars, "CODEXK8S_GITHUB_WEBHOOK_EVENTS", "push,pull_request,issues,issue_comment,pull_request_review,pull_request_review_comment")),
		"CODEXK8S_PUBLIC_BASE_URL":                   []byte(valueOr(vars, "CODEXK8S_PUBLIC_BASE_URL", "https://example.invalid")),
		"CODEXK8S_BOOTSTRAP_OWNER_EMAIL":             []byte(valueOr(vars, "CODEXK8S_BOOTSTRAP_OWNER_EMAIL", "owner@example.invalid")),
		"CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS":          []byte(valueOr(vars, "CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS", "")),
		"CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS":   []byte(valueOr(vars, "CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS", "")),
		"CODEXK8S_GITHUB_OAUTH_CLIENT_ID":            []byte(oauthClientID),
		"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET":        []byte(oauthClientSecret),
		"CODEXK8S_JWT_SIGNING_KEY":                   []byte(jwtSigningKey),
		"CODEXK8S_JWT_TTL":                           []byte(valueOr(vars, "CODEXK8S_JWT_TTL", "15m")),
		// UI hot-reload is only supported in ai slots. For ai-staging/production we
		// intentionally keep the value empty so api-gateway serves embedded static UI.
		"CODEXK8S_VITE_DEV_UPSTREAM": []byte(resolveViteDevUpstream(vars)),
	}
	if err := s.k8s.UpsertSecret(ctx, namespace, "codex-k8s-runtime", runtimeSecret); err != nil {
		return fmt.Errorf("upsert codex-k8s-runtime secret: %w", err)
	}

	oauthCookie := valueOr(vars, "OAUTH2_PROXY_COOKIE_SECRET", "")
	if oauthCookie == "" {
		existing, found, err := s.k8s.GetSecretData(ctx, namespace, "codex-k8s-oauth2-proxy")
		if err != nil {
			return fmt.Errorf("load oauth2-proxy secret: %w", err)
		}
		if found {
			if value, ok := existing["OAUTH2_PROXY_COOKIE_SECRET"]; ok {
				existingCookie := string(value)
				if isValidOAuthCookieSecret(existingCookie) {
					oauthCookie = existingCookie
				}
			}
		}
		if oauthCookie == "" {
			oauthCookie, err = randomHex(16)
			if err != nil {
				return fmt.Errorf("generate oauth2-proxy cookie secret: %w", err)
			}
		}
	}
	oauthSecret := map[string][]byte{
		"OAUTH2_PROXY_CLIENT_ID":     []byte(oauthClientID),
		"OAUTH2_PROXY_CLIENT_SECRET": []byte(oauthClientSecret),
		"OAUTH2_PROXY_COOKIE_SECRET": []byte(oauthCookie),
	}
	if err := s.k8s.UpsertSecret(ctx, namespace, "codex-k8s-oauth2-proxy", oauthSecret); err != nil {
		return fmt.Errorf("upsert codex-k8s-oauth2-proxy secret: %w", err)
	}

	labels := make(map[string]string, len(labelCatalogDefaults))
	for key, fallback := range labelCatalogDefaults {
		labels[key] = valueOr(vars, key, fallback)
	}
	if err := s.k8s.UpsertConfigMap(ctx, namespace, "codex-k8s-label-catalog", labels); err != nil {
		return fmt.Errorf("upsert codex-k8s-label-catalog configmap: %w", err)
	}

	migrationsData, err := readMigrationFiles(filepath.Join(s.cfg.RepositoryRoot, "services/internal/control-plane/cmd/cli/migrations"))
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	if len(migrationsData) > 0 {
		if err := s.k8s.UpsertConfigMap(ctx, namespace, "codex-k8s-migrations", migrationsData); err != nil {
			return fmt.Errorf("upsert codex-k8s-migrations configmap: %w", err)
		}
	}

	return nil
}

func resolveViteDevUpstream(vars map[string]string) string {
	// Prefer explicit config env when present (control-plane sets it).
	targetEnv := strings.ToLower(strings.TrimSpace(valueOr(vars, "CODEXK8S_SERVICES_CONFIG_ENV", "")))
	if targetEnv == "" {
		targetEnv = strings.ToLower(strings.TrimSpace(valueOr(vars, "CODEXK8S_ENV", "")))
	}

	// Keep ai-staging/prod prod-like by default. Hot-reload is only for ai slots.
	if targetEnv != "ai" {
		return ""
	}

	return strings.TrimSpace(valueOr(vars, "CODEXK8S_VITE_DEV_UPSTREAM", "http://codex-k8s-web-console:5173"))
}

func requiredNonEmptyValue(values map[string]string, key string) (string, error) {
	value := strings.TrimSpace(valueOr(values, key, ""))
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func valueOrRandomHex(values map[string]string, key string, numBytes int) (string, error) {
	if value := strings.TrimSpace(valueOr(values, key, "")); value != "" {
		return value, nil
	}
	return randomHex(numBytes)
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
