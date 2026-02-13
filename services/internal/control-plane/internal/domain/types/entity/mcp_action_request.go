package entity

import (
	"encoding/json"
	"time"
)

// MCPApprovalMode defines who must approve a privileged MCP action.
type MCPApprovalMode string

const (
	MCPApprovalModeNone      MCPApprovalMode = "none"
	MCPApprovalModeOwner     MCPApprovalMode = "owner"
	MCPApprovalModeDelegated MCPApprovalMode = "delegated"
)

// MCPApprovalState is lifecycle state for one MCP action request.
type MCPApprovalState string

const (
	MCPApprovalStateRequested MCPApprovalState = "requested"
	MCPApprovalStateApproved  MCPApprovalState = "approved"
	MCPApprovalStateDenied    MCPApprovalState = "denied"
	MCPApprovalStateExpired   MCPApprovalState = "expired"
	MCPApprovalStateFailed    MCPApprovalState = "failed"
	MCPApprovalStateApplied   MCPApprovalState = "applied"
)

// MCPActionRequest stores approval and execution lifecycle for privileged MCP tools.
type MCPActionRequest struct {
	ID            int64
	CorrelationID string
	RunID         string
	ProjectID     string
	ProjectSlug   string
	ProjectName   string
	IssueNumber   int
	PRNumber      int
	TriggerLabel  string
	ToolName      string
	Action        string
	TargetRef     json.RawMessage
	ApprovalMode  MCPApprovalMode
	ApprovalState MCPApprovalState
	RequestedBy   string
	AppliedBy     string
	Payload       json.RawMessage
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
