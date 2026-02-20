package worker

import (
	"context"
	"fmt"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

func (s *Service) finishRuntimePrepareCanceledRun(ctx context.Context, run runqueuerepo.RunningRun, execution valuetypes.RunExecutionContext, skipNamespaceCleanup bool) error {
	if err := s.finishRun(ctx, finishRunParams{
		Run:                  run,
		Execution:            execution,
		Status:               rundomain.StatusCanceled,
		EventType:            floweventdomain.EventTypeRunCanceled,
		Ref:                  s.launcher.JobRef(run.RunID, execution.Namespace),
		SkipNamespaceCleanup: skipNamespaceCleanup,
		Extra: runFinishedEventExtra{
			Reason: runFailureReasonRuntimeDeployCanceled,
		},
	}); err != nil {
		return fmt.Errorf("mark run canceled after runtime deploy cancellation: %w", err)
	}
	return nil
}
