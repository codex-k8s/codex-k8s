package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/cmd/codex-bootstrap/internal/envfile"
	gh "github.com/google/go-github/v82/github"
)

const (
	preflightDefaultTimeout   = 15 * time.Second
	preflightDefaultLabelName = "run:dev"
)

var preflightRequiredEnvKeys = []string{
	"TARGET_HOST",
	"TARGET_ROOT_USER",
	"TARGET_ROOT_SSH_KEY",
	"OPERATOR_USER",
	"OPERATOR_SSH_PUBKEY_PATH",
	"CODEXK8S_GITHUB_REPO",
	"CODEXK8S_GITHUB_PAT",
	"CODEXK8S_PRODUCTION_DOMAIN",
	"CODEXK8S_LETSENCRYPT_EMAIL",
	"CODEXK8S_PUBLIC_BASE_URL",
	"CODEXK8S_BOOTSTRAP_OWNER_EMAIL",
	"CODEXK8S_GITHUB_OAUTH_CLIENT_ID",
	"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET",
	"CODEXK8S_GIT_BOT_TOKEN",
}

func runPreflight(args []string, stdout io.Writer, stderr io.Writer) int {
	var vars kvList
	fs := flag.NewFlagSet("preflight", flag.ContinueOnError)
	fs.SetOutput(stderr)

	envPath := fs.String("env-file", "bootstrap/host/config.env", "Path to bootstrap env file")
	timeout := fs.Duration("timeout", preflightDefaultTimeout, "Timeout for each network check")
	skipSSH := fs.Bool("skip-ssh", false, "Skip SSH connectivity check")
	skipGitHub := fs.Bool("skip-github", false, "Skip GitHub API checks")
	fs.Var(&vars, "var", "Template variable in KEY=VALUE format (repeatable)")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *timeout <= 0 {
		writef(stderr, "preflight failed: --timeout must be positive\n")
		return 2
	}

	absEnv, err := filepath.Abs(*envPath)
	if err != nil {
		writef(stderr, "preflight failed: resolve env-file path: %v\n", err)
		return 1
	}

	loadedEnv, err := envfile.Load(absEnv)
	if err != nil {
		writef(stderr, "preflight failed: load env-file: %v\n", err)
		return 1
	}
	for key, value := range vars.Map() {
		loadedEnv[key] = value
	}

	failures := make([]string, 0)
	warnings := make([]string, 0)

	missing := missingRequiredKeys(loadedEnv, preflightRequiredEnvKeys)
	if len(missing) > 0 {
		failures = append(failures, fmt.Sprintf("missing required env keys: %s", strings.Join(missing, ", ")))
	}

	sshKey := expandUserPath(loadedEnv["TARGET_ROOT_SSH_KEY"])
	if sshKey != "" {
		if err := assertFileReadable(sshKey); err != nil {
			failures = append(failures, fmt.Sprintf("TARGET_ROOT_SSH_KEY: %v", err))
		}
	}
	operatorPubKey := expandUserPath(loadedEnv["OPERATOR_SSH_PUBKEY_PATH"])
	if operatorPubKey != "" {
		if err := assertFileReadable(operatorPubKey); err != nil {
			failures = append(failures, fmt.Sprintf("OPERATOR_SSH_PUBKEY_PATH: %v", err))
		}
	}

	if !*skipSSH {
		if err := checkSSHConnectivity(loadedEnv, *timeout); err != nil {
			failures = append(failures, fmt.Sprintf("ssh connectivity: %v", err))
		}
	}

	if !*skipGitHub {
		githubWarnings, err := checkGitHubAccess(loadedEnv, *timeout)
		if len(githubWarnings) > 0 {
			warnings = append(warnings, githubWarnings...)
		}
		if err != nil {
			failures = append(failures, fmt.Sprintf("github access: %v", err))
		}
	}

	if strings.TrimSpace(loadedEnv["CODEXK8S_OPENAI_API_KEY"]) == "" && strings.TrimSpace(loadedEnv["CODEXK8S_OPENAI_AUTH_FILE"]) == "" {
		warnings = append(warnings, "both CODEXK8S_OPENAI_API_KEY and CODEXK8S_OPENAI_AUTH_FILE are empty (agent runs will not have LLM credentials)")
	}

	writef(stdout, "preflight env-file=%s\n", absEnv)
	if len(failures) == 0 {
		writeln(stdout, "status: ok")
	} else {
		writeln(stdout, "status: failed")
	}

	if len(warnings) > 0 {
		writeln(stdout, "warnings:")
		for _, warning := range warnings {
			writef(stdout, "  - %s\n", warning)
		}
	}
	if len(failures) > 0 {
		writeln(stdout, "failures:")
		for _, failure := range failures {
			writef(stdout, "  - %s\n", failure)
		}
		return 1
	}

	return 0
}

func missingRequiredKeys(values map[string]string, required []string) []string {
	if len(required) == 0 {
		return nil
	}
	out := make([]string, 0)
	for _, key := range required {
		if strings.TrimSpace(values[key]) != "" {
			continue
		}
		out = append(out, key)
	}
	slices.Sort(out)
	return out
}

func expandUserPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	expanded := os.ExpandEnv(path)
	if strings.HasPrefix(expanded, "~"+string(os.PathSeparator)) || expanded == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			if expanded == "~" {
				expanded = home
			} else {
				expanded = filepath.Join(home, strings.TrimPrefix(expanded, "~"+string(os.PathSeparator)))
			}
		}
	}
	return expanded
}

func assertFileReadable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	return nil
}

