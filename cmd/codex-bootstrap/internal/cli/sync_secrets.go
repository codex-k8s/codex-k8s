package cli

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/cmd/codex-bootstrap/internal/envfile"
	"github.com/codex-k8s/codex-k8s/libs/go/k8s/clientcfg"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const defaultSyncSecretsTimeout = 3 * time.Minute

var syncSecretsRequiredKeys = []string{
	"CODEXK8S_GITHUB_REPO",
	"CODEXK8S_GIT_BOT_TOKEN",
	"CODEXK8S_BOOTSTRAP_OWNER_EMAIL",
	"CODEXK8S_GITHUB_OAUTH_CLIENT_ID",
	"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET",
	"CODEXK8S_PRODUCTION_DOMAIN",
	"CODEXK8S_AI_DOMAIN",
}

func runSyncSecrets(args []string, stdout io.Writer, stderr io.Writer) int {
	var vars kvList
	fs := flag.NewFlagSet("sync-secrets", flag.ContinueOnError)
	fs.SetOutput(stderr)

	envPath := fs.String("env-file", "bootstrap/host/config.env", "Path to bootstrap env file")
	kubeconfigPath := fs.String("kubeconfig", "", "Optional kubeconfig path")
	namespace := fs.String("namespace", "", "Override target namespace for secret sync")
	timeout := fs.Duration("timeout", defaultSyncSecretsTimeout, "Timeout for kubernetes operations")
	dryRun := fs.Bool("dry-run", false, "Print planned updates without writing to env file or kubernetes")
	fs.Var(&vars, "var", "Template variable in KEY=VALUE format (repeatable)")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *timeout <= 0 {
		writef(stderr, "sync-secrets failed: --timeout must be positive\n")
		return 2
	}

	absEnv, err := filepath.Abs(*envPath)
	if err != nil {
		writef(stderr, "sync-secrets failed: resolve env-file path: %v\n", err)
		return 1
	}
	values, err := envfile.Load(absEnv)
	if err != nil {
		writef(stderr, "sync-secrets failed: load env-file: %v\n", err)
		return 1
	}
	for key, value := range vars.Map() {
		values[key] = value
	}

	applyGitHubSyncDefaults(values)
	applySyncSecretsDefaults(values)
	missing := missingRequiredKeys(values, syncSecretsRequiredKeys)
	if len(missing) > 0 {
		writef(stderr, "sync-secrets failed: missing required env keys: %s\n", strings.Join(missing, ", "))
		return 1
	}

	targetNamespace := strings.TrimSpace(*namespace)
	if targetNamespace == "" {
		targetNamespace = strings.TrimSpace(values["CODEXK8S_PRODUCTION_NAMESPACE"])
	}
	if targetNamespace == "" {
		targetNamespace = "codex-k8s-prod"
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	restConfig, err := clientcfg.BuildRESTConfig(strings.TrimSpace(*kubeconfigPath))
	if err != nil {
		writef(stderr, "sync-secrets failed: build kubernetes rest config: %v\n", err)
		return 1
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		writef(stderr, "sync-secrets failed: build kubernetes clientset: %v\n", err)
		return 1
	}

	existingPostgres, err := getSecretData(ctx, clientset, targetNamespace, "codex-k8s-postgres")
	if err != nil {
		writef(stderr, "sync-secrets failed: load codex-k8s-postgres: %v\n", err)
		return 1
	}
	existingRuntime, err := getSecretData(ctx, clientset, targetNamespace, "codex-k8s-runtime")
	if err != nil {
		writef(stderr, "sync-secrets failed: load codex-k8s-runtime: %v\n", err)
		return 1
	}
	existingOAuth, err := getSecretData(ctx, clientset, targetNamespace, "codex-k8s-oauth2-proxy")
	if err != nil {
		writef(stderr, "sync-secrets failed: load codex-k8s-oauth2-proxy: %v\n", err)
		return 1
	}

	if err := hydrateValuesFromExistingSecrets(values, existingPostgres, existingRuntime, existingOAuth); err != nil {
		writef(stderr, "sync-secrets failed: hydrate values from existing secrets: %v\n", err)
		return 1
	}

	postgresSecret := buildPostgresSecretValues(values)
	runtimeSecret := buildRuntimeSecretValues(values)
	oauthSecret := buildOAuthSecretValues(values)

	if *dryRun {
		writef(stdout, "sync-secrets env-file=%s namespace=%s dry-run=true\n", absEnv, targetNamespace)
		writef(stdout, "resolved postgres keys=%d runtime keys=%d oauth keys=%d\n", len(postgresSecret), len(runtimeSecret), len(oauthSecret))
		printResolvedSecretKeys(stdout, postgresSecret, runtimeSecret, oauthSecret)
		return 0
	}

	if err := upsertSecretData(ctx, clientset, targetNamespace, "codex-k8s-postgres", postgresSecret); err != nil {
		writef(stderr, "sync-secrets failed: upsert codex-k8s-postgres: %v\n", err)
		return 1
	}
	if err := upsertSecretData(ctx, clientset, targetNamespace, "codex-k8s-runtime", runtimeSecret); err != nil {
		writef(stderr, "sync-secrets failed: upsert codex-k8s-runtime: %v\n", err)
		return 1
	}
	if len(oauthSecret) > 0 {
		if err := upsertSecretData(ctx, clientset, targetNamespace, "codex-k8s-oauth2-proxy", oauthSecret); err != nil {
			writef(stderr, "sync-secrets failed: upsert codex-k8s-oauth2-proxy: %v\n", err)
			return 1
		}
	}

	if err := envfile.Upsert(absEnv, values); err != nil {
		writef(stderr, "sync-secrets failed: persist env-file: %v\n", err)
		return 1
	}

	writef(stdout, "sync-secrets env-file=%s namespace=%s\n", absEnv, targetNamespace)
	writef(stdout, "upserted postgres keys=%d runtime keys=%d oauth keys=%d\n", len(postgresSecret), len(runtimeSecret), len(oauthSecret))
	return 0
}

func applySyncSecretsDefaults(values map[string]string) {
	setEnvDefault(values, "CODEXK8S_POSTGRES_DB", "codex_k8s")
	setEnvDefault(values, "CODEXK8S_POSTGRES_USER", "codex_k8s")
	setEnvDefault(values, "CODEXK8S_LEARNING_MODE_DEFAULT", "true")
	setEnvDefault(values, "CODEXK8S_JWT_TTL", "15m")
	setEnvDefault(values, "CODEXK8S_MCP_TOKEN_TTL", "24h")
	setEnvDefault(values, "CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS", "14")
	setEnvDefault(values, "CODEXK8S_GIT_BOT_USERNAME", "codex-bot")
	setEnvDefault(values, "CODEXK8S_GIT_BOT_MAIL", "codex-bot@codex-k8s.local")
	setEnvDefault(values, "CODEXK8S_PROJECT_DB_ADMIN_HOST", "postgres")
	setEnvDefault(values, "CODEXK8S_PROJECT_DB_ADMIN_PORT", "5432")
	setEnvDefault(values, "CODEXK8S_PROJECT_DB_ADMIN_SSLMODE", "disable")
	setEnvDefault(values, "CODEXK8S_PROJECT_DB_ADMIN_DATABASE", "postgres")
	setEnvDefault(values, "CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS", "dev,production,prod")
	setEnvDefault(values, "CODEXK8S_GITHUB_WEBHOOK_URL", resolveWebhookURL(values))
	setEnvDefault(values, "CODEXK8S_GITHUB_WEBHOOK_EVENTS", defaultGitHubWebhookEvents)
	setEnvDefault(values, "CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE", "20Gi")
}

func hydrateValuesFromExistingSecrets(values map[string]string, existingPostgres map[string][]byte, existingRuntime map[string][]byte, existingOAuth map[string][]byte) error {
	transferFromSecret(values, existingPostgres, []string{
		"CODEXK8S_POSTGRES_DB",
		"CODEXK8S_POSTGRES_USER",
		"CODEXK8S_POSTGRES_PASSWORD",
	})
	transferFromSecret(values, existingRuntime, []string{
		"CODEXK8S_GITHUB_PAT",
		"CODEXK8S_GITHUB_REPO",
		"CODEXK8S_FIRST_PROJECT_GITHUB_REPO",
		"CODEXK8S_OPENAI_API_KEY",
		"CODEXK8S_OPENAI_AUTH_FILE",
		"CODEXK8S_PROJECT_DB_ADMIN_HOST",
		"CODEXK8S_PROJECT_DB_ADMIN_PORT",
		"CODEXK8S_PROJECT_DB_ADMIN_USER",
		"CODEXK8S_PROJECT_DB_ADMIN_PASSWORD",
		"CODEXK8S_PROJECT_DB_ADMIN_SSLMODE",
		"CODEXK8S_PROJECT_DB_ADMIN_DATABASE",
		"CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS",
		"CODEXK8S_GIT_BOT_TOKEN",
		"CODEXK8S_GIT_BOT_USERNAME",
		"CODEXK8S_GIT_BOT_MAIL",
		"CODEXK8S_CONTEXT7_API_KEY",
		"CODEXK8S_APP_SECRET_KEY",
		"CODEXK8S_TOKEN_ENCRYPTION_KEY",
		"CODEXK8S_MCP_TOKEN_SIGNING_KEY",
		"CODEXK8S_MCP_TOKEN_TTL",
		"CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS",
		"CODEXK8S_LEARNING_MODE_DEFAULT",
		"CODEXK8S_GITHUB_WEBHOOK_SECRET",
		"CODEXK8S_GITHUB_WEBHOOK_URL",
		"CODEXK8S_GITHUB_WEBHOOK_EVENTS",
		"CODEXK8S_PRODUCTION_DOMAIN",
		"CODEXK8S_AI_DOMAIN",
		"CODEXK8S_PUBLIC_BASE_URL",
		"CODEXK8S_BOOTSTRAP_OWNER_EMAIL",
		"CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS",
		"CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS",
		"CODEXK8S_GITHUB_OAUTH_CLIENT_ID",
		"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET",
		"CODEXK8S_JWT_SIGNING_KEY",
		"CODEXK8S_JWT_TTL",
		"CODEXK8S_INTERNAL_REGISTRY_SERVICE",
		"CODEXK8S_INTERNAL_REGISTRY_PORT",
		"CODEXK8S_INTERNAL_REGISTRY_HOST",
		"CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE",
		"CODEXK8S_VITE_DEV_UPSTREAM",
	})
	transferFromSecret(values, existingOAuth, []string{
		"OAUTH2_PROXY_COOKIE_SECRET",
	})

	var err error
	values["CODEXK8S_POSTGRES_PASSWORD"], err = valueOrRandomHex(values, "CODEXK8S_POSTGRES_PASSWORD", 24)
	if err != nil {
		return err
	}
	values["CODEXK8S_APP_SECRET_KEY"], err = valueOrRandomHex(values, "CODEXK8S_APP_SECRET_KEY", 32)
	if err != nil {
		return err
	}
	values["CODEXK8S_TOKEN_ENCRYPTION_KEY"], err = valueOrRandomHex(values, "CODEXK8S_TOKEN_ENCRYPTION_KEY", 32)
	if err != nil {
		return err
	}
	values["CODEXK8S_GITHUB_WEBHOOK_SECRET"], err = valueOrRandomHex(values, "CODEXK8S_GITHUB_WEBHOOK_SECRET", 32)
	if err != nil {
		return err
	}
	values["CODEXK8S_JWT_SIGNING_KEY"], err = valueOrRandomHex(values, "CODEXK8S_JWT_SIGNING_KEY", 32)
	if err != nil {
		return err
	}
	values["OAUTH2_PROXY_COOKIE_SECRET"], err = valueOrRandomHex(values, "OAUTH2_PROXY_COOKIE_SECRET", 16)
	if err != nil {
		return err
	}

	if strings.TrimSpace(values["CODEXK8S_MCP_TOKEN_SIGNING_KEY"]) == "" {
		values["CODEXK8S_MCP_TOKEN_SIGNING_KEY"] = strings.TrimSpace(values["CODEXK8S_TOKEN_ENCRYPTION_KEY"])
	}
	if strings.TrimSpace(values["CODEXK8S_FIRST_PROJECT_GITHUB_REPO"]) == "" {
		values["CODEXK8S_FIRST_PROJECT_GITHUB_REPO"] = strings.TrimSpace(values["CODEXK8S_GITHUB_REPO"])
	}
	if strings.TrimSpace(values["CODEXK8S_PUBLIC_BASE_URL"]) == "" {
		domain := strings.TrimSpace(values["CODEXK8S_PRODUCTION_DOMAIN"])
		if domain != "" {
			values["CODEXK8S_PUBLIC_BASE_URL"] = "https://" + domain
		}
	}
	if strings.TrimSpace(values["CODEXK8S_PROJECT_DB_ADMIN_USER"]) == "" {
		values["CODEXK8S_PROJECT_DB_ADMIN_USER"] = strings.TrimSpace(values["CODEXK8S_POSTGRES_USER"])
	}
	if strings.TrimSpace(values["CODEXK8S_PROJECT_DB_ADMIN_PASSWORD"]) == "" {
		values["CODEXK8S_PROJECT_DB_ADMIN_PASSWORD"] = strings.TrimSpace(values["CODEXK8S_POSTGRES_PASSWORD"])
	}
	return nil
}

func buildPostgresSecretValues(values map[string]string) map[string]string {
	return compactStringMap(map[string]string{
		"CODEXK8S_POSTGRES_DB":       strings.TrimSpace(values["CODEXK8S_POSTGRES_DB"]),
		"CODEXK8S_POSTGRES_USER":     strings.TrimSpace(values["CODEXK8S_POSTGRES_USER"]),
		"CODEXK8S_POSTGRES_PASSWORD": strings.TrimSpace(values["CODEXK8S_POSTGRES_PASSWORD"]),
	})
}

func buildRuntimeSecretValues(values map[string]string) map[string]string {
	return compactStringMap(map[string]string{
		"CODEXK8S_INTERNAL_REGISTRY_SERVICE":         strings.TrimSpace(values["CODEXK8S_INTERNAL_REGISTRY_SERVICE"]),
		"CODEXK8S_INTERNAL_REGISTRY_PORT":            strings.TrimSpace(values["CODEXK8S_INTERNAL_REGISTRY_PORT"]),
		"CODEXK8S_INTERNAL_REGISTRY_HOST":            strings.TrimSpace(values["CODEXK8S_INTERNAL_REGISTRY_HOST"]),
		"CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE":    strings.TrimSpace(values["CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE"]),
		"CODEXK8S_GITHUB_PAT":                        strings.TrimSpace(values["CODEXK8S_GITHUB_PAT"]),
		"CODEXK8S_GITHUB_REPO":                       strings.TrimSpace(values["CODEXK8S_GITHUB_REPO"]),
		"CODEXK8S_FIRST_PROJECT_GITHUB_REPO":         strings.TrimSpace(values["CODEXK8S_FIRST_PROJECT_GITHUB_REPO"]),
		"CODEXK8S_OPENAI_API_KEY":                    strings.TrimSpace(values["CODEXK8S_OPENAI_API_KEY"]),
		"CODEXK8S_OPENAI_AUTH_FILE":                  strings.TrimSpace(values["CODEXK8S_OPENAI_AUTH_FILE"]),
		"CODEXK8S_PROJECT_DB_ADMIN_HOST":             strings.TrimSpace(values["CODEXK8S_PROJECT_DB_ADMIN_HOST"]),
		"CODEXK8S_PROJECT_DB_ADMIN_PORT":             strings.TrimSpace(values["CODEXK8S_PROJECT_DB_ADMIN_PORT"]),
		"CODEXK8S_PROJECT_DB_ADMIN_USER":             strings.TrimSpace(values["CODEXK8S_PROJECT_DB_ADMIN_USER"]),
		"CODEXK8S_PROJECT_DB_ADMIN_PASSWORD":         strings.TrimSpace(values["CODEXK8S_PROJECT_DB_ADMIN_PASSWORD"]),
		"CODEXK8S_PROJECT_DB_ADMIN_SSLMODE":          strings.TrimSpace(values["CODEXK8S_PROJECT_DB_ADMIN_SSLMODE"]),
		"CODEXK8S_PROJECT_DB_ADMIN_DATABASE":         strings.TrimSpace(values["CODEXK8S_PROJECT_DB_ADMIN_DATABASE"]),
		"CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS": strings.TrimSpace(values["CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS"]),
		"CODEXK8S_GIT_BOT_TOKEN":                     strings.TrimSpace(values["CODEXK8S_GIT_BOT_TOKEN"]),
		"CODEXK8S_GIT_BOT_USERNAME":                  strings.TrimSpace(values["CODEXK8S_GIT_BOT_USERNAME"]),
		"CODEXK8S_GIT_BOT_MAIL":                      strings.TrimSpace(values["CODEXK8S_GIT_BOT_MAIL"]),
		"CODEXK8S_CONTEXT7_API_KEY":                  strings.TrimSpace(values["CODEXK8S_CONTEXT7_API_KEY"]),
		"CODEXK8S_APP_SECRET_KEY":                    strings.TrimSpace(values["CODEXK8S_APP_SECRET_KEY"]),
		"CODEXK8S_TOKEN_ENCRYPTION_KEY":              strings.TrimSpace(values["CODEXK8S_TOKEN_ENCRYPTION_KEY"]),
		"CODEXK8S_MCP_TOKEN_SIGNING_KEY":             strings.TrimSpace(values["CODEXK8S_MCP_TOKEN_SIGNING_KEY"]),
		"CODEXK8S_MCP_TOKEN_TTL":                     strings.TrimSpace(values["CODEXK8S_MCP_TOKEN_TTL"]),
		"CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS":     strings.TrimSpace(values["CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS"]),
		"CODEXK8S_LEARNING_MODE_DEFAULT":             strings.TrimSpace(values["CODEXK8S_LEARNING_MODE_DEFAULT"]),
		"CODEXK8S_GITHUB_WEBHOOK_SECRET":             strings.TrimSpace(values["CODEXK8S_GITHUB_WEBHOOK_SECRET"]),
		"CODEXK8S_GITHUB_WEBHOOK_URL":                strings.TrimSpace(values["CODEXK8S_GITHUB_WEBHOOK_URL"]),
		"CODEXK8S_GITHUB_WEBHOOK_EVENTS":             strings.TrimSpace(values["CODEXK8S_GITHUB_WEBHOOK_EVENTS"]),
		"CODEXK8S_PRODUCTION_DOMAIN":                 strings.TrimSpace(values["CODEXK8S_PRODUCTION_DOMAIN"]),
		"CODEXK8S_AI_DOMAIN":                         strings.TrimSpace(values["CODEXK8S_AI_DOMAIN"]),
		"CODEXK8S_PUBLIC_BASE_URL":                   strings.TrimSpace(values["CODEXK8S_PUBLIC_BASE_URL"]),
		"CODEXK8S_BOOTSTRAP_OWNER_EMAIL":             strings.TrimSpace(values["CODEXK8S_BOOTSTRAP_OWNER_EMAIL"]),
		"CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS":          strings.TrimSpace(values["CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS"]),
		"CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS":   strings.TrimSpace(values["CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS"]),
		"CODEXK8S_GITHUB_OAUTH_CLIENT_ID":            strings.TrimSpace(values["CODEXK8S_GITHUB_OAUTH_CLIENT_ID"]),
		"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET":        strings.TrimSpace(values["CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET"]),
		"CODEXK8S_JWT_SIGNING_KEY":                   strings.TrimSpace(values["CODEXK8S_JWT_SIGNING_KEY"]),
		"CODEXK8S_JWT_TTL":                           strings.TrimSpace(values["CODEXK8S_JWT_TTL"]),
		"CODEXK8S_VITE_DEV_UPSTREAM":                 strings.TrimSpace(values["CODEXK8S_VITE_DEV_UPSTREAM"]),
	})
}

func buildOAuthSecretValues(values map[string]string) map[string]string {
	return compactStringMap(map[string]string{
		"OAUTH2_PROXY_CLIENT_ID":     strings.TrimSpace(values["CODEXK8S_GITHUB_OAUTH_CLIENT_ID"]),
		"OAUTH2_PROXY_CLIENT_SECRET": strings.TrimSpace(values["CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET"]),
		"OAUTH2_PROXY_COOKIE_SECRET": strings.TrimSpace(values["OAUTH2_PROXY_COOKIE_SECRET"]),
	})
}

func transferFromSecret(values map[string]string, secret map[string][]byte, keys []string) {
	if len(secret) == 0 || len(keys) == 0 {
		return
	}
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" || strings.TrimSpace(values[key]) != "" {
			continue
		}
		raw, ok := secret[key]
		if !ok {
			continue
		}
		value := strings.TrimSpace(string(raw))
		if value == "" {
			continue
		}
		values[key] = value
	}
}

func compactStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	return out
}

func printResolvedSecretKeys(out io.Writer, postgres map[string]string, runtime map[string]string, oauth map[string]string) {
	writeln(out, "postgres secret keys:")
	printSortedKeys(out, postgres)
	writeln(out, "runtime secret keys:")
	printSortedKeys(out, runtime)
	writeln(out, "oauth2-proxy secret keys:")
	printSortedKeys(out, oauth)
}

func printSortedKeys(out io.Writer, values map[string]string) {
	if len(values) == 0 {
		writeln(out, "  - <none>")
		return
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		writef(out, "  - %s\n", key)
	}
}

func getSecretData(ctx context.Context, clientset kubernetes.Interface, namespace string, name string) (map[string][]byte, error) {
	secret, err := clientset.CoreV1().Secrets(namespace).Get(ctx, strings.TrimSpace(name), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return map[string][]byte{}, nil
		}
		return nil, err
	}
	out := make(map[string][]byte, len(secret.Data))
	for key, value := range secret.Data {
		out[key] = append([]byte(nil), value...)
	}
	return out, nil
}

func upsertSecretData(ctx context.Context, clientset kubernetes.Interface, namespace string, name string, data map[string]string) error {
	targetNamespace := strings.TrimSpace(namespace)
	targetName := strings.TrimSpace(name)
	if targetNamespace == "" || targetName == "" {
		return fmt.Errorf("namespace and secret name are required")
	}

	existing, err := clientset.CoreV1().Secrets(targetNamespace).Get(ctx, targetName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	merged := make(map[string][]byte)
	if existing != nil && existing.Data != nil {
		for key, value := range existing.Data {
			merged[key] = append([]byte(nil), value...)
		}
	}
	for key, value := range data {
		trimmedKey := strings.TrimSpace(key)
		trimmedValue := strings.TrimSpace(value)
		if trimmedKey == "" || trimmedValue == "" {
			continue
		}
		merged[trimmedKey] = []byte(trimmedValue)
	}
	if len(merged) == 0 {
		return nil
	}

	if apierrors.IsNotFound(err) {
		_, createErr := clientset.CoreV1().Secrets(targetNamespace).Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: targetName},
			Type:       corev1.SecretTypeOpaque,
			Data:       merged,
		}, metav1.CreateOptions{})
		return createErr
	}

	existing.Data = merged
	if existing.Type == "" {
		existing.Type = corev1.SecretTypeOpaque
	}
	_, updateErr := clientset.CoreV1().Secrets(targetNamespace).Update(ctx, existing, metav1.UpdateOptions{})
	return updateErr
}

func valueOrRandomHex(values map[string]string, key string, numBytes int) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("key is required")
	}
	value := strings.TrimSpace(values[key])
	if value != "" {
		return value, nil
	}
	generated, err := randomHexString(numBytes)
	if err != nil {
		return "", err
	}
	values[key] = generated
	return generated, nil
}

func randomHexString(numBytes int) (string, error) {
	if numBytes <= 0 {
		numBytes = 16
	}
	raw := make([]byte, numBytes)
	if _, err := cryptorand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
