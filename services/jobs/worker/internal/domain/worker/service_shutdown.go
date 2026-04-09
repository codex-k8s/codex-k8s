package worker

import (
	"context"
	"fmt"

	floweventdomain "github.com/codex-k8s/kodex/libs/go/domain/flowevent"
	runqueuerepo "github.com/codex-k8s/kodex/services/jobs/worker/internal/domain/repository/runqueue"
)

// ReleaseOwnedRunLeasesOnShutdown releases running-run leases held by current worker before pod termination.
func (s *Service) ReleaseOwnedRunLeasesOnShutdown(ctx context.Context) error {
	released, err := s.runs.ReleaseOwnedLeases(ctx, runqueuerepo.ReleaseOwnedLeasesParams{
		WorkerID: s.cfg.WorkerID,
	})
	if err != nil {
		return fmt.Errorf("release owned running leases on shutdown: %w", err)
	}

	for _, item := range released {
		if err := s.insertRunLeaseStaleEvent(ctx, item, floweventdomain.EventTypeRunLeaseReleased); err != nil {
			return fmt.Errorf("insert graceful run lease release event: %w", err)
		}
	}

	return nil
}
