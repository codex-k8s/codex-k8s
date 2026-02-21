package runstatus

const (
	localeRU = "ru"
	localeEN = "en"
)

const (
	commentMarkerPrefix = "<!-- codex-k8s:run-status "
	commentMarkerSuffix = " -->"
)

const (
	runManagementPathPrefix = "/runs/"
)

const (
	triggerKindDev       = "dev"
	triggerKindDevRevise = "dev_revise"
)

const (
	triggerSourceIssueLabel        = "issue_label"
	triggerSourcePullRequestReview = "pull_request_review"
)

const (
	runtimeModeFullEnv = "full-env"
	runtimeModeCode    = "code-only"
)

const (
	runStatusSucceeded = "succeeded"
	runStatusFailed    = "failed"
)

const (
	githubIssueReactionEyes = "eyes"
)

type commentTargetKind string

const (
	commentTargetKindIssue       commentTargetKind = "issue"
	commentTargetKindPullRequest commentTargetKind = "pull_request"
)

// TriggerWarningReasonCode is a stable machine-readable reason code
// for webhook events where run was not created and warning comment is posted.
type TriggerWarningReasonCode string

const (
	TriggerWarningReasonPullRequestReviewMissingStageLabel  TriggerWarningReasonCode = "pull_request_review_missing_stage_label"
	TriggerWarningReasonPullRequestReviewStageLabelConflict TriggerWarningReasonCode = "pull_request_review_stage_label_conflict"
	TriggerWarningReasonPullRequestReviewStageNotResolved   TriggerWarningReasonCode = "pull_request_review_stage_not_resolved"
	TriggerWarningReasonPullRequestReviewStageAmbiguous     TriggerWarningReasonCode = "pull_request_review_stage_ambiguous"
	TriggerWarningReasonRepositoryNotBoundForIssueLabel     TriggerWarningReasonCode = "repository_not_bound_for_issue_label"
)
