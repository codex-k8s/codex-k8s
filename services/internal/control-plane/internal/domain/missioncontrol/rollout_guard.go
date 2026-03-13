package missioncontrol

import (
	"fmt"

	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

// ResolveRolloutCapabilities validates rollout dependencies and returns the effective allowed paths.
func ResolveRolloutCapabilities(state valuetypes.MissionControlRolloutState) (valuetypes.MissionControlRolloutCapabilities, error) {
	if err := ValidateRolloutState(state); err != nil {
		return valuetypes.MissionControlRolloutCapabilities{}, err
	}

	canWarmup := state.CoreFeatureEnabled && state.SchemaReady && state.DomainReady
	canSnapshot := canWarmup && state.WarmupVerified && state.ReadPathEnabled
	canRealtime := canSnapshot && state.RealtimeEnabled
	canWrite := canSnapshot && state.WritePathEnabled

	return valuetypes.MissionControlRolloutCapabilities{
		CanRunWarmup:         canWarmup,
		CanServeSnapshot:     canSnapshot,
		CanOpenRealtime:      canRealtime,
		CanSubmitCoreCommand: canWrite,
		CanUseVoicePath:      canWrite && state.VoiceFeatureEnabled,
	}, nil
}

// ValidateRolloutState ensures Mission Control gates follow the documented rollout order.
func ValidateRolloutState(state valuetypes.MissionControlRolloutState) error {
	if !state.CoreFeatureEnabled {
		if state.VoiceFeatureEnabled || state.SchemaReady || state.DomainReady || state.WarmupVerified || state.ReadPathEnabled || state.RealtimeEnabled || state.WritePathEnabled {
			return fmt.Errorf("mission control rollout: core feature flag disabled but rollout state is partially enabled")
		}
		return nil
	}
	if state.DomainReady && !state.SchemaReady {
		return fmt.Errorf("mission control rollout: domain readiness requires schema readiness")
	}
	if state.WarmupVerified && (!state.SchemaReady || !state.DomainReady) {
		return fmt.Errorf("mission control rollout: warmup verification requires schema and domain readiness")
	}
	if state.ReadPathEnabled && !state.WarmupVerified {
		return fmt.Errorf("mission control rollout: read-path enablement requires warmup verification")
	}
	if state.RealtimeEnabled && !state.ReadPathEnabled {
		return fmt.Errorf("mission control rollout: realtime enablement requires read-path enablement")
	}
	if state.WritePathEnabled && !state.ReadPathEnabled {
		return fmt.Errorf("mission control rollout: write-path enablement requires read-path enablement")
	}
	if state.VoiceFeatureEnabled && !state.WritePathEnabled {
		return fmt.Errorf("mission control rollout: voice feature requires core write-path enablement")
	}
	return nil
}
