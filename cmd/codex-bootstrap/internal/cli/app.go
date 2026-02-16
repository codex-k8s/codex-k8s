package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/cmd/codex-bootstrap/internal/envfile"
	"github.com/codex-k8s/codex-k8s/libs/go/manifesttpl"
	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
)

// Run executes codex-bootstrap CLI and returns process exit code.
func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return 0
	}

	switch args[0] {
	case "validate":
		return runValidate(args[1:], stdout, stderr)
	case "render":
		return runRender(args[1:], stdout, stderr)
	case "render-manifest":
		return runRenderManifest(args[1:], stdout, stderr)
	case "preflight":
		return runPreflight(args[1:], stdout, stderr)
	case "github-sync":
		return runGitHubSync(args[1:], stdout, stderr)
	case "sync-secrets":
		return runSyncSecrets(args[1:], stdout, stderr)
	case "cleanup":
		return runCleanup(args[1:], stdout, stderr)
	case "emergency-deploy":
		return runEmergencyDeploy(args[1:], stdout, stderr)
	case "bootstrap":
		return runBootstrap(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		writef(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runValidate(args []string, stdout io.Writer, stderr io.Writer) int {
	var vars kvList
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)

	configPath := fs.String("config", "services.yaml", "Path to services.yaml")
	envName := fs.String("env", "production", "Environment name")
	slotNo := fs.Int("slot", 0, "Slot number")
	fs.Var(&vars, "var", "Template variable in KEY=VALUE format (repeatable)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	result, err := servicescfg.Load(*configPath, servicescfg.LoadOptions{
		Env:  *envName,
		Slot: *slotNo,
		Vars: vars.Map(),
	})
	if err != nil {
		writef(stderr, "validate failed: %v\n", err)
		return 1
	}

	writef(stdout, "ok project=%s env=%s namespace=%s services=%d\n",
		result.Context.Project,
		result.Context.Env,
		result.Context.Namespace,
		len(result.Stack.Spec.Services),
	)
	return 0
}

func runRender(args []string, stdout io.Writer, stderr io.Writer) int {
	var vars kvList
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	fs.SetOutput(stderr)

	configPath := fs.String("config", "services.yaml", "Path to services.yaml")
	envName := fs.String("env", "production", "Environment name")
	slotNo := fs.Int("slot", 0, "Slot number")
	outputPath := fs.String("output", "", "Optional output path for rendered YAML")
	fs.Var(&vars, "var", "Template variable in KEY=VALUE format (repeatable)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	rendered, ctx, err := servicescfg.Render(*configPath, servicescfg.LoadOptions{
		Env:  *envName,
		Slot: *slotNo,
		Vars: vars.Map(),
	})
	if err != nil {
		writef(stderr, "render failed: %v\n", err)
		return 1
	}

	if strings.TrimSpace(*outputPath) != "" {
		absOutput, err := filepath.Abs(*outputPath)
		if err != nil {
			writef(stderr, "resolve output path: %v\n", err)
			return 1
		}
		if err := os.WriteFile(absOutput, rendered, 0o644); err != nil {
			writef(stderr, "write output %q: %v\n", absOutput, err)
			return 1
		}
		writef(stdout, "rendered project=%s env=%s namespace=%s -> %s\n", ctx.Project, ctx.Env, ctx.Namespace, absOutput)
		return 0
	}

	if _, err := stdout.Write(rendered); err != nil {
		writef(stderr, "write output: %v\n", err)
		return 1
	}
	return 0
}

