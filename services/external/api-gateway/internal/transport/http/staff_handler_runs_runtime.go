package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

func (h *staffHandler) ListRunWaits(c *echo.Context) error {
	return h.listRunsByFilter(c, true, h.listRunWaitsAsGetter)
}

func (h *staffHandler) ListPendingApprovals(c *echo.Context) error {
	return listByLimitResp(c, 200, h.listPendingApprovalsCall, casters.ApprovalRequests)
}

func (h *staffHandler) ResolveApprovalDecision(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePath("approval_request_id"), func(principal *controlplanev1.Principal, rawID string) error {
		approvalRequestID, err := strconv.ParseInt(strings.TrimSpace(rawID), 10, 64)
		if err != nil || approvalRequestID <= 0 {
			return errs.Validation{Field: "approval_request_id", Msg: "must be a positive int64"}
		}

		var req models.ResolveApprovalDecisionRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}

		item, err := h.resolveApprovalDecisionCall(c.Request().Context(), principal, approvalDecisionArg{
			approvalRequestID: approvalRequestID,
			body:              req,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.ResolveApprovalDecision(item))
	})
}

func (h *staffHandler) GetRun(c *echo.Context) error {
	return getByPathResp(c, "run_id", h.getRunCall, casters.Run)
}

func (h *staffHandler) CancelRun(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePath("run_id"), func(principal *controlplanev1.Principal, runID string) error {
		var req models.RunActionRequest
		if err := bindBodyOptional(c, &req); err != nil {
			return err
		}
		resp, err := h.cancelRunCall(c.Request().Context(), principal, runActionArg{
			runID: strings.TrimSpace(runID),
			body:  req,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.RunAction(resp))
	})
}

func (h *staffHandler) GetRunLogs(c *echo.Context) error {
	return withPrincipalAndResolvedJSON(c, resolveRunLogsArg(200), h.getRunLogsCall, casters.RunLogs)
}

func (h *staffHandler) DeleteRunNamespace(c *echo.Context) error {
	return withPrincipalAndResolvedJSON(c, resolvePath("run_id"), h.deleteRunNamespaceCall, casters.RunNamespaceDelete)
}

func (h *staffHandler) ListRunEvents(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolveRunEventsArg(500), func(principal *controlplanev1.Principal, arg runEventsArg) error {
		resp, err := h.listRunEventsCall(c.Request().Context(), principal, arg)
		if err != nil {
			return err
		}
		items := resp.GetItems()
		if !arg.includePayload {
			return c.JSON(http.StatusOK, models.ItemsResponse[models.FlowEvent]{Items: casters.FlowEventsSummary(items)})
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[models.FlowEvent]{Items: casters.FlowEvents(items)})
	})
}

func (h *staffHandler) ListRunLearningFeedback(c *echo.Context) error {
	return listByPathLimitResp(c, "run_id", 200, h.listRunLearningFeedbackCall, casters.LearningFeedbackList)
}

func (h *staffHandler) GetRuntimeDeployTask(c *echo.Context) error {
	return getByPathResp(c, "run_id", h.getRuntimeDeployTaskCall, casters.RuntimeDeployTask)
}

func (h *staffHandler) CancelRuntimeDeployTask(c *echo.Context) error {
	return h.runtimeDeployTaskAction(c, h.cancelRuntimeDeployTaskCall)
}

func (h *staffHandler) StopRuntimeDeployTask(c *echo.Context) error {
	return h.runtimeDeployTaskAction(c, h.stopRuntimeDeployTaskCall)
}

func (h *staffHandler) runtimeDeployTaskAction(
	c *echo.Context,
	call func(ctx context.Context, principal *controlplanev1.Principal, arg runtimeDeployActionArg) (*controlplanev1.RuntimeDeployTaskActionResponse, error),
) error {
	return withPrincipalAndResolved(c, resolvePath("run_id"), func(principal *controlplanev1.Principal, runID string) error {
		var req models.RuntimeDeployTaskActionRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		resp, err := call(c.Request().Context(), principal, runtimeDeployActionArg{
			runID: strings.TrimSpace(runID),
			body:  req,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.RuntimeDeployTaskAction(resp))
	})
}
