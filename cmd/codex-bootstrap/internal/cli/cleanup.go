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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/cmd/codex-bootstrap/internal/envfile"
	"github.com/codex-k8s/codex-k8s/libs/go/k8s/clientcfg"
	gh "github.com/google/go-github/v82/github"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const defaultCleanupTimeout = 45 * time.Second

func runCleanup(args []string, stdout io.Writer, stderr io.Writer) int {
	var vars kvList
	fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
	fs.SetOutput(stderr)

	envPath := fs.String("env-file", "bootstrap/host/config.env", "Path to bootstrap env file")
	timeout := fs.Duration("timeout", defaultCleanupTimeout, "Timeout for each cleanup stage")
	workers := fs.Int("workers", defaultGitHubSyncWorkers, "Parallel workers for GitHub cleanup")
	namespace := fs.String("namespace", "", "Override production namespace for Kubernetes cleanup")
	kubeconfig := fs.String("kubeconfig", "", "Path to kubeconfig for Kubernetes cleanup")
	skipGitHub := fs.Bool("skip-github", false, "Skip GitHub cleanup")
	skipK8s := fs.Bool("skip-k8s", false, "Skip Kubernetes cleanup")
	yes := fs.Bool("yes", false, "Confirm destructive cleanup")
	fs.Var(&vars, "var", "Template variable in KEY=VALUE format (repeatable)")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if !*yes {
		writef(stderr, "cleanup failed: pass --yes to confirm destructive cleanup\n")
		return 2
	}
	if *timeout <= 0 {
		writef(stderr, "cleanup failed: --timeout must be positive\n")
		return 2
	}
	if *workers <= 0 {
		writef(stderr, "cleanup failed: --workers must be positive\n")
		return 2
	}

	absEnv, err := filepath.Abs(*envPath)
	if err != nil {
		writef(stderr, "cleanup failed: resolve env-file path: %v\n", err)
		return 1
	}
	values, err := envfile.Load(absEnv)
	if err != nil {
		writef(stderr, "cleanup failed: load env-file: %v\n", err)
		return 1
	}
	for key, value := range vars.Map() {
		values[key] = value
	}
	applyGitHubSyncDefaults(values)

	targetNamespace := strings.TrimSpace(*namespace)
	if targetNamespace == "" {
		targetNamespace = strings.TrimSpace(values["CODEXK8S_PRODUCTION_NAMESPACE"])
	}
	if targetNamespace == "" {
		targetNamespace = "codex-k8s-prod"
	}

	writef(stdout, "cleanup env-file=%s\n", absEnv)
	writef(stdout, "namespace=%s\n", targetNamespace)

	if !*skipGitHub {
		if err := cleanupGitHub(values, *timeout, *workers); err != nil {
			writef(stderr, "cleanup failed: github cleanup: %v\n", err)
			return 1
		}
	}
	if !*skipK8s {
		if err := cleanupKubernetes(strings.TrimSpace(*kubeconfig), targetNamespace); err != nil {
			writef(stderr, "cleanup failed: kubernetes cleanup: %v\n", err)
			return 1
		}
	}

	writeln(stdout, "cleanup completed")
	return 0
}

