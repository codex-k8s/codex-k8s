package runtimedeploy

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	defaultServicesConfigPath = "services.yaml"
	defaultRepositoryRoot     = "."
	defaultRolloutTimeout     = 20 * time.Minute
	defaultKanikoTimeout      = 30 * time.Minute
	defaultFieldManager       = "codex-k8s-control-plane"
)

var placeholderPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
var imageTagSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

// Service prepares runtime environments from services.yaml contract.
type Service struct {
	cfg    Config
	k8s    KubernetesClient
	logger *slog.Logger
}

// NewService creates runtime deployment service.
func NewService(cfg Config, deps Dependencies) (*Service, error) {
	if deps.Kubernetes == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}
	cfg.ServicesConfigPath = strings.TrimSpace(cfg.ServicesConfigPath)
	if cfg.ServicesConfigPath == "" {
		cfg.ServicesConfigPath = defaultServicesConfigPath
	}
	cfg.RepositoryRoot = strings.TrimSpace(cfg.RepositoryRoot)
	if cfg.RepositoryRoot == "" {
		cfg.RepositoryRoot = defaultRepositoryRoot
	}
	cfg.KanikoFieldManager = strings.TrimSpace(cfg.KanikoFieldManager)
	if cfg.KanikoFieldManager == "" {
		cfg.KanikoFieldManager = defaultFieldManager
	}
	if cfg.RolloutTimeout <= 0 {
		cfg.RolloutTimeout = defaultRolloutTimeout
	}
	if cfg.KanikoTimeout <= 0 {
		cfg.KanikoTimeout = defaultKanikoTimeout
	}
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	return &Service{
		cfg:    cfg,
		k8s:    deps.Kubernetes,
		logger: deps.Logger,
	}, nil
}

// PrepareRunEnvironment builds images and applies infrastructure/services for one runtime target namespace.
func (s *Service) PrepareRunEnvironment(ctx context.Context, params PrepareParams) (PrepareResult, error) {
	zero := PrepareResult{}
	runID := strings.TrimSpace(params.RunID)
	if runID == "" {
		return zero, fmt.Errorf("run_id is required")
	}

	targetEnv := strings.TrimSpace(params.TargetEnv)
	if targetEnv == "" {
		targetEnv = "ai"
	}
	targetNamespace := strings.TrimSpace(params.Namespace)
	if targetNamespace == "" && strings.EqualFold(targetEnv, "ai-staging") {
		targetNamespace = buildAIStagingNamespace(params.RepositoryFullName)
	}

	templateVars := s.buildTemplateVars(params, targetNamespace)
	servicesConfigPath := s.resolveServicesConfigPath(params.ServicesYAMLPath)
	loaded, err := servicescfg.Load(servicesConfigPath, servicescfg.LoadOptions{
		Env:       targetEnv,
		Namespace: targetNamespace,
		Slot:      params.SlotNo,
		Vars:      templateVars,
	})
	if err != nil {
		return zero, fmt.Errorf("load services config: %w", err)
	}

	if targetNamespace == "" {
		targetNamespace = strings.TrimSpace(loaded.Context.Namespace)
	}
	if targetNamespace == "" {
		return zero, fmt.Errorf("resolved target namespace is empty")
	}
	if effectiveEnv := strings.TrimSpace(loaded.Context.Env); effectiveEnv != "" {
		targetEnv = effectiveEnv
	}
	templateVars["CODEXK8S_STAGING_NAMESPACE"] = targetNamespace
	templateVars["CODEXK8S_WORKER_K8S_NAMESPACE"] = targetNamespace
	templateVars["CODEXK8S_GITHUB_REPO"] = strings.TrimSpace(params.RepositoryFullName)
	if strings.TrimSpace(templateVars["CODEXK8S_WORKER_JOB_IMAGE"]) == "" {
		if v := strings.TrimSpace(templateVars["CODEXK8S_AGENT_RUNNER_IMAGE"]); v != "" {
			templateVars["CODEXK8S_WORKER_JOB_IMAGE"] = v
		}
	}

	if strings.EqualFold(strings.TrimSpace(loaded.Stack.Spec.Project), "codex-k8s") {
		if err := s.ensureCodexK8sPrerequisites(ctx, targetNamespace, templateVars); err != nil {
			return zero, fmt.Errorf("ensure codex-k8s prerequisites: %w", err)
		}
	}

	if err := s.buildImages(ctx, params, loaded.Stack, targetNamespace, templateVars); err != nil {
		return zero, fmt.Errorf("build images: %w", err)
	}

	appliedInfra, err := s.applyInfrastructure(ctx, loaded.Stack, targetNamespace, templateVars)
	if err != nil {
		return zero, fmt.Errorf("apply infrastructure: %w", err)
	}
	if err := s.applyServices(ctx, loaded.Stack, targetNamespace, templateVars, appliedInfra); err != nil {
		return zero, fmt.Errorf("apply services: %w", err)
	}
	return PrepareResult{
		Namespace: targetNamespace,
		TargetEnv: targetEnv,
	}, nil
}

