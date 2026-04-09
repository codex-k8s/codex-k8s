package runner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	floweventdomain "github.com/codex-k8s/kodex/libs/go/domain/flowevent"
	cpclient "github.com/codex-k8s/kodex/services/jobs/agent-runner/internal/controlplane"
)

func TestRestoreLatestSession_AllowsInteractionResumeWithoutPR(t *testing.T) {
	t.Parallel()

	service := NewService(Config{
		RunID:                    "run-resume",
		RepositoryFullName:       "codex-k8s/kodex",
		AgentKey:                 "dev",
		InteractionResumePayload: `{"interaction_id":"interaction-1"}`,
	}, &fakeSessionRestoreControlPlane{
		snapshot: cpclient.AgentSessionSnapshot{
			RunID:            "run-source",
			SessionID:        "sess-resume",
			CodexSessionJSON: json.RawMessage(`{"session_id":"sess-resume","cwd":"/workspace"}`),
		},
		found: true,
	}, nil)

	restored, err := service.restoreLatestSession(context.Background(), "codex/issue-394", t.TempDir())
	if err != nil {
		t.Fatalf("restoreLatestSession() error = %v", err)
	}
	if restored.prNotFound {
		t.Fatal("expected interaction resume flow to proceed without PR precondition failure")
	}
	if got, want := restored.sessionID, "sess-resume"; got != want {
		t.Fatalf("sessionID = %q, want %q", got, want)
	}
	if restored.restoredSessionPath == "" {
		t.Fatal("expected restored session path")
	}
	restoredJSON, err := os.ReadFile(restored.restoredSessionPath)
	if err != nil {
		t.Fatalf("read restored session file: %v", err)
	}
	if got, want := string(restoredJSON), `{"session_id":"sess-resume","cwd":"/workspace"}`; got != want {
		t.Fatalf("restored session file = %q, want %q", got, want)
	}
}

func TestResolveInteractionResumePayload_SkipsLookupForPlainRun(t *testing.T) {
	t.Parallel()

	service := NewService(Config{
		RunID:         "run-plain",
		CorrelationID: "corr-plain",
	}, &fakeSessionRestoreControlPlane{
		resumeErr: assertResumeLookupShouldNotRunError{},
	}, nil)

	payload, err := service.resolveInteractionResumePayload(context.Background())
	if err != nil {
		t.Fatalf("resolveInteractionResumePayload() error = %v", err)
	}
	if payload != "" {
		t.Fatalf("payload = %q, want empty", payload)
	}
}

func TestResolveInteractionResumePayload_FetchesCurrentRunPayload(t *testing.T) {
	t.Parallel()

	service := NewService(Config{
		RunID:         "run-resume",
		CorrelationID: "interaction-resume:interaction-1",
	}, &fakeSessionRestoreControlPlane{
		resumePayload: cpclient.RunInteractionResumePayload{
			Payload: json.RawMessage(`{"interaction_id":"interaction-1","tool_name":"user.decision.request"}`),
		},
		resumeFound: true,
	}, nil)

	payload, err := service.resolveInteractionResumePayload(context.Background())
	if err != nil {
		t.Fatalf("resolveInteractionResumePayload() error = %v", err)
	}
	if got, want := payload, `{"interaction_id":"interaction-1","tool_name":"user.decision.request"}`; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}
}

func TestResolveInteractionResumePayload_RequiresPersistedPayloadForResumeRun(t *testing.T) {
	t.Parallel()

	service := NewService(Config{
		RunID:         "run-resume",
		CorrelationID: "interaction-resume:interaction-1",
	}, &fakeSessionRestoreControlPlane{}, nil)

	_, err := service.resolveInteractionResumePayload(context.Background())
	if err == nil {
		t.Fatal("expected missing persisted resume payload to fail interaction resume run")
	}
}

