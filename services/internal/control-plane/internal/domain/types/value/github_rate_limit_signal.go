package value

import (
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// GitHubRateLimitHeaders is a sanitized snapshot of provider headers attached to one signal.
type GitHubRateLimitHeaders struct {
	RateLimitLimit     *int       `json:"rate_limit_limit,omitempty"`
	RateLimitRemaining *int       `json:"rate_limit_remaining,omitempty"`
	RateLimitUsed      *int       `json:"rate_limit_used,omitempty"`
	RateLimitResetAt   *time.Time `json:"rate_limit_reset_at,omitempty"`
	RateLimitResource  string     `json:"rate_limit_resource,omitempty"`
	RetryAfterSeconds  *int       `json:"retry_after_seconds,omitempty"`
	GitHubRequestID    string     `json:"github_request_id,omitempty"`
	DocumentationURL   string     `json:"documentation_url,omitempty"`
}

// GitHubRateLimitSignal is the canonical raw provider evidence sent into the domain owner.
type GitHubRateLimitSignal struct {
	SignalID               string                                  `json:"signal_id"`
	CorrelationID          string                                  `json:"correlation_id,omitempty"`
	ContourKind            enumtypes.GitHubRateLimitContourKind    `json:"contour_kind"`
	SignalOrigin           enumtypes.GitHubRateLimitSignalOrigin   `json:"signal_origin"`
	OperationClass         enumtypes.GitHubRateLimitOperationClass `json:"operation_class"`
	ProviderStatusCode     int                                     `json:"provider_status_code"`
	OccurredAt             time.Time                               `json:"occurred_at"`
	RequestFingerprint     string                                  `json:"request_fingerprint,omitempty"`
	StderrExcerpt          string                                  `json:"stderr_excerpt,omitempty"`
	MessageExcerpt         string                                  `json:"message_excerpt,omitempty"`
	Headers                GitHubRateLimitHeaders                  `json:"github_headers,omitempty"`
	SessionSnapshotVersion *int64                                  `json:"session_snapshot_version,omitempty"`
}

// GitHubRateLimitClassification is domain output of signal normalization and retry policy election.
type GitHubRateLimitClassification struct {
	HardFailure            bool
	FailureReason          string
	LimitKind              enumtypes.GitHubRateLimitLimitKind
	Confidence             enumtypes.GitHubRateLimitConfidence
	RecoveryHintKind       enumtypes.GitHubRateLimitRecoveryHintKind
	RecoveryHintSource     enumtypes.GitHubRateLimitRecoveryHintSource
	State                  enumtypes.GitHubRateLimitWaitState
	NextStepKind           enumtypes.GitHubRateLimitNextStepKind
	ResumeActionKind       enumtypes.GitHubRateLimitResumeActionKind
	ManualActionKind       enumtypes.GitHubRateLimitManualActionKind
	ResumeNotBefore        *time.Time
	AutoResumeAttemptsUsed int
	MaxAutoResumeAttempts  int
}
