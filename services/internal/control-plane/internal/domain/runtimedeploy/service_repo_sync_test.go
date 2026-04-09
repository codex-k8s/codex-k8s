package runtimedeploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldUseDirectRepositoryRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "deploy", "base"), 0o755); err != nil {
		t.Fatalf("create deploy/base: %v", err)
	}

	if got := shouldUseDirectRepositoryRoot(root, ""); !got {
		t.Fatalf("shouldUseDirectRepositoryRoot(root, \"\") = false, want true")
	}
	if got := shouldUseDirectRepositoryRoot(root, "codex-k8s/kodex"); got {
		t.Fatalf("shouldUseDirectRepositoryRoot(root, repository) = true, want false")
	}
}

func TestShouldUseDirectRepositoryRoot_FalseForNonRepositoryPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if got := shouldUseDirectRepositoryRoot(root, ""); got {
		t.Fatalf("shouldUseDirectRepositoryRoot(non-repo-root, \"\") = true, want false")
	}
}

func TestResolveSourceRepoSyncNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		targetNamespace string
		vars            map[string]string
		want            string
	}{
		{
			name:            "prefers platform namespace for full env source snapshot",
			targetNamespace: "codex-issue-3278207d1cd3-i473-r4a5958f0757f",
			vars: map[string]string{
				"KODEX_PLATFORM_NAMESPACE":   "kodex-prod",
				"KODEX_PRODUCTION_NAMESPACE": "codex-issue-3278207d1cd3-i473-r4a5958f0757f",
			},
			want: "kodex-prod",
		},
		{
			name:            "falls back to production namespace when platform missing",
			targetNamespace: "codex-issue-3278207d1cd3-i473-r4a5958f0757f",
			vars: map[string]string{
				"KODEX_PRODUCTION_NAMESPACE": "kodex-prod",
			},
			want: "kodex-prod",
		},
		{
			name:            "falls back to target namespace when no platform vars available",
			targetNamespace: "kodex-dev-1",
			vars:            map[string]string{},
			want:            "kodex-dev-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := resolveSourceRepoSyncNamespace(tt.targetNamespace, tt.vars); got != tt.want {
				t.Fatalf("resolveSourceRepoSyncNamespace() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShouldSyncRepoSnapshotToRuntimeNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		configuredRoot string
		targetEnv      string
		namespace      string
		repository     string
		hotReload      string
		want           bool
	}{
		{
			name:           "ai env always syncs with absolute root",
			configuredRoot: "/repo-cache",
			targetEnv:      "ai",
			namespace:      "kodex-dev-1",
			repository:     "codex-k8s/kodex",
			hotReload:      "false",
			want:           true,
		},
		{
			name:           "non ai with hot reload enabled syncs",
			configuredRoot: "/repo-cache",
			targetEnv:      "staging",
			namespace:      "staging-ns",
			repository:     "codex-k8s/kodex",
			hotReload:      "true",
			want:           true,
		},
		{
			name:           "relative root does not sync",
			configuredRoot: ".",
			targetEnv:      "ai",
			namespace:      "kodex-dev-1",
			repository:     "codex-k8s/kodex",
			hotReload:      "true",
			want:           false,
		},
		{
			name:           "missing namespace does not sync",
			configuredRoot: "/repo-cache",
			targetEnv:      "ai",
			namespace:      "",
			repository:     "codex-k8s/kodex",
			hotReload:      "true",
			want:           false,
		},
		{
			name:           "missing repository does not sync",
			configuredRoot: "/repo-cache",
			targetEnv:      "ai",
			namespace:      "kodex-dev-1",
			repository:     "",
			hotReload:      "true",
			want:           false,
		},
		{
			name:           "non ai without hot reload does not sync",
			configuredRoot: "/repo-cache",
			targetEnv:      "production",
			namespace:      "kodex-prod",
			repository:     "codex-k8s/kodex",
			hotReload:      "false",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := shouldSyncRepoSnapshotToRuntimeNamespace(tt.configuredRoot, tt.targetEnv, tt.namespace, tt.repository, tt.hotReload)
			if got != tt.want {
				t.Fatalf("shouldSyncRepoSnapshotToRuntimeNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepoSnapshotPath_UsesBuildRefScopedSnapshotForAIEnv(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			RepositoryRoot: "/repo-cache",
		},
	}

	got := svc.repoSnapshotPath("ai", "codex-k8s", "kodex", "3e7c26a0470fbab877607df5f82f5874a8119b5f")
	want := "/repo-cache/github/codex-k8s/kodex/3e7c26a0470fbab877607df5f82f5874a8119b5f"
	if got != want {
		t.Fatalf("repoSnapshotPath(ai) = %q, want %q", got, want)
	}
}

func TestRepoSnapshotPath_FallsBackToMainWhenBuildRefIsEmpty(t *testing.T) {
	t.Parallel()

	svc := &Service{
		cfg: Config{
			RepositoryRoot: "/repo-cache",
		},
	}

	got := svc.repoSnapshotPath("ai", "codex-k8s", "kodex", "")
	want := "/repo-cache/github/codex-k8s/kodex/main"
	if got != want {
		t.Fatalf("repoSnapshotPath(empty build ref) = %q, want %q", got, want)
	}
}
