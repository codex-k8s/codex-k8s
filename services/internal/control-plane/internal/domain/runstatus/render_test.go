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
