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

	"github.com/codex-k8s/kodex/libs/go/manifesttpl"
	"github.com/codex-k8s/kodex/libs/go/registry"
	"github.com/codex-k8s/kodex/libs/go/servicescfg"
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

type mirrorJobSpec struct {
	EntryName   string
	SourceImage string
	TargetImage string
	JobName     string
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

	prebuilt, allPrebuilt, err := s.resolveReusableBuildImages(ctx, buildEntries, vars, runID)
	if err != nil {
		return err
	}
	if allPrebuilt {
		for _, result := range prebuilt {
			applyBuiltImageResult(vars, result.Name, result.ImageRef)
		}
		s.appendTaskLogBestEffort(ctx, runID, "build", "info", "All build images already exist in registry, skipping mirror/codegen/kaniko stages")
		return nil
	}

	githubPAT := strings.TrimSpace(s.cfg.GitHubPAT)
	if githubPAT == "" {
		githubPAT = strings.TrimSpace(vars["KODEX_GITHUB_PAT"])
	}
	if githubPAT == "" {
		return fmt.Errorf("KODEX_GITHUB_PAT is required for kaniko build jobs")
	}
	repositoryFullName := strings.TrimSpace(params.RepositoryFullName)
	if repositoryFullName == "" {
		repositoryFullName = strings.TrimSpace(vars["KODEX_GITHUB_REPO"])
	}
	if repositoryFullName == "" {
		return fmt.Errorf("repository_full_name is required for kaniko build jobs")
	}
	buildRef := resolveRuntimeBuildRef(
		params.BuildRef,
		vars["KODEX_BUILD_REF"],
		vars["KODEX_AGENT_BASE_BRANCH"],
	)
	checkoutBuildRef := normalizeRuntimeBuildCheckoutRef(buildRef)
	vars["KODEX_BUILD_REF"] = buildRef

	runToken := sanitizeNameToken(params.RunID, 12)
	if runToken == "" {
		generatedToken, tokenErr := randomHex(6)
		if tokenErr != nil {
			return fmt.Errorf("generate kaniko run token: %w", tokenErr)
		}
		runToken = generatedToken
	}

	if err := s.k8s.UpsertSecret(ctx, namespace, "kodex-git-token", map[string][]byte{
		"token": []byte(githubPAT),
	}); err != nil {
		return fmt.Errorf("upsert kodex-git-token secret: %w", err)
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
		codegenTemplatePath := filepath.Join(repoRoot, "deploy/base/kodex/codegen-check-job.yaml.tpl")
		codegenTemplateRaw, codegenErr := os.ReadFile(codegenTemplatePath)
		if codegenErr != nil {
			return fmt.Errorf("read codegen check template %s: %w", codegenTemplatePath, codegenErr)
		}
		if err := s.runCodegenCheck(ctx, namespace, repositoryFullName, checkoutBuildRef, runToken, runID, vars, codegenTemplatePath, codegenTemplateRaw); err != nil {
			return err
		}
	}

	maxParallel := parsePositiveInt(vars["KODEX_KANIKO_MAX_PARALLEL"], 1)
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

			result, runErr := s.runKanikoBuild(ctxBuild, namespace, repositoryFullName, checkoutBuildRef, runToken, runID, entry, vars, kanikoTemplatePath, kanikoTemplateRaw)
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
	if stack == nil || !strings.EqualFold(strings.TrimSpace(stack.Spec.Project), "kodex") {
		return false
	}
	enabledRaw := strings.TrimSpace(vars["KODEX_CODEGEN_CHECK_ENABLED"])
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
	jobName := "kodex-codegen-check-" + sanitizeNameToken(runToken, 20)
	if len(jobName) > 63 {
		jobName = strings.TrimRight(jobName[:63], "-")
	}

	jobVars := cloneStringMap(vars)
	jobVars["KODEX_PRODUCTION_NAMESPACE"] = namespace
	jobVars["KODEX_CODEGEN_CHECK_JOB_NAME"] = jobName
	jobVars["KODEX_GITHUB_REPO"] = repositoryFullName
	jobVars["KODEX_BUILD_REF"] = buildRef

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
	if rawTimeout := strings.TrimSpace(vars["KODEX_CODEGEN_CHECK_TIMEOUT"]); rawTimeout != "" {
		parsedTimeout, parseErr := time.ParseDuration(rawTimeout)
		if parseErr == nil && parsedTimeout > 0 {
			timeout = parsedTimeout
		}
	}
	if err := s.waitForJobCompletionWithFailureLogs(
		ctx,
		namespace,
		jobName,
		timeout,
		runID,
		"codegen-check",
		"wait codegen check job",
		"Codegen check failed logs:",
	); err != nil {
		return err
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
	jobName := fmt.Sprintf("kodex-kaniko-%s-%s", sanitizeNameToken(entry.Name, 24), runToken)
	if len(jobName) > 63 {
		jobName = strings.TrimRight(jobName[:63], "-")
	}
	jobVars := cloneStringMap(vars)
	jobVars["KODEX_PRODUCTION_NAMESPACE"] = namespace
	jobVars["KODEX_GITHUB_REPO"] = repositoryFullName
	jobVars["KODEX_BUILD_REF"] = buildRef
	jobVars["KODEX_KANIKO_JOB_NAME"] = jobName
	jobVars["KODEX_KANIKO_COMPONENT"] = sanitizeNameToken(entry.Name, 30)
	jobVars["KODEX_KANIKO_CONTEXT"] = contextArg
	jobVars["KODEX_KANIKO_DOCKERFILE"] = dockerfileArg
	jobVars["KODEX_KANIKO_DESTINATION_LATEST"] = destinationLatest
	jobVars["KODEX_KANIKO_DESTINATION_SHA"] = destinationTagged

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
			noCacheVars["KODEX_KANIKO_CACHE_ENABLED"] = "false"
			noCacheVars["KODEX_KANIKO_CACHE_REPO"] = ""
			noCacheVars["KODEX_KANIKO_CACHE_TTL"] = ""
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
	enabled := strings.TrimSpace(vars["KODEX_IMAGE_MIRROR_ENABLED"])
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
	internalHost := strings.TrimSpace(vars["KODEX_INTERNAL_REGISTRY_HOST"])
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

	mirrorToolImage := strings.TrimSpace(valueOr(vars, "KODEX_IMAGE_MIRROR_TOOL_IMAGE", "gcr.io/go-containerregistry/crane:debug"))
	jobToken, err := randomHex(4)
	if err != nil {
		return fmt.Errorf("generate image mirror token: %w", err)
	}
	maxParallel := parsePositiveInt(vars["KODEX_IMAGE_MIRROR_MAX_PARALLEL"], 1)
	if maxParallel <= 0 {
		maxParallel = 1
	}

	jobs := make([]mirrorJobSpec, 0, len(entries))
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

		targetRepo, targetTag := registry.SplitImageRef(targetImage)
		if targetRepo == "" || targetTag == "" {
			s.appendTaskLogBestEffort(ctx, runID, "mirror", "warning", "Skip mirror for "+entry.Name+": invalid target "+targetImage)
			continue
		}
		repoPath := registry.ExtractRepositoryPath(targetRepo, internalHost)
		if repoPath == "" {
			s.appendTaskLogBestEffort(ctx, runID, "mirror", "warning", "Skip mirror for "+entry.Name+": invalid registry path for "+targetImage)
			continue
		}

		jobName := fmt.Sprintf("kodex-mirror-%s-%s", sanitizeNameToken(entry.Name, 20), jobToken)
		if len(jobName) > 63 {
			jobName = strings.TrimRight(jobName[:63], "-")
		}
		jobs = append(jobs, mirrorJobSpec{
			EntryName:   entry.Name,
			SourceImage: sourceImage,
			TargetImage: targetImage,
			JobName:     jobName,
		})
	}
	if len(jobs) == 0 {
		return nil
	}
	if maxParallel > len(jobs) {
		maxParallel = len(jobs)
	}

	ctxMirror, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, len(jobs))
	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup

	for _, job := range jobs {
		job := job
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctxMirror.Done():
				return
			}
			defer func() { <-sem }()

			if runErr := s.runMirrorJob(ctxMirror, namespace, vars, runID, templatePath, templateRaw, mirrorToolImage, job); runErr != nil {
				errCh <- runErr
				cancel()
			}
		}()
	}
	wg.Wait()
	close(errCh)
	if len(errCh) > 0 {
		return <-errCh
	}
	return nil
}

