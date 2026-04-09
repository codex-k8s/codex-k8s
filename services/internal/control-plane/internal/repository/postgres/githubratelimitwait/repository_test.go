package githubratelimitwait

import (
	"strings"
	"testing"
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

func TestSetRunWaitContextQueryGuardsExistingNonGitHubWaits(t *testing.T) {
	t.Parallel()

	required := []string{
		"AND (wait_reason IS NULL OR wait_reason = '' OR wait_reason = $2)",
		"AND (wait_target_kind IS NULL OR wait_target_kind = '' OR wait_target_kind = $3)",
	}
	for _, item := range required {
		if !strings.Contains(querySetRunWaitContext, item) {
			t.Fatalf("set_run_wait_context query must contain %q", item)
		}
	}
}

func TestSetSessionBackpressureQueryGuardsExistingNonBackpressureWaits(t *testing.T) {
	t.Parallel()

	if !strings.Contains(querySetSessionBackpressure, "AND (wait_state IS NULL OR wait_state = '' OR wait_state = $2)") {
		t.Fatal("set_session_backpressure query must guard existing non-backpressure wait_state")
	}
}

func TestValidateWaitLifecycleRejectsResolvedWithoutResolvedAt(t *testing.T) {
	t.Parallel()

	input := validWaitLifecycleValidationInput()
	input.State = enumtypes.GitHubRateLimitWaitStateResolved
	input.ResolvedAt = nil

	err := validateWaitLifecycle(input)
	if err == nil || !strings.Contains(err.Error(), "resolved_at is required") {
		t.Fatalf("expected resolved_at validation error, got %v", err)
	}
}

func TestValidateWaitLifecycleRejectsManualActionWithoutKind(t *testing.T) {
	t.Parallel()

	input := validWaitLifecycleValidationInput()
	input.State = enumtypes.GitHubRateLimitWaitStateManualActionRequired
	input.ManualActionKind = ""
	input.AutoResumeAttemptsUsed = input.MaxAutoResumeAttempts

	err := validateWaitLifecycle(input)
	if err == nil || !strings.Contains(err.Error(), "manual_action_kind is required") {
		t.Fatalf("expected manual_action_kind validation error, got %v", err)
	}
}

func TestValidateWaitLifecycleRejectsManualActionWithoutExhaustedBudget(t *testing.T) {
	t.Parallel()

	input := validWaitLifecycleValidationInput()
	input.State = enumtypes.GitHubRateLimitWaitStateManualActionRequired
	input.ManualActionKind = enumtypes.GitHubRateLimitManualActionKindRetryAfterReview

	err := validateWaitLifecycle(input)
	if err == nil || !strings.Contains(err.Error(), "exhausted auto-resume budget") {
		t.Fatalf("expected exhausted-budget validation error, got %v", err)
	}
}

func TestValidateWaitLifecycleRejectsResumeNotBeforeOutsideManualOnlyProviderUncertainPath(t *testing.T) {
	t.Parallel()

	input := validWaitLifecycleValidationInput()
	input.ResumeNotBefore = nil

	err := validateWaitLifecycle(input)
	if err == nil || !strings.Contains(err.Error(), "resume_not_before is required") {
		t.Fatalf("expected resume_not_before validation error, got %v", err)
	}
}

func TestValidateWaitLifecycleAllowsExhaustedManualOnlyProviderUncertainPathWithoutResumeNotBefore(t *testing.T) {
	t.Parallel()

	input := validWaitLifecycleValidationInput()
	input.RecoveryHintKind = enumtypes.GitHubRateLimitRecoveryHintKindManualOnly
	input.Confidence = enumtypes.GitHubRateLimitConfidenceProviderUnclear
	input.ResumeNotBefore = nil
	input.AutoResumeAttemptsUsed = input.MaxAutoResumeAttempts

	if err := validateWaitLifecycle(input); err != nil {
		t.Fatalf("validateWaitLifecycle() error = %v", err)
	}
}

func TestValidateWaitLifecycleRejectsAttemptsUsedAboveBudget(t *testing.T) {
	t.Parallel()

	input := validWaitLifecycleValidationInput()
	input.AutoResumeAttemptsUsed = input.MaxAutoResumeAttempts + 1

	err := validateWaitLifecycle(input)
	if err == nil || !strings.Contains(err.Error(), "must be <= max_auto_resume_attempts") {
		t.Fatalf("expected budget validation error, got %v", err)
	}
}

func validWaitLifecycleValidationInput() waitLifecycleValidationInput {
	resumeNotBefore := time.Date(2026, time.March, 14, 12, 0, 0, 0, time.UTC)
	return waitLifecycleValidationInput{
		State:                  enumtypes.GitHubRateLimitWaitStateOpen,
		Confidence:             enumtypes.GitHubRateLimitConfidenceDeterministic,
		RecoveryHintKind:       enumtypes.GitHubRateLimitRecoveryHintKindRetryAfter,
		AutoResumeAttemptsUsed: 1,
		MaxAutoResumeAttempts:  2,
		ResumeNotBefore:        &resumeNotBefore,
	}
}
