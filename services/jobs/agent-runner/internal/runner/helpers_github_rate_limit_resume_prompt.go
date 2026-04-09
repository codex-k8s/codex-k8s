package runner

import (
	"fmt"
	"time"

	sharedgithubratelimit "github.com/codex-k8s/kodex/libs/go/domain/githubratelimit"
)

const (
	githubRateLimitWaitReason = "github_rate_limit"

	githubRateLimitContourPlatformPAT   = "platform_pat"
	githubRateLimitContourAgentBotToken = "agent_bot_token"

	githubRateLimitLimitPrimary   = "primary"
	githubRateLimitLimitSecondary = "secondary"

	githubRateLimitResolutionAutoResumed      = "auto_resumed"
	githubRateLimitResolutionManuallyResolved = "manually_resolved"
	githubRateLimitResolutionCancelled        = "cancelled"

	githubRateLimitOperationRunStatusComment   = "run_status_comment"
	githubRateLimitOperationIssueLabel         = "issue_label_transition"
	githubRateLimitOperationRepositoryProvider = "repository_provider_call"
	githubRateLimitOperationAgentGitHubCall    = "agent_github_call"
)

type githubRateLimitResumePayload struct {
	WaitID                 string `json:"wait_id"`
	WaitReason             string `json:"wait_reason"`
	ContourKind            string `json:"contour_kind"`
	LimitKind              string `json:"limit_kind"`
	ResolutionKind         string `json:"resolution_kind"`
	RecoveredAt            string `json:"recovered_at"`
	AttemptNo              int    `json:"attempt_no"`
	AffectedOperationClass string `json:"affected_operation_class"`
	Guidance               string `json:"guidance"`
}

func buildGitHubRateLimitResumePromptBlock(locale string, rawPayload string, resume bool) (string, error) {
	return buildDeterministicResumePromptBlock(
		locale,
		rawPayload,
		resume,
		"github rate-limit resume payload requires restored codex session",
		parseGitHubRateLimitResumePayload,
		"Детерминированный resume context (GitHub rate-limit wait):",
		"Ниже machine-readable outcome platform-owned GitHub rate-limit wait. Используйте этот JSON как единственный authoritative source для resume semantics, не выводите wait-state заново из старых stderr/headers и продолжайте работу из восстановленного snapshot.",
		"Deterministic resume context (GitHub rate-limit wait):",
		"Below is the machine-readable outcome of a platform-owned GitHub rate-limit wait. Treat this JSON as the only authoritative source for resume semantics, do not re-derive wait state from stale stderr/headers, and continue from the restored snapshot.",
	)
}

func parseGitHubRateLimitResumePayload(rawPayload string) (githubRateLimitResumePayload, error) {
	return parseDeterministicResumePayload(
		rawPayload,
		sharedgithubratelimit.ResumePayloadMaxBytes,
		"github rate-limit resume payload",
		normalizeGitHubRateLimitResumePayload,
		validateGitHubRateLimitResumePayload,
	)
}

func normalizeGitHubRateLimitResumePayload(payload githubRateLimitResumePayload) githubRateLimitResumePayload {
	trimStringFields(
		&payload.WaitID,
		&payload.WaitReason,
		&payload.ContourKind,
		&payload.LimitKind,
		&payload.ResolutionKind,
		&payload.RecoveredAt,
		&payload.AffectedOperationClass,
		&payload.Guidance,
	)
	return payload
}

func validateGitHubRateLimitResumePayload(payload githubRateLimitResumePayload) error {
	if payload.WaitID == "" {
		return fmt.Errorf("github rate-limit resume payload: wait_id is required")
	}
	if payload.WaitReason != githubRateLimitWaitReason {
		return fmt.Errorf("github rate-limit resume payload: wait_reason must be %q", githubRateLimitWaitReason)
	}
	switch payload.ContourKind {
	case githubRateLimitContourPlatformPAT, githubRateLimitContourAgentBotToken:
	default:
		return fmt.Errorf("github rate-limit resume payload: contour_kind %q is not supported", payload.ContourKind)
	}
	switch payload.LimitKind {
	case githubRateLimitLimitPrimary, githubRateLimitLimitSecondary:
	default:
		return fmt.Errorf("github rate-limit resume payload: limit_kind %q is not supported", payload.LimitKind)
	}
	switch payload.ResolutionKind {
	case githubRateLimitResolutionAutoResumed, githubRateLimitResolutionManuallyResolved, githubRateLimitResolutionCancelled:
	default:
		return fmt.Errorf("github rate-limit resume payload: resolution_kind %q is not supported", payload.ResolutionKind)
	}
	if payload.RecoveredAt == "" {
		return fmt.Errorf("github rate-limit resume payload: recovered_at is required")
	}
	if _, err := time.Parse(time.RFC3339Nano, payload.RecoveredAt); err != nil {
		return fmt.Errorf("github rate-limit resume payload: recovered_at must be RFC3339: %w", err)
	}
	if payload.AttemptNo <= 0 {
		return fmt.Errorf("github rate-limit resume payload: attempt_no must be > 0")
	}
	switch payload.AffectedOperationClass {
	case githubRateLimitOperationRunStatusComment,
		githubRateLimitOperationIssueLabel,
		githubRateLimitOperationRepositoryProvider,
		githubRateLimitOperationAgentGitHubCall:
	default:
		return fmt.Errorf("github rate-limit resume payload: affected_operation_class %q is not supported", payload.AffectedOperationClass)
	}
	if payload.Guidance == "" {
		return fmt.Errorf("github rate-limit resume payload: guidance is required")
	}
	return nil
}
