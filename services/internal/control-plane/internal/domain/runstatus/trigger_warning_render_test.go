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
		ReasonCode: TriggerWarningReasonPullRequestReviewMissingStageLabel,
	})
	if err != nil {
		t.Fatalf("renderTriggerWarningCommentBody() error = %v", err)
	}
	if !strings.Contains(body, "Запуск не создан") {
		t.Fatalf("missing ru title in body: %q", body)
	}
	if !strings.Contains(body, string(TriggerWarningReasonPullRequestReviewMissingStageLabel)) {
		t.Fatalf("missing reason code in body: %q", body)
	}
}
