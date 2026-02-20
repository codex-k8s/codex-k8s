package worker

import (
	"context"
	"fmt"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

func classifyAgentContextResolveError(err error) (floweventdomain.EventType, runFailureReason) {
	if isFailedPreconditionError(err) {
		return floweventdomain.EventTypeRunFailedPrecondition, runFailureReasonPreconditionFailed
	}
	return floweventdomain.EventTypeRunFailedLaunchError, runFailureReasonAgentContextResolve
}

func (s *Service) failRunAfterAgentContextResolve(ctx context.Context, run runqueuerepo.RunningRun, execution valuetypes.RunExecutionContext, resolveErr error) error {
	eventType, reason := classifyAgentContextResolveError(resolveErr)
	if finishErr := s.finishRun(ctx, finishRunParams{
		Run:       run,
		Execution: execution,
		Status:    rundomain.StatusFailed,
		EventType: eventType,
		Ref:       s.launcher.JobRef(run.RunID, execution.Namespace),
		Extra: runFinishedEventExtra{
			Error:  resolveErr.Error(),
			Reason: reason,
		},
	}); finishErr != nil {
		return fmt.Errorf("mark run failed after context resolve error: %w", finishErr)
	}
	return nil
}
