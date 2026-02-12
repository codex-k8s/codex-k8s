package mcp

import (
	"context"
	"fmt"
	"strings"

	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

func (s *Service) resolveGitHubIssueRunContext(ctx context.Context, session SessionContext, tool ToolCapability, explicitIssue int) (resolvedRunContext, int, error) {
	runCtx, err := s.resolveRunContext(ctx, session, true)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return resolvedRunContext{}, 0, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	issueNumber, err := resolveIssueNumber(explicitIssue, runCtx.Payload)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return resolvedRunContext{}, 0, err
	}
	return runCtx, issueNumber, nil
}

func githubIssueScopedRead[T any](
	ctx context.Context,
	svc *Service,
	session SessionContext,
	toolName ToolName,
	explicitIssue int,
	errorPrefix string,
	readFn func(context.Context, resolvedRunContext, int) (T, error),
) (T, error) {
	var zero T

	tool, err := svc.toolCapability(toolName)
	if err != nil {
		return zero, err
	}

	runCtx, issueNumber, err := svc.resolveGitHubIssueRunContext(ctx, session, tool, explicitIssue)
	if err != nil {
		return zero, err
	}

	value, err := readFn(ctx, runCtx, issueNumber)
	if err != nil {
		svc.auditToolFailed(ctx, runCtx.Session, tool, err)
		return zero, fmt.Errorf("%s: %w", errorPrefix, err)
	}

	svc.auditToolSucceeded(ctx, runCtx.Session, tool)
	return value, nil
}

func (s *Service) githubLabelsMutate(
	ctx context.Context,
	session SessionContext,
	toolName ToolName,
	issueNumber int,
	labels []string,
	errorPrefix string,
	mutate func(context.Context, GitHubMutateLabelsParams) ([]GitHubLabel, error),
) (GitHubLabelsMutationResult, error) {
	tool, err := s.toolCapability(toolName)
	if err != nil {
		return GitHubLabelsMutationResult{}, err
	}

	runCtx, resolvedIssueNumber, err := s.resolveGitHubIssueRunContext(ctx, session, tool, issueNumber)
	if err != nil {
		return GitHubLabelsMutationResult{}, err
	}
	if tool.Approval == ToolApprovalRequired {
		message := "approval is required by policy before labels mutation"
		s.auditToolApprovalPending(ctx, runCtx.Session, tool, message)
		return GitHubLabelsMutationResult{
			Status:  ToolExecutionStatusApprovalRequired,
			Message: message,
		}, nil
	}

	normalizedLabels := normalizeLabels(labels)
	if len(normalizedLabels) == 0 {
		err := fmt.Errorf("labels are required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubLabelsMutationResult{}, err
	}

	mutated, err := mutate(ctx, GitHubMutateLabelsParams{
		Token:       runCtx.Token,
		Owner:       runCtx.Repository.Owner,
		Repository:  runCtx.Repository.Name,
		IssueNumber: resolvedIssueNumber,
		Labels:      normalizedLabels,
	})
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return GitHubLabelsMutationResult{}, fmt.Errorf("%s: %w", errorPrefix, err)
	}

	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return GitHubLabelsMutationResult{
		Status: ToolExecutionStatusOK,
		Labels: mutated,
	}, nil
}

func resolveIssueNumber(explicit int, payload querytypes.RunPayload) (int, error) {
	if explicit > 0 {
		return explicit, nil
	}
	if payload.Issue != nil && payload.Issue.Number > 0 {
		return int(payload.Issue.Number), nil
	}
	return 0, fmt.Errorf("issue_number is required")
}

func normalizeLabels(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, value := range in {
		label := strings.TrimSpace(value)
		if label == "" {
			continue
		}
		key := strings.ToLower(label)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, label)
	}
	return out
}
