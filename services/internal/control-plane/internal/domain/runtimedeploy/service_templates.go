package runtimedeploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
	if strings.TrimSpace(vars["CODEXK8S_PLATFORM_DEPLOYMENT_REPLICAS"]) == "" {
		vars["CODEXK8S_PLATFORM_DEPLOYMENT_REPLICAS"] = defaultPlatformDeploymentReplicas(params.TargetEnv)
	}
	if strings.TrimSpace(vars["CODEXK8S_WORKER_REPLICAS"]) == "" {
		vars["CODEXK8S_WORKER_REPLICAS"] = vars["CODEXK8S_PLATFORM_DEPLOYMENT_REPLICAS"]
	}

	return vars
}

func defaultPlatformDeploymentReplicas(targetEnv string) string {
	switch strings.ToLower(strings.TrimSpace(targetEnv)) {
	case "ai-staging", "production", "prod":
		return "2"
	default:
		return "1"
	}
}
