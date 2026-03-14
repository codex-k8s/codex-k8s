package value

import "time"

// MissionControlWorkItemProjectionPayload stores persisted work-item detail fragments for Mission Control.
type MissionControlWorkItemProjectionPayload struct {
	RepositoryFullName string     `json:"repository_full_name"`
	IssueNumber        int64      `json:"issue_number"`
	IssueURL           string     `json:"issue_url,omitempty"`
	LastRunID          string     `json:"last_run_id,omitempty"`
	LastStatus         string     `json:"last_status,omitempty"`
	TriggerKind        string     `json:"trigger_kind,omitempty"`
	WorkItemType       string     `json:"work_item_type,omitempty"`
	StageLabel         string     `json:"stage_label,omitempty"`
	Labels             []string   `json:"labels,omitempty"`
	Owner              string     `json:"owner,omitempty"`
	Assignees          []string   `json:"assignees,omitempty"`
	LastProviderSyncAt *time.Time `json:"last_provider_sync_at,omitempty"`
}

// MissionControlDiscussionProjectionPayload stores persisted discussion detail fragments for Mission Control.
type MissionControlDiscussionProjectionPayload struct {
	DiscussionKind       string `json:"discussion_kind,omitempty"`
	Status               string `json:"status,omitempty"`
	Author               string `json:"author,omitempty"`
	ParticipantCount     int32  `json:"participant_count,omitempty"`
	LatestCommentExcerpt string `json:"latest_comment_excerpt,omitempty"`
	FormalizationTarget  string `json:"formalization_target,omitempty"`
}

// MissionControlPullRequestProjectionPayload stores persisted pull-request detail fragments for Mission Control.
type MissionControlPullRequestProjectionPayload struct {
	RepositoryFullName string   `json:"repository_full_name"`
	PullRequestNumber  int64    `json:"pull_request_number"`
	PullRequestURL     string   `json:"pull_request_url,omitempty"`
	LastRunID          string   `json:"last_run_id,omitempty"`
	LastStatus         string   `json:"last_status,omitempty"`
	BranchHead         string   `json:"branch_head,omitempty"`
	BranchBase         string   `json:"branch_base,omitempty"`
	MergeState         string   `json:"merge_state,omitempty"`
	ReviewDecision     string   `json:"review_decision,omitempty"`
	ChecksSummary      string   `json:"checks_summary,omitempty"`
	LinkedIssueRefs    []string `json:"linked_issue_refs,omitempty"`
}

// MissionControlAgentProjectionPayload stores persisted agent detail fragments for Mission Control.
type MissionControlAgentProjectionPayload struct {
	AgentKey          string     `json:"agent_key"`
	LastRunID         string     `json:"last_run_id,omitempty"`
	LastStatus        string     `json:"last_status,omitempty"`
	RuntimeMode       string     `json:"runtime_mode,omitempty"`
	WaitingReason     string     `json:"waiting_reason,omitempty"`
	LastHeartbeatAt   *time.Time `json:"last_heartbeat_at,omitempty"`
	LastRunRepository string     `json:"last_run_repository,omitempty"`
}
