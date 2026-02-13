package staff

import (
	"context"
	"fmt"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	runstatusdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runstatus"
)

const (
	projectRoleReadWrite = "read_write"
	projectRoleAdmin     = "admin"
)

type runNamespaceService interface {
	DeleteRunNamespace(ctx context.Context, params runstatusdomain.DeleteNamespaceParams) (runstatusdomain.DeleteNamespaceResult, error)
	GetRunRuntimeState(ctx context.Context, runID string) (runstatusdomain.RuntimeState, error)
}

// RunNamespaceDeleteResult describes one manual namespace deletion outcome.
type RunNamespaceDeleteResult struct {
	RunID          string
	Namespace      string
	Deleted        bool
	AlreadyDeleted bool
	CommentURL     string
}

// DeleteRunNamespace removes one run namespace by run id (staff-only, write access required).
func (s *Service) DeleteRunNamespace(ctx context.Context, principal Principal, runID string) (RunNamespaceDeleteResult, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return RunNamespaceDeleteResult{}, errs.Validation{Field: "run_id", Msg: "is required"}
	}

	_, projectID, err := s.resolveRunAccess(ctx, principal, runID)
	if err != nil {
		return RunNamespaceDeleteResult{}, err
	}

	if err := s.requireRunWriteAccess(ctx, principal, projectID); err != nil {
		return RunNamespaceDeleteResult{}, err
	}

	if s.runStatus == nil {
		return RunNamespaceDeleteResult{}, fmt.Errorf("run status service is not configured")
	}

	result, err := s.runStatus.DeleteRunNamespace(ctx, runstatusdomain.DeleteNamespaceParams{
		RunID:           runID,
		RequestedByType: runstatusdomain.RequestedByTypeStaffUser,
		RequestedByID:   strings.TrimSpace(principal.UserID),
	})
	if err != nil {
		return RunNamespaceDeleteResult{}, err
	}

	return RunNamespaceDeleteResult{
		RunID:          result.RunID,
		Namespace:      result.Namespace,
		Deleted:        result.Deleted,
		AlreadyDeleted: result.AlreadyDeleted,
		CommentURL:     result.CommentURL,
	}, nil
}

func (s *Service) requireRunWriteAccess(ctx context.Context, principal Principal, projectID string) error {
	if principal.IsPlatformAdmin {
		return nil
	}
	if strings.TrimSpace(projectID) == "" {
		return errs.Forbidden{Msg: "run is not assigned to a project"}
	}

	role, hasRole, err := s.members.GetRole(ctx, projectID, principal.UserID)
	if err != nil {
		return err
	}
	if !hasRole {
		return errs.Forbidden{Msg: "project access required"}
	}
	if role != projectRoleAdmin && role != projectRoleReadWrite {
		return errs.Forbidden{Msg: "project write access required"}
	}
	return nil
}
