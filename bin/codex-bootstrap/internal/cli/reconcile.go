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

const (
	defaultNamespaceEnvVar = "CODEXK8S_STAGING_NAMESPACE"
	defaultNamespaceName   = "default"
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

type actionExecutor struct {
	clientset   *kubernetes.Clientset
	envCfg      servicescfg.DeployEnvironment
	environment string
	namespace   string
	project     string
	rootDir     string
	env         map[string]string
	renderer    *servicescfg.Renderer
	stdout      io.Writer
}

func runReconcile(ctx context.Context, params reconcileParams) error {
	envCfg, ok := params.Config.Deploy.Environments[params.Environment]
	if !ok {
		return fmt.Errorf("deploy environment %q is not configured in services.yaml", params.Environment)
	}

	namespaceEnvVar := strings.TrimSpace(envCfg.Namespace.EnvVar)
	if namespaceEnvVar == "" {
		namespaceEnvVar = defaultNamespaceEnvVar
	}

	namespace := strings.TrimSpace(params.EnvMap[namespaceEnvVar])
	if namespace == "" {
		namespace = strings.TrimSpace(envCfg.Namespace.Default)
	}
	if namespace == "" {
		namespace = defaultNamespaceName
	}

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

	if strings.TrimSpace(envCfg.Namespace.Pattern) != "" && strings.TrimSpace(params.EnvMap[namespaceEnvVar]) == "" {
		resolvedNamespace, resolveErr := renderer.ResolveNamespace(envCfg.Namespace.Pattern)
		if resolveErr != nil {
			return fmt.Errorf("resolve namespace pattern: %w", resolveErr)
		}
		namespace = strings.TrimSpace(resolvedNamespace)
		if namespace == "" {
			return fmt.Errorf("resolved namespace is empty for environment %q", params.Environment)
		}
		params.EnvMap[namespaceEnvVar] = namespace
		renderer, err = servicescfg.NewRenderer(params.Config, params.RootDir, servicescfg.RenderContext{
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
	}

	restCfg, err := clientcfg.BuildRESTConfig(strings.TrimSpace(params.Kubeconfig))
	if err != nil {
		return fmt.Errorf("build kubernetes rest config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("create kubernetes clientset: %w", err)
	}

	executor := &actionExecutor{
		clientset:   clientset,
		envCfg:      envCfg,
		environment: params.Environment,
		namespace:   namespace,
		project:     params.Config.Project,
		rootDir:     params.RootDir,
		env:         params.EnvMap,
		renderer:    renderer,
		stdout:      params.Stdout,
	}

	for _, phase := range envCfg.ManifestPhases {
		if !phaseEnabled(phase, params.EnvMap) {
			continue
		}
		if _, err := fmt.Fprintf(params.Stdout, "Phase %s: start\n", phase.Name); err != nil {
			return err
		}
		for _, action := range phase.Actions {
			if err := executor.executeAction(ctx, action); err != nil {
				return fmt.Errorf("phase %s action %s: %w", phase.Name, action.Type, err)
			}
		}
	}

	if _, err := fmt.Fprintf(params.Stdout, "Reconcile completed for namespace %s\n", namespace); err != nil {
		return err
	}
	return nil
}

func (e *actionExecutor) executeAction(ctx context.Context, action servicescfg.DeployAction) error {
	switch strings.TrimSpace(action.Type) {
	case "set_defaults":
		e.setDefaults(action.Defaults)
		return nil
	case "set_from_env":
		e.setFromEnv(action.Assign)
		return nil
	case "import_secret":
		return e.importSecret(ctx, action)
	case "generate_hex":
		return e.generateHex(action.Generate)
	case "assert_env":
		return e.assertEnv(action.Keys)
	case "upsert_secret_from_env":
		return e.upsertSecretFromEnv(ctx, action)
	case "upsert_configmap_from_env":
		return e.upsertConfigMapFromEnv(ctx, action)
	case "upsert_configmap_from_dir":
		return e.upsertConfigMapFromDir(ctx, action)
	case "run_script":
		return e.runScript(ctx, action)
	case "apply_manifests":
		return e.applyManifests(ctx, action)
	case "delete_resources":
		return e.deleteResources(ctx, action.Resources)
	case "rollout_restart":
		return e.rolloutRestart(ctx, action.Resources)
	case "wait_targets":
		return e.waitTargets(ctx, action.WaitFor)
	default:
		return fmt.Errorf("unsupported action type %q", action.Type)
	}
}

func (e *actionExecutor) setDefaults(defaults map[string]string) {
	for key, rawValue := range defaults {
		if strings.TrimSpace(key) == "" {
			continue
		}
		if strings.TrimSpace(e.env[key]) != "" {
			continue
		}
		e.env[key] = expandWithEnv(rawValue, e.env)
	}
}

func (e *actionExecutor) setFromEnv(assignments []servicescfg.EnvAssign) {
	for _, assignment := range assignments {
		target := strings.TrimSpace(assignment.Target)
		source := strings.TrimSpace(assignment.Source)
		if target == "" || source == "" {
			continue
		}
		if strings.TrimSpace(e.env[target]) != "" {
			continue
		}
		sourceValue := strings.TrimSpace(e.env[source])
		if sourceValue != "" {
			e.env[target] = sourceValue
		}
	}
}

func (e *actionExecutor) importSecret(ctx context.Context, action servicescfg.DeployAction) error {
	secretName := strings.TrimSpace(action.Name)
	if secretName == "" {
		return fmt.Errorf("import_secret requires action.name")
	}
	values, err := getSecretData(ctx, e.clientset, e.namespace, secretName)
	if err != nil {
		return err
	}

	for _, key := range action.Keys {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			continue
		}
		if strings.TrimSpace(e.env[normalized]) != "" {
			continue
		}
		if value := strings.TrimSpace(values[normalized]); value != "" {
			e.env[normalized] = value
		}
	}
	for _, mapping := range action.Mappings {
		valueKey := strings.TrimSpace(mapping.Key)
		if valueKey == "" {
			continue
		}
		targetEnv := strings.TrimSpace(mapping.Env)
		if targetEnv == "" {
			targetEnv = valueKey
		}
		if strings.TrimSpace(e.env[targetEnv]) != "" {
			continue
		}
		if value := strings.TrimSpace(values[valueKey]); value != "" {
			e.env[targetEnv] = value
		}
	}
	return nil
}

func (e *actionExecutor) generateHex(rules []servicescfg.EnvGenerate) error {
	for _, rule := range rules {
		key := strings.TrimSpace(rule.Key)
		if key == "" {
			continue
		}
		currentValue := strings.TrimSpace(e.env[key])
		generateForEmpty := rule.IfEmpty || (!rule.IfEmpty && len(rule.RegenerateIfLengthNotIn) == 0)
		needsGeneration := generateForEmpty && currentValue == ""
		if !needsGeneration && currentValue != "" && len(rule.RegenerateIfLengthNotIn) > 0 && !intInSlice(len(currentValue), rule.RegenerateIfLengthNotIn) {
			needsGeneration = true
		}
		if !needsGeneration {
			continue
		}
		if rule.HexBytes <= 0 {
			return fmt.Errorf("generate_hex requires positive hex_bytes for key %q", key)
		}
		generated, err := randomHex(rule.HexBytes)
		if err != nil {
			return err
		}
		e.env[key] = generated
	}
	return nil
}

func (e *actionExecutor) assertEnv(keys []string) error {
	missing := make([]string, 0, len(keys))
	for _, key := range keys {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			continue
		}
		if strings.TrimSpace(e.env[normalized]) == "" {
			missing = append(missing, normalized)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env keys: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (e *actionExecutor) upsertSecretFromEnv(ctx context.Context, action servicescfg.DeployAction) error {
	secretName := strings.TrimSpace(action.Name)
	if secretName == "" {
		return fmt.Errorf("upsert_secret_from_env requires action.name")
	}
	data, err := e.collectData(action)
	if err != nil {
		return err
	}
	return upsertSecret(ctx, e.clientset, e.namespace, secretName, data)
}

func (e *actionExecutor) upsertConfigMapFromEnv(ctx context.Context, action servicescfg.DeployAction) error {
	configMapName := strings.TrimSpace(action.Name)
	if configMapName == "" {
		return fmt.Errorf("upsert_configmap_from_env requires action.name")
	}
	data, err := e.collectData(action)
	if err != nil {
		return err
	}
	return upsertConfigMap(ctx, e.clientset, e.namespace, configMapName, data)
}

func (e *actionExecutor) upsertConfigMapFromDir(ctx context.Context, action servicescfg.DeployAction) error {
	configMapName := strings.TrimSpace(action.Name)
	if configMapName == "" {
		return fmt.Errorf("upsert_configmap_from_dir requires action.name")
	}
	directory := strings.TrimSpace(action.Directory)
	if directory == "" {
		return fmt.Errorf("upsert_configmap_from_dir requires action.directory")
	}
	if !filepath.IsAbs(directory) {
		directory = filepath.Join(e.rootDir, directory)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read directory %q: %w", directory, err)
	}
	data := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			return fmt.Errorf("read file %q: %w", entry.Name(), readErr)
		}
		data[entry.Name()] = string(content)
	}
	return upsertConfigMap(ctx, e.clientset, e.namespace, configMapName, data)
}

func (e *actionExecutor) runScript(ctx context.Context, action servicescfg.DeployAction) error {
	scriptPath := strings.TrimSpace(action.Path)
	if scriptPath == "" {
		return fmt.Errorf("run_script requires action.path")
	}
	return runLocalScript(ctx, e.rootDir, scriptPath, e.env, e.stdout)
}

func (e *actionExecutor) applyManifests(ctx context.Context, action servicescfg.DeployAction) error {
	if len(action.Paths) == 0 {
		return fmt.Errorf("apply_manifests requires action.paths")
	}
	for _, manifestPath := range action.Paths {
		rendered, err := e.renderer.RenderFile(manifestPath, manifestTemplateData{
			Environment: e.environment,
			Namespace:   e.namespace,
			Project:     e.project,
			Env:         e.env,
		})
		if err != nil {
			return fmt.Errorf("render manifest %q: %w", manifestPath, err)
		}
		if err := kubectlApplyYAML(ctx, e.namespace, rendered, e.env, e.stdout); err != nil {
			return fmt.Errorf("apply manifest %q: %w", manifestPath, err)
		}
	}
	return nil
}

func (e *actionExecutor) deleteResources(ctx context.Context, resources []string) error {
	for _, resource := range resources {
		if err := kubectlDeleteResource(ctx, e.namespace, resource, e.env, e.stdout); err != nil {
			return err
		}
	}
	return nil
}

func (e *actionExecutor) rolloutRestart(ctx context.Context, resources []string) error {
	for _, resource := range resources {
		if err := kubectlRolloutRestart(ctx, e.namespace, resource, e.env, e.stdout); err != nil {
			return err
		}
	}
	return nil
}

func (e *actionExecutor) waitTargets(ctx context.Context, targets []servicescfg.WaitTarget) error {
	if len(targets) == 0 {
		return nil
	}
	waitRolloutEnabled := true
	waitRolloutKey := strings.TrimSpace(e.envCfg.WaitRolloutEnvVar)
	if waitRolloutKey != "" {
		waitRolloutEnabled = isTrue(e.env[waitRolloutKey])
	}
	for _, target := range targets {
		if strings.TrimSpace(target.Type) == "rollout" && !waitRolloutEnabled {
			continue
		}
		if err := waitForTarget(ctx, e.namespace, target, e.envCfg, e.env, e.stdout); err != nil {
			if target.Optional {
				if _, writeErr := fmt.Fprintf(e.stdout, "Optional wait failed for %s: %v\n", target.Resource, err); writeErr != nil {
					return writeErr
				}
				continue
			}
			return err
		}
	}
	return nil
}

func (e *actionExecutor) collectData(action servicescfg.DeployAction) (map[string]string, error) {
	data := make(map[string]string, len(action.Keys)+len(action.Mappings)+len(action.Values))
	for key, value := range action.Values {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			continue
		}
		data[normalized] = expandWithEnv(value, e.env)
	}
	for _, key := range action.Keys {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			continue
		}
		data[normalized] = strings.TrimSpace(e.env[normalized])
	}
	for _, mapping := range action.Mappings {
		targetKey := strings.TrimSpace(mapping.Key)
		if targetKey == "" {
			continue
		}
		sourceKey := strings.TrimSpace(mapping.Env)
		if sourceKey == "" {
			sourceKey = targetKey
		}
		value := strings.TrimSpace(e.env[sourceKey])
		if mapping.Required && value == "" {
			return nil, fmt.Errorf("required env %q for mapping %q is empty", sourceKey, targetKey)
		}
		data[targetKey] = value
	}
	return data, nil
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

func intInSlice(value int, allowed []int) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func isTrue(raw string) bool {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	switch normalized {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func expandWithEnv(raw string, env map[string]string) string {
	return os.Expand(raw, func(key string) string {
		if value, ok := env[key]; ok {
			return value
		}
		return "${" + key + "}"
	})
}
