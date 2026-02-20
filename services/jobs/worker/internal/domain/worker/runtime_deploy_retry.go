package worker

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errRuntimeDeployTaskCanceled = errors.New("runtime deploy task canceled")

func (s *Service) prepareRuntimeEnvironmentPoll(ctx context.Context, params PrepareRunEnvironmentParams) (PrepareRunEnvironmentResult, bool, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, runtimePrepareAttemptTimeout(s.cfg.RuntimePrepareRetryInterval))
	prepared, err := s.deployer.PrepareRunEnvironment(attemptCtx, params)
	cancel()
	if err != nil {
		if isRuntimeDeployTaskCanceledError(err) {
			return PrepareRunEnvironmentResult{}, false, errRuntimeDeployTaskCanceled
		}
		if isRetryableRuntimeDeployError(err) {
			return PrepareRunEnvironmentResult{}, false, nil
		}
		return PrepareRunEnvironmentResult{}, false, err
	}
	if sanitizeDNSLabelValue(prepared.Namespace) == "" {
		return prepared, false, nil
	}
	return prepared, true, nil
}

func runtimePrepareAttemptTimeout(retryInterval time.Duration) time.Duration {
	// PrepareRunEnvironment is a blocking unary RPC on control-plane side
	// (it waits until runtime deploy task becomes terminal). Keep per-attempt context short
	// to avoid long-lived idle gRPC calls being terminated by infrastructure timeouts.
	attemptTimeout := retryInterval * 4
	if attemptTimeout <= 0 {
		attemptTimeout = 15 * time.Second
	}
	if attemptTimeout < 5*time.Second {
		attemptTimeout = 5 * time.Second
	}
	if attemptTimeout > 30*time.Second {
		attemptTimeout = 30 * time.Second
	}
	return attemptTimeout
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
	case codes.Unavailable, codes.DeadlineExceeded, codes.Aborted, codes.ResourceExhausted:
		return true
	case codes.Canceled:
		return !isRuntimeDeployTaskCanceledError(err)
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

func isRuntimeDeployTaskCanceledError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, errRuntimeDeployTaskCanceled) {
		return true
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Canceled {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(st.Message()))
	return strings.Contains(msg, "runtime deploy task canceled")
}
