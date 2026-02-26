package runtimedeploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (s *Service) resolveServicesConfigPath(repositoryRoot string, pathFromRun string) string {
	repoRoot := strings.TrimSpace(repositoryRoot)
	if repoRoot == "" {
		repoRoot = s.cfg.RepositoryRoot
	}
	trimmed := strings.TrimSpace(pathFromRun)
	if trimmed != "" {
		if filepath.IsAbs(trimmed) {
			if servicesConfigExists(trimmed) {
				return trimmed
			}
		} else {
			candidate := filepath.Join(repoRoot, trimmed)
			if servicesConfigExists(candidate) {
				return candidate
			}
		}
	}
	configPath := strings.TrimSpace(s.cfg.ServicesConfigPath)
	if configPath == "" {
		configPath = defaultServicesConfigPath
	}
	if filepath.IsAbs(configPath) {
		// In runtime deploy flows we should prefer repository snapshot config,
		// otherwise an absolute in-image path (for example /app/services.yaml)
		// can pin deploy logic to stale versions.
		if repoRoot != "" {
			repoByBase := filepath.Join(repoRoot, filepath.Base(configPath))
			if servicesConfigExists(repoByBase) {
				return repoByBase
			}
			repoDefault := filepath.Join(repoRoot, defaultServicesConfigPath)
			if servicesConfigExists(repoDefault) {
				return repoDefault
			}
		}
		return configPath
	}
	return filepath.Join(repoRoot, configPath)
}

func servicesConfigExists(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
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
	// Preserve platform namespace for cross-namespace secret reads (ai env uses shared secrets from platform ns).
	if strings.TrimSpace(vars["CODEXK8S_PLATFORM_NAMESPACE"]) == "" {
		vars["CODEXK8S_PLATFORM_NAMESPACE"] = strings.TrimSpace(vars["CODEXK8S_PRODUCTION_NAMESPACE"])
	}

	targetEnv := strings.TrimSpace(params.TargetEnv)
	if targetEnv == "" {
		targetEnv = "ai"
	}
	// Manifests and runtime prerequisites rely on CODEXK8S_ENV / CODEXK8S_SERVICES_CONFIG_ENV.
	vars["CODEXK8S_ENV"] = targetEnv
	vars["CODEXK8S_SERVICES_CONFIG_ENV"] = targetEnv
	vars["CODEXK8S_HOT_RELOAD"] = resolveHotReloadFlag(targetEnv, vars["CODEXK8S_HOT_RELOAD"])
	// AI hot-reload requires Go sources to stay in the image.
	// Force cleanup=false for AI builds even if production env exported true.
	if strings.EqualFold(strings.TrimSpace(targetEnv), "ai") {
		vars["CODEXK8S_KANIKO_CLEANUP"] = "false"
	}

	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace != "" {
		vars["CODEXK8S_PRODUCTION_NAMESPACE"] = targetNamespace
		vars["CODEXK8S_WORKER_K8S_NAMESPACE"] = targetNamespace
		if strings.TrimSpace(vars["CODEXK8S_CONTROL_PLANE_GRPC_TARGET"]) == "" {
			vars["CODEXK8S_CONTROL_PLANE_GRPC_TARGET"] = fmt.Sprintf("codex-k8s-control-plane.%s.svc.cluster.local:9090", targetNamespace)
		}
		if strings.TrimSpace(vars["CODEXK8S_CONTROL_PLANE_MCP_BASE_URL"]) == "" {
			vars["CODEXK8S_CONTROL_PLANE_MCP_BASE_URL"] = fmt.Sprintf("http://codex-k8s-control-plane.%s.svc.cluster.local:8081/mcp", targetNamespace)
		}
	}

	publicDomain := resolvePublicDomain(targetEnv, targetNamespace, vars)
	if publicDomain != "" {
		vars["CODEXK8S_PUBLIC_DOMAIN"] = publicDomain
		if strings.EqualFold(targetEnv, "ai") || strings.TrimSpace(vars["CODEXK8S_PUBLIC_BASE_URL"]) == "" {
			vars["CODEXK8S_PUBLIC_BASE_URL"] = "https://" + publicDomain
		}
	}
	if strings.TrimSpace(vars["CODEXK8S_SHARED_OAUTH2_PROXY_AUTH_URL"]) == "" {
		if productionDomain := strings.TrimSpace(valueOr(vars, "CODEXK8S_PRODUCTION_DOMAIN", "")); productionDomain != "" {
			vars["CODEXK8S_SHARED_OAUTH2_PROXY_AUTH_URL"] = "https://" + productionDomain + "/oauth2/auth"
		}
	}
	if strings.TrimSpace(vars["CODEXK8S_SHARED_OAUTH2_PROXY_SIGNIN_URL"]) == "" {
		if productionDomain := strings.TrimSpace(valueOr(vars, "CODEXK8S_PRODUCTION_DOMAIN", "")); productionDomain != "" {
			vars["CODEXK8S_SHARED_OAUTH2_PROXY_SIGNIN_URL"] = "https://" + productionDomain + "/oauth2/start?rd=$scheme://$host$request_uri"
		}
	}
	if strings.TrimSpace(vars["CODEXK8S_OAUTH2_PROXY_COOKIE_DOMAIN"]) == "" {
		if productionDomain := strings.TrimSpace(valueOr(vars, "CODEXK8S_PRODUCTION_DOMAIN", "")); productionDomain != "" {
			vars["CODEXK8S_OAUTH2_PROXY_COOKIE_DOMAIN"] = "." + strings.TrimPrefix(productionDomain, ".")
		}
	}
	if strings.TrimSpace(vars["CODEXK8S_TLS_SECRET_NAME"]) == "" {
		vars["CODEXK8S_TLS_SECRET_NAME"] = defaultTLSSecretName(targetEnv)
	}

	buildRef := resolveRuntimeBuildRef(
		params.BuildRef,
		vars["CODEXK8S_BUILD_REF"],
		vars["CODEXK8S_AGENT_BASE_BRANCH"],
	)
	vars["CODEXK8S_BUILD_REF"] = buildRef
	vars["CODEXK8S_BUILD_TAG"] = sanitizeImageTag(buildRef)
	if repo := strings.TrimSpace(params.RepositoryFullName); repo != "" {
		vars["CODEXK8S_GITHUB_REPO"] = repo
	}
	if strings.TrimSpace(vars["CODEXK8S_PLATFORM_DEPLOYMENT_REPLICAS"]) == "" {
		vars["CODEXK8S_PLATFORM_DEPLOYMENT_REPLICAS"] = defaultPlatformDeploymentReplicas(params.TargetEnv)
	}
	if strings.TrimSpace(vars["CODEXK8S_WORKER_REPLICAS"]) == "" {
		vars["CODEXK8S_WORKER_REPLICAS"] = defaultWorkerReplicas(params.TargetEnv, vars["CODEXK8S_PLATFORM_DEPLOYMENT_REPLICAS"])
	}

	return vars
}