func (s *Service) resolveServicesConfigPath(pathFromRun string) string {
	trimmed := strings.TrimSpace(pathFromRun)
	if trimmed != "" {
		if filepath.IsAbs(trimmed) {
			if _, err := os.Stat(trimmed); err == nil {
				return trimmed
			}
		} else {
			candidate := filepath.Join(s.cfg.RepositoryRoot, trimmed)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}
	if filepath.IsAbs(s.cfg.ServicesConfigPath) {
		return s.cfg.ServicesConfigPath
	}
	return filepath.Join(s.cfg.RepositoryRoot, s.cfg.ServicesConfigPath)
}

func (s *Service) buildTemplateVars(params PrepareParams, namespace string) map[string]string {
	vars := defaultTemplateVars()
	for _, item := range os.Environ() {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key == "" {
			continue
		}
		vars[key] = value
	}

	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace == "" && strings.EqualFold(strings.TrimSpace(params.TargetEnv), "ai-staging") {
		targetNamespace = buildAIStagingNamespace(params.RepositoryFullName)
	}
	if targetNamespace != "" {
		vars["CODEXK8S_STAGING_NAMESPACE"] = targetNamespace
		vars["CODEXK8S_WORKER_K8S_NAMESPACE"] = targetNamespace
		if strings.TrimSpace(vars["CODEXK8S_CONTROL_PLANE_GRPC_TARGET"]) == "" {
			vars["CODEXK8S_CONTROL_PLANE_GRPC_TARGET"] = fmt.Sprintf("codex-k8s-control-plane.%s.svc.cluster.local:9090", targetNamespace)
		}
		if strings.TrimSpace(vars["CODEXK8S_CONTROL_PLANE_MCP_BASE_URL"]) == "" {
			vars["CODEXK8S_CONTROL_PLANE_MCP_BASE_URL"] = fmt.Sprintf("http://codex-k8s-control-plane.%s.svc.cluster.local:8081/mcp", targetNamespace)
		}
	}

	buildRef := strings.TrimSpace(params.BuildRef)
	if buildRef == "" {
		buildRef = strings.TrimSpace(vars["CODEXK8S_BUILD_REF"])
	}
	if buildRef == "" {
		buildRef = strings.TrimSpace(vars["CODEXK8S_AGENT_BASE_BRANCH"])
	}
	if buildRef == "" {
		buildRef = "main"
	}
	vars["CODEXK8S_BUILD_REF"] = buildRef
	vars["CODEXK8S_BUILD_TAG"] = sanitizeImageTag(buildRef)
	if repo := strings.TrimSpace(params.RepositoryFullName); repo != "" {
		vars["CODEXK8S_GITHUB_REPO"] = repo
	}
	if strings.TrimSpace(vars["CODEXK8S_WORKER_JOB_IMAGE"]) == "" {
		vars["CODEXK8S_WORKER_JOB_IMAGE"] = strings.TrimSpace(vars["CODEXK8S_AGENT_RUNNER_IMAGE"])
	}

	return vars
}

func (s *Service) ensureCodexK8sPrerequisites(ctx context.Context, namespace string, vars map[string]string) error {
	postgresData := map[string][]byte{
		"CODEXK8S_POSTGRES_DB":       []byte(valueOr(vars, "CODEXK8S_POSTGRES_DB", "codex_k8s")),
		"CODEXK8S_POSTGRES_USER":     []byte(valueOr(vars, "CODEXK8S_POSTGRES_USER", "codex_k8s")),
		"CODEXK8S_POSTGRES_PASSWORD": []byte(valueOr(vars, "CODEXK8S_POSTGRES_PASSWORD", randomHex(24))),
	}
	if err := s.k8s.UpsertSecret(ctx, namespace, "codex-k8s-postgres", postgresData); err != nil {
		return fmt.Errorf("upsert codex-k8s-postgres secret: %w", err)
	}

	runtimeSecret := map[string][]byte{
		"CODEXK8S_GITHUB_PAT":                        []byte(valueOr(vars, "CODEXK8S_GITHUB_PAT", "")),
		"CODEXK8S_OPENAI_API_KEY":                    []byte(valueOr(vars, "CODEXK8S_OPENAI_API_KEY", "")),
		"CODEXK8S_OPENAI_AUTH_FILE":                  []byte(valueOr(vars, "CODEXK8S_OPENAI_AUTH_FILE", "")),
		"CODEXK8S_PROJECT_DB_ADMIN_HOST":             []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_ADMIN_HOST", "postgres")),
		"CODEXK8S_PROJECT_DB_ADMIN_PORT":             []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_ADMIN_PORT", "5432")),
		"CODEXK8S_PROJECT_DB_ADMIN_USER":             []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_ADMIN_USER", valueOr(vars, "CODEXK8S_POSTGRES_USER", "codex_k8s"))),
		"CODEXK8S_PROJECT_DB_ADMIN_PASSWORD":         []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_ADMIN_PASSWORD", string(postgresData["CODEXK8S_POSTGRES_PASSWORD"]))),
		"CODEXK8S_PROJECT_DB_ADMIN_SSLMODE":          []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_ADMIN_SSLMODE", "disable")),
		"CODEXK8S_PROJECT_DB_ADMIN_DATABASE":         []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_ADMIN_DATABASE", "postgres")),
		"CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS": []byte(valueOr(vars, "CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS", "dev,staging,ai-staging,prod")),
		"CODEXK8S_GIT_BOT_TOKEN":                     []byte(valueOr(vars, "CODEXK8S_GIT_BOT_TOKEN", "")),
		"CODEXK8S_GIT_BOT_USERNAME":                  []byte(valueOr(vars, "CODEXK8S_GIT_BOT_USERNAME", "codex-bot")),
		"CODEXK8S_GIT_BOT_MAIL":                      []byte(valueOr(vars, "CODEXK8S_GIT_BOT_MAIL", "codex-bot@codex-k8s.local")),
		"CODEXK8S_CONTEXT7_API_KEY":                  []byte(valueOr(vars, "CODEXK8S_CONTEXT7_API_KEY", "")),
		"CODEXK8S_APP_SECRET_KEY":                    []byte(valueOr(vars, "CODEXK8S_APP_SECRET_KEY", randomHex(32))),
		"CODEXK8S_TOKEN_ENCRYPTION_KEY":              []byte(valueOr(vars, "CODEXK8S_TOKEN_ENCRYPTION_KEY", randomHex(32))),
		"CODEXK8S_MCP_TOKEN_SIGNING_KEY":             []byte(valueOr(vars, "CODEXK8S_MCP_TOKEN_SIGNING_KEY", valueOr(vars, "CODEXK8S_TOKEN_ENCRYPTION_KEY", randomHex(32)))),
		"CODEXK8S_MCP_TOKEN_TTL":                     []byte(valueOr(vars, "CODEXK8S_MCP_TOKEN_TTL", "24h")),
		"CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS":     []byte(valueOr(vars, "CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS", "14")),
		"CODEXK8S_LEARNING_MODE_DEFAULT":             []byte(valueOr(vars, "CODEXK8S_LEARNING_MODE_DEFAULT", "true")),
		"CODEXK8S_GITHUB_WEBHOOK_SECRET":             []byte(valueOr(vars, "CODEXK8S_GITHUB_WEBHOOK_SECRET", randomHex(32))),
		"CODEXK8S_GITHUB_WEBHOOK_URL":                []byte(valueOr(vars, "CODEXK8S_GITHUB_WEBHOOK_URL", "")),
		"CODEXK8S_GITHUB_WEBHOOK_EVENTS":             []byte(valueOr(vars, "CODEXK8S_GITHUB_WEBHOOK_EVENTS", "push,pull_request,issues,issue_comment,pull_request_review,pull_request_review_comment")),
		"CODEXK8S_PUBLIC_BASE_URL":                   []byte(valueOr(vars, "CODEXK8S_PUBLIC_BASE_URL", "https://example.invalid")),
		"CODEXK8S_BOOTSTRAP_OWNER_EMAIL":             []byte(valueOr(vars, "CODEXK8S_BOOTSTRAP_OWNER_EMAIL", "owner@example.invalid")),
		"CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS":          []byte(valueOr(vars, "CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS", "")),
		"CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS":   []byte(valueOr(vars, "CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS", "")),
		"CODEXK8S_GITHUB_OAUTH_CLIENT_ID":            []byte(valueOr(vars, "CODEXK8S_GITHUB_OAUTH_CLIENT_ID", "placeholder-client-id")),
		"CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET":        []byte(valueOr(vars, "CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET", "placeholder-client-secret")),
		"CODEXK8S_JWT_SIGNING_KEY":                   []byte(valueOr(vars, "CODEXK8S_JWT_SIGNING_KEY", randomHex(32))),
		"CODEXK8S_JWT_TTL":                           []byte(valueOr(vars, "CODEXK8S_JWT_TTL", "15m")),
		"CODEXK8S_VITE_DEV_UPSTREAM":                 []byte(valueOr(vars, "CODEXK8S_VITE_DEV_UPSTREAM", "http://codex-k8s-web-console:5173")),
	}
	if err := s.k8s.UpsertSecret(ctx, namespace, "codex-k8s-runtime", runtimeSecret); err != nil {
		return fmt.Errorf("upsert codex-k8s-runtime secret: %w", err)
	}

	oauthCookie := valueOr(vars, "OAUTH2_PROXY_COOKIE_SECRET", "")
	if oauthCookie == "" {
		existing, found, err := s.k8s.GetSecretData(ctx, namespace, "codex-k8s-oauth2-proxy")
		if err != nil {
			return fmt.Errorf("load oauth2-proxy secret: %w", err)
		}
		if found {
			if value, ok := existing["OAUTH2_PROXY_COOKIE_SECRET"]; ok {
				existingCookie := string(value)
				if isValidOAuthCookieSecret(existingCookie) {
					oauthCookie = existingCookie
				}
			}
		}
		if oauthCookie == "" {
			oauthCookie = randomHex(16)
		}
	}
	oauthSecret := map[string][]byte{
		"OAUTH2_PROXY_CLIENT_ID":     []byte(valueOr(vars, "CODEXK8S_GITHUB_OAUTH_CLIENT_ID", "placeholder-client-id")),
		"OAUTH2_PROXY_CLIENT_SECRET": []byte(valueOr(vars, "CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET", "placeholder-client-secret")),
		"OAUTH2_PROXY_COOKIE_SECRET": []byte(oauthCookie),
	}
	if err := s.k8s.UpsertSecret(ctx, namespace, "codex-k8s-oauth2-proxy", oauthSecret); err != nil {
		return fmt.Errorf("upsert codex-k8s-oauth2-proxy secret: %w", err)
	}

	labels := make(map[string]string, len(labelCatalogDefaults))
	for key, fallback := range labelCatalogDefaults {
		labels[key] = valueOr(vars, key, fallback)
	}
	if err := s.k8s.UpsertConfigMap(ctx, namespace, "codex-k8s-label-catalog", labels); err != nil {
		return fmt.Errorf("upsert codex-k8s-label-catalog configmap: %w", err)
	}

	migrationsData, err := readMigrationFiles(filepath.Join(s.cfg.RepositoryRoot, "services/internal/control-plane/cmd/cli/migrations"))
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	if len(migrationsData) > 0 {
		if err := s.k8s.UpsertConfigMap(ctx, namespace, "codex-k8s-migrations", migrationsData); err != nil {
			return fmt.Errorf("upsert codex-k8s-migrations configmap: %w", err)
		}
	}

	return nil
}

func (s *Service) buildImages(ctx context.Context, params PrepareParams, stack *servicescfg.Stack, namespace string, vars map[string]string) error {
	if stack == nil {
		return fmt.Errorf("stack is nil")
	}

	type buildImageEntry struct {
		Name  string
		Image servicescfg.Image
	}
	buildEntries := make([]buildImageEntry, 0, len(stack.Spec.Images))
	for name, image := range stack.Spec.Images {
		if strings.EqualFold(strings.TrimSpace(image.Type), "build") {
			buildEntries = append(buildEntries, buildImageEntry{
				Name:  strings.TrimSpace(name),
				Image: image,
			})
		}
	}
	if len(buildEntries) == 0 {
		return nil
	}
	sort.Slice(buildEntries, func(i, j int) bool { return buildEntries[i].Name < buildEntries[j].Name })

	githubPAT := strings.TrimSpace(s.cfg.GitHubPAT)
	if githubPAT == "" {
		githubPAT = strings.TrimSpace(vars["CODEXK8S_GITHUB_PAT"])
	}
	if githubPAT == "" {
		return fmt.Errorf("CODEXK8S_GITHUB_PAT is required for kaniko build jobs")
	}
	repositoryFullName := strings.TrimSpace(params.RepositoryFullName)
	if repositoryFullName == "" {
		repositoryFullName = strings.TrimSpace(vars["CODEXK8S_GITHUB_REPO"])
	}
	if repositoryFullName == "" {
		return fmt.Errorf("repository_full_name is required for kaniko build jobs")
	}

	if err := s.k8s.UpsertSecret(ctx, namespace, "codex-k8s-git-token", map[string][]byte{
		"token": []byte(githubPAT),
	}); err != nil {
		return fmt.Errorf("upsert codex-k8s-git-token secret: %w", err)
	}

	templatePath := filepath.Join(s.cfg.RepositoryRoot, "deploy/base/kaniko/kaniko-build-job.yaml.tpl")
	templateRaw, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read kaniko template %s: %w", templatePath, err)
	}

	buildRef := strings.TrimSpace(params.BuildRef)
	if buildRef == "" {
		buildRef = strings.TrimSpace(vars["CODEXK8S_BUILD_REF"])
	}
	if buildRef == "" {
		buildRef = strings.TrimSpace(vars["CODEXK8S_AGENT_BASE_BRANCH"])
	}
	if buildRef == "" {
		buildRef = "main"
	}
	vars["CODEXK8S_BUILD_REF"] = buildRef

	runToken := sanitizeNameToken(params.RunID, 12)
	if runToken == "" {
		runToken = randomHex(6)
	}

	for _, entry := range buildEntries {
		repository := strings.TrimSpace(entry.Image.Repository)
		if repository == "" {
			return fmt.Errorf("image %q repository is required for build type", entry.Name)
		}
		tag := sanitizeImageTag(strings.TrimSpace(entry.Image.TagTemplate))
		if tag == "" {
			tag = "latest"
		}
		destinationLatest := repository + ":latest"
		destinationTagged := repository + ":" + tag

		contextArg := resolveKanikoContext(entry.Image.Context)
		dockerfileArg, dockerfileErr := resolveKanikoDockerfile(entry.Image.Dockerfile)
		if dockerfileErr != nil {
			return fmt.Errorf("image %q: %w", entry.Name, dockerfileErr)
		}
		jobName := fmt.Sprintf("codex-k8s-kaniko-%s-%s", sanitizeNameToken(entry.Name, 24), runToken)
		if len(jobName) > 63 {
			jobName = strings.TrimRight(jobName[:63], "-")
		}
		jobVars := cloneStringMap(vars)
		jobVars["CODEXK8S_STAGING_NAMESPACE"] = namespace
		jobVars["CODEXK8S_GITHUB_REPO"] = repositoryFullName
		jobVars["CODEXK8S_BUILD_REF"] = buildRef
		jobVars["CODEXK8S_KANIKO_JOB_NAME"] = jobName
		jobVars["CODEXK8S_KANIKO_COMPONENT"] = sanitizeNameToken(entry.Name, 30)
		jobVars["CODEXK8S_KANIKO_CONTEXT"] = contextArg
		jobVars["CODEXK8S_KANIKO_DOCKERFILE"] = dockerfileArg
		jobVars["CODEXK8S_KANIKO_DESTINATION_LATEST"] = destinationLatest
		jobVars["CODEXK8S_KANIKO_DESTINATION_SHA"] = destinationTagged

		renderedJob := renderPlaceholders(string(templateRaw), jobVars)
		if err := s.k8s.DeleteJobIfExists(ctx, namespace, jobName); err != nil {
			return fmt.Errorf("delete previous kaniko job %s: %w", jobName, err)
		}
		if _, err := s.k8s.ApplyManifest(ctx, []byte(renderedJob), namespace, s.cfg.KanikoFieldManager); err != nil {
			return fmt.Errorf("apply kaniko job %s: %w", jobName, err)
		}
		if err := s.k8s.WaitForJobComplete(ctx, namespace, jobName, s.cfg.KanikoTimeout); err != nil {
			return fmt.Errorf("wait kaniko job %s: %w", jobName, err)
		}

		applyBuiltImageResult(vars, entry.Name, destinationTagged)
	}

	return nil
}

