package cli

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codex-k8s/codex-k8s/cmd/codex-bootstrap/internal/envfile"
	gh "github.com/google/go-github/v82/github"
	"golang.org/x/crypto/nacl/box"
)

const (
	defaultGitHubSyncTimeout      = 15 * time.Minute
	defaultGitHubSyncWorkers      = 2
	defaultGitHubWebhookEvents    = "push,pull_request,issues,issue_comment,pull_request_review,pull_request_review_comment"
	defaultGitHubLabelDescription = "codex-k8s managed label"
	defaultGitHubLabelColor       = "1f6feb"

	githubEnvironmentProduction = "production"
	githubEnvironmentAI         = "ai"
)

var (
	githubEnvVariableKeys = []string{
		"CODEXK8S_PRODUCTION_NAMESPACE",
		"CODEXK8S_PRODUCTION_DOMAIN",
		"CODEXK8S_AI_DOMAIN",
		"CODEXK8S_PUBLIC_BASE_URL",
	}
	githubRepoSecretKeys = []string{
		"CODEXK8S_OPENAI_API_KEY",
		"CODEXK8S_OPENAI_AUTH_FILE",
		"CODEXK8S_POSTGRES_PASSWORD",
		"CODEXK8S_APP_SECRET_KEY",
		"CODEXK8S_TOKEN_ENCRYPTION_KEY",
		"CODEXK8S_GITHUB_WEBHOOK_SECRET",
		"CODEXK8S_GITHUB_PAT",
		"CODEXK8S_GIT_BOT_TOKEN",
		"CODEXK8S_PROJECT_DB_ADMIN_HOST",
		"CODEXK8S_PROJECT_DB_ADMIN_PORT",
		"CODEXK8S_PROJECT_DB_ADMIN_USER",
		"CODEXK8S_PROJECT_DB_ADMIN_PASSWORD",
		"CODEXK8S_PROJECT_DB_ADMIN_SSLMODE",
		"CODEXK8S_PROJECT_DB_ADMIN_DATABASE",
		"CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS",
		"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET",
		"CODEXK8S_JWT_SIGNING_KEY",
		"CODEXK8S_MCP_TOKEN_SIGNING_KEY",
		"CODEXK8S_CONTEXT7_API_KEY",
	}
	githubLegacyVariableKeys = []string{
		"POSTGRES_DB",
		"POSTGRES_USER",
		"LEARNING_MODE_DEFAULT",
		"RUNNER_NAMESPACE",
		"RUNNER_SCALE_SET_NAME",
		"RUNNER_MIN",
		"RUNNER_MAX",
		"RUNNER_IMAGE",
		"CODEXK8S_GITHUB_USERNAME",
		"CODEXK8S_IMAGE",
		"CODEXK8S_INTERNAL_IMAGE_REPOSITORY",
	}
	githubLegacySecretKeys = []string{
		"OPENAI_API_KEY",
		"CONTEXT7_API_KEY",
		"CODEXK8S_GITHUB_TOKEN",
		"CODEXK8S_GITHUB_USERNAME",
	}
	githubRequiredSyncKeys = []string{
		"CODEXK8S_GITHUB_REPO",
		"CODEXK8S_GITHUB_PAT",
		"CODEXK8S_GITHUB_WEBHOOK_SECRET",
		"CODEXK8S_PRODUCTION_DOMAIN",
		"CODEXK8S_AI_DOMAIN",
	}
)

type githubRepositoryRef struct {
	Owner    string
	Name     string
	FullName string
}

type githubRepoPublicKey struct {
	KeyID string
	Key   [32]byte
}

