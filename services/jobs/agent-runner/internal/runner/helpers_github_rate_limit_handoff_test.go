package runner

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	sharedgithubratelimit "github.com/codex-k8s/codex-k8s/libs/go/domain/githubratelimit"
	cpclient "github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/controlplane"
)

func TestDetectGitHubRateLimitSignal_ParsesHeaders(t *testing.T) {
	t.Parallel()

	candidate, ok := detectGitHubRateLimitSignal(nil, strings.Join([]string{
		"HTTP 403 Forbidden",
		"API rate limit exceeded for token",
		"x-ratelimit-remaining: 0",
		"x-ratelimit-reset: 1773507900",
		"x-ratelimit-resource: core",
		"retry-after: 60",
		"x-github-request-id: ABCD:1234",
		"https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting",
	}, "\n"), assertGitHubRateLimitExecError{})
	if !ok {
		t.Fatal("expected github rate-limit detection")
	}
	if candidate.ProviderStatusCode != 403 {
		t.Fatalf("provider status code = %d, want 403", candidate.ProviderStatusCode)
	}
	if candidate.Headers.RateLimitRemaining == nil || *candidate.Headers.RateLimitRemaining != 0 {
		t.Fatalf("rate_limit_remaining = %v, want 0", candidate.Headers.RateLimitRemaining)
	}
	if candidate.Headers.RetryAfterSeconds == nil || *candidate.Headers.RetryAfterSeconds != 60 {
		t.Fatalf("retry_after_seconds = %v, want 60", candidate.Headers.RetryAfterSeconds)
	}
	if candidate.Headers.GitHubRequestID != "ABCD:1234" {
		t.Fatalf("github_request_id = %q", candidate.Headers.GitHubRequestID)
	}
	if !strings.Contains(candidate.Headers.DocumentationURL, "docs.github.com") {
		t.Fatalf("documentation_url = %q", candidate.Headers.DocumentationURL)
	}
	if !strings.HasPrefix(candidate.SignalID, "ghrlsig-") {
		t.Fatalf("signal_id = %q", candidate.SignalID)
	}
}

func TestRunCodexExecWithAuthRecovery_GitHubRateLimitAcceptedPersistsWaitingSnapshot(t *testing.T) {
	t.Setenv("PATH", buildFakeGitHubRateLimitCodexPath(t))

	sessionsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sessionsDir, "session.json"), []byte(`{"session_id":"sess-rate-limit","cwd":"/workspace"}`), 0o600); err != nil {
		t.Fatalf("write fake session file: %v", err)
	}

	controlPlane := &fakeGitHubRateLimitControlPlane{
		reportResult: cpclient.ReportGitHubRateLimitSignalResult{
			WaitID:       "wait-1",
			WaitState:    runStatusWaitingBackpressure,
			WaitReason:   githubRateLimitWaitReason,
			NextStepKind: "auto_resume_scheduled",
			RunnerAction: sharedgithubratelimit.RunnerActionPersistSessionAndExitWait,
		},
	}
	service := NewService(Config{
		RunID:              "run-428",
		CorrelationID:      "corr-428",
		RepositoryFullName: "codex-k8s/codex-k8s",
		AgentKey:           "dev",
	}, controlPlane, nil)

	result := runResult{targetBranch: "codex/issue-428", triggerKind: "dev", templateKind: promptTemplateKindWork}
	_, err := service.runCodexExecWithAuthRecovery(
		context.Background(),
		codexState{repoDir: t.TempDir(), sessionsDir: sessionsDir},
		&result,
		time.Date(2026, time.March, 14, 17, 0, 0, 0, time.UTC),
		codexExecParams{RepoDir: t.TempDir(), Prompt: "prompt"},
	)
	if err == nil {
		t.Fatal("expected github rate-limit wait handoff error")
	}
	if !isGitHubRateLimitWaitAccepted(err) {
		t.Fatalf("expected github rate-limit wait acceptance, got %v", err)
	}

	if got, want := controlPlane.persistedStatuses, []string{runStatusRunning, runStatusWaitingBackpressure}; !equalStringSlices(got, want) {
		t.Fatalf("persisted statuses = %v, want %v", got, want)
	}
	if controlPlane.reportParams.ProviderStatusCode != 403 {
		t.Fatalf("provider_status_code = %d, want 403", controlPlane.reportParams.ProviderStatusCode)
	}
	if got := controlPlane.reportParams.SignalOrigin; got != "agent_runner" {
		t.Fatalf("signal_origin = %q, want agent_runner", got)
	}
	if got := controlPlane.reportParams.OperationClass; got != githubRateLimitOperationAgentGitHubCall {
		t.Fatalf("operation_class = %q, want %q", got, githubRateLimitOperationAgentGitHubCall)
	}
	if controlPlane.reportParams.SessionSnapshotVersion == nil || *controlPlane.reportParams.SessionSnapshotVersion != 1 {
		t.Fatalf("session_snapshot_version = %v, want 1", controlPlane.reportParams.SessionSnapshotVersion)
	}
	if got := result.sessionID; got != "sess-rate-limit" {
		t.Fatalf("result.sessionID = %q, want sess-rate-limit", got)
	}
}

