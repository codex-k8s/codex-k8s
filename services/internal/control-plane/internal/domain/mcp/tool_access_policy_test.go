package mcp

import (
	"testing"

	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
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
			name:        "dev gets labels and user interactions",
			triggerKind: "dev",
			agentKey:    "dev",
			wantAllowedTools: []ToolName{
				ToolGitHubLabelsAdd,
				ToolGitHubLabelsList,
				ToolGitHubLabelsRemove,
				ToolGitHubLabelsTransition,
				ToolRunStatusReport,
				ToolMCPUserDecisionRequest,
				ToolMCPUserNotify,
			},
		},
		{
			name:        "discussion gets labels and user interactions",
			triggerKind: "discussion",
			agentKey:    "dev",
			wantAllowedTools: []ToolName{
				ToolGitHubLabelsAdd,
				ToolGitHubLabelsList,
				ToolGitHubLabelsRemove,
				ToolGitHubLabelsTransition,
				ToolRunStatusReport,
				ToolMCPUserDecisionRequest,
				ToolMCPUserNotify,
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
			name:        "self-improve revise keeps diagnostics without user interactions",
			triggerKind: "self_improve_revise",
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
			name:        "ops sre gets labels, user interactions and control tools",
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
				ToolMCPUserDecisionRequest,
				ToolMCPUserNotify,
			},
		},
		{
			name:        "ops qa gets labels and user interactions",
			triggerKind: "ops",
			agentKey:    "qa",
			wantAllowedTools: []ToolName{
				ToolGitHubLabelsAdd,
				ToolGitHubLabelsList,
				ToolGitHubLabelsRemove,
				ToolGitHubLabelsTransition,
				ToolRunStatusReport,
				ToolMCPUserDecisionRequest,
				ToolMCPUserNotify,
			},
		},
		{
			name:        "ai-repair sre gets labels, user interactions and control tools",
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
				ToolMCPUserDecisionRequest,
				ToolMCPUserNotify,
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
