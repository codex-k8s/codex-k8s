package githubratelimit

import (
	"fmt"

	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

// ResolveRolloutCapabilities validates rollout sequencing and returns effective allowed paths.
func ResolveRolloutCapabilities(state valuetypes.GitHubRateLimitRolloutState) (valuetypes.GitHubRateLimitRolloutCapabilities, error) {
	if err := ValidateRolloutState(state); err != nil {
		return valuetypes.GitHubRateLimitRolloutCapabilities{}, err
	}

	canPersistWaits := state.CoreFeatureEnabled && state.SchemaReady && state.DomainReady
	canRunWorkerSweep := canPersistWaits && state.WorkerReady
	canAcceptSignals := canRunWorkerSweep && state.RunnerReady
	canServeTransport := canAcceptSignals && state.TransportReady

	return valuetypes.GitHubRateLimitRolloutCapabilities{
		CanPersistWaits:   canPersistWaits,
		CanRunWorkerSweep: canRunWorkerSweep,
		CanAcceptSignals:  canAcceptSignals,
		CanServeTransport: canServeTransport,
		CanExposeUI:       canServeTransport && state.UIFeatureEnabled,
	}, nil
}

// ValidateRolloutState enforces the documented schema -> domain -> worker -> runner -> transport -> UI order.
func ValidateRolloutState(state valuetypes.GitHubRateLimitRolloutState) error {
	if !state.CoreFeatureEnabled {
		if state.SchemaReady || state.DomainReady || state.WorkerReady || state.RunnerReady || state.TransportReady || state.UIFeatureEnabled {
			return fmt.Errorf("github rate-limit rollout: core feature flag disabled but rollout state is partially enabled")
		}
		return nil
	}
	if state.DomainReady && !state.SchemaReady {
		return fmt.Errorf("github rate-limit rollout: domain readiness requires schema readiness")
	}
	if state.WorkerReady && !state.DomainReady {
		return fmt.Errorf("github rate-limit rollout: worker enablement requires domain readiness")
	}
	if state.RunnerReady && !state.WorkerReady {
		return fmt.Errorf("github rate-limit rollout: runner enablement requires worker readiness")
	}
	if state.TransportReady && !state.RunnerReady {
		return fmt.Errorf("github rate-limit rollout: transport enablement requires runner readiness")
	}
	if state.UIFeatureEnabled && !state.TransportReady {
		return fmt.Errorf("github rate-limit rollout: ui enablement requires transport readiness")
	}
	return nil
}