func TestResolveGitHubRateLimitResumePayload_FetchesCurrentRunPayload(t *testing.T) {
	t.Parallel()

	service := NewService(Config{
		RunID:         "run-resume",
		CorrelationID: "github-rate-limit-resume:wait-1",
	}, &fakeSessionRestoreControlPlane{
		githubResumePayload: cpclient.RunGitHubRateLimitResumePayload{
			Payload: json.RawMessage(`{"wait_id":"wait-1","wait_reason":"github_rate_limit","contour_kind":"agent_bot_token","limit_kind":"secondary","resolution_kind":"auto_resumed","recovered_at":"2026-03-14T17:00:00Z","attempt_no":2,"affected_operation_class":"agent_github_call","guidance":"reuse the restored snapshot"}`),
		},
		githubResumeFound: true,
	}, nil)

	payload, err := service.resolveGitHubRateLimitResumePayload(context.Background())
	if err != nil {
		t.Fatalf("resolveGitHubRateLimitResumePayload() error = %v", err)
	}
	if got, want := payload, `{"wait_id":"wait-1","wait_reason":"github_rate_limit","contour_kind":"agent_bot_token","limit_kind":"secondary","resolution_kind":"auto_resumed","recovered_at":"2026-03-14T17:00:00Z","attempt_no":2,"affected_operation_class":"agent_github_call","guidance":"reuse the restored snapshot"}`; got != want {
		t.Fatalf("payload = %q, want %q", got, want)
	}
}

func TestResolveGitHubRateLimitResumePayload_RequiresPersistedPayloadForResumeRun(t *testing.T) {
	t.Parallel()

	service := NewService(Config{
		RunID:         "run-resume",
		CorrelationID: "github-rate-limit-resume:wait-1",
	}, &fakeSessionRestoreControlPlane{}, nil)

	_, err := service.resolveGitHubRateLimitResumePayload(context.Background())
	if err == nil {
		t.Fatal("expected missing persisted github rate-limit resume payload to fail resume run")
	}
}

func TestRestoreLatestSession_WithoutPRAndWithoutInteractionResumeMarksPRNotFound(t *testing.T) {
	t.Parallel()

	service := NewService(Config{
		RunID:              "run-revise",
		RepositoryFullName: "codex-k8s/kodex",
		AgentKey:           "dev",
	}, &fakeSessionRestoreControlPlane{
		snapshot: cpclient.AgentSessionSnapshot{
			RunID:            "run-source",
			SessionID:        "sess-revise",
			CodexSessionJSON: json.RawMessage(`{"session_id":"sess-revise"}`),
		},
		found: true,
	}, nil)

	restored, err := service.restoreLatestSession(context.Background(), "codex/issue-394", t.TempDir())
	if err != nil {
		t.Fatalf("restoreLatestSession() error = %v", err)
	}
	if !restored.prNotFound {
		t.Fatal("expected missing PR to remain a precondition failure outside interaction resume or discussion flows")
	}
	if restored.restoredSessionPath != "" {
		t.Fatalf("expected no restored session path, got %q", restored.restoredSessionPath)
	}
}

func TestRestoreLatestSession_AllowsGitHubRateLimitResumeWithoutPR(t *testing.T) {
	t.Parallel()

	service := NewService(Config{
		RunID:                        "run-github-resume",
		RepositoryFullName:           "codex-k8s/kodex",
		AgentKey:                     "dev",
		GitHubRateLimitResumePayload: `{"wait_id":"wait-1"}`,
	}, &fakeSessionRestoreControlPlane{
		snapshot: cpclient.AgentSessionSnapshot{
			RunID:            "run-source",
			SessionID:        "sess-rate-limit",
			CodexSessionJSON: json.RawMessage(`{"session_id":"sess-rate-limit","cwd":"/workspace"}`),
		},
		found: true,
	}, nil)

	restored, err := service.restoreLatestSession(context.Background(), "codex/issue-394", t.TempDir())
	if err != nil {
		t.Fatalf("restoreLatestSession() error = %v", err)
	}
	if restored.prNotFound {
		t.Fatal("expected github rate-limit resume flow to proceed without PR precondition failure")
	}
	if got, want := restored.sessionID, "sess-rate-limit"; got != want {
		t.Fatalf("sessionID = %q, want %q", got, want)
	}
}

