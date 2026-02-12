package worker

import (
	"context"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
)

// IssueMCPTokenParams describes one MCP token issuance request for run.
type IssueMCPTokenParams struct {
	RunID       string
	Namespace   string
	RuntimeMode agentdomain.RuntimeMode
}

// IssuedMCPToken stores issued MCP token and expiration timestamp.
type IssuedMCPToken struct {
	Token     string
	ExpiresAt time.Time
}

// MCPTokenIssuer issues run-bound MCP token via control-plane contract.
type MCPTokenIssuer interface {
	IssueRunMCPToken(ctx context.Context, params IssueMCPTokenParams) (IssuedMCPToken, error)
}
