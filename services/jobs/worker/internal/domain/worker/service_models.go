package worker

import (
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

// finishRunParams carries all fields required to finalize a run and publish final events.
type finishRunParams struct {
	Run       runqueuerepo.RunningRun
	Execution valuetypes.RunExecutionContext
	Status    rundomain.Status
	EventType floweventdomain.EventType
	Ref       JobRef
	Extra     runFinishedEventExtra
	// SkipNamespaceCleanup disables namespace cleanup for deploy-only full-env runs.
	SkipNamespaceCleanup bool
}

// namespaceLifecycleEventParams describes one namespace lifecycle flow event.
type namespaceLifecycleEventParams struct {
	CorrelationID string
	EventType     floweventdomain.EventType
	RunID         string
	ProjectID     string
	Execution     valuetypes.RunExecutionContext
	Extra         namespaceLifecycleEventExtra
}

type namespaceLeaseSpec struct {
	AgentKey    string
	IssueNumber int64
	TTL         time.Duration
}

type runLaunchOptions struct {
	ServiceAccountName string
}
