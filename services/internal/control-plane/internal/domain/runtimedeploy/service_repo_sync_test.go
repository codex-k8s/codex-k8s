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
