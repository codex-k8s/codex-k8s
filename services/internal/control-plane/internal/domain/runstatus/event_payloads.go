package runstatus

// runStatusCommentUpsertedPayload is stored in flow_events for run status comment updates.
type runStatusCommentUpsertedPayload struct {
	RunID              string `json:"run_id"`
	IssueNumber        int    `json:"issue_number"`
	RepositoryFullName string `json:"repository_full_name"`
	CommentID          int64  `json:"comment_id"`
	CommentURL         string `json:"comment_url"`
	Phase              Phase  `json:"phase"`
}

// runNamespaceDeleteByTokenPayload is stored in flow_events for force cleanup requests.
type runNamespaceDeleteByTokenPayload struct {
	RunID              string `json:"run_id"`
	Namespace          string `json:"namespace"`
	Deleted            bool   `json:"deleted"`
	AlreadyDeleted     bool   `json:"already_deleted"`
	RunStatusCommentID int64  `json:"run_status_comment_id"`
	RunStatusURL       string `json:"run_status_url"`
}
