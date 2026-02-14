package mcp

import (
	"context"
	"strings"

	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func (s *Service) PromptContext(ctx context.Context, session SessionContext) (PromptContextResult, error) {
	tool, err := s.toolCapability(ToolPromptContextGet)
	if err != nil {
		return PromptContextResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, false)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return PromptContextResult{}, err
	}
	allowedTools := s.allowedToolsForRunContext(runCtx)

	s.auditToolCalled(ctx, runCtx.Session, tool)

	issueCtx := buildPromptIssueContext(runCtx.Payload.Issue)
	result := PromptContextResult{
		Status: ToolExecutionStatusOK,
		Context: PromptContext{
			Version: promptContextVersion,
			Run: PromptRunContext{
				RunID:         runCtx.Session.RunID,
				CorrelationID: runCtx.Session.CorrelationID,
				ProjectID:     runCtx.Session.ProjectID,
				Namespace:     runCtx.Session.Namespace,
				RuntimeMode:   runCtx.Session.RuntimeMode,
			},
			Repository: PromptRepositoryContext{
				Provider:     runCtx.Repository.Provider,
				Owner:        runCtx.Repository.Owner,
				Name:         runCtx.Repository.Name,
				FullName:     runCtx.Repository.Owner + "/" + runCtx.Repository.Name,
				ServicesYAML: runCtx.Repository.ServicesYAMLPath,
			},
			Issue: issueCtx,
			Environment: PromptEnvironmentContext{
				ServiceName: s.cfg.ServerName,
				MCPBaseURL:  s.cfg.InternalMCPBaseURL,
			},
			Services: buildPromptServices(s.cfg.PublicBaseURL, s.cfg.InternalMCPBaseURL),
			MCP: PromptMCPContext{
				ServerName: s.cfg.ServerName,
				Tools:      allowedTools,
			},
		},
	}

	s.auditPromptContextAssembled(ctx, runCtx)
	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return result, nil
}

func buildPromptIssueContext(issue *querytypes.RunPayloadIssue) *PromptIssueContext {
	if issue == nil {
		return nil
	}
	if issue.Number <= 0 {
		return nil
	}
	return &PromptIssueContext{
		Number: issue.Number,
		Title:  strings.TrimSpace(issue.Title),
		State:  strings.TrimSpace(issue.State),
		URL:    strings.TrimSpace(issue.HTMLURL),
	}
}

func buildPromptServices(publicBaseURL string, internalMCPBaseURL string) []PromptServiceContext {
	items := []PromptServiceContext{
		{Name: "control-plane-grpc", Endpoint: "codex-k8s-control-plane:9090", Kind: "grpc"},
		{Name: "control-plane-mcp", Endpoint: strings.TrimSpace(internalMCPBaseURL), Kind: "mcp-http"},
		{Name: "worker", Endpoint: "codex-k8s-worker", Kind: "job-orchestrator"},
	}
	if strings.TrimSpace(publicBaseURL) != "" {
		items = append(items, PromptServiceContext{
			Name:     "api-gateway",
			Endpoint: strings.TrimSpace(publicBaseURL),
			Kind:     "http",
		})
	}
	return items
}
