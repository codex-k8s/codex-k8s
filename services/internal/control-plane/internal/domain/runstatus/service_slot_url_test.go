package runstatus

import (
	"testing"

	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func TestResolveRunSlotURL_RecomputesStaleURLForAISlot(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			AIDomain:         "ai.platform.codex-k8s.dev",
			ProductionDomain: "platform.codex-k8s.dev",
		},
	}

	got := svc.resolveRunSlotURL(runContext{
		payload: querytypes.RunPayload{
			Runtime: &querytypes.RunPayloadRuntime{},
		},
	}, commentState{
		RuntimeMode: runtimeModeFullEnv,
		Namespace:   "codex-k8s-dev-2",
		SlotURL:     "https://platform.codex-k8s.dev",
	})

	want := "https://codex-k8s-dev-2.ai.platform.codex-k8s.dev"
	if got != want {
		t.Fatalf("resolveRunSlotURL() = %q, want %q", got, want)
	}
}

func TestResolveRunSlotURL_HidesURLWhenTargetEnvAndNamespaceAreNotFinal(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			AIDomain:         "ai.platform.codex-k8s.dev",
			ProductionDomain: "platform.codex-k8s.dev",
		},
	}

	got := svc.resolveRunSlotURL(runContext{
		payload: querytypes.RunPayload{
			Runtime: &querytypes.RunPayloadRuntime{},
		},
	}, commentState{
		RuntimeMode: runtimeModeFullEnv,
		Namespace:   "codex-issue-3278207d1cd3-i77-ra335a61f755",
		SlotURL:     "https://platform.codex-k8s.dev",
	})

	if got != "" {
		t.Fatalf("resolveRunSlotURL() = %q, want empty string", got)
	}
}

func TestResolveRunSlotURL_UsesProductionDomainForExplicitProductionTarget(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			AIDomain:         "ai.platform.codex-k8s.dev",
			ProductionDomain: "platform.codex-k8s.dev",
		},
	}

	got := svc.resolveRunSlotURL(runContext{
		payload: querytypes.RunPayload{
			Runtime: &querytypes.RunPayloadRuntime{
				TargetEnv: "production",
			},
		},
	}, commentState{
		RuntimeMode: runtimeModeFullEnv,
		Namespace:   "codex-k8s-prod",
	})

	want := "https://platform.codex-k8s.dev"
	if got != want {
		t.Fatalf("resolveRunSlotURL() = %q, want %q", got, want)
	}
}

func TestResolveRunSlotURL_UsesRuntimePublicHostOverride(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			AIDomain:         "ai.platform.codex-k8s.dev",
			ProductionDomain: "platform.codex-k8s.dev",
		},
	}

	got := svc.resolveRunSlotURL(runContext{
		payload: querytypes.RunPayload{
			Runtime: &querytypes.RunPayloadRuntime{
				PublicHost: "codex-k8s-dev-3.ai.platform.codex-k8s.dev",
			},
		},
	}, commentState{
		RuntimeMode: runtimeModeFullEnv,
	})

	want := "https://codex-k8s-dev-3.ai.platform.codex-k8s.dev"
	if got != want {
		t.Fatalf("resolveRunSlotURL() = %q, want %q", got, want)
	}
}
