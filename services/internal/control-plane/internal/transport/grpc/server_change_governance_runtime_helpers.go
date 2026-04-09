package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/codex-k8s/kodex/libs/go/errs"
	mcpdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/mcp"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
)

func (s *Server) validateChangeGovernanceRunContext(runSession mcpdomain.SessionContext, requestedProjectID string) error {
	sessionProjectID := strings.TrimSpace(runSession.ProjectID)
	if sessionProjectID == "" {
		return errs.Forbidden{Msg: "authenticated run token is not bound to a project"}
	}
	requestedProjectID = strings.TrimSpace(requestedProjectID)
	if requestedProjectID != "" && requestedProjectID != sessionProjectID {
		return errs.Forbidden{Msg: "project_id does not match authenticated run token"}
	}
	return nil
}

func (s *Server) validateChangeGovernanceRunLineage(ctx context.Context, runID string, repositoryFullName string, issueNumber int, prNumber *int) error {
	if s == nil || s.runs == nil {
		return fmt.Errorf("agent run repository is not configured")
	}
	run, found, err := s.runs.GetByID(ctx, strings.TrimSpace(runID))
	if err != nil {
		return fmt.Errorf("load authenticated run: %w", err)
	}
	if !found {
		return errs.NotFound{Msg: fmt.Sprintf("run %q not found", strings.TrimSpace(runID))}
	}

	payload, err := querytypes.DecodeRunPayload(run.RunPayload)
	if err != nil {
		return err
	}
	if expected := strings.TrimSpace(payload.Repository.FullName); expected != "" && expected != strings.TrimSpace(repositoryFullName) {
		return errs.Forbidden{Msg: "repository_full_name does not match authenticated run context"}
	}
	if payload.Issue != nil && payload.Issue.Number > 0 && int(payload.Issue.Number) != issueNumber {
		return errs.Forbidden{Msg: "issue_number does not match authenticated run context"}
	}
	if payload.PullRequest != nil && payload.PullRequest.Number > 0 {
		switch {
		case prNumber == nil:
			return errs.Forbidden{Msg: "pr_number is required for authenticated revise run context"}
		case int(payload.PullRequest.Number) != *prNumber:
			return errs.Forbidden{Msg: "pr_number does not match authenticated run context"}
		}
	}
	return nil
}
