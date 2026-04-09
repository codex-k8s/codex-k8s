package value

import (
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// GitHubRateLimitRunWaitProjection is the canonical typed read model for one GitHub rate-limit wait.
type GitHubRateLimitRunWaitProjection struct {
	WaitState          string                                      `json:"wait_state"`
	WaitReason         enumtypes.AgentRunWaitReason                `json:"wait_reason"`
	DominantWait       GitHubRateLimitWaitProjectionItem           `json:"dominant_wait"`
	RelatedWaits       []GitHubRateLimitWaitProjectionItem         `json:"related_waits"`
	CommentMirrorState enumtypes.GitHubRateLimitCommentMirrorState `json:"comment_mirror_state"`
}

// GitHubRateLimitWaitProjectionItem is one visible wait aggregate in run details and wait queue.
type GitHubRateLimitWaitProjectionItem struct {
	WaitID          string                                  `json:"wait_id"`
	ContourKind     enumtypes.GitHubRateLimitContourKind    `json:"contour_kind"`
	LimitKind       enumtypes.GitHubRateLimitLimitKind      `json:"limit_kind"`
	OperationClass  enumtypes.GitHubRateLimitOperationClass `json:"operation_class"`
	State           enumtypes.GitHubRateLimitWaitState      `json:"state"`
	Confidence      enumtypes.GitHubRateLimitConfidence     `json:"confidence"`
	EnteredAt       time.Time                               `json:"entered_at"`
	ResumeNotBefore *time.Time                              `json:"resume_not_before,omitempty"`
	AttemptsUsed    int                                     `json:"attempts_used"`
	MaxAttempts     int                                     `json:"max_attempts"`
	RecoveryHint    GitHubRateLimitRecoveryHint             `json:"recovery_hint"`
	ManualAction    *GitHubRateLimitManualAction            `json:"manual_action,omitempty"`
}

// GitHubRateLimitRecoveryHint is the typed operator-facing retry hint derived from provider evidence.
type GitHubRateLimitRecoveryHint struct {
	HintKind        enumtypes.GitHubRateLimitRecoveryHintKind   `json:"hint_kind"`
	ResumeNotBefore *time.Time                                  `json:"resume_not_before,omitempty"`
	SourceHeaders   enumtypes.GitHubRateLimitRecoveryHintSource `json:"source_headers"`
	DetailsMarkdown string                                      `json:"details_markdown"`
}

// GitHubRateLimitManualAction is the typed operator guidance emitted after auto-resume budget exhaustion.
type GitHubRateLimitManualAction struct {
	Kind               enumtypes.GitHubRateLimitManualActionKind `json:"kind"`
	Summary            string                                    `json:"summary"`
	DetailsMarkdown    string                                    `json:"details_markdown"`
	SuggestedNotBefore *time.Time                                `json:"suggested_not_before,omitempty"`
}

// GitHubRateLimitCommentRenderContext is the canonical source for best-effort GitHub service-comment mirroring.
type GitHubRateLimitCommentRenderContext struct {
	Headline             string                                  `json:"headline"`
	DominantContour      enumtypes.GitHubRateLimitContourKind    `json:"dominant_contour"`
	LimitKind            enumtypes.GitHubRateLimitLimitKind      `json:"limit_kind"`
	OperationClass       enumtypes.GitHubRateLimitOperationClass `json:"operation_class"`
	NextStepKind         enumtypes.GitHubRateLimitNextStepKind   `json:"next_step_kind"`
	ResumeNotBefore      *time.Time                              `json:"resume_not_before,omitempty"`
	ManualActionSummary  string                                  `json:"manual_action_summary,omitempty"`
	RelatedContourBadges []GitHubRateLimitCommentContourBadge    `json:"related_contour_badges"`
}

// GitHubRateLimitCommentContourBadge is the compact related-wait badge rendered into the service comment.
type GitHubRateLimitCommentContourBadge struct {
	ContourKind enumtypes.GitHubRateLimitContourKind `json:"contour_kind"`
	LimitKind   enumtypes.GitHubRateLimitLimitKind   `json:"limit_kind"`
	State       enumtypes.GitHubRateLimitWaitState   `json:"state"`
}
