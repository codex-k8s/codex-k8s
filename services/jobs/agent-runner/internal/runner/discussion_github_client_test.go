package runner

import "testing"

func TestSplitRepositoryFullName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		owner, repository, err := splitRepositoryFullName(" codex-k8s/codex-k8s ")
		if err != nil {
			t.Fatalf("splitRepositoryFullName returned error: %v", err)
		}
		if owner != "codex-k8s" || repository != "codex-k8s" {
			t.Fatalf("unexpected repository parts: owner=%q repo=%q", owner, repository)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if _, _, err := splitRepositoryFullName("codex-k8s"); err == nil {
			t.Fatal("expected splitRepositoryFullName to fail for invalid repository name")
		}
	})
}
