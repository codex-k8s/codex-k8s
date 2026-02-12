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
