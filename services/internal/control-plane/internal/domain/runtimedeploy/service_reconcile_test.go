package runtimedeploy

import (
	"testing"
	"time"
)

func TestRuntimeDeployLeaseTiming(t *testing.T) {
	t.Parallel()

	if got := runtimeDeployLeaseRenewInterval(10 * time.Minute); got != 30*time.Second {
		t.Fatalf("runtimeDeployLeaseRenewInterval(10m) = %s, want %s", got, 30*time.Second)
	}
	if got := runtimeDeployLeaseRenewInterval(20 * time.Second); got != 10*time.Second {
		t.Fatalf("runtimeDeployLeaseRenewInterval(20s) = %s, want %s", got, 10*time.Second)
	}
	if got := runtimeDeployLeaseRenewInterval(500 * time.Millisecond); got != time.Second {
		t.Fatalf("runtimeDeployLeaseRenewInterval(500ms) = %s, want %s", got, time.Second)
	}

	if got := runtimeDeployStaleRunningTimeout(30 * time.Second); got != 65*time.Second {
		t.Fatalf("runtimeDeployStaleRunningTimeout(30s) = %s, want %s", got, 65*time.Second)
	}
	if got := runtimeDeployStaleRunningTimeout(10 * time.Second); got != 30*time.Second {
		t.Fatalf("runtimeDeployStaleRunningTimeout(10s) = %s, want %s", got, 30*time.Second)
	}
	if got := runtimeDeployStaleRunningTimeout(2 * time.Minute); got != 2*time.Minute {
		t.Fatalf("runtimeDeployStaleRunningTimeout(2m) = %s, want %s", got, 2*time.Minute)
	}
}

func TestRepositoryRootForRuntimeEnv_PrefersConfiguredRoot(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			RepositoryRoot: "/repo-cache",
		},
	}
	if got := svc.repositoryRootForRuntimeEnv("/repo-cache/github/codex-k8s/codex-k8s/main"); got != "/repo-cache" {
		t.Fatalf("repositoryRootForRuntimeEnv() = %q, want %q", got, "/repo-cache")
	}
}

func TestRepositoryRootForRuntimeEnv_FallsBackToResolvedWhenConfigEmpty(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{},
	}
	if got := svc.repositoryRootForRuntimeEnv("/repo-cache/github/codex-k8s/codex-k8s/main"); got != "/repo-cache" {
		t.Fatalf("repositoryRootForRuntimeEnv() = %q, want %q", got, "/repo-cache")
	}
}

func TestNormalizeRepositoryCacheRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "snapshot path collapses to cache root",
			in:   "/repo-cache/github/codex-k8s/codex-k8s/0acb7d5",
			want: "/repo-cache",
		},
		{
			name: "cache root stays unchanged",
			in:   "/repo-cache",
			want: "/repo-cache",
		},
		{
			name: "non github path stays unchanged",
			in:   "/var/lib/codex",
			want: "/var/lib/codex",
		},
		{
			name: "empty path stays empty",
			in:   "",
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeRepositoryCacheRoot(tt.in); got != tt.want {
				t.Fatalf("normalizeRepositoryCacheRoot(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
