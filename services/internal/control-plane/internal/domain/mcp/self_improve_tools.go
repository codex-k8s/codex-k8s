package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

var selfImprovePathSegmentSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func (s *Service) SelfImproveRunsList(ctx context.Context, session SessionContext, input SelfImproveRunsListInput) (SelfImproveRunsListResult, error) {
	tool, err := s.toolCapability(ToolSelfImproveRunsList)
	if err != nil {
		return SelfImproveRunsListResult{}, err
	}

	runCtx, projectID, err := s.resolveSelfImproveRunContext(ctx, session)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return SelfImproveRunsListResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	limit := normalizeSelfImproveLimit(input.Limit)
	page := normalizeSelfImprovePage(input.Page)
	offset := (page - 1) * limit
	repositoryFullName := resolveSelfImproveRepositoryFullName(strings.TrimSpace(input.RepositoryFullName), runCtx)

	items, err := s.runs.ListRecentByProject(ctx, projectID, repositoryFullName, limit+1, offset)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SelfImproveRunsListResult{}, fmt.Errorf("list recent runs: %w", err)
	}
	hasNext := len(items) > limit
	if hasNext {
		items = items[:limit]
	}

	resultItems := make([]SelfImproveRunRef, 0, len(items))
	for _, item := range items {
		resultItems = append(resultItems, selfImproveRunRefFromLookupItem(item))
	}

	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return SelfImproveRunsListResult{
		Status:  ToolExecutionStatusOK,
		Page:    page,
		Limit:   limit,
		HasNext: hasNext,
		Items:   resultItems,
	}, nil
}

func (s *Service) SelfImproveRunLookup(ctx context.Context, session SessionContext, input SelfImproveRunLookupInput) (SelfImproveRunLookupResult, error) {
	tool, err := s.toolCapability(ToolSelfImproveRunLookup)
	if err != nil {
		return SelfImproveRunLookupResult{}, err
	}

	runCtx, projectID, err := s.resolveSelfImproveRunContext(ctx, session)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return SelfImproveRunLookupResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	if input.IssueNumber <= 0 && input.PullRequestNumber <= 0 {
		err := fmt.Errorf("issue_number or pull_request_number is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SelfImproveRunLookupResult{}, err
	}

	limit := normalizeSelfImproveLimit(input.Limit)
	repositoryFullName := resolveSelfImproveRepositoryFullName(strings.TrimSpace(input.RepositoryFullName), runCtx)
	items, err := s.runs.SearchRecentByProjectIssueOrPullRequest(ctx, projectID, repositoryFullName, input.IssueNumber, input.PullRequestNumber, limit)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SelfImproveRunLookupResult{}, fmt.Errorf("search runs by references: %w", err)
	}

	resultItems := make([]SelfImproveRunRef, 0, len(items))
	for _, item := range items {
		resultItems = append(resultItems, selfImproveRunRefFromLookupItem(item))
	}

	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return SelfImproveRunLookupResult{
		Status: ToolExecutionStatusOK,
		Items:  resultItems,
	}, nil
}

func (s *Service) SelfImproveSessionGet(ctx context.Context, session SessionContext, input SelfImproveSessionGetInput) (SelfImproveSessionGetResult, error) {
	tool, err := s.toolCapability(ToolSelfImproveSessionGet)
	if err != nil {
		return SelfImproveSessionGetResult{}, err
	}

	runCtx, projectID, err := s.resolveSelfImproveRunContext(ctx, session)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return SelfImproveSessionGetResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	runID := strings.TrimSpace(input.RunID)
	if runID == "" {
		err := fmt.Errorf("run_id is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SelfImproveSessionGetResult{}, err
	}

	targetRun, found, err := s.runs.GetByID(ctx, runID)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SelfImproveSessionGetResult{}, fmt.Errorf("get run by id: %w", err)
	}
	if !found {
		err := fmt.Errorf("run not found")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SelfImproveSessionGetResult{}, err
	}
	if strings.TrimSpace(targetRun.ProjectID) != "" && !strings.EqualFold(strings.TrimSpace(targetRun.ProjectID), projectID) {
		err := fmt.Errorf("run belongs to another project")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SelfImproveSessionGetResult{}, err
	}

	sessionSnapshot, found, err := s.sessions.GetByRunID(ctx, runID)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SelfImproveSessionGetResult{}, fmt.Errorf("get run session: %w", err)
	}
	if !found {
		err := fmt.Errorf("run session not found")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SelfImproveSessionGetResult{}, err
	}

	codexSessionJSON := sessionSnapshot.CodexSessionJSON
	if len(codexSessionJSON) == 0 {
		err := fmt.Errorf("codex session json is empty")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SelfImproveSessionGetResult{}, err
	}
	if !json.Valid(codexSessionJSON) {
		err := fmt.Errorf("codex session json is invalid")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SelfImproveSessionGetResult{}, err
	}

	targetPayload, err := parseRunPayload(targetRun.RunPayload)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SelfImproveSessionGetResult{}, err
	}

	runRef := selfImproveRunRefFromRunAndPayload(targetRun, targetPayload)
	tmpDirectory := filepath.Join(selfImproveSessionTmpRoot, sanitizePathSegment(runID))
	tmpFilePath := filepath.Join(tmpDirectory, selfImproveSessionFileName)

	s.auditToolSucceeded(ctx, runCtx.Session, tool)
	return SelfImproveSessionGetResult{
		Status:           ToolExecutionStatusOK,
		Run:              runRef,
		TmpDirectory:     tmpDirectory,
		TmpFilePath:      tmpFilePath,
		CodexSessionJSON: codexSessionJSON,
	}, nil
}

