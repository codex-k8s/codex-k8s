package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	runstatusdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runstatus"
)

func TestExtractPullRequestHeadBuildRefFromNormalizedRunPayload(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"pull_request":{"head":{"ref":"feature/from-normalized"}}}`)
	if got, want := extractPullRequestHeadBuildRefFromNormalizedRunPayload(payload), "feature/from-normalized"; got != want {
		t.Fatalf("extractPullRequestHeadBuildRefFromNormalizedRunPayload() = %q, want %q", got, want)
	}
}

func TestExtractPullRequestHeadBuildRefFromNormalizedRunPayload_PrefersSHA(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"pull_request":{"head":{"ref":"feature/from-normalized","sha":"0123456789abcdef0123456789abcdef01234567"}}}`)
	if got, want := extractPullRequestHeadBuildRefFromNormalizedRunPayload(payload), "0123456789abcdef0123456789abcdef01234567"; got != want {
		t.Fatalf("extractPullRequestHeadBuildRefFromNormalizedRunPayload() = %q, want %q", got, want)
	}
}

func TestExtractPullRequestHeadBuildRefFromNormalizedRunPayload_FallbackToRawPayload(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"raw_payload":{"pull_request":{"head":{"ref":"feature/from-raw"}}}}`)
	if got, want := extractPullRequestHeadBuildRefFromNormalizedRunPayload(payload), "feature/from-raw"; got != want {
		t.Fatalf("extractPullRequestHeadBuildRefFromNormalizedRunPayload() = %q, want %q", got, want)
	}
}

func TestResolveIssueTriggerRuntimeProfile_QAUsesCandidateNamespaceAndCurrentPRHead(t *testing.T) {
	t.Parallel()

	runs := &inMemoryRunRepo{
		byRunID: map[string]agentrunrepo.Run{
			"run-100": {
				ID:         "run-100",
				RunPayload: json.RawMessage(`{"runtime":{"namespace":"codex-issue-205"},"pull_request":{"head":{"ref":"feature/pr-100"}}}`),
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
	svc := &Service{
		agentRuns:   runs,
		githubToken: "token",
		githubMgmt: &inMemoryPushMainVersionBumpClient{
			prHeads: map[string]GitHubPullRequestHeadDetails{
				"codex-k8s/codex-k8s#100": {
					State:   "open",
					HeadRef: "feature/pr-100",
					HeadSHA: "0123456789abcdef0123456789abcdef01234567",
				},
			},
		},
		runStatus: &inMemoryRunStatusService{
			runtimeStates: map[string]runstatusdomain.RuntimeState{
				"run-100": {Namespace: "codex-issue-205"},
			},
		},
	}

	envelope := githubWebhookEnvelope{
		Repository: githubRepositoryRecord{FullName: "codex-k8s/codex-k8s"},
		Issue:      githubIssueRecord{Number: 205},
	}
	profile := svc.resolveIssueTriggerRuntimeProfile(
		context.Background(),
		"project-1",
		envelope,
		issueRunTrigger{Kind: webhookdomain.TriggerKindQA},
		"main",
		agentdomain.RuntimeModeFullEnv,
	)
	if got, want := profile.TargetEnv, "ai"; got != want {
		t.Fatalf("target env = %q, want %q", got, want)
	}
	if got, want := profile.Namespace, "codex-issue-205"; got != want {
		t.Fatalf("namespace = %q, want %q", got, want)
	}
	if got, want := profile.BuildRef, "0123456789abcdef0123456789abcdef01234567"; got != want {
		t.Fatalf("build ref = %q, want %q", got, want)
	}
	if got, want := profile.AccessProfile, agentdomain.RuntimeAccessProfileCandidate; got != want {
		t.Fatalf("access profile = %q, want %q", got, want)
	}
}

func TestResolveIssueTriggerRuntimeProfile_CodeOnlyKeepsDefaultRef(t *testing.T) {
	t.Parallel()

	svc := &Service{}
	envelope := githubWebhookEnvelope{
		Repository: githubRepositoryRecord{FullName: "codex-k8s/codex-k8s"},
		Issue:      githubIssueRecord{Number: 205},
	}
	profile := svc.resolveIssueTriggerRuntimeProfile(
		context.Background(),
		"project-1",
		envelope,
		issueRunTrigger{Kind: webhookdomain.TriggerKindQA},
		"main",
		agentdomain.RuntimeModeCodeOnly,
	)
	if got, want := profile.BuildRef, "main"; got != want {
		t.Fatalf("build ref = %q, want %q", got, want)
	}
}

func TestResolveIssueTriggerRuntimeProfile_ReleaseWithoutCandidateReturnsWarning(t *testing.T) {
	t.Parallel()

	svc := &Service{
		agentRuns: &inMemoryRunRepo{},
	}
	envelope := githubWebhookEnvelope{
		Repository: githubRepositoryRecord{FullName: "codex-k8s/codex-k8s"},
		Issue:      githubIssueRecord{Number: 240},
	}
	profile := svc.resolveIssueTriggerRuntimeProfile(
		context.Background(),
		"project-1",
		envelope,
		issueRunTrigger{Kind: webhookdomain.TriggerKindRelease},
		"main",
		agentdomain.RuntimeModeFullEnv,
	)
	if got, want := profile.WarningReason, string(runstatusdomain.TriggerWarningReasonIssueTriggerCandidateNotFound); got != want {
		t.Fatalf("warning reason = %q, want %q", got, want)
	}
	if len(profile.SuggestedLabels) != 2 {
		t.Fatalf("expected 2 suggested labels, got %d", len(profile.SuggestedLabels))
	}
}

func TestResolveIssueTriggerRuntimeProfile_QAWithPullRequestLookupErrorReturnsWarning(t *testing.T) {
	t.Parallel()

	runs := &inMemoryRunRepo{
		byRunID: map[string]agentrunrepo.Run{
			"run-lookup-error": {
				ID:         "run-lookup-error",
				RunPayload: json.RawMessage(`{"runtime":{"namespace":"codex-issue-205"},"pull_request":{"head":{"ref":"feature/stale","sha":"stale-sha"}}}`),
			},
		},
		searchItems: []agentrunrepo.RunLookupItem{
			{
				RunID:              "run-lookup-error",
				ProjectID:          "project-1",
				RepositoryFullName: "codex-k8s/codex-k8s",
				IssueNumber:        205,
				PullRequestNumber:  100,
			},
		},
	}
	svc := &Service{
		agentRuns:   runs,
		githubToken: "token",
		githubMgmt: &inMemoryPushMainVersionBumpClient{
			prHeadErr: errors.New("github unavailable"),
		},
		runStatus: &inMemoryRunStatusService{
			runtimeStates: map[string]runstatusdomain.RuntimeState{
				"run-lookup-error": {Namespace: "codex-issue-205"},
			},
		},
	}

	envelope := githubWebhookEnvelope{
		Repository: githubRepositoryRecord{FullName: "codex-k8s/codex-k8s"},
		Issue:      githubIssueRecord{Number: 205},
	}
	profile := svc.resolveIssueTriggerRuntimeProfile(
		context.Background(),
		"project-1",
		envelope,
		issueRunTrigger{Kind: webhookdomain.TriggerKindQA},
		"main",
		agentdomain.RuntimeModeFullEnv,
	)
	if got, want := profile.WarningReason, string(runstatusdomain.TriggerWarningReasonIssueTriggerCandidateNotFound); got != want {
		t.Fatalf("warning reason = %q, want %q", got, want)
	}
	if profile.Namespace != "" {
		t.Fatalf("expected empty namespace when PR head lookup fails, got %q", profile.Namespace)
	}
	if got, want := profile.BuildRef, "main"; got != want {
		t.Fatalf("build ref = %q, want %q", got, want)
	}
}

func TestResolveIssueTriggerRuntimeProfile_PostdeployUsesProductionReadOnly(t *testing.T) {
	t.Parallel()

	svc := &Service{
		platformNamespace: "codex-k8s-prod",
		githubToken:       "token",
		githubMgmt: &inMemoryPushMainVersionBumpClient{
			refToSHA: map[string]string{
				"codex-k8s/codex-k8s@main": "89abcdef0123456789abcdef0123456789abcdef",
			},
		},
	}
	envelope := githubWebhookEnvelope{
		Repository: githubRepositoryRecord{FullName: "codex-k8s/codex-k8s"},
		Issue:      githubIssueRecord{Number: 241},
	}
	profile := svc.resolveIssueTriggerRuntimeProfile(
		context.Background(),
		"project-1",
		envelope,
		issueRunTrigger{Kind: webhookdomain.TriggerKindPostDeploy},
		"main",
		agentdomain.RuntimeModeFullEnv,
	)
	if got, want := profile.TargetEnv, "production"; got != want {
		t.Fatalf("target env = %q, want %q", got, want)
	}
	if got, want := profile.Namespace, "codex-k8s-prod"; got != want {
		t.Fatalf("namespace = %q, want %q", got, want)
	}
	if got, want := profile.BuildRef, "89abcdef0123456789abcdef0123456789abcdef"; got != want {
		t.Fatalf("build ref = %q, want %q", got, want)
	}
	if got, want := profile.AccessProfile, agentdomain.RuntimeAccessProfileProductionReadOnly; got != want {
		t.Fatalf("access profile = %q, want %q", got, want)
	}
}
