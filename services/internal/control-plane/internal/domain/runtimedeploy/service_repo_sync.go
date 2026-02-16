package runtimedeploy

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/manifesttpl"
)

const (
	repoSyncGitTokenSecretName = "codex-k8s-git-token"

	repoSyncJobTemplatePath  = "deploy/base/codex-k8s/repo-sync-job.yaml.tpl"
	repoCachePVCTemplatePath = "deploy/base/codex-k8s/repo-cache-pvc.yaml.tpl"

	defaultRepoCachePVCName = "codex-k8s-repo-cache"
	defaultRepoSyncTimeout  = 10 * time.Minute
)

var (
	// Keep this template in sync with deploy/base/codex-k8s/repo-sync-job.yaml.tpl
	//go:embed assets/repo-sync-job.yaml.tpl
	embeddedRepoSyncJobTemplate []byte

	// Keep this template in sync with deploy/base/codex-k8s/repo-cache-pvc.yaml.tpl
	//go:embed assets/repo-cache-pvc.yaml.tpl
	embeddedRepoCachePVCTemplate []byte
)

func (s *Service) resolveRunRepositoryRoot(ctx context.Context, params PrepareParams, vars map[string]string, runID string) (string, error) {
	configuredRoot := strings.TrimSpace(s.cfg.RepositoryRoot)
	if configuredRoot == "" {
		return s.cfg.RepositoryRoot, nil
	}
	// Prefer "direct filesystem" mode when the configured root already contains deploy/templates.
	// This keeps runtime-deploy CLI (repository-root=/opt/codex-k8s) working and avoids an
	// unnecessary repo-sync roundtrip when the image already ships sources.
	if looksLikeRepositoryRoot(configuredRoot) {
		return configuredRoot, nil
	}
	// Keep local/dev mode shell-free: do not attempt repo-sync when the root is relative.
	if !filepath.IsAbs(configuredRoot) {
		return configuredRoot, nil
	}

	repositoryFullName := strings.TrimSpace(params.RepositoryFullName)
	if repositoryFullName == "" {
		repositoryFullName = strings.TrimSpace(valueOr(vars, "CODEXK8S_GITHUB_REPO", ""))
	}
	if repositoryFullName == "" {
		return "", fmt.Errorf("repository_full_name is required to resolve repository snapshot")
	}
	owner, name, ok := strings.Cut(repositoryFullName, "/")
	if !ok || strings.TrimSpace(owner) == "" || strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("repository_full_name must be in owner/name form, got %q", repositoryFullName)
	}
	owner = strings.TrimSpace(owner)
	name = strings.TrimSpace(name)

	buildRef := strings.TrimSpace(params.BuildRef)
	if buildRef == "" {
		buildRef = strings.TrimSpace(valueOr(vars, "CODEXK8S_BUILD_REF", ""))
	}
	if buildRef == "" {
		buildRef = strings.TrimSpace(valueOr(vars, "CODEXK8S_AGENT_BASE_BRANCH", "main"))
	}
	if buildRef == "" {
		buildRef = "main"
	}

	platformNamespace := strings.TrimSpace(valueOr(vars, "CODEXK8S_PLATFORM_NAMESPACE", ""))
	if platformNamespace == "" {
		platformNamespace = strings.TrimSpace(valueOr(vars, "CODEXK8S_PRODUCTION_NAMESPACE", ""))
	}
	if platformNamespace == "" {
		return "", fmt.Errorf("CODEXK8S_PLATFORM_NAMESPACE is required for repo sync")
	}

	repoRoot := s.repoSnapshotPath(owner, name, buildRef)
	if repoRoot == "" {
		return "", fmt.Errorf("resolve repository snapshot path: empty")
	}

	if _, err := os.Stat(filepath.Join(repoRoot, ".git")); err == nil {
		return repoRoot, nil
	}

	if err := s.ensureRepoCachePVC(ctx, platformNamespace, vars, runID); err != nil {
		return "", err
	}
	if err := s.ensureRepoSnapshot(ctx, platformNamespace, repositoryFullName, buildRef, repoRoot, vars, runID); err != nil {
		return "", err
	}

	if _, err := os.Stat(filepath.Join(repoRoot, ".git")); err != nil {
		return "", fmt.Errorf("repo snapshot is missing after repo sync: %s: %w", repoRoot, err)
	}

	return repoRoot, nil
}

func looksLikeRepositoryRoot(root string) bool {
	root = strings.TrimSpace(root)
	if root == "" {
		return false
	}
	if stat, err := os.Stat(filepath.Join(root, "deploy", "base")); err == nil && stat.IsDir() {
		return true
	}
	if _, err := os.Stat(filepath.Join(root, "services.yaml")); err == nil {
		return true
	}
	return false
}

func (s *Service) repoSnapshotPath(owner string, name string, buildRef string) string {
	cacheRoot := strings.TrimSpace(s.cfg.RepositoryRoot)
	if cacheRoot == "" {
		return ""
	}
	refToken := sanitizeNameToken(buildRef, 120)
	if refToken == "" {
		refToken = "main"
	}
	// Layout: <cacheRoot>/github/<owner>/<repo>/<refToken>
	return filepath.Join(cacheRoot, "github", owner, name, refToken)
}