func (s *Service) applyInfrastructure(ctx context.Context, stack *servicescfg.Stack, namespace string, vars map[string]string) (map[string]struct{}, error) {
	enabled := make(map[string]servicescfg.InfrastructureItem, len(stack.Spec.Infrastructure))
	for _, item := range stack.Spec.Infrastructure {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return nil, fmt.Errorf("infrastructure item name is required")
		}
		include, err := evaluateWhen(item.When)
		if err != nil {
			return nil, fmt.Errorf("infrastructure %q when expression: %w", name, err)
		}
		if !include {
			continue
		}
		enabled[name] = item
	}
	order, err := topoSortInfrastructure(enabled)
	if err != nil {
		return nil, err
	}

	applied := make(map[string]struct{}, len(enabled))
	for _, name := range order {
		item := enabled[name]
		if err := s.applyUnit(ctx, name, item.Manifests, namespace, vars); err != nil {
			return nil, err
		}
		applied[name] = struct{}{}
	}
	return applied, nil
}

func (s *Service) applyServices(
	ctx context.Context,
	stack *servicescfg.Stack,
	namespace string,
	vars map[string]string,
	applied map[string]struct{},
) error {
	enabledByName := make(map[string]servicescfg.Service, len(stack.Spec.Services))
	groupToNames := make(map[string][]string)
	for _, service := range stack.Spec.Services {
		name := strings.TrimSpace(service.Name)
		if name == "" {
			return fmt.Errorf("service name is required")
		}
		include, err := evaluateWhen(service.When)
		if err != nil {
			return fmt.Errorf("service %q when expression: %w", name, err)
		}
		if !include {
			continue
		}
		enabledByName[name] = service
		group := strings.TrimSpace(service.DeployGroup)
		groupToNames[group] = append(groupToNames[group], name)
	}

	if len(enabledByName) == 0 {
		return nil
	}

	groupOrder := buildServiceGroupOrder(stack.Spec.Orchestration.DeployOrder, groupToNames)
	for _, group := range groupOrder {
		names := append([]string(nil), groupToNames[group]...)
		sort.Strings(names)
		for len(names) > 0 {
			progress := false
			for idx := 0; idx < len(names); idx++ {
				name := names[idx]
				service := enabledByName[name]
				if !dependenciesSatisfied(service.DependsOn, applied) {
					continue
				}
				if err := s.applyUnit(ctx, name, service.Manifests, namespace, vars); err != nil {
					return err
				}
				applied[name] = struct{}{}
				names = append(names[:idx], names[idx+1:]...)
				progress = true
				break
			}
			if !progress {
				return fmt.Errorf("service dependency deadlock in group %q: unresolved %s", group, strings.Join(names, ", "))
			}
		}
	}

	return nil
}