func runGitHubSync(args []string, stdout io.Writer, stderr io.Writer) int {
	var vars kvList
	fs := flag.NewFlagSet("github-sync", flag.ContinueOnError)
	fs.SetOutput(stderr)

	envPath := fs.String("env-file", "bootstrap/host/config.env", "Path to bootstrap env file")
	timeout := fs.Duration("timeout", defaultGitHubSyncTimeout, "GitHub API timeout")
	workers := fs.Int("workers", defaultGitHubSyncWorkers, "Parallel workers for variable/secret sync")
	dryRun := fs.Bool("dry-run", false, "Print planned changes without applying them")
	skipVariables := fs.Bool("skip-variables", false, "Skip environment variables sync")
	skipSecrets := fs.Bool("skip-secrets", false, "Skip environment secrets sync")
	skipWebhook := fs.Bool("skip-webhook", false, "Skip repository webhook sync")
	skipLabels := fs.Bool("skip-labels", false, "Skip repository label sync")
	fs.Var(&vars, "var", "Template variable in KEY=VALUE format (repeatable)")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *timeout <= 0 {
		writef(stderr, "github-sync failed: --timeout must be positive\n")
		return 2
	}
	if *workers <= 0 {
		writef(stderr, "github-sync failed: --workers must be positive\n")
		return 2
	}

	absEnv, err := filepath.Abs(*envPath)
	if err != nil {
		writef(stderr, "github-sync failed: resolve env-file path: %v\n", err)
		return 1
	}
	values, err := envfile.Load(absEnv)
	if err != nil {
		writef(stderr, "github-sync failed: load env-file: %v\n", err)
		return 1
	}
	for key, value := range vars.Map() {
		values[key] = value
	}
	applyGitHubSyncDefaults(values)

	missing := missingRequiredKeys(values, githubRequiredSyncKeys)
	if len(missing) > 0 {
		writef(stderr, "github-sync failed: missing required env keys: %s\n", strings.Join(missing, ", "))
		return 1
	}

	platformRepo, err := parseGitHubRepository(values["CODEXK8S_GITHUB_REPO"])
	if err != nil {
		writef(stderr, "github-sync failed: CODEXK8S_GITHUB_REPO: %v\n", err)
		return 1
	}
	firstProjectRaw := strings.TrimSpace(values["CODEXK8S_FIRST_PROJECT_GITHUB_REPO"])
	firstProjectRepo := platformRepo
	if firstProjectRaw != "" {
		firstProjectRepo, err = parseGitHubRepository(firstProjectRaw)
		if err != nil {
			writef(stderr, "github-sync failed: CODEXK8S_FIRST_PROJECT_GITHUB_REPO: %v\n", err)
			return 1
		}
	}

	reposForWebhookAndLabels := []githubRepositoryRef{platformRepo}
	if firstProjectRepo.FullName != platformRepo.FullName {
		reposForWebhookAndLabels = append(reposForWebhookAndLabels, firstProjectRepo)
	}

	labels := collectGitHubLabels(values)
	webhookURL := resolveWebhookURL(values)
	webhookEvents := normalizeGitHubEvents(values["CODEXK8S_GITHUB_WEBHOOK_EVENTS"])
	webhookSecret := strings.TrimSpace(values["CODEXK8S_GITHUB_WEBHOOK_SECRET"])
	if webhookSecret == "" {
		writef(stderr, "github-sync failed: CODEXK8S_GITHUB_WEBHOOK_SECRET is required\n")
		return 1
	}

	writef(stdout, "github-sync env-file=%s\n", absEnv)
	writef(stdout, "platform-repo=%s\n", platformRepo.FullName)
	writef(stdout, "webhook-label-repos=%s\n", joinRepositoryNames(reposForWebhookAndLabels))
	writef(stdout, "labels=%d\n", len(labels))

	if *dryRun {
		writeln(stdout, "dry-run: no GitHub mutations were applied")
		return 0
	}

	client := gh.NewClient(&http.Client{Timeout: *timeout}).WithAuthToken(strings.TrimSpace(values["CODEXK8S_GITHUB_PAT"]))
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	repoID, err := getGitHubRepositoryID(ctx, client, platformRepo)
	if err != nil {
		writef(stderr, "github-sync failed: resolve repository id: %v\n", err)
		return 1
	}

	environments := []string{githubEnvironmentProduction, githubEnvironmentAI}
	for _, envName := range environments {
		if err := ensureGitHubEnvironment(ctx, client, platformRepo, envName); err != nil {
			writef(stderr, "github-sync failed: ensure environment %s: %v\n", envName, err)
			return 1
		}
	}

	if !*skipVariables {
		for _, envName := range environments {
			envValues := cloneStringMap(values)
			applyEnvironmentOverrides(envValues, envName, githubEnvVariableKeys)
			variableKeys := collectGitHubVariableKeys(envValues)
			writef(stdout, "sync %s environment variables=%d\n", envName, len(variableKeys))
			if err := syncGitHubEnvVariables(ctx, client, platformRepo, envName, envValues, variableKeys, *workers); err != nil {
				writef(stderr, "github-sync failed: sync %s environment variables: %v\n", envName, err)
				return 1
			}
		}
	}
	if !*skipSecrets {
		for _, envName := range environments {
			envValues := cloneStringMap(values)
			applyEnvironmentOverrides(envValues, envName, githubRepoSecretKeys)
			secretValues := collectGitHubSecretValues(envValues)
			writef(stdout, "sync %s environment secrets=%d\n", envName, len(secretValues))
			if err := syncGitHubEnvSecrets(ctx, client, platformRepo, repoID, envName, envValues, githubRepoSecretKeys, *workers); err != nil {
				writef(stderr, "github-sync failed: sync %s environment secrets: %v\n", envName, err)
				return 1
			}
		}
	}
	if !*skipWebhook {
		for _, repo := range reposForWebhookAndLabels {
			if err := ensureGitHubWebhook(ctx, client, repo, webhookURL, webhookSecret, webhookEvents); err != nil {
				writef(stderr, "github-sync failed: ensure webhook in %s: %v\n", repo.FullName, err)
				return 1
			}
		}
	}
	if !*skipLabels {
		for _, repo := range reposForWebhookAndLabels {
			if err := ensureGitHubLabels(ctx, client, repo, labels, *workers); err != nil {
				writef(stderr, "github-sync failed: ensure labels in %s: %v\n", repo.FullName, err)
				return 1
			}
		}
	}
	if err := cleanupLegacyGitHubMetadata(ctx, client, platformRepo); err != nil {
		writef(stderr, "github-sync failed: cleanup legacy metadata: %v\n", err)
		return 1
	}

	writeln(stdout, "github-sync completed")
	return 0
}

