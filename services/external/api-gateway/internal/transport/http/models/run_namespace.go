package models

// RunNamespaceCleanupResponse describes force namespace cleanup result.
type RunNamespaceCleanupResponse struct {
	RunID          string `json:"run_id"`
	Namespace      string `json:"namespace"`
	Deleted        bool   `json:"deleted"`
	AlreadyDeleted bool   `json:"already_deleted"`
	CommentURL     string `json:"comment_url,omitempty"`
}