func (s *Service) applyUnit(ctx context.Context, unitName string, manifests []servicescfg.ManifestRef, namespace string, vars map[string]string) error {
	for _, manifest := range manifests {
		path := strings.TrimSpace(manifest.Path)
		if path == "" {
			continue
		}
		fullPath := path
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(s.cfg.RepositoryRoot, path)
		}
		raw, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("read manifest %s for %s: %w", fullPath, unitName, err)
		}
		rendered := renderPlaceholders(string(raw), vars)

		refs, err := parseManifestRefs([]byte(rendered), namespace)
		if err != nil {
			return fmt.Errorf("parse manifest refs %s for %s: %w", fullPath, unitName, err)
		}
		for _, ref := range refs {
			if strings.EqualFold(ref.Kind, "Job") && strings.TrimSpace(ref.Name) != "" {
				jobNamespace := strings.TrimSpace(ref.Namespace)
				if jobNamespace == "" {
					jobNamespace = namespace
				}
				if jobNamespace != "" {
					if err := s.k8s.DeleteJobIfExists(ctx, jobNamespace, ref.Name); err != nil {
						return fmt.Errorf("delete previous job %s/%s before apply: %w", jobNamespace, ref.Name, err)
					}
				}
			}
		}

		appliedRefs, err := s.k8s.ApplyManifest(ctx, []byte(rendered), namespace, s.cfg.KanikoFieldManager)
		if err != nil {
			return fmt.Errorf("apply manifest %s for %s: %w", fullPath, unitName, err)
		}
		for _, ref := range appliedRefs {
			if err := s.waitAppliedResource(ctx, ref, namespace); err != nil {
				return fmt.Errorf("wait applied resource %s/%s for %s: %w", ref.Kind, ref.Name, unitName, err)
			}
		}
	}
	return nil
}

