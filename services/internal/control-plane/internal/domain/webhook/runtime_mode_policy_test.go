package webhook

import (
	"testing"

	agentdomain "github.com/codex-k8s/kodex/libs/go/domain/agent"
	webhookdomain "github.com/codex-k8s/kodex/libs/go/domain/webhook"
)

func TestRuntimeModePolicyResolve_DefaultTriggerBehavior(t *testing.T) {
	t.Parallel()

	mode, source := DefaultRuntimeModePolicy().resolve(&issueRunTrigger{Kind: webhookdomain.TriggerKindDev})
	if mode != agentdomain.RuntimeModeFullEnv {
		t.Fatalf("unexpected mode: %q", mode)
	}
	if source != runtimeModeSourceTriggerDefault {
		t.Fatalf("unexpected source: %q", source)
	}
}

func TestRuntimeModePolicyResolve_UsesServicesYAMLMap(t *testing.T) {
	t.Parallel()

	policy := RuntimeModePolicy{
		Configured:  true,
		Source:      "services.yaml",
		DefaultMode: agentdomain.RuntimeModeFullEnv,
		TriggerModes: map[webhookdomain.TriggerKind]agentdomain.RuntimeMode{
			webhookdomain.TriggerKindSelfImprove: agentdomain.RuntimeModeCodeOnly,
		},
	}

	mode, source := policy.resolve(&issueRunTrigger{Kind: webhookdomain.TriggerKindSelfImprove})
	if mode != agentdomain.RuntimeModeCodeOnly {
		t.Fatalf("unexpected mode: %q", mode)
	}
	if source != runtimeModeSourceServicesYAML {
		t.Fatalf("unexpected source: %q", source)
	}
}

func TestRuntimeModePolicyResolve_UsesServicesYAMLMapForSelfImproveRevise(t *testing.T) {
	t.Parallel()

	policy := RuntimeModePolicy{
		Configured:  true,
		Source:      "services.yaml",
		DefaultMode: agentdomain.RuntimeModeFullEnv,
		TriggerModes: map[webhookdomain.TriggerKind]agentdomain.RuntimeMode{
			webhookdomain.TriggerKindSelfImproveRevise: agentdomain.RuntimeModeCodeOnly,
		},
	}

	mode, source := policy.resolve(&issueRunTrigger{Kind: webhookdomain.TriggerKindSelfImproveRevise})
	if mode != agentdomain.RuntimeModeCodeOnly {
		t.Fatalf("unexpected mode: %q", mode)
	}
	if source != runtimeModeSourceServicesYAML {
		t.Fatalf("unexpected source: %q", source)
	}
}

func TestRuntimeModePolicyResolve_AIRepairDefaultsToCodeOnly(t *testing.T) {
	t.Parallel()

	mode, source := DefaultRuntimeModePolicy().resolve(&issueRunTrigger{Kind: webhookdomain.TriggerKindAIRepair})
	if mode != agentdomain.RuntimeModeCodeOnly {
		t.Fatalf("unexpected mode: %q", mode)
	}
	if source != runtimeModeSourceTriggerDefault {
		t.Fatalf("unexpected source: %q", source)
	}
}
