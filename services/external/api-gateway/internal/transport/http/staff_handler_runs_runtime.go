package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

func (h *staffHandler) ListRuns(c *echo.Context) error {
	return listByLimitResp(c, 200, h.listRunsCall, casters.Runs)
}

func (h *staffHandler) ListRunJobs(c *echo.Context) error {
	return h.listRunsByFilter(c, false, h.listRunJobsAsGetter)
}

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

func (h *staffHandler) ListRuntimeDeployTasks(c *echo.Context) error {
	return listByResolvedResp(c, resolveRuntimeDeployListFilters(200), h.listRuntimeDeployTasksCall, casters.RuntimeDeployTaskListItems)
}

func (h *staffHandler) GetRuntimeDeployTask(c *echo.Context) error {
	return getByPathResp(c, "run_id", h.getRuntimeDeployTaskCall, casters.RuntimeDeployTask)
}

func (h *staffHandler) ListRegistryImages(c *echo.Context) error {
	return listByResolvedResp(c, resolveRegistryImagesListFilters(100, 50), h.listRegistryImagesCall, casters.RegistryImageRepositories)
}

func (h *staffHandler) DeleteRegistryImageTag(c *echo.Context) error {
	return createByBodyResp(c, http.StatusOK, h.deleteRegistryImageTagCall, casters.RegistryImageDeleteResult)
}

func (h *staffHandler) CleanupRegistryImages(c *echo.Context) error {
	return createByBodyResp(c, http.StatusOK, h.cleanupRegistryImagesCall, casters.RegistryImageCleanupResult)
}