func cleanupGitHub(values map[string]string, timeout time.Duration, workers int) error {
	platformRepo, err := parseGitHubRepository(values["CODEXK8S_GITHUB_REPO"])
	if err != nil {
		return fmt.Errorf("CODEXK8S_GITHUB_REPO: %w", err)
	}
	firstProject := platformRepo
	if raw := strings.TrimSpace(values["CODEXK8S_FIRST_PROJECT_GITHUB_REPO"]); raw != "" {
		firstProject, err = parseGitHubRepository(raw)
		if err != nil {
			return fmt.Errorf("CODEXK8S_FIRST_PROJECT_GITHUB_REPO: %w", err)
		}
	}
	repos := []githubRepositoryRef{platformRepo}
	if firstProject.FullName != platformRepo.FullName {
		repos = append(repos, firstProject)
	}

	token := strings.TrimSpace(values["CODEXK8S_GITHUB_PAT"])
	if token == "" {
		return fmt.Errorf("CODEXK8S_GITHUB_PAT is required")
	}
	client := gh.NewClient(&http.Client{Timeout: timeout}).WithAuthToken(token)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	webhookURL := resolveWebhookURL(values)
	labels := collectGitHubLabels(values)
	labelNames := make([]string, 0, len(labels))
	for labelName := range labels {
		labelNames = append(labelNames, labelName)
	}
	sortStrings(labelNames)

	for _, repo := range repos {
		if err := deleteGitHubWebhook(ctx, client, repo, webhookURL); err != nil {
			return fmt.Errorf("delete webhook in %s: %w", repo.FullName, err)
		}
		if err := deleteGitHubLabels(ctx, client, repo, labelNames, workers); err != nil {
			return fmt.Errorf("delete labels in %s: %w", repo.FullName, err)
		}
	}
	if err := deleteGitHubRepoVariables(ctx, client, platformRepo, workers); err != nil {
		return fmt.Errorf("delete repository variables: %w", err)
	}
	if err := deleteGitHubRepoSecrets(ctx, client, platformRepo, workers); err != nil {
		return fmt.Errorf("delete repository secrets: %w", err)
	}
	return nil
}

func deleteGitHubWebhook(ctx context.Context, client *gh.Client, repo githubRepositoryRef, webhookURL string) error {
	webhookURL = strings.TrimSpace(webhookURL)
	if webhookURL == "" {
		return nil
	}
	hooks, _, err := client.Repositories.ListHooks(ctx, repo.Owner, repo.Name, &gh.ListOptions{PerPage: 100})
	if err != nil {
		return err
	}
	for _, hook := range hooks {
		if hook == nil || hook.Config == nil {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(hook.Config.GetURL()), webhookURL) {
			continue
		}
		if _, err := client.Repositories.DeleteHook(ctx, repo.Owner, repo.Name, hook.GetID()); err != nil && !isGitHubNotFound(err) {
			return err
		}
	}
	return nil
}

func deleteGitHubLabels(ctx context.Context, client *gh.Client, repo githubRepositoryRef, labels []string, workers int) error {
	operations := make([]githubOperation, 0, len(labels))
	for _, labelName := range labels {
		labelName := strings.TrimSpace(labelName)
		if labelName == "" {
			continue
		}
		operations = append(operations, githubOperation{
			Name: "delete label " + labelName,
			Run: func(ctx context.Context) error {
				_, err := client.Issues.DeleteLabel(ctx, repo.Owner, repo.Name, labelName)
				if err != nil && !isGitHubNotFound(err) {
					return err
				}
				return nil
			},
		})
	}
	return runGitHubOperations(ctx, workers, operations)
}

func deleteGitHubRepoVariables(ctx context.Context, client *gh.Client, repo githubRepositoryRef, workers int) error {
	names, err := listGitHubRepoVariableNames(ctx, client, repo)
	if err != nil {
		return err
	}
	allowedLegacy := make(map[string]struct{}, len(githubLegacyVariableKeys))
	for _, key := range githubLegacyVariableKeys {
		allowedLegacy[key] = struct{}{}
	}
	operations := make([]githubOperation, 0, len(names))
	for _, name := range names {
		name := strings.TrimSpace(name)
		if !strings.HasPrefix(name, "CODEXK8S_") {
			if _, ok := allowedLegacy[name]; !ok {
				continue
			}
		}
		operations = append(operations, githubOperation{
			Name: "delete variable " + name,
			Run: func(ctx context.Context) error {
				if _, err := client.Actions.DeleteRepoVariable(ctx, repo.Owner, repo.Name, name); err != nil && !isGitHubNotFound(err) {
					return err
				}
				return nil
			},
		})
	}
	return runGitHubOperations(ctx, workers, operations)
}

