package runstatus

import (
	"strings"
	"testing"
)

func mustRenderCommentBody(t *testing.T, state commentState, managementURL string) string {
	t.Helper()

	body, err := renderCommentBody(state, managementURL, "https://platform.codex-k8s.dev")
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
