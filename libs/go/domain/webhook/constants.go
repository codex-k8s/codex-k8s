package webhook

// IngestStatus represents normalized webhook ingestion state.
type IngestStatus string

const (
	IngestStatusAccepted  IngestStatus = "accepted"
	IngestStatusDuplicate IngestStatus = "duplicate"
	IngestStatusIgnored   IngestStatus = "ignored"
)

// GitHubEventType is a GitHub webhook event name from headers.
type GitHubEventType string

const (
	GitHubEventIssues      GitHubEventType = "issues"
	GitHubEventPullRequest GitHubEventType = "pull_request"
	GitHubEventPush        GitHubEventType = "push"
)

// GitHubAction is an action field from GitHub webhook payload.
type GitHubAction string

const (
	GitHubActionLabeled GitHubAction = "labeled"
)

// TriggerKind is an issue-label trigger flavor that maps to run behavior.
type TriggerKind string

const (
	TriggerKindDev       TriggerKind = "dev"
	TriggerKindDevRevise TriggerKind = "dev_revise"
)

const (
	TriggerSourceIssueLabel  = "issue_label"
	DefaultRunDevLabel       = "run:dev"
	DefaultRunDevReviseLabel = "run:dev:revise"
)
