package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsurePreparedFullEnvBranchAttachesDetachedHead(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repoDir := initGitRepoForPreparedBranchTests(t)

	if err := runCommandQuiet(ctx, repoDir, "git", "checkout", "--detach"); err != nil {
		t.Fatalf("detach HEAD: %v", err)
	}

	if err := ensurePreparedFullEnvBranch(ctx, repoDir, "codex/issue-355"); err != nil {
		t.Fatalf("ensurePreparedFullEnvBranch(detached) error = %v", err)
	}

	currentBranch, err := currentLocalBranch(ctx, repoDir)
	if err != nil {
		t.Fatalf("currentLocalBranch() error = %v", err)
	}
	if got, want := currentBranch, "codex/issue-355"; got != want {
		t.Fatalf("currentLocalBranch() = %q, want %q", got, want)
	}
}

func TestEnsurePreparedFullEnvBranchRejectsUnexpectedLiveBranch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repoDir := initGitRepoForPreparedBranchTests(t)

	if err := ensurePreparedFullEnvBranch(ctx, repoDir, "codex/issue-355"); err == nil {
		t.Fatal("ensurePreparedFullEnvBranch(unexpected branch) expected error")
	} else if !strings.Contains(err.Error(), "branch mismatch") {
		t.Fatalf("ensurePreparedFullEnvBranch(unexpected branch) error = %v, want branch mismatch", err)
	}
}

func initGitRepoForPreparedBranchTests(t *testing.T) string {
	t.Helper()

	ctx := context.Background()
	repoDir := t.TempDir()
	if err := runCommandQuiet(ctx, repoDir, "git", "init", "-b", "main"); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := runCommandQuiet(ctx, repoDir, "git", "config", "user.name", "Test Runner"); err != nil {
		t.Fatalf("git config user.name: %v", err)
	}
	if err := runCommandQuiet(ctx, repoDir, "git", "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config user.email: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("test\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := runCommandQuiet(ctx, repoDir, "git", "add", "README.md"); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := runCommandQuiet(ctx, repoDir, "git", "commit", "-m", "test"); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	return repoDir
}
