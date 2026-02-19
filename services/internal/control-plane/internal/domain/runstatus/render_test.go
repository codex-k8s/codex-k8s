package runstatus

import (
	"strings"
	"testing"
)

func TestRenderCommentBody_RendersTemplateByLocale(t *testing.T) {
	t.Parallel()

	body, err := renderCommentBody(commentState{
		RunID:        "run-1",
		Phase:        PhaseStarted,
		TriggerKind:  triggerKindDev,
		PromptLocale: localeRU,
	}, "https://platform.codex-k8s.dev/runs/run-1")
	if err != nil {
		t.Fatalf("renderCommentBody returned error: %v", err)
	}
	if !strings.Contains(body, "### 🧠 Запуск ИИ-агента") {
		t.Fatalf("rendered body does not contain russian title: %q", body)
	}
	if !strings.Contains(body, "`run-1`") {
		t.Fatalf("rendered body does not contain run id: %q", body)
	}
}

func TestRenderCommentBody_RendersSlotURLAndAuthTimeline(t *testing.T) {
	t.Parallel()

	body, err := renderCommentBody(commentState{
		RunID:        "run-2",
		Phase:        PhaseAuthResolved,
		RuntimeMode:  runtimeModeFullEnv,
		Namespace:    "codex-k8s-dev-2",
		SlotURL:      "https://codex-k8s-dev-2.ai.platform.codex-k8s.dev",
		RunStatus:    "running",
		PromptLocale: localeRU,
	}, "https://platform.codex-k8s.dev/runs/run-2")
	if err != nil {
		t.Fatalf("renderCommentBody returned error: %v", err)
	}
	if !strings.Contains(body, "Ссылка на слот") {
		t.Fatalf("rendered body does not contain slot url label: %q", body)
	}
	if !strings.Contains(body, "Авторизация Codex подтверждена") {
		t.Fatalf("rendered body does not contain auth resolved timeline item: %q", body)
	}
}

func TestRenderCommentBody_RendersAuthVerificationPayload(t *testing.T) {
	t.Parallel()

	body, err := renderCommentBody(commentState{
		RunID:                    "run-auth",
		Phase:                    PhaseAuthRequired,
		PromptLocale:             localeRU,
		CodexAuthVerificationURL: "https://example.com/device",
		CodexAuthUserCode:        "ABCD-EFGH",
	}, "https://platform.codex-k8s.dev/runs/run-auth")
	if err != nil {
		t.Fatalf("renderCommentBody returned error: %v", err)
	}
	if !strings.Contains(body, "Ссылка авторизации") {
		t.Fatalf("rendered body does not contain auth verification url label: %q", body)
	}
	if !strings.Contains(body, "ABCD-EFGH") {
		t.Fatalf("rendered body does not contain auth user code: %q", body)
	}
}
