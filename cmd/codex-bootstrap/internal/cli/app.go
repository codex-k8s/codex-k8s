package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/codex-k8s/codex-k8s/cmd/codex-bootstrap/internal/envfile"
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
	case "bootstrap":
		return runBootstrap(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runValidate(args []string, stdout io.Writer, stderr io.Writer) int {
	var vars kvList
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)

	configPath := fs.String("config", "services.yaml", "Path to services.yaml")
	envName := fs.String("env", "ai-staging", "Environment name")
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
		fmt.Fprintf(stderr, "validate failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "ok project=%s env=%s namespace=%s services=%d\n",
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
	envName := fs.String("env", "ai-staging", "Environment name")
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
		fmt.Fprintf(stderr, "render failed: %v\n", err)
		return 1
	}

	if strings.TrimSpace(*outputPath) != "" {
		absOutput, err := filepath.Abs(*outputPath)
		if err != nil {
			fmt.Fprintf(stderr, "resolve output path: %v\n", err)
			return 1
		}
		if err := os.WriteFile(absOutput, rendered, 0o644); err != nil {
			fmt.Fprintf(stderr, "write output %q: %v\n", absOutput, err)
			return 1
		}
		fmt.Fprintf(stdout, "rendered project=%s env=%s namespace=%s -> %s\n", ctx.Project, ctx.Env, ctx.Namespace, absOutput)
		return 0
	}

	if _, err := stdout.Write(rendered); err != nil {
		fmt.Fprintf(stderr, "write output: %v\n", err)
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
	scriptPath := fs.String("script", "bootstrap/host/bootstrap_remote_staging.sh", "Path to host bootstrap script")
	envName := fs.String("env", "ai-staging", "Environment name for services.yaml validation")
	slotNo := fs.Int("slot", 0, "Slot number for context rendering")
	dryRun := fs.Bool("dry-run", false, "Print resolved plan without running host script")
	fs.Var(&vars, "var", "Template variable in KEY=VALUE format (repeatable)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	absEnv, err := filepath.Abs(*envPath)
	if err != nil {
		fmt.Fprintf(stderr, "resolve env-file path: %v\n", err)
		return 1
	}
	loadedEnv, err := envfile.Load(absEnv)
	if err != nil {
		fmt.Fprintf(stderr, "load env-file: %v\n", err)
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
		fmt.Fprintf(stderr, "validate services config before bootstrap: %v\n", err)
		return 1
	}

	absScript, err := filepath.Abs(*scriptPath)
	if err != nil {
		fmt.Fprintf(stderr, "resolve script path: %v\n", err)
		return 1
	}
	if _, err := os.Stat(absScript); err != nil {
		fmt.Fprintf(stderr, "bootstrap script is not available: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "project=%s env=%s namespace=%s\n", result.Context.Project, result.Context.Env, result.Context.Namespace)
	fmt.Fprintf(stdout, "env-file=%s\n", absEnv)
	fmt.Fprintf(stdout, "script=%s\n", absScript)
	fmt.Fprintf(stdout, "env-vars-loaded=%d\n", len(loadedEnv))
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
		fmt.Fprintf(stderr, "run bootstrap script: %v\n", err)
		return 1
	}
	return 0
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "codex-bootstrap - bootstrap helper for codex-k8s")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  codex-bootstrap validate [flags]")
	fmt.Fprintln(out, "  codex-bootstrap render [flags]")
	fmt.Fprintln(out, "  codex-bootstrap bootstrap [flags]")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  go run ./cmd/codex-bootstrap validate --config services.yaml --env ai-staging")
	fmt.Fprintln(out, "  go run ./cmd/codex-bootstrap render --config services.yaml --env ai-staging --output /tmp/rendered.yaml")
	fmt.Fprintln(out, "  go run ./cmd/codex-bootstrap bootstrap --env-file bootstrap/host/config-e2e-test.env --dry-run")
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
		fmt.Fprintln(out, "env-keys: <none>")
		return
	}
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fmt.Fprintln(out, "env-keys:")
	for _, key := range keys {
		fmt.Fprintf(out, "  - %s\n", key)
	}
}

func mustAbs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}
