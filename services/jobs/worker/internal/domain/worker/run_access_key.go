package worker

import (
	"context"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
)

// IssueRunAccessKeyParams describes one run access key issuance request.
type IssueRunAccessKeyParams struct {
	RunID       string
	Namespace   string
	RuntimeMode agentdomain.RuntimeMode
	TargetEnv   string
	CreatedBy   string
	TTL         time.Duration
}

// IssuedRunAccessKey stores one plaintext run access key returned by control-plane.
type IssuedRunAccessKey struct {
	AccessKey string
}

// RunAccessKeyIssuer issues run-scoped OAuth bypass keys.
type RunAccessKeyIssuer interface {
	IssueRunAccessKey(ctx context.Context, params IssueRunAccessKeyParams) (IssuedRunAccessKey, error)
}
