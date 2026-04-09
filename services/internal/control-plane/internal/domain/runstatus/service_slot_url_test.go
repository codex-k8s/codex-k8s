package runstatus

import (
	"testing"

	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
)

func TestResolveRunSlotURL_RecomputesStaleURLForAISlot(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			AIDomain:         "ai.platform.kodex.works",
			ProductionDomain: "platform.kodex.works",
		},
	}

	got := svc.resolveRunSlotURL(runContext{
		payload: querytypes.RunPayload{
			Runtime: &querytypes.RunPayloadRuntime{},
		},
	}, commentState{
		RuntimeMode: runtimeModeFullEnv,
		Namespace:   "kodex-dev-2",
		SlotURL:     "https://platform.kodex.works",
	})

	want := "https://kodex-dev-2.ai.platform.kodex.works"
	if got != want {
		t.Fatalf("resolveRunSlotURL() = %q, want %q", got, want)
	}
}

func TestResolveRunSlotURL_HidesURLWhenTargetEnvAndNamespaceAreNotFinal(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			AIDomain:         "ai.platform.kodex.works",
			ProductionDomain: "platform.kodex.works",
		},
	}

	got := svc.resolveRunSlotURL(runContext{
		payload: querytypes.RunPayload{
			Runtime: &querytypes.RunPayloadRuntime{},
		},
	}, commentState{
		RuntimeMode: runtimeModeFullEnv,
		Namespace:   "codex-issue-3278207d1cd3-i77-ra335a61f755",
		SlotURL:     "https://platform.kodex.works",
	})

	if got != "" {
		t.Fatalf("resolveRunSlotURL() = %q, want empty string", got)
	}
}

func TestResolveRunSlotURL_UsesProductionDomainForExplicitProductionTarget(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			AIDomain:         "ai.platform.kodex.works",
			ProductionDomain: "platform.kodex.works",
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
		Namespace:   "kodex-prod",
	})

	want := "https://platform.kodex.works"
	if got != want {
		t.Fatalf("resolveRunSlotURL() = %q, want %q", got, want)
	}
}

func TestResolveRunSlotURL_UsesRuntimePublicHostOverride(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			AIDomain:         "ai.platform.kodex.works",
			ProductionDomain: "platform.kodex.works",
		},
	}

	got := svc.resolveRunSlotURL(runContext{
		payload: querytypes.RunPayload{
			Runtime: &querytypes.RunPayloadRuntime{
				PublicHost: "kodex-dev-3.ai.platform.kodex.works",
			},
		},
	}, commentState{
		RuntimeMode: runtimeModeFullEnv,
	})

	want := "https://kodex-dev-3.ai.platform.kodex.works"
	if got != want {
		t.Fatalf("resolveRunSlotURL() = %q, want %q", got, want)
	}
}