func (s *Service) waitAppliedResource(ctx context.Context, ref AppliedResourceRef, fallbackNamespace string) error {
	namespace := strings.TrimSpace(ref.Namespace)
	if namespace == "" {
		namespace = strings.TrimSpace(fallbackNamespace)
	}

	switch strings.ToLower(strings.TrimSpace(ref.Kind)) {
	case "deployment":
		if namespace == "" {
			return nil
		}
		return s.k8s.WaitForDeploymentReady(ctx, namespace, ref.Name, s.cfg.RolloutTimeout)
	case "statefulset":
		if namespace == "" {
			return nil
		}
		return s.k8s.WaitForStatefulSetReady(ctx, namespace, ref.Name, s.cfg.RolloutTimeout)
	case "daemonset":
		if namespace == "" {
			return nil
		}
		return s.k8s.WaitForDaemonSetReady(ctx, namespace, ref.Name, s.cfg.RolloutTimeout)
	case "job":
		if namespace == "" {
			return nil
		}
		return s.k8s.WaitForJobComplete(ctx, namespace, ref.Name, s.cfg.RolloutTimeout)
	default:
		return nil
	}
}

func parseManifestRefs(manifest []byte, namespaceOverride string) ([]AppliedResourceRef, error) {
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 4096)
	overrideNamespace := strings.TrimSpace(namespaceOverride)
	out := make([]AppliedResourceRef, 0, 8)
	for {
		var objectMap map[string]any
		if err := decoder.Decode(&objectMap); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(objectMap) == 0 {
			continue
		}
		obj := &unstructured.Unstructured{Object: objectMap}
		name := strings.TrimSpace(obj.GetName())
		if name == "" {
			continue
		}
		namespace := strings.TrimSpace(obj.GetNamespace())
		if namespace == "" {
			namespace = overrideNamespace
		}
		out = append(out, AppliedResourceRef{
			APIVersion: obj.GetAPIVersion(),
			Kind:       obj.GetKind(),
			Namespace:  namespace,
			Name:       name,
		})
	}
	return out, nil
}

