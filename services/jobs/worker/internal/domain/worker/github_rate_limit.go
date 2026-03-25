package worker

import (
	"context"
	"fmt"
	"time"
)

// GitHubRateLimitProcessResult describes one processed GitHub rate-limit wait sweep.
type GitHubRateLimitProcessResult struct {
	WaitID                string
	RunID                 string
	State                 string
	ResolutionKind        string
	AttemptNo             int
	ManualActionKind      string
	ResumeNotBefore       *time.Time
	RequeuedCorrelationID string
}

// GitHubRateLimitWaitProcessor exposes worker-facing control-plane sweep RPCs.
type GitHubRateLimitWaitProcessor interface {
	ProcessNextGitHubRateLimitWait(ctx context.Context, workerID string) (GitHubRateLimitProcessResult, bool, error)
}

func (s *Service) reconcileGitHubRateLimitWaits(ctx context.Context) error {
	if !s.githubRateLimitWaitEnabled() {
		return nil
	}

	for i := 0; i < s.cfg.GitHubRateLimitSweepLimit; i++ {
		result, found, err := s.githubRateLimits.ProcessNextGitHubRateLimitWait(ctx, s.cfg.WorkerID)
		if err != nil {
			return fmt.Errorf("process next github rate-limit wait: %w", err)
		}
		if !found {
			return nil
		}

		s.logger.Info(
			"github rate-limit wait processed",
			"worker_id", s.cfg.WorkerID,
			"wait_id", result.WaitID,
			"run_id", result.RunID,
			"state", result.State,
			"resolution_kind", result.ResolutionKind,
			"attempt_no", result.AttemptNo,
			"manual_action_kind", result.ManualActionKind,
			"requeued_correlation_id", result.RequeuedCorrelationID,
		)
	}
	return nil
}

func (s *Service) githubRateLimitWaitEnabled() bool {
	if s == nil {
		return false
	}
	if s.systemSettings != nil {
		return s.systemSettings.GitHubRateLimitWaitEnabled()
	}
	return s.cfg.GitHubRateLimitWaitEnabledFallback
}

func (s *Service) qualityGovernanceEnabled() bool {
	if s == nil || s.systemSettings == nil {
		return false
	}
	return s.systemSettings.QualityGovernanceEnabled()
}
