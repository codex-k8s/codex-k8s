package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/codex-k8s/codex-k8s/libs/go/k8s/clientcfg"
	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
)

type reconcileParams struct {
	Config      *servicescfg.Config
	RootDir     string
	Environment string
	EnvMap      map[string]string
	Kubeconfig  string
	Stdout      io.Writer
}

type manifestTemplateData struct {
	Environment string
	Namespace   string
	Project     string
	Env         map[string]string
}

func runReconcile(ctx context.Context, params reconcileParams) error {
	envCfg, ok := params.Config.Deploy.Environments[params.Environment]
	if !ok {
		return fmt.Errorf("deploy environment %q is not configured in services.yaml", params.Environment)
	}
	namespaceEnv := strings.TrimSpace(envCfg.NamespaceEnvVar)
	if namespaceEnv == "" {
		namespaceEnv = "CODEXK8S_STAGING_NAMESPACE"
	}
	namespace := strings.TrimSpace(params.EnvMap[namespaceEnv])
	if namespace == "" {
		namespace = "codex-k8s-ai-staging"
		params.EnvMap[namespaceEnv] = namespace
	}

	applyDeployDefaults(params.EnvMap)
	renderer, err := servicescfg.NewRenderer(params.Config, params.RootDir, servicescfg.RenderContext{
		Env:        params.Environment,
		Namespace:  namespace,
		Project:    params.Config.Project,
		ProjectDir: params.RootDir,
		Now:        time.Now().UTC(),
		EnvMap:     params.EnvMap,
	})
	if err != nil {
		return err
	}

	restCfg, err := clientcfg.BuildRESTConfig(strings.TrimSpace(params.Kubeconfig))
	if err != nil {
		return fmt.Errorf("build kubernetes rest config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("create kubernetes clientset: %w", err)
	}

	if err := hydrateFromExistingSecrets(ctx, clientset, namespace, envCfg, params.EnvMap); err != nil {
		return err
	}
	if err := ensurePostgresSecret(ctx, clientset, namespace, envCfg, params.EnvMap); err != nil {
		return err
	}
	if err := ensureRuntimeSecret(ctx, clientset, namespace, envCfg, params.EnvMap); err != nil {
		return err
	}
	if err := ensureResourcesConfigMap(ctx, clientset, namespace, envCfg, params.EnvMap); err != nil {
		return err
	}
	if err := ensureLabelCatalogConfigMap(ctx, clientset, namespace, envCfg, params.EnvMap); err != nil {
		return err
	}
	if err := ensureOAuthSecret(ctx, clientset, namespace, envCfg, params.EnvMap); err != nil {
		return err
	}
	if err := ensureMigrationsConfigMap(ctx, clientset, namespace, envCfg, params.RootDir); err != nil {
		return err
	}

	if script := strings.TrimSpace(envCfg.NetworkPolicyScript); script != "" {
		if err := runLocalScript(ctx, params.RootDir, script, params.EnvMap, params.Stdout); err != nil {
			return fmt.Errorf("run network policy baseline script: %w", err)
		}
	}

	for _, phase := range envCfg.ManifestPhases {
		if !phaseEnabled(phase, params.EnvMap) {
			continue
		}
		if _, err := fmt.Fprintf(params.Stdout, "Phase %s: start\n", phase.Name); err != nil {
			return err
		}
		for _, resource := range phase.PreDelete {
			if err := kubectlDeleteResource(ctx, namespace, resource, params.EnvMap, params.Stdout); err != nil {
				return fmt.Errorf("phase %s pre-delete %s: %w", phase.Name, resource, err)
			}
		}
		for _, manifestPath := range phase.Manifests {
			manifestData := manifestTemplateData{
				Environment: params.Environment,
				Namespace:   namespace,
				Project:     params.Config.Project,
				Env:         params.EnvMap,
			}
			rendered, err := renderer.RenderFile(manifestPath, manifestData)
			if err != nil {
				return fmt.Errorf("phase %s render manifest %s: %w", phase.Name, manifestPath, err)
			}
			if err := kubectlApplyYAML(ctx, namespace, rendered, params.EnvMap, params.Stdout); err != nil {
				return fmt.Errorf("phase %s apply manifest %s: %w", phase.Name, manifestPath, err)
			}
		}
		for _, resource := range phase.RolloutRestart {
			if err := kubectlRolloutRestart(ctx, namespace, resource, params.EnvMap, params.Stdout); err != nil {
				return fmt.Errorf("phase %s rollout restart %s: %w", phase.Name, resource, err)
			}
		}
		waitEnabled := strings.EqualFold(strings.TrimSpace(params.EnvMap[envCfg.WaitRolloutEnvVar]), "true") || strings.TrimSpace(envCfg.WaitRolloutEnvVar) == ""
		for _, wait := range phase.WaitFor {
			if !waitEnabled && wait.Type == "rollout" {
				continue
			}
			if err := waitForTarget(ctx, namespace, wait, envCfg, params.EnvMap, params.Stdout); err != nil {
				if wait.Optional {
					if _, writeErr := fmt.Fprintf(params.Stdout, "Phase %s optional wait failed (%s): %v\n", phase.Name, wait.Resource, err); writeErr != nil {
						return writeErr
					}
					continue
				}
				return fmt.Errorf("phase %s wait %s: %w", phase.Name, wait.Resource, err)
			}
		}
	}

	if _, err := fmt.Fprintf(params.Stdout, "Reconcile completed for namespace %s\n", namespace); err != nil {
		return err
	}
	return nil
}

func phaseEnabled(phase servicescfg.DeployPhase, envMap map[string]string) bool {
	key := strings.TrimSpace(phase.EnabledWhenEnv)
	if key == "" {
		return true
	}
	expected := strings.TrimSpace(phase.EnabledWhenEquals)
	actual := strings.TrimSpace(envMap[key])
	if expected == "" {
		return actual != ""
	}
	return strings.EqualFold(actual, expected)
}

func applyDeployDefaults(env map[string]string) {
	defaults := map[string]string{
		"CODEXK8S_POSTGRES_DB":                             "codex_k8s",
		"CODEXK8S_POSTGRES_USER":                           "codex_k8s",
		"CODEXK8S_PROJECT_DB_ADMIN_HOST":                   "postgres",
		"CODEXK8S_PROJECT_DB_ADMIN_PORT":                   "5432",
		"CODEXK8S_PROJECT_DB_ADMIN_SSLMODE":                "disable",
		"CODEXK8S_PROJECT_DB_ADMIN_DATABASE":               "postgres",
		"CODEXK8S_WAIT_ROLLOUT":                            "true",
		"CODEXK8S_ROLLOUT_TIMEOUT":                         "1800s",
		"CODEXK8S_WORKER_RUN_NAMESPACE_PREFIX":             "codex-issue",
		"CODEXK8S_WORKER_RUN_NAMESPACE_CLEANUP":            "true",
		"CODEXK8S_WORKER_RUN_SERVICE_ACCOUNT":              "codex-runner",
		"CODEXK8S_WORKER_RUN_ROLE_NAME":                    "codex-runner",
		"CODEXK8S_WORKER_RUN_ROLE_BINDING_NAME":            "codex-runner",
		"CODEXK8S_WORKER_RUN_RESOURCE_QUOTA_NAME":          "codex-run-quota",
		"CODEXK8S_WORKER_RUN_LIMIT_RANGE_NAME":             "codex-run-limits",
		"CODEXK8S_WORKER_RUN_CREDENTIALS_SECRET_NAME":      "codex-run-credentials",
		"CODEXK8S_WORKER_RUN_QUOTA_PODS":                   "20",
		"CODEXK8S_WORKER_RUN_QUOTA_REQUESTS_CPU":           "6",
		"CODEXK8S_WORKER_RUN_QUOTA_REQUESTS_MEMORY":        "24Gi",
		"CODEXK8S_WORKER_RUN_QUOTA_LIMITS_CPU":             "8",
		"CODEXK8S_WORKER_RUN_QUOTA_LIMITS_MEMORY":          "32Gi",
		"CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_REQUEST_CPU":    "4",
		"CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_REQUEST_MEMORY": "16Gi",
		"CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_CPU":            "6",
		"CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_MEMORY":         "24Gi",
		"CODEXK8S_RUN_INTAKE_LABEL":                        "run:intake",
		"CODEXK8S_RUN_INTAKE_REVISE_LABEL":                 "run:intake:revise",
		"CODEXK8S_RUN_VISION_LABEL":                        "run:vision",
		"CODEXK8S_RUN_VISION_REVISE_LABEL":                 "run:vision:revise",
		"CODEXK8S_RUN_PRD_LABEL":                           "run:prd",
		"CODEXK8S_RUN_PRD_REVISE_LABEL":                    "run:prd:revise",
		"CODEXK8S_RUN_ARCH_LABEL":                          "run:arch",
		"CODEXK8S_RUN_ARCH_REVISE_LABEL":                   "run:arch:revise",
		"CODEXK8S_RUN_DESIGN_LABEL":                        "run:design",
		"CODEXK8S_RUN_DESIGN_REVISE_LABEL":                 "run:design:revise",
		"CODEXK8S_RUN_PLAN_LABEL":                          "run:plan",
		"CODEXK8S_RUN_PLAN_REVISE_LABEL":                   "run:plan:revise",
		"CODEXK8S_RUN_DEV_LABEL":                           "run:dev",
		"CODEXK8S_RUN_DEV_REVISE_LABEL":                    "run:dev:revise",
		"CODEXK8S_RUN_DEBUG_LABEL":                         "run:debug",
		"CODEXK8S_RUN_DOC_AUDIT_LABEL":                     "run:doc-audit",
		"CODEXK8S_RUN_QA_LABEL":                            "run:qa",
		"CODEXK8S_RUN_RELEASE_LABEL":                       "run:release",
		"CODEXK8S_RUN_POSTDEPLOY_LABEL":                    "run:postdeploy",
		"CODEXK8S_RUN_OPS_LABEL":                           "run:ops",
		"CODEXK8S_RUN_SELF_IMPROVE_LABEL":                  "run:self-improve",
		"CODEXK8S_RUN_RETHINK_LABEL":                       "run:rethink",
		"CODEXK8S_MODE_DISCUSSION_LABEL":                   "mode:discussion",
		"CODEXK8S_STATE_BLOCKED_LABEL":                     "state:blocked",
		"CODEXK8S_STATE_IN_REVIEW_LABEL":                   "state:in-review",
		"CODEXK8S_STATE_APPROVED_LABEL":                    "state:approved",
		"CODEXK8S_STATE_SUPERSEDED_LABEL":                  "state:superseded",
		"CODEXK8S_STATE_ABANDONED_LABEL":                   "state:abandoned",
		"CODEXK8S_NEED_INPUT_LABEL":                        "need:input",
		"CODEXK8S_NEED_PM_LABEL":                           "need:pm",
		"CODEXK8S_NEED_SA_LABEL":                           "need:sa",
		"CODEXK8S_NEED_QA_LABEL":                           "need:qa",
		"CODEXK8S_NEED_SRE_LABEL":                          "need:sre",
		"CODEXK8S_NEED_EM_LABEL":                           "need:em",
		"CODEXK8S_NEED_KM_LABEL":                           "need:km",
		"CODEXK8S_NEED_REVIEWER_LABEL":                     "need:reviewer",
		"CODEXK8S_VITE_DEV_UPSTREAM":                       "http://codex-k8s-web-console:5173",
	}
	for key, value := range defaults {
		if strings.TrimSpace(env[key]) == "" {
			env[key] = value
		}
	}
	if strings.TrimSpace(env["CODEXK8S_PROJECT_DB_ADMIN_USER"]) == "" {
		env["CODEXK8S_PROJECT_DB_ADMIN_USER"] = env["CODEXK8S_POSTGRES_USER"]
	}
	if strings.TrimSpace(env["CODEXK8S_PROJECT_DB_ADMIN_PASSWORD"]) == "" {
		env["CODEXK8S_PROJECT_DB_ADMIN_PASSWORD"] = env["CODEXK8S_POSTGRES_PASSWORD"]
	}
}

func hydrateFromExistingSecrets(ctx context.Context, clientset *kubernetes.Clientset, namespace string, envCfg servicescfg.DeployEnvironment, env map[string]string) error {
	runtimeSecret := strings.TrimSpace(envCfg.RuntimeSecretName)
	if runtimeSecret == "" {
		runtimeSecret = "codex-k8s-runtime"
	}
	postgresSecret := strings.TrimSpace(envCfg.PostgresSecretName)
	if postgresSecret == "" {
		postgresSecret = "codex-k8s-postgres"
	}

	existingRuntime, err := getSecretData(ctx, clientset, namespace, runtimeSecret)
	if err != nil {
		return err
	}
	for _, key := range runtimeSecretKeys {
		if strings.TrimSpace(env[key]) == "" && strings.TrimSpace(existingRuntime[key]) != "" {
			env[key] = existingRuntime[key]
		}
	}

	existingPostgres, err := getSecretData(ctx, clientset, namespace, postgresSecret)
	if err != nil {
		return err
	}
	if strings.TrimSpace(env["CODEXK8S_POSTGRES_PASSWORD"]) == "" {
		env["CODEXK8S_POSTGRES_PASSWORD"] = strings.TrimSpace(existingPostgres["CODEXK8S_POSTGRES_PASSWORD"])
	}
	if strings.TrimSpace(env["CODEXK8S_POSTGRES_PASSWORD"]) == "" {
		generated, genErr := randomHex(24)
		if genErr != nil {
			return genErr
		}
		env["CODEXK8S_POSTGRES_PASSWORD"] = generated
	}
	if strings.TrimSpace(env["CODEXK8S_APP_SECRET_KEY"]) == "" {
		generated, genErr := randomHex(32)
		if genErr != nil {
			return genErr
		}
		env["CODEXK8S_APP_SECRET_KEY"] = generated
	}
	if strings.TrimSpace(env["CODEXK8S_TOKEN_ENCRYPTION_KEY"]) == "" {
		generated, genErr := randomHex(32)
		if genErr != nil {
			return genErr
		}
		env["CODEXK8S_TOKEN_ENCRYPTION_KEY"] = generated
	}
	return nil
}

func ensurePostgresSecret(ctx context.Context, clientset *kubernetes.Clientset, namespace string, envCfg servicescfg.DeployEnvironment, env map[string]string) error {
	secretName := strings.TrimSpace(envCfg.PostgresSecretName)
	if secretName == "" {
		secretName = "codex-k8s-postgres"
	}
	data := map[string]string{
		"CODEXK8S_POSTGRES_DB":       env["CODEXK8S_POSTGRES_DB"],
		"CODEXK8S_POSTGRES_USER":     env["CODEXK8S_POSTGRES_USER"],
		"CODEXK8S_POSTGRES_PASSWORD": env["CODEXK8S_POSTGRES_PASSWORD"],
	}
	return upsertSecret(ctx, clientset, namespace, secretName, data)
}

func ensureRuntimeSecret(ctx context.Context, clientset *kubernetes.Clientset, namespace string, envCfg servicescfg.DeployEnvironment, env map[string]string) error {
	secretName := strings.TrimSpace(envCfg.RuntimeSecretName)
	if secretName == "" {
		secretName = "codex-k8s-runtime"
	}
	data := make(map[string]string, len(runtimeSecretKeys))
	for _, key := range runtimeSecretKeys {
		data[key] = strings.TrimSpace(env[key])
	}
	return upsertSecret(ctx, clientset, namespace, secretName, data)
}

func ensureResourcesConfigMap(ctx context.Context, clientset *kubernetes.Clientset, namespace string, envCfg servicescfg.DeployEnvironment, env map[string]string) error {
	configMapName := strings.TrimSpace(envCfg.ResourcesConfigMap)
	if configMapName == "" {
		configMapName = "codex-k8s-deploy-resources"
	}
	data := make(map[string]string, len(resourcesConfigMapKeys))
	for _, key := range resourcesConfigMapKeys {
		data[key] = strings.TrimSpace(env[key])
	}
	return upsertConfigMap(ctx, clientset, namespace, configMapName, data)
}

func ensureLabelCatalogConfigMap(ctx context.Context, clientset *kubernetes.Clientset, namespace string, envCfg servicescfg.DeployEnvironment, env map[string]string) error {
	configMapName := strings.TrimSpace(envCfg.LabelCatalogConfigMap)
	if configMapName == "" {
		configMapName = "codex-k8s-label-catalog"
	}
	data := make(map[string]string, len(labelCatalogKeys))
	for _, key := range labelCatalogKeys {
		data[key] = strings.TrimSpace(env[key])
	}
	return upsertConfigMap(ctx, clientset, namespace, configMapName, data)
}

func ensureOAuthSecret(ctx context.Context, clientset *kubernetes.Clientset, namespace string, envCfg servicescfg.DeployEnvironment, env map[string]string) error {
	secretName := strings.TrimSpace(envCfg.OAuthSecretName)
	if secretName == "" {
		secretName = "codex-k8s-oauth2-proxy"
	}
	existing, err := getSecretData(ctx, clientset, namespace, secretName)
	if err != nil {
		return err
	}
	cookieSecret := strings.TrimSpace(existing["OAUTH2_PROXY_COOKIE_SECRET"])
	if len(cookieSecret) != 16 && len(cookieSecret) != 24 && len(cookieSecret) != 32 {
		generated, genErr := randomHex(16)
		if genErr != nil {
			return genErr
		}
		cookieSecret = generated
	}
	return upsertSecret(ctx, clientset, namespace, secretName, map[string]string{
		"OAUTH2_PROXY_CLIENT_ID":     env["CODEXK8S_GITHUB_OAUTH_CLIENT_ID"],
		"OAUTH2_PROXY_CLIENT_SECRET": env["CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET"],
		"OAUTH2_PROXY_COOKIE_SECRET": cookieSecret,
	})
}

func ensureMigrationsConfigMap(ctx context.Context, clientset *kubernetes.Clientset, namespace string, envCfg servicescfg.DeployEnvironment, rootDir string) error {
	configMapName := strings.TrimSpace(envCfg.MigrationsConfigMap)
	if configMapName == "" {
		configMapName = "codex-k8s-migrations"
	}
	migrationsDir := strings.TrimSpace(envCfg.MigrationsDirectory)
	if migrationsDir == "" {
		migrationsDir = "services/internal/control-plane/cmd/cli/migrations"
	}
	if !filepath.IsAbs(migrationsDir) {
		migrationsDir = filepath.Join(rootDir, migrationsDir)
	}
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations directory %q: %w", migrationsDir, err)
	}
	data := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
		if readErr != nil {
			return fmt.Errorf("read migration file %q: %w", entry.Name(), readErr)
		}
		data[entry.Name()] = string(content)
	}
	return upsertConfigMap(ctx, clientset, namespace, configMapName, data)
}

