package service

import (
	"strings"
	"testing"
)

func TestMessageRenderer_RendersDecisionTemplateWithLinks(t *testing.T) {
	t.Parallel()

	renderer, err := newMessageRenderer()
	if err != nil {
		t.Fatalf("newMessageRenderer() error = %v", err)
	}

	text := renderer.Render("ru", "decision_message", decisionMessageData{
		Question:         "Подтвердить выкладку?",
		DetailsMarkdown:  "Нужно подтвердить rollout в candidate.",
		ReplyInstruction: "Можно ответить сообщением в этот чат.",
		Links: messageLinks{
			RunURL:         "https://example.test/run/1",
			IssueURL:       "https://example.test/issues/473",
			PullRequestURL: "https://example.test/pr/483",
		},
	})

	for _, want := range []string{
		"🧠 Подтвердить выкладку?",
		"✍️ Можно ответить сообщением в этот чат.",
		"🏃 Run: https://example.test/run/1",
		"🐞 Issue: https://example.test/issues/473",
		"🔀 PR: https://example.test/pr/483",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("rendered decision message %q does not contain %q", text, want)
		}
	}
}

func TestDecorateOptionLabel_UsesSemanticEmoji(t *testing.T) {
	t.Parallel()

	if got := decorateOptionLabel(DecisionOption{OptionID: "approve", Label: "Одобрить"}, 0); got != "✅ Одобрить" {
		t.Fatalf("approve label = %q", got)
	}
	if got := decorateOptionLabel(DecisionOption{OptionID: "reject", Label: "Отклонить"}, 1); got != "❌ Отклонить" {
		t.Fatalf("reject label = %q", got)
	}
	if got := decorateOptionLabel(DecisionOption{OptionID: "later", Label: "Позже"}, 2); got != "⏳ Позже" {
		t.Fatalf("later label = %q", got)
	}
}
