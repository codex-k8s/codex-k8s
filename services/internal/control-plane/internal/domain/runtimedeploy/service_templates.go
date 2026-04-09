package runtimedeploy

import (
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
	if strings.TrimSpace(vars["KODEX_PLATFORM_NAMESPACE"]) == "" {
		vars["KODEX_PLATFORM_NAMESPACE"] = strings.TrimSpace(vars["KODEX_PRODUCTION_NAMESPACE"])
	}

	targetEnv := strings.TrimSpace(params.TargetEnv)
	if targetEnv == "" {
		targetEnv = "ai"
	}
	// Manifests and runtime prerequisites rely on KODEX_ENV / KODEX_SERVICES_CONFIG_ENV.
	vars["KODEX_ENV"] = targetEnv
	vars["KODEX_SERVICES_CONFIG_ENV"] = targetEnv
	vars["KODEX_HOT_RELOAD"] = resolveHotReloadFlag(targetEnv, vars["KODEX_HOT_RELOAD"])
	// AI hot-reload requires Go sources to stay in the image.
	// Force cleanup=false for AI builds even if production env exported true.
	if strings.EqualFold(strings.TrimSpace(targetEnv), "ai") {
		vars["KODEX_KANIKO_CLEANUP"] = "false"
	}

	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace != "" {
		vars["KODEX_PRODUCTION_NAMESPACE"] = targetNamespace
		vars["KODEX_WORKER_K8S_NAMESPACE"] = targetNamespace
	}

	publicDomain := resolvePublicDomain(targetEnv, targetNamespace, vars)
	if publicDomain != "" {
		vars["KODEX_PUBLIC_DOMAIN"] = publicDomain
		if strings.EqualFold(targetEnv, "ai") || strings.TrimSpace(vars["KODEX_PUBLIC_BASE_URL"]) == "" {
			vars["KODEX_PUBLIC_BASE_URL"] = "https://" + publicDomain
		}
	}
	if strings.TrimSpace(vars["KODEX_SHARED_OAUTH2_PROXY_AUTH_URL"]) == "" {
		if productionDomain := strings.TrimSpace(valueOr(vars, "KODEX_PRODUCTION_DOMAIN", "")); productionDomain != "" {
			vars["KODEX_SHARED_OAUTH2_PROXY_AUTH_URL"] = "https://" + productionDomain + "/oauth2/auth"
		}
	}
	if strings.TrimSpace(vars["KODEX_SHARED_OAUTH2_PROXY_SIGNIN_URL"]) == "" {
		if productionDomain := strings.TrimSpace(valueOr(vars, "KODEX_PRODUCTION_DOMAIN", "")); productionDomain != "" {
			vars["KODEX_SHARED_OAUTH2_PROXY_SIGNIN_URL"] = "https://" + productionDomain + "/oauth2/start?rd=$scheme://$host$request_uri"
		}
	}
	if strings.TrimSpace(vars["KODEX_OAUTH2_PROXY_COOKIE_DOMAIN"]) == "" {
		if productionDomain := strings.TrimSpace(valueOr(vars, "KODEX_PRODUCTION_DOMAIN", "")); productionDomain != "" {
			vars["KODEX_OAUTH2_PROXY_COOKIE_DOMAIN"] = "." + strings.TrimPrefix(productionDomain, ".")
		}
	}
	if strings.TrimSpace(vars["KODEX_TLS_SECRET_NAME"]) == "" {
		vars["KODEX_TLS_SECRET_NAME"] = defaultTLSSecretName(targetEnv)
	}

	buildRef := resolveRuntimeBuildRef(
		params.BuildRef,
		vars["KODEX_BUILD_REF"],
		vars["KODEX_AGENT_BASE_BRANCH"],
	)
	vars["KODEX_BUILD_REF"] = buildRef
	vars["KODEX_BUILD_TAG"] = sanitizeImageTag(buildRef)
	if repo := strings.TrimSpace(params.RepositoryFullName); repo != "" {
		vars["KODEX_GITHUB_REPO"] = repo
	}
	if strings.TrimSpace(vars["KODEX_PLATFORM_DEPLOYMENT_REPLICAS"]) == "" {
		vars["KODEX_PLATFORM_DEPLOYMENT_REPLICAS"] = defaultPlatformDeploymentReplicas(params.TargetEnv)
	}
	if strings.TrimSpace(vars["KODEX_WORKER_REPLICAS"]) == "" {
		vars["KODEX_WORKER_REPLICAS"] = defaultWorkerReplicas(params.TargetEnv, vars["KODEX_PLATFORM_DEPLOYMENT_REPLICAS"])
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
		base := strings.TrimSpace(valueOr(vars, "KODEX_AI_DOMAIN", ""))
		ns := strings.TrimSpace(namespace)
		if base != "" && ns != "" {
			return ns + "." + base
		}
	}
	return strings.TrimSpace(valueOr(vars, "KODEX_PRODUCTION_DOMAIN", ""))
}

func defaultTLSSecretName(targetEnv string) string {
	if strings.EqualFold(strings.TrimSpace(targetEnv), "ai") {
		return "kodex-ai-tls"
	}
	return "kodex-production-tls"
}
