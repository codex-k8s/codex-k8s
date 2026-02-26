package runtimedeploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/manifesttpl"
	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
)

type buildImageEntry struct {
	Name  string
	Image servicescfg.Image
}

type buildImageResult struct {
	Name       string
	ImageRef   string
	Repository string
}

func (s *Service) buildImages(ctx context.Context, repositoryRoot string, params PrepareParams, stack *servicescfg.Stack, namespace string, vars map[string]string) error {
	runID := strings.TrimSpace(params.RunID)
	if stack == nil {
		return fmt.Errorf("stack is nil")
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
		s.appendTaskLogBestEffort(ctx, runID, "build", "info", "No build images configured, skipping kaniko stage")
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
	buildRef := resolveRuntimeBuildRef(
		params.BuildRef,
		vars["CODEXK8S_BUILD_REF"],
		vars["CODEXK8S_AGENT_BASE_BRANCH"],
	)
	vars["CODEXK8S_BUILD_REF"] = buildRef

	runToken := sanitizeNameToken(params.RunID, 12)
	if runToken == "" {
		generatedToken, tokenErr := randomHex(6)
		if tokenErr != nil {
			return fmt.Errorf("generate kaniko run token: %w", tokenErr)
		}
		runToken = generatedToken
	}

	if err := s.k8s.UpsertSecret(ctx, namespace, "codex-k8s-git-token", map[string][]byte{
		"token": []byte(githubPAT),
	}); err != nil {
		return fmt.Errorf("upsert codex-k8s-git-token secret: %w", err)
	}

	repoRoot := strings.TrimSpace(repositoryRoot)
	if repoRoot == "" {
		repoRoot = s.cfg.RepositoryRoot
	}

	kanikoTemplatePath := filepath.Join(repoRoot, "deploy/base/kaniko/kaniko-build-job.yaml.tpl")
	kanikoTemplateRaw, err := os.ReadFile(kanikoTemplatePath)
	if err != nil {
		return fmt.Errorf("read kaniko template %s: %w", kanikoTemplatePath, err)
	}
	mirrorTemplatePath := filepath.Join(repoRoot, "deploy/base/kaniko/mirror-image-job.yaml.tpl")
	mirrorTemplateRaw, mirrorTemplateErr := os.ReadFile(mirrorTemplatePath)
	if mirrorTemplateErr == nil {
		if err := s.mirrorExternalDependencies(ctx, namespace, vars, runID, stack, mirrorTemplatePath, mirrorTemplateRaw); err != nil {
			return fmt.Errorf("mirror external dependencies: %w", err)
		}
	}

	if shouldRunCodegenCheck(stack, vars) {
		codegenTemplatePath := filepath.Join(repoRoot, "deploy/base/codex-k8s/codegen-check-job.yaml.tpl")
		codegenTemplateRaw, codegenErr := os.ReadFile(codegenTemplatePath)
		if codegenErr != nil {
			return fmt.Errorf("read codegen check template %s: %w", codegenTemplatePath, codegenErr)
		}
		if err := s.runCodegenCheck(ctx, namespace, repositoryFullName, buildRef, runToken, runID, vars, codegenTemplatePath, codegenTemplateRaw); err != nil {
			return err
		}
	}

	maxParallel := parsePositiveInt(vars["CODEXK8S_KANIKO_MAX_PARALLEL"], 1)
	if maxParallel <= 0 {
		maxParallel = 1
	}
	if maxParallel > len(buildEntries) {
		maxParallel = len(buildEntries)
	}

	ctxBuild, cancel := context.WithCancel(ctx)
	defer cancel()

	resultsCh := make(chan buildImageResult, len(buildEntries))
	errCh := make(chan error, len(buildEntries))
	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup

	for _, entry := range buildEntries {
		entry := entry
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctxBuild.Done():
				return
			}
			defer func() { <-sem }()

			result, runErr := s.runKanikoBuild(ctxBuild, namespace, repositoryFullName, buildRef, runToken, runID, entry, vars, kanikoTemplatePath, kanikoTemplateRaw)
			if runErr != nil {
				errCh <- runErr
				cancel()
				return
			}
			resultsCh <- result
		}()
	}

	wg.Wait()
	close(resultsCh)
	close(errCh)

	if len(errCh) > 0 {
		return <-errCh
	}

	builtRepositories := make(map[string]struct{}, len(buildEntries))
	for result := range resultsCh {
		applyBuiltImageResult(vars, result.Name, result.ImageRef)
		if strings.TrimSpace(result.Repository) != "" {
			builtRepositories[result.Repository] = struct{}{}
		}
	}

	if err := s.cleanupBuiltImageRepositories(ctx, vars, runID, builtRepositories); err != nil {
		return err
	}
	return nil
}

