package query

import (
	"encoding/json"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

// MCPActionRequestCreateParams describes one mcp_action_requests insert.
type MCPActionRequestCreateParams struct {
	CorrelationID string
	RunID         string
	ToolName      string
	Action        string
	TargetRef     json.RawMessage
	ApprovalMode  entitytypes.MCPApprovalMode
	ApprovalState entitytypes.MCPApprovalState
	RequestedBy   string
	AppliedBy     string
	Payload       json.RawMessage
}

// MCPActionRequestUpdateStateParams describes one approval state update.
type MCPActionRequestUpdateStateParams struct {
	ID            int64
	ApprovalState entitytypes.MCPApprovalState
	AppliedBy     string
	Payload       json.RawMessage
}
