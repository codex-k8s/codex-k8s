package runner

import (
	"strings"
	"testing"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

func TestRenderPromptArtifactContractBlocks_UsesFullIssueURLAndRoleSpecificSections(t *testing.T) {
	t.Parallel()

	issueBlock, prBlock, err := renderPromptArtifactContractBlocks("codex-k8s/codex-k8s", 253, "dev", "dev", promptTemplateKindWork, promptLocaleRU)
	if err != nil {
		t.Fatalf("renderPromptArtifactContractBlocks() error = %v", err)
	}
	if !strings.Contains(issueBlock, "Dev follow-up") {
		t.Fatalf("issue contract must contain dev issue pattern, got: %q", issueBlock)
	}
	if !strings.Contains(prBlock, "## Логи и runtime-диагностика") {
		t.Fatalf("pr contract must contain dev-specific diagnostics section, got: %q", prBlock)
	}
	if !strings.Contains(prBlock, "Closes https://github.com/codex-k8s/codex-k8s/issues/253") {
		t.Fatalf("pr contract must contain full issue URL closes directive, got: %q", prBlock)
	}
}

func TestBuildPrompt_EmbedsRenderedPromptBlocks(t *testing.T) {
	t.Parallel()

	service := &Service{
		cfg: Config{
			RunID:              "run-123",
			RepositoryFullName: "codex-k8s/codex-k8s",
			AgentKey:           "qa",
			IssueNumber:        246,
			RuntimeMode:        runtimeModeCodeOnly,
			PromptConfig: PromptConfig{
				TriggerKind:          "qa",
				TriggerLabel:         "run:qa",
				StateInReviewLabel:   "state:in-review",
				PromptTemplateLocale: promptLocaleRU,
				AgentBaseBranch:      "main",
			},
		},
	}

	prompt, err := service.buildPrompt("task body", runResult{targetBranch: "codex/issue-246", triggerKind: "qa", templateKind: promptTemplateKindWork}, t.TempDir())
	if err != nil {
		t.Fatalf("buildPrompt() error = %v", err)
	}
	if !strings.Contains(prompt, "Ролевой профиль:") {
		t.Fatalf("prompt must include rendered role profile block, got: %q", prompt)
	}
	if !strings.Contains(prompt, "Контракт оформления follow-up Issue:") {
		t.Fatalf("prompt must include issue contract block, got: %q", prompt)
	}
	if !strings.Contains(prompt, "## Тестовые сценарии и запросы") {
		t.Fatalf("prompt must include qa-specific PR contract section, got: %q", prompt)
	}
}

func TestBuildPrompt_DiscussionIncludesDiscussionContinuationContract(t *testing.T) {
	t.Parallel()

	service := &Service{
		cfg: Config{
			RunID:              "run-discussion",
			RepositoryFullName: "codex-k8s/codex-k8s",
			AgentKey:           "dev",
			IssueNumber:        289,
			RuntimeMode:        runtimeModeCodeOnly,
			PromptConfig: PromptConfig{
				TriggerKind:          "dev",
				TriggerLabel:         webhookdomain.DefaultModeDiscussionLabel,
				DiscussionMode:       true,
				PromptTemplateKind:   promptTemplateKindDiscussion,
				PromptTemplateLocale: promptLocaleRU,
				AgentBaseBranch:      "main",
			},
		},
	}

	prompt, err := service.buildPrompt("task body", runResult{
		targetBranch: "codex/issue-289",
		triggerKind:  "dev",
		templateKind: promptTemplateKindDiscussion,
	}, t.TempDir())
	if err != nil {
		t.Fatalf("buildPrompt() error = %v", err)
	}
	if !strings.Contains(prompt, "Каждый новый человеческий комментарий") {
		t.Fatalf("prompt must include discussion continuation contract, got: %q", prompt)
	}
	if !strings.Contains(prompt, "публикуйте его под Issue #289 через `gh issue comment`") {
		t.Fatalf("prompt must keep issue comment completion requirement, got: %q", prompt)
	}
}
