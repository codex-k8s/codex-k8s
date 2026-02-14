package query

import "time"

// AgentRunLookupItem is one run record for diagnostic lookup flows.
type AgentRunLookupItem struct {
	RunID              string
	CorrelationID      string
	ProjectID          string
	RepositoryFullName string
	AgentKey           string
	IssueNumber        int64
	IssueURL           string
	PullRequestNumber  int64
	PullRequestURL     string
	TriggerKind        string
	TriggerLabel       string
	Status             string
	CreatedAt          time.Time
	StartedAt          *time.Time
	FinishedAt         *time.Time
}
