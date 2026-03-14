package runner

import (
	"strings"
	"testing"
)

func TestBuildGitHubRateLimitResumePromptBlock_ValidatesResumeContext(t *testing.T) {
	t.Parallel()

	_, err := buildGitHubRateLimitResumePromptBlock(promptLocaleRU, `{"wait_id":"wait-1"}`, false)
	if err == nil {
		t.Fatal("expected error when github rate-limit resume payload is provided without restored session")
	}
	if !strings.Contains(err.Error(), "requires restored codex session") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildPrompt_PrependsGitHubRateLimitResumePromptBlock(t *testing.T) {
	t.Parallel()

	service := &Service{
		cfg: Config{
			RunID:                        "run-resume",
			RepositoryFullName:           "codex-k8s/codex-k8s",
			AgentKey:                     "dev",
			IssueNumber:                  428,
			RuntimeMode:                  runtimeModeFullEnv,
			GitHubRateLimitResumePayload: `{"wait_id":"wait-1","wait_reason":"github_rate_limit","contour_kind":"agent_bot_token","limit_kind":"secondary","resolution_kind":"auto_resumed","recovered_at":"2026-03-14T17:00:00Z","attempt_no":2,"affected_operation_class":"agent_github_call","guidance":"resume from the persisted snapshot"}`,
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
		targetBranch:        "codex/issue-428",
		triggerKind:         "dev",
		templateKind:        promptTemplateKindWork,
		restoredSessionPath: "/tmp/restored-session.json",
	}, t.TempDir())
	if err != nil {
		t.Fatalf("buildPrompt() error = %v", err)
	}
	if !strings.HasPrefix(prompt, "Детерминированный resume context (GitHub rate-limit wait):") {
		t.Fatalf("prompt must start with github rate-limit resume block, got: %q", prompt)
	}
	if !strings.Contains(prompt, `"wait_id": "wait-1"`) {
		t.Fatalf("prompt must contain pretty-printed github rate-limit payload, got: %q", prompt)
	}
	if !strings.Contains(prompt, "единственный authoritative source для resume semantics") {
		t.Fatalf("prompt must instruct deterministic github wait resume behavior, got: %q", prompt)
	}
}

func TestParseGitHubRateLimitResumePayload_RejectsInvalidWaitReason(t *testing.T) {
	t.Parallel()

	_, err := parseGitHubRateLimitResumePayload(`{"wait_id":"wait-1","wait_reason":"other","contour_kind":"agent_bot_token","limit_kind":"secondary","resolution_kind":"auto_resumed","recovered_at":"2026-03-14T17:00:00Z","attempt_no":2,"affected_operation_class":"agent_github_call","guidance":"resume from the persisted snapshot"}`)
	if err == nil {
		t.Fatal("expected invalid wait_reason error")
	}
	if !strings.Contains(err.Error(), "wait_reason") {
		t.Fatalf("unexpected error: %v", err)
	}
}
