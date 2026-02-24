package mcp

import (
	"testing"

	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func TestAllowedToolsForRunContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		triggerKind      string
		agentKey         string
		wantAllowedTools []ToolName
	}{
		{
			name:        "dev gets labels only",
			triggerKind: "dev",
			agentKey:    "dev",
			wantAllowedTools: []ToolName{
				ToolGitHubLabelsAdd,
				ToolGitHubLabelsList,
				ToolGitHubLabelsRemove,
				ToolGitHubLabelsTransition,
				ToolRunStatusReport,
			},
		},
		{
			name:        "self-improve gets labels and diagnostics",
			triggerKind: "self_improve",
			agentKey:    "km",
			wantAllowedTools: []ToolName{
				ToolGitHubLabelsAdd,
				ToolGitHubLabelsList,
				ToolGitHubLabelsRemove,
				ToolGitHubLabelsTransition,
				ToolRunStatusReport,
				ToolSelfImproveRunLookup,
				ToolSelfImproveRunsList,
				ToolSelfImproveSessionGet,
			},
		},
		{
			name:        "ops sre gets labels and control tools",
			triggerKind: "ops",
			agentKey:    "sre",
			wantAllowedTools: []ToolName{
				ToolMCPDatabaseLifecycle,
				ToolGitHubLabelsAdd,
				ToolGitHubLabelsList,
				ToolGitHubLabelsRemove,
				ToolGitHubLabelsTransition,
				ToolMCPOwnerFeedbackRequest,
				ToolRunStatusReport,
				ToolMCPSecretSyncEnv,
			},
		},
		{
			name:        "ops qa gets labels only",
			triggerKind: "ops",
			agentKey:    "qa",
			wantAllowedTools: []ToolName{
				ToolGitHubLabelsAdd,
				ToolGitHubLabelsList,
				ToolGitHubLabelsRemove,
				ToolGitHubLabelsTransition,
				ToolRunStatusReport,
			},
		},
		{
			name:        "ai-repair sre gets labels and control tools",
			triggerKind: "ai_repair",
			agentKey:    "sre",
			wantAllowedTools: []ToolName{
				ToolMCPDatabaseLifecycle,
				ToolGitHubLabelsAdd,
				ToolGitHubLabelsList,
				ToolGitHubLabelsRemove,
				ToolGitHubLabelsTransition,
				ToolMCPOwnerFeedbackRequest,
				ToolRunStatusReport,
				ToolMCPSecretSyncEnv,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			service := &Service{toolCatalog: DefaultToolCatalog()}
			runCtx := resolvedRunContext{
				Payload: querytypes.RunPayload{
					Agent: &querytypes.RunPayloadAgent{
						Key: testCase.agentKey,
					},
					Trigger: &querytypes.RunPayloadTrigger{
						Kind: testCase.triggerKind,
					},
				},
			}

			got := service.allowedToolsForRunContext(runCtx)
			if len(got) != len(testCase.wantAllowedTools) {
				t.Fatalf("allowed tools count = %d, want %d", len(got), len(testCase.wantAllowedTools))
			}
			for idx, tool := range got {
				if tool.Name != testCase.wantAllowedTools[idx] {
					t.Fatalf("allowed tool at index %d = %q, want %q", idx, tool.Name, testCase.wantAllowedTools[idx])
				}
			}
		})
	}
}