func (s *Service) resolveSelfImproveRunContext(ctx context.Context, session SessionContext) (resolvedRunContext, string, error) {
	runCtx, err := s.resolveRunContext(ctx, session, false)
	if err != nil {
		return resolvedRunContext{}, "", err
	}
	if runCtx.Payload.Trigger == nil {
		return resolvedRunContext{}, "", fmt.Errorf("run trigger is required")
	}
	if webhookdomain.NormalizeTriggerKind(runCtx.Payload.Trigger.Kind) != webhookdomain.TriggerKindSelfImprove {
		return resolvedRunContext{}, "", fmt.Errorf("tool is available only for run:self-improve")
	}

	projectID := strings.TrimSpace(runCtx.Session.ProjectID)
	if projectID == "" {
		projectID = strings.TrimSpace(runCtx.Run.ProjectID)
	}
	if projectID == "" {
		projectID = strings.TrimSpace(runCtx.Repository.ProjectID)
	}
	if projectID == "" {
		projectID = strings.TrimSpace(runCtx.Payload.Project.ID)
	}
	if projectID == "" {
		return resolvedRunContext{}, "", fmt.Errorf("project_id is required")
	}
	return runCtx, projectID, nil
}

func resolveSelfImproveRepositoryFullName(explicit string, runCtx resolvedRunContext) string {
	if explicit != "" {
		return explicit
	}
	repositoryFullName := strings.TrimSpace(runCtx.Payload.Repository.FullName)
	if repositoryFullName != "" {
		return repositoryFullName
	}
	owner := strings.TrimSpace(runCtx.Repository.Owner)
	name := strings.TrimSpace(runCtx.Repository.Name)
	if owner != "" && name != "" {
		return owner + "/" + name
	}
	return ""
}

func normalizeSelfImproveLimit(value int) int {
	switch {
	case value <= 0:
		return defaultSelfImproveRunsLimit
	case value > maxSelfImproveRunsLimit:
		return maxSelfImproveRunsLimit
	default:
		return value
	}
}

func normalizeSelfImprovePage(value int) int {
	if value <= 0 {
		return 1
	}
	return value
}

func selfImproveRunRefFromLookupItem(item agentrunrepo.RunLookupItem) SelfImproveRunRef {
	result := SelfImproveRunRef{
		RunID:              item.RunID,
		CorrelationID:      item.CorrelationID,
		ProjectID:          item.ProjectID,
		RepositoryFullName: item.RepositoryFullName,
		AgentKey:           item.AgentKey,
		IssueNumber:        item.IssueNumber,
		IssueURL:           item.IssueURL,
		PullRequestNumber:  item.PullRequestNumber,
		PullRequestURL:     item.PullRequestURL,
		TriggerKind:        item.TriggerKind,
		TriggerLabel:       item.TriggerLabel,
		Status:             item.Status,
		CreatedAt:          nowRFC3339Nano(item.CreatedAt),
	}
	if item.StartedAt != nil {
		result.StartedAt = nowRFC3339Nano(*item.StartedAt)
	}
	if item.FinishedAt != nil {
		result.FinishedAt = nowRFC3339Nano(*item.FinishedAt)
	}
	return result
}

func selfImproveRunRefFromRunAndPayload(run agentrunrepo.Run, payload querytypes.RunPayload) SelfImproveRunRef {
	result := SelfImproveRunRef{
		RunID:              strings.TrimSpace(run.ID),
		CorrelationID:      strings.TrimSpace(run.CorrelationID),
		ProjectID:          strings.TrimSpace(run.ProjectID),
		RepositoryFullName: strings.TrimSpace(payload.Repository.FullName),
		Status:             strings.TrimSpace(run.Status),
	}
	if payload.Agent != nil {
		result.AgentKey = strings.TrimSpace(payload.Agent.Key)
	}
	if payload.Trigger != nil {
		result.TriggerKind = strings.TrimSpace(payload.Trigger.Kind)
		result.TriggerLabel = strings.TrimSpace(payload.Trigger.Label)
	}
	if payload.Issue != nil {
		result.IssueNumber = payload.Issue.Number
		result.IssueURL = strings.TrimSpace(payload.Issue.HTMLURL)
	}
	if payload.PullRequest != nil {
		result.PullRequestNumber = payload.PullRequest.Number
		result.PullRequestURL = strings.TrimSpace(payload.PullRequest.HTMLURL)
	}
	return result
}

func sanitizePathSegment(value string) string {
	sanitized := selfImprovePathSegmentSanitizer.ReplaceAllString(strings.TrimSpace(value), "-")
	sanitized = strings.Trim(sanitized, "-.")
	if sanitized == "" {
		return "run"
	}
	return sanitized
}
