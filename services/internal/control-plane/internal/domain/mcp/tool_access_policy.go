package mcp

import (
	"context"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

var (
	baseLabelTools = []ToolName{
		ToolGitHubLabelsList,
		ToolGitHubLabelsAdd,
		ToolGitHubLabelsRemove,
		ToolGitHubLabelsTransition,
		ToolRunStatusReport,
	}
	selfImproveDiagnosticTools = []ToolName{
		ToolSelfImproveRunsList,
		ToolSelfImproveRunLookup,
		ToolSelfImproveSessionGet,
	}
	controlPlaneControlTools = []ToolName{
		ToolMCPSecretSyncEnv,
		ToolMCPDatabaseLifecycle,
		ToolMCPOwnerFeedbackRequest,
	}
)

// AllowedTools resolves MCP tool catalog visible for one authenticated run session.
func (s *Service) AllowedTools(ctx context.Context, session SessionContext) ([]ToolCapability, error) {
	runCtx, err := s.resolveRunContext(ctx, session, false)
	if err != nil {
		return nil, err
	}
	return s.allowedToolsForRunContext(runCtx), nil
}

// IsToolAllowed reports whether tool is allowed for current authenticated run session.
func (s *Service) IsToolAllowed(ctx context.Context, session SessionContext, toolName ToolName) (bool, error) {
	tools, err := s.AllowedTools(ctx, session)
	if err != nil {
		return false, err
	}
	for _, tool := range tools {
		if tool.Name == toolName {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) allowedToolsForRunContext(runCtx resolvedRunContext) []ToolCapability {
	allowedNames := make(map[ToolName]struct{}, len(baseLabelTools)+len(selfImproveDiagnosticTools)+len(controlPlaneControlTools))
	addAllowedToolNames(allowedNames, baseLabelTools...)

	triggerKind := webhookdomain.TriggerKindDev
	if runCtx.Payload.Trigger != nil {
		triggerKind = webhookdomain.NormalizeTriggerKind(strings.TrimSpace(runCtx.Payload.Trigger.Kind))
	}
	agentKey := ""
	if runCtx.Payload.Agent != nil {
		agentKey = strings.ToLower(strings.TrimSpace(runCtx.Payload.Agent.Key))
	}

	switch triggerKind {
	case webhookdomain.TriggerKindSelfImprove:
		addAllowedToolNames(allowedNames, selfImproveDiagnosticTools...)
	case webhookdomain.TriggerKindOps, webhookdomain.TriggerKindAIRepair:
		if agentKey == agentKeySRE || agentKey == agentKeyDev {
			addAllowedToolNames(allowedNames, controlPlaneControlTools...)
		}
	}

	out := make([]ToolCapability, 0, len(allowedNames))
	for _, tool := range s.toolCatalog {
		if _, ok := allowedNames[tool.Name]; ok {
			out = append(out, tool)
		}
	}
	return out
}

func addAllowedToolNames(allowed map[ToolName]struct{}, names ...ToolName) {
	for _, name := range names {
		allowed[name] = struct{}{}
	}
}
