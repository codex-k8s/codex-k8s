package runner

import (
	"fmt"
	"path/filepath"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

func normalizeTriggerKind(value string) string {
	return string(webhookdomain.NormalizeTriggerKind(value))
}

func normalizeRuntimeMode(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), runtimeModeFullEnv) {
		return runtimeModeFullEnv
	}
	return runtimeModeCodeOnly
}

func usesPreparedFullEnvRepository(runtimeMode string) bool {
	return normalizeRuntimeMode(runtimeMode) == runtimeModeFullEnv
}

func runnerWorkspaceDir(runtimeMode string, agentKey string, targetBranch string) string {
	if !usesPreparedFullEnvRepository(runtimeMode) {
		return "/workspace"
	}
	agentToken := sanitizePathComponent(agentKey)
	if agentToken == "" {
		agentToken = "agent"
	}
	branchToken := sanitizePathComponent(targetBranch)
	if branchToken == "" {
		branchToken = "branch"
	}
	return filepath.Join("/tmp", "codex-runner", agentToken, branchToken)
}

func runnerRepoDir(runtimeMode string, agentKey string, targetBranch string) string {
	if !usesPreparedFullEnvRepository(runtimeMode) {
		return filepath.Join("/workspace", "repo")
	}
	return "/workspace"
}

func gitCleanArgs(runtimeMode string) []string {
	return []string{"clean", "-fdx"}
}

func normalizeTemplateKind(value string, triggerKind string) string {
	if strings.EqualFold(strings.TrimSpace(value), promptTemplateKindDiscussion) {
		return promptTemplateKindDiscussion
	}
	normalizedTrigger := webhookdomain.NormalizeTriggerKind(triggerKind)
	if webhookdomain.IsReviseTriggerKind(normalizedTrigger) {
		return promptTemplateKindRevise
	}
	if strings.EqualFold(strings.TrimSpace(value), promptTemplateKindRevise) {
		return promptTemplateKindRevise
	}
	return promptTemplateKindWork
}

func buildTargetBranch(explicitBranch string, runID string, issueNumber int64, triggerKind string, baseBranch string) string {
	trimmedExplicit := strings.TrimSpace(explicitBranch)
	if trimmedExplicit != "" {
		return trimmedExplicit
	}
	if isAIRepairMainDirectTrigger(triggerKind) {
		base := strings.TrimSpace(baseBranch)
		if base != "" {
			return base
		}
		return "main"
	}
	if issueNumber > 0 {
		return fmt.Sprintf("codex/issue-%d", issueNumber)
	}
	trimmedRunID := strings.TrimSpace(runID)
	if len(trimmedRunID) > 12 {
		trimmedRunID = trimmedRunID[:12]
	}
	return "codex/run-" + trimmedRunID
}

func isAIRepairMainDirectTrigger(triggerKind string) bool {
	return webhookdomain.NormalizeTriggerKind(strings.TrimSpace(triggerKind)) == webhookdomain.TriggerKindAIRepair
}

func optionalIssueNumber(value int64) *int {
	if value <= 0 {
		return nil
	}
	intValue := int(value)
	return &intValue
}

func optionalInt(value int) *int {
	if value <= 0 {
		return nil
	}
	intValue := value
	return &intValue
}
