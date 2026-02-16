package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/controlplane"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

// staffHandler implements staff/private JSON endpoints protected by JWT.
type staffHandler struct {
	cp *controlplane.Client
}

func newStaffHandler(cp *controlplane.Client) *staffHandler {
	return &staffHandler{cp: cp}
}

func parseLimit(c *echo.Context, def int) (int, error) {
	limitStr := c.QueryParam("limit")
	if limitStr == "" {
		return def, nil
	}
	n, err := strconv.Atoi(limitStr)
	if err != nil || n <= 0 {
		return 0, errs.Validation{Field: "limit", Msg: "must be a positive integer"}
	}
	if n > 1000 {
		n = 1000
	}
	return n, nil
}

func requirePrincipal(c *echo.Context) (*controlplanev1.Principal, error) {
	p, ok := getPrincipal(c)
	if !ok || p == nil || strings.TrimSpace(p.UserId) == "" {
		return nil, errs.Unauthorized{Msg: "not authenticated"}
	}
	return p, nil
}

func requirePathParam(c *echo.Context, name string) (string, error) {
	v := strings.TrimSpace(c.Param(name))
	if v == "" {
		return "", errs.Validation{Field: name, Msg: "is required"}
	}
	return v, nil
}

func bindBody(c *echo.Context, target interface{}) error {
	if err := c.Bind(target); err != nil {
		return errs.Validation{Field: "body", Msg: "invalid JSON"}
	}
	return nil
}

func resolvePath(param string) func(c *echo.Context) (string, error) {
	return func(c *echo.Context) (string, error) {
		return requirePathParam(c, param)
	}
}

func resolveLimit(defLimit int) func(c *echo.Context) (int, error) {
	return func(c *echo.Context) (int, error) {
		return parseLimit(c, defLimit)
	}
}

func resolvePathLimit(param string, defLimit int) func(c *echo.Context) (pathLimit, error) {
	return func(c *echo.Context) (pathLimit, error) {
		id, err := requirePathParam(c, param)
		if err != nil {
			return pathLimit{}, err
		}
		limit, err := parseLimit(c, defLimit)
		if err != nil {
			return pathLimit{}, err
		}
		return pathLimit{id: id, limit: limit}, nil
	}
}

func resolveRunListFilters(defLimit int, includeWaitState bool) func(c *echo.Context) (runListFilterArg, error) {
	return func(c *echo.Context) (runListFilterArg, error) {
		limit, err := parseLimit(c, defLimit)
		if err != nil {
			return runListFilterArg{}, err
		}
		result := runListFilterArg{
			limit:       int32(limit),
			triggerKind: strings.TrimSpace(c.QueryParam("trigger_kind")),
			status:      strings.TrimSpace(c.QueryParam("status")),
			agentKey:    strings.TrimSpace(c.QueryParam("agent_key")),
		}
		if includeWaitState {
			result.waitState = strings.TrimSpace(c.QueryParam("wait_state"))
		}
		return result, nil
	}
}

func resolveRuntimeDeployListFilters(defLimit int) func(c *echo.Context) (runtimeDeployListArg, error) {
	return func(c *echo.Context) (runtimeDeployListArg, error) {
		limit, err := parseLimit(c, defLimit)
		if err != nil {
			return runtimeDeployListArg{}, err
		}
		return runtimeDeployListArg{
			limit:     int32(limit),
			status:    strings.TrimSpace(c.QueryParam("status")),
			targetEnv: strings.TrimSpace(c.QueryParam("target_env")),
		}, nil
	}
}

func parseOptionalPositiveInt(raw string, field string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(trimmed)
	if err != nil || value <= 0 {
		return 0, errs.Validation{Field: field, Msg: "must be a positive integer"}
	}
	if value > 1000 {
		value = 1000
	}
	return value, nil
}

