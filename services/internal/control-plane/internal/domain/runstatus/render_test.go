package runstatus

import (
	"strings"
	"testing"
)

func mustRenderCommentBody(t *testing.T, state commentState, managementURL string) string {
	t.Helper()

	body, err := renderCommentBody(state, managementURL, "https://platform.codex-k8s.dev", nil)
	if err != nil {
		t.Fatalf("renderCommentBody returned error: %v", err)
	}
	return body
}

func assertRenderedBodyContains(t *testing.T, state commentState, managementURL string, expected ...string) {
	t.Helper()

	body := mustRenderCommentBody(t, state, managementURL)
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
	}, "https://platform.codex-k8s.dev/runs/run-1")
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
	}, "https://platform.codex-k8s.dev/runs/run-planned")
	if !strings.Contains(body, "Планируется запуск агента") {
		t.Fatalf("rendered body does not contain planned launch marker: %q", body)
	}
	if !strings.Contains(body, "Ожидание подготовки окружения") {
		t.Fatalf("rendered body does not contain waiting runtime preparation marker: %q", body)
	}
}

func TestRenderCommentBody_RendersSlotURLAndAuthTimeline(t *testing.T) {
	t.Parallel()

	body := mustRenderCommentBody(t, commentState{
		RunID:        "run-2",
		Phase:        PhaseAuthResolved,
		RuntimeMode:  runtimeModeFullEnv,
		Namespace:    "codex-k8s-dev-2",
		SlotURL:      "https://codex-k8s-dev-2.ai.platform.codex-k8s.dev",
		RunStatus:    "running",
		PromptLocale: localeRU,
	}, "https://platform.codex-k8s.dev/runs/run-2")
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
	}, "https://platform.codex-k8s.dev/runs/run-auth", "Ссылка авторизации", "ABCD-EFGH")
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
				Namespace:    "codex-k8s-dev-2",
				RunStatus:    "running",
				PromptLocale: localeRU,
			},
			managementURL: "https://platform.codex-k8s.dev/runs/run-preparing",
			mustContain: []string{
				"Идёт подготовка окружения",
				"namespace, runtime stack, slot URL",
			},
		},
		{
			name: "namespace_kept",
			state: commentState{
				RunID:        "run-debug",
				Phase:        PhaseNamespaceDeleted,
				RuntimeMode:  runtimeModeFullEnv,
				Namespace:    "codex-k8s-dev-2",
				RunStatus:    "succeeded",
				PromptLocale: localeRU,
			},
			managementURL: "https://platform.codex-k8s.dev/runs/run-debug",
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

			body := mustRenderCommentBody(t, testCase.state, testCase.managementURL)
			for _, expected := range testCase.mustContain {
				if !strings.Contains(body, expected) {
					t.Fatalf("rendered body does not contain %q: %q", expected, body)
				}
			}
		})
	}
}

func TestRenderCommentBody_RendersStageAwareActionCards(t *testing.T) {
	t.Parallel()

	body := mustRenderCommentBody(t, commentState{
		RunID:              "run-dev",
		Phase:              PhaseStarted,
		TriggerKind:        triggerKindDev,
		PromptLocale:       localeRU,
		RepositoryFullName: "codex-k8s/codex-k8s",
		IssueNumber:        95,
		LaunchProfile:      "quick-fix",
		StagePath:          "intake -> plan -> dev -> qa -> release -> postdeploy -> ops",
		PrimaryAction:      "https://platform.codex-k8s.dev/governance/labels-stages?repo=codex-k8s%2Fcodex-k8s&issue=95&target=run%3Aqa",
		FallbackAction: "gh issue view 95 --json labels --jq '.labels[].name'\n" +
			`gh issue edit 95 --remove-label "run:dev" --remove-label "run:dev:revise" --add-label "run:qa"`,
		GuardrailNote: guardrailNotePrecheckRequired,
	}, "https://platform.codex-k8s.dev/runs/run-dev")
	if !strings.Contains(body, "Следующие шаги") {
		t.Fatalf("expected next steps section in body: %q", body)
	}
	if !strings.Contains(body, "`run:dev:revise`") {
		t.Fatalf("expected revise action label in body: %q", body)
	}
	if !strings.Contains(body, "`run:qa`") {
		t.Fatalf("expected next stage action label in body: %q", body)
	}
	if !strings.Contains(body, "/governance/labels-stages?") {
		t.Fatalf("expected deep-link action url in body: %q", body)
	}
	if !strings.Contains(body, "target=run%3Aqa") {
		t.Fatalf("expected deep-link target label in body: %q", body)
	}
	if !strings.Contains(body, "`launch_profile`") || !strings.Contains(body, "`quick-fix`") {
		t.Fatalf("expected launch_profile contract field in body: %q", body)
	}
	if !strings.Contains(body, "`fallback_action`") {
		t.Fatalf("expected fallback_action contract field in body: %q", body)
	}
	if !strings.Contains(body, `gh issue edit 95 --remove-label "run:dev" --remove-label "run:dev:revise" --add-label "run:qa"`) {
		t.Fatalf("expected fallback transition command in body: %q", body)
	}
	if !strings.Contains(body, "`guardrail_note`") {
		t.Fatalf("expected guardrail_note contract field in body: %q", body)
	}
}

