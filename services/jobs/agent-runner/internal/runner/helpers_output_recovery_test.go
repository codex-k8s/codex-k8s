package runner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	cpclient "github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/controlplane"
)

func TestResolveCodexReport_RecoversPRMetadataFromControlPlane(t *testing.T) {
	service := NewService(Config{
		RunID:                  "run-1",
		ProjectID:              "project-1",
		RepositoryFullName:     "codex-k8s/codex-k8s",
		PromptConfig:           PromptConfig{PromptTemplateLocale: promptLocaleRU},
		ExistingPRNumber:       354,
		RunTargetBranch:        "codex/issue-347",
		IssueNumber:            347,
		AgentKey:               "dev",
		ControlPlaneGRPCTarget: "codex-k8s-control-plane:9090",
	}, &fakeOutputRecoveryControlPlane{
		lookupResult: cpclient.RunPullRequestLookupResult{
			PRNumber: 354,
			PRURL:    "https://example.test/pr/354",
			PRState:  "open",
		},
		lookupFound: true,
	}, nil)

	state := codexState{repoDir: t.TempDir()}
	result := runResult{
		targetBranch:     "codex/issue-347",
		existingPRNumber: 354,
		sessionID:        "sess-1",
	}

	report, repairedOutput, err := service.resolveCodexReport(context.Background(), state, &result, time.Now().UTC(), filepath.Join(t.TempDir(), "schema.json"), []byte(`{"status":"ok","summary":["done"]}`), true)
	if err != nil {
		t.Fatalf("resolveCodexReport() error = %v", err)
	}
	if len(repairedOutput) != 0 {
		t.Fatalf("expected no repaired output, got %q", string(repairedOutput))
	}
	if report.PRNumber != 354 {
		t.Fatalf("pr_number = %d, want 354", report.PRNumber)
	}
	if report.PRURL != "https://example.test/pr/354" {
		t.Fatalf("pr_url = %q", report.PRURL)
	}
	if report.Summary != "done" {
		t.Fatalf("summary = %q, want done", report.Summary)
	}
}

func TestResolveCodexReport_RepairsStructuredOutputWhenRequiredFieldsMissing(t *testing.T) {
	t.Setenv("PATH", buildFakeOutputRecoveryPath(t, `{"summary":"repaired","branch":"codex/issue-347","pr_number":354,"pr_url":"https://example.test/pr/354","session_id":"sess-1","model":"gpt-5.4","reasoning_effort":"high"}`))

	service := NewService(Config{
		RunID:              "run-2",
		ProjectID:          "project-1",
		RepositoryFullName: "codex-k8s/codex-k8s",
		PromptConfig:       PromptConfig{PromptTemplateLocale: promptLocaleRU},
		ExistingPRNumber:   354,
		RunTargetBranch:    "codex/issue-347",
		IssueNumber:        347,
		AgentKey:           "dev",
	}, &fakeOutputRecoveryControlPlane{}, nil)

	state := codexState{repoDir: t.TempDir()}
	result := runResult{
		targetBranch:     "codex/issue-347",
		existingPRNumber: 0,
		sessionID:        "sess-1",
	}

	report, repairedOutput, err := service.resolveCodexReport(context.Background(), state, &result, time.Now().UTC(), filepath.Join(t.TempDir(), "schema.json"), []byte(`{"status":"ok","summary":["done"]}`), true)
	if err != nil {
		t.Fatalf("resolveCodexReport() error = %v", err)
	}
	if strings.TrimSpace(string(repairedOutput)) == "" {
		t.Fatalf("expected repaired output to be present")
	}
	if report.PRNumber != 354 {
		t.Fatalf("pr_number = %d, want 354", report.PRNumber)
	}
	if report.PRURL != "https://example.test/pr/354" {
		t.Fatalf("pr_url = %q", report.PRURL)
	}
	if report.Summary != "repaired" {
		t.Fatalf("summary = %q, want repaired", report.Summary)
	}
}

type fakeOutputRecoveryControlPlane struct {
	lookupParams  []cpclient.RunPullRequestLookupParams
	lookupResult  cpclient.RunPullRequestLookupResult
	lookupFound   bool
	lookupErr     error
	sessionResult cpclient.AgentSessionUpsertResult
}

func (f *fakeOutputRecoveryControlPlane) UpsertAgentSession(context.Context, cpclient.AgentSessionUpsertParams) (cpclient.AgentSessionUpsertResult, error) {
	return f.sessionResult, nil
}

func (f *fakeOutputRecoveryControlPlane) GetLatestAgentSession(context.Context, cpclient.LatestAgentSessionQuery) (cpclient.AgentSessionSnapshot, bool, error) {
	return cpclient.AgentSessionSnapshot{}, false, nil
}

func (f *fakeOutputRecoveryControlPlane) GetRunInteractionResumePayload(context.Context) (cpclient.RunInteractionResumePayload, bool, error) {
	return cpclient.RunInteractionResumePayload{}, false, nil
}

func (f *fakeOutputRecoveryControlPlane) GetRunGitHubRateLimitResumePayload(context.Context) (cpclient.RunGitHubRateLimitResumePayload, bool, error) {
	return cpclient.RunGitHubRateLimitResumePayload{}, false, nil
}

func (f *fakeOutputRecoveryControlPlane) ReportGitHubRateLimitSignal(context.Context, cpclient.ReportGitHubRateLimitSignalParams) (cpclient.ReportGitHubRateLimitSignalResult, error) {
	return cpclient.ReportGitHubRateLimitSignalResult{}, nil
}

func (f *fakeOutputRecoveryControlPlane) LookupRunPullRequest(_ context.Context, params cpclient.RunPullRequestLookupParams) (cpclient.RunPullRequestLookupResult, bool, error) {
	f.lookupParams = append(f.lookupParams, params)
	return f.lookupResult, f.lookupFound, f.lookupErr
}

func (f *fakeOutputRecoveryControlPlane) InsertRunFlowEvent(context.Context, string, floweventdomain.EventType, json.RawMessage) error {
	return nil
}

func (f *fakeOutputRecoveryControlPlane) GetCodexAuth(context.Context) ([]byte, bool, error) {
	return nil, false, nil
}

func (f *fakeOutputRecoveryControlPlane) UpsertCodexAuth(context.Context, []byte) error {
	return nil
}

func (f *fakeOutputRecoveryControlPlane) UpsertRunStatusComment(context.Context, cpclient.UpsertRunStatusCommentParams) error {
	return nil
}

func buildFakeOutputRecoveryPath(t *testing.T, repairOutput string) string {
	t.Helper()

	binDir := t.TempDir()
	codexScriptPath := filepath.Join(binDir, "codex")

	codexScript := `#!/usr/bin/env bash
set -euo pipefail
output_file=""
prev=""
for arg in "$@"; do
  if [[ "$prev" == "--output-last-message" ]]; then
    output_file="$arg"
  fi
  prev="$arg"
done
if [[ -n "$output_file" ]]; then
  printf '%s' "$FAKE_CODEX_REPAIR_OUTPUT" > "$output_file"
fi
printf 'stdout-fallback'
`
	if err := os.WriteFile(codexScriptPath, []byte(codexScript), 0o755); err != nil {
		t.Fatalf("write fake codex script: %v", err)
	}

	t.Setenv("FAKE_CODEX_REPAIR_OUTPUT", repairOutput)

	return binDir + string(os.PathListSeparator) + os.Getenv("PATH")
}
