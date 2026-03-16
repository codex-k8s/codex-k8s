package worker

import (
	"testing"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
)

func TestResolveRunControlPlaneGRPCTarget(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		runtimeMode         agentdomain.RuntimeMode
		productionNamespace string
		fallbackTarget      string
		want                string
	}{
		{
			name:                "full env uses production fqdn",
			runtimeMode:         agentdomain.RuntimeModeFullEnv,
			productionNamespace: "codex-k8s-prod",
			fallbackTarget:      "codex-k8s-control-plane:9090",
			want:                "codex-k8s-control-plane.codex-k8s-prod.svc.cluster.local:9090",
		},
		{
			name:                "full env falls back when production namespace is empty",
			runtimeMode:         agentdomain.RuntimeModeFullEnv,
			productionNamespace: "",
			fallbackTarget:      "codex-k8s-control-plane:9090",
			want:                "codex-k8s-control-plane:9090",
		},
		{
			name:                "code only keeps fallback target",
			runtimeMode:         agentdomain.RuntimeModeCodeOnly,
			productionNamespace: "codex-k8s-prod",
			fallbackTarget:      "codex-k8s-control-plane:9090",
			want:                "codex-k8s-control-plane:9090",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := resolveRunControlPlaneGRPCTarget(tc.runtimeMode, tc.productionNamespace, tc.fallbackTarget)
			if got != tc.want {
				t.Fatalf("resolveRunControlPlaneGRPCTarget() = %q, want %q", got, tc.want)
			}
		})
	}
}
