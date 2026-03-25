package changegovernance

import (
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

// ResolveRolloutCapabilities validates rollout sequencing and returns effective capabilities.
func ResolveRolloutCapabilities(state valuetypes.ChangeGovernanceRolloutState) (valuetypes.ChangeGovernanceRolloutCapabilities, error) {
	if err := ValidateRolloutState(state); err != nil {
		return valuetypes.ChangeGovernanceRolloutCapabilities{}, err
	}
	canPersistFoundation := state.CoreFeatureEnabled && state.SchemaReady && state.DomainReady
	return valuetypes.ChangeGovernanceRolloutCapabilities{
		CanPersistFoundation:   canPersistFoundation,
		CanAcceptRunnerSignals: canPersistFoundation && state.RunnerReady,
	}, nil
}

// ValidateRolloutState enforces schema -> domain -> runner sequencing for the foundation path.
func ValidateRolloutState(state valuetypes.ChangeGovernanceRolloutState) error {
	if !state.CoreFeatureEnabled {
		if state.SchemaReady || state.DomainReady || state.RunnerReady {
			return errs.FailedPrecondition{Msg: "change governance rollout: core feature flag disabled but rollout state is partially enabled"}
		}
		return nil
	}
	if state.DomainReady && !state.SchemaReady {
		return errs.FailedPrecondition{Msg: "change governance rollout: domain readiness requires schema readiness"}
	}
	if state.RunnerReady && !state.DomainReady {
		return errs.FailedPrecondition{Msg: "change governance rollout: runner enablement requires domain readiness"}
	}
	return nil
}
