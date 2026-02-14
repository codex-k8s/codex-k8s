package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/codex-k8s/codex-k8s/bin/codex-bootstrap/internal/envfile"
	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
)

const (
	defaultServicesPath = "services.yaml"
	defaultEnvFilePath  = "bootstrap/host/config.env"
	defaultEnvironment  = "ai-staging"
)

type rootOptions struct {
	servicesPath string
	envFilePath  string
	environment  string
	kubeconfig   string
	noPrompt     bool
}

type loadedConfig struct {
	Config      *servicescfg.Config
	RootDir     string
	EnvFilePath string
	FileEnv     map[string]string
	RuntimeEnv  map[string]string
}

// ExecuteContext executes codex-bootstrap CLI.
func ExecuteContext(ctx context.Context) error {
	opts := rootOptions{}

	rootCmd := &cobra.Command{
		Use:   "codex-bootstrap",
		Short: "Declarative bootstrap/deploy CLI for codex-k8s",
	}
	rootCmd.PersistentFlags().StringVar(&opts.servicesPath, "services", defaultServicesPath, "Path to services.yaml")
	rootCmd.PersistentFlags().StringVar(&opts.envFilePath, "env-file", "", "Path to bootstrap env file (optional)")
	rootCmd.PersistentFlags().StringVar(&opts.environment, "environment", defaultEnvironment, "Target deploy environment from services.yaml")
	rootCmd.PersistentFlags().StringVar(&opts.kubeconfig, "kubeconfig", "", "Explicit kubeconfig path (optional)")
	rootCmd.PersistentFlags().BoolVar(&opts.noPrompt, "no-prompt", false, "Fail instead of interactive prompt for missing values")

	rootCmd.AddCommand(newValidateCommand(&opts))
	rootCmd.AddCommand(newInstallCommand(&opts))
	rootCmd.AddCommand(newReconcileCommand(&opts))

	return rootCmd.ExecuteContext(ctx)
}

func newValidateCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate services.yaml, partial templates and required env variables",
		RunE: func(cmd *cobra.Command, _ []string) error {
			loaded, err := loadConfigAndEnv(opts.servicesPath, opts.envFilePath)
			if err != nil {
				return err
			}
			renderer, err := servicescfg.NewRenderer(loaded.Config, loaded.RootDir, servicescfg.RenderContext{
				Env:        opts.environment,
				Project:    loaded.Config.Project,
				ProjectDir: loaded.RootDir,
				Now:        time.Now().UTC(),
				EnvMap:     loaded.RuntimeEnv,
			})
			if err != nil {
				return err
			}
			_ = renderer

			required := collectRequiredEnv(loaded.Config.Bootstrap.RequiredEnvValidate, loaded.Config.Bootstrap.RequiredEnvReconcile)
			missing := missingKeys(required, loaded.RuntimeEnv)
			if len(missing) > 0 {
				return fmt.Errorf("missing required env keys: %s", strings.Join(missing, ", "))
			}
			cmd.Printf("Validation passed for %s (%s)\n", opts.servicesPath, opts.environment)
			return nil
		},
	}
}

func newInstallCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Run first-install bootstrap flow",
		RunE: func(cmd *cobra.Command, _ []string) error {
			loaded, err := loadConfigAndEnv(opts.servicesPath, opts.envFilePath)
			if err != nil {
				return err
			}
			required := loaded.Config.Bootstrap.RequiredEnvInstall
			if len(required) == 0 {
				required = defaultInstallRequiredEnv
			}
			secretSet := toSet(loaded.Config.Bootstrap.SecretEnvKeys)
			if len(secretSet) == 0 {
				secretSet = toSet(defaultSecretEnvKeys)
			}
			promptedKeys, err := fillMissingValues(required, secretSet, loaded.RuntimeEnv, opts.noPrompt)
			if err != nil {
				return err
			}
			if len(promptedKeys) > 0 {
				persistPromptedValues(loaded.FileEnv, loaded.RuntimeEnv, promptedKeys)
				if err := envfile.Save(loaded.EnvFilePath, loaded.FileEnv); err != nil {
					return err
				}
			}

			scriptPath := strings.TrimSpace(loaded.Config.Bootstrap.InstallScriptPath)
			if scriptPath == "" {
				scriptPath = "bootstrap/host/bootstrap_remote_staging.sh"
			}
			if !filepath.IsAbs(scriptPath) {
				scriptPath = filepath.Join(loaded.RootDir, scriptPath)
			}
			if _, err := os.Stat(scriptPath); err != nil {
				return fmt.Errorf("install script %q is not available: %w", scriptPath, err)
			}

			cmd.Printf("Run install script: %s\n", scriptPath)
			execCmd := exec.CommandContext(cmd.Context(), "bash", scriptPath)
			execCmd.Env = mapToEnv(loaded.RuntimeEnv)
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			execCmd.Stdin = os.Stdin
			if err := execCmd.Run(); err != nil {
				return fmt.Errorf("execute install script: %w", err)
			}
			return nil
		},
	}
}

func newReconcileCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "reconcile",
		Short: "Apply declarative bootstrap/deploy state to Kubernetes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			loaded, err := loadConfigAndEnv(opts.servicesPath, opts.envFilePath)
			if err != nil {
				return err
			}
			required := loaded.Config.Bootstrap.RequiredEnvReconcile
			if len(required) == 0 {
				required = defaultReconcileRequiredEnv
			}
			secretSet := toSet(loaded.Config.Bootstrap.SecretEnvKeys)
			if len(secretSet) == 0 {
				secretSet = toSet(defaultSecretEnvKeys)
			}
			promptedKeys, err := fillMissingValues(required, secretSet, loaded.RuntimeEnv, opts.noPrompt)
			if err != nil {
				return err
			}
			if len(promptedKeys) > 0 {
				persistPromptedValues(loaded.FileEnv, loaded.RuntimeEnv, promptedKeys)
				if err := envfile.Save(loaded.EnvFilePath, loaded.FileEnv); err != nil {
					return err
				}
			}
			return runReconcile(cmd.Context(), reconcileParams{
				Config:      loaded.Config,
				RootDir:     loaded.RootDir,
				Environment: opts.environment,
				EnvMap:      loaded.RuntimeEnv,
				Kubeconfig:  opts.kubeconfig,
				Stdout:      cmd.OutOrStdout(),
			})
		},
	}
}

func loadConfigAndEnv(servicesPath string, envFilePath string) (*loadedConfig, error) {
	cfg, rootDir, err := servicescfg.Load(servicesPath)
	if err != nil {
		return nil, err
	}
	resolvedEnvFilePath := strings.TrimSpace(envFilePath)
	if resolvedEnvFilePath == "" {
		resolvedEnvFilePath = strings.TrimSpace(cfg.Bootstrap.EnvFilePath)
	}
	if resolvedEnvFilePath == "" {
		resolvedEnvFilePath = defaultEnvFilePath
	}
	if !filepath.IsAbs(resolvedEnvFilePath) {
		resolvedEnvFilePath = filepath.Join(rootDir, resolvedEnvFilePath)
	}

	fileVars, err := envfile.Load(resolvedEnvFilePath)
	if err != nil {
		return nil, err
	}

	merged := make(map[string]string, len(fileVars)+64)
	for key, value := range fileVars {
		merged[key] = value
	}
	for _, kv := range os.Environ() {
		idx := strings.IndexRune(kv, '=')
		if idx <= 0 {
			continue
		}
		key := kv[:idx]
		value := kv[idx+1:]
		if strings.TrimSpace(value) == "" {
			continue
		}
		merged[key] = value
	}
	return &loadedConfig{
		Config:      cfg,
		RootDir:     rootDir,
		EnvFilePath: resolvedEnvFilePath,
		FileEnv:     fileVars,
		RuntimeEnv:  merged,
	}, nil
}

func fillMissingValues(required []string, secretKeys map[string]struct{}, envMap map[string]string, noPrompt bool) ([]string, error) {
	promptedKeys := make([]string, 0, len(required))
	for _, key := range required {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			continue
		}
		if strings.TrimSpace(envMap[normalized]) != "" {
			continue
		}
		if noPrompt {
			return nil, fmt.Errorf("missing required env key: %s", normalized)
		}
		value, err := promptValue(normalized, isSecretKey(normalized, secretKeys))
		if err != nil {
			return nil, err
		}
		envMap[normalized] = value
		promptedKeys = append(promptedKeys, normalized)
	}
	return promptedKeys, nil
}

func persistPromptedValues(fileEnv map[string]string, runtimeEnv map[string]string, keys []string) {
	for _, key := range keys {
		if value, ok := runtimeEnv[key]; ok {
			fileEnv[key] = value
		}
	}
}

