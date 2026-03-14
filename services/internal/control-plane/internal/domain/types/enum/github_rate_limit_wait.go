package enum

// GitHubRateLimitContourKind identifies which GitHub auth contour is saturated.
type GitHubRateLimitContourKind string

const (
	GitHubRateLimitContourKindPlatformPAT   GitHubRateLimitContourKind = "platform_pat"
	GitHubRateLimitContourKindAgentBotToken GitHubRateLimitContourKind = "agent_bot_token"
)

// GitHubRateLimitSignalOrigin identifies which service emitted the current signal.
type GitHubRateLimitSignalOrigin string

const (
	GitHubRateLimitSignalOriginControlPlane GitHubRateLimitSignalOrigin = "control_plane"
	GitHubRateLimitSignalOriginWorker       GitHubRateLimitSignalOrigin = "worker"
	GitHubRateLimitSignalOriginAgentRunner  GitHubRateLimitSignalOrigin = "agent_runner"
)

// GitHubRateLimitOperationClass classifies the blocked GitHub operation family.
type GitHubRateLimitOperationClass string

const (
	GitHubRateLimitOperationClassRunStatusComment     GitHubRateLimitOperationClass = "run_status_comment"
	GitHubRateLimitOperationClassIssueLabelTransition GitHubRateLimitOperationClass = "issue_label_transition"
	GitHubRateLimitOperationClassRepositoryProvider   GitHubRateLimitOperationClass = "repository_provider_call"
	GitHubRateLimitOperationClassAgentGitHubCall      GitHubRateLimitOperationClass = "agent_github_call"
)

// GitHubRateLimitWaitState is lifecycle state of one persisted rate-limit wait aggregate.
type GitHubRateLimitWaitState string

const (
	GitHubRateLimitWaitStateOpen                 GitHubRateLimitWaitState = "open"
	GitHubRateLimitWaitStateAutoResumeScheduled  GitHubRateLimitWaitState = "auto_resume_scheduled"
	GitHubRateLimitWaitStateAutoResumeInProgress GitHubRateLimitWaitState = "auto_resume_in_progress"
	GitHubRateLimitWaitStateResolved             GitHubRateLimitWaitState = "resolved"
	GitHubRateLimitWaitStateManualActionRequired GitHubRateLimitWaitState = "manual_action_required"
	GitHubRateLimitWaitStateCancelled            GitHubRateLimitWaitState = "cancelled"
)

// IsOpen reports whether the wait state still participates in dominant-wait election.
func (s GitHubRateLimitWaitState) IsOpen() bool {
	switch s {
	case GitHubRateLimitWaitStateOpen,
		GitHubRateLimitWaitStateAutoResumeScheduled,
		GitHubRateLimitWaitStateAutoResumeInProgress,
		GitHubRateLimitWaitStateManualActionRequired:
		return true
	default:
		return false
	}
}

// GitHubRateLimitLimitKind is provider classification for the detected limit.
type GitHubRateLimitLimitKind string

const (
	GitHubRateLimitLimitKindPrimary   GitHubRateLimitLimitKind = "primary"
	GitHubRateLimitLimitKindSecondary GitHubRateLimitLimitKind = "secondary"
)

// GitHubRateLimitConfidence describes classification confidence presented to operators.
type GitHubRateLimitConfidence string

const (
	GitHubRateLimitConfidenceDeterministic   GitHubRateLimitConfidence = "deterministic"
	GitHubRateLimitConfidenceConservative    GitHubRateLimitConfidence = "conservative"
	GitHubRateLimitConfidenceProviderUnclear GitHubRateLimitConfidence = "provider_uncertain"
)

// GitHubRateLimitRecoveryHintKind is policy type used for next retry guidance.
type GitHubRateLimitRecoveryHintKind string

const (
	GitHubRateLimitRecoveryHintKindReset              GitHubRateLimitRecoveryHintKind = "rate_limit_reset"
	GitHubRateLimitRecoveryHintKindRetryAfter         GitHubRateLimitRecoveryHintKind = "retry_after"
	GitHubRateLimitRecoveryHintKindExponentialBackoff GitHubRateLimitRecoveryHintKind = "exponential_backoff"
	GitHubRateLimitRecoveryHintKindManualOnly         GitHubRateLimitRecoveryHintKind = "manual_only"
)

// GitHubRateLimitResumeActionKind identifies replay path used after recovery.
type GitHubRateLimitResumeActionKind string

const (
	GitHubRateLimitResumeActionKindRunStatusCommentRetry GitHubRateLimitResumeActionKind = "run_status_comment_retry"
	GitHubRateLimitResumeActionKindPlatformCallReplay    GitHubRateLimitResumeActionKind = "platform_github_call_replay"
	GitHubRateLimitResumeActionKindAgentSessionResume    GitHubRateLimitResumeActionKind = "agent_session_resume"
)

// GitHubRateLimitManualActionKind identifies terminal operator guidance.
type GitHubRateLimitManualActionKind string

const (
	GitHubRateLimitManualActionKindRequeuePlatformOperation GitHubRateLimitManualActionKind = "requeue_platform_operation"
	GitHubRateLimitManualActionKindResumeAgentSession       GitHubRateLimitManualActionKind = "resume_agent_session"
	GitHubRateLimitManualActionKindRetryAfterReview         GitHubRateLimitManualActionKind = "retry_after_operator_review"
)

// GitHubRateLimitEvidenceEventKind is append-only evidence lifecycle marker.
type GitHubRateLimitEvidenceEventKind string

const (
	GitHubRateLimitEvidenceEventSignalDetected       GitHubRateLimitEvidenceEventKind = "signal_detected"
	GitHubRateLimitEvidenceEventClassified           GitHubRateLimitEvidenceEventKind = "classified"
	GitHubRateLimitEvidenceEventResumeScheduled      GitHubRateLimitEvidenceEventKind = "resume_scheduled"
	GitHubRateLimitEvidenceEventResumeAttempted      GitHubRateLimitEvidenceEventKind = "resume_attempted"
	GitHubRateLimitEvidenceEventResumeFailed         GitHubRateLimitEvidenceEventKind = "resume_failed"
	GitHubRateLimitEvidenceEventResolved             GitHubRateLimitEvidenceEventKind = "resolved"
	GitHubRateLimitEvidenceEventManualActionRequired GitHubRateLimitEvidenceEventKind = "manual_action_required"
	GitHubRateLimitEvidenceEventCommentMirrorFailed  GitHubRateLimitEvidenceEventKind = "comment_mirror_failed"
)
