package runstatus

import (
	"strings"
	"testing"
)

func TestRenderTriggerWarningCommentBody_RU(t *testing.T) {
	t.Parallel()

	body, err := renderTriggerWarningCommentBody(triggerWarningRenderParams{
		Locale:     localeRU,
		ThreadKind: string(commentTargetKindPullRequest),
		ReasonCode: "pull_request_review_missing_stage_label",
	})
	if err != nil {
		t.Fatalf("renderTriggerWarningCommentBody() error = %v", err)
	}
	if !strings.Contains(body, "Запуск не создан") {
		t.Fatalf("missing ru title in body: %q", body)
	}
	if !strings.Contains(body, "pull_request_review_missing_stage_label") {
		t.Fatalf("missing reason code in body: %q", body)
	}
}
