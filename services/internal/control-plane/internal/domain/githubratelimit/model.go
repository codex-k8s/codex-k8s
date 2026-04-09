package githubratelimit

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	floweventdomain "github.com/codex-k8s/kodex/libs/go/domain/flowevent"
	agentrunrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/flowevent"
	waitrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/githubratelimitwait"
	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

const (
	primaryLimitGuardDelay              = 5 * time.Second
	secondaryRetryAfterGuardDelay       = 15 * time.Second
	primaryConservativeRetryDelay       = time.Minute
	secondaryBackoffBaseDelay           = time.Minute
	secondaryBackoffMaxDelay            = 15 * time.Minute
	signalExcerptMaxBytes               = 4 * 1024
	rateLimitResumePayloadMaxBytes      = 12 * 1024
	waitStateWaitingBackpressure        = "waiting_backpressure"
	runStatusPending                    = "pending"
	runStatusRunning                    = "running"
	runStatusWaitingBackpressure        = "waiting_backpressure"
	runStatusSucceeded                  = "succeeded"
	runStatusFailed                     = "failed"
	runStatusCanceled                   = "canceled"
	githubRateLimitWaitEnteredEventKey  = "github_rate_limit.wait.entered"
	githubRateLimitManualActionEventKey = "github_rate_limit.manual_action_required"
)

// Config controls rollout gates for GitHub rate-limit ownership inside control-plane.
type Config struct {
	RolloutState valuetypes.GitHubRateLimitRolloutState
}

// Dependencies contains collaborators required by the domain service.
type Dependencies struct {
	Runs           runRepository
	Waits          waitrepo.Repository
	FlowEvents     floweventrepo.Repository
	RunStatusRetry runStatusCommentRetrier
	PlatformReplay platformCallReplayer
	RolloutState   rolloutStateProvider
}

// ReportSignalParams carries one canonical provider signal into control-plane.
type ReportSignalParams struct {
	RunID             string
	Signal            Signal
	ReplayPayloadJSON json.RawMessage
}

// ReportSignalResult returns persisted wait state, projection, and comment context for one signal.
type ReportSignalResult struct {
	HardFailure          bool
	DuplicateSignal      bool
	Wait                 Wait
	Classification       Classification
	Projection           WaitProjection
	CommentRenderContext CommentRenderContext
	ProjectionRefresh    waitrepo.RefreshProjectionResult
}

// BuildResumePayloadParams describes deterministic agent-session resume JSON emitted after recovery.
type BuildResumePayloadParams struct {
	Wait           Wait
	ResolutionKind enumtypes.GitHubRateLimitResolutionKind
	RecoveredAt    time.Time
	AttemptNo      int
}

// ProcessNextAutoResumeParams describes one worker-triggered sweep attempt.
type ProcessNextAutoResumeParams struct {
	WorkerID string
}

// ProcessNextAutoResumeResult reports worker sweep outcome for one due wait.
type ProcessNextAutoResumeResult struct {
	Found                 bool
	Wait                  Wait
	ResolutionKind        enumtypes.GitHubRateLimitResolutionKind
	AttemptNo             int
	ManualActionKind      enumtypes.GitHubRateLimitManualActionKind
	ResumeNotBefore       *time.Time
	RequeuedCorrelationID string
}

// Service implements canonical GitHub rate-limit domain ownership under control-plane.
type Service struct {
	cfg        Config
	runs       runRepository
	waits      waitrepo.Repository
	flowEvents floweventrepo.Repository
	runStatus  runStatusCommentRetrier
	platform   platformCallReplayer
	rollout    rolloutStateProvider
	now        func() time.Time
}

type runRepository interface {
	GetByID(ctx context.Context, runID string) (agentrunrepo.Run, bool, error)
	CreatePendingIfAbsent(ctx context.Context, params agentrunrepo.CreateParams) (agentrunrepo.CreateResult, error)
}

type runStatusCommentRetrier interface {
	RetryGitHubRateLimitComment(ctx context.Context, payload valuetypes.GitHubRateLimitRunStatusCommentRetryPayload) error
}

type platformCallReplayer interface {
	ReplayGitHubRateLimitPlatformCall(ctx context.Context, payload valuetypes.GitHubRateLimitPlatformCallReplayPayload) error
}

type rolloutStateProvider interface {
	CurrentGitHubRateLimitRolloutState() valuetypes.GitHubRateLimitRolloutState
}

type waitSignalEvidencePayload struct {
	ContourKind            enumtypes.GitHubRateLimitContourKind    `json:"contour_kind"`
	OperationClass         enumtypes.GitHubRateLimitOperationClass `json:"operation_class"`
	SessionSnapshotVersion *int64                                  `json:"session_snapshot_version,omitempty"`
}

type waitClassificationEvidencePayload struct {
	LimitKind           enumtypes.GitHubRateLimitLimitKind           `json:"limit_kind"`
	Confidence          enumtypes.GitHubRateLimitConfidence          `json:"confidence"`
	RecoveryHintKind    enumtypes.GitHubRateLimitRecoveryHintKind    `json:"recovery_hint_kind"`
	RecoveryHintSource  enumtypes.GitHubRateLimitRecoveryHintSource  `json:"recovery_hint_source"`
	NextStepKind        enumtypes.GitHubRateLimitNextStepKind        `json:"next_step_kind"`
	ResumeActionKind    enumtypes.GitHubRateLimitResumeActionKind    `json:"resume_action_kind"`
	ManualActionKind    enumtypes.GitHubRateLimitManualActionKind    `json:"manual_action_kind,omitempty"`
	ProjectionSyncState enumtypes.GitHubRateLimitProjectionSyncState `json:"projection_sync_state"`
}

