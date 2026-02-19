package runner

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

type runWriteScopeMode string

const (
	runWriteScopeModeAny             runWriteScopeMode = "any"
	runWriteScopeModeMarkdownOnly    runWriteScopeMode = "markdown_only"
	runWriteScopeModeNoRepoChanges   runWriteScopeMode = "no_repo_changes"
	runWriteScopeModeSelfImproveOnly runWriteScopeMode = "self_improve_only"
)

type runWriteScopePolicy struct {
	Mode              runWriteScopeMode
	RequireExistingPR bool
}

var markdownOnlyTriggerKinds = map[webhookdomain.TriggerKind]struct{}{
	webhookdomain.TriggerKindIntake:       {},
	webhookdomain.TriggerKindIntakeRevise: {},
	webhookdomain.TriggerKindVision:       {},
	webhookdomain.TriggerKindVisionRevise: {},
	webhookdomain.TriggerKindPRD:          {},
	webhookdomain.TriggerKindPRDRevise:    {},
	webhookdomain.TriggerKindArch:         {},
	webhookdomain.TriggerKindArchRevise:   {},
	webhookdomain.TriggerKindDesign:       {},
	webhookdomain.TriggerKindDesignRevise: {},
	webhookdomain.TriggerKindPlan:         {},
	webhookdomain.TriggerKindPlanRevise:   {},
	webhookdomain.TriggerKindDocAudit:     {},
	webhookdomain.TriggerKindQA:           {},
	webhookdomain.TriggerKindRelease:      {},
	webhookdomain.TriggerKindPostDeploy:   {},
	webhookdomain.TriggerKindOps:          {},
	webhookdomain.TriggerKindRethink:      {},
}

func resolveRunWriteScopePolicy(triggerKind string, agentKey string) runWriteScopePolicy {
	normalizedAgentKey := strings.ToLower(strings.TrimSpace(agentKey))
	normalizedTriggerKind := webhookdomain.NormalizeTriggerKind(triggerKind)

	if normalizedTriggerKind == webhookdomain.TriggerKindSelfImprove {
		return runWriteScopePolicy{Mode: runWriteScopeModeSelfImproveOnly}
	}
	if normalizedAgentKey == "reviewer" {
		return runWriteScopePolicy{
			Mode:              runWriteScopeModeNoRepoChanges,
			RequireExistingPR: true,
		}
	}
	if _, ok := markdownOnlyTriggerKinds[normalizedTriggerKind]; ok {
		return runWriteScopePolicy{Mode: runWriteScopeModeMarkdownOnly}
	}
	return runWriteScopePolicy{Mode: runWriteScopeModeAny}
}

func isMarkdownOnlyScope(triggerKind string, agentKey string) bool {
	return resolveRunWriteScopePolicy(triggerKind, agentKey).Mode == runWriteScopeModeMarkdownOnly
}

func isReviewerCommentOnlyScope(triggerKind string, agentKey string) bool {
	return resolveRunWriteScopePolicy(triggerKind, agentKey).Mode == runWriteScopeModeNoRepoChanges
}

func isSelfImproveRestrictedScope(triggerKind string, agentKey string) bool {
	return resolveRunWriteScopePolicy(triggerKind, agentKey).Mode == runWriteScopeModeSelfImproveOnly
}

func enforceRunWriteScope(ctx context.Context, repoDir string, baselineHead string, triggerKind string, agentKey string, existingPRNumber int) error {
	policy := resolveRunWriteScopePolicy(triggerKind, agentKey)
	if policy.RequireExistingPR && existingPRNumber <= 0 {
		return fmt.Errorf("failed_precondition: reviewer run requires existing PR context")
	}
	if policy.Mode == runWriteScopeModeAny {
		return nil
	}

	currentHead, err := gitCurrentHead(ctx, repoDir)
	if err != nil {
		return err
	}
	changedPaths, err := collectChangedPathsSince(ctx, repoDir, baselineHead, currentHead)
	if err != nil {
		return err
	}

	switch policy.Mode {
	case runWriteScopeModeNoRepoChanges:
		if strings.TrimSpace(baselineHead) != "" && strings.TrimSpace(currentHead) != strings.TrimSpace(baselineHead) {
			return fmt.Errorf("failed_precondition: reviewer mode forbids new commits in repository")
		}
		if len(changedPaths) > 0 {
			return fmt.Errorf("failed_precondition: reviewer mode forbids repository file changes, got: %s", formatChangedPathList(changedPaths, 20))
		}
	case runWriteScopeModeMarkdownOnly:
		invalid := collectInvalidPaths(changedPaths, isMarkdownDocumentationPath)
		if len(invalid) > 0 {
			return fmt.Errorf("failed_precondition: trigger %q allows markdown docs only, forbidden paths: %s", webhookdomain.NormalizeTriggerKind(triggerKind), formatChangedPathList(invalid, 20))
		}
	case runWriteScopeModeSelfImproveOnly:
		invalid := collectInvalidPaths(changedPaths, isSelfImproveAllowedPath)
		if len(invalid) > 0 {
			return fmt.Errorf(
				"failed_precondition: run:self-improve allows only prompts/instructions/agent-runner Dockerfile changes, forbidden paths: %s",
				formatChangedPathList(invalid, 20),
			)
		}
	}

	return nil
}

