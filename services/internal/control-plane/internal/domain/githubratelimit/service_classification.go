package githubratelimit

import (
	"fmt"
	"strings"
	"time"

	waitrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/githubratelimitwait"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

var (
	hardFailureKeywords = []string{
		"bad credentials",
		"requires authentication",
		"resource not accessible by integration",
		"insufficient permission",
		"insufficient permissions",
		"missing scope",
		"must have admin rights",
		"not permitted",
		"forbidden for this token",
	}
	secondaryRateLimitKeywords = []string{
		"secondary rate limit",
		"secondary-rate-limits",
		"abuse detection",
		"abuse-rate-limits",
		"api rate limit exceeded",
		"rate limit error",
	}
)

func normalizeSignal(signal Signal, now time.Time) (Signal, error) {
	normalized := signal
	normalized.SignalID = strings.TrimSpace(normalized.SignalID)
	normalized.CorrelationID = strings.TrimSpace(normalized.CorrelationID)
	normalized.RequestFingerprint = strings.TrimSpace(normalized.RequestFingerprint)
	normalized.MessageExcerpt = trimToMaxBytes(strings.TrimSpace(normalized.MessageExcerpt), signalExcerptMaxBytes)
	normalized.StderrExcerpt = trimToMaxBytes(strings.TrimSpace(normalized.StderrExcerpt), signalExcerptMaxBytes)
	normalized.Headers.RateLimitResource = strings.TrimSpace(normalized.Headers.RateLimitResource)
	normalized.Headers.GitHubRequestID = strings.TrimSpace(normalized.Headers.GitHubRequestID)
	normalized.Headers.DocumentationURL = strings.TrimSpace(normalized.Headers.DocumentationURL)

	if normalized.SignalID == "" {
		return Signal{}, fmt.Errorf("signal_id is required")
	}
	if strings.TrimSpace(string(normalized.ContourKind)) == "" {
		return Signal{}, fmt.Errorf("contour_kind is required")
	}
	if strings.TrimSpace(string(normalized.SignalOrigin)) == "" {
		return Signal{}, fmt.Errorf("signal_origin is required")
	}
	if strings.TrimSpace(string(normalized.OperationClass)) == "" {
		return Signal{}, fmt.Errorf("operation_class is required")
	}
	if normalized.ProviderStatusCode != 403 && normalized.ProviderStatusCode != 429 {
		return Signal{}, fmt.Errorf("provider_status_code must be 403 or 429")
	}
	if normalized.OccurredAt.IsZero() {
		normalized.OccurredAt = now.UTC()
	} else {
		normalized.OccurredAt = normalized.OccurredAt.UTC()
	}
	if normalized.Headers.RateLimitResetAt != nil {
		resetAt := normalized.Headers.RateLimitResetAt.UTC()
		normalized.Headers.RateLimitResetAt = &resetAt
	}
	if normalized.Headers.RetryAfterSeconds != nil && *normalized.Headers.RetryAfterSeconds < 0 {
		return Signal{}, fmt.Errorf("retry_after_seconds must be >= 0")
	}
	return normalized, nil
}

func trimToMaxBytes(value string, max int) string {
	if len([]byte(value)) <= max {
		return value
	}

	trimmed := value
	for len([]byte(trimmed)) > max && len(trimmed) > 0 {
		trimmed = trimmed[:len(trimmed)-1]
	}
	return trimmed
}

func classifySignal(signal Signal, existing waitrepo.Wait, now time.Time) (Classification, error) {
	resumeActionKind, err := resolveResumeActionKind(signal.OperationClass)
	if err != nil {
		return Classification{}, err
	}

	attemptsUsed := existing.AutoResumeAttemptsUsed
	lowerEvidence := strings.ToLower(strings.Join([]string{
		signal.MessageExcerpt,
		signal.StderrExcerpt,
		signal.Headers.DocumentationURL,
	}, " "))
	if containsAnyKeyword(lowerEvidence, hardFailureKeywords) {
		return Classification{
			HardFailure:      true,
			FailureReason:    "non_rate_limit_forbidden",
			ResumeActionKind: resumeActionKind,
		}, nil
	}

	if signal.Headers.RetryAfterSeconds != nil {
		return classifySecondaryRetryAfter(*signal.Headers.RetryAfterSeconds, attemptsUsed, resumeActionKind, now), nil
	}
	if hasPrimaryResetSignal(signal.Headers) {
		return classifyPrimaryReset(*signal.Headers.RateLimitResetAt, attemptsUsed, resumeActionKind, now), nil
	}
	if signal.ProviderStatusCode == 429 || containsAnyKeyword(lowerEvidence, secondaryRateLimitKeywords) || strings.Contains(lowerEvidence, "rate limit") {
		return classifySecondaryBackoff(attemptsUsed, resumeActionKind, now), nil
	}

	return Classification{
		HardFailure:      true,
		FailureReason:    "non_rate_limit_forbidden",
		ResumeActionKind: resumeActionKind,
	}, nil
}

func classifyPrimaryReset(resetAt time.Time, attemptsUsed int, resumeActionKind enumtypes.GitHubRateLimitResumeActionKind, now time.Time) Classification {
	maxAttempts := 2
	resumeNotBefore := resetAt.UTC().Add(primaryLimitGuardDelay)
	if attemptsUsed > 0 {
		conservativeLowerBound := now.Add(primaryConservativeRetryDelay)
		if resumeNotBefore.Before(conservativeLowerBound) {
			resumeNotBefore = conservativeLowerBound
		}
	}

	classification := Classification{
		LimitKind:              enumtypes.GitHubRateLimitLimitKindPrimary,
		Confidence:             enumtypes.GitHubRateLimitConfidenceDeterministic,
		RecoveryHintKind:       enumtypes.GitHubRateLimitRecoveryHintKindReset,
		RecoveryHintSource:     enumtypes.GitHubRateLimitRecoveryHintSourceResetAt,
		State:                  enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
		NextStepKind:           enumtypes.GitHubRateLimitNextStepKindAutoResumeScheduled,
		ResumeActionKind:       resumeActionKind,
		ResumeNotBefore:        &resumeNotBefore,
		AutoResumeAttemptsUsed: attemptsUsed,
		MaxAutoResumeAttempts:  maxAttempts,
	}
	if attemptsUsed > 0 {
		classification.Confidence = enumtypes.GitHubRateLimitConfidenceConservative
	}
	if attemptsUsed >= maxAttempts {
		classification.State = enumtypes.GitHubRateLimitWaitStateManualActionRequired
		classification.NextStepKind = enumtypes.GitHubRateLimitNextStepKindManualActionRequired
		classification.RecoveryHintKind = enumtypes.GitHubRateLimitRecoveryHintKindManualOnly
		classification.ManualActionKind = resolveManualActionKind(resumeActionKind, classification.Confidence, classification.RecoveryHintKind)
	}
	return classification
}

func classifySecondaryRetryAfter(retryAfterSeconds int, attemptsUsed int, resumeActionKind enumtypes.GitHubRateLimitResumeActionKind, now time.Time) Classification {
	maxAttempts := 2
	resumeNotBefore := now.Add(time.Duration(retryAfterSeconds)*time.Second + secondaryRetryAfterGuardDelay)
	classification := Classification{
		LimitKind:              enumtypes.GitHubRateLimitLimitKindSecondary,
		Confidence:             enumtypes.GitHubRateLimitConfidenceConservative,
		RecoveryHintKind:       enumtypes.GitHubRateLimitRecoveryHintKindRetryAfter,
		RecoveryHintSource:     enumtypes.GitHubRateLimitRecoveryHintSourceRetryAfter,
		State:                  enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
		NextStepKind:           enumtypes.GitHubRateLimitNextStepKindAutoResumeScheduled,
		ResumeActionKind:       resumeActionKind,
		ResumeNotBefore:        &resumeNotBefore,
		AutoResumeAttemptsUsed: attemptsUsed,
		MaxAutoResumeAttempts:  maxAttempts,
	}
	if attemptsUsed >= maxAttempts {
		classification.State = enumtypes.GitHubRateLimitWaitStateManualActionRequired
		classification.NextStepKind = enumtypes.GitHubRateLimitNextStepKindManualActionRequired
		classification.RecoveryHintKind = enumtypes.GitHubRateLimitRecoveryHintKindManualOnly
		classification.ManualActionKind = resolveManualActionKind(resumeActionKind, classification.Confidence, classification.RecoveryHintKind)
	}
	return classification
}

func classifySecondaryBackoff(attemptsUsed int, resumeActionKind enumtypes.GitHubRateLimitResumeActionKind, now time.Time) Classification {
	maxAttempts := 3
	delay := secondaryBackoffDelay(attemptsUsed)
	resumeNotBefore := now.Add(delay)
	classification := Classification{
		LimitKind:              enumtypes.GitHubRateLimitLimitKindSecondary,
		Confidence:             enumtypes.GitHubRateLimitConfidenceProviderUnclear,
		RecoveryHintKind:       enumtypes.GitHubRateLimitRecoveryHintKindExponentialBackoff,
		RecoveryHintSource:     enumtypes.GitHubRateLimitRecoveryHintSourceProviderUncertain,
		State:                  enumtypes.GitHubRateLimitWaitStateAutoResumeScheduled,
		NextStepKind:           enumtypes.GitHubRateLimitNextStepKindAutoResumeScheduled,
		ResumeActionKind:       resumeActionKind,
		ResumeNotBefore:        &resumeNotBefore,
		AutoResumeAttemptsUsed: attemptsUsed,
		MaxAutoResumeAttempts:  maxAttempts,
	}
	if attemptsUsed >= maxAttempts {
		classification.State = enumtypes.GitHubRateLimitWaitStateManualActionRequired
		classification.NextStepKind = enumtypes.GitHubRateLimitNextStepKindManualActionRequired
		classification.RecoveryHintKind = enumtypes.GitHubRateLimitRecoveryHintKindManualOnly
		classification.ManualActionKind = resolveManualActionKind(resumeActionKind, classification.Confidence, classification.RecoveryHintKind)
	}
	return classification
}

func secondaryBackoffDelay(attemptsUsed int) time.Duration {
	multiplier := 1 << attemptsUsed
	delay := time.Duration(multiplier) * secondaryBackoffBaseDelay
	if delay > secondaryBackoffMaxDelay {
		return secondaryBackoffMaxDelay
	}
	return delay
}

func hasPrimaryResetSignal(headers Headers) bool {
	return headers.RateLimitRemaining != nil &&
		*headers.RateLimitRemaining == 0 &&
		headers.RateLimitResetAt != nil
}

func containsAnyKeyword(haystack string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(haystack, keyword) {
			return true
		}
	}
	return false
}