func TestRunCodexExecWithAuthRecovery_GitHubRateLimitAcceptedWhenWaitingSnapshotPersistFails(t *testing.T) {
	t.Setenv("PATH", buildFakeGitHubRateLimitCodexPath(t))

	sessionsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sessionsDir, "session.json"), []byte(`{"session_id":"sess-rate-limit","cwd":"/workspace"}`), 0o600); err != nil {
		t.Fatalf("write fake session file: %v", err)
	}

	controlPlane := &fakeGitHubRateLimitControlPlane{
		reportResult: cpclient.ReportGitHubRateLimitSignalResult{
			WaitID:       "wait-2",
			WaitState:    runStatusWaitingBackpressure,
			WaitReason:   githubRateLimitWaitReason,
			NextStepKind: "auto_resume_scheduled",
			RunnerAction: sharedgithubratelimit.RunnerActionPersistSessionAndExitWait,
		},
		failPersistAt: 2,
	}
	service := NewService(Config{
		RunID:              "run-428",
		CorrelationID:      "corr-428",
		RepositoryFullName: "codex-k8s/codex-k8s",
		AgentKey:           "dev",
	}, controlPlane, nil)

	result := runResult{targetBranch: "codex/issue-428", triggerKind: "dev", templateKind: promptTemplateKindWork}
	_, err := service.runCodexExecWithAuthRecovery(
		context.Background(),
		codexState{repoDir: t.TempDir(), sessionsDir: sessionsDir},
		&result,
		time.Date(2026, time.March, 14, 17, 0, 0, 0, time.UTC),
		codexExecParams{RepoDir: t.TempDir(), Prompt: "prompt"},
	)
	if err == nil {
		t.Fatal("expected github rate-limit wait handoff error")
	}
	if !isGitHubRateLimitWaitAccepted(err) {
		t.Fatalf("expected github rate-limit wait acceptance, got %v", err)
	}

	var waitErr githubRateLimitWaitAcceptedError
	if !errors.As(err, &waitErr) {
		t.Fatalf("expected githubRateLimitWaitAcceptedError, got %T", err)
	}
	if waitErr.PersistErr == nil {
		t.Fatal("expected persist error to be attached to wait acceptance")
	}
	if !strings.Contains(waitErr.PersistErr.Error(), "waiting_backpressure") {
		t.Fatalf("persist error = %v, want waiting_backpressure context", waitErr.PersistErr)
	}
	if got, want := controlPlane.persistedStatuses, []string{runStatusRunning, runStatusWaitingBackpressure}; !equalStringSlices(got, want) {
		t.Fatalf("persisted statuses = %v, want %v", got, want)
	}
	if got := result.snapshotVersion; got != 1 {
		t.Fatalf("snapshot_version = %d, want 1 after failed waiting snapshot persist", got)
	}
}