func runRenderManifest(args []string, stdout io.Writer, stderr io.Writer) int {
	var vars kvList
	fs := flag.NewFlagSet("render-manifest", flag.ContinueOnError)
	fs.SetOutput(stderr)

	templatePath := fs.String("template", "", "Path to *.yaml.tpl manifest template")
	envPath := fs.String("env-file", "", "Optional path to env file with KEY=VALUE values")
	fs.Var(&vars, "var", "Template variable in KEY=VALUE format (repeatable)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	trimmedTemplatePath := strings.TrimSpace(*templatePath)
	if trimmedTemplatePath == "" {
		writef(stderr, "render-manifest failed: --template is required\n")
		return 2
	}

	raw, err := os.ReadFile(trimmedTemplatePath)
	if err != nil {
		writef(stderr, "render-manifest failed: read template %q: %v\n", trimmedTemplatePath, err)
		return 1
	}

	templateVars := make(map[string]string)
	trimmedEnvPath := strings.TrimSpace(*envPath)
	if trimmedEnvPath != "" {
		loadedEnv, loadErr := envfile.Load(trimmedEnvPath)
		if loadErr != nil {
			writef(stderr, "render-manifest failed: load env-file %q: %v\n", trimmedEnvPath, loadErr)
			return 1
		}
		for key, value := range loadedEnv {
			templateVars[key] = value
		}
	}
	for key, value := range vars.Map() {
		templateVars[key] = value
	}

	rendered, err := manifesttpl.Render(trimmedTemplatePath, raw, templateVars)
	if err != nil {
		writef(stderr, "render-manifest failed: %v\n", err)
		return 1
	}

	if _, err := stdout.Write(rendered); err != nil {
		writef(stderr, "render-manifest failed: write output: %v\n", err)
		return 1
	}
	return 0
}

func runBootstrap(args []string, stdout io.Writer, stderr io.Writer) int {
	var vars kvList
	fs := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
	fs.SetOutput(stderr)

	configPath := fs.String("config", "services.yaml", "Path to services.yaml")
	envPath := fs.String("env-file", "bootstrap/host/config.env", "Path to bootstrap env file")
	scriptPath := fs.String("script", "bootstrap/host/bootstrap_remote_production.sh", "Path to host bootstrap script")
	envName := fs.String("env", "production", "Environment name for services.yaml validation")
	slotNo := fs.Int("slot", 0, "Slot number for context rendering")
	dryRun := fs.Bool("dry-run", false, "Print resolved plan without running host script")
	skipRemoteDeploy := fs.Bool("skip-remote-deploy", false, "Skip remote GitHub sync and deploy pipeline after host provisioning")
	fs.Var(&vars, "var", "Template variable in KEY=VALUE format (repeatable)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	absEnv, err := filepath.Abs(*envPath)
	if err != nil {
		writef(stderr, "resolve env-file path: %v\n", err)
		return 1
	}
	loadedEnv, err := envfile.Load(absEnv)
	if err != nil {
		writef(stderr, "load env-file: %v\n", err)
		return 1
	}
	for key, value := range vars.Map() {
		loadedEnv[key] = value
	}

	result, err := servicescfg.Load(*configPath, servicescfg.LoadOptions{
		Env:  *envName,
		Slot: *slotNo,
		Vars: loadedEnv,
	})
	if err != nil {
		writef(stderr, "validate services config before bootstrap: %v\n", err)
		return 1
	}

	absScript, err := filepath.Abs(*scriptPath)
	if err != nil {
		writef(stderr, "resolve script path: %v\n", err)
		return 1
	}
	if _, err := os.Stat(absScript); err != nil {
		writef(stderr, "bootstrap script is not available: %v\n", err)
		return 1
	}

	writef(stdout, "project=%s env=%s namespace=%s\n", result.Context.Project, result.Context.Env, result.Context.Namespace)
	writef(stdout, "env-file=%s\n", absEnv)
	writef(stdout, "script=%s\n", absScript)
	writef(stdout, "env-vars-loaded=%d\n", len(loadedEnv))
	if *dryRun {
		printEnvKeys(stdout, loadedEnv)
		return 0
	}

	command := exec.Command("bash", absScript)
	command.Stdin = os.Stdin
	command.Stdout = stdout
	command.Stderr = stderr
	command.Env = mergeEnv(os.Environ(), loadedEnv)
	command.Env = append(command.Env, "CODEXK8S_BOOTSTRAP_CONFIG_FILE="+absEnv)
	command.Env = append(command.Env, "CODEXK8S_SERVICES_CONFIG="+mustAbs(*configPath))

	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		writef(stderr, "run bootstrap script: %v\n", err)
		return 1
	}
	if *skipRemoteDeploy {
		writeln(stdout, "remote deploy pipeline skipped")
		return 0
	}
	if err := runRemoteBootstrapDeployPipeline(loadedEnv, stdout, stderr); err != nil {
		writef(stderr, "run remote deploy pipeline: %v\n", err)
		return 1
	}
	return 0
}

func runRemoteBootstrapDeployPipeline(values map[string]string, stdout io.Writer, stderr io.Writer) error {
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

	remoteEnvPath := strings.TrimSpace(values["CODEXK8S_REMOTE_BOOTSTRAP_ENV_FILE"])
	if remoteEnvPath == "" {
		remoteEnvPath = "/root/codex-k8s-bootstrap/bootstrap.env"
	}
	remoteKubeconfig := strings.TrimSpace(values["CODEXK8S_REMOTE_KUBECONFIG"])
	if remoteKubeconfig == "" {
		remoteKubeconfig = "/etc/rancher/k3s/k3s.yaml"
	}
	githubSyncWorkers := strings.TrimSpace(values["CODEXK8S_GITHUB_SYNC_WORKERS"])
	if githubSyncWorkers == "" {
		githubSyncWorkers = "2"
	}
	githubSyncTimeout := strings.TrimSpace(values["CODEXK8S_GITHUB_SYNC_TIMEOUT"])
	if githubSyncTimeout == "" {
		githubSyncTimeout = "15m"
	}
	remoteTimeoutRaw := strings.TrimSpace(values["CODEXK8S_REMOTE_DEPLOY_TIMEOUT"])
	if remoteTimeoutRaw == "" {
		remoteTimeoutRaw = "90m"
	}
	remoteTimeout, err := time.ParseDuration(remoteTimeoutRaw)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_REMOTE_DEPLOY_TIMEOUT=%q: %w", remoteTimeoutRaw, err)
	}

	command := strings.Join([]string{
		"set -euo pipefail",
		"cd /opt/codex-k8s",
		"go run ./services/internal/control-plane/cmd/runtime-deploy --env-file " + shellQuote(remoteEnvPath) +
			" --kubeconfig " + shellQuote(remoteKubeconfig) +
			" --prerequisites-only" +
			" --timeout " + shellQuote(remoteTimeoutRaw),
		"go run ./cmd/codex-bootstrap sync-secrets --env-file " + shellQuote(remoteEnvPath) + " --kubeconfig " + shellQuote(remoteKubeconfig),
		"go run ./cmd/codex-bootstrap github-sync --env-file " + shellQuote(remoteEnvPath) + " --workers " + shellQuote(githubSyncWorkers) + " --timeout " + shellQuote(githubSyncTimeout),
		"go run ./services/internal/control-plane/cmd/runtime-deploy --env-file " + shellQuote(remoteEnvPath) +
			" --kubeconfig " + shellQuote(remoteKubeconfig) +
			" --repository-root /opt/codex-k8s" +
			" --services-config /opt/codex-k8s/services.yaml" +
			" --target-env production" +
			" --timeout " + shellQuote(remoteTimeoutRaw),
	}, "; ")
	writef(stdout, "remote deploy: host=%s user=%s env=%s\n", host, user, remoteEnvPath)

	ctx, cancel := context.WithTimeout(context.Background(), remoteTimeout)
	defer cancel()
	sshArgs := []string{
		"-i", keyPath,
		"-p", port,
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		user + "@" + host,
		command,
	}
	run := exec.CommandContext(ctx, "ssh", sshArgs...)
	run.Stdout = stdout
	run.Stderr = stderr
	if err := run.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("remote deploy timeout after %s", remoteTimeout)
		}
		return err
	}
	return nil
}

