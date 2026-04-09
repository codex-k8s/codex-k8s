package staff

import (
	"context"
	"fmt"
	"strings"

	"github.com/codex-k8s/kodex/libs/go/errs"
	runstatusdomain "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/runstatus"
)

// CancelRun stops one run and active runtime artifacts by run id (staff-only, write access required).
func (s *Service) CancelRun(ctx context.Context, principal Principal, runID string, reason string) (RunCancelResult, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return RunCancelResult{}, errs.Validation{Field: "run_id", Msg: "is required"}
	}

	_, projectID, err := s.resolveRunAccess(ctx, principal, runID)
	if err != nil {
		return RunCancelResult{}, err
	}

	if err := s.requireRunWriteAccess(ctx, principal, projectID); err != nil {
		return RunCancelResult{}, err
	}

	if s.runStatus == nil {
		return RunCancelResult{}, fmt.Errorf("run status service is not configured")
	}

	result, err := s.runStatus.CancelRun(ctx, runstatusdomain.CancelRunParams{
		RunID:             runID,
		Reason:            strings.TrimSpace(reason),
		RequestedByType:   runstatusdomain.RequestedByTypeStaffUser,
		RequestedByID:     strings.TrimSpace(principal.UserID),
		RequestedByEmail:  strings.TrimSpace(principal.Email),
		RequestedByGitHub: strings.TrimSpace(principal.GitHubLogin),
	})
	if err != nil {
		return RunCancelResult{}, err
	}

	return RunCancelResult{
		RunID:                        result.RunID,
		PreviousStatus:               result.PreviousStatus,
		CurrentStatus:                result.CurrentStatus,
		AlreadyTerminal:              result.AlreadyTerminal,
		RuntimeDeployCancelRequested: result.RuntimeDeployCancelRequested,
		JobStopped:                   result.JobStopped,
		CanceledGitHubWaits:          result.CanceledGitHubWaits,
		CommentURL:                   result.CommentURL,
	}, nil
}