func defaultHotReloadFlag(targetEnv string) string {
	return "false"
}

func resolveHotReloadFlag(targetEnv string, currentValue string) string {
	if strings.EqualFold(strings.TrimSpace(targetEnv), "ai") {
		return "true"
	}
	trimmed := strings.TrimSpace(currentValue)
	if trimmed != "" {
		return trimmed
	}
	return defaultHotReloadFlag(targetEnv)
}

func defaultPlatformDeploymentReplicas(targetEnv string) string {
	switch strings.ToLower(strings.TrimSpace(targetEnv)) {
	case "production", "prod":
		return "2"
	default:
		return "1"
	}
}

func defaultWorkerReplicas(targetEnv string, platformReplicas string) string {
	normalized := strings.TrimSpace(platformReplicas)
	switch strings.ToLower(strings.TrimSpace(targetEnv)) {
	case "production", "prod":
		if replicas, err := strconv.Atoi(normalized); err == nil && replicas >= 3 {
			return strconv.Itoa(replicas)
		}
		return "3"
	default:
		if normalized == "" {
			return "1"
		}
		return normalized
	}
}

func resolvePublicDomain(targetEnv string, namespace string, vars map[string]string) string {
	if strings.EqualFold(strings.TrimSpace(targetEnv), "ai") {
		base := strings.TrimSpace(valueOr(vars, "CODEXK8S_AI_DOMAIN", ""))
		ns := strings.TrimSpace(namespace)
		if base != "" && ns != "" {
			return ns + "." + base
		}
	}
	return strings.TrimSpace(valueOr(vars, "CODEXK8S_PRODUCTION_DOMAIN", ""))
}

func defaultTLSSecretName(targetEnv string) string {
	if strings.EqualFold(strings.TrimSpace(targetEnv), "ai") {
		return "codex-k8s-ai-tls"
	}
	return "codex-k8s-production-tls"
}
