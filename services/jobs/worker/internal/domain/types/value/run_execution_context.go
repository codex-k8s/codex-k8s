package value

import agentdomain "github.com/codex-k8s/kodex/libs/go/domain/agent"

// RunExecutionContext contains resolved execution mode and namespace metadata for one run.
type RunExecutionContext struct {
	RuntimeMode agentdomain.RuntimeMode
	Namespace   string
	IssueNumber int64
}
