package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

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
	if runPayload.Agent.Key != "dev" {
		t.Fatalf("unexpected agent key: %#v", runPayload.Agent.Key)
	}
	if runPayload.Agent.Name != "AI Developer" {
		t.Fatalf("unexpected agent name: %#v", runPayload.Agent.Name)
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

func TestIngestGitHubWebhook_PullRequestReviewChangesRequested_CreatesReviseRun(t *testing.T) {
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
	if got.Status != webhookdomain.IngestStatusAccepted || got.Duplicate {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got.RunID == "" {
		t.Fatalf("expected run id for pull_request_review trigger")
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

func TestIngestGitHubWebhook_IssueRunDev_DeniesUnknownSender(t *testing.T) {
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
	svc := NewService(Config{
		AgentRuns:  runs,
		Agents:     agents,
		FlowEvents: events,
		Repos:      repos,
		Users:      &inMemoryUserRepo{},
		Members:    &inMemoryProjectMemberRepo{},
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
	items map[string]string
	last  agentrunrepo.CreateParams
}

func (r *inMemoryRunRepo) CreatePendingIfAbsent(_ context.Context, params agentrunrepo.CreateParams) (agentrunrepo.CreateResult, error) {
	r.last = params
	if id, ok := r.items[params.CorrelationID]; ok {
		return agentrunrepo.CreateResult{
			RunID:    id,
			Inserted: false,
		}, nil
	}
	id := "run-" + params.CorrelationID
	r.items[params.CorrelationID] = id
	return agentrunrepo.CreateResult{
		RunID:    id,
		Inserted: true,
	}, nil
}

func (r *inMemoryRunRepo) GetByID(_ context.Context, runID string) (agentrunrepo.Run, bool, error) {
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

func (r *inMemoryRunRepo) ListRunIDsByRepositoryIssue(_ context.Context, _ string, _ int64, _ int) ([]string, error) {
	return nil, nil
}

func (r *inMemoryRunRepo) ListRunIDsByRepositoryPullRequest(_ context.Context, _ string, _ int64, _ int) ([]string, error) {
	return nil, nil
}

type inMemoryRunStatusService struct {
	issueCleanupCalls       int
	pullRequestCleanupCalls int
	lastIssueCleanup        runstatusdomain.CleanupByIssueParams
	lastPullRequestCleanup  runstatusdomain.CleanupByPullRequestParams
	conflictCommentCalls    int
	lastConflictComment     runstatusdomain.TriggerLabelConflictCommentParams
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

func (r *inMemoryRepoCfgRepo) GetTokenEncrypted(_ context.Context, _ string) ([]byte, bool, error) {
	return nil, false, nil
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
