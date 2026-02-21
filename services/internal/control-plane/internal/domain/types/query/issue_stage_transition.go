package query

// IssueStageLabelTransitionParams defines parameters for stage transition via issue labels.
type IssueStageLabelTransitionParams struct {
	RepositoryFullName string
	IssueNumber        int
	TargetLabel        string
}

// IssueStageLabelTransitionResult describes applied label transition for one issue.
type IssueStageLabelTransitionResult struct {
	RepositoryFullName string
	IssueNumber        int
	IssueURL           string
	RemovedLabels      []string
	AddedLabels        []string
	FinalLabels        []string
}