func promptValue(key string, secret bool) (string, error) {
	if secret {
		fmt.Printf("%s: ", key)
		bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", fmt.Errorf("read value for %s: %w", key, err)
		}
		return strings.TrimSpace(string(bytes)), nil
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", key)
	line, err := reader.ReadString('\n')
	if err != nil && !strings.Contains(err.Error(), "EOF") {
		return "", fmt.Errorf("read value for %s: %w", key, err)
	}
	return strings.TrimSpace(line), nil
}

func mapToEnv(vars map[string]string) []string {
	out := make([]string, 0, len(vars))
	for key, value := range vars {
		out = append(out, key+"="+value)
	}
	return out
}

func missingKeys(required []string, vars map[string]string) []string {
	out := make([]string, 0, len(required))
	for _, key := range required {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			continue
		}
		if strings.TrimSpace(vars[normalized]) == "" {
			out = append(out, normalized)
		}
	}
	return out
}

func toSet(keys []string) map[string]struct{} {
	out := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		normalized := strings.TrimSpace(key)
		if normalized != "" {
			out[normalized] = struct{}{}
		}
	}
	return out
}

func isSecretKey(key string, secretSet map[string]struct{}) bool {
	if _, ok := secretSet[key]; ok {
		return true
	}
	return strings.Contains(strings.ToLower(key), "token") ||
		strings.Contains(strings.ToLower(key), "secret") ||
		strings.Contains(strings.ToLower(key), "password") ||
		strings.Contains(strings.ToLower(key), "key")
}

func collectRequiredEnv(chunks ...[]string) []string {
	seen := make(map[string]struct{}, 32)
	out := make([]string, 0, 32)
	for _, chunk := range chunks {
		for _, item := range chunk {
			normalized := strings.TrimSpace(item)
			if normalized == "" {
				continue
			}
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
			out = append(out, normalized)
		}
	}
	return out
}

var defaultInstallRequiredEnv = []string{
	"TARGET_HOST",
	"TARGET_PORT",
	"TARGET_ROOT_USER",
	"TARGET_ROOT_SSH_KEY",
	"OPERATOR_USER",
	"OPERATOR_SSH_PUBKEY_PATH",
	"CODEXK8S_GITHUB_REPO",
	"CODEXK8S_GITHUB_PAT",
	"CODEXK8S_GIT_BOT_TOKEN",
	"CODEXK8S_GIT_BOT_USERNAME",
	"CODEXK8S_GIT_BOT_MAIL",
	"CODEXK8S_BOOTSTRAP_OWNER_EMAIL",
	"CODEXK8S_STAGING_DOMAIN",
	"CODEXK8S_GITHUB_OAUTH_CLIENT_ID",
	"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET",
	"CODEXK8S_LETSENCRYPT_EMAIL",
}

var defaultReconcileRequiredEnv = []string{
	"CODEXK8S_STAGING_NAMESPACE",
	"CODEXK8S_STAGING_DOMAIN",
	"CODEXK8S_GITHUB_PAT",
	"CODEXK8S_GIT_BOT_TOKEN",
	"CODEXK8S_GIT_BOT_USERNAME",
	"CODEXK8S_GIT_BOT_MAIL",
	"CODEXK8S_GITHUB_WEBHOOK_SECRET",
	"CODEXK8S_PUBLIC_BASE_URL",
	"CODEXK8S_BOOTSTRAP_OWNER_EMAIL",
	"CODEXK8S_GITHUB_OAUTH_CLIENT_ID",
	"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET",
	"CODEXK8S_JWT_SIGNING_KEY",
}

var defaultSecretEnvKeys = []string{
	"CODEXK8S_GITHUB_PAT",
	"CODEXK8S_OPENAI_API_KEY",
	"CODEXK8S_OPENAI_AUTH_FILE",
	"CODEXK8S_GIT_BOT_TOKEN",
	"CODEXK8S_PROJECT_DB_ADMIN_PASSWORD",
	"CODEXK8S_APP_SECRET_KEY",
	"CODEXK8S_TOKEN_ENCRYPTION_KEY",
	"CODEXK8S_MCP_TOKEN_SIGNING_KEY",
	"CODEXK8S_GITHUB_WEBHOOK_SECRET",
	"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET",
	"CODEXK8S_JWT_SIGNING_KEY",
	"CODEXK8S_POSTGRES_PASSWORD",
}