func deleteGitHubRepoSecrets(ctx context.Context, client *gh.Client, repo githubRepositoryRef, workers int) error {
	names, err := listGitHubRepoSecretNames(ctx, client, repo)
	if err != nil {
		return err
	}
	allowedLegacy := make(map[string]struct{}, len(githubLegacySecretKeys))
	for _, key := range githubLegacySecretKeys {
		allowedLegacy[key] = struct{}{}
	}
	operations := make([]githubOperation, 0, len(names))
	for _, name := range names {
		name := strings.TrimSpace(name)
		if !strings.HasPrefix(name, "CODEXK8S_") {
			if _, ok := allowedLegacy[name]; !ok {
				continue
			}
		}
		operations = append(operations, githubOperation{
			Name: "delete secret " + name,
			Run: func(ctx context.Context) error {
				if _, err := client.Actions.DeleteRepoSecret(ctx, repo.Owner, repo.Name, name); err != nil && !isGitHubNotFound(err) {
					return err
				}
				return nil
			},
		})
	}
	return runGitHubOperations(ctx, workers, operations)
}

func listGitHubRepoVariableNames(ctx context.Context, client *gh.Client, repo githubRepositoryRef) ([]string, error) {
	page := 1
	out := make([]string, 0)
	for {
		varsPage, resp, err := client.Actions.ListRepoVariables(ctx, repo.Owner, repo.Name, &gh.ListOptions{
			Page:    page,
			PerPage: 100,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range varsPage.Variables {
			if item == nil {
				continue
			}
			out = append(out, strings.TrimSpace(item.Name))
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	sortStrings(out)
	return out, nil
}

func listGitHubRepoSecretNames(ctx context.Context, client *gh.Client, repo githubRepositoryRef) ([]string, error) {
	page := 1
	out := make([]string, 0)
	for {
		secretsPage, resp, err := client.Actions.ListRepoSecrets(ctx, repo.Owner, repo.Name, &gh.ListOptions{
			Page:    page,
			PerPage: 100,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range secretsPage.Secrets {
			if item == nil {
				continue
			}
			out = append(out, strings.TrimSpace(item.Name))
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	sortStrings(out)
	return out, nil
}

func cleanupKubernetes(kubeconfigPath string, namespace string) error {
	restConfig, err := clientcfg.BuildRESTConfig(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("build kubernetes rest config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("build kubernetes clientset: %w", err)
	}
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("build kubernetes dynamic client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := clientset.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete namespace %s: %w", namespace, err)
	}

	issuerGVR := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "clusterissuers",
	}
	if err := dynamicClient.Resource(issuerGVR).Delete(ctx, "codex-k8s-letsencrypt", metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete clusterissuer codex-k8s-letsencrypt: %w", err)
	}

	waitCtx, waitCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer waitCancel()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		_, err := clientset.CoreV1().Namespaces().Get(waitCtx, namespace, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("wait namespace %s delete: %w", namespace, err)
		}
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("wait namespace %s delete: %w", namespace, waitCtx.Err())
		case <-ticker.C:
		}
	}
}

func runEmergencyDeploy(args []string, stdout io.Writer, stderr io.Writer) int {
	var vars kvList
	fs := flag.NewFlagSet("emergency-deploy", flag.ContinueOnError)
	fs.SetOutput(stderr)

	configPath := fs.String("config", "services.yaml", "Path to services.yaml")
	envPath := fs.String("env-file", "bootstrap/host/config.env", "Path to bootstrap env file")
	scriptPath := fs.String("script", "bootstrap/host/bootstrap_remote_production.sh", "Path to host bootstrap script")
	fullCleanup := fs.Bool("full-cleanup", false, "Run full cleanup before bootstrap")
	cleanupTimeout := fs.Duration("cleanup-timeout", defaultCleanupTimeout, "Cleanup timeout")
	fs.Var(&vars, "var", "Template variable in KEY=VALUE format (repeatable)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	absEnv, err := filepath.Abs(*envPath)
	if err != nil {
		writef(stderr, "emergency-deploy failed: resolve env-file path: %v\n", err)
		return 1
	}
	values, err := envfile.Load(absEnv)
	if err != nil {
		writef(stderr, "emergency-deploy failed: load env-file: %v\n", err)
		return 1
	}
	for key, value := range vars.Map() {
		values[key] = value
	}

	if *fullCleanup {
		writef(stdout, "emergency-deploy: run GitHub cleanup before deploy\n")
		cleanupArgs := []string{
			"--env-file", absEnv,
			"--yes",
			"--skip-k8s",
			"--timeout", cleanupTimeout.String(),
		}
		for _, entry := range vars {
			cleanupArgs = append(cleanupArgs, "--var", entry)
		}
		if code := runCleanup(cleanupArgs, stdout, stderr); code != 0 {
			return code
		}

		writef(stdout, "emergency-deploy: run remote Kubernetes cleanup\n")
		if err := runRemoteKubernetesCleanup(values, *cleanupTimeout); err != nil {
			writef(stderr, "emergency-deploy failed: remote cleanup: %v\n", err)
			return 1
		}
	}

	bootstrapArgs := []string{
		"--config", *configPath,
		"--env-file", absEnv,
		"--script", *scriptPath,
	}
	for _, entry := range vars {
		bootstrapArgs = append(bootstrapArgs, "--var", entry)
	}
	return runBootstrap(bootstrapArgs, stdout, stderr)
}

func runRemoteKubernetesCleanup(values map[string]string, timeout time.Duration) error {
	host := strings.TrimSpace(values["TARGET_HOST"])
	user := strings.TrimSpace(values["TARGET_ROOT_USER"])
	keyPath := expandUserPath(values["TARGET_ROOT_SSH_KEY"])
	port := strings.TrimSpace(values["TARGET_PORT"])
	if port == "" {
		port = "22"
	}
	if host == "" || user == "" || keyPath == "" {
		return fmt.Errorf("TARGET_HOST, TARGET_ROOT_USER and TARGET_ROOT_SSH_KEY are required")
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("TARGET_PORT must be numeric, got %q", port)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	remoteEnvPath := strings.TrimSpace(values["CODEXK8S_REMOTE_BOOTSTRAP_ENV_FILE"])
	if remoteEnvPath == "" {
		remoteEnvPath = "/root/codex-k8s-bootstrap/bootstrap.env"
	}
	remoteKubeconfig := strings.TrimSpace(values["CODEXK8S_REMOTE_KUBECONFIG"])
	if remoteKubeconfig == "" {
		remoteKubeconfig = "/etc/rancher/k3s/k3s.yaml"
	}
	command := "set -euo pipefail; " +
		"if [ -d /opt/codex-k8s ]; then " +
		"cd /opt/codex-k8s; " +
		"go run ./cmd/codex-bootstrap cleanup --env-file " + shellQuote(remoteEnvPath) +
		" --kubeconfig " + shellQuote(remoteKubeconfig) +
		" --yes --skip-github; " +
		"fi"

	sshArgs := []string{
		"-i", keyPath,
		"-p", port,
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		user + "@" + host,
		command,
	}
	run := exec.CommandContext(ctx, "ssh", sshArgs...)
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	if err := run.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("timeout after %s", timeout)
		}
		return err
	}
	return nil
}

func shellQuote(value string) string {
	quoted := strings.ReplaceAll(value, "'", "'\"'\"'")
	return "'" + quoted + "'"
}

func sortStrings(items []string) {
	if len(items) <= 1 {
		return
	}
	sort.Slice(items, func(i, j int) bool { return items[i] < items[j] })
}
