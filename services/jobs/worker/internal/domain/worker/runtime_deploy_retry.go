package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) prepareRuntimeEnvironmentWithRetry(ctx context.Context, params PrepareRunEnvironmentParams) (PrepareRunEnvironmentResult, error) {
	if s.cfg.RuntimePrepareRetryTimeout <= 0 {
		return s.deployer.PrepareRunEnvironment(ctx, params)
	}

	deadline := time.Now().Add(s.cfg.RuntimePrepareRetryTimeout)
	attempt := 0
	for {
		attempt++
		prepared, err := s.deployer.PrepareRunEnvironment(ctx, params)
		if err == nil {
			return prepared, nil
		}
		if !isRetryableRuntimeDeployError(err) {
			return PrepareRunEnvironmentResult{}, err
		}
		if time.Now().After(deadline) {
			return PrepareRunEnvironmentResult{}, fmt.Errorf("runtime deploy prepare retry timeout exceeded after %d attempts: %w", attempt, err)
		}

		wait := s.cfg.RuntimePrepareRetryInterval
		if wait <= 0 {
			wait = 3 * time.Second
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return PrepareRunEnvironmentResult{}, ctx.Err()
		case <-timer.C:
		}
	}
}

func isRetryableRuntimeDeployError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled, codes.Aborted, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}