func shouldRunCodegenCheck(stack *servicescfg.Stack, vars map[string]string) bool {
	if stack == nil || !strings.EqualFold(strings.TrimSpace(stack.Spec.Project), "codex-k8s") {
		return false
	}
	enabledRaw := strings.TrimSpace(vars["CODEXK8S_CODEGEN_CHECK_ENABLED"])
	if enabledRaw == "" {
		return true
	}
	enabled, err := strconv.ParseBool(strings.ToLower(enabledRaw))
	if err != nil {
		return true
	}
	return enabled
}

func (s *Service) runCodegenCheck(ctx context.Context, namespace string, repositoryFullName string, buildRef string, runToken string, runID string, vars map[string]string, templatePath string, templateRaw []byte) error {
	jobName := "codex-k8s-codegen-check-" + sanitizeNameToken(runToken, 20)
	if len(jobName) > 63 {
		jobName = strings.TrimRight(jobName[:63], "-")
	}

	jobVars := cloneStringMap(vars)
	jobVars["CODEXK8S_PRODUCTION_NAMESPACE"] = namespace
	jobVars["CODEXK8S_CODEGEN_CHECK_JOB_NAME"] = jobName
	jobVars["CODEXK8S_GITHUB_REPO"] = repositoryFullName
	jobVars["CODEXK8S_BUILD_REF"] = buildRef

	renderedRaw, err := manifesttpl.Render(templatePath, templateRaw, jobVars)
	if err != nil {
		return fmt.Errorf("render codegen check job: %w", err)
	}

	s.appendTaskLogBestEffort(ctx, runID, "codegen-check", "info", "Codegen check started")
	if err := s.k8s.DeleteJobIfExists(ctx, namespace, jobName); err != nil {
		return fmt.Errorf("delete previous codegen check job %s: %w", jobName, err)
	}
	if _, err := s.k8s.ApplyManifest(ctx, renderedRaw, namespace, s.cfg.KanikoFieldManager); err != nil {
		return fmt.Errorf("apply codegen check job %s: %w", jobName, err)
	}

	timeout := s.cfg.KanikoTimeout
	if rawTimeout := strings.TrimSpace(vars["CODEXK8S_CODEGEN_CHECK_TIMEOUT"]); rawTimeout != "" {
		parsedTimeout, parseErr := time.ParseDuration(rawTimeout)
		if parseErr == nil && parsedTimeout > 0 {
			timeout = parsedTimeout
		}
	}
	if err := s.k8s.WaitForJobComplete(ctx, namespace, jobName, timeout); err != nil {
		jobLogs, logsErr := s.k8s.GetJobLogs(ctx, namespace, jobName, s.cfg.KanikoJobLogTailLines)
		if logsErr == nil && strings.TrimSpace(jobLogs) != "" {
			s.appendTaskLogBestEffort(ctx, runID, "codegen-check", "error", "Codegen check failed logs:\n"+jobLogs)
			return fmt.Errorf("wait codegen check job %s: %w; logs: %s", jobName, err, trimLogForError(jobLogs))
		}
		return fmt.Errorf("wait codegen check job %s: %w", jobName, err)
	}

	jobLogs, logsErr := s.k8s.GetJobLogs(ctx, namespace, jobName, s.cfg.KanikoJobLogTailLines)
	if logsErr == nil && strings.TrimSpace(jobLogs) != "" {
		s.appendTaskLogBestEffort(ctx, runID, "codegen-check", "info", "Codegen check logs:\n"+jobLogs)
	}
	s.appendTaskLogBestEffort(ctx, runID, "codegen-check", "info", "Codegen check finished")
	return nil
}