func resolveRegistryImagesListFilters(defLimitRepositories int, defLimitTags int) func(c *echo.Context) (registryImagesListArg, error) {
	return func(c *echo.Context) (registryImagesListArg, error) {
		limitRepositories, err := parseOptionalPositiveInt(c.QueryParam("limit_repositories"), "limit_repositories")
		if err != nil {
			return registryImagesListArg{}, err
		}
		if limitRepositories == 0 {
			limitRepositories = defLimitRepositories
		}

		limitTags, err := parseOptionalPositiveInt(c.QueryParam("limit_tags"), "limit_tags")
		if err != nil {
			return registryImagesListArg{}, err
		}
		if limitTags == 0 {
			limitTags = defLimitTags
		}

		return registryImagesListArg{
			repository:        strings.TrimSpace(c.QueryParam("repository")),
			limitRepositories: int32(limitRepositories),
			limitTags:         int32(limitTags),
		}, nil
	}
}

func resolveRunLogsArg(defTailLines int) func(c *echo.Context) (runLogsArg, error) {
	return func(c *echo.Context) (runLogsArg, error) {
		runID, err := requirePathParam(c, "run_id")
		if err != nil {
			return runLogsArg{}, err
		}

		tailLines := defTailLines
		if rawTailLines := strings.TrimSpace(c.QueryParam("tail_lines")); rawTailLines != "" {
			value, convErr := strconv.Atoi(rawTailLines)
			if convErr != nil || value <= 0 {
				return runLogsArg{}, errs.Validation{Field: "tail_lines", Msg: "must be a positive integer"}
			}
			if value > 2000 {
				value = 2000
			}
			tailLines = value
		}

		return runLogsArg{runID: runID, tailLines: int32(tailLines)}, nil
	}
}

func withPrincipal(c *echo.Context, fn func(principal *controlplanev1.Principal) error) error {
	principal, err := requirePrincipal(c)
	if err != nil {
		return err
	}
	return fn(principal)
}

func withPrincipalAndResolved[T any](
	c *echo.Context,
	resolve func(c *echo.Context) (T, error),
	fn func(principal *controlplanev1.Principal, value T) error,
) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		value, err := resolve(c)
		if err != nil {
			return err
		}
		return fn(principal, value)
	})
}

func withPrincipalAndTwoPaths(
	c *echo.Context,
	param1 string,
	param2 string,
	fn func(principal *controlplanev1.Principal, id1 string, id2 string) error,
) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		id1, err := requirePathParam(c, param1)
		if err != nil {
			return err
		}
		id2, err := requirePathParam(c, param2)
		if err != nil {
			return err
		}
		return fn(principal, id1, id2)
	})
}

type itemsGetter[Proto any] interface {
	GetItems() []Proto
}

type runItemsGetter interface {
	GetItems() []*controlplanev1.Run
}

type runListCallFn func(ctx context.Context, principal *controlplanev1.Principal, arg runListFilterArg) (runItemsGetter, error)

func listByLimitResp[Proto any, Resp itemsGetter[Proto], Out any](
	c *echo.Context,
	defLimit int,
	call func(ctx context.Context, principal *controlplanev1.Principal, limit int32) (Resp, error),
	cast func(items []Proto) []Out,
) error {
	return withPrincipalAndResolved(c, resolveLimit(defLimit), func(principal *controlplanev1.Principal, limit int) error {
		resp, err := call(c.Request().Context(), principal, int32(limit))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[Out]{Items: cast(resp.GetItems())})
	})
}

func listByPathLimitResp[Proto any, Resp itemsGetter[Proto], Out any](
	c *echo.Context,
	param string,
	defLimit int,
	call func(ctx context.Context, principal *controlplanev1.Principal, id string, limit int32) (Resp, error),
	cast func(items []Proto) []Out,
) error {
	return withPrincipalAndResolved(c, resolvePathLimit(param, defLimit), func(principal *controlplanev1.Principal, value pathLimit) error {
		resp, err := call(c.Request().Context(), principal, value.id, int32(value.limit))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[Out]{Items: cast(resp.GetItems())})
	})
}

func getByPathResp[Proto any, Out any](
	c *echo.Context,
	param string,
	call func(ctx context.Context, principal *controlplanev1.Principal, id string) (Proto, error),
	cast func(item Proto) Out,
) error {
	return withPrincipalAndResolved(c, resolvePath(param), func(principal *controlplanev1.Principal, id string) error {
		item, err := call(c.Request().Context(), principal, id)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, cast(item))
	})
}

