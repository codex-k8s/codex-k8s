package value

import agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"

// RunExecutionContext contains resolved execution mode and namespace metadata for one run.
type RunExecutionContext struct {
	RuntimeMode agentdomain.RuntimeMode
	Namespace   string
	IssueNumber int64
}
