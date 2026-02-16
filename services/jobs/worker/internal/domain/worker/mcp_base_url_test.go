package worker

import "testing"

func TestResolveControlPlaneMCPBaseURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		explicitURL string
		grpcTarget  string
		want        string
	}{
		{
			name:        "uses explicit URL as-is",
			explicitURL: "http://custom-mcp:8081/mcp",
			grpcTarget:  "codex-k8s-control-plane.ns.svc.cluster.local:9090",
			want:        "http://custom-mcp:8081/mcp",
		},
		{
			name:       "derives from grpc host and port",
			grpcTarget: "codex-k8s-control-plane.codex-k8s-prod.svc.cluster.local:9090",
			want:       "http://codex-k8s-control-plane.codex-k8s-prod.svc.cluster.local:8081/mcp",
		},
		{
			name:       "falls back to default when grpc target is empty",
			grpcTarget: "",
			want:       defaultControlPlaneMCPBaseURL,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := resolveControlPlaneMCPBaseURL(tc.explicitURL, tc.grpcTarget)
			if got != tc.want {
				t.Fatalf("resolveControlPlaneMCPBaseURL() = %q, want %q", got, tc.want)
			}
		})
	}
}
