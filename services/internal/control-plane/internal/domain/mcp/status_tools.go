package mcp

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

const maxRunStatusReportChars = 100

func (s *Service) RunStatusReport(ctx context.Context, session SessionContext, input RunStatusReportInput) (RunStatusReportResult, error) {
	tool, err := s.toolCapability(ToolRunStatusReport)
	if err != nil {
		return RunStatusReportResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, false)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return RunStatusReportResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	reportedStatus, err := normalizeRunStatusReportText(input.Status)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return RunStatusReportResult{}, err
	}

	s.auditRunAgentStatusReported(ctx, runCtx, reportedStatus)
	s.auditToolSucceeded(ctx, runCtx.Session, tool)

	return RunStatusReportResult{
		Status:         ToolExecutionStatusOK,
		ReportedStatus: reportedStatus,
	}, nil
}

func normalizeRunStatusReportText(value string) (string, error) {
	status := strings.TrimSpace(value)
	if status == "" {
		return "", fmt.Errorf("status is required")
	}
	if utf8.RuneCountInString(status) > maxRunStatusReportChars {
		return "", fmt.Errorf("status must be at most %d characters", maxRunStatusReportChars)
	}
	return status, nil
}
