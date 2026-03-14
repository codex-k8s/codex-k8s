package enum

// AgentRunWaitTargetKind identifies domain aggregate linked from agent_runs wait fields.
type AgentRunWaitTargetKind string

const (
	AgentRunWaitTargetKindApprovalRequest     AgentRunWaitTargetKind = "approval_request"
	AgentRunWaitTargetKindInteractionRequest  AgentRunWaitTargetKind = "interaction_request"
	AgentRunWaitTargetKindGitHubRateLimitWait AgentRunWaitTargetKind = "github_rate_limit_wait"
)
