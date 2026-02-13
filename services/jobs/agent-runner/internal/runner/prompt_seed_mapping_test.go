package runner

import "testing"

func TestPromptSeedStageByTriggerKind(t *testing.T) {
	testCases := []struct {
		name      string
		trigger   string
		wantStage string
	}{
		{name: "dev", trigger: "dev", wantStage: "dev"},
		{name: "dev revise", trigger: "dev_revise", wantStage: "dev"},
		{name: "intake", trigger: "intake", wantStage: "intake"},
		{name: "intake revise", trigger: "intake_revise", wantStage: "intake"},
		{name: "vision", trigger: "vision", wantStage: "vision"},
		{name: "prd", trigger: "prd", wantStage: "prd"},
		{name: "arch", trigger: "arch", wantStage: "arch"},
		{name: "design", trigger: "design", wantStage: "design"},
		{name: "plan", trigger: "plan", wantStage: "plan"},
		{name: "doc audit", trigger: "doc_audit", wantStage: "doc-audit"},
		{name: "qa", trigger: "qa", wantStage: "qa"},
		{name: "release", trigger: "release", wantStage: "release"},
		{name: "postdeploy", trigger: "postdeploy", wantStage: "postdeploy"},
		{name: "ops", trigger: "ops", wantStage: "ops"},
		{name: "self improve", trigger: "self_improve", wantStage: "self-improve"},
		{name: "abort", trigger: "abort", wantStage: "abort"},
		{name: "rethink", trigger: "rethink", wantStage: "rethink"},
		{name: "unknown fallback", trigger: "unknown", wantStage: "dev"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := promptSeedStageByTriggerKind(tc.trigger); got != tc.wantStage {
				t.Fatalf("unexpected stage: got %q, want %q", got, tc.wantStage)
			}
		})
	}
}

func TestPromptSeedCandidates(t *testing.T) {
	candidates := promptSeedCandidates("design_revise", "review", "ru")
	if len(candidates) < 2 {
		t.Fatalf("expected at least two candidates, got %d", len(candidates))
	}
	if candidates[0] != "design-review_ru.md" {
		t.Fatalf("unexpected first candidate: %q", candidates[0])
	}
	if candidates[1] != "design-review.md" {
		t.Fatalf("unexpected second candidate: %q", candidates[1])
	}

	fallback := promptSeedCandidates("nonexistent", "work", "ru")
	if len(fallback) < 2 {
		t.Fatalf("expected fallback candidates, got %d", len(fallback))
	}
	if fallback[0] != "dev-work_ru.md" {
		t.Fatalf("unexpected fallback first candidate: %q", fallback[0])
	}
	if fallback[1] != "dev-work.md" {
		t.Fatalf("unexpected fallback second candidate: %q", fallback[1])
	}
}