func withPrincipalAndResolvedJSON[Req any, Proto any, Out any](
	c *echo.Context,
	resolve func(c *echo.Context) (Req, error),
	call func(ctx context.Context, principal *controlplanev1.Principal, req Req) (Proto, error),
	cast func(item Proto) Out,
) error {
	return withPrincipalAndResolved(c, resolve, func(principal *controlplanev1.Principal, req Req) error {
		item, err := call(c.Request().Context(), principal, req)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, cast(item))
	})
}

func (h *staffHandler) listRunsByFilter(
	c *echo.Context,
	includeWaitState bool,
	call runListCallFn,
) error {
	return withPrincipalAndResolved(c, resolveRunListFilters(200, includeWaitState), func(principal *controlplanev1.Principal, arg runListFilterArg) error {
		resp, err := call(c.Request().Context(), principal, arg)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[models.Run]{Items: casters.Runs(resp.GetItems())})
	})
}

func createByBodyResp[Req any, Proto any, Out any](
	c *echo.Context,
	statusCode int,
	call func(ctx context.Context, principal *controlplanev1.Principal, req Req) (Proto, error),
	cast func(item Proto) Out,
) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		var req Req
		if err := bindBody(c, &req); err != nil {
			return err
		}
		item, err := call(c.Request().Context(), principal, req)
		if err != nil {
			return err
		}
		return c.JSON(statusCode, cast(item))
	})
}

func (h *staffHandler) deleteWith1Param(c *echo.Context, paramName string, fn func(ctx context.Context, principal *controlplanev1.Principal, id string) error) error {
	return withPrincipalAndResolved(c, resolvePath(paramName), func(principal *controlplanev1.Principal, id string) error {
		if err := fn(c.Request().Context(), principal, id); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}

func (h *staffHandler) deleteWith2Params(
	c *echo.Context,
	param1 string,
	param2 string,
	fn func(ctx context.Context, principal *controlplanev1.Principal, id1 string, id2 string) error,
) error {
	return withPrincipalAndTwoPaths(c, param1, param2, func(principal *controlplanev1.Principal, id1 string, id2 string) error {
		if err := fn(c.Request().Context(), principal, id1, id2); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}

func (h *staffHandler) ListProjects(c *echo.Context) error {
	return listByLimitResp(c, 200, h.listProjectsCall, casters.Projects)
}

func (h *staffHandler) GetProject(c *echo.Context) error {
	return getByPathResp(c, "project_id", h.getProjectCall, casters.Project)
}

func (h *staffHandler) UpsertProject(c *echo.Context) error {
	return createByBodyResp(c, http.StatusCreated, h.upsertProjectCall, casters.Project)
}

func (h *staffHandler) DeleteProject(c *echo.Context) error {
	return h.deleteWith1Param(c, "project_id", h.deleteProject)
}

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
	return listByPathLimitResp(c, "run_id", 500, h.listRunEventsCall, casters.FlowEvents)
}

func (h *staffHandler) ListRunLearningFeedback(c *echo.Context) error {
	return listByPathLimitResp(c, "run_id", 200, h.listRunLearningFeedbackCall, casters.LearningFeedbackList)
}

func (h *staffHandler) ListRuntimeDeployTasks(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolveRuntimeDeployListFilters(200), func(principal *controlplanev1.Principal, arg runtimeDeployListArg) error {
		resp, err := h.listRuntimeDeployTasksCall(c.Request().Context(), principal, arg)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[models.RuntimeDeployTask]{Items: casters.RuntimeDeployTasks(resp.GetItems())})
	})
}

func (h *staffHandler) GetRuntimeDeployTask(c *echo.Context) error {
	return getByPathResp(c, "run_id", h.getRuntimeDeployTaskCall, casters.RuntimeDeployTask)
}

func (h *staffHandler) ListRegistryImages(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolveRegistryImagesListFilters(100, 50), func(principal *controlplanev1.Principal, arg registryImagesListArg) error {
		resp, err := h.listRegistryImagesCall(c.Request().Context(), principal, arg)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, models.ItemsResponse[models.RegistryImageRepository]{Items: casters.RegistryImageRepositories(resp.GetItems())})
	})
}

