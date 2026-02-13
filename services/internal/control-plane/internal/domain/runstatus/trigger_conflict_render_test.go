package runstatus

import (
	"strings"
	"testing"
)

func TestRenderTriggerLabelConflictCommentBody_RU(t *testing.T) {
	t.Parallel()

	body, err := renderTriggerLabelConflictCommentBody(localeRU, "run:vision", []string{"run:vision", "run:dev"})
	if err != nil {
		t.Fatalf("renderTriggerLabelConflictCommentBody() error = %v", err)
	}
	if !strings.Contains(body, "Конфликт trigger-лейблов") {
		t.Fatalf("unexpected body: %q", body)
	}
	if !strings.Contains(body, "`run:dev`") || !strings.Contains(body, "`run:vision`") {
		t.Fatalf("labels are missing in body: %q", body)
	}
}

func TestRenderTriggerLabelConflictCommentBody_EN(t *testing.T) {
	t.Parallel()

	body, err := renderTriggerLabelConflictCommentBody(localeEN, "run:plan", []string{"run:plan", "run:ops"})
	if err != nil {
		t.Fatalf("renderTriggerLabelConflictCommentBody() error = %v", err)
	}
	if !strings.Contains(body, "Trigger Label Conflict") {
		t.Fatalf("unexpected body: %q", body)
	}
	if !strings.Contains(body, "`run:ops`") {
		t.Fatalf("expected conflicting label in body: %q", body)
	}
}