func (s *Service) runMirrorJob(ctx context.Context, namespace string, vars map[string]string, runID string, templatePath string, templateRaw []byte, mirrorToolImage string, job mirrorJobSpec) error {
	s.appendTaskLogBestEffort(ctx, runID, "mirror", "info", "Ensuring mirror "+job.SourceImage+" -> "+job.TargetImage)

	jobVars := cloneStringMap(vars)
	jobVars["KODEX_PRODUCTION_NAMESPACE"] = namespace
	jobVars["KODEX_IMAGE_MIRROR_JOB_NAME"] = job.JobName
	jobVars["KODEX_IMAGE_MIRROR_SOURCE"] = job.SourceImage
	jobVars["KODEX_IMAGE_MIRROR_TARGET"] = job.TargetImage
	jobVars["KODEX_IMAGE_MIRROR_TOOL_IMAGE"] = mirrorToolImage

	renderedRaw, renderErr := manifesttpl.Render(templatePath, templateRaw, jobVars)
	if renderErr != nil {
		return fmt.Errorf("render mirror job for %s: %w", job.EntryName, renderErr)
	}
	if err := s.k8s.DeleteJobIfExists(ctx, namespace, job.JobName); err != nil {
		return fmt.Errorf("delete previous mirror job %s: %w", job.JobName, err)
	}
	if _, err := s.k8s.ApplyManifest(ctx, renderedRaw, namespace, s.cfg.KanikoFieldManager); err != nil {
		return fmt.Errorf("apply mirror job %s: %w", job.JobName, err)
	}
	if err := s.k8s.WaitForJobComplete(ctx, namespace, job.JobName, s.cfg.KanikoTimeout); err != nil {
		jobLogs, logsErr := s.k8s.GetJobLogs(ctx, namespace, job.JobName, s.cfg.KanikoJobLogTailLines)
		if logsErr == nil && strings.TrimSpace(jobLogs) != "" {
			s.appendTaskLogBestEffort(ctx, runID, "mirror", "error", "Mirror failed for "+job.TargetImage+":\n"+jobLogs)
		}
		return fmt.Errorf("wait mirror job %s: %w", job.JobName, err)
	}
	s.appendTaskLogBestEffort(ctx, runID, "mirror", "info", "Mirrored "+job.SourceImage+" -> "+job.TargetImage)
	return nil
}

