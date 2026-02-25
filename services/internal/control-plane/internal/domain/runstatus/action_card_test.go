package runstatus

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestExtractThreadLabelsFromRunPayload_PrefersPullRequestLabelsForPullRequestTarget(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{
		"raw_payload": {
			"issue": {"labels":[{"name":"run:plan"}]},
			"pull_request": {"labels":[{"name":"run:dev:revise"},{"name":"risk:rbac"}]}
		},
		"trigger": {"label":"run:plan:revise"},
		"profile_hints": {"last_run_issue_labels":["cross-service"]}
	}`)

	labels := extractThreadLabelsFromRunPayload(runPayload, commentTargetKindPullRequest)
	if !slices.Equal(labels, []string{"risk:rbac", "run:dev:revise"}) {
		t.Fatalf("unexpected pull request labels: %#v", labels)
	}
}

func TestExtractThreadLabelsFromRunPayload_FallsBackToTriggerAndProfileHintsWhenRawPayloadHasNoLabels(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{
		"raw_payload": {
			"pull_request": {"number": 203}
		},
		"trigger": {"label":"run:plan:revise"},
		"profile_hints": {
			"last_run_issue_labels":["run:plan","cross-service"],
			"last_run_pull_request_labels":["run:plan:revise","risk:rbac"]
		}
	}`)

	labels := extractThreadLabelsFromRunPayload(runPayload, commentTargetKindPullRequest)
	if !slices.Contains(labels, "run:plan:revise") {
		t.Fatalf("expected fallback to include trigger stage label, got %#v", labels)
	}
	if slices.Contains(labels, "run:plan") {
		t.Fatalf("expected fallback to keep a single stage label from trigger context, got %#v", labels)
	}
	if !slices.Contains(labels, "cross-service") || !slices.Contains(labels, "risk:rbac") {
		t.Fatalf("expected fallback to preserve non-stage profile hints, got %#v", labels)
	}
	stageLabels := collectStageLabels(labels)
	if !slices.Equal(stageLabels, []string{"run:plan:revise"}) {
		t.Fatalf("expected one resolved stage label from trigger fallback, got %#v", stageLabels)
	}
}

func TestExtractThreadLabelsFromRunPayload_UsesProfileHintsForAmbiguityWhenTriggerLabelMissing(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{
		"raw_payload": {"pull_request": {"number": 204}},
		"profile_hints": {
			"last_run_issue_labels": ["run:plan","run:dev"],
			"last_run_pull_request_labels": ["risk:rbac"]
		}
	}`)

	labels := extractThreadLabelsFromRunPayload(runPayload, commentTargetKindPullRequest)
	stageLabels := collectStageLabels(labels)
	if !slices.Equal(stageLabels, []string{"run:dev", "run:plan"}) {
		t.Fatalf("expected ambiguous stage labels from profile hints, got %#v", stageLabels)
	}
}
