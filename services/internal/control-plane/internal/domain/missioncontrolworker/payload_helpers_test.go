package missioncontrolworker

import (
	"encoding/json"
	"testing"
)

func TestDecodeProjectionRunContext(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage(`{
		"trigger":{"label":"run:dev:revise"},
		"runtime":{"mode":"full-env"},
		"issue":{"title":"Mission issue","state":"open","user":{"login":"normalized-owner"}},
		"pull_request":{"title":"Mission PR","state":"open","head":{"ref":"feature/from-normalized"},"base":{"ref":"main"},"user":{"login":"normalized-pr-owner"}},
		"raw_payload":{
			"issue":{"labels":[{"name":"bug"},{"name":"run:qa"}],"user":{"login":"raw-owner"}},
			"pull_request":{"labels":[{"name":"state:in-review"}],"head":{"ref":"feature/from-raw"},"base":{"ref":"release"},"user":{"login":"raw-pr-owner"}}
		}
	}`)

	ctx := decodeProjectionRunContext(raw)
	if got, want := ctx.TriggerLabel, "run:dev:revise"; got != want {
		t.Fatalf("trigger label = %q, want %q", got, want)
	}
	if got, want := ctx.RuntimeMode, "full-env"; got != want {
		t.Fatalf("runtime mode = %q, want %q", got, want)
	}
	if got, want := ctx.IssueTitle, "Mission issue"; got != want {
		t.Fatalf("issue title = %q, want %q", got, want)
	}
	if got, want := ctx.IssueOwner, "normalized-owner"; got != want {
		t.Fatalf("issue owner = %q, want %q", got, want)
	}
	if got, want := len(ctx.IssueLabels), 2; got != want {
		t.Fatalf("issue labels len = %d, want %d", got, want)
	}
	if got, want := ctx.PullRequestHead, "feature/from-normalized"; got != want {
		t.Fatalf("pull request head = %q, want %q", got, want)
	}
	if got, want := ctx.PullRequestBase, "main"; got != want {
		t.Fatalf("pull request base = %q, want %q", got, want)
	}
	if got, want := resolveProjectionStageLabel(ctx.IssueLabels, ctx.PullRequestLabels, ctx.TriggerLabel), "run:qa"; got != want {
		t.Fatalf("resolved stage label = %q, want %q", got, want)
	}
}

func TestResolveProjectionStageLabelFallsBackToTrigger(t *testing.T) {
	t.Parallel()

	if got, want := resolveProjectionStageLabel(nil, nil, "run:dev"), "run:dev"; got != want {
		t.Fatalf("resolveProjectionStageLabel() = %q, want %q", got, want)
	}
}
