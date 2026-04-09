package worker

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

func TestTickReconcilesGitHubRateLimitWaitsBeforeMissionControl(t *testing.T) {
	t.Parallel()

	processor := &fakeGitHubRateLimitWaitProcessor{
		results: []GitHubRateLimitProcessResult{
			{WaitID: "wait-1", RunID: "run-1", State: "resolved", ResolutionKind: "auto_resumed", AttemptNo: 1},
		},
		found: []bool{true, false},
	}
	missionControl := &fakeMissionControlClient{
		warmupProjects: []MissionControlWarmupProject{
			{ProjectID: "project-1", ProjectName: "Project 1", RepositoryFullName: "codex-k8s/kodex"},
		},
	}

	svc := NewService(Config{
		WorkerID:                           "worker-1",
		GitHubRateLimitWaitEnabledFallback: true,
		GitHubRateLimitSweepLimit:          5,
	}, Dependencies{
		Runs:             &fakeRunQueue{},
		Events:           &fakeFlowEvents{},
		Launcher:         &fakeLauncher{states: map[string]JobState{}},
		GitHubRateLimits: processor,
		MissionControl:   missionControl,
		Logger:           slog.New(slog.NewJSONHandler(io.Discard, nil)),
	})

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() error = %v", err)
	}
	if got, want := processor.calls, 2; got != want {
		t.Fatalf("github rate-limit sweep calls = %d, want %d", got, want)
	}
	if len(missionControl.runWarmupCalls) != 1 {
		t.Fatalf("expected mission control reconciliation after github rate-limit sweep")
	}
}

type fakeGitHubRateLimitWaitProcessor struct {
	results []GitHubRateLimitProcessResult
	found   []bool
	err     error
	calls   int
}

func (f *fakeGitHubRateLimitWaitProcessor) ProcessNextGitHubRateLimitWait(_ context.Context, _ string) (GitHubRateLimitProcessResult, bool, error) {
	if f.err != nil {
		return GitHubRateLimitProcessResult{}, false, f.err
	}
	index := f.calls
	f.calls++
	if index >= len(f.found) || !f.found[index] {
		return GitHubRateLimitProcessResult{}, false, nil
	}
	if index >= len(f.results) {
		return GitHubRateLimitProcessResult{}, true, nil
	}
	return f.results[index], true, nil
}
