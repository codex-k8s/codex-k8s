package runstatus

import (
	"strings"
	"testing"
	"time"

	nextstepdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/nextstep"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
)

func mustRenderCommentBody(t *testing.T, state commentState, managementURL string, nextStepActions []nextStepCommentAction, recentStatuses []recentAgentStatus) string {
	t.Helper()

	body, err := renderCommentBody(state, managementURL, nextStepActions, recentStatuses)
	if err != nil {
		t.Fatalf("renderCommentBody returned error: %v", err)
	}
	return body
}

func assertRenderedBodyContains(t *testing.T, state commentState, managementURL string, nextStepActions []nextStepCommentAction, recentStatuses []recentAgentStatus, expected ...string) {
	t.Helper()

	body := mustRenderCommentBody(t, state, managementURL, nextStepActions, recentStatuses)
	for _, item := range expected {
		if !strings.Contains(body, item) {
			t.Fatalf("rendered body does not contain %q: %q", item, body)
		}
	}
}

func TestRenderCommentBody_RendersTemplateByLocale(t *testing.T) {
	t.Parallel()

	body := mustRenderCommentBody(t, commentState{
		RunID:        "run-1",
		Phase:        PhaseStarted,
		TriggerKind:  triggerKindDev,
		PromptLocale: localeRU,
	}, "https://platform.kodex.works/runs/run-1", nil, nil)
	if !strings.Contains(body, "### 🧠 Запуск ИИ-агента") {
		t.Fatalf("rendered body does not contain russian title: %q", body)
	}
	if !strings.Contains(body, "`run-1`") {
		t.Fatalf("rendered body does not contain run id: %q", body)
	}
}

func TestRenderCommentBody_RendersPlannedLaunchState(t *testing.T) {
	t.Parallel()

	body := mustRenderCommentBody(t, commentState{
		RunID:        "run-planned",
		Phase:        PhaseCreated,
		RuntimeMode:  runtimeModeFullEnv,
		RunStatus:    "pending",
		PromptLocale: localeRU,
	}, "https://platform.kodex.works/runs/run-planned", nil, nil)
	if !strings.Contains(body, "Планируется запуск агента") {
		t.Fatalf("rendered body does not contain planned launch marker: %q", body)
	}
	if !strings.Contains(body, "Ожидание сборки и деплоя") {
		t.Fatalf("rendered body does not contain waiting runtime preparation marker: %q", body)
	}
}

func TestRenderCommentBody_RendersSlotURLAndAuthTimeline(t *testing.T) {
	t.Parallel()

	body := mustRenderCommentBody(t, commentState{
		RunID:         "run-2",
		Phase:         PhaseReady,
		AuthRequested: true,
		RuntimeMode:   runtimeModeFullEnv,
		Namespace:     "kodex-dev-2",
		SlotURL:       "https://kodex-dev-2.ai.platform.kodex.works",
		RunStatus:     "running",
		PromptLocale:  localeRU,
	}, "https://platform.kodex.works/runs/run-2", nil, nil)
	if !strings.Contains(body, "Ссылка на слот") {
		t.Fatalf("rendered body does not contain slot url label: %q", body)
	}
	if !strings.Contains(body, "Авторизация Codex подтверждена") {
		t.Fatalf("rendered body does not contain auth resolved timeline item: %q", body)
	}
}

func TestRenderCommentBody_RendersAuthVerificationPayload(t *testing.T) {
	t.Parallel()

	assertRenderedBodyContains(t, commentState{
		RunID:                    "run-auth",
		Phase:                    PhaseAuthRequired,
		PromptLocale:             localeRU,
		CodexAuthVerificationURL: "https://example.com/device",
		CodexAuthUserCode:        "ABCD-EFGH",
	}, "https://platform.kodex.works/runs/run-auth", nil, nil, "Ссылка авторизации", "ABCD-EFGH")
}

