package runqueue

import (
	"testing"

	"github.com/google/uuid"

	querytypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/query"
)

func TestParseRunQueuePayload_ReturnsDeployOnlyRuntime(t *testing.T) {
	payload := parseRunQueuePayload([]byte(`{"repository":{"full_name":"Codex-K8S/Repo"},"runtime":{"deploy_only":true}}`))

	if got, want := payload.Repository.FullName, "Codex-K8S/Repo"; got != want {
		t.Fatalf("Repository.FullName mismatch: got %q want %q", got, want)
	}
	if payload.Runtime == nil {
		t.Fatal("expected runtime payload")
	}
	if !payload.Runtime.DeployOnly {
		t.Fatal("expected deploy_only=true")
	}
}

func TestParseRunQueuePayload_InvalidJSON(t *testing.T) {
	payload := parseRunQueuePayload([]byte(`{"repository":`))
	if payload.Repository.FullName != "" {
		t.Fatalf("expected empty repository full_name for invalid json, got %q", payload.Repository.FullName)
	}
	if payload.Runtime != nil {
		t.Fatal("expected nil runtime for invalid json")
	}
}

func TestIsDeployOnlyRun(t *testing.T) {
	tests := []struct {
		name    string
		payload querytypes.RunQueuePayload
		want    bool
	}{
		{
			name: "runtime missing",
			payload: querytypes.RunQueuePayload{
				Repository: querytypes.RepositoryPayload{FullName: "codex-k8s/repo"},
			},
			want: false,
		},
		{
			name: "deploy_only false",
			payload: querytypes.RunQueuePayload{
				Runtime: &querytypes.RunRuntimeProfile{DeployOnly: false},
			},
			want: false,
		},
		{
			name: "deploy_only true",
			payload: querytypes.RunQueuePayload{
				Runtime: &querytypes.RunRuntimeProfile{DeployOnly: true},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDeployOnlyRun(tt.payload); got != tt.want {
				t.Fatalf("isDeployOnlyRun()=%v want %v", got, tt.want)
			}
		})
	}
}

func TestDeriveProjectID(t *testing.T) {
	t.Run("from repository full_name", func(t *testing.T) {
		payload := querytypes.RunQueuePayload{
			Repository: querytypes.RepositoryPayload{FullName: "Codex-K8S/Repo"},
		}
		got := deriveProjectID("corr-1", payload)
		want := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("repo:codex-k8s/repo")).String()
		if got != want {
			t.Fatalf("deriveProjectID()=%q want %q", got, want)
		}
	})

	t.Run("fallback to correlation", func(t *testing.T) {
		got := deriveProjectID("corr-2", querytypes.RunQueuePayload{})
		want := uuid.NewSHA1(uuid.NameSpaceDNS, []byte("correlation:corr-2")).String()
		if got != want {
			t.Fatalf("deriveProjectID()=%q want %q", got, want)
		}
	})
}
