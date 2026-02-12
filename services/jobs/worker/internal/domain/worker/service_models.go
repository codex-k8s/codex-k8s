package worker

import (
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