func TestRenderCommentBody_RuntimePreparationAndNamespaceMessages(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		state         commentState
		managementURL string
		mustContain   []string
	}{
		{
			name: "runtime_preparing",
			state: commentState{
				RunID:        "run-preparing",
				Phase:        PhasePreparingRuntime,
				RuntimeMode:  runtimeModeFullEnv,
				Namespace:    "kodex-dev-2",
				RunStatus:    "running",
				PromptLocale: localeRU,
			},
			managementURL: "https://platform.kodex.works/runs/run-preparing",
			mustContain: []string{
				"Идёт сборка и деплой",
				"namespace, runtime stack, slot URL",
			},
		},
		{
			name: "namespace_kept",
			state: commentState{
				RunID:        "run-debug",
				Phase:        PhaseNamespaceDeleted,
				RuntimeMode:  runtimeModeFullEnv,
				Namespace:    "kodex-dev-2",
				RunStatus:    "succeeded",
				PromptLocale: localeRU,
			},
			managementURL: "https://platform.kodex.works/runs/run-debug",
			mustContain: []string{
				"Namespace не удален",
				"Удалить его можно на странице запуска",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			body := mustRenderCommentBody(t, testCase.state, testCase.managementURL, nil, nil)
			for _, expected := range testCase.mustContain {
				if !strings.Contains(body, expected) {
					t.Fatalf("rendered body does not contain %q: %q", expected, body)
				}
			}
		})
	}
}

func TestRenderCommentBody_RendersNextStepMatrix(t *testing.T) {
	t.Parallel()

	nextStepActions := []nextStepCommentAction{
		{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: nextStepDisplayRevise,
			TargetLabel:    "run:dev:revise",
			URL:            "https://platform.kodex.works/?modal=next-step&repository_full_name=codex-k8s%2Fkodex&issue_number=95&action_kind=issue_stage_transition&target_label=run%3Adev%3Arevise&display_variant=revise",
		},
		{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: nextStepDisplayVeryShortFlow,
			TargetLabel:    "run:qa",
			URL:            "https://platform.kodex.works/?modal=next-step&repository_full_name=codex-k8s%2Fkodex&issue_number=95&pull_request_number=123&action_kind=issue_stage_transition&target_label=run%3Aqa&display_variant=very_short_flow",
		},
		{
			ActionKind:     querytypes.NextStepActionKindPullRequestLabelAdd,
			DisplayVariant: nextStepDisplayReviewer,
			TargetLabel:    "need:reviewer",
			URL:            "https://platform.kodex.works/?modal=next-step&repository_full_name=codex-k8s%2Fkodex&issue_number=95&pull_request_number=123&action_kind=pull_request_label_add&target_label=need%3Areviewer&display_variant=reviewer",
		},
		{
			ActionKind:     querytypes.NextStepActionKindIssueStageTransition,
			DisplayVariant: nextStepDisplayDocAudit,
			TargetLabel:    "run:doc-audit",
			URL:            "https://platform.kodex.works/?modal=next-step&repository_full_name=codex-k8s%2Fkodex&issue_number=95&action_kind=issue_stage_transition&target_label=run%3Adoc-audit&display_variant=doc_audit",
		},
	}

	body := mustRenderCommentBody(t, commentState{
		RunID:              "run-dev",
		Phase:              PhaseStarted,
		TriggerKind:        triggerKindDev,
		PromptLocale:       localeRU,
		RepositoryFullName: "codex-k8s/kodex",
		IssueNumber:        95,
	}, "https://platform.kodex.works/runs/run-dev", nextStepActions, nil)

	if !strings.Contains(body, "Следующие шаги") {
		t.Fatalf("expected next steps section in body: %q", body)
	}
	if !strings.Contains(body, "`run:dev:revise`") || !strings.Contains(body, "`run:qa`") || !strings.Contains(body, "`need:reviewer`") {
		t.Fatalf("expected next step labels in body: %q", body)
	}
	if !strings.Contains(body, "/?modal=next-step") {
		t.Fatalf("expected root next-step deep-link in body: %q", body)
	}
	if strings.Contains(body, "/governance/labels-stages?") {
		t.Fatalf("expected legacy labels-stages link to be removed: %q", body)
	}
	if strings.Contains(body, "Контракт next-step action-card") || strings.Contains(body, "`fallback_action`") || strings.Contains(body, "`launch_profile`") {
		t.Fatalf("expected legacy action-card contract to be removed: %q", body)
	}
}

func TestRenderCommentBody_RendersDesignFastTrackAction(t *testing.T) {
	t.Parallel()

	body := mustRenderCommentBody(t, commentState{
		RunID:              "run-design",
		Phase:              PhaseStarted,
		TriggerKind:        "design",
		PromptLocale:       localeRU,
		RepositoryFullName: "codex-k8s/kodex",
		IssueNumber:        95,
	}, "https://platform.kodex.works/runs/run-design", buildNextStepActions("https://platform.kodex.works", nextstepdomain.DefaultLabels(), runContext{}, commentState{
		TriggerKind:        "design",
		RepositoryFullName: "codex-k8s/kodex",
		IssueNumber:        95,
	}), nil)
	if !strings.Contains(body, "`run:dev`") {
		t.Fatalf("expected fast-track run:dev action label in body: %q", body)
	}
	if !strings.Contains(body, "target_label=run%3Adev") {
		t.Fatalf("expected fast-track deep-link target in body: %q", body)
	}
}

