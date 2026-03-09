package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

func (h *staffHandler) PreviewNextStepAction(c *echo.Context) error {
	return h.handleNextStepAction(c, true)
}

func (h *staffHandler) ExecuteNextStepAction(c *echo.Context) error {
	return h.handleNextStepAction(c, false)
}

func (h *staffHandler) handleNextStepAction(c *echo.Context, preview bool) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		var req models.NextStepActionRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}

		repositoryFullName := strings.TrimSpace(req.RepositoryFullName)
		if repositoryFullName == "" {
			return errs.Validation{Field: "repository_full_name", Msg: "is required"}
		}
		actionKind := strings.TrimSpace(req.ActionKind)
		if actionKind == "" {
			return errs.Validation{Field: "action_kind", Msg: "is required"}
		}
		targetLabel := strings.TrimSpace(req.TargetLabel)
		if targetLabel == "" {
			return errs.Validation{Field: "target_label", Msg: "is required"}
		}

		grpcReq := &controlplanev1.NextStepActionRequest{
			Principal:          principal,
			RepositoryFullName: repositoryFullName,
			ActionKind:         actionKind,
			TargetLabel:        targetLabel,
		}
		if req.IssueNumber != nil {
			grpcReq.IssueNumber = req.IssueNumber
		}
		if req.PullRequestNumber != nil {
			grpcReq.PullRequestNumber = req.PullRequestNumber
		}

		var (
			resp *controlplanev1.NextStepActionResponse
			err  error
		)
		if preview {
			resp, err = h.cp.Service().PreviewNextStepAction(c.Request().Context(), grpcReq)
		} else {
			resp, err = h.cp.Service().ExecuteNextStepAction(c.Request().Context(), grpcReq)
		}
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, casters.NextStepActionResponse(resp))
	})
}
