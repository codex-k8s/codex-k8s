package runstatus

// runStatusCommentUpsertedPayload is stored in flow_events for run status comment updates.
type runStatusCommentUpsertedPayload struct {
	RunID              string `json:"run_id"`
	IssueNumber        int    `json:"issue_number"`
	ThreadKind         string `json:"thread_kind,omitempty"`
	RepositoryFullName string `json:"repository_full_name"`
	CommentID          int64  `json:"comment_id"`
	CommentURL         string `json:"comment_url"`
	Phase              Phase  `json:"phase"`
}

// runNamespaceDeleteByStaffPayload is stored in flow_events for manual cleanup requests from staff UI.
type runNamespaceDeleteByStaffPayload struct {
	RunID              string `json:"run_id"`
	Namespace          string `json:"namespace"`
	Deleted            bool   `json:"deleted"`
	AlreadyDeleted     bool   `json:"already_deleted"`
	RunStatusCommentID int64  `json:"run_status_comment_id"`
	RunStatusURL       string `json:"run_status_url"`
	RequestedByType    string `json:"requested_by_type"`
	RequestedByID      string `json:"requested_by_id,omitempty"`
}

// triggerLabelConflictCommentPayload is stored in flow_events for conflict diagnostics.
type triggerLabelConflictCommentPayload struct {
	RepositoryFullName string   `json:"repository_full_name"`
	IssueNumber        int      `json:"issue_number"`
	TriggerLabel       string   `json:"trigger_label"`
	ConflictingLabels  []string `json:"conflicting_labels"`
	CommentID          int64    `json:"comment_id"`
	CommentURL         string   `json:"comment_url"`
}