func (s *Service) runKanikoBuild(ctx context.Context, namespace string, repositoryFullName string, buildRef string, runToken string, runID string, entry buildImageEntry, vars map[string]string, templatePath string, templateRaw []byte) (buildImageResult, error) {
	s.appendTaskLogBestEffort(ctx, runID, "build", "info", "Build image "+entry.Name+" started")
	repository := strings.TrimSpace(entry.Image.Repository)
	if repository == "" {
		return buildImageResult{}, fmt.Errorf("image %q repository is required for build type", entry.Name)
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
		return buildImageResult{}, fmt.Errorf("image %q: %w", entry.Name, dockerfileErr)
	}
	shouldSkip, skipErr := s.shouldSkipKanikoBuild(ctx, repository, tag, vars, runID, entry.Name)
	if skipErr != nil {
		return buildImageResult{}, skipErr
	}
	if shouldSkip {
		return buildImageResult{
			Name:       entry.Name,
			ImageRef:   destinationTagged,
			Repository: repository,
		}, nil
	}
	jobName := fmt.Sprintf("codex-k8s-kaniko-%s-%s", sanitizeNameToken(entry.Name, 24), runToken)
	if len(jobName) > 63 {
		jobName = strings.TrimRight(jobName[:63], "-")
	}
	jobVars := cloneStringMap(vars)
	jobVars["CODEXK8S_PRODUCTION_NAMESPACE"] = namespace
	jobVars["CODEXK8S_GITHUB_REPO"] = repositoryFullName
	jobVars["CODEXK8S_BUILD_REF"] = buildRef
	jobVars["CODEXK8S_KANIKO_JOB_NAME"] = jobName
	jobVars["CODEXK8S_KANIKO_COMPONENT"] = sanitizeNameToken(entry.Name, 30)
	jobVars["CODEXK8S_KANIKO_CONTEXT"] = contextArg
	jobVars["CODEXK8S_KANIKO_DOCKERFILE"] = dockerfileArg
	jobVars["CODEXK8S_KANIKO_DESTINATION_LATEST"] = destinationLatest
	jobVars["CODEXK8S_KANIKO_DESTINATION_SHA"] = destinationTagged

	runKanikoOnce := func(attemptVars map[string]string, attemptLabel string) (string, error) {
		renderedJobRaw, renderErr := manifesttpl.Render(templatePath, templateRaw, attemptVars)
		if renderErr != nil {
			return "", fmt.Errorf("render kaniko job template %s for image %s: %w", templatePath, entry.Name, renderErr)
		}
		renderedJob := string(renderedJobRaw)
		if err := s.k8s.DeleteJobIfExists(ctx, namespace, jobName); err != nil {
			return "", fmt.Errorf("delete previous kaniko job %s: %w", jobName, err)
		}
		if _, err := s.k8s.ApplyManifest(ctx, []byte(renderedJob), namespace, s.cfg.KanikoFieldManager); err != nil {
			return "", fmt.Errorf("apply kaniko job %s: %w", jobName, err)
		}
		waitErr := s.k8s.WaitForJobComplete(ctx, namespace, jobName, s.cfg.KanikoTimeout)
		jobLogs, logsErr := s.k8s.GetJobLogs(ctx, namespace, jobName, s.cfg.KanikoJobLogTailLines)
		if waitErr != nil {
			if logsErr == nil && strings.TrimSpace(jobLogs) != "" {
				return jobLogs, fmt.Errorf("wait kaniko job %s: %w; logs: %s", jobName, waitErr, trimLogForError(jobLogs))
			}
			return "", fmt.Errorf("wait kaniko job %s: %w", jobName, waitErr)
		}
		if logsErr == nil && strings.TrimSpace(jobLogs) != "" {
			logSuffix := ""
			if attemptLabel != "" {
				logSuffix = " (" + attemptLabel + ")"
			}
			s.appendTaskLogBestEffort(ctx, runID, "build", "info", "Build image "+entry.Name+" logs"+logSuffix+":\n"+jobLogs)
		}
		return jobLogs, nil
	}

	failureLogs, buildErr := runKanikoOnce(jobVars, "")
	if buildErr != nil {
		if isKanikoCacheEnabled(jobVars) && shouldRetryKanikoWithoutCache(failureLogs) {
			s.appendTaskLogBestEffort(ctx, runID, "build", "warning", "Build image "+entry.Name+" failed due stale kaniko cache, retrying without cache")
			noCacheVars := cloneStringMap(jobVars)
			noCacheVars["CODEXK8S_KANIKO_CACHE_ENABLED"] = "false"
			noCacheVars["CODEXK8S_KANIKO_CACHE_REPO"] = ""
			noCacheVars["CODEXK8S_KANIKO_CACHE_TTL"] = ""
			retryLogs, retryErr := runKanikoOnce(noCacheVars, "retry-no-cache")
			if retryErr == nil {
				if strings.TrimSpace(retryLogs) != "" {
					s.appendTaskLogBestEffort(ctx, runID, "build", "info", "Build image "+entry.Name+" retry without cache succeeded")
				}
			} else {
				if strings.TrimSpace(retryLogs) != "" {
					s.appendTaskLogBestEffort(ctx, runID, "build", "error", "Build image "+entry.Name+" retry without cache failed logs:\n"+retryLogs)
				}
				return buildImageResult{}, fmt.Errorf("build image %s retry without cache failed: %w", entry.Name, retryErr)
			}
		} else {
			if strings.TrimSpace(failureLogs) != "" {
				s.appendTaskLogBestEffort(ctx, runID, "build", "error", "Build image "+entry.Name+" failed logs:\n"+failureLogs)
			}
			return buildImageResult{}, buildErr
		}
	}

	s.appendTaskLogBestEffort(ctx, runID, "build", "info", "Build image "+entry.Name+" finished: "+destinationTagged)

	return buildImageResult{
		Name:       entry.Name,
		ImageRef:   destinationTagged,
		Repository: repository,
	}, nil
}

