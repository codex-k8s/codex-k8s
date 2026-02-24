package runner

import "testing"

func TestBuildTargetBranch_AIRepairUsesBaseBranch(t *testing.T) {
	t.Parallel()

	got := buildTargetBranch("", "run-123", 45, "ai_repair", "main")
	if want := "main"; got != want {
		t.Fatalf("buildTargetBranch() = %q, want %q", got, want)
	}
}

func TestBuildTargetBranch_ExplicitBranchHasPriority(t *testing.T) {
	t.Parallel()

	got := buildTargetBranch("hotfix/incident-1", "run-123", 45, "ai_repair", "main")
	if want := "hotfix/incident-1"; got != want {
		t.Fatalf("buildTargetBranch() = %q, want %q", got, want)
	}
}