func applyGitHubSyncDefaults(values map[string]string) {
	setEnvDefault(values, "CODEXK8S_PRODUCTION_NAMESPACE", "codex-k8s-prod")
	setEnvDefault(values, "CODEXK8S_PRODUCTION_DOMAIN", "platform.codex-k8s.dev")
	if strings.TrimSpace(values["CODEXK8S_AI_DOMAIN"]) == "" {
		if productionDomain := strings.TrimSpace(values["CODEXK8S_PRODUCTION_DOMAIN"]); productionDomain != "" {
			values["CODEXK8S_AI_DOMAIN"] = "ai." + productionDomain
		}
	}
	setEnvDefault(values, "CODEXK8S_INTERNAL_REGISTRY_SERVICE", "codex-k8s-registry")
	setEnvDefault(values, "CODEXK8S_INTERNAL_REGISTRY_PORT", "5000")
	setEnvDefault(values, "CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE", "20Gi")
	setEnvDefault(values, "CODEXK8S_INTERNAL_REGISTRY_HOST", "127.0.0.1:"+strings.TrimSpace(values["CODEXK8S_INTERNAL_REGISTRY_PORT"]))
	setEnvDefault(values, "CODEXK8S_GITHUB_WEBHOOK_EVENTS", defaultGitHubWebhookEvents)
	setEnvDefault(values, "CODEXK8S_GITHUB_WEBHOOK_URL", resolveWebhookURL(values))
}

func setEnvDefault(values map[string]string, key string, fallback string) {
	if strings.TrimSpace(values[key]) != "" {
		return
	}
	values[key] = fallback
}

func parseGitHubRepository(value string) (githubRepositoryRef, error) {
	owner, repo, err := splitRepositoryFullName(value)
	if err != nil {
		return githubRepositoryRef{}, err
	}
	return githubRepositoryRef{
		Owner:    owner,
		Name:     repo,
		FullName: owner + "/" + repo,
	}, nil
}

