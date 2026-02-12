package runner

import (
	"encoding/json"
	"testing"
)

func TestBuildSessionLogJSON(t *testing.T) {
	logs := buildSessionLogJSON(runResult{
		targetBranch:     "codex/issue-13",
		existingPRNumber: 200,
		report: codexReport{
			Summary:         "done",
			PRNumber:        200,
			PRURL:           "https://example/pull/200",
			Model:           "gpt-5.2-codex",
			ReasoningEffort: "high",
		},
		codexExecOutput: "stdout",
		gitPushOutput:   "push ok",
	}, runStatusSucceeded)

	if !json.Valid(logs) {
		t.Fatal("expected valid json output")
	}

	var parsed sessionLogSnapshot
	if err := json.Unmarshal(logs, &parsed); err != nil {
		t.Fatalf("unmarshal session logs: %v", err)
	}
	if parsed.Status != runStatusSucceeded {
		t.Fatalf("expected status %q, got %q", runStatusSucceeded, parsed.Status)
	}
	if parsed.Runtime.TargetBranch != "codex/issue-13" {
		t.Fatalf("unexpected target branch %q", parsed.Runtime.TargetBranch)
	}
	if parsed.Report.PRNumber != 200 {
		t.Fatalf("expected report pr number 200, got %d", parsed.Report.PRNumber)
	}
}