func (h *staffHandler) DeleteRegistryImageTag(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		var req models.DeleteRegistryImageTagRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		item, err := h.deleteRegistryImageTagCall(c.Request().Context(), principal, req)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.RegistryImageDeleteResult(item))
	})
}

func (h *staffHandler) CleanupRegistryImages(c *echo.Context) error {
	return withPrincipal(c, func(principal *controlplanev1.Principal) error {
		var req models.CleanupRegistryImagesRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		item, err := h.cleanupRegistryImagesCall(c.Request().Context(), principal, req)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, casters.RegistryImageCleanupResult(item))
	})
}

func (h *staffHandler) ListUsers(c *echo.Context) error {
	return listByLimitResp(c, 200, h.listUsersCall, casters.Users)
}

func (h *staffHandler) DeleteUser(c *echo.Context) error {
	return h.deleteWith1Param(c, "user_id", h.deleteUser)
}

func (h *staffHandler) CreateUser(c *echo.Context) error {
	return createByBodyResp(c, http.StatusCreated, h.createUserCall, casters.User)
}

func (h *staffHandler) ListProjectMembers(c *echo.Context) error {
	return listByPathLimitResp(c, "project_id", 200, h.listProjectMembersCall, casters.ProjectMembers)
}

func (h *staffHandler) UpsertProjectMember(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePath("project_id"), func(principal *controlplanev1.Principal, projectID string) error {
		var req models.UpsertProjectMemberRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}

		email := strings.TrimSpace(req.Email)
		userID := strings.TrimSpace(req.UserID)
		if email != "" && userID != "" {
			return errs.Validation{Field: "user_id", Msg: "either user_id or email must be set"}
		}
		if email == "" && userID == "" {
			return errs.Validation{Field: "user_id", Msg: "is required"}
		}

		if _, err := h.cp.Service().UpsertProjectMember(c.Request().Context(), &controlplanev1.UpsertProjectMemberRequest{
			Principal: principal,
			ProjectId: projectID,
			UserId:    optionalStringPtr(userID),
			Email:     optionalStringPtr(email),
			Role:      req.Role,
		}); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}

func (h *staffHandler) DeleteProjectMember(c *echo.Context) error {
	return h.deleteWith2Params(c, "project_id", "user_id", h.deleteProjectMember)
}

