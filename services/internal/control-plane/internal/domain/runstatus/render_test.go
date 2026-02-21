package runstatus

import (
	"strings"
	"testing"
)

func mustRenderCommentBody(t *testing.T, state commentState, managementURL string) string {
	t.Helper()

	body, err := renderCommentBody(state, managementURL)
	if err != nil {
		t.Fatalf("renderCommentBody returned error: %v", err)
	}
	return body
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

	body := mustRenderCommentBody(t, commentState{
		RunID:                    "run-auth",
		Phase:                    PhaseAuthRequired,
		PromptLocale:             localeRU,
		CodexAuthVerificationURL: "https://example.com/device",
		CodexAuthUserCode:        "ABCD-EFGH",
	}, "https://platform.codex-k8s.dev/runs/run-auth")
	if !strings.Contains(body, "Ссылка авторизации") {
		t.Fatalf("rendered body does not contain auth verification url label: %q", body)
	}
	if !strings.Contains(body, "ABCD-EFGH") {
		t.Fatalf("rendered body does not contain auth user code: %q", body)
	}
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
		RunID:        "run-dev",
		Phase:        PhaseStarted,
		TriggerKind:  triggerKindDev,
		PromptLocale: localeRU,
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
}

func TestRenderCommentBody_RendersIssueAndPRLinks(t *testing.T) {
	t.Parallel()

	body := mustRenderCommentBody(t, commentState{
		RunID:          "run-links",
		Phase:          PhaseStarted,
		PromptLocale:   localeRU,
		IssueURL:       "https://github.com/codex-k8s/codex-k8s/issues/95",
		PullRequestURL: "https://github.com/codex-k8s/codex-k8s/pull/123",
	}, "https://platform.codex-k8s.dev/runs/run-links")
	if !strings.Contains(body, "issues/95") {
		t.Fatalf("expected issue link in body: %q", body)
	}
	if !strings.Contains(body, "pull/123") {
		t.Fatalf("expected pr link in body: %q", body)
	}
}
