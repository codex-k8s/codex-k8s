package webhook

import (
	"context"
	"encoding/json"
	"testing"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
)

func TestExtractPullRequestHeadRefFromNormalizedRunPayload(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"pull_request":{"head":{"ref":"feature/from-normalized"}}}`)
	if got, want := extractPullRequestHeadRefFromNormalizedRunPayload(payload), "feature/from-normalized"; got != want {
		t.Fatalf("extractPullRequestHeadRefFromNormalizedRunPayload() = %q, want %q", got, want)
	}
}

func TestExtractPullRequestHeadRefFromNormalizedRunPayload_FallbackToRawPayload(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"raw_payload":{"pull_request":{"head":{"ref":"feature/from-raw"}}}}`)
	if got, want := extractPullRequestHeadRefFromNormalizedRunPayload(payload), "feature/from-raw"; got != want {
		t.Fatalf("extractPullRequestHeadRefFromNormalizedRunPayload() = %q, want %q", got, want)
	}
}

func TestResolveRuntimeBuildRefForIssueTrigger_UsesRunHistoryPullRequestRef(t *testing.T) {
	t.Parallel()

	runs := &inMemoryRunRepo{
		byRunID: map[string]agentrunrepo.Run{
			"run-100": {
				ID:         "run-100",
				RunPayload: json.RawMessage(`{"pull_request":{"head":{"ref":"feature/pr-100"}}}`),
			},
		},
		searchItems: []agentrunrepo.RunLookupItem{
			{
				RunID:              "run-100",
				ProjectID:          "project-1",
				RepositoryFullName: "codex-k8s/codex-k8s",
				IssueNumber:        205,
				PullRequestNumber:  100,
			},
		},
	}
	svc := &Service{agentRuns: runs}

	envelope := githubWebhookEnvelope{
		Repository: githubRepositoryRecord{FullName: "codex-k8s/codex-k8s"},
		Issue:      githubIssueRecord{Number: 205},
	}
	got := svc.resolveRuntimeBuildRefForIssueTrigger(context.Background(), "project-1", envelope, "main", agentdomain.RuntimeModeFullEnv)
	if got != "feature/pr-100" {
		t.Fatalf("resolveRuntimeBuildRefForIssueTrigger() = %q, want %q", got, "feature/pr-100")
	}
}

func TestResolveRuntimeBuildRefForIssueTrigger_CodeOnlyKeepsDefault(t *testing.T) {
	t.Parallel()

	svc := &Service{}
	envelope := githubWebhookEnvelope{
		Repository: githubRepositoryRecord{FullName: "codex-k8s/codex-k8s"},
		Issue:      githubIssueRecord{Number: 205},
	}
	got, want := svc.resolveRuntimeBuildRefForIssueTrigger(context.Background(), "project-1", envelope, "main", agentdomain.RuntimeModeCodeOnly), "main"
	if got != want {
		t.Fatalf("resolveRuntimeBuildRefForIssueTrigger() = %q, want %q", got, want)
	}
}
