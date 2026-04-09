package missioncontrol

import (
	"fmt"

	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

// ResolveRolloutCapabilities validates rollout dependencies and returns the effective allowed paths.
func ResolveRolloutCapabilities(state valuetypes.MissionControlRolloutState) (valuetypes.MissionControlRolloutCapabilities, error) {
	if err := ValidateRolloutState(state); err != nil {
		return valuetypes.MissionControlRolloutCapabilities{}, err
	}

	canSnapshot := state.SchemaReady && state.DomainReady
	canRealtime := canSnapshot
	canWarmup := canSnapshot
	canWrite := canWarmup

	return valuetypes.MissionControlRolloutCapabilities{
		CanRunWarmup:         canWarmup,
		CanServeSnapshot:     canSnapshot,
		CanOpenRealtime:      canRealtime,
		CanSubmitCoreCommand: canWrite,
		CanUseVoicePath:      canWrite,
	}, nil
}

// ValidateRolloutState ensures Mission Control gates follow the documented rollout order.
func ValidateRolloutState(state valuetypes.MissionControlRolloutState) error {
	if state.DomainReady && !state.SchemaReady {
		return fmt.Errorf("mission control rollout: domain readiness requires schema readiness")
	}
	return nil
}