func collectGitHubVariableKeys(values map[string]string) []string {
	secretSet := make(map[string]struct{}, len(githubRepoSecretKeys))
	for _, key := range githubRepoSecretKeys {
		secretSet[key] = struct{}{}
	}

	keysSet := make(map[string]struct{})
	for _, key := range githubEnvVariableKeys {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" {
			continue
		}
		if _, isSecret := secretSet[trimmed]; isSecret {
			continue
		}
		if strings.TrimSpace(values[trimmed]) == "" {
			continue
		}
		keysSet[trimmed] = struct{}{}
	}

	keys := make([]string, 0, len(keysSet))
	for key := range keysSet {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func collectGitHubSecretValues(values map[string]string) map[string]string {
	out := make(map[string]string)
	for _, key := range githubRepoSecretKeys {
		value := strings.TrimSpace(values[key])
		if value == "" {
			continue
		}
		out[key] = value
	}
	return out
}

func collectGitHubLabels(values map[string]string) map[string]string {
	out := make(map[string]string)
	keys := make([]string, 0)
	for key, value := range values {
		if !strings.HasSuffix(key, "_LABEL") {
			continue
		}
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out[trimmed] = labelDescriptionForKey(key)
		keys = append(keys, trimmed)
	}
	sort.Strings(keys)
	return out
}

func labelDescriptionForKey(key string) string {
	labelKey := strings.TrimPrefix(strings.TrimSpace(key), "CODEXK8S_")
	labelKey = strings.TrimSuffix(labelKey, "_LABEL")
	labelKey = strings.ToLower(strings.ReplaceAll(labelKey, "_", " "))
	if labelKey == "" {
		return defaultGitHubLabelDescription
	}
	return "codex-k8s managed label: " + labelKey
}

func normalizeGitHubEvents(raw string) []string {
	items := strings.Split(strings.TrimSpace(raw), ",")
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		event := strings.TrimSpace(item)
		if event == "" {
			continue
		}
		if _, ok := seen[event]; ok {
			continue
		}
		seen[event] = struct{}{}
		out = append(out, event)
	}
	if len(out) == 0 {
		out = append(out, "push")
	}
	return out
}

func joinRepositoryNames(repos []githubRepositoryRef) string {
	items := make([]string, 0, len(repos))
	for _, repo := range repos {
		items = append(items, repo.FullName)
	}
	return strings.Join(items, ",")
}

func getGitHubRepositoryID(ctx context.Context, client *gh.Client, repo githubRepositoryRef) (int, error) {
	repository, _, err := client.Repositories.Get(ctx, repo.Owner, repo.Name)
	if err != nil {
		return 0, fmt.Errorf("get repository %s: %w", repo.FullName, err)
	}
	id := repository.GetID()
	if id <= 0 {
		return 0, fmt.Errorf("repository %s has invalid id", repo.FullName)
	}
	return int(id), nil
}

func ensureGitHubEnvironment(ctx context.Context, client *gh.Client, repo githubRepositoryRef, name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil
	}
	if _, _, err := client.Repositories.CreateUpdateEnvironment(ctx, repo.Owner, repo.Name, trimmed, &gh.CreateUpdateEnvironment{}); err != nil {
		return err
	}
	return nil
}

func syncGitHubEnvVariables(ctx context.Context, client *gh.Client, repo githubRepositoryRef, env string, values map[string]string, keys []string, workers int) error {
	existingVars, err := listGitHubEnvVariableValues(ctx, client, repo, env)
	if err != nil {
		return err
	}

	ops := make([]githubOperation, 0, len(keys))
	for _, key := range keys {
		trimmedKey := strings.TrimSpace(key)
		value := strings.TrimSpace(values[trimmedKey])
		if value == "" {
			continue
		}
		existingValue, exists := existingVars[trimmedKey]
		if exists && strings.TrimSpace(existingValue) == value {
			continue
		}

		keyCopy := trimmedKey
		valueCopy := value
		existsCopy := exists
		ops = append(ops, githubOperation{
			Name: "variable " + keyCopy,
			Run: func(ctx context.Context) error {
				return upsertGitHubEnvVariable(ctx, client, repo, env, keyCopy, valueCopy, existsCopy)
			},
		})
	}
	// Environment variables endpoint is aggressively rate-limited; keep it sequential to
	// avoid secondary rate-limit storms.
	return runGitHubOperations(ctx, 1, ops)
}

