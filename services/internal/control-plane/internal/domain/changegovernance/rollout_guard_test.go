package changegovernance

import (
	"strings"
	"testing"

	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

func TestResolveRolloutCapabilities_AllowsRunnerSignalsWhenFullyEnabled(t *testing.T) {
	t.Parallel()

	caps, err := ResolveRolloutCapabilities(valuetypes.ChangeGovernanceRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		DomainReady:        true,
		RunnerReady:        true,
	})
	if err != nil {
		t.Fatalf("ResolveRolloutCapabilities() error = %v", err)
	}
	if !caps.CanPersistFoundation || !caps.CanAcceptRunnerSignals {
		t.Fatalf("unexpected rollout capabilities: %+v", caps)
	}
}

func TestValidateRolloutState_RejectsRunnerBeforeDomain(t *testing.T) {
	t.Parallel()

	err := ValidateRolloutState(valuetypes.ChangeGovernanceRolloutState{
		CoreFeatureEnabled: true,
		SchemaReady:        true,
		RunnerReady:        true,
	})
	if err == nil || !strings.Contains(err.Error(), "runner enablement requires domain readiness") {
		t.Fatalf("ValidateRolloutState() error = %v", err)
	}
}
