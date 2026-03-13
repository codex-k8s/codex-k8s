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
	if !strings.Contains(issueBlock, "Dev follow-up[ Sprint S<спринт> Day<день>]") {
		t.Fatalf("issue contract must describe optional sprint/day placement, got: %q", issueBlock)
	}
	if !strings.Contains(prBlock, "## Логи и runtime-диагностика") {
		t.Fatalf("pr contract must contain dev-specific diagnostics section, got: %q", prBlock)
	}
	if !strings.Contains(prBlock, "Closes https://github.com/codex-k8s/codex-k8s/issues/253") {
		t.Fatalf("pr contract must contain full issue URL closes directive, got: %q", prBlock)
	}
}

func TestRenderPromptArtifactContractBlocks_WorkPatternsKeepOptionalSprintDayForRoleTitles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		agentKey       string
		locale         string
		triggerKind    string
		expectedIssue  string
		expectedPR     string
	}{
		{
			name:          "qa en",
			agentKey:      "qa",
			locale:        promptLocaleEN,
			triggerKind:   "qa",
			expectedIssue: "QA gap[ Sprint S<sprint> Day<day>]",
			expectedPR:    "Issue #246: qa[ Sprint S<sprint> Day<day>] — <short verification result> (#246)",
		},
		{
			name:          "sre ru",
			agentKey:      "sre",
			locale:        promptLocaleRU,
			triggerKind:   "ops",
			expectedIssue: "SRE remediation[ Sprint S<спринт> Day<день>]",
			expectedPR:    "Issue #246: sre[ Sprint S<спринт> Day<день>] — <краткий итог remediation> (#246)",
		},
		{
			name:          "default ru",
			agentKey:      "unknown",
			locale:        promptLocaleRU,
			triggerKind:   "plan",
			expectedIssue: "default[ Sprint S<спринт> Day<день>]",
			expectedPR:    "Issue #246: plan-package — <краткая цель пакета>",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			issueBlock, prBlock, err := renderPromptArtifactContractBlocks("codex-k8s/codex-k8s", 246, tc.agentKey, tc.triggerKind, promptTemplateKindWork, tc.locale)
			if err != nil {
				t.Fatalf("renderPromptArtifactContractBlocks() error = %v", err)
			}
			if !strings.Contains(issueBlock, tc.expectedIssue) {
				t.Fatalf("issue contract must contain optional sprint/day pattern %q, got: %q", tc.expectedIssue, issueBlock)
			}
			if !strings.Contains(prBlock, tc.expectedPR) {
				t.Fatalf("pr contract must contain expected pattern %q, got: %q", tc.expectedPR, prBlock)
			}
		})
	}
}

func TestRenderPromptArtifactContractBlocks_ReviseInstructsAppendNotOverwrite(t *testing.T) {
	t.Parallel()

	_, prBlock, err := renderPromptArtifactContractBlocks("codex-k8s/codex-k8s", 310, "dev", "dev_revise", promptTemplateKindRevise, promptLocaleRU)
	if err != nil {
		t.Fatalf("renderPromptArtifactContractBlocks() error = %v", err)
	}
	if !strings.Contains(prBlock, "сохраняйте текущий заголовок существующего PR") {
		t.Fatalf("revise pr contract must preserve the existing title, got: %q", prBlock)
	}
	if !strings.Contains(prBlock, "gh pr view <current-pr> --json title,body") {
		t.Fatalf("revise pr contract must instruct to read current pr title and body, got: %q", prBlock)
	}
	if !strings.Contains(prBlock, "Не перезатирайте PR title/body новым revise-summary") {
		t.Fatalf("revise pr contract must preserve the existing title/body, got: %q", prBlock)
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

func TestBuildPrompt_RevisePreservesExistingPRBodyInstructions(t *testing.T) {
	t.Parallel()

	service := &Service{
		cfg: Config{
			RunID:              "run-revise",
			RepositoryFullName: "codex-k8s/codex-k8s",
			AgentKey:           "dev",
			IssueNumber:        310,
			RuntimeMode:        runtimeModeCodeOnly,
			ExistingPRNumber:   512,
			PromptConfig: PromptConfig{
				TriggerKind:          "dev_revise",
				TriggerLabel:         "run:dev:revise",
				StateInReviewLabel:   "state:in-review",
				PromptTemplateKind:   promptTemplateKindRevise,
				PromptTemplateLocale: promptLocaleRU,
				AgentBaseBranch:      "main",
			},
		},
	}

	prompt, err := service.buildPrompt("task body", runResult{
		targetBranch:     "codex/issue-310",
		triggerKind:      "dev_revise",
		templateKind:     promptTemplateKindRevise,
		existingPRNumber: 512,
	}, t.TempDir())
	if err != nil {
		t.Fatalf("buildPrompt() error = %v", err)
	}
	if !strings.Contains(prompt, "Не перезатирайте title/body существующего PR") {
		t.Fatalf("revise prompt must preserve existing pr title/body, got: %q", prompt)
	}
	if !strings.Contains(prompt, "Не перезатирайте PR title/body новым revise-summary") {
		t.Fatalf("revise prompt must include revise append contract for title/body, got: %q", prompt)
	}
	if !strings.Contains(prompt, "gh pr view <current-pr> --json title,body") {
		t.Fatalf("revise prompt must instruct to read current pr title/body, got: %q", prompt)
	}
}
