package webhook

import (
	"testing"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
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