func upsertSecret(ctx context.Context, clientset *kubernetes.Clientset, namespace string, name string, values map[string]string) error {
	data := make(map[string][]byte, len(values))
	for key, value := range values {
		data[key] = []byte(value)
	}

	existing, err := clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			_, createErr := clientset.CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: name},
				Type:       corev1.SecretTypeOpaque,
				Data:       data,
			}, metav1.CreateOptions{})
			return createErr
		}
		return err
	}
	existing.Data = data
	_, err = clientset.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func upsertConfigMap(ctx context.Context, clientset *kubernetes.Clientset, namespace string, name string, values map[string]string) error {
	existing, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			_, createErr := clientset.CoreV1().ConfigMaps(namespace).Create(ctx, &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: name},
				Data:       values,
			}, metav1.CreateOptions{})
			return createErr
		}
		return err
	}
	existing.Data = values
	_, err = clientset.CoreV1().ConfigMaps(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func getSecretData(ctx context.Context, clientset *kubernetes.Clientset, namespace string, name string) (map[string]string, error) {
	secret, err := clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	out := make(map[string]string, len(secret.Data))
	for key, value := range secret.Data {
		out[key] = string(value)
	}
	return out, nil
}

func waitForTarget(ctx context.Context, namespace string, target servicescfg.WaitTarget, envCfg servicescfg.DeployEnvironment, env map[string]string, stdout io.Writer) error {
	timeoutKey := strings.TrimSpace(envCfg.RolloutTimeoutEnvVar)
	if timeoutKey == "" {
		timeoutKey = "CODEXK8S_ROLLOUT_TIMEOUT"
	}
	timeout := strings.TrimSpace(env[timeoutKey])
	if timeout == "" {
		timeout = "1800s"
	}
	if customTimeout := strings.TrimSpace(target.TimeoutEnv); customTimeout != "" {
		if resolved := strings.TrimSpace(env[customTimeout]); resolved != "" {
			timeout = resolved
		}
	}
	switch strings.TrimSpace(target.Type) {
	case "job-complete":
		return runCommandWithEnv(ctx, env, stdout, "kubectl", "-n", namespace, "wait", "--for=condition=complete", target.Resource, "--timeout="+timeout)
	case "rollout":
		return runCommandWithEnv(ctx, env, stdout, "kubectl", "-n", namespace, "rollout", "status", target.Resource, "--timeout="+timeout)
	default:
		return fmt.Errorf("unsupported wait type %q", target.Type)
	}
}

func kubectlDeleteResource(ctx context.Context, namespace string, resource string, env map[string]string, stdout io.Writer) error {
	return runCommandWithEnv(ctx, env, stdout, "kubectl", "-n", namespace, "delete", resource, "--ignore-not-found")
}

func kubectlApplyYAML(ctx context.Context, namespace string, content []byte, env map[string]string, stdout io.Writer) error {
	cmd := exec.CommandContext(ctx, "kubectl", "-n", namespace, "apply", "-f", "-")
	cmd.Env = mapToEnv(env)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = bytes.NewReader(content)
	return cmd.Run()
}

func kubectlRolloutRestart(ctx context.Context, namespace string, resource string, env map[string]string, stdout io.Writer) error {
	return runCommandWithEnv(ctx, env, stdout, "kubectl", "-n", namespace, "rollout", "restart", resource)
}

func runLocalScript(ctx context.Context, rootDir string, relativePath string, env map[string]string, stdout io.Writer) error {
	scriptPath := relativePath
	if !filepath.IsAbs(scriptPath) {
		scriptPath = filepath.Join(rootDir, relativePath)
	}
	return runCommandWithEnv(ctx, env, stdout, "bash", scriptPath)
}

func runCommandWithEnv(ctx context.Context, env map[string]string, stdout io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = mapToEnv(env)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func randomHex(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("random hex length must be positive")
	}
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

var runtimeSecretKeys = []string{
	"CODEXK8S_GITHUB_PAT",
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
	"CODEXK8S_PUBLIC_BASE_URL",
	"CODEXK8S_BOOTSTRAP_OWNER_EMAIL",
	"CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS",
	"CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS",
	"CODEXK8S_GITHUB_OAUTH_CLIENT_ID",
	"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET",
	"CODEXK8S_JWT_SIGNING_KEY",
	"CODEXK8S_JWT_TTL",
	"CODEXK8S_VITE_DEV_UPSTREAM",
}

var resourcesConfigMapKeys = []string{
	"CODEXK8S_API_GATEWAY_RESOURCES_REQUEST_CPU",
	"CODEXK8S_API_GATEWAY_RESOURCES_REQUEST_MEMORY",
	"CODEXK8S_API_GATEWAY_RESOURCES_LIMIT_CPU",
	"CODEXK8S_API_GATEWAY_RESOURCES_LIMIT_MEMORY",
	"CODEXK8S_CONTROL_PLANE_RESOURCES_REQUEST_CPU",
	"CODEXK8S_CONTROL_PLANE_RESOURCES_REQUEST_MEMORY",
	"CODEXK8S_CONTROL_PLANE_RESOURCES_LIMIT_CPU",
	"CODEXK8S_CONTROL_PLANE_RESOURCES_LIMIT_MEMORY",
	"CODEXK8S_WORKER_RESOURCES_REQUEST_CPU",
	"CODEXK8S_WORKER_RESOURCES_REQUEST_MEMORY",
	"CODEXK8S_WORKER_RESOURCES_LIMIT_CPU",
	"CODEXK8S_WORKER_RESOURCES_LIMIT_MEMORY",
	"CODEXK8S_WORKER_RUN_NAMESPACE_PREFIX",
	"CODEXK8S_WORKER_RUN_NAMESPACE_CLEANUP",
	"CODEXK8S_WORKER_RUN_SERVICE_ACCOUNT",
	"CODEXK8S_WORKER_RUN_ROLE_NAME",
	"CODEXK8S_WORKER_RUN_ROLE_BINDING_NAME",
	"CODEXK8S_WORKER_RUN_RESOURCE_QUOTA_NAME",
	"CODEXK8S_WORKER_RUN_LIMIT_RANGE_NAME",
	"CODEXK8S_WORKER_RUN_CREDENTIALS_SECRET_NAME",
	"CODEXK8S_WORKER_RUN_QUOTA_PODS",
	"CODEXK8S_WORKER_RUN_QUOTA_REQUESTS_CPU",
	"CODEXK8S_WORKER_RUN_QUOTA_REQUESTS_MEMORY",
	"CODEXK8S_WORKER_RUN_QUOTA_LIMITS_CPU",
	"CODEXK8S_WORKER_RUN_QUOTA_LIMITS_MEMORY",
	"CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_REQUEST_CPU",
	"CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_REQUEST_MEMORY",
	"CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_CPU",
	"CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_MEMORY",
	"CODEXK8S_WORKER_JOB_IMAGE",
	"CODEXK8S_WORKER_JOB_COMMAND",
	"CODEXK8S_GIT_BOT_USERNAME",
	"CODEXK8S_GIT_BOT_MAIL",
	"CODEXK8S_AGENT_DEFAULT_MODEL",
	"CODEXK8S_AGENT_DEFAULT_REASONING_EFFORT",
	"CODEXK8S_AGENT_DEFAULT_LOCALE",
	"CODEXK8S_AGENT_BASE_BRANCH",
	"CODEXK8S_WEB_CONSOLE_RESOURCES_REQUEST_CPU",
	"CODEXK8S_WEB_CONSOLE_RESOURCES_REQUEST_MEMORY",
	"CODEXK8S_WEB_CONSOLE_RESOURCES_LIMIT_CPU",
	"CODEXK8S_WEB_CONSOLE_RESOURCES_LIMIT_MEMORY",
	"CODEXK8S_KANIKO_RESOURCES_REQUEST_CPU",
	"CODEXK8S_KANIKO_RESOURCES_REQUEST_MEMORY",
	"CODEXK8S_KANIKO_RESOURCES_LIMIT_CPU",
	"CODEXK8S_KANIKO_RESOURCES_LIMIT_MEMORY",
	"CODEXK8S_KANIKO_CACHE_ENABLED",
	"CODEXK8S_KANIKO_CACHE_REPO",
	"CODEXK8S_KANIKO_CACHE_TTL",
	"CODEXK8S_KANIKO_CACHE_COMPRESSED",
	"CODEXK8S_KANIKO_MATRIX_MAX_PARALLEL",
}

var labelCatalogKeys = []string{
	"CODEXK8S_RUN_INTAKE_LABEL",
	"CODEXK8S_RUN_INTAKE_REVISE_LABEL",
	"CODEXK8S_RUN_VISION_LABEL",
	"CODEXK8S_RUN_VISION_REVISE_LABEL",
	"CODEXK8S_RUN_PRD_LABEL",
	"CODEXK8S_RUN_PRD_REVISE_LABEL",
	"CODEXK8S_RUN_ARCH_LABEL",
	"CODEXK8S_RUN_ARCH_REVISE_LABEL",
	"CODEXK8S_RUN_DESIGN_LABEL",
	"CODEXK8S_RUN_DESIGN_REVISE_LABEL",
	"CODEXK8S_RUN_PLAN_LABEL",
	"CODEXK8S_RUN_PLAN_REVISE_LABEL",
	"CODEXK8S_RUN_DEV_LABEL",
	"CODEXK8S_RUN_DEV_REVISE_LABEL",
	"CODEXK8S_RUN_DEBUG_LABEL",
	"CODEXK8S_RUN_DOC_AUDIT_LABEL",
	"CODEXK8S_RUN_QA_LABEL",
	"CODEXK8S_RUN_RELEASE_LABEL",
	"CODEXK8S_RUN_POSTDEPLOY_LABEL",
	"CODEXK8S_RUN_OPS_LABEL",
	"CODEXK8S_RUN_SELF_IMPROVE_LABEL",
	"CODEXK8S_RUN_RETHINK_LABEL",
	"CODEXK8S_MODE_DISCUSSION_LABEL",
	"CODEXK8S_STATE_BLOCKED_LABEL",
	"CODEXK8S_STATE_IN_REVIEW_LABEL",
	"CODEXK8S_STATE_APPROVED_LABEL",
	"CODEXK8S_STATE_SUPERSEDED_LABEL",
	"CODEXK8S_STATE_ABANDONED_LABEL",
	"CODEXK8S_NEED_INPUT_LABEL",
	"CODEXK8S_NEED_PM_LABEL",
	"CODEXK8S_NEED_SA_LABEL",
	"CODEXK8S_NEED_QA_LABEL",
	"CODEXK8S_NEED_SRE_LABEL",
	"CODEXK8S_NEED_EM_LABEL",
	"CODEXK8S_NEED_KM_LABEL",
	"CODEXK8S_NEED_REVIEWER_LABEL",
	"CODEXK8S_AI_MODEL_GPT_5_3_CODEX_LABEL",
	"CODEXK8S_AI_MODEL_GPT_5_3_CODEX_SPARK_LABEL",
	"CODEXK8S_AI_MODEL_GPT_5_2_CODEX_LABEL",
	"CODEXK8S_AI_MODEL_GPT_5_1_CODEX_MAX_LABEL",
	"CODEXK8S_AI_MODEL_GPT_5_2_LABEL",
	"CODEXK8S_AI_MODEL_GPT_5_1_CODEX_MINI_LABEL",
	"CODEXK8S_AI_REASONING_LOW_LABEL",
	"CODEXK8S_AI_REASONING_MEDIUM_LABEL",
	"CODEXK8S_AI_REASONING_HIGH_LABEL",
	"CODEXK8S_AI_REASONING_EXTRA_HIGH_LABEL",
}
