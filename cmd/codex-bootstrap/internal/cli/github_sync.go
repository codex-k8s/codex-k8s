package cli

import (
	"context"
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

	"github.com/codex-k8s/kodex/cmd/codex-bootstrap/internal/envfile"
	gh "github.com/google/go-github/v82/github"
)

const (
	defaultGitHubSyncTimeout      = 15 * time.Minute
	defaultGitHubSyncWorkers      = 2
	defaultGitHubWebhookEvents    = "push,pull_request,issues,issue_comment,pull_request_review,pull_request_review_comment"
	defaultGitHubLabelDescription = "kodex managed label"
	defaultGitHubLabelColor       = "1f6feb"
)

var githubRequiredSyncKeys = []string{
	"KODEX_GITHUB_REPO",
	"KODEX_GITHUB_PAT",
	"KODEX_GITHUB_WEBHOOK_SECRET",
	"KODEX_PRODUCTION_DOMAIN",
}

type githubRepositoryRef struct {
	Owner    string
	Name     string
	FullName string
}

func runGitHubSync(args []string, stdout io.Writer, stderr io.Writer) int {
	var vars kvList
	fs := flag.NewFlagSet("github-sync", flag.ContinueOnError)
	fs.SetOutput(stderr)

	envPath := fs.String("env-file", "bootstrap/host/config.env", "Path to bootstrap env file")
	timeout := fs.Duration("timeout", defaultGitHubSyncTimeout, "GitHub API timeout")
	workers := fs.Int("workers", defaultGitHubSyncWorkers, "Parallel workers for label sync")
	dryRun := fs.Bool("dry-run", false, "Print planned changes without applying them")
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

	platformRepo, err := parseGitHubRepository(values["KODEX_GITHUB_REPO"])
	if err != nil {
		writef(stderr, "github-sync failed: KODEX_GITHUB_REPO: %v\n", err)
		return 1
	}
	firstProjectRaw := strings.TrimSpace(values["KODEX_FIRST_PROJECT_GITHUB_REPO"])
	firstProjectRepo := platformRepo
	if firstProjectRaw != "" {
		firstProjectRepo, err = parseGitHubRepository(firstProjectRaw)
		if err != nil {
			writef(stderr, "github-sync failed: KODEX_FIRST_PROJECT_GITHUB_REPO: %v\n", err)
			return 1
		}
	}

	reposForWebhookAndLabels := []githubRepositoryRef{platformRepo}
	if firstProjectRepo.FullName != platformRepo.FullName {
		reposForWebhookAndLabels = append(reposForWebhookAndLabels, firstProjectRepo)
	}

	labels := collectGitHubLabels(values)
	webhookURL := resolveWebhookURL(values)
	webhookEvents := normalizeGitHubEvents(values["KODEX_GITHUB_WEBHOOK_EVENTS"])
	webhookSecret := strings.TrimSpace(values["KODEX_GITHUB_WEBHOOK_SECRET"])
	if webhookSecret == "" {
		writef(stderr, "github-sync failed: KODEX_GITHUB_WEBHOOK_SECRET is required\n")
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

	client := gh.NewClient(&http.Client{Timeout: *timeout}).WithAuthToken(strings.TrimSpace(values["KODEX_GITHUB_PAT"]))
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

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

	writeln(stdout, "github-sync completed")
	return 0
}

func applyGitHubSyncDefaults(values map[string]string) {
	setEnvDefault(values, "KODEX_PRODUCTION_NAMESPACE", "kodex-prod")
	setEnvDefault(values, "KODEX_PRODUCTION_DOMAIN", "platform.kodex.works")
	if strings.TrimSpace(values["KODEX_AI_DOMAIN"]) == "" {
		if productionDomain := strings.TrimSpace(values["KODEX_PRODUCTION_DOMAIN"]); productionDomain != "" {
			values["KODEX_AI_DOMAIN"] = "ai." + productionDomain
		}
	}
	setEnvDefault(values, "KODEX_INTERNAL_REGISTRY_SERVICE", "kodex-registry")
	setEnvDefault(values, "KODEX_INTERNAL_REGISTRY_PORT", "5000")
	setEnvDefault(values, "KODEX_INTERNAL_REGISTRY_STORAGE_SIZE", "20Gi")
	setEnvDefault(values, "KODEX_INTERNAL_REGISTRY_HOST", "127.0.0.1:"+strings.TrimSpace(values["KODEX_INTERNAL_REGISTRY_PORT"]))
	setEnvDefault(values, "KODEX_GITHUB_WEBHOOK_EVENTS", defaultGitHubWebhookEvents)
	setEnvDefault(values, "KODEX_GITHUB_WEBHOOK_URL", resolveWebhookURL(values))
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
	return githubRepositoryRef{Owner: owner, Name: repo, FullName: owner + "/" + repo}, nil
}

func collectGitHubLabels(values map[string]string) map[string]string {
	out := make(map[string]string)
	for key, value := range values {
		if !strings.HasSuffix(key, "_LABEL") {
			continue
		}
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		out[trimmed] = labelDescriptionForKey(key)
	}
	return out
}

func labelDescriptionForKey(key string) string {
	labelKey := strings.TrimPrefix(strings.TrimSpace(key), "KODEX_")
	labelKey = strings.TrimSuffix(labelKey, "_LABEL")
	labelKey = strings.ToLower(strings.ReplaceAll(labelKey, "_", " "))
	if labelKey == "" {
		return defaultGitHubLabelDescription
	}
	return "kodex managed label: " + labelKey
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