func gitCurrentHead(ctx context.Context, repoDir string) (string, error) {
	output, err := runCommandCaptureCombinedOutput(ctx, repoDir, "git", "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("resolve current git HEAD: %w", err)
	}
	head := strings.TrimSpace(output)
	if head == "" {
		return "", fmt.Errorf("resolve current git HEAD: empty")
	}
	return head, nil
}

func collectChangedPathsSince(ctx context.Context, repoDir string, baselineHead string, currentHead string) ([]string, error) {
	paths := make([]string, 0, 16)
	if strings.TrimSpace(baselineHead) != "" && strings.TrimSpace(currentHead) != "" && strings.TrimSpace(baselineHead) != strings.TrimSpace(currentHead) {
		committed, err := runCommandCaptureCombinedOutput(ctx, repoDir, "git", "diff", "--name-only", strings.TrimSpace(baselineHead)+".."+strings.TrimSpace(currentHead))
		if err != nil {
			return nil, fmt.Errorf("collect committed changed files: %w", err)
		}
		paths = append(paths, splitGitPathOutput(committed)...)
	}

	unstaged, err := runCommandCaptureCombinedOutput(ctx, repoDir, "git", "diff", "--name-only")
	if err != nil {
		return nil, fmt.Errorf("collect unstaged changed files: %w", err)
	}
	paths = append(paths, splitGitPathOutput(unstaged)...)

	staged, err := runCommandCaptureCombinedOutput(ctx, repoDir, "git", "diff", "--name-only", "--cached")
	if err != nil {
		return nil, fmt.Errorf("collect staged changed files: %w", err)
	}
	paths = append(paths, splitGitPathOutput(staged)...)

	untracked, err := runCommandCaptureCombinedOutput(ctx, repoDir, "git", "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, fmt.Errorf("collect untracked files: %w", err)
	}
	paths = append(paths, splitGitPathOutput(untracked)...)

	if len(paths) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(paths))
	unique := make([]string, 0, len(paths))
	for _, raw := range paths {
		normalized := normalizeRepoRelativePath(raw)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		unique = append(unique, normalized)
	}
	sort.Strings(unique)
	return unique, nil
}

func splitGitPathOutput(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		path := strings.TrimSpace(line)
		if path == "" {
			continue
		}
		result = append(result, path)
	}
	return result
}

func normalizeRepoRelativePath(path string) string {
	normalized := filepath.ToSlash(strings.TrimSpace(path))
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimPrefix(normalized, "/")
	return normalized
}

func collectInvalidPaths(paths []string, isAllowed func(string) bool) []string {
	if len(paths) == 0 {
		return nil
	}
	invalid := make([]string, 0, len(paths))
	for _, path := range paths {
		if isAllowed(path) {
			continue
		}
		invalid = append(invalid, path)
	}
	return invalid
}

func isMarkdownDocumentationPath(path string) bool {
	lower := strings.ToLower(strings.TrimSpace(path))
	return strings.HasSuffix(lower, ".md")
}

func isSelfImproveAllowedPath(path string) bool {
	normalized := normalizeRepoRelativePath(path)
	lower := strings.ToLower(normalized)

	if isMarkdownDocumentationPath(normalized) {
		return true
	}
	if strings.HasPrefix(lower, "services/jobs/agent-runner/internal/runner/promptseeds/") {
		return true
	}
	if lower == "services/jobs/agent-runner/internal/runner/templates/prompt_envelope.tmpl" {
		return true
	}
	if lower == "services/jobs/agent-runner/dockerfile" {
		return true
	}
	return false
}

func formatChangedPathList(paths []string, limit int) string {
	if len(paths) == 0 {
		return ""
	}
	normalizedLimit := limit
	if normalizedLimit <= 0 {
		normalizedLimit = 10
	}

	if len(paths) <= normalizedLimit {
		return strings.Join(paths, ", ")
	}
	items := slices.Clone(paths[:normalizedLimit])
	items = append(items, fmt.Sprintf("... and %d more", len(paths)-normalizedLimit))
	return strings.Join(items, ", ")
}
