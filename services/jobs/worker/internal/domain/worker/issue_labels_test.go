package worker

import (
	"encoding/json"
	"testing"
)

func TestHasIssueLabelInRunPayload(t *testing.T) {
	t.Parallel()

	runPayload := json.RawMessage(`{"raw_payload":{"issue":{"labels":[{"name":"run:debug"},{"name":"[ai-model-gpt-5.3-codex]"}]}}}`)

	if !hasIssueLabelInRunPayload(runPayload, "run:debug") {
		t.Fatal("expected to find run:debug label")
	}
	if !hasIssueLabelInRunPayload(runPayload, "[ai-model-gpt-5.3-codex]") {
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