type fakeSessionRestoreControlPlane struct {
	snapshot            cpclient.AgentSessionSnapshot
	found               bool
	err                 error
	resumePayload       cpclient.RunInteractionResumePayload
	resumeFound         bool
	resumeErr           error
	githubResumePayload cpclient.RunGitHubRateLimitResumePayload
	githubResumeFound   bool
	githubResumeErr     error
}

type assertResumeLookupShouldNotRunError struct{}

func (assertResumeLookupShouldNotRunError) Error() string {
	return "resume lookup should not have been called"
}

func (f *fakeSessionRestoreControlPlane) UpsertAgentSession(context.Context, cpclient.AgentSessionUpsertParams) (cpclient.AgentSessionUpsertResult, error) {
	return cpclient.AgentSessionUpsertResult{}, nil
}

func (f *fakeSessionRestoreControlPlane) GetLatestAgentSession(context.Context, cpclient.LatestAgentSessionQuery) (cpclient.AgentSessionSnapshot, bool, error) {
	if f.err != nil {
		return cpclient.AgentSessionSnapshot{}, false, f.err
	}
	return f.snapshot, f.found, nil
}

func (f *fakeSessionRestoreControlPlane) GetRunInteractionResumePayload(context.Context) (cpclient.RunInteractionResumePayload, bool, error) {
	if f.resumeErr != nil {
		return cpclient.RunInteractionResumePayload{}, false, f.resumeErr
	}
	return f.resumePayload, f.resumeFound, nil
}

func (f *fakeSessionRestoreControlPlane) GetRunGitHubRateLimitResumePayload(context.Context) (cpclient.RunGitHubRateLimitResumePayload, bool, error) {
	if f.githubResumeErr != nil {
		return cpclient.RunGitHubRateLimitResumePayload{}, false, f.githubResumeErr
	}
	return f.githubResumePayload, f.githubResumeFound, nil
}

func (f *fakeSessionRestoreControlPlane) ReportGitHubRateLimitSignal(context.Context, cpclient.ReportGitHubRateLimitSignalParams) (cpclient.ReportGitHubRateLimitSignalResult, error) {
	return cpclient.ReportGitHubRateLimitSignalResult{}, nil
}

func (f *fakeSessionRestoreControlPlane) LookupRunPullRequest(context.Context, cpclient.RunPullRequestLookupParams) (cpclient.RunPullRequestLookupResult, bool, error) {
	return cpclient.RunPullRequestLookupResult{}, false, nil
}

func (f *fakeSessionRestoreControlPlane) InsertRunFlowEvent(context.Context, string, floweventdomain.EventType, json.RawMessage) error {
	return nil
}

func (f *fakeSessionRestoreControlPlane) GetCodexAuth(context.Context) ([]byte, bool, error) {
	return nil, false, nil
}

func (f *fakeSessionRestoreControlPlane) UpsertCodexAuth(context.Context, []byte) error {
	return nil
}

func (f *fakeSessionRestoreControlPlane) UpsertRunStatusComment(context.Context, cpclient.UpsertRunStatusCommentParams) error {
	return nil
}

func TestRestoreLatestSession_WritesSnapshotWithRunScopedFilename(t *testing.T) {
	t.Parallel()

	sessionsDir := t.TempDir()
	service := NewService(Config{
		RunID:                    "run-snapshot",
		RepositoryFullName:       "codex-k8s/kodex",
		AgentKey:                 "dev",
		InteractionResumePayload: `{"interaction_id":"interaction-1"}`,
	}, &fakeSessionRestoreControlPlane{
		snapshot: cpclient.AgentSessionSnapshot{
			CodexSessionJSON: json.RawMessage(`{"session_id":"sess-snapshot"}`),
		},
		found: true,
	}, nil)

	restored, err := service.restoreLatestSession(context.Background(), "codex/issue-394", sessionsDir)
	if err != nil {
		t.Fatalf("restoreLatestSession() error = %v", err)
	}
	if got, want := restored.restoredSessionPath, filepath.Join(sessionsDir, "restored-run-snapshot.json"); got != want {
		t.Fatalf("restoredSessionPath = %q, want %q", got, want)
	}
}
