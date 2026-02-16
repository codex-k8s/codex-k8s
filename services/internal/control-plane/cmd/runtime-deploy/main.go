package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/registry"
	kubernetesclient "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/clients/kubernetes"
	runtimedeploydomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runtimedeploy"
)

const (
	defaultRuntimeDeployEnvPath         = "bootstrap/host/config.env"
	defaultRuntimeDeployTimeout         = 2 * time.Hour
	defaultRuntimeDeployRolloutTimeout  = 20 * time.Minute
	defaultRuntimeDeployIngressTimeout  = 20 * time.Minute
	defaultRuntimeDeployCertMgrTimeout  = 20 * time.Minute
	defaultRuntimeDeployKanikoTimeout   = 30 * time.Minute
	defaultRuntimeDeployWaitPollTimeout = 2 * time.Second
	defaultRegistryHTTPTimeout          = 15 * time.Second
	defaultIngressManifestURL           = "https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml"
	defaultCertManagerManifestURL       = "https://github.com/cert-manager/cert-manager/releases/download/v1.15.3/cert-manager.yaml"
	defaultPrerequisitesFieldManager    = "codex-k8s-bootstrap"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout *os.File, stderr *os.File) int {
	fs := flag.NewFlagSet("runtime-deploy", flag.ContinueOnError)
	fs.SetOutput(stderr)

	envPath := fs.String("env-file", defaultRuntimeDeployEnvPath, "Path to env-file with CODEXK8S_* variables")
	kubeconfigPath := fs.String("kubeconfig", "", "Optional kubeconfig path")
	timeout := fs.Duration("timeout", defaultRuntimeDeployTimeout, "Overall deploy timeout")
	prerequisitesOnly := fs.Bool("prerequisites-only", false, "Install/upgrade Kubernetes prerequisites (ingress-nginx, cert-manager) and exit")
	prerequisitesFieldManager := fs.String("prerequisites-field-manager", defaultPrerequisitesFieldManager, "Field manager for prerequisites server-side apply")
	ingressManifestURL := fs.String("ingress-manifest-url", "", "Ingress controller manifest URL")
	certManagerManifestURL := fs.String("cert-manager-manifest-url", "", "cert-manager manifest URL")
	ingressReadyTimeout := fs.Duration("ingress-ready-timeout", defaultRuntimeDeployIngressTimeout, "Ingress controller readiness timeout")
	certManagerReadyTimeout := fs.Duration("cert-manager-ready-timeout", defaultRuntimeDeployCertMgrTimeout, "cert-manager readiness timeout")
	ingressHostNetwork := fs.String("ingress-host-network", "", "Override ingress host network mode (true/false)")
	namespace := fs.String("namespace", "", "Override target namespace")
	targetEnv := fs.String("target-env", "production", "Target services.yaml environment")
	slotNo := fs.Int("slot", 0, "Slot number for services.yaml context")
	runID := fs.String("run-id", "", "Run identifier used for runtime deploy logs")
	runtimeMode := fs.String("runtime-mode", "full-env", "Runtime mode")
	repositoryFullName := fs.String("repository-full-name", "", "Repository full name (owner/name)")
	servicesConfigPath := fs.String("services-config", "services.yaml", "Path to services.yaml")
	repositoryRoot := fs.String("repository-root", ".", "Repository root path")
	buildRef := fs.String("build-ref", "", "Git ref/tag used for image tag rendering")
	deployOnly := fs.Bool("deploy-only", false, "Skip build stage and apply manifests only")
	rolloutTimeout := fs.Duration("rollout-timeout", defaultRuntimeDeployRolloutTimeout, "Workload rollout timeout")
	kanikoTimeout := fs.Duration("kaniko-timeout", defaultRuntimeDeployKanikoTimeout, "Kaniko job timeout")
	waitPoll := fs.Duration("wait-poll-interval", defaultRuntimeDeployWaitPollTimeout, "Task wait polling interval")
	fieldManager := fs.String("field-manager", "codex-k8s-control-plane", "Kubernetes server-side apply field manager")
	registryHost := fs.String("registry-host", "", "Internal registry host:port override")
	registryScheme := fs.String("registry-scheme", "", "Internal registry scheme override")
	registryTimeout := fs.Duration("registry-http-timeout", defaultRegistryHTTPTimeout, "Registry API timeout")
	registryCleanupKeepTags := fs.Int("registry-cleanup-keep-tags", 5, "Keep this many latest tags when cleaning stale images")
	kanikoLogTailLines := fs.Int64("kaniko-log-tail-lines", 200, "Tail lines collected from kaniko jobs")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *timeout <= 0 {
		writeRuntimeDeployError(stderr, "runtime-deploy failed: --timeout must be positive")
		return 2
	}
	if *ingressReadyTimeout <= 0 {
		writeRuntimeDeployError(stderr, "runtime-deploy failed: --ingress-ready-timeout must be positive")
		return 2
	}
	if *certManagerReadyTimeout <= 0 {
		writeRuntimeDeployError(stderr, "runtime-deploy failed: --cert-manager-ready-timeout must be positive")
		return 2
	}

	values, err := loadEnvFile(*envPath)
	if err != nil {
		writeRuntimeDeployError(stderr, "runtime-deploy failed: load env-file: %v", err)
		return 1
	}
	applyRuntimeDeployEnvDefaults(values)
	if strings.TrimSpace(values["CODEXK8S_INGRESS_READY_TIMEOUT"]) != "" && *ingressReadyTimeout == defaultRuntimeDeployIngressTimeout {
		parsedDuration, parseErr := time.ParseDuration(strings.TrimSpace(values["CODEXK8S_INGRESS_READY_TIMEOUT"]))
		if parseErr != nil {
			writeRuntimeDeployError(stderr, "runtime-deploy failed: parse CODEXK8S_INGRESS_READY_TIMEOUT: %v", parseErr)
			return 2
		}
		*ingressReadyTimeout = parsedDuration
	}
	if strings.TrimSpace(values["CODEXK8S_CERT_MANAGER_READY_TIMEOUT"]) != "" && *certManagerReadyTimeout == defaultRuntimeDeployCertMgrTimeout {
		parsedDuration, parseErr := time.ParseDuration(strings.TrimSpace(values["CODEXK8S_CERT_MANAGER_READY_TIMEOUT"]))
		if parseErr != nil {
			writeRuntimeDeployError(stderr, "runtime-deploy failed: parse CODEXK8S_CERT_MANAGER_READY_TIMEOUT: %v", parseErr)
			return 2
		}
		*certManagerReadyTimeout = parsedDuration
	}
	for key, value := range values {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		if setErr := os.Setenv(key, value); setErr != nil {
			writeRuntimeDeployError(stderr, "runtime-deploy failed: set env %s: %v", key, setErr)
			return 1
		}
	}

	targetNamespace := strings.TrimSpace(*namespace)
	if targetNamespace == "" {
		targetNamespace = strings.TrimSpace(values["CODEXK8S_PRODUCTION_NAMESPACE"])
	}
	if targetNamespace == "" {
		targetNamespace = "codex-k8s-prod"
	}

	k8sClient, err := kubernetesclient.NewClient(strings.TrimSpace(*kubeconfigPath))
	if err != nil {
		writeRuntimeDeployError(stderr, "runtime-deploy failed: init kubernetes client: %v", err)
		return 1
	}
	if *prerequisitesOnly {
		targetIngressManifestURL := strings.TrimSpace(*ingressManifestURL)
		if targetIngressManifestURL == "" {
			targetIngressManifestURL = strings.TrimSpace(values["CODEXK8S_INGRESS_MANIFEST_URL"])
		}
		if targetIngressManifestURL == "" {
			targetIngressManifestURL = defaultIngressManifestURL
		}

		targetCertManagerManifestURL := strings.TrimSpace(*certManagerManifestURL)
		if targetCertManagerManifestURL == "" {
			targetCertManagerManifestURL = strings.TrimSpace(values["CODEXK8S_CERT_MANAGER_MANIFEST_URL"])
		}
		if targetCertManagerManifestURL == "" {
			targetCertManagerManifestURL = defaultCertManagerManifestURL
		}

		hostNetworkRaw := strings.TrimSpace(*ingressHostNetwork)
		if hostNetworkRaw == "" {
			hostNetworkRaw = strings.TrimSpace(values["CODEXK8S_INGRESS_HOST_NETWORK"])
		}
		hostNetworkEnabled, parseErr := parseBoolDefault(hostNetworkRaw, true)
		if parseErr != nil {
			writeRuntimeDeployError(stderr, "runtime-deploy failed: --ingress-host-network: %v", parseErr)
			return 2
		}

		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		if err := installKubernetesPrerequisites(ctx, runtimePrerequisitesOptions{
			Values:                  values,
			Namespace:               targetNamespace,
			FieldManager:            strings.TrimSpace(*prerequisitesFieldManager),
			IngressManifestURL:      targetIngressManifestURL,
			CertManagerManifestURL:  targetCertManagerManifestURL,
			IngressReadyTimeout:     *ingressReadyTimeout,
			CertManagerReadyTimeout: *certManagerReadyTimeout,
			IngressHostNetwork:      hostNetworkEnabled,
			Client:                  k8sClient,
		}); err != nil {
			writeRuntimeDeployError(stderr, "runtime-deploy failed: install prerequisites: %v", err)
			return 1
		}
		_, _ = fmt.Fprintf(stdout, "runtime-deploy prerequisites completed namespace=%s\n", targetNamespace)
		return 0
	}

	targetRepository := strings.TrimSpace(*repositoryFullName)
	if targetRepository == "" {
		targetRepository = strings.TrimSpace(values["CODEXK8S_GITHUB_REPO"])
	}
	if targetRepository == "" {
		writeRuntimeDeployError(stderr, "runtime-deploy failed: repository-full-name is required")
		return 1
	}

	targetRunID := strings.TrimSpace(*runID)
	if targetRunID == "" {
		targetRunID = "bootstrap-" + time.Now().UTC().Format("20060102t150405")
	}
	targetBuildRef := strings.TrimSpace(*buildRef)
	if targetBuildRef == "" {
		targetBuildRef = strings.TrimSpace(values["CODEXK8S_BUILD_REF"])
	}
	if targetBuildRef == "" {
		targetBuildRef = "main"
	}

	targetRegistryHost := strings.TrimSpace(*registryHost)
	if targetRegistryHost == "" {
		targetRegistryHost = strings.TrimSpace(values["CODEXK8S_INTERNAL_REGISTRY_HOST"])
	}
	if targetRegistryHost == "" {
		targetRegistryHost = "codex-k8s-registry:5000"
	}
	targetRegistryScheme := strings.TrimSpace(*registryScheme)
	if targetRegistryScheme == "" {
		targetRegistryScheme = strings.TrimSpace(values["CODEXK8S_INTERNAL_REGISTRY_SCHEME"])
	}
	if targetRegistryScheme == "" {
		targetRegistryScheme = "http"
	}

	registryClient, err := registry.NewClient(targetRegistryScheme+"://"+targetRegistryHost, *registryTimeout)
	if err != nil {
		writeRuntimeDeployError(stderr, "runtime-deploy failed: init registry client: %v", err)
		return 1
	}

	logger := slog.New(slog.NewJSONHandler(stdout, nil))
	service, err := runtimedeploydomain.NewService(runtimedeploydomain.Config{
		ServicesConfigPath:      strings.TrimSpace(*servicesConfigPath),
		RepositoryRoot:          strings.TrimSpace(*repositoryRoot),
		RolloutTimeout:          *rolloutTimeout,
		KanikoTimeout:           *kanikoTimeout,
		WaitPollInterval:        *waitPoll,
		KanikoFieldManager:      strings.TrimSpace(*fieldManager),
		GitHubPAT:               strings.TrimSpace(values["CODEXK8S_GITHUB_PAT"]),
		RegistryCleanupKeepTags: *registryCleanupKeepTags,
		KanikoJobLogTailLines:   *kanikoLogTailLines,
	}, runtimedeploydomain.Dependencies{
		Kubernetes: runtimeDeployKubernetesAdapter{client: k8sClient},
		Tasks:      noopRuntimeDeployTaskRepository{},
		Registry:   registryClient,
		Logger:     logger,
	})
	if err != nil {
		writeRuntimeDeployError(stderr, "runtime-deploy failed: init runtime deploy service: %v", err)
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	result, err := service.ApplyNow(ctx, runtimedeploydomain.PrepareParams{
		RunID:              targetRunID,
		RuntimeMode:        strings.TrimSpace(*runtimeMode),
		Namespace:          targetNamespace,
		TargetEnv:          strings.TrimSpace(*targetEnv),
		SlotNo:             *slotNo,
		RepositoryFullName: targetRepository,
		ServicesYAMLPath:   strings.TrimSpace(*servicesConfigPath),
		BuildRef:           targetBuildRef,
		DeployOnly:         *deployOnly,
	})
	if err != nil {
		writeRuntimeDeployError(stderr, "runtime-deploy failed: %v", err)
		return 1
	}

	_, _ = fmt.Fprintf(stdout, "runtime-deploy completed run_id=%s namespace=%s target_env=%s\n", targetRunID, result.Namespace, result.TargetEnv)
	return 0
}

func writeRuntimeDeployError(stderr *os.File, format string, args ...any) {
	log.New(stderr, "", 0).Printf(format, args...)
}

type runtimePrerequisitesOptions struct {
	Values                  map[string]string
	Namespace               string
	FieldManager            string
	IngressManifestURL      string
	CertManagerManifestURL  string
	IngressReadyTimeout     time.Duration
	CertManagerReadyTimeout time.Duration
	IngressHostNetwork      bool
	Client                  *kubernetesclient.Client
}

func installKubernetesPrerequisites(ctx context.Context, opts runtimePrerequisitesOptions) error {
	if opts.Client == nil {
		return fmt.Errorf("kubernetes client is nil")
	}
	fieldManager := strings.TrimSpace(opts.FieldManager)
	if fieldManager == "" {
		fieldManager = defaultPrerequisitesFieldManager
	}

	productionNamespace := strings.TrimSpace(opts.Namespace)
	if productionNamespace == "" {
		productionNamespace = "codex-k8s-prod"
	}

	baseNamespacesManifest := fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
---
apiVersion: v1
kind: Namespace
metadata:
  name: cert-manager
---
apiVersion: v1
kind: Namespace
metadata:
  name: ingress-nginx
`, productionNamespace)
	if _, err := opts.Client.ApplyManifest(ctx, []byte(baseNamespacesManifest), "", fieldManager); err != nil {
		return fmt.Errorf("apply base namespaces manifest: %w", err)
	}

	ingressManifest, err := fetchManifest(ctx, opts.IngressManifestURL)
	if err != nil {
		return fmt.Errorf("fetch ingress manifest: %w", err)
	}
	if _, err := opts.Client.ApplyManifest(ctx, ingressManifest, "", fieldManager); err != nil {
		return fmt.Errorf("apply ingress manifest: %w", err)
	}

	if opts.IngressHostNetwork {
		if err := opts.Client.EnableIngressControllerHostNetwork(ctx, "ingress-nginx", "ingress-nginx-controller"); err != nil {
			return fmt.Errorf("enable ingress host-network: %w", err)
		}
	}

	if err := opts.Client.WaitForDeploymentReady(ctx, "ingress-nginx", "ingress-nginx-controller", opts.IngressReadyTimeout); err != nil {
		return fmt.Errorf("wait ingress-nginx-controller ready: %w", err)
	}

	certManagerManifest, err := fetchManifest(ctx, opts.CertManagerManifestURL)
	if err != nil {
		return fmt.Errorf("fetch cert-manager manifest: %w", err)
	}
	if _, err := opts.Client.ApplyManifest(ctx, certManagerManifest, "", fieldManager); err != nil {
		return fmt.Errorf("apply cert-manager manifest: %w", err)
	}
	if err := opts.Client.WaitForDeploymentReady(ctx, "cert-manager", "cert-manager", opts.CertManagerReadyTimeout); err != nil {
		return fmt.Errorf("wait cert-manager ready: %w", err)
	}
	if err := opts.Client.WaitForDeploymentReady(ctx, "cert-manager", "cert-manager-webhook", opts.CertManagerReadyTimeout); err != nil {
		return fmt.Errorf("wait cert-manager-webhook ready: %w", err)
	}
	if err := opts.Client.WaitForDeploymentReady(ctx, "cert-manager", "cert-manager-cainjector", opts.CertManagerReadyTimeout); err != nil {
		return fmt.Errorf("wait cert-manager-cainjector ready: %w", err)
	}
	return nil
}

func fetchManifest(ctx context.Context, url string) ([]byte, error) {
	trimmedURL := strings.TrimSpace(url)
	if trimmedURL == "" {
		return nil, fmt.Errorf("url is empty")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, trimmedURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 60 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %s", response.Status)
	}
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("empty manifest payload")
	}
	return payload, nil
}

func parseBoolDefault(raw string, fallback bool) (bool, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback, nil
	}
	switch strings.ToLower(trimmed) {
	case "true", "1", "yes", "y", "on":
		return true, nil
	case "false", "0", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("unsupported boolean value %q", raw)
	}
}
