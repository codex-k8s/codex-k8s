package query

const (
	NextStepActionKindIssueStageTransition = "issue_stage_transition"
	NextStepActionKindPullRequestLabelAdd  = "pull_request_label_add"

	NextStepThreadKindIssue       = "issue"
	NextStepThreadKindPullRequest = "pull_request"
)

// NextStepActionParams defines preview/execute parameters for one next-step action.
type NextStepActionParams struct {
	RepositoryFullName string
	IssueNumber        int
	PullRequestNumber  int
	ActionKind         string
	TargetLabel        string
}

// NextStepActionResult describes one next-step action preview/execute outcome.
type NextStepActionResult struct {
	RepositoryFullName string
	ThreadKind         string
	ThreadNumber       int
	ThreadURL          string
	RemovedLabels      []string
	AddedLabels        []string
	FinalLabels        []string
}
