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
	GitHubEventIssues            GitHubEventType = "issues"
	GitHubEventPullRequest       GitHubEventType = "pull_request"
	GitHubEventPullRequestReview GitHubEventType = "pull_request_review"
	GitHubEventPush              GitHubEventType = "push"
)

// GitHubAction is an action field from GitHub webhook payload.
type GitHubAction string

const (
	GitHubActionLabeled   GitHubAction = "labeled"
	GitHubActionSubmitted GitHubAction = "submitted"
)

// TriggerKind is an issue-label trigger flavor that maps to run behavior.
type TriggerKind string

const (
	TriggerKindIntake       TriggerKind = "intake"
	TriggerKindIntakeRevise TriggerKind = "intake_revise"
	TriggerKindVision       TriggerKind = "vision"
	TriggerKindVisionRevise TriggerKind = "vision_revise"
	TriggerKindPRD          TriggerKind = "prd"
	TriggerKindPRDRevise    TriggerKind = "prd_revise"
	TriggerKindArch         TriggerKind = "arch"
	TriggerKindArchRevise   TriggerKind = "arch_revise"
	TriggerKindDesign       TriggerKind = "design"
	TriggerKindDesignRevise TriggerKind = "design_revise"
	TriggerKindPlan         TriggerKind = "plan"
	TriggerKindPlanRevise   TriggerKind = "plan_revise"
	TriggerKindDev          TriggerKind = "dev"
	TriggerKindDevRevise    TriggerKind = "dev_revise"
	TriggerKindDocAudit     TriggerKind = "doc_audit"
	TriggerKindQA           TriggerKind = "qa"
	TriggerKindRelease      TriggerKind = "release"
	TriggerKindPostDeploy   TriggerKind = "postdeploy"
	TriggerKindOps          TriggerKind = "ops"
	TriggerKindSelfImprove  TriggerKind = "self_improve"
	TriggerKindRethink      TriggerKind = "rethink"
)

const (
	GitHubReviewStateChangesRequested = "changes_requested"
	TriggerSourceIssueLabel           = "issue_label"
	TriggerSourcePullRequestReview    = "pull_request_review"
)

const (
	DefaultRunIntakeLabel       = "run:intake"
	DefaultRunIntakeReviseLabel = "run:intake:revise"
	DefaultRunVisionLabel       = "run:vision"
)

const (
	DefaultRunVisionReviseLabel = "run:vision:revise"
	DefaultRunPRDLabel          = "run:prd"
	DefaultRunPRDReviseLabel    = "run:prd:revise"
)

const (
	DefaultRunArchLabel       = "run:arch"
	DefaultRunArchReviseLabel = "run:arch:revise"
	DefaultRunDesignLabel     = "run:design"
)

const (
	DefaultRunDesignReviseLabel = "run:design:revise"
	DefaultRunPlanLabel         = "run:plan"
	DefaultRunPlanReviseLabel   = "run:plan:revise"
)

const (
	DefaultRunDevLabel       = "run:dev"
	DefaultRunDevReviseLabel = "run:dev:revise"
	DefaultRunDocAuditLabel  = "run:doc-audit"
)

const (
	DefaultRunQALabel         = "run:qa"
	DefaultRunReleaseLabel    = "run:release"
	DefaultRunPostDeployLabel = "run:postdeploy"
)

const (
	DefaultRunOpsLabel         = "run:ops"
	DefaultRunSelfImproveLabel = "run:self-improve"
	DefaultRunRethinkLabel     = "run:rethink"
)
