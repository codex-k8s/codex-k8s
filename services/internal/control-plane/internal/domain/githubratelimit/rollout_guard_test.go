package githubratelimit

import (
	"strings"
	"testing"

	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

func TestResolveRolloutCapabilities_AllowsOnlySequencedPaths(t *testing.T) {
	caps, err := ResolveRolloutCapabilities(valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
		WorkerReady:        true,
		RunnerReady:        true,
		TransportReady:     true,
		UIFeatureEnabled:   true,
	})
	if err != nil {
		t.Fatalf("ResolveRolloutCapabilities error = %v", err)
	}
	if !caps.CanPersistWaits || !caps.CanRunWorkerSweep || !caps.CanAcceptSignals || !caps.CanServeTransport || !caps.CanExposeUI {
		t.Fatalf("unexpected rollout capabilities: %+v", caps)
	}
}

func TestResolveRolloutCapabilities_StopsBeforeTransport(t *testing.T) {
	caps, err := ResolveRolloutCapabilities(valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
	})
	if err != nil {
		t.Fatalf("ResolveRolloutCapabilities error = %v", err)
	}
	if !caps.CanPersistWaits {
		t.Fatalf("expected CanPersistWaits to be true")
	}
	if caps.CanRunWorkerSweep || caps.CanAcceptSignals || caps.CanServeTransport || caps.CanExposeUI {
		t.Fatalf("unexpected rollout capabilities: %+v", caps)
	}
}

func TestValidateRolloutState_RejectsRunnerBeforeWorker(t *testing.T) {
	err := ValidateRolloutState(valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
		RunnerReady:        true,
	})
	if err == nil || !strings.Contains(err.Error(), "runner enablement requires worker readiness") {
		t.Fatalf("ValidateRolloutState error = %v, want runner readiness failure", err)
	}
}

func TestValidateRolloutState_RejectsUIBeforeTransport(t *testing.T) {
	err := ValidateRolloutState(valuetypes.GitHubRateLimitRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
		WorkerReady:        true,
		RunnerReady:        true,
		UIFeatureEnabled:   true,
	})
	if err == nil || !strings.Contains(err.Error(), "ui enablement requires transport readiness") {
		t.Fatalf("ValidateRolloutState error = %v, want ui gating failure", err)
	}
}