type waitResumeScheduledEvidencePayload struct {
	WaitID          string                                `json:"wait_id"`
	ResumeNotBefore *time.Time                            `json:"resume_not_before,omitempty"`
	AttemptsUsed    int                                   `json:"attempts_used"`
	MaxAttempts     int                                   `json:"max_attempts"`
	NextStepKind    enumtypes.GitHubRateLimitNextStepKind `json:"next_step_kind"`
	SignalOrigin    enumtypes.GitHubRateLimitSignalOrigin `json:"signal_origin"`
}

type waitResumeAttemptEvidencePayload struct {
	WaitID    string `json:"wait_id"`
	AttemptNo int    `json:"attempt_no"`
	WorkerID  string `json:"worker_id,omitempty"`
}

type waitResumeFailureEvidencePayload struct {
	WaitID          string                                `json:"wait_id"`
	AttemptNo       int                                   `json:"attempt_no"`
	WorkerID        string                                `json:"worker_id,omitempty"`
	Error           string                                `json:"error,omitempty"`
	NextStepKind    enumtypes.GitHubRateLimitNextStepKind `json:"next_step_kind"`
	ResumeNotBefore *time.Time                            `json:"resume_not_before,omitempty"`
}

type waitResolvedEvidencePayload struct {
	WaitID                string                                  `json:"wait_id"`
	AttemptNo             int                                     `json:"attempt_no"`
	ResolutionKind        enumtypes.GitHubRateLimitResolutionKind `json:"resolution_kind"`
	RequeuedCorrelationID string                                  `json:"requeued_correlation_id,omitempty"`
}

type waitManualActionEvidencePayload struct {
	WaitID           string                                    `json:"wait_id"`
	AttemptNo        int                                       `json:"attempt_no"`
	ManualActionKind enumtypes.GitHubRateLimitManualActionKind `json:"manual_action_kind"`
	WorkerID         string                                    `json:"worker_id,omitempty"`
}

type runWaitFlowEventPayload struct {
	RunID              string                                      `json:"run_id"`
	WaitID             string                                      `json:"wait_id"`
	ContourKind        enumtypes.GitHubRateLimitContourKind        `json:"contour_kind"`
	LimitKind          enumtypes.GitHubRateLimitLimitKind          `json:"limit_kind"`
	State              enumtypes.GitHubRateLimitWaitState          `json:"state"`
	NextStepKind       enumtypes.GitHubRateLimitNextStepKind       `json:"next_step_kind"`
	ResumeNotBefore    *time.Time                                  `json:"resume_not_before,omitempty"`
	OpenWaitCount      int                                         `json:"open_wait_count"`
	DominantWaitID     string                                      `json:"dominant_wait_id,omitempty"`
	EventKey           string                                      `json:"event_key"`
	CommentMirrorState enumtypes.GitHubRateLimitCommentMirrorState `json:"comment_mirror_state"`
}

type runWaitResolvedFlowEventPayload struct {
	RunID                 string                                      `json:"run_id"`
	WaitID                string                                      `json:"wait_id"`
	ContourKind           enumtypes.GitHubRateLimitContourKind        `json:"contour_kind"`
	LimitKind             enumtypes.GitHubRateLimitLimitKind          `json:"limit_kind"`
	OperationClass        enumtypes.GitHubRateLimitOperationClass     `json:"operation_class"`
	ResolutionKind        enumtypes.GitHubRateLimitResolutionKind     `json:"resolution_kind"`
	AttemptNo             int                                         `json:"attempt_no"`
	EventKey              string                                      `json:"event_key"`
	RequeuedCorrelationID string                                      `json:"requeued_correlation_id,omitempty"`
	CommentMirrorState    enumtypes.GitHubRateLimitCommentMirrorState `json:"comment_mirror_state,omitempty"`
}

type marshalErrorPayload struct {
	Error string `json:"error"`
}

var errGitHubRateLimitReplayPayloadMissing = errors.New("github rate-limit replay payload is required")

type messageTemplateData struct {
	ContourKind               string
	LimitKind                 string
	OperationClass            string
	WaitState                 string
	Confidence                string
	RecoveryHintKind          string
	NextStepKind              string
	ResumeNotBeforeRFC3339    string
	SuggestedNotBeforeRFC3339 string
	AttemptsUsed              int
	MaxAttempts               int
	AttemptNo                 int
}

func marshalJSONPayload(payload any) json.RawMessage {
	raw, err := json.Marshal(payload)
	if err == nil {
		return raw
	}
	fallback, fallbackErr := json.Marshal(marshalErrorPayload{Error: "payload_marshal_failed"})
	if fallbackErr != nil {
		return json.RawMessage(`{"error":"payload_marshal_failed"}`)
	}
	return fallback
}

func waitPausedEventType() floweventdomain.EventType {
	return floweventdomain.EventTypeRunWaitPaused
}