func resolveResumeActionKind(operationClass enumtypes.GitHubRateLimitOperationClass) (enumtypes.GitHubRateLimitResumeActionKind, error) {
	switch operationClass {
	case enumtypes.GitHubRateLimitOperationClassRunStatusComment:
		return enumtypes.GitHubRateLimitResumeActionKindRunStatusCommentRetry, nil
	case enumtypes.GitHubRateLimitOperationClassIssueLabelTransition, enumtypes.GitHubRateLimitOperationClassRepositoryProvider:
		return enumtypes.GitHubRateLimitResumeActionKindPlatformCallReplay, nil
	case enumtypes.GitHubRateLimitOperationClassAgentGitHubCall:
		return enumtypes.GitHubRateLimitResumeActionKindAgentSessionResume, nil
	default:
		return "", fmt.Errorf("unsupported github rate-limit operation_class %q", operationClass)
	}
}

func resolveManualActionKind(resumeActionKind enumtypes.GitHubRateLimitResumeActionKind, confidence enumtypes.GitHubRateLimitConfidence, recoveryHintKind enumtypes.GitHubRateLimitRecoveryHintKind) enumtypes.GitHubRateLimitManualActionKind {
	if resumeActionKind == enumtypes.GitHubRateLimitResumeActionKindAgentSessionResume {
		return enumtypes.GitHubRateLimitManualActionKindResumeAgentSession
	}
	if confidence == enumtypes.GitHubRateLimitConfidenceProviderUnclear {
		return enumtypes.GitHubRateLimitManualActionKindRetryAfterReview
	}
	return enumtypes.GitHubRateLimitManualActionKindRequeuePlatformOperation
}

func nextStepKindForWaitState(state enumtypes.GitHubRateLimitWaitState) enumtypes.GitHubRateLimitNextStepKind {
	if state == enumtypes.GitHubRateLimitWaitStateManualActionRequired {
		return enumtypes.GitHubRateLimitNextStepKindManualActionRequired
	}
	return enumtypes.GitHubRateLimitNextStepKindAutoResumeScheduled
}

func classificationFromWait(wait Wait) Classification {
	return Classification{
		LimitKind:              wait.LimitKind,
		Confidence:             wait.Confidence,
		RecoveryHintKind:       wait.RecoveryHintKind,
		RecoveryHintSource:     recoveryHintSourceForWait(wait),
		State:                  wait.State,
		NextStepKind:           nextStepKindForWaitState(wait.State),
		ResumeActionKind:       wait.ResumeActionKind,
		ManualActionKind:       wait.ManualActionKind,
		ResumeNotBefore:        wait.ResumeNotBefore,
		AutoResumeAttemptsUsed: wait.AutoResumeAttemptsUsed,
		MaxAutoResumeAttempts:  wait.MaxAutoResumeAttempts,
	}
}