func TestRenderCommentBody_RendersIssueAndPRLinks(t *testing.T) {
	t.Parallel()

	assertRenderedBodyContains(t, commentState{
		RunID:          "run-links",
		Phase:          PhaseStarted,
		PromptLocale:   localeRU,
		IssueURL:       "https://github.com/codex-k8s/kodex/issues/95",
		PullRequestURL: "https://github.com/codex-k8s/kodex/pull/123",
	}, "https://platform.kodex.works/runs/run-links", nil, nil, "issues/95", "pull/123")
}

func TestRenderCommentBody_RendersRecentAgentStatusesRU(t *testing.T) {
	t.Parallel()

	body := mustRenderCommentBody(t, commentState{
		RunID:        "run-statuses-ru",
		Phase:        PhaseStarted,
		PromptLocale: localeRU,
	}, "https://platform.kodex.works/runs/run-statuses-ru", nil, []recentAgentStatus{
		{StatusText: "Обновляю API", ReportedAt: nowUTC().Format(time.RFC3339Nano)},
		{StatusText: "Проверяю тесты", ReportedAt: nowUTC().Format(time.RFC3339Nano), RepeatCount: 2},
	})
	if !strings.Contains(body, "Последние статусы агента") {
		t.Fatalf("expected recent agent statuses section in body: %q", body)
	}
	if !strings.Contains(body, "Проверяю тесты") {
		t.Fatalf("expected status text in body: %q", body)
	}
	if strings.Contains(body, "agent_key") || strings.Contains(body, "`dev`:") {
		t.Fatalf("expected agent key to be hidden in body: %q", body)
	}
	if !strings.Contains(body, "(x2)") {
		t.Fatalf("expected dedupe marker in body: %q", body)
	}
}

func TestRenderCommentBody_RendersRecentAgentStatusesEN(t *testing.T) {
	t.Parallel()

	body := mustRenderCommentBody(t, commentState{
		RunID:        "run-statuses-en",
		Phase:        PhaseStarted,
		PromptLocale: localeEN,
	}, "https://platform.kodex.works/runs/run-statuses-en", nil, []recentAgentStatus{
		{StatusText: "Running regression", ReportedAt: nowUTC().Format(time.RFC3339Nano)},
	})
	if !strings.Contains(body, "Latest Agent Statuses") {
		t.Fatalf("expected recent agent statuses section in body: %q", body)
	}
	if !strings.Contains(body, "Running regression") {
		t.Fatalf("expected status text in body: %q", body)
	}
}

func TestRenderCommentBody_RendersDiscussionRunAsPod(t *testing.T) {
	t.Parallel()

	body := mustRenderCommentBody(t, commentState{
		RunID:        "run-discussion",
		Phase:        PhaseStarted,
		TriggerKind:  triggerKindDev,
		TriggerLabel: "mode:discussion",
		RuntimeMode:  runtimeModeCode,
		JobName:      "kodex-run-run-discussion",
		JobNamespace: "codex-issue-demo",
		Namespace:    "codex-issue-demo",
		PromptLocale: localeRU,
	}, "https://platform.kodex.works/runs/run-discussion", nil, nil)

	if !strings.Contains(body, "Режим запуска: `discussion`") {
		t.Fatalf("expected discussion trigger display in body: %q", body)
	}
	if !strings.Contains(body, "Runtime mode: `code-only`") {
		t.Fatalf("expected code-only runtime mode in body: %q", body)
	}
	if !strings.Contains(body, "Pod: `codex-issue-demo/kodex-run-run-discussion`") {
		t.Fatalf("expected pod workload reference in body: %q", body)
	}
	if strings.Contains(body, "Job: `codex-issue-demo/kodex-run-run-discussion`") {
		t.Fatalf("expected job workload reference to be hidden for discussion run: %q", body)
	}
	if strings.Contains(body, "Ожидание подготовки окружения") {
		t.Fatalf("expected runtime preparation timeline to be hidden for discussion run: %q", body)
	}
}