func (s *Service) mirrorExternalDependencies(ctx context.Context, namespace string, vars map[string]string, runID string, stack *servicescfg.Stack, templatePath string, templateRaw []byte) error {
	enabled := strings.TrimSpace(vars["CODEXK8S_IMAGE_MIRROR_ENABLED"])
	if enabled == "" {
		enabled = "true"
	}
	isEnabled, err := strconv.ParseBool(strings.ToLower(enabled))
	if err != nil || !isEnabled {
		return nil
	}
	if s.registry == nil {
		s.appendTaskLogBestEffort(ctx, runID, "mirror", "warning", "Registry client is not configured, skipping external image mirror")
		return nil
	}
	if stack == nil || len(stack.Spec.Images) == 0 {
		return nil
	}
	internalHost := strings.TrimSpace(vars["CODEXK8S_INTERNAL_REGISTRY_HOST"])
	if internalHost == "" {
		return nil
	}

	type mirrorEntry struct {
		Name  string
		Image servicescfg.Image
	}
	entries := make([]mirrorEntry, 0, len(stack.Spec.Images))
	for name, image := range stack.Spec.Images {
		if strings.EqualFold(strings.TrimSpace(image.Type), "external") {
			entries = append(entries, mirrorEntry{
				Name:  strings.TrimSpace(name),
				Image: image,
			})
		}
	}
	if len(entries) == 0 {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })

	mirrorToolImage := strings.TrimSpace(valueOr(vars, "CODEXK8S_IMAGE_MIRROR_TOOL_IMAGE", "gcr.io/go-containerregistry/crane:debug"))
	jobToken, err := randomHex(4)
	if err != nil {
		return fmt.Errorf("generate image mirror token: %w", err)
	}

	for _, entry := range entries {
		if entry.Name == "" {
			continue
		}
		envKey := imageEnvVar(entry.Name)
		sourceImage := strings.TrimSpace(entry.Image.From)
		targetImage := strings.TrimSpace(entry.Image.Local)
		if targetImage == "" {
			targetImage = resolveStackImageRef(entry.Image)
		}
		if sourceImage == "" || targetImage == "" {
			s.appendTaskLogBestEffort(ctx, runID, "mirror", "warning", "Skip mirror for "+entry.Name+": missing source/target image")
			continue
		}
		if envKey != "" {
			vars[envKey] = targetImage
		}

		targetRepo, targetTag := splitImageRef(targetImage)
		if targetRepo == "" || targetTag == "" {
			s.appendTaskLogBestEffort(ctx, runID, "mirror", "warning", "Skip mirror for "+entry.Name+": invalid target "+targetImage)
			continue
		}
		repoPath := extractRegistryRepositoryPath(targetRepo, internalHost)
		if repoPath == "" {
			s.appendTaskLogBestEffort(ctx, runID, "mirror", "warning", "Skip mirror for "+entry.Name+": invalid registry path for "+targetImage)
			continue
		}

		s.appendTaskLogBestEffort(ctx, runID, "mirror", "info", "Ensuring mirror "+sourceImage+" -> "+targetImage)

		jobName := fmt.Sprintf("codex-k8s-mirror-%s-%s", sanitizeNameToken(entry.Name, 20), jobToken)
		if len(jobName) > 63 {
			jobName = strings.TrimRight(jobName[:63], "-")
		}
		jobVars := cloneStringMap(vars)
		jobVars["CODEXK8S_PRODUCTION_NAMESPACE"] = namespace
		jobVars["CODEXK8S_IMAGE_MIRROR_JOB_NAME"] = jobName
		jobVars["CODEXK8S_IMAGE_MIRROR_SOURCE"] = sourceImage
		jobVars["CODEXK8S_IMAGE_MIRROR_TARGET"] = targetImage
		jobVars["CODEXK8S_IMAGE_MIRROR_TOOL_IMAGE"] = mirrorToolImage

		renderedRaw, renderErr := manifesttpl.Render(templatePath, templateRaw, jobVars)
		if renderErr != nil {
			return fmt.Errorf("render mirror job for %s: %w", entry.Name, renderErr)
		}
		if err := s.k8s.DeleteJobIfExists(ctx, namespace, jobName); err != nil {
			return fmt.Errorf("delete previous mirror job %s: %w", jobName, err)
		}
		if _, err := s.k8s.ApplyManifest(ctx, renderedRaw, namespace, s.cfg.KanikoFieldManager); err != nil {
			return fmt.Errorf("apply mirror job %s: %w", jobName, err)
		}
		if err := s.k8s.WaitForJobComplete(ctx, namespace, jobName, s.cfg.KanikoTimeout); err != nil {
			jobLogs, logsErr := s.k8s.GetJobLogs(ctx, namespace, jobName, s.cfg.KanikoJobLogTailLines)
			if logsErr == nil && strings.TrimSpace(jobLogs) != "" {
				s.appendTaskLogBestEffort(ctx, runID, "mirror", "error", "Mirror failed for "+targetImage+":\n"+jobLogs)
			}
			return fmt.Errorf("wait mirror job %s: %w", jobName, err)
		}
		s.appendTaskLogBestEffort(ctx, runID, "mirror", "info", "Mirrored "+sourceImage+" -> "+targetImage)
	}
	return nil
}