func (h *staffHandler) SetProjectMemberLearningModeOverride(c *echo.Context) error {
	return withPrincipalAndTwoPaths(c, "project_id", "user_id", func(principal *controlplanev1.Principal, projectID string, userID string) error {
		var req models.SetProjectMemberLearningModeRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		var enabled *wrapperspb.BoolValue
		if req.Enabled != nil {
			enabled = wrapperspb.Bool(*req.Enabled)
		}
		if _, err := h.cp.Service().SetProjectMemberLearningModeOverride(c.Request().Context(), &controlplanev1.SetProjectMemberLearningModeOverrideRequest{
			Principal: principal,
			ProjectId: projectID,
			UserId:    userID,
			Enabled:   enabled,
		}); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
}

func (h *staffHandler) ListProjectRepositories(c *echo.Context) error {
	return listByPathLimitResp(c, "project_id", 200, h.listProjectRepositoriesCall, casters.RepositoryBindings)
}

func (h *staffHandler) UpsertProjectRepository(c *echo.Context) error {
	return withPrincipalAndResolved(c, resolvePath("project_id"), func(principal *controlplanev1.Principal, projectID string) error {
		var req models.UpsertProjectRepositoryRequest
		if err := bindBody(c, &req); err != nil {
			return err
		}
		item, err := h.cp.Service().UpsertProjectRepository(c.Request().Context(), &controlplanev1.UpsertProjectRepositoryRequest{
			Principal:        principal,
			ProjectId:        projectID,
			Provider:         req.Provider,
			Owner:            req.Owner,
			Name:             req.Name,
			Token:            req.Token,
			ServicesYamlPath: req.ServicesYAMLPath,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, casters.RepositoryBinding(item))
	})
}

func (h *staffHandler) DeleteProjectRepository(c *echo.Context) error {
	return h.deleteWith2Params(c, "project_id", "repository_id", h.deleteProjectRepository)
}

func callUnaryWithArg[Arg any, Req any, Resp any](
	ctx context.Context,
	principal *controlplanev1.Principal,
	arg Arg,
	build func(principal *controlplanev1.Principal, arg Arg) *Req,
	call func(ctx context.Context, req *Req, opts ...grpc.CallOption) (Resp, error),
) (Resp, error) {
	return call(ctx, build(principal, arg))
}

func buildListProjectsRequest(principal *controlplanev1.Principal, limit int32) *controlplanev1.ListProjectsRequest {
	return &controlplanev1.ListProjectsRequest{Principal: principal, Limit: limit}
}

func buildGetProjectRequest(principal *controlplanev1.Principal, id string) *controlplanev1.GetProjectRequest {
	return &controlplanev1.GetProjectRequest{Principal: principal, ProjectId: id}
}

func buildUpsertProjectRequest(principal *controlplanev1.Principal, req models.UpsertProjectRequest) *controlplanev1.UpsertProjectRequest {
	return &controlplanev1.UpsertProjectRequest{Principal: principal, Slug: req.Slug, Name: req.Name}
}

func buildListRunsRequest(principal *controlplanev1.Principal, limit int32) *controlplanev1.ListRunsRequest {
	return &controlplanev1.ListRunsRequest{Principal: principal, Limit: limit}
}

func buildListRunJobsRequest(principal *controlplanev1.Principal, arg runListFilterArg) *controlplanev1.ListRunJobsRequest {
	return &controlplanev1.ListRunJobsRequest{
		Principal:   principal,
		Limit:       arg.limit,
		TriggerKind: optionalStringPtr(arg.triggerKind),
		Status:      optionalStringPtr(arg.status),
		AgentKey:    optionalStringPtr(arg.agentKey),
	}
}

func buildListRunWaitsRequest(principal *controlplanev1.Principal, arg runListFilterArg) *controlplanev1.ListRunWaitsRequest {
	return &controlplanev1.ListRunWaitsRequest{
		Principal:   principal,
		Limit:       arg.limit,
		TriggerKind: optionalStringPtr(arg.triggerKind),
		Status:      optionalStringPtr(arg.status),
		AgentKey:    optionalStringPtr(arg.agentKey),
		WaitState:   optionalStringPtr(arg.waitState),
	}
}

func buildListPendingApprovalsRequest(principal *controlplanev1.Principal, limit int32) *controlplanev1.ListPendingApprovalsRequest {
	return &controlplanev1.ListPendingApprovalsRequest{Principal: principal, Limit: limit}
}

func buildGetRunRequest(principal *controlplanev1.Principal, id string) *controlplanev1.GetRunRequest {
	return &controlplanev1.GetRunRequest{Principal: principal, RunId: id}
}

func buildGetRunLogsRequest(principal *controlplanev1.Principal, arg runLogsArg) *controlplanev1.GetRunLogsRequest {
	return &controlplanev1.GetRunLogsRequest{
		Principal: principal,
		RunId:     arg.runID,
		TailLines: arg.tailLines,
	}
}

type approvalDecisionArg struct {
	approvalRequestID int64
	body              models.ResolveApprovalDecisionRequest
}

func buildResolveApprovalDecisionRequest(principal *controlplanev1.Principal, arg approvalDecisionArg) *controlplanev1.ResolveApprovalDecisionRequest {
	return &controlplanev1.ResolveApprovalDecisionRequest{
		Principal:         principal,
		ApprovalRequestId: arg.approvalRequestID,
		Decision:          arg.body.Decision,
		Reason:            optionalStringPtr(arg.body.Reason),
	}
}

func buildDeleteRunNamespaceRequest(principal *controlplanev1.Principal, id string) *controlplanev1.DeleteRunNamespaceRequest {
	return &controlplanev1.DeleteRunNamespaceRequest{Principal: principal, RunId: id}
}

func buildListRunEventsRequest(principal *controlplanev1.Principal, arg idLimitArg) *controlplanev1.ListRunEventsRequest {
	return &controlplanev1.ListRunEventsRequest{Principal: principal, RunId: arg.id, Limit: arg.limit}
}

func buildListRunLearningFeedbackRequest(principal *controlplanev1.Principal, arg idLimitArg) *controlplanev1.ListRunLearningFeedbackRequest {
	return &controlplanev1.ListRunLearningFeedbackRequest{Principal: principal, RunId: arg.id, Limit: arg.limit}
}

func buildListRuntimeDeployTasksRequest(principal *controlplanev1.Principal, arg runtimeDeployListArg) *controlplanev1.ListRuntimeDeployTasksRequest {
	return &controlplanev1.ListRuntimeDeployTasksRequest{
		Principal: principal,
		Limit:     arg.limit,
		Status:    optionalStringPtr(arg.status),
		TargetEnv: optionalStringPtr(arg.targetEnv),
	}
}

func buildGetRuntimeDeployTaskRequest(principal *controlplanev1.Principal, runID string) *controlplanev1.GetRuntimeDeployTaskRequest {
	return &controlplanev1.GetRuntimeDeployTaskRequest{
		Principal: principal,
		RunId:     strings.TrimSpace(runID),
	}
}

func buildListRegistryImagesRequest(principal *controlplanev1.Principal, arg registryImagesListArg) *controlplanev1.ListRegistryImagesRequest {
	return &controlplanev1.ListRegistryImagesRequest{
		Principal:         principal,
		Repository:        optionalStringPtr(arg.repository),
		LimitRepositories: arg.limitRepositories,
		LimitTags:         arg.limitTags,
	}
}

func buildDeleteRegistryImageTagRequest(principal *controlplanev1.Principal, req models.DeleteRegistryImageTagRequest) *controlplanev1.DeleteRegistryImageTagRequest {
	return &controlplanev1.DeleteRegistryImageTagRequest{
		Principal:  principal,
		Repository: strings.TrimSpace(req.Repository),
		Tag:        strings.TrimSpace(req.Tag),
	}
}

func buildCleanupRegistryImagesRequest(principal *controlplanev1.Principal, req models.CleanupRegistryImagesRequest) *controlplanev1.CleanupRegistryImagesRequest {
	return &controlplanev1.CleanupRegistryImagesRequest{
		Principal:         principal,
		RepositoryPrefix:  optionalStringPtr(req.RepositoryPrefix),
		LimitRepositories: req.LimitRepositories,
		KeepTags:          req.KeepTags,
		DryRun:            req.DryRun,
	}
}

func buildListUsersRequest(principal *controlplanev1.Principal, limit int32) *controlplanev1.ListUsersRequest {
	return &controlplanev1.ListUsersRequest{Principal: principal, Limit: limit}
}

func buildCreateUserRequest(principal *controlplanev1.Principal, req models.CreateUserRequest) *controlplanev1.CreateUserRequest {
	return &controlplanev1.CreateUserRequest{Principal: principal, Email: req.Email, IsPlatformAdmin: req.IsPlatformAdmin}
}

func buildListProjectMembersRequest(principal *controlplanev1.Principal, arg idLimitArg) *controlplanev1.ListProjectMembersRequest {
	return &controlplanev1.ListProjectMembersRequest{Principal: principal, ProjectId: arg.id, Limit: arg.limit}
}

func buildListProjectRepositoriesRequest(principal *controlplanev1.Principal, arg idLimitArg) *controlplanev1.ListProjectRepositoriesRequest {
	return &controlplanev1.ListProjectRepositoriesRequest{Principal: principal, ProjectId: arg.id, Limit: arg.limit}
}

func (h *staffHandler) listProjectsCall(ctx context.Context, principal *controlplanev1.Principal, limit int32) (*controlplanev1.ListProjectsResponse, error) {
	return callUnaryWithArg(ctx, principal, limit, buildListProjectsRequest, h.cp.Service().ListProjects)
}

func (h *staffHandler) getProjectCall(ctx context.Context, principal *controlplanev1.Principal, id string) (*controlplanev1.Project, error) {
	return callUnaryWithArg(ctx, principal, id, buildGetProjectRequest, h.cp.Service().GetProject)
}

func (h *staffHandler) upsertProjectCall(ctx context.Context, principal *controlplanev1.Principal, req models.UpsertProjectRequest) (*controlplanev1.Project, error) {
	return callUnaryWithArg(ctx, principal, req, buildUpsertProjectRequest, h.cp.Service().UpsertProject)
}

func (h *staffHandler) listRunsCall(ctx context.Context, principal *controlplanev1.Principal, limit int32) (*controlplanev1.ListRunsResponse, error) {
	return callUnaryWithArg(ctx, principal, limit, buildListRunsRequest, h.cp.Service().ListRuns)
}

func (h *staffHandler) listRunJobsCall(ctx context.Context, principal *controlplanev1.Principal, arg runListFilterArg) (*controlplanev1.ListRunJobsResponse, error) {
	return callUnaryWithArg(ctx, principal, arg, buildListRunJobsRequest, h.cp.Service().ListRunJobs)
}

func (h *staffHandler) listRunJobsAsGetter(ctx context.Context, principal *controlplanev1.Principal, arg runListFilterArg) (runItemsGetter, error) {
	return h.listRunJobsCall(ctx, principal, arg)
}

func (h *staffHandler) listRunWaitsCall(ctx context.Context, principal *controlplanev1.Principal, arg runListFilterArg) (*controlplanev1.ListRunWaitsResponse, error) {
	return callUnaryWithArg(ctx, principal, arg, buildListRunWaitsRequest, h.cp.Service().ListRunWaits)
}

func (h *staffHandler) listRunWaitsAsGetter(ctx context.Context, principal *controlplanev1.Principal, arg runListFilterArg) (runItemsGetter, error) {
	return h.listRunWaitsCall(ctx, principal, arg)
}

func (h *staffHandler) listPendingApprovalsCall(ctx context.Context, principal *controlplanev1.Principal, limit int32) (*controlplanev1.ListPendingApprovalsResponse, error) {
	return callUnaryWithArg(ctx, principal, limit, buildListPendingApprovalsRequest, h.cp.Service().ListPendingApprovals)
}

func (h *staffHandler) getRunCall(ctx context.Context, principal *controlplanev1.Principal, id string) (*controlplanev1.Run, error) {
	return callUnaryWithArg(ctx, principal, id, buildGetRunRequest, h.cp.Service().GetRun)
}

func (h *staffHandler) getRunLogsCall(ctx context.Context, principal *controlplanev1.Principal, arg runLogsArg) (*controlplanev1.RunLogs, error) {
	return callUnaryWithArg(ctx, principal, arg, buildGetRunLogsRequest, h.cp.Service().GetRunLogs)
}

func (h *staffHandler) resolveApprovalDecisionCall(
	ctx context.Context,
	principal *controlplanev1.Principal,
	arg approvalDecisionArg,
) (*controlplanev1.ResolveApprovalDecisionResponse, error) {
	return callUnaryWithArg(ctx, principal, arg, buildResolveApprovalDecisionRequest, h.cp.Service().ResolveApprovalDecision)
}

func (h *staffHandler) deleteRunNamespaceCall(ctx context.Context, principal *controlplanev1.Principal, id string) (*controlplanev1.DeleteRunNamespaceResponse, error) {
	return callUnaryWithArg(ctx, principal, id, buildDeleteRunNamespaceRequest, h.cp.Service().DeleteRunNamespace)
}

func (h *staffHandler) listRunEventsCall(ctx context.Context, principal *controlplanev1.Principal, id string, limit int32) (*controlplanev1.ListRunEventsResponse, error) {
	arg := idLimitArg{id: id, limit: limit}
	return callUnaryWithArg(ctx, principal, arg, buildListRunEventsRequest, h.cp.Service().ListRunEvents)
}

func (h *staffHandler) listRunLearningFeedbackCall(ctx context.Context, principal *controlplanev1.Principal, id string, limit int32) (*controlplanev1.ListRunLearningFeedbackResponse, error) {
	req := buildListRunLearningFeedbackRequest(principal, idLimitArg{id: id, limit: limit})
	return h.cp.Service().ListRunLearningFeedback(ctx, req)
}

func (h *staffHandler) listRuntimeDeployTasksCall(ctx context.Context, principal *controlplanev1.Principal, arg runtimeDeployListArg) (*controlplanev1.ListRuntimeDeployTasksResponse, error) {
	return callUnaryWithArg(ctx, principal, arg, buildListRuntimeDeployTasksRequest, h.cp.Service().ListRuntimeDeployTasks)
}

func (h *staffHandler) getRuntimeDeployTaskCall(ctx context.Context, principal *controlplanev1.Principal, runID string) (*controlplanev1.RuntimeDeployTask, error) {
	return callUnaryWithArg(ctx, principal, runID, buildGetRuntimeDeployTaskRequest, h.cp.Service().GetRuntimeDeployTask)
}

func (h *staffHandler) listRegistryImagesCall(ctx context.Context, principal *controlplanev1.Principal, arg registryImagesListArg) (*controlplanev1.ListRegistryImagesResponse, error) {
	return callUnaryWithArg(ctx, principal, arg, buildListRegistryImagesRequest, h.cp.Service().ListRegistryImages)
}

func (h *staffHandler) deleteRegistryImageTagCall(ctx context.Context, principal *controlplanev1.Principal, req models.DeleteRegistryImageTagRequest) (*controlplanev1.RegistryImageDeleteResult, error) {
	return callUnaryWithArg(ctx, principal, req, buildDeleteRegistryImageTagRequest, h.cp.Service().DeleteRegistryImageTag)
}

func (h *staffHandler) cleanupRegistryImagesCall(ctx context.Context, principal *controlplanev1.Principal, req models.CleanupRegistryImagesRequest) (*controlplanev1.CleanupRegistryImagesResponse, error) {
	return callUnaryWithArg(ctx, principal, req, buildCleanupRegistryImagesRequest, h.cp.Service().CleanupRegistryImages)
}

func (h *staffHandler) listUsersCall(ctx context.Context, principal *controlplanev1.Principal, limit int32) (*controlplanev1.ListUsersResponse, error) {
	return callUnaryWithArg(ctx, principal, limit, buildListUsersRequest, h.cp.Service().ListUsers)
}

func (h *staffHandler) createUserCall(ctx context.Context, principal *controlplanev1.Principal, req models.CreateUserRequest) (*controlplanev1.User, error) {
	return callUnaryWithArg(ctx, principal, req, buildCreateUserRequest, h.cp.Service().CreateUser)
}

func (h *staffHandler) listProjectMembersCall(ctx context.Context, principal *controlplanev1.Principal, id string, limit int32) (*controlplanev1.ListProjectMembersResponse, error) {
	builder := buildListProjectMembersRequest
	return callUnaryWithArg(ctx, principal, idLimitArg{id: id, limit: limit}, builder, h.cp.Service().ListProjectMembers)
}

func (h *staffHandler) listProjectRepositoriesCall(ctx context.Context, principal *controlplanev1.Principal, id string, limit int32) (*controlplanev1.ListProjectRepositoriesResponse, error) {
	req := buildListProjectRepositoriesRequest(principal, idLimitArg{id: id, limit: limit})
	svc := h.cp.Service()
	return svc.ListProjectRepositories(ctx, req)
}

func (h *staffHandler) deleteProject(ctx context.Context, principal *controlplanev1.Principal, id string) error {
	req := &controlplanev1.DeleteProjectRequest{Principal: principal, ProjectId: id}
	_, err := h.cp.Service().DeleteProject(ctx, req)
	return err
}

func (h *staffHandler) deleteUser(ctx context.Context, principal *controlplanev1.Principal, id string) error {
	_, err := h.cp.Service().DeleteUser(ctx, &controlplanev1.DeleteUserRequest{Principal: principal, UserId: id})
	return err
}

func (h *staffHandler) deleteProjectMember(ctx context.Context, principal *controlplanev1.Principal, projectID string, userID string) error {
	_, err := h.cp.Service().DeleteProjectMember(ctx, &controlplanev1.DeleteProjectMemberRequest{Principal: principal, ProjectId: projectID, UserId: userID})
	return err
}

func (h *staffHandler) deleteProjectRepository(ctx context.Context, principal *controlplanev1.Principal, projectID string, repositoryID string) error {
	req := controlplanev1.DeleteProjectRepositoryRequest{Principal: principal, ProjectId: projectID, RepositoryId: repositoryID}
	_, err := h.cp.Service().DeleteProjectRepository(ctx, &req)
	return err
}

func optionalStringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