func (s *Service) cleanupBuiltImageRepositories(ctx context.Context, vars map[string]string, runID string, repositories map[string]struct{}) error {
	if s.registry == nil {
		return nil
	}
	if len(repositories) == 0 {
		return nil
	}
	keepTags := parsePositiveInt(vars["KODEX_REGISTRY_CLEANUP_KEEP_TAGS"], s.cfg.RegistryCleanupKeepTags)
	if keepTags <= 0 {
		keepTags = s.cfg.RegistryCleanupKeepTags
	}
	if keepTags <= 0 {
		keepTags = 5
	}
	internalHost := strings.TrimSpace(vars["KODEX_INTERNAL_REGISTRY_HOST"])

	for repository := range repositories {
		repoPath := registry.ExtractRepositoryPath(repository, internalHost)
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
		vars["KODEX_API_GATEWAY_IMAGE"] = imageRef
	case "control-plane":
		vars["KODEX_CONTROL_PLANE_IMAGE"] = imageRef
	case "worker":
		vars["KODEX_WORKER_IMAGE"] = imageRef
	case "agent-runner":
		vars["KODEX_AGENT_RUNNER_IMAGE"] = imageRef
		if strings.TrimSpace(vars["KODEX_WORKER_JOB_IMAGE"]) == "" {
			vars["KODEX_WORKER_JOB_IMAGE"] = imageRef
		}
	case "web-console":
		vars["KODEX_WEB_CONSOLE_IMAGE"] = imageRef
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
	internalHost := strings.TrimSpace(vars["KODEX_INTERNAL_REGISTRY_HOST"])
	repoPath := registry.ExtractRepositoryPath(repository, internalHost)
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

func isKanikoCacheEnabled(vars map[string]string) bool {
	value := strings.TrimSpace(vars["KODEX_KANIKO_CACHE_ENABLED"])
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

func (s *Service) resolveReusableBuildImages(ctx context.Context, buildEntries []buildImageEntry, vars map[string]string, runID string) ([]buildImageResult, bool, error) {
	if len(buildEntries) == 0 {
		return nil, true, nil
	}
	if s.registry == nil {
		return nil, false, nil
	}

	results := make([]buildImageResult, 0, len(buildEntries))
	for _, entry := range buildEntries {
		repository := strings.TrimSpace(entry.Image.Repository)
		if repository == "" {
			return nil, false, fmt.Errorf("image %q repository is required for build type", entry.Name)
		}
		tag := sanitizeImageTag(strings.TrimSpace(entry.Image.TagTemplate))
		if tag == "" {
			tag = "latest"
		}

		shouldSkip, err := s.shouldSkipKanikoBuild(ctx, repository, tag, vars, runID, entry.Name)
		if err != nil {
			return nil, false, err
		}
		if !shouldSkip {
			return nil, false, nil
		}

		results = append(results, buildImageResult{
			Name:       entry.Name,
			ImageRef:   repository + ":" + tag,
			Repository: repository,
		})
	}
	return results, len(results) == len(buildEntries), nil
}