func renderPlaceholders(input string, vars map[string]string) string {
	return placeholderPattern.ReplaceAllStringFunc(input, func(token string) string {
		matches := placeholderPattern.FindStringSubmatch(token)
		if len(matches) != 2 {
			return token
		}
		key := matches[1]
		if value, ok := vars[key]; ok {
			return value
		}
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
		return ""
	})
}

func evaluateWhen(value string) (bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return true, nil
	}
	parsed, err := strconv.ParseBool(strings.ToLower(trimmed))
	if err != nil {
		return false, err
	}
	return parsed, nil
}

func dependenciesSatisfied(dependsOn []string, applied map[string]struct{}) bool {
	for _, dependency := range dependsOn {
		name := strings.TrimSpace(dependency)
		if name == "" {
			continue
		}
		if _, ok := applied[name]; !ok {
			return false
		}
	}
	return true
}

func topoSortInfrastructure(items map[string]servicescfg.InfrastructureItem) ([]string, error) {
	perm := make(map[string]struct{}, len(items))
	temp := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))

	var visit func(name string) error
	visit = func(name string) error {
		if _, ok := perm[name]; ok {
			return nil
		}
		if _, ok := temp[name]; ok {
			return fmt.Errorf("infrastructure dependency cycle detected at %q", name)
		}
		temp[name] = struct{}{}
		item := items[name]
		for _, dependency := range item.DependsOn {
			depName := strings.TrimSpace(dependency)
			if depName == "" {
				continue
			}
			if _, exists := items[depName]; !exists {
				continue
			}
			if err := visit(depName); err != nil {
				return err
			}
		}
		delete(temp, name)
		perm[name] = struct{}{}
		out = append(out, name)
		return nil
	}

	names := make([]string, 0, len(items))
	for name := range items {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := visit(name); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func buildServiceGroupOrder(deployOrder []string, groupToNames map[string][]string) []string {
	seen := make(map[string]struct{}, len(groupToNames))
	out := make([]string, 0, len(groupToNames))

	for _, group := range deployOrder {
		trimmed := strings.TrimSpace(group)
		if _, ok := groupToNames[trimmed]; !ok {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}

	rest := make([]string, 0, len(groupToNames))
	for group := range groupToNames {
		if _, ok := seen[group]; ok {
			continue
		}
		rest = append(rest, group)
	}
	sort.Strings(rest)
	out = append(out, rest...)
	return out
}

func resolveKanikoContext(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || trimmed == "." {
		return "dir:///workspace"
	}
	if strings.HasPrefix(trimmed, "dir://") {
		return trimmed
	}
	normalized := strings.TrimPrefix(trimmed, "./")
	return "dir:///workspace/" + normalized
}

func resolveKanikoDockerfile(path string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", fmt.Errorf("dockerfile is required for build image")
	}
	if strings.HasPrefix(trimmed, "/") {
		return trimmed, nil
	}
	normalized := strings.TrimPrefix(trimmed, "./")
	return "/workspace/" + normalized, nil
}

func applyBuiltImageResult(vars map[string]string, imageName string, imageRef string) {
	switch strings.ToLower(strings.TrimSpace(imageName)) {
	case "api-gateway":
		vars["CODEXK8S_API_GATEWAY_IMAGE"] = imageRef
	case "control-plane":
		vars["CODEXK8S_CONTROL_PLANE_IMAGE"] = imageRef
	case "worker":
		vars["CODEXK8S_WORKER_IMAGE"] = imageRef
	case "agent-runner":
		vars["CODEXK8S_AGENT_RUNNER_IMAGE"] = imageRef
		if strings.TrimSpace(vars["CODEXK8S_WORKER_JOB_IMAGE"]) == "" {
			vars["CODEXK8S_WORKER_JOB_IMAGE"] = imageRef
		}
	case "web-console":
		vars["CODEXK8S_WEB_CONSOLE_IMAGE"] = imageRef
	}
}

func buildAIStagingNamespace(repositoryFullName string) string {
	repoName := repositoryName(repositoryFullName)
	if repoName == "" {
		return ""
	}
	return repoName + "-ai-staging"
}

func repositoryName(repositoryFullName string) string {
	owner, repo, ok := strings.Cut(strings.TrimSpace(repositoryFullName), "/")
	if !ok {
		return sanitizeNameToken(owner, 40)
	}
	_ = owner
	return sanitizeNameToken(repo, 40)
}

func sanitizeNameToken(value string, max int) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return ""
	}
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, ".", "-")
	normalized = imageTagSanitizer.ReplaceAllString(normalized, "-")
	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}
	normalized = strings.Trim(normalized, "-")
	if max > 0 && len(normalized) > max {
		normalized = strings.TrimRight(normalized[:max], "-")
	}
	return normalized
}