func upsertGitHubEnvVariable(ctx context.Context, client *gh.Client, repo githubRepositoryRef, env string, key string, value string, exists bool) error {
	trimmedKey := strings.TrimSpace(key)
	trimmedEnv := strings.TrimSpace(env)
	if trimmedKey == "" || trimmedEnv == "" {
		return nil
	}
	if strings.TrimSpace(value) == "" {
		// Empty values mean "do not overwrite GitHub".
		return nil
	}

	payload := &gh.ActionsVariable{Name: trimmedKey, Value: value}
	if exists {
		if _, err := client.Actions.UpdateEnvVariable(ctx, repo.Owner, repo.Name, trimmedEnv, payload); err == nil {
			return nil
		} else if !isGitHubNotFound(err) {
			if !isGitHubConflict(err) && !isGitHubUnprocessable(err) {
				return fmt.Errorf("update variable %s: %w", trimmedKey, err)
			}
		}
	}

	if _, err := client.Actions.CreateEnvVariable(ctx, repo.Owner, repo.Name, trimmedEnv, payload); err != nil {
		if isGitHubConflict(err) || isGitHubUnprocessable(err) {
			if _, updateErr := client.Actions.UpdateEnvVariable(ctx, repo.Owner, repo.Name, trimmedEnv, payload); updateErr != nil {
				return fmt.Errorf("update variable %s after conflict: %w", trimmedKey, updateErr)
			}
			return nil
		}
		return fmt.Errorf("create variable %s: %w", trimmedKey, err)
	}
	return nil
}

func listGitHubEnvVariableValues(ctx context.Context, client *gh.Client, repo githubRepositoryRef, envName string) (map[string]string, error) {
	page := 1
	out := make(map[string]string)
	for {
		vars, resp, err := client.Actions.ListEnvVariables(ctx, repo.Owner, repo.Name, envName, &gh.ListOptions{PerPage: 100, Page: page})
		if err != nil {
			return nil, err
		}
		if vars != nil {
			for _, item := range vars.Variables {
				if item == nil {
					continue
				}
				name := strings.TrimSpace(item.Name)
				if name == "" {
					continue
				}
				out[name] = strings.TrimSpace(item.Value)
			}
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return out, nil
}

func syncGitHubEnvSecrets(ctx context.Context, client *gh.Client, repo githubRepositoryRef, repoID int, env string, values map[string]string, keys []string, workers int) error {
	publicKey, err := getGitHubEnvPublicKey(ctx, client, repo, repoID, env)
	if err != nil {
		return err
	}

	ops := make([]githubOperation, 0, len(keys))
	for _, key := range keys {
		key := strings.TrimSpace(key)
		if key == "" {
			continue
		}
		value := strings.TrimSpace(values[key])
		ops = append(ops, githubOperation{
			Name: "secret " + key,
			Run: func(ctx context.Context) error {
				return upsertGitHubEnvSecret(ctx, client, repoID, env, publicKey, key, value)
			},
		})
	}
	return runGitHubOperations(ctx, workers, ops)
}

func getGitHubEnvPublicKey(ctx context.Context, client *gh.Client, repo githubRepositoryRef, repoID int, env string) (githubRepoPublicKey, error) {
	publicKey, _, err := client.Actions.GetEnvPublicKey(ctx, repoID, strings.TrimSpace(env))
	if err != nil {
		return githubRepoPublicKey{}, fmt.Errorf("get environment public key for %s env %s: %w", repo.FullName, env, err)
	}

	keyID := strings.TrimSpace(publicKey.GetKeyID())
	keyEncoded := strings.TrimSpace(publicKey.GetKey())
	if keyID == "" || keyEncoded == "" {
		return githubRepoPublicKey{}, fmt.Errorf("environment public key for %s env %s is invalid", repo.FullName, env)
	}
	decoded, err := base64.StdEncoding.DecodeString(keyEncoded)
	if err != nil {
		return githubRepoPublicKey{}, fmt.Errorf("decode environment public key for %s env %s: %w", repo.FullName, env, err)
	}
	if len(decoded) != 32 {
		return githubRepoPublicKey{}, fmt.Errorf("environment public key for %s env %s has invalid length", repo.FullName, env)
	}
	var key [32]byte
	copy(key[:], decoded)
	return githubRepoPublicKey{KeyID: keyID, Key: key}, nil
}

func upsertGitHubEnvSecret(ctx context.Context, client *gh.Client, repoID int, env string, publicKey githubRepoPublicKey, key string, value string) error {
	name := strings.TrimSpace(key)
	targetEnv := strings.TrimSpace(env)
	if name == "" || targetEnv == "" {
		return nil
	}
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		// Empty values mean "do not overwrite GitHub".
		return nil
	}

	encrypted, err := encryptGitHubSecretValue(trimmedValue, publicKey.Key)
	if err != nil {
		return fmt.Errorf("encrypt secret %s: %w", name, err)
	}
	if _, err := client.Actions.CreateOrUpdateEnvSecret(ctx, repoID, targetEnv, &gh.EncryptedSecret{
		Name:           name,
		KeyID:          publicKey.KeyID,
		EncryptedValue: encrypted,
	}); err != nil {
		return fmt.Errorf("upsert secret %s: %w", name, err)
	}
	return nil
}

func ensureGitHubWebhook(ctx context.Context, client *gh.Client, repo githubRepositoryRef, webhookURL string, webhookSecret string, events []string) error {
	webhookURL = strings.TrimSpace(webhookURL)
	webhookSecret = strings.TrimSpace(webhookSecret)
	if webhookURL == "" {
		return fmt.Errorf("webhook url is required")
	}
	if webhookSecret == "" {
		return fmt.Errorf("webhook secret is required")
	}

	hooks, _, err := client.Repositories.ListHooks(ctx, repo.Owner, repo.Name, &gh.ListOptions{PerPage: 100})
	if err != nil {
		return fmt.Errorf("list hooks for %s: %w", repo.FullName, err)
	}

	desired := &gh.Hook{
		Name:   gh.Ptr("web"),
		Active: gh.Ptr(true),
		Events: events,
		Config: &gh.HookConfig{
			URL:         gh.Ptr(webhookURL),
			ContentType: gh.Ptr("json"),
			InsecureSSL: gh.Ptr("0"),
			Secret:      gh.Ptr(webhookSecret),
		},
	}
	for _, hook := range hooks {
		if hook == nil || hook.Config == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(hook.Config.GetURL()), webhookURL) {
			if _, _, err := client.Repositories.EditHook(ctx, repo.Owner, repo.Name, hook.GetID(), desired); err != nil {
				return fmt.Errorf("edit webhook in %s: %w", repo.FullName, err)
			}
			return nil
		}
	}
	if _, _, err := client.Repositories.CreateHook(ctx, repo.Owner, repo.Name, desired); err != nil {
		return fmt.Errorf("create webhook in %s: %w", repo.FullName, err)
	}
	return nil
}

