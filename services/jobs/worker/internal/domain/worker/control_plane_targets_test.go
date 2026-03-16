package worker

import (
	"testing"
)

func TestResolveRunControlPlaneGRPCTarget(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		productionNamespace string
		fallbackTarget      string
		want                string
	}{
		{
			name:                "uses production fqdn when namespace is configured",
			productionNamespace: "codex-k8s-prod",
			fallbackTarget:      "codex-k8s-control-plane:9090",
			want:                "codex-k8s-control-plane.codex-k8s-prod.svc.cluster.local:9090",
		},
		{
			name:                "falls back when production namespace is empty",
			productionNamespace: "",
			fallbackTarget:      "codex-k8s-control-plane:9090",
			want:                "codex-k8s-control-plane:9090",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := resolveRunControlPlaneGRPCTarget(tc.productionNamespace, tc.fallbackTarget)
			if got != tc.want {
				t.Fatalf("resolveRunControlPlaneGRPCTarget() = %q, want %q", got, tc.want)
			}
		})
	}
}