func sanitizeImageTag(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return ""
	}
	normalized = imageTagSanitizer.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, ".-")
	if normalized == "" {
		return ""
	}
	if len(normalized) > 120 {
		normalized = normalized[:120]
	}
	return normalized
}

func valueOr(values map[string]string, key string, fallback string) string {
	if values != nil {
		if value, ok := values[key]; ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func cloneStringMap(input map[string]string) map[string]string {
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func randomHex(numBytes int) string {
	if numBytes <= 0 {
		numBytes = 16
	}
	raw := make([]byte, numBytes)
	if _, err := rand.Read(raw); err != nil {
		return "fallback-random-value"
	}
	return hex.EncodeToString(raw)
}

func isValidOAuthCookieSecret(value string) bool {
	size := len(strings.TrimSpace(value))
	return size == 16 || size == 24 || size == 32
}

func readMigrationFiles(dir string) (map[string]string, error) {
	items, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	files := make([]string, 0, len(items))
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		name := strings.TrimSpace(item.Name())
		if !strings.HasSuffix(strings.ToLower(name), ".sql") {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)

	data := make(map[string]string, len(files))
	for _, name := range files {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		data[name] = string(raw)
	}
	return data, nil
}

func defaultTemplateVars() map[string]string {
	return map[string]string{
		"CODEXK8S_STAGING_NAMESPACE":                       "codex-k8s-ai-staging",
		"CODEXK8S_INTERNAL_REGISTRY_HOST":                  "127.0.0.1:5000",
		"CODEXK8S_KANIKO_CACHE_ENABLED":                    "true",
		"CODEXK8S_KANIKO_CACHE_TTL":                        "168h",
		"CODEXK8S_KANIKO_CACHE_COMPRESSED":                 "false",
		"CODEXK8S_KANIKO_RESOURCES_REQUEST_CPU":            "8",
		"CODEXK8S_KANIKO_RESOURCES_REQUEST_MEMORY":         "16Gi",
		"CODEXK8S_KANIKO_RESOURCES_LIMIT_CPU":              "16",
		"CODEXK8S_KANIKO_RESOURCES_LIMIT_MEMORY":           "32Gi",
		"CODEXK8S_API_GATEWAY_IMAGE":                       "127.0.0.1:5000/codex-k8s/api-gateway:latest",
		"CODEXK8S_CONTROL_PLANE_IMAGE":                     "127.0.0.1:5000/codex-k8s/control-plane:latest",
		"CODEXK8S_WORKER_IMAGE":                            "127.0.0.1:5000/codex-k8s/worker:latest",
		"CODEXK8S_AGENT_RUNNER_IMAGE":                      "127.0.0.1:5000/codex-k8s/agent-runner:latest",
		"CODEXK8S_WEB_CONSOLE_IMAGE":                       "127.0.0.1:5000/codex-k8s/web-console:latest",
		"CODEXK8S_WORKER_JOB_IMAGE":                        "127.0.0.1:5000/codex-k8s/agent-runner:latest",
		"CODEXK8S_WORKER_REPLICAS":                         "1",
		"CODEXK8S_WORKER_POLL_INTERVAL":                    "5s",
		"CODEXK8S_WORKER_CLAIM_LIMIT":                      "2",
		"CODEXK8S_WORKER_RUNNING_CHECK_LIMIT":              "200",
		"CODEXK8S_WORKER_SLOTS_PER_PROJECT":                "2",
		"CODEXK8S_WORKER_SLOT_LEASE_TTL":                   "10m",
		"CODEXK8S_WORKER_K8S_NAMESPACE":                    "codex-k8s-ai-staging",
		"CODEXK8S_WORKER_JOB_COMMAND":                      "/usr/local/bin/codex-k8s-agent-runner",
		"CODEXK8S_WORKER_JOB_TTL_SECONDS":                  "600",
		"CODEXK8S_WORKER_JOB_BACKOFF_LIMIT":                "0",
		"CODEXK8S_WORKER_JOB_ACTIVE_DEADLINE_SECONDS":      "900",
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
		"CODEXK8S_AGENT_DEFAULT_MODEL":                     "gpt-5.3-codex",
		"CODEXK8S_AGENT_DEFAULT_REASONING_EFFORT":          "high",
		"CODEXK8S_AGENT_DEFAULT_LOCALE":                    "ru",
		"CODEXK8S_AGENT_BASE_BRANCH":                       "main",
		"CODEXK8S_STAGING_DOMAIN":                          "staging.example.com",
		"CODEXK8S_API_GATEWAY_RESOURCES_REQUEST_CPU":       "100m",
		"CODEXK8S_API_GATEWAY_RESOURCES_REQUEST_MEMORY":    "256Mi",
		"CODEXK8S_API_GATEWAY_RESOURCES_LIMIT_CPU":         "1000m",
		"CODEXK8S_API_GATEWAY_RESOURCES_LIMIT_MEMORY":      "1Gi",
		"CODEXK8S_CONTROL_PLANE_RESOURCES_REQUEST_CPU":     "100m",
		"CODEXK8S_CONTROL_PLANE_RESOURCES_REQUEST_MEMORY":  "256Mi",
		"CODEXK8S_CONTROL_PLANE_RESOURCES_LIMIT_CPU":       "1000m",
		"CODEXK8S_CONTROL_PLANE_RESOURCES_LIMIT_MEMORY":    "1Gi",
		"CODEXK8S_WORKER_RESOURCES_REQUEST_CPU":            "100m",
		"CODEXK8S_WORKER_RESOURCES_REQUEST_MEMORY":         "256Mi",
		"CODEXK8S_WORKER_RESOURCES_LIMIT_CPU":              "1000m",
		"CODEXK8S_WORKER_RESOURCES_LIMIT_MEMORY":           "1Gi",
		"CODEXK8S_WEB_CONSOLE_RESOURCES_REQUEST_CPU":       "100m",
		"CODEXK8S_WEB_CONSOLE_RESOURCES_REQUEST_MEMORY":    "128Mi",
		"CODEXK8S_WEB_CONSOLE_RESOURCES_LIMIT_CPU":         "500m",
		"CODEXK8S_WEB_CONSOLE_RESOURCES_LIMIT_MEMORY":      "512Mi",
	}
}

var labelCatalogDefaults = map[string]string{
	"CODEXK8S_RUN_INTAKE_LABEL":                   "run:intake",
	"CODEXK8S_RUN_INTAKE_REVISE_LABEL":            "run:intake:revise",
	"CODEXK8S_RUN_VISION_LABEL":                   "run:vision",
	"CODEXK8S_RUN_VISION_REVISE_LABEL":            "run:vision:revise",
	"CODEXK8S_RUN_PRD_LABEL":                      "run:prd",
	"CODEXK8S_RUN_PRD_REVISE_LABEL":               "run:prd:revise",
	"CODEXK8S_RUN_ARCH_LABEL":                     "run:arch",
	"CODEXK8S_RUN_ARCH_REVISE_LABEL":              "run:arch:revise",
	"CODEXK8S_RUN_DESIGN_LABEL":                   "run:design",
	"CODEXK8S_RUN_DESIGN_REVISE_LABEL":            "run:design:revise",
	"CODEXK8S_RUN_PLAN_LABEL":                     "run:plan",
	"CODEXK8S_RUN_PLAN_REVISE_LABEL":              "run:plan:revise",
	"CODEXK8S_RUN_DEV_LABEL":                      "run:dev",
	"CODEXK8S_RUN_DEV_REVISE_LABEL":               "run:dev:revise",
	"CODEXK8S_RUN_DEBUG_LABEL":                    "run:debug",
	"CODEXK8S_RUN_DOC_AUDIT_LABEL":                "run:doc-audit",
	"CODEXK8S_RUN_QA_LABEL":                       "run:qa",
	"CODEXK8S_RUN_RELEASE_LABEL":                  "run:release",
	"CODEXK8S_RUN_POSTDEPLOY_LABEL":               "run:postdeploy",
	"CODEXK8S_RUN_OPS_LABEL":                      "run:ops",
	"CODEXK8S_RUN_SELF_IMPROVE_LABEL":             "run:self-improve",
	"CODEXK8S_RUN_RETHINK_LABEL":                  "run:rethink",
	"CODEXK8S_MODE_DISCUSSION_LABEL":              "mode:discussion",
	"CODEXK8S_STATE_BLOCKED_LABEL":                "state:blocked",
	"CODEXK8S_STATE_IN_REVIEW_LABEL":              "state:in-review",
	"CODEXK8S_STATE_APPROVED_LABEL":               "state:approved",
	"CODEXK8S_STATE_SUPERSEDED_LABEL":             "state:superseded",
	"CODEXK8S_STATE_ABANDONED_LABEL":              "state:abandoned",
	"CODEXK8S_NEED_INPUT_LABEL":                   "need:input",
	"CODEXK8S_NEED_PM_LABEL":                      "need:pm",
	"CODEXK8S_NEED_SA_LABEL":                      "need:sa",
	"CODEXK8S_NEED_QA_LABEL":                      "need:qa",
	"CODEXK8S_NEED_SRE_LABEL":                     "need:sre",
	"CODEXK8S_NEED_EM_LABEL":                      "need:em",
	"CODEXK8S_NEED_KM_LABEL":                      "need:km",
	"CODEXK8S_NEED_REVIEWER_LABEL":                "need:reviewer",
	"CODEXK8S_AI_MODEL_GPT_5_3_CODEX_LABEL":       "[ai-model-gpt-5.3-codex]",
	"CODEXK8S_AI_MODEL_GPT_5_3_CODEX_SPARK_LABEL": "[ai-model-gpt-5.3-codex-spark]",
	"CODEXK8S_AI_MODEL_GPT_5_2_CODEX_LABEL":       "[ai-model-gpt-5.2-codex]",
	"CODEXK8S_AI_MODEL_GPT_5_1_CODEX_MAX_LABEL":   "[ai-model-gpt-5.1-codex-max]",
	"CODEXK8S_AI_MODEL_GPT_5_2_LABEL":             "[ai-model-gpt-5.2]",
	"CODEXK8S_AI_MODEL_GPT_5_1_CODEX_MINI_LABEL":  "[ai-model-gpt-5.1-codex-mini]",
	"CODEXK8S_AI_REASONING_LOW_LABEL":             "[ai-reasoning-low]",
	"CODEXK8S_AI_REASONING_MEDIUM_LABEL":          "[ai-reasoning-medium]",
	"CODEXK8S_AI_REASONING_HIGH_LABEL":            "[ai-reasoning-high]",
	"CODEXK8S_AI_REASONING_EXTRA_HIGH_LABEL":      "[ai-reasoning-extra-high]",
}
