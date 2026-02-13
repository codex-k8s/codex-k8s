package models

// MCPApprovalCallbackRequest describes decision callback from external approver/executor adapters.
type MCPApprovalCallbackRequest struct {
	ApprovalRequestID int64  `json:"approval_request_id"`
	Decision          string `json:"decision"`
	Reason            string `json:"reason"`
	ActorID           string `json:"actor_id"`
}