func checkSSHConnectivity(values map[string]string, timeout time.Duration) error {
	host := strings.TrimSpace(values["TARGET_HOST"])
	user := strings.TrimSpace(values["TARGET_ROOT_USER"])
	keyPath := expandUserPath(values["TARGET_ROOT_SSH_KEY"])
	port := strings.TrimSpace(values["TARGET_PORT"])
	if port == "" {
		port = "22"
	}
	if host == "" || user == "" || keyPath == "" {
		return errors.New("TARGET_HOST, TARGET_ROOT_USER and TARGET_ROOT_SSH_KEY must be set for SSH check")
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("TARGET_PORT must be numeric, got %q", port)
	}

	timeoutSeconds := int(timeout.Seconds())
	if timeoutSeconds <= 0 {
		timeoutSeconds = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := []string{
		"-i", keyPath,
		"-p", port,
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=" + strconv.Itoa(timeoutSeconds),
		user + "@" + host,
		"true",
	}
	command := exec.CommandContext(ctx, "ssh", args...)
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("timeout after %s", timeout.String())
		}
		return err
	}
	return nil
}

func checkGitHubAccess(values map[string]string, timeout time.Duration) ([]string, error) {
	token := strings.TrimSpace(values["CODEXK8S_GITHUB_PAT"])
	repoFullName := strings.TrimSpace(values["CODEXK8S_GITHUB_REPO"])
	if token == "" || repoFullName == "" {
		return nil, errors.New("CODEXK8S_GITHUB_PAT and CODEXK8S_GITHUB_REPO must be set for GitHub checks")
	}

	owner, repo, err := splitRepositoryFullName(repoFullName)
	if err != nil {
		return nil, fmt.Errorf("CODEXK8S_GITHUB_REPO: %w", err)
	}

	httpClient := &http.Client{Timeout: timeout}
	client := gh.NewClient(httpClient).WithAuthToken(token)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if _, _, err := client.Repositories.Get(ctx, owner, repo); err != nil {
		return nil, fmt.Errorf("get repository %s: %w", repoFullName, err)
	}

	warnings := make([]string, 0)
	hooks, _, err := client.Repositories.ListHooks(ctx, owner, repo, &gh.ListOptions{PerPage: 100})
	if err != nil {
		return warnings, fmt.Errorf("list webhooks %s: %w", repoFullName, err)
	}

	labels, _, err := client.Issues.ListLabels(ctx, owner, repo, &gh.ListOptions{PerPage: 100})
	if err != nil {
		return warnings, fmt.Errorf("list labels %s: %w", repoFullName, err)
	}

	webhookURL := resolveWebhookURL(values)
	if webhookURL != "" && !hasWebhookURL(hooks, webhookURL) {
		warnings = append(warnings, fmt.Sprintf("webhook %q is not configured in %s yet", webhookURL, repoFullName))
	}

	expectedLabel := strings.TrimSpace(values["CODEXK8S_RUN_DEV_LABEL"])
	if expectedLabel == "" {
		expectedLabel = preflightDefaultLabelName
	}
	if !hasLabel(labels, expectedLabel) {
		warnings = append(warnings, fmt.Sprintf("label %q is not configured in %s yet", expectedLabel, repoFullName))
	}

	firstProjectRepo := strings.TrimSpace(values["CODEXK8S_FIRST_PROJECT_GITHUB_REPO"])
	if firstProjectRepo == "" {
		return warnings, nil
	}
	firstOwner, firstRepo, splitErr := splitRepositoryFullName(firstProjectRepo)
	if splitErr != nil {
		return warnings, fmt.Errorf("CODEXK8S_FIRST_PROJECT_GITHUB_REPO: %w", splitErr)
	}
	if _, _, err := client.Repositories.Get(ctx, firstOwner, firstRepo); err != nil {
		return warnings, fmt.Errorf("get first project repository %s: %w", firstProjectRepo, err)
	}
	if _, _, err := client.Repositories.ListHooks(ctx, firstOwner, firstRepo, &gh.ListOptions{PerPage: 1}); err != nil {
		return warnings, fmt.Errorf("list webhooks %s: %w", firstProjectRepo, err)
	}
	if _, _, err := client.Issues.ListLabels(ctx, firstOwner, firstRepo, &gh.ListOptions{PerPage: 1}); err != nil {
		return warnings, fmt.Errorf("list labels %s: %w", firstProjectRepo, err)
	}
	return warnings, nil
}

func splitRepositoryFullName(value string) (string, string, error) {
	trimmed := strings.TrimSpace(value)
	owner, name, ok := strings.Cut(trimmed, "/")
	if !ok {
		return "", "", fmt.Errorf("expected owner/repository, got %q", value)
	}
	owner = strings.TrimSpace(owner)
	name = strings.TrimSpace(name)
	if owner == "" || name == "" || strings.Contains(name, "/") {
		return "", "", fmt.Errorf("expected owner/repository, got %q", value)
	}
	return owner, name, nil
}

func resolveWebhookURL(values map[string]string) string {
	if explicit := strings.TrimSpace(values["CODEXK8S_GITHUB_WEBHOOK_URL"]); explicit != "" {
		return explicit
	}
	domain := strings.TrimSpace(values["CODEXK8S_PRODUCTION_DOMAIN"])
	if domain == "" {
		return ""
	}
	return fmt.Sprintf("https://%s/api/v1/webhooks/github", domain)
}

func hasWebhookURL(hooks []*gh.Hook, webhookURL string) bool {
	for _, hook := range hooks {
		if hook == nil || hook.Config == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(hook.Config.GetURL()), webhookURL) {
			return true
		}
	}
	return false
}

func hasLabel(labels []*gh.Label, label string) bool {
	for _, item := range labels {
		if item == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(item.GetName()), label) {
			return true
		}
	}
	return false
}
