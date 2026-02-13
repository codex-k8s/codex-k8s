package mcpactionrequest

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	Item              = entitytypes.MCPActionRequest
	CreateParams      = querytypes.MCPActionRequestCreateParams
	UpdateStateParams = querytypes.MCPActionRequestUpdateStateParams
	ApprovalState     = entitytypes.MCPApprovalState
)

// Repository stores approval lifecycle for privileged MCP actions.
type Repository interface {
	// Create inserts a new action request row.
	Create(ctx context.Context, params CreateParams) (Item, error)
	// GetByID returns one request by id.
	GetByID(ctx context.Context, id int64) (Item, bool, error)
	// FindLatestBySignature returns latest request with the same action signature.
	FindLatestBySignature(ctx context.Context, runID string, toolName string, action string, targetRefJSON []byte) (Item, bool, error)
	// FindPendingBySignature returns latest pending request for idempotent retries.
	FindPendingBySignature(ctx context.Context, runID string, toolName string, action string, targetRefJSON []byte) (Item, bool, error)
	// ListPending returns pending approval queue.
	ListPending(ctx context.Context, limit int) ([]Item, error)
	// UpdateState updates approval state and returns updated row.
	UpdateState(ctx context.Context, params UpdateStateParams) (Item, bool, error)
}