func (s *Service) cleanupBuiltImageRepositories(ctx context.Context, vars map[string]string, runID string, repositories map[string]struct{}) error {
	if s.registry == nil {
		return nil
	}
	if len(repositories) == 0 {
		return nil
	}
	keepTags := parsePositiveInt(vars["CODEXK8S_REGISTRY_CLEANUP_KEEP_TAGS"], s.cfg.RegistryCleanupKeepTags)
	if keepTags <= 0 {
		keepTags = s.cfg.RegistryCleanupKeepTags
	}
	if keepTags <= 0 {
		keepTags = 5
	}
	internalHost := strings.TrimSpace(vars["CODEXK8S_INTERNAL_REGISTRY_HOST"])

	for repository := range repositories {
		repoPath := extractRegistryRepositoryPath(repository, internalHost)
		if repoPath == "" {
			continue
		}
		tags, err := s.registry.ListTagInfos(ctx, repoPath)
		if err != nil {
			s.appendTaskLogBestEffort(ctx, runID, "cleanup", "warning", "List tags failed for "+repoPath+": "+err.Error())
			continue
		}
		if len(tags) <= keepTags {
			continue
		}
		for idx := keepTags; idx < len(tags); idx++ {
			tag := strings.TrimSpace(tags[idx].Tag)
			if tag == "" {
				continue
			}
			if _, err := s.registry.DeleteTag(ctx, repoPath, tag); err != nil {
				s.appendTaskLogBestEffort(ctx, runID, "cleanup", "warning", "Delete stale tag failed for "+repoPath+":"+tag+": "+err.Error())
				continue
			}
			s.appendTaskLogBestEffort(ctx, runID, "cleanup", "info", "Deleted stale tag "+repoPath+":"+tag)
		}
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

func (s *Service) shouldSkipKanikoBuild(ctx context.Context, repository string, tag string, vars map[string]string, runID string, imageName string) (bool, error) {
	if s.registry == nil {
		return false, nil
	}
	repository = strings.TrimSpace(repository)
	tag = strings.TrimSpace(tag)
	if repository == "" || tag == "" {
		return false, nil
	}
	internalHost := strings.TrimSpace(vars["CODEXK8S_INTERNAL_REGISTRY_HOST"])
	repoPath := extractRegistryRepositoryPath(repository, internalHost)
	if repoPath == "" {
		return false, nil
	}
	exists, err := s.checkRegistryTagExists(ctx, repoPath, tag, runID, "build", imageName)
	if err != nil {
		return false, fmt.Errorf("check build target %s:%s: %w", repoPath, tag, err)
	}
	if !exists {
		s.appendTaskLogBestEffort(ctx, runID, "build", "info", "Build image "+imageName+" required: missing tag "+repository+":"+tag)
		return false, nil
	}
	s.appendTaskLogBestEffort(ctx, runID, "build", "info", "Build image "+imageName+" skipped: tag exists "+repository+":"+tag)
	return true, nil
}

func (s *Service) checkRegistryTagExists(ctx context.Context, repositoryPath string, tag string, runID string, stage string, imageName string) (bool, error) {
	repositoryPath = strings.TrimSpace(repositoryPath)
	tag = strings.TrimSpace(tag)
	if repositoryPath == "" || tag == "" {
		return false, fmt.Errorf("registry repository path and tag are required")
	}
	s.appendTaskLogBestEffort(ctx, runID, stage, "info", "Registry check for "+imageName+": "+repositoryPath+":"+tag)
	info, found, err := s.registry.GetTagInfo(ctx, repositoryPath, tag)
	if err != nil {
		s.appendTaskLogBestEffort(ctx, runID, stage, "error", "Registry check failed for "+repositoryPath+":"+tag+": "+err.Error())
		return false, err
	}
	if !found {
		s.appendTaskLogBestEffort(ctx, runID, stage, "info", "Registry check result for "+imageName+": missing "+repositoryPath+":"+tag)
		return false, nil
	}
	digest := strings.TrimSpace(info.Digest)
	if digest == "" {
		s.appendTaskLogBestEffort(ctx, runID, stage, "warning", "Registry check result for "+imageName+": tag exists without digest, treat as missing "+repositoryPath+":"+tag)
		return false, nil
	}
	s.appendTaskLogBestEffort(ctx, runID, stage, "info", "Registry check result for "+imageName+": found "+repositoryPath+":"+tag+" digest="+digest)
	return true, nil
}

func parsePositiveInt(raw string, fallback int) int {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func extractRegistryRepositoryPath(imageRepository string, internalHost string) string {
	repository := strings.TrimSpace(imageRepository)
	host := strings.TrimSpace(internalHost)
	if repository == "" {
		return ""
	}
	repository = strings.TrimPrefix(repository, "http://")
	repository = strings.TrimPrefix(repository, "https://")
	if host == "" {
		return repository
	}
	prefix := host + "/"
	if strings.HasPrefix(repository, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(repository, prefix))
	}
	return ""
}

func splitImageRef(ref string) (string, string) {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return "", ""
	}
	if at := strings.Index(trimmed, "@"); at >= 0 {
		trimmed = trimmed[:at]
	}
	lastSlash := strings.LastIndex(trimmed, "/")
	lastColon := strings.LastIndex(trimmed, ":")
	if lastColon == -1 || lastColon < lastSlash {
		return trimmed, ""
	}
	return trimmed[:lastColon], trimmed[lastColon+1:]
}

func isKanikoCacheEnabled(vars map[string]string) bool {
	value := strings.TrimSpace(vars["CODEXK8S_KANIKO_CACHE_ENABLED"])
	if value == "" {
		return true
	}
	enabled, err := strconv.ParseBool(strings.ToLower(value))
	if err != nil {
		return true
	}
	return enabled
}

func shouldRetryKanikoWithoutCache(jobLogs string) bool {
	lower := strings.ToLower(strings.TrimSpace(jobLogs))
	if lower == "" {
		return false
	}
	return strings.Contains(lower, "error while retrieving image from cache") ||
		(strings.Contains(lower, "manifest_unknown") && strings.Contains(lower, "from cache"))
}

func trimLogForError(logs string) string {
	trimmed := strings.TrimSpace(logs)
	if len(trimmed) > 500 {
		return trimmed[:500] + "..."
	}
	return trimmed
}
