package worker

import (
	"encoding/json"
	"testing"
)

func TestHasIssueLabelInRunPayload(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{"raw_payload":{"issue":{"labels":[{"name":"run:debug"},{"name":"[ai-model-gpt-5.2-codex]"}]}}}`)

	if !hasIssueLabelInRunPayload(runPayload, "run:debug") {
		t.Fatal("expected to find run:debug label")
	}
	if !hasIssueLabelInRunPayload(runPayload, "[ai-model-gpt-5.2-codex]") {
		t.Fatal("expected to find bracketed ai-model label")
	}
	if hasIssueLabelInRunPayload(runPayload, "run:dev") {
		t.Fatal("did not expect to find run:dev label")
	}
}

func TestExtractIssueLabelsFromRunPayload_InvalidPayload(t *testing.T) {
	t.Parallel()

	labels := extractIssueLabelsFromRunPayload(json.RawMessage(`{"raw_payload":"invalid"}`))
	if labels != nil {
		t.Fatalf("expected nil labels for invalid payload, got %#v", labels)
	}
}

func TestExtractIssueLabelsFromRunPayload_PullRequestLabels(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{
		"raw_payload":{
			"pull_request":{
				"labels":[
					{"name":"run:dev:revise"},
					{"name":"[ai-model-gpt-5.2-codex]"}
				]
			}
		}
	}`)

	if !hasIssueLabelInRunPayload(runPayload, "run:dev:revise") {
		t.Fatal("expected to find run:dev:revise label in pull_request.labels")
	}
	if !hasIssueLabelInRunPayload(runPayload, "[ai-model-gpt-5.2-codex]") {
		t.Fatal("expected to find ai-model label in pull_request.labels")
	}
}

func TestExtractIssueAndPullRequestLabels(t *testing.T) {
	t.Parallel()

	rawPayload := json.RawMessage(`{
		"issue":{"labels":[{"name":"issue-label-1"},{"name":"issue-label-2"}]},
		"pull_request":{"labels":[{"name":"pr-label-1"}]}
	}`)

	issueLabels, pullRequestLabels := extractIssueAndPullRequestLabels(rawPayload)
	if len(issueLabels) != 2 {
		t.Fatalf("expected 2 issue labels, got %d", len(issueLabels))
	}
	if len(pullRequestLabels) != 1 {
		t.Fatalf("expected 1 pull request label, got %d", len(pullRequestLabels))
	}
	if issueLabels[0] != "issue-label-1" || issueLabels[1] != "issue-label-2" {
		t.Fatalf("unexpected issue labels: %#v", issueLabels)
	}
	if pullRequestLabels[0] != "pr-label-1" {
		t.Fatalf("unexpected pull request labels: %#v", pullRequestLabels)
	}
}
