package runtimedeploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
)

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
		runToken, err = randomHex(6)
		if err != nil {
			return fmt.Errorf("generate kaniko run token: %w", err)
		}
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
