package enum

// AgentSessionWaitState is persisted wait-state snapshot for one resumable run session.
type AgentSessionWaitState string

const (
	AgentSessionWaitStateNone         AgentSessionWaitState = ""
	AgentSessionWaitStateOwnerReview  AgentSessionWaitState = "owner_review"
	AgentSessionWaitStateMCP          AgentSessionWaitState = "mcp"
	AgentSessionWaitStateBackpressure AgentSessionWaitState = "backpressure"
)