func TestRenderCommentBody_RendersDesignFastTrackAction(t *testing.T) {
	t.Parallel()

	body := mustRenderCommentBody(t, commentState{
		RunID:              "run-design",
		Phase:              PhaseStarted,
		TriggerKind:        "design",
		PromptLocale:       localeRU,
		RepositoryFullName: "codex-k8s/codex-k8s",
		IssueNumber:        95,
	}, "https://platform.codex-k8s.dev/runs/run-design")
	if !strings.Contains(body, "`run:dev`") {
		t.Fatalf("expected fast-track run:dev action label in body: %q", body)
	}
	if !strings.Contains(body, "target=run%3Adev") {
		t.Fatalf("expected fast-track deep-link target in body: %q", body)
	}
}

func TestRenderCommentBody_BlocksNextStageActionsWhenActionCardGuardrailRequiresNeedInput(t *testing.T) {
	t.Parallel()

	body := mustRenderCommentBody(t, commentState{
		RunID:              "run-dev-ambiguous",
		Phase:              PhaseStarted,
		TriggerKind:        triggerKindDev,
		PromptLocale:       localeRU,
		RepositoryFullName: "codex-k8s/codex-k8s",
		IssueNumber:        95,
		LaunchProfile:      "quick-fix",
		StagePath:          "intake -> plan -> dev -> qa -> release -> postdeploy -> ops",
		FallbackAction:     `gh issue edit 95 --add-label "need:input"`,
		GuardrailNote:      guardrailNoteAmbiguousStageLabel,
	}, "https://platform.codex-k8s.dev/runs/run-dev-ambiguous")

	if strings.Contains(body, "`run:qa`") {
		t.Fatalf("expected stage transition links to be hidden for blocked action card: %q", body)
	}
	if !strings.Contains(body, `gh issue edit 95 --add-label "need:input"`) {
		t.Fatalf("expected need:input remediation command in body: %q", body)
	}
}

func TestRenderCommentBody_RendersIssueAndPRLinks(t *testing.T) {
	t.Parallel()

	assertRenderedBodyContains(t, commentState{
		RunID:          "run-links",
		Phase:          PhaseStarted,
		PromptLocale:   localeRU,
		IssueURL:       "https://github.com/codex-k8s/codex-k8s/issues/95",
		PullRequestURL: "https://github.com/codex-k8s/codex-k8s/pull/123",
	}, "https://platform.codex-k8s.dev/runs/run-links", "issues/95", "pull/123")
}

func TestRenderCommentBody_RendersRecentAgentStatusesRU(t *testing.T) {
	t.Parallel()

	body, err := renderCommentBody(commentState{
		RunID:        "run-statuses-ru",
		Phase:        PhaseStarted,
		PromptLocale: localeRU,
	}, "https://platform.codex-k8s.dev/runs/run-statuses-ru", "https://platform.codex-k8s.dev", []recentAgentStatus{
		{AgentKey: "dev", StatusText: "Обновляю API"},
		{AgentKey: "dev", StatusText: "Проверяю тесты"},
	})
	if err != nil {
		t.Fatalf("renderCommentBody returned error: %v", err)
	}
	if !strings.Contains(body, "Последние статусы агента") {
		t.Fatalf("expected recent agent statuses section in body: %q", body)
	}
	if !strings.Contains(body, "Проверяю тесты") {
		t.Fatalf("expected status text in body: %q", body)
	}
}

func TestRenderCommentBody_RendersRecentAgentStatusesEN(t *testing.T) {
	t.Parallel()

	body, err := renderCommentBody(commentState{
		RunID:        "run-statuses-en",
		Phase:        PhaseStarted,
		PromptLocale: localeEN,
	}, "https://platform.codex-k8s.dev/runs/run-statuses-en", "https://platform.codex-k8s.dev", []recentAgentStatus{
		{AgentKey: "qa", StatusText: "Running regression"},
	})
	if err != nil {
		t.Fatalf("renderCommentBody returned error: %v", err)
	}
	if !strings.Contains(body, "Latest Agent Statuses") {
		t.Fatalf("expected recent agent statuses section in body: %q", body)
	}
	if !strings.Contains(body, "Running regression") {
		t.Fatalf("expected status text in body: %q", body)
	}
}
