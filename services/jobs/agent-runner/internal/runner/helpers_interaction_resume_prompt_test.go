package runner

import (
	"fmt"
	"strings"
	"testing"

	"github.com/codex-k8s/codex-k8s/libs/go/mcp/userinteraction"
)

func TestBuildInteractionResumePromptBlock_ValidatesResumeContext(t *testing.T) {
	t.Parallel()

	_, err := buildInteractionResumePromptBlock(promptLocaleRU, `{"interaction_id":"interaction-1"}`, false)
	if err == nil {
		t.Fatal("expected error when interaction resume payload is provided without restored session")
	}
	if !strings.Contains(err.Error(), "requires restored codex session") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildPrompt_PrependsInteractionResumePromptBlock(t *testing.T) {
	t.Parallel()

	service := &Service{
		cfg: Config{
			RunID:                    "run-resume",
			RepositoryFullName:       "codex-k8s/codex-k8s",
			AgentKey:                 "dev",
			IssueNumber:              394,
			RuntimeMode:              runtimeModeFullEnv,
			InteractionResumePayload: `{"interaction_id":"interaction-1","tool_name":"user.decision.request","request_status":"answered","response_kind":"option","selected_option_id":"approve","resolved_at":"2026-03-13T16:05:00Z","resolution_reason":"accepted"}`,
			PromptConfig: PromptConfig{
				TriggerKind:          "dev",
				TriggerLabel:         "run:dev",
				StateInReviewLabel:   "state:in-review",
				PromptTemplateLocale: promptLocaleRU,
				AgentBaseBranch:      "main",
			},
		},
	}

	prompt, err := service.buildPrompt("task body", runResult{
		targetBranch:        "codex/issue-394",
		triggerKind:         "dev",
		templateKind:        promptTemplateKindWork,
		restoredSessionPath: "/tmp/restored-session.json",
	}, t.TempDir())
	if err != nil {
		t.Fatalf("buildPrompt() error = %v", err)
	}
	if !strings.HasPrefix(prompt, "Детерминированный resume context:") {
		t.Fatalf("prompt must start with interaction resume block, got: %q", prompt)
	}
	if !strings.Contains(prompt, `"interaction_id": "interaction-1"`) {
		t.Fatalf("prompt must contain pretty-printed interaction resume payload, got: %q", prompt)
	}
	if !strings.Contains(prompt, "не задавайте пользователю тот же вопрос повторно") {
		t.Fatalf("prompt must instruct deterministic resume behavior, got: %q", prompt)
	}
}

func TestParseInteractionResumePayload_RejectsOversizedPayload(t *testing.T) {
	t.Parallel()

	rawPayload := fmt.Sprintf(
		`{"interaction_id":"interaction-1","tool_name":"user.decision.request","request_status":"answered","response_kind":"free_text","free_text":"%s","resolved_at":"2026-03-13T16:05:00Z","resolution_reason":"accepted"}`,
		strings.Repeat("a", userinteraction.ResumePayloadMaxBytes),
	)

	_, err := parseInteractionResumePayload(rawPayload)
	if err == nil {
		t.Fatal("expected oversized interaction resume payload error")
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("unexpected error: %v", err)
	}
}
