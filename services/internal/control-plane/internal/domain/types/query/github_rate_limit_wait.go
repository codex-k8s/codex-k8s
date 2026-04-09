package query

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// GitHubRateLimitWaitCreateParams creates one persisted wait aggregate.
type GitHubRateLimitWaitCreateParams struct {
	ProjectID              string
	RunID                  string
	ContourKind            enumtypes.GitHubRateLimitContourKind
	SignalOrigin           enumtypes.GitHubRateLimitSignalOrigin
	OperationClass         enumtypes.GitHubRateLimitOperationClass
	State                  enumtypes.GitHubRateLimitWaitState
	LimitKind              enumtypes.GitHubRateLimitLimitKind
	Confidence             enumtypes.GitHubRateLimitConfidence
	RecoveryHintKind       enumtypes.GitHubRateLimitRecoveryHintKind
	SignalID               string
	RequestFingerprint     string
	CorrelationID          string
	ResumeActionKind       enumtypes.GitHubRateLimitResumeActionKind
	ResumePayloadJSON      json.RawMessage
	ManualActionKind       enumtypes.GitHubRateLimitManualActionKind
	AutoResumeAttemptsUsed int
	MaxAutoResumeAttempts  int
	ResumeNotBefore        *time.Time
	LastResumeAttemptAt    *time.Time
	FirstDetectedAt        time.Time
	LastSignalAt           time.Time
	ResolvedAt             *time.Time
}

// GitHubRateLimitWaitUpdateParams updates mutable fields of one existing wait aggregate.
type GitHubRateLimitWaitUpdateParams struct {
	ID                     string
	SignalOrigin           enumtypes.GitHubRateLimitSignalOrigin
	OperationClass         enumtypes.GitHubRateLimitOperationClass
	State                  enumtypes.GitHubRateLimitWaitState
	LimitKind              enumtypes.GitHubRateLimitLimitKind
	Confidence             enumtypes.GitHubRateLimitConfidence
	RecoveryHintKind       enumtypes.GitHubRateLimitRecoveryHintKind
	SignalID               string
	RequestFingerprint     string
	CorrelationID          string
	ResumeActionKind       enumtypes.GitHubRateLimitResumeActionKind
	ResumePayloadJSON      json.RawMessage
	ManualActionKind       enumtypes.GitHubRateLimitManualActionKind
	AutoResumeAttemptsUsed int
	MaxAutoResumeAttempts  int
	ResumeNotBefore        *time.Time
	LastResumeAttemptAt    *time.Time
	LastSignalAt           time.Time
	ResolvedAt             *time.Time
}

// GitHubRateLimitWaitEvidenceCreateParams appends one evidence ledger row.
type GitHubRateLimitWaitEvidenceCreateParams struct {
	WaitID             string
	EventKind          enumtypes.GitHubRateLimitEvidenceEventKind
	SignalID           string
	SignalOrigin       enumtypes.GitHubRateLimitSignalOrigin
	ProviderStatusCode *int
	RetryAfterSeconds  *int
	RateLimitLimit     *int
	RateLimitRemaining *int
	RateLimitUsed      *int
	RateLimitResetAt   *time.Time
	RateLimitResource  string
	GitHubRequestID    string
	DocumentationURL   string
	MessageExcerpt     string
	StderrExcerpt      string
	PayloadJSON        json.RawMessage
	ObservedAt         time.Time
}
