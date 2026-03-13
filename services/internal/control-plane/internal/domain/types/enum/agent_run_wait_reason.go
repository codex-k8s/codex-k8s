package enum

// AgentRunWaitReason is typed business meaning of one run wait-state.
type AgentRunWaitReason string

const (
	AgentRunWaitReasonOwnerReview        AgentRunWaitReason = "owner_review"
	AgentRunWaitReasonApprovalPending    AgentRunWaitReason = "approval_pending"
	AgentRunWaitReasonInteractionReply   AgentRunWaitReason = "interaction_response"
)