func ensureGitHubLabels(ctx context.Context, client *gh.Client, repo githubRepositoryRef, labels map[string]string, workers int) error {
	if len(labels) == 0 {
		return nil
	}
	keys := make([]string, 0, len(labels))
	for label := range labels {
		keys = append(keys, label)
	}
	sort.Strings(keys)

	ops := make([]githubOperation, 0, len(keys))
	for _, labelName := range keys {
		labelName := labelName
		description := labels[labelName]
		ops = append(ops, githubOperation{
			Name: "label " + labelName,
			Run: func(ctx context.Context) error {
				return ensureGitHubLabel(ctx, client, repo, labelName, description)
			},
		})
	}
	return runGitHubOperations(ctx, workers, ops)
}

func ensureGitHubLabel(ctx context.Context, client *gh.Client, repo githubRepositoryRef, labelName string, description string) error {
	name := strings.TrimSpace(labelName)
	if name == "" {
		return nil
	}
	description = strings.TrimSpace(description)
	if description == "" {
		description = defaultGitHubLabelDescription
	}

	_, _, getErr := client.Issues.GetLabel(ctx, repo.Owner, repo.Name, name)
	if getErr == nil {
		if _, _, err := client.Issues.EditLabel(ctx, repo.Owner, repo.Name, name, &gh.Label{
			Name:        gh.Ptr(name),
			Description: gh.Ptr(description),
			Color:       gh.Ptr(defaultGitHubLabelColor),
		}); err != nil {
			return fmt.Errorf("edit label %s: %w", name, err)
		}
		return nil
	}
	if !isGitHubNotFound(getErr) {
		return fmt.Errorf("get label %s: %w", name, getErr)
	}

	if _, _, err := client.Issues.CreateLabel(ctx, repo.Owner, repo.Name, &gh.Label{
		Name:        gh.Ptr(name),
		Description: gh.Ptr(description),
		Color:       gh.Ptr(defaultGitHubLabelColor),
	}); err != nil {
		if isGitHubConflict(err) || isGitHubUnprocessable(err) {
			if _, _, updateErr := client.Issues.EditLabel(ctx, repo.Owner, repo.Name, name, &gh.Label{
				Name:        gh.Ptr(name),
				Description: gh.Ptr(description),
				Color:       gh.Ptr(defaultGitHubLabelColor),
			}); updateErr != nil {
				return fmt.Errorf("edit label %s after conflict: %w", name, updateErr)
			}
			return nil
		}
		return fmt.Errorf("create label %s: %w", name, err)
	}
	return nil
}

