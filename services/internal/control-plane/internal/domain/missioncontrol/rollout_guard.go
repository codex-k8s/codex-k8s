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

	canSnapshot := state.SchemaReady && state.DomainReady
	canRealtime := canSnapshot
	canWarmup := state.CoreFeatureEnabled && canSnapshot
	canWrite := canWarmup && state.WarmupVerified && state.WritePathEnabled

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
	if state.DomainReady && !state.SchemaReady {
		return fmt.Errorf("mission control rollout: domain readiness requires schema readiness")
	}
	if state.WarmupVerified && (!state.SchemaReady || !state.DomainReady) {
		return fmt.Errorf("mission control rollout: warmup verification requires schema and domain readiness")
	}
	if state.WarmupVerified && !state.CoreFeatureEnabled {
		return fmt.Errorf("mission control rollout: warmup verification requires core feature enablement")
	}
	if state.WritePathEnabled && !state.CoreFeatureEnabled {
		return fmt.Errorf("mission control rollout: write-path enablement requires core feature enablement")
	}
	if state.WritePathEnabled && !state.WarmupVerified {
		return fmt.Errorf("mission control rollout: write-path enablement requires warmup verification")
	}
	if state.VoiceFeatureEnabled && !state.WritePathEnabled {
		return fmt.Errorf("mission control rollout: voice feature requires core write-path enablement")
	}
	return nil
}
