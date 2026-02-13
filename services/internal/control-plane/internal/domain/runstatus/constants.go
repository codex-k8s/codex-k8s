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

type commentTargetKind string

const (
	commentTargetKindIssue       commentTargetKind = "issue"
	commentTargetKindPullRequest commentTargetKind = "pull_request"
)
