package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	agentrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agent"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	projectmemberrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectmember"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
	runstatusdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runstatus"
)

func TestIngestGitHubWebhook_Dedup(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{
		byLogin: map[string]userrepo.User{
			"member": {
				ID:          "user-1",
				GitHubLogin: "member",
			},
		},
	}
	members := &inMemoryProjectMemberRepo{
		roles: map[string]string{
			"project-1|user-1": "read_write",
		},
	}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      users,
		Members:    members,
	})

	payload := json.RawMessage(`{
		"action":"labeled",
		"label":{"name":"run:dev"},
		"issue":{"id":1001,"number":77,"title":"Implement feature","html_url":"https://github.com/codex-k8s/codex-k8s/issues/77","state":"open"},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-1",
		DeliveryID:    "delivery-1",
		EventType:     string(webhookdomain.GitHubEventIssues),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	first, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("first ingest failed: %v", err)
	}
	if first.Duplicate {
		t.Fatalf("expected first event to be accepted")
	}

	second, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("second ingest failed: %v", err)
	}
	if !second.Duplicate {
		t.Fatalf("expected duplicate event on second delivery")
	}

	if len(events.items) != 2 {
		t.Fatalf("expected 2 flow events, got %d", len(events.items))
	}
	if events.items[0].EventType != floweventdomain.EventTypeWebhookReceived {
		t.Fatalf("expected first event webhook.received, got %s", events.items[0].EventType)
	}
	if events.items[1].EventType != floweventdomain.EventTypeWebhookDuplicate {
		t.Fatalf("expected second event webhook.duplicate, got %s", events.items[1].EventType)
	}
}

func TestIngestGitHubWebhook_NonTriggerEventsDoNotCreateRun(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
	})

	payload := json.RawMessage(`{
		"action":"created",
		"issue":{"id":1001,"number":77,"title":"Implement feature","html_url":"https://github.com/codex-k8s/codex-k8s/issues/77","state":"open"},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-nt-1",
		DeliveryID:    "delivery-nt-1",
		EventType:     "issue_comment",
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != webhookdomain.IngestStatusAccepted || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.RunID != "" {
		t.Fatalf("expected no run for non-trigger event, got run id %q", got.RunID)
	}
	if len(runs.items) != 0 {
		t.Fatalf("expected no run records for non-trigger event, got %d", len(runs.items))
	}
	if len(events.items) != 1 {
		t.Fatalf("expected 1 flow event, got %d", len(events.items))
	}
	if events.items[0].EventType != floweventdomain.EventTypeWebhookReceived {
		t.Fatalf("expected webhook.received event, got %s", events.items[0].EventType)
	}
}

func TestIngestGitHubWebhook_PushMain_CreatesDeployOnlyProductionRun(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	svc := NewService(Config{
		AgentRuns:  runs,
		FlowEvents: events,
		Repos:      repos,
	})

	buildRef := "0123456789abcdef0123456789abcdef01234567"
	payload := json.RawMessage(fmt.Sprintf(`{
		"ref":"refs/heads/main",
		"before":"0000000000000000000000000000000000000000",
		"after":"%s",
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`, buildRef))
	cmd := IngestCommand{
		CorrelationID: "delivery-push-main-1",
		DeliveryID:    "delivery-push-main-1",
		EventType:     string(webhookdomain.GitHubEventPush),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != webhookdomain.IngestStatusAccepted || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.RunID == "" {
		t.Fatal("expected run id for push main deploy-only trigger")
	}

	var runPayload githubRunPayload
	if err := json.Unmarshal(runs.last.RunPayload, &runPayload); err != nil {
		t.Fatalf("unmarshal run payload: %v", err)
	}
	if runPayload.Trigger != nil {
		t.Fatalf("expected trigger to be nil for push deploy-only run, got %#v", runPayload.Trigger)
	}
	if got, want := runPayload.Runtime.Mode, "full-env"; got != want {
		t.Fatalf("unexpected runtime mode: got %q want %q", got, want)
	}
	if got, want := runPayload.Runtime.Source, runtimeModeSourcePushMain; got != want {
		t.Fatalf("unexpected runtime source: got %q want %q", got, want)
	}
	if got, want := runPayload.Runtime.TargetEnv, "production"; got != want {
		t.Fatalf("unexpected runtime target env: got %q want %q", got, want)
	}
	if runPayload.Runtime.Namespace != "" {
		t.Fatalf("unexpected runtime namespace: got %q want empty (resolved via services.yaml)", runPayload.Runtime.Namespace)
	}
	if got, want := runPayload.Runtime.BuildRef, buildRef; got != want {
		t.Fatalf("unexpected runtime build ref: got %q want %q", got, want)
	}
	if !runPayload.Runtime.DeployOnly {
		t.Fatal("expected runtime deploy_only=true for push main trigger")
	}
	if len(events.items) != 1 {
		t.Fatalf("expected one flow event, got %d", len(events.items))
	}
	if events.items[0].EventType != floweventdomain.EventTypeWebhookReceived {
		t.Fatalf("unexpected event type: %s", events.items[0].EventType)
	}
}

func TestIngestGitHubWebhook_PushMainFork_CreatesDeployOnlyProductionRun(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	svc := NewService(Config{
		AgentRuns:  runs,
		FlowEvents: events,
		Repos:      repos,
	})

	buildRef := "89abcdef0123456789abcdef0123456789abcdef"
	payload := json.RawMessage(fmt.Sprintf(`{
		"ref":"refs/heads/main",
		"before":"0000000000000000000000000000000000000000",
		"after":"%s",
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s-fork","name":"codex-k8s-fork","fork":true},
		"sender":{"id":10,"login":"member"}
	}`, buildRef))
	cmd := IngestCommand{
		CorrelationID: "delivery-push-main-fork-1",
		DeliveryID:    "delivery-push-main-fork-1",
		EventType:     string(webhookdomain.GitHubEventPush),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != webhookdomain.IngestStatusAccepted || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.RunID == "" {
		t.Fatal("expected run id for push main deploy-only trigger")
	}

	var runPayload githubRunPayload
	if err := json.Unmarshal(runs.last.RunPayload, &runPayload); err != nil {
		t.Fatalf("unmarshal run payload: %v", err)
	}
	if got, want := runPayload.Runtime.TargetEnv, "production"; got != want {
		t.Fatalf("unexpected runtime target env: got %q want %q", got, want)
	}
	if runPayload.Runtime.Namespace != "" {
		t.Fatalf("unexpected runtime namespace: got %q want empty (resolved via services.yaml)", runPayload.Runtime.Namespace)
	}
	if got, want := runPayload.Runtime.BuildRef, buildRef; got != want {
		t.Fatalf("unexpected runtime build ref: got %q want %q", got, want)
	}
	if !runPayload.Runtime.DeployOnly {
		t.Fatal("expected runtime deploy_only=true for push main fork trigger")
	}
	if !runPayload.Repository.Fork {
		t.Fatal("expected run payload repository.fork=true for fork repository")
	}
}

func TestIngestGitHubWebhook_PushMain_AutoBumpsVersionsAndSkipsCurrentRun(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	githubMgmt := &inMemoryPushMainVersionBumpClient{
		filesByRef: map[string][]byte{
			"services.yaml@0123456789abcdef0123456789abcdef01234567": []byte(strings.TrimSpace(`
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: codex-k8s
spec:
  versions:
    control-plane:
      value: "0.1.2"
      bumpOn:
        - services/internal/control-plane
    worker:
      value: "0.1.1"
      bumpOn:
        - services/jobs/worker
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
`)),
		},
		changedPaths: []string{
			"services/internal/control-plane/internal/app/app.go",
		},
	}
	svc := NewService(Config{
		AgentRuns:        runs,
		FlowEvents:       events,
		Repos:            repos,
		GitHubToken:      "token",
		GitHubMgmt:       githubMgmt,
		PushMainAutoBump: true,
	})

	buildRef := "0123456789abcdef0123456789abcdef01234567"
	beforeRef := "1111111111111111111111111111111111111111"
	payload := json.RawMessage(fmt.Sprintf(`{
		"ref":"refs/heads/main",
		"before":"%s",
		"after":"%s",
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`, beforeRef, buildRef))
	cmd := IngestCommand{
		CorrelationID: "delivery-push-main-bump-1",
		DeliveryID:    "delivery-push-main-bump-1",
		EventType:     string(webhookdomain.GitHubEventPush),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != webhookdomain.IngestStatusIgnored || got.Duplicate {
		t.Fatalf("unexpected ingest result: %+v", got)
	}
	if got.RunID != "" {
		t.Fatalf("expected no run id when auto bump commits follow-up push, got %q", got.RunID)
	}
	if len(runs.items) != 0 {
		t.Fatalf("expected no run creation on auto bump path, got %d", len(runs.items))
	}
	if githubMgmt.commitCalls != 1 {
		t.Fatalf("expected one commit call, got %d", githubMgmt.commitCalls)
	}
	if got, want := githubMgmt.lastCommitBranch, "main"; got != want {
		t.Fatalf("unexpected commit branch: got %q want %q", got, want)
	}
	if got, want := githubMgmt.lastCommitBaseSHA, buildRef; got != want {
		t.Fatalf("unexpected commit base sha: got %q want %q", got, want)
	}
	updated := string(githubMgmt.lastCommitFiles["services.yaml"])
	if !strings.Contains(updated, `value: "0.1.3"`) {
		t.Fatalf("expected control-plane version bump to 0.1.3, got:\n%s", updated)
	}
	if !strings.Contains(updated, `value: "0.1.1"`) {
		t.Fatalf("worker version should stay 0.1.1, got:\n%s", updated)
	}

	if len(events.items) != 1 {
		t.Fatalf("expected one flow event, got %d", len(events.items))
	}
	if events.items[0].EventType != floweventdomain.EventTypeWebhookIgnored {
		t.Fatalf("unexpected event type: %s", events.items[0].EventType)
	}
	var eventPayload githubFlowEventPayload
	if err := json.Unmarshal(events.items[0].Payload, &eventPayload); err != nil {
		t.Fatalf("unmarshal flow event payload: %v", err)
	}
	if got, want := strings.TrimSpace(eventPayload.Reason), "push_main_versions_autobumped"; got != want {
		t.Fatalf("unexpected ignored reason: got %q want %q", got, want)
	}
}

func TestIngestGitHubWebhook_PushMain_AutoBumpNoMatchesCreatesRun(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	githubMgmt := &inMemoryPushMainVersionBumpClient{
		filesByRef: map[string][]byte{
			"services.yaml@89abcdef0123456789abcdef0123456789abcdef": []byte(strings.TrimSpace(`
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: codex-k8s
spec:
  versions:
    control-plane:
      value: "0.1.2"
      bumpOn:
        - services/internal/control-plane
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
`)),
		},
		changedPaths: []string{
			"docs/architecture/c4_container.md",
		},
	}
	svc := NewService(Config{
		AgentRuns:        runs,
		FlowEvents:       events,
		Repos:            repos,
		GitHubToken:      "token",
		GitHubMgmt:       githubMgmt,
		PushMainAutoBump: true,
	})

	buildRef := "89abcdef0123456789abcdef0123456789abcdef"
	payload := json.RawMessage(fmt.Sprintf(`{
		"ref":"refs/heads/main",
		"before":"1111111111111111111111111111111111111111",
		"after":"%s",
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`, buildRef))
	cmd := IngestCommand{
		CorrelationID: "delivery-push-main-no-bump-1",
		DeliveryID:    "delivery-push-main-no-bump-1",
		EventType:     string(webhookdomain.GitHubEventPush),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != webhookdomain.IngestStatusAccepted || got.Duplicate {
		t.Fatalf("unexpected ingest result: %+v", got)
	}
	if got.RunID == "" {
		t.Fatal("expected run id when no auto bump changes matched")
	}
	if githubMgmt.commitCalls != 0 {
		t.Fatalf("expected no commit calls, got %d", githubMgmt.commitCalls)
	}
}

func TestIngestGitHubWebhook_ClosedEvents_TriggersNamespaceCleanup(t *testing.T) {
	t.Parallel()

	runCase := func(t *testing.T, name string, correlationID string, eventType string, payload json.RawMessage, expectedIssueNumber int64, expectedPRNumber int64) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			runs := &inMemoryRunRepo{items: map[string]string{}}
			events := &inMemoryEventRepo{}
			agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
			repos := &inMemoryRepoCfgRepo{
				byExternalID: map[int64]repocfgrepo.FindResult{
					42: {
						ProjectID:        "project-1",
						RepositoryID:     "repo-1",
						ServicesYAMLPath: "services.yaml",
					},
				},
			}
			runStatus := &inMemoryRunStatusService{}
			svc := NewService(Config{
				AgentRuns:  runs,
				Agents:     agents,
				FlowEvents: events,
				Repos:      repos,
				RunStatus:  runStatus,
			})

			cmd := IngestCommand{
				CorrelationID: correlationID,
				DeliveryID:    correlationID,
				EventType:     eventType,
				ReceivedAt:    time.Now().UTC(),
				Payload:       payload,
			}

			if _, err := svc.IngestGitHubWebhook(ctx, cmd); err != nil {
				t.Fatalf("ingest failed: %v", err)
			}
			if expectedIssueNumber > 0 {
				if runStatus.issueCleanupCalls != 1 {
					t.Fatalf("expected one issue cleanup call, got %d", runStatus.issueCleanupCalls)
				}
				if runStatus.lastIssueCleanup.RepositoryFullName != "codex-k8s/codex-k8s" {
					t.Fatalf("unexpected repository full name: %s", runStatus.lastIssueCleanup.RepositoryFullName)
				}
				if runStatus.lastIssueCleanup.IssueNumber != expectedIssueNumber {
					t.Fatalf("unexpected issue number: %d", runStatus.lastIssueCleanup.IssueNumber)
				}
				return
			}

			if runStatus.pullRequestCleanupCalls != 1 {
				t.Fatalf("expected one pull request cleanup call, got %d", runStatus.pullRequestCleanupCalls)
			}
			if runStatus.lastPullRequestCleanup.RepositoryFullName != "codex-k8s/codex-k8s" {
				t.Fatalf("unexpected repository full name: %s", runStatus.lastPullRequestCleanup.RepositoryFullName)
			}
			if runStatus.lastPullRequestCleanup.PRNumber != expectedPRNumber {
				t.Fatalf("unexpected pull request number: %d", runStatus.lastPullRequestCleanup.PRNumber)
			}
		})
	}

	runCase(t, "issue_closed", "delivery-issue-close-1", string(webhookdomain.GitHubEventIssues), json.RawMessage(`{
		"action":"closed",
		"issue":{"id":1001,"number":77},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`), 77, 0)

	runCase(t, "pull_request_closed", "delivery-pr-close-1", string(webhookdomain.GitHubEventPullRequest), json.RawMessage(`{
		"action":"closed",
		"pull_request":{"id":501,"number":200},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`), 0, 200)
}

func TestIngestGitHubWebhook_LearningMode_DefaultFallback(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{
		byLogin: map[string]userrepo.User{
			"member": {
				ID:          "user-1",
				GitHubLogin: "member",
			},
		},
	}
	members := &inMemoryProjectMemberRepo{
		roles: map[string]string{
			"project-1|user-1": "read_write",
		},
	}
	svc := NewService(Config{
		AgentRuns:           runs,
		Agents:              agents,
		FlowEvents:          events,
		LearningModeDefault: true,
		Repos:               repos,
		Users:               users,
		Members:             members,
	})

	payload := json.RawMessage(`{
		"action":"labeled",
		"label":{"name":"run:dev"},
		"issue":{"id":1001,"number":77,"title":"Implement feature","html_url":"https://github.com/codex-k8s/codex-k8s/issues/77","state":"open"},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-1",
		DeliveryID:    "delivery-1",
		EventType:     string(webhookdomain.GitHubEventIssues),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	if _, err := svc.IngestGitHubWebhook(ctx, cmd); err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if !runs.last.LearningMode {
		t.Fatalf("expected learning mode to fallback to default=true")
	}
}

func TestIngestGitHubWebhook_IssueRunDev_CreatesRunForAllowedMember(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{
		byLogin: map[string]userrepo.User{
			"member": {
				ID:          "user-1",
				GitHubLogin: "member",
			},
		},
	}
	members := &inMemoryProjectMemberRepo{
		roles: map[string]string{
			"project-1|user-1": "read_write",
		},
	}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      users,
		Members:    members,
	})

	payload := json.RawMessage(`{
		"action":"labeled",
		"label":{"name":"run:dev"},
		"issue":{"id":1001,"number":77,"title":"Implement feature","html_url":"https://github.com/codex-k8s/codex-k8s/issues/77","state":"open","user":{"id":55,"login":"owner"}},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-77",
		DeliveryID:    "delivery-77",
		EventType:     string(webhookdomain.GitHubEventIssues),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != webhookdomain.IngestStatusAccepted || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.RunID == "" {
		t.Fatalf("expected run id for issue trigger")
	}

	var runPayload githubRunPayload
	if err := json.Unmarshal(runs.last.RunPayload, &runPayload); err != nil {
		t.Fatalf("unmarshal run payload: %v", err)
	}
	if runPayload.Trigger == nil {
		t.Fatalf("expected trigger object in run payload")
	}
	if runPayload.Trigger.Kind != webhookdomain.TriggerKindDev {
		t.Fatalf("unexpected trigger kind: %#v", runPayload.Trigger.Kind)
	}
	if runPayload.Trigger.Label != webhookdomain.DefaultRunDevLabel {
		t.Fatalf("unexpected trigger label: %#v", runPayload.Trigger.Label)
	}
	if got, want := runPayload.Runtime.Mode, "full-env"; got != want {
		t.Fatalf("unexpected runtime mode: got %q want %q", got, want)
	}
	if got, want := runPayload.Runtime.Source, runtimeModeSourceTriggerDefault; got != want {
		t.Fatalf("unexpected runtime source: got %q want %q", got, want)
	}
	if runPayload.Agent.Key != "dev" {
		t.Fatalf("unexpected agent key: %#v", runPayload.Agent.Key)
	}
	if runPayload.Agent.Name != "AI Developer" {
		t.Fatalf("unexpected agent name: %#v", runPayload.Agent.Name)
	}
}

func TestIngestGitHubWebhook_IssueRunDev_PostsPlannedRunStatusImmediately(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	runStatus := &inMemoryRunStatusService{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{
		byLogin: map[string]userrepo.User{
			"member": {
				ID:          "user-1",
				GitHubLogin: "member",
			},
		},
	}
	members := &inMemoryProjectMemberRepo{
		roles: map[string]string{
			"project-1|user-1": "read_write",
		},
	}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      users,
		Members:    members,
		RunStatus:  runStatus,
	})

	payload := json.RawMessage(`{
		"action":"labeled",
		"label":{"name":"run:dev"},
		"issue":{"id":1001,"number":177,"title":"Implement feature","html_url":"https://github.com/codex-k8s/codex-k8s/issues/177","state":"open","user":{"id":55,"login":"owner"}},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-177-planned",
		DeliveryID:    "delivery-177-planned",
		EventType:     string(webhookdomain.GitHubEventIssues),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.RunID == "" {
		t.Fatal("expected run id")
	}
	if len(runStatus.statusCommentUpsertCalls) != 1 {
		t.Fatalf("expected one status comment upsert call, got %d", len(runStatus.statusCommentUpsertCalls))
	}
	comment := runStatus.statusCommentUpsertCalls[0]
	if comment.RunID != got.RunID {
		t.Fatalf("unexpected run id in status comment call: got %q want %q", comment.RunID, got.RunID)
	}
	if comment.Phase != runstatusdomain.PhaseCreated {
		t.Fatalf("expected phase %q, got %q", runstatusdomain.PhaseCreated, comment.Phase)
	}
	if comment.RunStatus != "pending" {
		t.Fatalf("expected pending run status, got %q", comment.RunStatus)
	}
	if comment.TriggerKind != string(webhookdomain.TriggerKindDev) {
		t.Fatalf("expected trigger kind %q, got %q", webhookdomain.TriggerKindDev, comment.TriggerKind)
	}
	if comment.PromptLocale != "ru" {
		t.Fatalf("expected prompt locale ru, got %q", comment.PromptLocale)
	}
}

func TestIngestGitHubWebhook_RuntimePolicyOverrideFromServicesYAML(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{
		byLogin: map[string]userrepo.User{
			"member": {
				ID:          "user-1",
				GitHubLogin: "member",
			},
		},
	}
	members := &inMemoryProjectMemberRepo{
		roles: map[string]string{
			"project-1|user-1": "read_write",
		},
	}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      users,
		Members:    members,
		RuntimeModePolicy: RuntimeModePolicy{
			Configured:  true,
			Source:      "services.yaml",
			DefaultMode: agentdomain.RuntimeModeFullEnv,
			TriggerModes: map[webhookdomain.TriggerKind]agentdomain.RuntimeMode{
				webhookdomain.TriggerKindDev: agentdomain.RuntimeModeCodeOnly,
			},
		},
	})

	payload := json.RawMessage(`{
		"action":"labeled",
		"label":{"name":"run:dev"},
		"issue":{"id":1001,"number":177,"title":"Implement feature","html_url":"https://github.com/codex-k8s/codex-k8s/issues/177","state":"open","user":{"id":55,"login":"owner"}},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-177",
		DeliveryID:    "delivery-177",
		EventType:     string(webhookdomain.GitHubEventIssues),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	if _, err := svc.IngestGitHubWebhook(ctx, cmd); err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	var runPayload githubRunPayload
	if err := json.Unmarshal(runs.last.RunPayload, &runPayload); err != nil {
		t.Fatalf("unmarshal run payload: %v", err)
	}
	if got, want := runPayload.Runtime.Mode, "code-only"; got != want {
		t.Fatalf("unexpected runtime mode: got %q want %q", got, want)
	}
	if got, want := runPayload.Runtime.Source, runtimeModeSourceServicesYAML; got != want {
		t.Fatalf("unexpected runtime source: got %q want %q", got, want)
	}
}

func TestIngestGitHubWebhook_IssueRunVision_CreatesStageRunForAllowedMember(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{
		"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"},
		"pm":  {ID: "agent-pm", AgentKey: "pm", Name: "AI Product Manager"},
	}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{
		byLogin: map[string]userrepo.User{
			"member": {
				ID:          "user-1",
				GitHubLogin: "member",
			},
		},
	}
	members := &inMemoryProjectMemberRepo{
		roles: map[string]string{
			"project-1|user-1": "read_write",
		},
	}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      users,
		Members:    members,
	})

	payload := json.RawMessage(`{
		"action":"labeled",
		"label":{"name":"run:vision"},
		"issue":{"id":1001,"number":78,"title":"Vision stage","html_url":"https://github.com/codex-k8s/codex-k8s/issues/78","state":"open","labels":[{"name":"run:vision"}],"user":{"id":55,"login":"owner"}},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-vision-78",
		DeliveryID:    "delivery-vision-78",
		EventType:     string(webhookdomain.GitHubEventIssues),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != webhookdomain.IngestStatusAccepted || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.RunID == "" {
		t.Fatalf("expected run id for issue trigger")
	}

	var runPayload githubRunPayload
	if err := json.Unmarshal(runs.last.RunPayload, &runPayload); err != nil {
		t.Fatalf("unmarshal run payload: %v", err)
	}
	if runPayload.Trigger == nil {
		t.Fatalf("expected trigger object in run payload")
	}
	if runPayload.Trigger.Kind != webhookdomain.TriggerKindVision {
		t.Fatalf("unexpected trigger kind: %#v", runPayload.Trigger.Kind)
	}
	if runPayload.Trigger.Label != webhookdomain.DefaultRunVisionLabel {
		t.Fatalf("unexpected trigger label: %#v", runPayload.Trigger.Label)
	}
	if runPayload.Agent.Key != "pm" {
		t.Fatalf("unexpected agent key: %#v", runPayload.Agent.Key)
	}
}

func TestResolveRunAgentKey_SelfImproveUsesKM(t *testing.T) {
	t.Parallel()

	key := resolveRunAgentKey(&issueRunTrigger{
		Kind: webhookdomain.TriggerKindSelfImprove,
	})
	if key != "km" {
		t.Fatalf("resolveRunAgentKey() = %q, want %q", key, "km")
	}
}

func TestResolveRunAgentKey_DocAuditUsesKM(t *testing.T) {
	t.Parallel()

	key := resolveRunAgentKey(&issueRunTrigger{
		Kind: webhookdomain.TriggerKindDocAudit,
	})
	if key != "km" {
		t.Fatalf("resolveRunAgentKey() = %q, want %q", key, "km")
	}
}

func TestIngestGitHubWebhook_IssueTriggerConflict_IgnoredWithDiagnosticComment(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{
		byLogin: map[string]userrepo.User{
			"member": {
				ID:          "user-1",
				GitHubLogin: "member",
			},
		},
	}
	members := &inMemoryProjectMemberRepo{
		roles: map[string]string{
			"project-1|user-1": "read_write",
		},
	}
	runStatus := &inMemoryRunStatusService{}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      users,
		Members:    members,
		RunStatus:  runStatus,
	})

	payload := json.RawMessage(`{
		"action":"labeled",
		"label":{"name":"run:vision"},
		"issue":{"id":1001,"number":79,"title":"Conflict stage","html_url":"https://github.com/codex-k8s/codex-k8s/issues/79","state":"open","labels":[{"name":"run:dev"},{"name":"run:vision"}],"user":{"id":55,"login":"owner"}},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-conflict-79",
		DeliveryID:    "delivery-conflict-79",
		EventType:     string(webhookdomain.GitHubEventIssues),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != webhookdomain.IngestStatusIgnored || got.RunID != "" || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if len(runs.items) != 0 {
		t.Fatalf("expected no run creation for conflicting trigger labels")
	}
	if runStatus.conflictCommentCalls != 1 {
		t.Fatalf("expected conflict comment call, got %d", runStatus.conflictCommentCalls)
	}
	if runStatus.lastConflictComment.IssueNumber != 79 {
		t.Fatalf("unexpected issue number in conflict comment params: %d", runStatus.lastConflictComment.IssueNumber)
	}
	if len(events.items) != 1 {
		t.Fatalf("expected one flow event, got %d", len(events.items))
	}
	if events.items[0].EventType != floweventdomain.EventTypeWebhookIgnored {
		t.Fatalf("unexpected event type: %s", events.items[0].EventType)
	}
	var payloadJSON map[string]any
	if err := json.Unmarshal(events.items[0].Payload, &payloadJSON); err != nil {
		t.Fatalf("decode ignored event payload: %v", err)
	}
	if payloadJSON["reason"] != "issue_trigger_label_conflict" {
		t.Fatalf("unexpected reason: %#v", payloadJSON["reason"])
	}
}

func TestIngestGitHubWebhook_PullRequestReviewChangesRequested_WithoutRunLabel_IsIgnored(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	runStatus := &inMemoryRunStatusService{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{
		byLogin: map[string]userrepo.User{
			"member": {
				ID:          "user-1",
				GitHubLogin: "member",
			},
		},
	}
	members := &inMemoryProjectMemberRepo{
		roles: map[string]string{
			"project-1|user-1": "read_write",
		},
	}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      users,
		Members:    members,
		RunStatus:  runStatus,
	})

	payload := json.RawMessage(`{
		"action":"submitted",
		"review":{"state":"changes_requested"},
		"pull_request":{
			"id":501,
			"number":200,
			"title":"WIP feature",
			"html_url":"https://github.com/codex-k8s/codex-k8s/pull/200",
			"state":"open",
			"head":{"ref":"codex/issue-13"},
			"user":{"id":55,"login":"member"}
		},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-pr-review-1",
		DeliveryID:    "delivery-pr-review-1",
		EventType:     string(webhookdomain.GitHubEventPullRequestReview),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != webhookdomain.IngestStatusIgnored || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.RunID != "" {
		t.Fatalf("expected no run id for pull_request_review without run label, got %q", got.RunID)
	}
	if len(runs.items) != 0 {
		t.Fatalf("expected no run creation, got %d runs", len(runs.items))
	}
	if len(runStatus.warningCommentCalls) != 1 {
		t.Fatalf("expected warning comment call, got %d", len(runStatus.warningCommentCalls))
	}
	if runStatus.warningCommentCalls[0].ReasonCode != runstatusdomain.TriggerWarningReasonPullRequestReviewStageNotResolved {
		t.Fatalf("unexpected warning reason: %q", runStatus.warningCommentCalls[0].ReasonCode)
	}
}

func TestIngestGitHubWebhook_PullRequestReviewChangesRequested_WithRunDevReviseLabel_CreatesReviseRun(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{
		byLogin: map[string]userrepo.User{
			"member": {
				ID:          "user-1",
				GitHubLogin: "member",
			},
		},
	}
	members := &inMemoryProjectMemberRepo{
		roles: map[string]string{
			"project-1|user-1": "read_write",
		},
	}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      users,
		Members:    members,
	})

	payload := json.RawMessage(`{
		"action":"submitted",
		"review":{"state":"changes_requested"},
		"pull_request":{
			"id":501,
			"number":200,
			"title":"WIP feature",
			"html_url":"https://github.com/codex-k8s/codex-k8s/pull/200",
			"state":"open",
			"labels":[{"name":"run:dev:revise"}],
			"head":{"ref":"codex/issue-13"},
			"user":{"id":55,"login":"member"}
		},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-pr-review-2",
		DeliveryID:    "delivery-pr-review-2",
		EventType:     string(webhookdomain.GitHubEventPullRequestReview),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != webhookdomain.IngestStatusAccepted || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.RunID == "" {
		t.Fatalf("expected run id for pull_request_review trigger with run label")
	}

	var runPayload githubRunPayload
	if err := json.Unmarshal(runs.last.RunPayload, &runPayload); err != nil {
		t.Fatalf("unmarshal run payload: %v", err)
	}
	if runPayload.Trigger == nil {
		t.Fatalf("expected trigger object in run payload")
	}
	if runPayload.Trigger.Source != webhookdomain.TriggerSourcePullRequestReview {
		t.Fatalf("unexpected trigger source: %#v", runPayload.Trigger.Source)
	}
	if runPayload.Trigger.Kind != webhookdomain.TriggerKindDevRevise {
		t.Fatalf("unexpected trigger kind: %#v", runPayload.Trigger.Kind)
	}
	if runPayload.Trigger.Label != webhookdomain.DefaultRunDevReviseLabel {
		t.Fatalf("unexpected trigger label: %#v", runPayload.Trigger.Label)
	}
	if runPayload.Issue == nil || runPayload.Issue.Number != 200 {
		t.Fatalf("expected issue payload with number=200, got %#v", runPayload.Issue)
	}
	if runPayload.PullRequest == nil || runPayload.PullRequest.Number != 200 {
		t.Fatalf("expected pull_request payload with number=200, got %#v", runPayload.PullRequest)
	}
}

func TestIngestGitHubWebhook_PullRequestReviewChangesRequested_WithRunIntakeLabel_CreatesIntakeReviseRun(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"pm": {ID: "agent-pm", AgentKey: "pm", Name: "AI PM"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{
		byLogin: map[string]userrepo.User{
			"member": {ID: "user-1", GitHubLogin: "member"},
		},
	}
	members := &inMemoryProjectMemberRepo{
		roles: map[string]string{"project-1|user-1": "read_write"},
	}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      users,
		Members:    members,
	})

	payload := json.RawMessage(`{
		"action":"submitted",
		"review":{"state":"changes_requested"},
		"pull_request":{
			"id":501,
			"number":201,
			"title":"Intake artifacts",
			"html_url":"https://github.com/codex-k8s/codex-k8s/pull/201",
			"state":"open",
			"labels":[{"name":"run:intake"}],
			"head":{"ref":"codex/issue-201"},
			"user":{"id":55,"login":"member"}
		},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-pr-review-intake",
		DeliveryID:    "delivery-pr-review-intake",
		EventType:     string(webhookdomain.GitHubEventPullRequestReview),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.RunID == "" {
		t.Fatalf("expected run id for pull_request_review trigger with run:intake label")
	}

	var runPayload githubRunPayload
	if err := json.Unmarshal(runs.last.RunPayload, &runPayload); err != nil {
		t.Fatalf("unmarshal run payload: %v", err)
	}
	if runPayload.Trigger == nil {
		t.Fatalf("expected trigger object in run payload")
	}
	if runPayload.Trigger.Kind != webhookdomain.TriggerKindIntakeRevise {
		t.Fatalf("unexpected trigger kind: %#v", runPayload.Trigger.Kind)
	}
	if runPayload.Trigger.Label != webhookdomain.DefaultRunIntakeReviseLabel {
		t.Fatalf("unexpected trigger label: %#v", runPayload.Trigger.Label)
	}
}

func TestIngestGitHubWebhook_PullRequestReviewChangesRequested_ResolvesFromIssueLabelsHistory(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{
		items: map[string]string{},
		byRunID: map[string]agentrunrepo.Run{
			"history-run-1": {
				ID:            "history-run-1",
				CorrelationID: "history-correlation-1",
				ProjectID:     "project-1",
				Status:        "succeeded",
				RunPayload:    json.RawMessage(`{"raw_payload":{"issue":{"labels":[{"name":"run:plan"},{"name":"[ai-model-gpt-5.2-codex]"}]}}}`),
			},
		},
		searchItems: []agentrunrepo.RunLookupItem{
			{
				RunID:              "history-run-1",
				ProjectID:          "project-1",
				RepositoryFullName: "codex-k8s/codex-k8s",
				IssueNumber:        95,
				PullRequestNumber:  203,
				TriggerKind:        "plan",
			},
		},
	}
	events := &inMemoryEventRepo{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"em": {ID: "agent-em", AgentKey: "em", Name: "AI EM"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{byLogin: map[string]userrepo.User{"member": {ID: "user-1", GitHubLogin: "member"}}}
	members := &inMemoryProjectMemberRepo{roles: map[string]string{"project-1|user-1": "read_write"}}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      users,
		Members:    members,
	})

	payload := json.RawMessage(`{
		"action":"submitted",
		"review":{"state":"changes_requested"},
		"pull_request":{
			"id":501,
			"number":203,
			"title":"Plan fixes",
			"html_url":"https://github.com/codex-k8s/codex-k8s/pull/203",
			"state":"open",
			"head":{"ref":"codex/issue-95"},
			"user":{"id":55,"login":"member"}
		},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-pr-review-history-issue-labels",
		DeliveryID:    "delivery-pr-review-history-issue-labels",
		EventType:     string(webhookdomain.GitHubEventPullRequestReview),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.RunID == "" {
		t.Fatal("expected run id")
	}

	var runPayload githubRunPayload
	if err := json.Unmarshal(runs.last.RunPayload, &runPayload); err != nil {
		t.Fatalf("unmarshal run payload: %v", err)
	}
	if runPayload.Trigger == nil {
		t.Fatal("expected trigger in run payload")
	}
	if runPayload.Trigger.Kind != webhookdomain.TriggerKindPlanRevise {
		t.Fatalf("unexpected trigger kind: %q", runPayload.Trigger.Kind)
	}
	if runPayload.Trigger.Label != webhookdomain.DefaultRunPlanReviseLabel {
		t.Fatalf("unexpected trigger label: %q", runPayload.Trigger.Label)
	}
	if runPayload.ProfileHints == nil {
		t.Fatal("expected profile hints in run payload")
	}
	if len(runPayload.ProfileHints.LastRunIssueLabels) == 0 {
		t.Fatalf("expected non-empty last run issue labels in profile hints: %#v", runPayload.ProfileHints)
	}
}

func TestIngestGitHubWebhook_PullRequestReviewChangesRequested_ResolvesFromLastRunContext(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{
		items: map[string]string{},
		searchItems: []agentrunrepo.RunLookupItem{
			{
				RunID:              "history-run-2",
				ProjectID:          "project-1",
				RepositoryFullName: "codex-k8s/codex-k8s",
				IssueNumber:        96,
				PullRequestNumber:  204,
				TriggerKind:        "design",
			},
		},
	}
	events := &inMemoryEventRepo{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"sa": {ID: "agent-sa", AgentKey: "sa", Name: "AI SA"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{byLogin: map[string]userrepo.User{"member": {ID: "user-1", GitHubLogin: "member"}}}
	members := &inMemoryProjectMemberRepo{roles: map[string]string{"project-1|user-1": "read_write"}}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      users,
		Members:    members,
	})

	payload := json.RawMessage(`{
		"action":"submitted",
		"review":{"state":"changes_requested"},
		"pull_request":{
			"id":501,
			"number":204,
			"title":"Design fixes",
			"html_url":"https://github.com/codex-k8s/codex-k8s/pull/204",
			"state":"open",
			"head":{"ref":"codex/issue-96"},
			"user":{"id":55,"login":"member"}
		},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-pr-review-last-run-context",
		DeliveryID:    "delivery-pr-review-last-run-context",
		EventType:     string(webhookdomain.GitHubEventPullRequestReview),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.RunID == "" {
		t.Fatal("expected run id")
	}

	var runPayload githubRunPayload
	if err := json.Unmarshal(runs.last.RunPayload, &runPayload); err != nil {
		t.Fatalf("unmarshal run payload: %v", err)
	}
	if runPayload.Trigger == nil {
		t.Fatal("expected trigger in run payload")
	}
	if runPayload.Trigger.Kind != webhookdomain.TriggerKindDesignRevise {
		t.Fatalf("unexpected trigger kind: %q", runPayload.Trigger.Kind)
	}
	if runPayload.Trigger.Label != webhookdomain.DefaultRunDesignReviseLabel {
		t.Fatalf("unexpected trigger label: %q", runPayload.Trigger.Label)
	}
}

func TestIngestGitHubWebhook_PullRequestReviewChangesRequested_WithMultipleStageLabels_IsIgnored(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	runStatus := &inMemoryRunStatusService{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	users := &inMemoryUserRepo{
		byLogin: map[string]userrepo.User{
			"member": {ID: "user-1", GitHubLogin: "member"},
		},
	}
	members := &inMemoryProjectMemberRepo{
		roles: map[string]string{"project-1|user-1": "read_write"},
	}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      users,
		Members:    members,
		RunStatus:  runStatus,
	})

	payload := json.RawMessage(`{
		"action":"submitted",
		"review":{"state":"changes_requested"},
		"pull_request":{
			"id":501,
			"number":202,
			"title":"Conflicting stage labels",
			"html_url":"https://github.com/codex-k8s/codex-k8s/pull/202",
			"state":"open",
			"labels":[{"name":"run:dev"},{"name":"run:plan"}],
			"head":{"ref":"codex/issue-202"},
			"user":{"id":55,"login":"member"}
		},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-pr-review-conflict",
		DeliveryID:    "delivery-pr-review-conflict",
		EventType:     string(webhookdomain.GitHubEventPullRequestReview),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.RunID != "" {
		t.Fatalf("expected no run id for pull_request_review with multiple stage labels, got %q", got.RunID)
	}
	if got.Status != webhookdomain.IngestStatusIgnored {
		t.Fatalf("expected ignored status, got %+v", got)
	}
	if len(runs.items) != 0 {
		t.Fatalf("expected no run creation, got %d runs", len(runs.items))
	}
	if len(runStatus.warningCommentCalls) != 1 {
		t.Fatalf("expected warning comment call, got %d", len(runStatus.warningCommentCalls))
	}
	if runStatus.warningCommentCalls[0].ReasonCode != runstatusdomain.TriggerWarningReasonPullRequestReviewStageAmbiguous {
		t.Fatalf("unexpected warning reason: %q", runStatus.warningCommentCalls[0].ReasonCode)
	}
}

func TestIngestGitHubWebhook_IssueRunDev_DeniesUnknownSender(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	runStatus := &inMemoryRunStatusService{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
	repos := &inMemoryRepoCfgRepo{
		byExternalID: map[int64]repocfgrepo.FindResult{
			42: {
				ProjectID:        "project-1",
				RepositoryID:     "repo-1",
				ServicesYAMLPath: "services.yaml",
			},
		},
	}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      &inMemoryUserRepo{},
		Members:    &inMemoryProjectMemberRepo{},
		RunStatus:  runStatus,
	})

	payload := json.RawMessage(`{
		"action":"labeled",
		"label":{"name":"run:dev"},
		"issue":{"id":1001,"number":77,"title":"Implement feature","html_url":"https://github.com/codex-k8s/codex-k8s/issues/77","state":"open"},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"unknown"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-78",
		DeliveryID:    "delivery-78",
		EventType:     string(webhookdomain.GitHubEventIssues),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != webhookdomain.IngestStatusIgnored || got.RunID != "" || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if len(runs.items) != 0 {
		t.Fatalf("expected no run creation for denied sender")
	}
	if len(events.items) != 1 {
		t.Fatalf("expected one flow event, got %d", len(events.items))
	}
	if events.items[0].EventType != floweventdomain.EventTypeWebhookIgnored {
		t.Fatalf("unexpected event type: %s", events.items[0].EventType)
	}
	if len(runStatus.warningCommentCalls) != 1 {
		t.Fatalf("expected warning comment call, got %d", len(runStatus.warningCommentCalls))
	}
	if runStatus.warningCommentCalls[0].ThreadKind != "issue" {
		t.Fatalf("expected issue thread warning, got %q", runStatus.warningCommentCalls[0].ThreadKind)
	}
}

func TestIngestGitHubWebhook_IssueNonTriggerLabelIgnored(t *testing.T) {
	ctx := context.Background()
	runs := &inMemoryRunRepo{items: map[string]string{}}
	events := &inMemoryEventRepo{}
	agents := &inMemoryAgentRepo{items: map[string]agentrepo.Agent{"dev": {ID: "agent-dev", AgentKey: "dev", Name: "AI Developer"}}}
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
	})

	payload := json.RawMessage(`{
		"action":"labeled",
		"label":{"name":"bug"},
		"issue":{"id":1001,"number":77},
		"repository":{"id":42,"full_name":"codex-k8s/codex-k8s","name":"codex-k8s"},
		"sender":{"id":10,"login":"member"}
	}`)
	cmd := IngestCommand{
		CorrelationID: "delivery-79",
		DeliveryID:    "delivery-79",
		EventType:     string(webhookdomain.GitHubEventIssues),
		ReceivedAt:    time.Now().UTC(),
		Payload:       payload,
	}

	got, err := svc.IngestGitHubWebhook(ctx, cmd)
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}
	if got.Status != webhookdomain.IngestStatusIgnored || got.RunID != "" || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if len(runs.items) != 0 {
		t.Fatalf("expected no run creation for non-trigger issue label")
	}
}

type inMemoryAgentRepo struct {
	items map[string]agentrepo.Agent
}

func (r *inMemoryAgentRepo) FindEffectiveByKey(_ context.Context, _ string, agentKey string) (agentrepo.Agent, bool, error) {
	if len(r.items) == 0 {
		return agentrepo.Agent{}, false, nil
	}
	lookupKey := strings.TrimSpace(agentKey)
	if lookupKey == "" {
		return agentrepo.Agent{}, false, nil
	}
	item, ok := r.items[lookupKey]
	if !ok {
		for key, value := range r.items {
			if strings.EqualFold(key, lookupKey) {
				return value, true, nil
			}
		}
		return agentrepo.Agent{}, false, nil
	}
	return item, true, nil
}

type inMemoryRunRepo struct {
	items       map[string]string
	last        agentrunrepo.CreateParams
	byRunID     map[string]agentrunrepo.Run
	searchItems []agentrunrepo.RunLookupItem
}

func (r *inMemoryRunRepo) CreatePendingIfAbsent(_ context.Context, params agentrunrepo.CreateParams) (agentrunrepo.CreateResult, error) {
	r.last = params
	if r.byRunID == nil {
		r.byRunID = make(map[string]agentrunrepo.Run)
	}
	if id, ok := r.items[params.CorrelationID]; ok {
		return agentrunrepo.CreateResult{
			RunID:    id,
			Inserted: false,
		}, nil
	}
	id := "run-" + params.CorrelationID
	r.items[params.CorrelationID] = id
	r.byRunID[id] = agentrunrepo.Run{
		ID:            id,
		CorrelationID: params.CorrelationID,
		ProjectID:     params.ProjectID,
		Status:        "pending",
		RunPayload:    params.RunPayload,
	}
	return agentrunrepo.CreateResult{
		RunID:    id,
		Inserted: true,
	}, nil
}

func (r *inMemoryRunRepo) GetByID(_ context.Context, runID string) (agentrunrepo.Run, bool, error) {
	if item, ok := r.byRunID[runID]; ok {
		return item, true, nil
	}
	for correlationID, existingRunID := range r.items {
		if existingRunID == runID {
			return agentrunrepo.Run{
				ID:            existingRunID,
				CorrelationID: correlationID,
				ProjectID:     r.last.ProjectID,
				Status:        "pending",
				RunPayload:    r.last.RunPayload,
			}, true, nil
		}
	}
	return agentrunrepo.Run{}, false, nil
}

func (r *inMemoryRunRepo) ListRecentByProject(_ context.Context, _ string, _ string, _ int, _ int) ([]agentrunrepo.RunLookupItem, error) {
	return nil, nil
}

func (r *inMemoryRunRepo) SearchRecentByProjectIssueOrPullRequest(_ context.Context, _ string, _ string, _ int64, _ int64, _ int) ([]agentrunrepo.RunLookupItem, error) {
	items := make([]agentrunrepo.RunLookupItem, 0, len(r.searchItems))
	items = append(items, r.searchItems...)
	return items, nil
}

func (r *inMemoryRunRepo) ListRunIDsByRepositoryIssue(_ context.Context, _ string, _ int64, _ int) ([]string, error) {
	return nil, nil
}

func (r *inMemoryRunRepo) ListRunIDsByRepositoryPullRequest(_ context.Context, _ string, _ int64, _ int) ([]string, error) {
	return nil, nil
}

type inMemoryRunStatusService struct {
	issueCleanupCalls        int
	pullRequestCleanupCalls  int
	lastIssueCleanup         runstatusdomain.CleanupByIssueParams
	lastPullRequestCleanup   runstatusdomain.CleanupByPullRequestParams
	conflictCommentCalls     int
	lastConflictComment      runstatusdomain.TriggerLabelConflictCommentParams
	warningCommentCalls      []runstatusdomain.TriggerWarningCommentParams
	statusCommentUpsertCalls []runstatusdomain.UpsertCommentParams
}

func (s *inMemoryRunStatusService) UpsertRunStatusComment(_ context.Context, params runstatusdomain.UpsertCommentParams) (runstatusdomain.UpsertCommentResult, error) {
	s.statusCommentUpsertCalls = append(s.statusCommentUpsertCalls, params)
	return runstatusdomain.UpsertCommentResult{
		CommentID:  1,
		CommentURL: "https://example.invalid/run-status",
	}, nil
}

func (s *inMemoryRunStatusService) CleanupNamespacesByIssue(_ context.Context, params runstatusdomain.CleanupByIssueParams) (runstatusdomain.CleanupByIssueResult, error) {
	s.issueCleanupCalls++
	s.lastIssueCleanup = params
	return runstatusdomain.CleanupByIssueResult{}, nil
}

func (s *inMemoryRunStatusService) CleanupNamespacesByPullRequest(_ context.Context, params runstatusdomain.CleanupByPullRequestParams) (runstatusdomain.CleanupByIssueResult, error) {
	s.pullRequestCleanupCalls++
	s.lastPullRequestCleanup = params
	return runstatusdomain.CleanupByIssueResult{}, nil
}

func (s *inMemoryRunStatusService) PostTriggerLabelConflictComment(_ context.Context, params runstatusdomain.TriggerLabelConflictCommentParams) (runstatusdomain.TriggerLabelConflictCommentResult, error) {
	s.conflictCommentCalls++
	s.lastConflictComment = params
	return runstatusdomain.TriggerLabelConflictCommentResult{
		CommentID:  1,
		CommentURL: "https://example.test/comment/1",
	}, nil
}

func (s *inMemoryRunStatusService) PostTriggerWarningComment(_ context.Context, params runstatusdomain.TriggerWarningCommentParams) (runstatusdomain.TriggerWarningCommentResult, error) {
	s.warningCommentCalls = append(s.warningCommentCalls, params)
	return runstatusdomain.TriggerWarningCommentResult{
		CommentID:  int64(len(s.warningCommentCalls)),
		CommentURL: "https://example.invalid/warning",
	}, nil
}

type inMemoryEventRepo struct {
	items []floweventrepo.InsertParams
}

func (r *inMemoryEventRepo) Insert(_ context.Context, params floweventrepo.InsertParams) error {
	r.items = append(r.items, params)
	return nil
}

type inMemoryRepoCfgRepo struct {
	byExternalID map[int64]repocfgrepo.FindResult
}

func (r *inMemoryRepoCfgRepo) ListForProject(_ context.Context, _ string, _ int) ([]repocfgrepo.RepositoryBinding, error) {
	return nil, nil
}

func (r *inMemoryRepoCfgRepo) GetByID(_ context.Context, repositoryID string) (repocfgrepo.RepositoryBinding, bool, error) {
	for _, item := range r.byExternalID {
		if item.RepositoryID == repositoryID {
			return repocfgrepo.RepositoryBinding{
				ID:               item.RepositoryID,
				ProjectID:        item.ProjectID,
				Provider:         "github",
				Owner:            "codex-k8s",
				Name:             "codex-k8s",
				ServicesYAMLPath: item.ServicesYAMLPath,
			}, true, nil
		}
	}
	return repocfgrepo.RepositoryBinding{}, false, nil
}

func (r *inMemoryRepoCfgRepo) Upsert(_ context.Context, _ repocfgrepo.UpsertParams) (repocfgrepo.RepositoryBinding, error) {
	return repocfgrepo.RepositoryBinding{}, fmt.Errorf("not implemented")
}

func (r *inMemoryRepoCfgRepo) Delete(_ context.Context, _, _ string) error {
	return nil
}

func (r *inMemoryRepoCfgRepo) FindByProviderExternalID(_ context.Context, _ string, externalID int64) (repocfgrepo.FindResult, bool, error) {
	res, ok := r.byExternalID[externalID]
	if !ok {
		return repocfgrepo.FindResult{}, false, nil
	}
	return res, true, nil
}

func (r *inMemoryRepoCfgRepo) FindByProviderOwnerName(_ context.Context, _ string, _ string, _ string) (repocfgrepo.FindResult, bool, error) {
	return repocfgrepo.FindResult{}, false, nil
}

func (r *inMemoryRepoCfgRepo) GetTokenEncrypted(_ context.Context, _ string) ([]byte, bool, error) {
	return nil, false, nil
}

func (r *inMemoryRepoCfgRepo) GetBotTokenEncrypted(_ context.Context, _ string) ([]byte, bool, error) {
	return nil, false, nil
}

func (r *inMemoryRepoCfgRepo) UpsertBotParams(_ context.Context, _ repocfgrepo.RepositoryBotParamsUpsertParams) error {
	return nil
}

func (r *inMemoryRepoCfgRepo) UpsertPreflightReport(_ context.Context, _ repocfgrepo.RepositoryPreflightReportUpsertParams) error {
	return nil
}

func (r *inMemoryRepoCfgRepo) AcquirePreflightLock(_ context.Context, params repocfgrepo.RepositoryPreflightLockAcquireParams) (string, bool, error) {
	return params.LockToken, true, nil
}

func (r *inMemoryRepoCfgRepo) ReleasePreflightLock(_ context.Context, _ string, _ string) error {
	return nil
}

func (r *inMemoryRepoCfgRepo) SetTokenEncryptedForAll(_ context.Context, _ []byte) (int64, error) {
	return 0, nil
}

type inMemoryUserRepo struct {
	byLogin map[string]userrepo.User
}

func (r *inMemoryUserRepo) EnsureOwner(_ context.Context, _ string) (userrepo.User, error) {
	return userrepo.User{}, nil
}

func (r *inMemoryUserRepo) GetByID(_ context.Context, _ string) (userrepo.User, bool, error) {
	return userrepo.User{}, false, nil
}

func (r *inMemoryUserRepo) GetByEmail(_ context.Context, _ string) (userrepo.User, bool, error) {
	return userrepo.User{}, false, nil
}

func (r *inMemoryUserRepo) GetByGitHubLogin(_ context.Context, githubLogin string) (userrepo.User, bool, error) {
	u, ok := r.byLogin[githubLogin]
	return u, ok, nil
}

func (r *inMemoryUserRepo) UpdateGitHubIdentity(_ context.Context, _ string, _ int64, _ string) error {
	return nil
}

func (r *inMemoryUserRepo) CreateAllowedUser(_ context.Context, _ string, _ bool) (userrepo.User, error) {
	return userrepo.User{}, nil
}

func (r *inMemoryUserRepo) List(_ context.Context, _ int) ([]userrepo.User, error) {
	return nil, nil
}

func (r *inMemoryUserRepo) DeleteByID(_ context.Context, _ string) error {
	return nil
}

type inMemoryProjectMemberRepo struct {
	roles map[string]string
}

func (r *inMemoryProjectMemberRepo) List(_ context.Context, _ string, _ int) ([]projectmemberrepo.Member, error) {
	return nil, nil
}

func (r *inMemoryProjectMemberRepo) Upsert(_ context.Context, _, _, _ string) error {
	return nil
}

func (r *inMemoryProjectMemberRepo) Delete(_ context.Context, _, _ string) error {
	return nil
}

func (r *inMemoryProjectMemberRepo) GetRole(_ context.Context, projectID string, userID string) (string, bool, error) {
	role, ok := r.roles[projectID+"|"+userID]
	return role, ok, nil
}

func (r *inMemoryProjectMemberRepo) SetLearningModeOverride(_ context.Context, _, _ string, _ *bool) error {
	return nil
}

func (r *inMemoryProjectMemberRepo) GetLearningModeOverride(_ context.Context, _, _ string) (*bool, bool, error) {
	return nil, false, nil
}

type inMemoryPushMainVersionBumpClient struct {
	filesByRef map[string][]byte

	changedPaths []string
	changedErr   error
	commitErr    error

	commitCalls       int
	lastCommitOwner   string
	lastCommitRepo    string
	lastCommitBranch  string
	lastCommitBaseSHA string
	lastCommitMessage string
	lastCommitFiles   map[string][]byte
}

func (c *inMemoryPushMainVersionBumpClient) GetFile(_ context.Context, _ string, _ string, _ string, filePath string, ref string) ([]byte, bool, error) {
	if c == nil {
		return nil, false, nil
	}
	if c.filesByRef == nil {
		return nil, false, nil
	}
	key := strings.TrimSpace(filePath) + "@" + strings.TrimSpace(ref)
	if raw, ok := c.filesByRef[key]; ok {
		return append([]byte(nil), raw...), true, nil
	}
	if raw, ok := c.filesByRef[strings.TrimSpace(filePath)]; ok {
		return append([]byte(nil), raw...), true, nil
	}
	return nil, false, nil
}

func (c *inMemoryPushMainVersionBumpClient) ListChangedFilesBetweenCommits(_ context.Context, _ string, _ string, _ string, _ string, _ string) ([]string, error) {
	if c == nil {
		return nil, nil
	}
	if c.changedErr != nil {
		return nil, c.changedErr
	}
	return append([]string(nil), c.changedPaths...), nil
}

func (c *inMemoryPushMainVersionBumpClient) CommitFilesOnBranch(_ context.Context, _ string, owner string, repo string, branch string, baseSHA string, message string, files map[string][]byte) (string, error) {
	if c == nil {
		return "", nil
	}
	if c.commitErr != nil {
		return "", c.commitErr
	}
	c.commitCalls++
	c.lastCommitOwner = owner
	c.lastCommitRepo = repo
	c.lastCommitBranch = branch
	c.lastCommitBaseSHA = baseSHA
	c.lastCommitMessage = message
	c.lastCommitFiles = make(map[string][]byte, len(files))
	for path, raw := range files {
		c.lastCommitFiles[path] = append([]byte(nil), raw...)
	}
	return "bumped-sha", nil
}