func printUsage(out io.Writer) {
	writeln(out, "codex-bootstrap - bootstrap helper for codex-k8s")
	writeln(out, "")
	writeln(out, "Usage:")
	writeln(out, "  codex-bootstrap validate [flags]")
	writeln(out, "  codex-bootstrap render [flags]")
	writeln(out, "  codex-bootstrap render-manifest [flags]")
	writeln(out, "  codex-bootstrap preflight [flags]")
	writeln(out, "  codex-bootstrap github-sync [flags]")
	writeln(out, "  codex-bootstrap sync-secrets [flags]")
	writeln(out, "  codex-bootstrap cleanup [flags]")
	writeln(out, "  codex-bootstrap emergency-deploy [flags]")
	writeln(out, "  codex-bootstrap bootstrap [flags]")
	writeln(out, "")
	writeln(out, "Examples:")
	writeln(out, "  go run ./cmd/codex-bootstrap validate --config services.yaml --env production")
	writeln(out, "  go run ./cmd/codex-bootstrap render --config services.yaml --env production --output /tmp/rendered.yaml")
	writeln(out, "  go run ./cmd/codex-bootstrap render-manifest --template deploy/base/namespace/namespace.yaml.tpl")
	writeln(out, "  go run ./cmd/codex-bootstrap preflight --env-file bootstrap/host/config.env")
	writeln(out, "  go run ./cmd/codex-bootstrap github-sync --env-file bootstrap/host/config.env")
	writeln(out, "  go run ./cmd/codex-bootstrap sync-secrets --env-file bootstrap/host/config.env")
	writeln(out, "  go run ./cmd/codex-bootstrap cleanup --env-file bootstrap/host/config.env --yes")
	writeln(out, "  go run ./cmd/codex-bootstrap emergency-deploy --env-file bootstrap/host/config.env --full-cleanup")
	writeln(out, "  go run ./cmd/codex-bootstrap bootstrap --env-file bootstrap/host/config-e2e-test.env --dry-run")
}

type kvList []string

func (k *kvList) String() string {
	return strings.Join(*k, ",")
}

func (k *kvList) Set(value string) error {
	if _, _, ok := strings.Cut(value, "="); !ok {
		return fmt.Errorf("expected KEY=VALUE, got %q", value)
	}
	*k = append(*k, value)
	return nil
}

func (k kvList) Map() map[string]string {
	if len(k) == 0 {
		return nil
	}
	out := make(map[string]string, len(k))
	for _, entry := range k {
		key, value, _ := strings.Cut(entry, "=")
		out[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return out
}

func mergeEnv(base []string, extras map[string]string) []string {
	if len(extras) == 0 {
		return append([]string(nil), base...)
	}

	combined := make(map[string]string, len(base)+len(extras))
	for _, item := range base {
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			continue
		}
		combined[key] = value
	}
	for key, value := range extras {
		combined[key] = value
	}

	out := make([]string, 0, len(combined))
	for key, value := range combined {
		out = append(out, key+"="+value)
	}
	sort.Strings(out)
	return out
}

func printEnvKeys(out io.Writer, env map[string]string) {
	if len(env) == 0 {
		writeln(out, "env-keys: <none>")
		return
	}
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	writeln(out, "env-keys:")
	for _, key := range keys {
		writef(out, "  - %s\n", key)
	}
}

func writef(out io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(out, format, args...)
}

func writeln(out io.Writer, args ...any) {
	_, _ = fmt.Fprintln(out, args...)
}

func mustAbs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}
