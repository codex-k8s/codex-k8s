package entity

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// GitHubRateLimitWait is canonical persisted aggregate for one recoverable GitHub rate-limit incident.
type GitHubRateLimitWait struct {
	ID                     string
	ProjectID              string
	RunID                  string
	ContourKind            enumtypes.GitHubRateLimitContourKind
	SignalOrigin           enumtypes.GitHubRateLimitSignalOrigin
	OperationClass         enumtypes.GitHubRateLimitOperationClass
	State                  enumtypes.GitHubRateLimitWaitState
	LimitKind              enumtypes.GitHubRateLimitLimitKind
	Confidence             enumtypes.GitHubRateLimitConfidence
	RecoveryHintKind       enumtypes.GitHubRateLimitRecoveryHintKind
	DominantForRun         bool
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
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// GitHubRateLimitWaitEvidence is one append-only evidence row for the wait aggregate.
type GitHubRateLimitWaitEvidence struct {
	ID                 int64
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
	CreatedAt          time.Time
}
