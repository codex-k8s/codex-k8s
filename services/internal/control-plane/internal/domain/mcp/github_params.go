package mcp

// GitHubGetIssueParams describes issue read operation in adapter.
type GitHubGetIssueParams struct {
	Token       string
	Owner       string
	Repository  string
	IssueNumber int
}

// GitHubGetPullRequestParams describes pull request read operation in adapter.
type GitHubGetPullRequestParams struct {
	Token             string
	Owner             string
	Repository        string
	PullRequestNumber int
}

// GitHubListIssueCommentsParams describes issue comments list operation in adapter.
type GitHubListIssueCommentsParams struct {
	Token       string
	Owner       string
	Repository  string
	IssueNumber int
	Limit       int
}

// GitHubListIssueLabelsParams describes issue labels list operation in adapter.
type GitHubListIssueLabelsParams struct {
	Token       string
	Owner       string
	Repository  string
	IssueNumber int
}

// GitHubListBranchesParams describes branches list operation in adapter.
type GitHubListBranchesParams struct {
	Token      string
	Owner      string
	Repository string
	Limit      int
}

// GitHubEnsureBranchParams describes branch create/sync operation in adapter.
type GitHubEnsureBranchParams struct {
	Token      string
	Owner      string
	Repository string
	BranchName string
	BaseBranch string
	BaseSHA    string
	Force      bool
}

// GitHubUpsertPullRequestParams describes create/update PR operation in adapter.
type GitHubUpsertPullRequestParams struct {
	Token             string
	Owner             string
	Repository        string
	PullRequestNumber int
	Title             string
	Body              string
	HeadBranch        string
	BaseBranch        string
	Draft             bool
}

// GitHubCreateIssueCommentParams describes issue/PR comment create operation in adapter.
type GitHubCreateIssueCommentParams struct {
	Token       string
	Owner       string
	Repository  string
	IssueNumber int
	Body        string
}

// GitHubEditIssueCommentParams describes issue/PR comment update operation in adapter.
type GitHubEditIssueCommentParams struct {
	Token      string
	Owner      string
	Repository string
	CommentID  int64
	Body       string
}

// GitHubMutateLabelsParams describes labels add/remove operation in adapter.
type GitHubMutateLabelsParams struct {
	Token       string
	Owner       string
	Repository  string
	IssueNumber int
	Labels      []string
}

// GitHubUpsertRepositorySecretParams describes repository secret upsert operation in adapter.
type GitHubUpsertRepositorySecretParams struct {
	Token      string
	Owner      string
	Repository string
	SecretName string
	Value      string
}