func cleanupLegacyGitHubMetadata(ctx context.Context, client *gh.Client, repo githubRepositoryRef) error {
	var cleanupErrors []error

	for _, key := range githubLegacySecretKeys {
		if _, err := client.Actions.DeleteRepoSecret(ctx, repo.Owner, repo.Name, key); err != nil && !isGitHubNotFound(err) {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("delete legacy secret %s: %w", key, err))
		}
	}
	for _, key := range githubLegacyVariableKeys {
		if _, err := client.Actions.DeleteRepoVariable(ctx, repo.Owner, repo.Name, key); err != nil && !isGitHubNotFound(err) {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("delete legacy variable %s: %w", key, err))
		}
	}
	return errors.Join(cleanupErrors...)
}

func encryptGitHubSecretValue(value string, publicKey [32]byte) (string, error) {
	encryptedRaw, err := box.SealAnonymous(nil, []byte(value), &publicKey, rand.Reader)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encryptedRaw), nil
}

type githubOperation struct {
	Name string
	Run  func(ctx context.Context) error
}

func runGitHubOperations(ctx context.Context, workers int, operations []githubOperation) error {
	if len(operations) == 0 {
		return nil
	}
	if workers <= 0 {
		workers = 1
	}

	semaphore := make(chan struct{}, workers)
	var waitGroup sync.WaitGroup
	errs := make([]error, 0)
	var errsMu sync.Mutex

	for _, operation := range operations {
		operation := operation
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				errsMu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", operation.Name, ctx.Err()))
				errsMu.Unlock()
				return
			}
			defer func() { <-semaphore }()

			if err := runGitHubOperationWithRetry(ctx, operation); err != nil {
				errsMu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", operation.Name, err))
				errsMu.Unlock()
			}
		}()
	}
	waitGroup.Wait()
	return errors.Join(errs...)
}

func runGitHubOperationWithRetry(ctx context.Context, operation githubOperation) error {
	var lastErr error
	const maxAttempts = 20

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := operation.Run(ctx)
		if err == nil {
			return nil
		}
		lastErr = err
		if attempt >= maxAttempts || !isGitHubRetryable(err) {
			return err
		}

		delay := gitHubRetryDelay(err, attempt)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return lastErr
}

func isGitHubRetryable(err error) bool {
	code := githubStatusCode(err)
	if code == http.StatusTooManyRequests {
		return true
	}
	if code >= 500 && code < 600 {
		return true
	}
	// GitHub sometimes responds with 403 for secondary rate limits.
	if code == http.StatusForbidden {
		_, ok := gitHubRetryAfter(err)
		return ok
	}
	return false
}

func gitHubRetryDelay(err error, attempt int) time.Duration {
	if delay, ok := gitHubRetryAfter(err); ok {
		if delay < 2*time.Second {
			return 2 * time.Second
		}
		return delay
	}

	// Exponential backoff with a tight cap to prevent secondary rate-limit storms.
	if attempt < 1 {
		attempt = 1
	}
	delay := 1 * time.Second
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= time.Minute {
			return time.Minute
		}
	}
	return delay
}

func gitHubRetryAfter(err error) (time.Duration, bool) {
	var apiErr *gh.ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.Response == nil {
		return 0, false
	}

	if raw := strings.TrimSpace(apiErr.Response.Header.Get("Retry-After")); raw != "" {
		seconds, parseErr := strconv.Atoi(raw)
		if parseErr == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second, true
		}
	}

	if raw := strings.TrimSpace(apiErr.Response.Header.Get("X-RateLimit-Reset")); raw != "" {
		epoch, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr == nil && epoch > 0 {
			until := time.Until(time.Unix(epoch, 0))
			if until > 0 {
				return until, true
			}
		}
	}

	return 0, false
}

func isGitHubNotFound(err error) bool {
	return githubStatusCode(err) == http.StatusNotFound
}

func isGitHubConflict(err error) bool {
	return githubStatusCode(err) == http.StatusConflict
}

func isGitHubUnprocessable(err error) bool {
	return githubStatusCode(err) == http.StatusUnprocessableEntity
}

func githubStatusCode(err error) int {
	var apiErr *gh.ErrorResponse
	if errors.As(err, &apiErr) && apiErr.Response != nil {
		return apiErr.Response.StatusCode
	}
	return 0
}