type assertGitHubRateLimitExecError struct{}

func (assertGitHubRateLimitExecError) Error() string {
	return "exit status 1"
}

type fakeGitHubRateLimitControlPlane struct {
	persistedStatuses []string
	reportParams      cpclient.ReportGitHubRateLimitSignalParams
	reportResult      cpclient.ReportGitHubRateLimitSignalResult
	failPersistAt     int
}

func (f *fakeGitHubRateLimitControlPlane) UpsertAgentSession(_ context.Context, params cpclient.AgentSessionUpsertParams) (cpclient.AgentSessionUpsertResult, error) {
	f.persistedStatuses = append(f.persistedStatuses, params.Runtime.Status)
	if f.failPersistAt > 0 && len(f.persistedStatuses) == f.failPersistAt {
		return cpclient.AgentSessionUpsertResult{}, errors.New("synthetic persist failure")
	}
	return cpclient.AgentSessionUpsertResult{
		SnapshotVersion:  int64(len(f.persistedStatuses)),
		SnapshotChecksum: "checksum",
	}, nil
}

func (f *fakeGitHubRateLimitControlPlane) GetLatestAgentSession(context.Context, cpclient.LatestAgentSessionQuery) (cpclient.AgentSessionSnapshot, bool, error) {
	return cpclient.AgentSessionSnapshot{}, false, nil
}

func (f *fakeGitHubRateLimitControlPlane) GetRunInteractionResumePayload(context.Context) (cpclient.RunInteractionResumePayload, bool, error) {
	return cpclient.RunInteractionResumePayload{}, false, nil
}

func (f *fakeGitHubRateLimitControlPlane) GetRunGitHubRateLimitResumePayload(context.Context) (cpclient.RunGitHubRateLimitResumePayload, bool, error) {
	return cpclient.RunGitHubRateLimitResumePayload{}, false, nil
}

func (f *fakeGitHubRateLimitControlPlane) ReportGitHubRateLimitSignal(_ context.Context, params cpclient.ReportGitHubRateLimitSignalParams) (cpclient.ReportGitHubRateLimitSignalResult, error) {
	f.reportParams = params
	return f.reportResult, nil
}

func (f *fakeGitHubRateLimitControlPlane) LookupRunPullRequest(context.Context, cpclient.RunPullRequestLookupParams) (cpclient.RunPullRequestLookupResult, bool, error) {
	return cpclient.RunPullRequestLookupResult{}, false, nil
}

func (f *fakeGitHubRateLimitControlPlane) InsertRunFlowEvent(context.Context, string, floweventdomain.EventType, json.RawMessage) error {
	return nil
}

func (f *fakeGitHubRateLimitControlPlane) GetCodexAuth(context.Context) ([]byte, bool, error) {
	return nil, false, nil
}

func (f *fakeGitHubRateLimitControlPlane) UpsertCodexAuth(context.Context, []byte) error {
	return nil
}

func (f *fakeGitHubRateLimitControlPlane) UpsertRunStatusComment(context.Context, cpclient.UpsertRunStatusCommentParams) error {
	return nil
}

func buildFakeGitHubRateLimitCodexPath(t *testing.T) string {
	t.Helper()

	binDir := t.TempDir()
	scriptPath := filepath.Join(binDir, "codex")
	script := `#!/usr/bin/env bash
set -euo pipefail
cat >&2 <<'EOF'
HTTP 403 Forbidden
API rate limit exceeded for token
x-ratelimit-remaining: 0
x-ratelimit-reset: 1773507900
x-ratelimit-resource: core
retry-after: 60
x-github-request-id: ABCD:1234
https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting
EOF
exit 1
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake codex script: %v", err)
	}
	return binDir + string(os.PathListSeparator) + os.Getenv("PATH")
}

func equalStringSlices(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
