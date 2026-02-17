package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

		// PrepareRunEnvironment is implemented as a blocking unary RPC on control-plane side
		// (it waits until runtime deploy task becomes terminal). Keep per-attempt context short
		// to avoid long-lived idle gRPC calls being terminated by infrastructure timeouts.
		attemptTimeout := s.cfg.RuntimePrepareRetryInterval * 4
		if attemptTimeout <= 0 {
			attemptTimeout = 15 * time.Second
		}
		if attemptTimeout < 5*time.Second {
			attemptTimeout = 5 * time.Second
		}
		if attemptTimeout > 30*time.Second {
			attemptTimeout = 30 * time.Second
		}

		attemptCtx, cancel := context.WithTimeout(ctx, attemptTimeout)
		prepared, err := s.deployer.PrepareRunEnvironment(attemptCtx, params)
		cancel()
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
	case codes.Internal:
		// Control-plane may wrap transient infra errors into Internal. Treat the most common
		// cases as retryable to avoid stuck runs when DB/control-plane temporarily restarts.
		msg := strings.ToLower(strings.TrimSpace(st.Message()))
		if msg == "" {
			return false
		}
		if strings.Contains(msg, "context deadline exceeded") {
			return true
		}
		if strings.Contains(msg, "context canceled") {
			return true
		}
		if strings.Contains(msg, "connection refused") || strings.Contains(msg, "dial tcp") {
			return true
		}
		if strings.Contains(msg, "connection reset") || strings.Contains(msg, "broken pipe") {
			return true
		}
		return false
	default:
		return false
	}
}