func (s *Service) ensureRepoCachePVC(ctx context.Context, platformNamespace string, vars map[string]string, runID string) error {
	namespace := strings.TrimSpace(platformNamespace)
	if namespace == "" {
		return fmt.Errorf("platform namespace is required")
	}

	renderVars := cloneStringMap(vars)
	renderVars["CODEXK8S_PRODUCTION_NAMESPACE"] = namespace
	renderVars["CODEXK8S_PLATFORM_NAMESPACE"] = namespace

	rendered, err := manifesttpl.Render(repoCachePVCTemplatePath, embeddedRepoCachePVCTemplate, renderVars)
	if err != nil {
		return fmt.Errorf("render repo cache pvc manifest: %w", err)
	}
	if _, err := s.k8s.ApplyManifest(ctx, rendered, namespace, s.cfg.KanikoFieldManager); err != nil {
		return fmt.Errorf("apply repo cache pvc manifest: %w", err)
	}
	s.appendTaskLogBestEffort(ctx, runID, "repo-cache", "info", "Repo cache PVC ensured in namespace "+namespace)
	return nil
}

func (s *Service) ensureRepoSnapshot(ctx context.Context, platformNamespace string, repositoryFullName string, buildRef string, repoRoot string, vars map[string]string, runID string) error {
	namespace := strings.TrimSpace(platformNamespace)
	if namespace == "" {
		return fmt.Errorf("platform namespace is required")
	}

	repositoryFullName = strings.TrimSpace(repositoryFullName)
	if repositoryFullName == "" {
		return fmt.Errorf("repository_full_name is required")
	}
	buildRef = strings.TrimSpace(buildRef)
	if buildRef == "" {
		buildRef = "main"
	}
	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == "" {
		return fmt.Errorf("repo_root is required")
	}

	// If snapshot is already present, skip the job.
	if _, err := os.Stat(filepath.Join(repoRoot, ".git")); err == nil {
		return nil
	}

	token := strings.TrimSpace(s.cfg.GitHubPAT)
	if token == "" {
		token = strings.TrimSpace(valueOr(vars, "CODEXK8S_GITHUB_PAT", ""))
	}
	if token == "" {
		return fmt.Errorf("CODEXK8S_GITHUB_PAT is required for repo sync")
	}

	if err := s.k8s.UpsertSecret(ctx, namespace, repoSyncGitTokenSecretName, map[string][]byte{
		"token": []byte(token),
	}); err != nil {
		return fmt.Errorf("upsert %s secret: %w", repoSyncGitTokenSecretName, err)
	}

	runToken := sanitizeNameToken(runID, 12)
	if runToken == "" {
		generated, err := randomHex(6)
		if err != nil {
			return fmt.Errorf("generate repo sync run token: %w", err)
		}
		runToken = generated
	}

	jobName := "codex-k8s-repo-sync-" + sanitizeNameToken(repositoryFullName, 20) + "-" + runToken
	if len(jobName) > 63 {
		jobName = strings.TrimRight(jobName[:63], "-")
	}

	jobVars := cloneStringMap(vars)
	jobVars["CODEXK8S_PLATFORM_NAMESPACE"] = namespace
	jobVars["CODEXK8S_PRODUCTION_NAMESPACE"] = namespace
	jobVars["CODEXK8S_REPO_SYNC_JOB_NAME"] = jobName
	jobVars["CODEXK8S_REPO_SYNC_DEST_DIR"] = repoRoot
	jobVars["CODEXK8S_REPO_CACHE_PVC_NAME"] = defaultRepoCachePVCName
	jobVars["CODEXK8S_REPOSITORY_ROOT"] = strings.TrimSpace(s.cfg.RepositoryRoot)
	jobVars["CODEXK8S_GITHUB_REPO"] = repositoryFullName
	jobVars["CODEXK8S_BUILD_REF"] = buildRef

	rendered, err := manifesttpl.Render(repoSyncJobTemplatePath, embeddedRepoSyncJobTemplate, jobVars)
	if err != nil {
		return fmt.Errorf("render repo sync job: %w", err)
	}

	s.appendTaskLogBestEffort(ctx, runID, "repo-sync", "info", "Repo sync started for "+repositoryFullName+" ref "+buildRef)
	if err := s.k8s.DeleteJobIfExists(ctx, namespace, jobName); err != nil {
		return fmt.Errorf("delete previous repo sync job %s: %w", jobName, err)
	}
	if _, err := s.k8s.ApplyManifest(ctx, rendered, namespace, s.cfg.KanikoFieldManager); err != nil {
		return fmt.Errorf("apply repo sync job %s: %w", jobName, err)
	}

	timeout := defaultRepoSyncTimeout
	if timeout <= 0 {
		timeout = s.cfg.KanikoTimeout
	}
	if err := s.k8s.WaitForJobComplete(ctx, namespace, jobName, timeout); err != nil {
		jobLogs, logsErr := s.k8s.GetJobLogs(ctx, namespace, jobName, s.cfg.KanikoJobLogTailLines)
		if logsErr == nil && strings.TrimSpace(jobLogs) != "" {
			s.appendTaskLogBestEffort(ctx, runID, "repo-sync", "error", "Repo sync failed logs:\n"+jobLogs)
			return fmt.Errorf("wait repo sync job %s: %w; logs: %s", jobName, err, trimLogForError(jobLogs))
		}
		return fmt.Errorf("wait repo sync job %s: %w", jobName, err)
	}

	s.appendTaskLogBestEffort(ctx, runID, "repo-sync", "info", "Repo sync finished for "+repositoryFullName)
	return nil
}
