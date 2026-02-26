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
	if got := shouldUseDirectRepositoryRoot(root, "codex-k8s/codex-k8s"); got {
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
			namespace:      "codex-k8s-dev-1",
			repository:     "codex-k8s/codex-k8s",
			hotReload:      "false",
			want:           true,
		},
		{
			name:           "non ai with hot reload enabled syncs",
			configuredRoot: "/repo-cache",
			targetEnv:      "staging",
			namespace:      "staging-ns",
			repository:     "codex-k8s/codex-k8s",
			hotReload:      "true",
			want:           true,
		},
		{
			name:           "relative root does not sync",
			configuredRoot: ".",
			targetEnv:      "ai",
			namespace:      "codex-k8s-dev-1",
			repository:     "codex-k8s/codex-k8s",
			hotReload:      "true",
			want:           false,
		},
		{
			name:           "missing namespace does not sync",
			configuredRoot: "/repo-cache",
			targetEnv:      "ai",
			namespace:      "",
			repository:     "codex-k8s/codex-k8s",
			hotReload:      "true",
			want:           false,
		},
		{
			name:           "missing repository does not sync",
			configuredRoot: "/repo-cache",
			targetEnv:      "ai",
			namespace:      "codex-k8s-dev-1",
			repository:     "",
			hotReload:      "true",
			want:           false,
		},
		{
			name:           "non ai without hot reload does not sync",
			configuredRoot: "/repo-cache",
			targetEnv:      "production",
			namespace:      "codex-k8s-prod",
			repository:     "codex-k8s/codex-k8s",
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
