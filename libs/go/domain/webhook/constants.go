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
	GitHubEventIssueComment      GitHubEventType = "issue_comment"
	GitHubEventPullRequest       GitHubEventType = "pull_request"
	GitHubEventPullRequestReview GitHubEventType = "pull_request_review"
	GitHubEventPush              GitHubEventType = "push"
)

// GitHubAction is an action field from GitHub webhook payload.
type GitHubAction string

const (
	GitHubActionLabeled   GitHubAction = "labeled"
	GitHubActionCreated   GitHubAction = "created"
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
	TriggerKindAIRepair     TriggerKind = "ai_repair"
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
	TriggerSourceIssueComment         = "issue_comment"
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
	DefaultRunDevLabel            = "run:dev"
	DefaultRunDevReviseLabel      = "run:dev:revise"
	DefaultRunDebugLabel          = "run:debug"
	DefaultRunDocAuditLabel       = "run:doc-audit"
	DefaultRunDocAuditReviseLabel = "run:doc-audit:revise"
	DefaultRunAIRepairLabel       = "run:ai-repair"
)

const (
	DefaultRunQALabel               = "run:qa"
	DefaultRunQAReviseLabel         = "run:qa:revise"
	DefaultRunReleaseLabel          = "run:release"
	DefaultRunReleaseReviseLabel    = "run:release:revise"
	DefaultRunPostDeployLabel       = "run:postdeploy"
	DefaultRunPostDeployReviseLabel = "run:postdeploy:revise"
)

const (
	DefaultRunOpsLabel               = "run:ops"
	DefaultRunOpsReviseLabel         = "run:ops:revise"
	DefaultRunSelfImproveLabel       = "run:self-improve"
	DefaultRunSelfImproveReviseLabel = "run:self-improve:revise"
	DefaultRunRethinkLabel           = "run:rethink"
	DefaultModeDiscussionLabel       = "mode:discussion"
)

const (
	DefaultStateBlockedLabel    = "state:blocked"
	DefaultStateInReviewLabel   = "state:in-review"
	DefaultStateApprovedLabel   = "state:approved"
	DefaultStateSupersededLabel = "state:superseded"
	DefaultStateAbandonedLabel  = "state:abandoned"
)

const (
	DefaultNeedInputLabel    = "need:input"
	DefaultNeedPMLabel       = "need:pm"
	DefaultNeedSALabel       = "need:sa"
	DefaultNeedQALabel       = "need:qa"
	DefaultNeedSRELabel      = "need:sre"
	DefaultNeedEMLabel       = "need:em"
	DefaultNeedKMLabel       = "need:km"
	DefaultNeedReviewerLabel = "need:reviewer"
)
