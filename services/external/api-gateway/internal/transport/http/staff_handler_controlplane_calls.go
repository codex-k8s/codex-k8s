package http

import (
	"context"
	"strings"

	"google.golang.org/grpc"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/models"
)

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

func buildListRunsRequest(principal *controlplanev1.Principal, arg runListPageArg) *controlplanev1.ListRunsRequest {
	return &controlplanev1.ListRunsRequest{
		Principal: principal,
		Page:      arg.page,
		PageSize:  arg.pageSize,
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

func buildListRunEventsRequest(principal *controlplanev1.Principal, arg runEventsArg) *controlplanev1.ListRunEventsRequest {
	return &controlplanev1.ListRunEventsRequest{Principal: principal, RunId: arg.runID, Limit: arg.limit}
}

func buildListRunLearningFeedbackRequest(principal *controlplanev1.Principal, arg idLimitArg) *controlplanev1.ListRunLearningFeedbackRequest {
	return &controlplanev1.ListRunLearningFeedbackRequest{Principal: principal, RunId: arg.id, Limit: arg.limit}
}

func buildListRuntimeDeployTasksRequest(principal *controlplanev1.Principal, arg runtimeDeployListArg) *controlplanev1.ListRuntimeDeployTasksRequest {
	return &controlplanev1.ListRuntimeDeployTasksRequest{
		Principal: principal,
		Page:      arg.page,
		PageSize:  arg.pageSize,
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

func buildCancelRuntimeDeployTaskRequest(principal *controlplanev1.Principal, arg runtimeDeployActionArg) *controlplanev1.CancelRuntimeDeployTaskRequest {
	return &controlplanev1.CancelRuntimeDeployTaskRequest{
		Principal: principal,
		RunId:     strings.TrimSpace(arg.runID),
		Reason:    optionalStringPtr(arg.body.Reason),
	}
}

func buildStopRuntimeDeployTaskRequest(principal *controlplanev1.Principal, arg runtimeDeployActionArg) *controlplanev1.StopRuntimeDeployTaskRequest {
	return &controlplanev1.StopRuntimeDeployTaskRequest{
		Principal: principal,
		RunId:     strings.TrimSpace(arg.runID),
		Reason:    optionalStringPtr(arg.body.Reason),
		Force:     arg.body.Force,
	}
}

func buildListRuntimeErrorsRequest(principal *controlplanev1.Principal, arg runtimeErrorsListArg) *controlplanev1.ListRuntimeErrorsRequest {
	return &controlplanev1.ListRuntimeErrorsRequest{
		Principal:     principal,
		Limit:         arg.limit,
		State:         optionalStringPtr(arg.state),
		Level:         optionalStringPtr(arg.level),
		Source:        optionalStringPtr(arg.source),
		RunId:         optionalStringPtr(arg.runID),
		CorrelationId: optionalStringPtr(arg.correlationID),
	}
}

func buildMarkRuntimeErrorViewedRequest(principal *controlplanev1.Principal, runtimeErrorID string) *controlplanev1.MarkRuntimeErrorViewedRequest {
	return &controlplanev1.MarkRuntimeErrorViewedRequest{
		Principal:      principal,
		RuntimeErrorId: strings.TrimSpace(runtimeErrorID),
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

func buildGetMissionControlDashboardRequest(principal *controlplanev1.Principal, arg missionControlDashboardArg) *controlplanev1.GetMissionControlSnapshotRequest {
	return &controlplanev1.GetMissionControlSnapshotRequest{
		Principal:    principal,
		ViewMode:     optionalStringPtr(arg.viewMode),
		ActiveFilter: optionalStringPtr(arg.activeFilter),
		Search:       optionalStringPtr(arg.search),
		Cursor:       optionalStringPtr(arg.cursor),
		Limit:        arg.limit,
	}
}

func buildGetMissionControlEntityRequest(principal *controlplanev1.Principal, arg missionControlEntityArg) *controlplanev1.GetMissionControlEntityRequest {
	return &controlplanev1.GetMissionControlEntityRequest{
		Principal:      principal,
		EntityKind:     strings.TrimSpace(arg.entityKind),
		EntityPublicId: strings.TrimSpace(arg.entityPublicID),
		TimelineLimit:  arg.timelineLimit,
	}
}

func buildListMissionControlTimelineRequest(principal *controlplanev1.Principal, arg missionControlTimelineArg) *controlplanev1.ListMissionControlTimelineRequest {
	return &controlplanev1.ListMissionControlTimelineRequest{
		Principal:      principal,
		EntityKind:     strings.TrimSpace(arg.entityKind),
		EntityPublicId: strings.TrimSpace(arg.entityPublicID),
		Cursor:         optionalStringPtr(arg.cursor),
		Limit:          arg.limit,
	}
}

func buildGetMissionControlCommandRequest(principal *controlplanev1.Principal, commandID string) *controlplanev1.GetMissionControlCommandRequest {
	return &controlplanev1.GetMissionControlCommandRequest{
		Principal: principal,
		CommandId: strings.TrimSpace(commandID),
	}
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

func (h *staffHandler) listRunsCall(ctx context.Context, principal *controlplanev1.Principal, arg runListPageArg) (*controlplanev1.ListRunsResponse, error) {
	return callUnaryWithArg(ctx, principal, arg, buildListRunsRequest, h.cp.Service().ListRuns)
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
	item, err := callUnaryWithArg(ctx, principal, arg, buildGetRunLogsRequest, h.cp.Service().GetRunLogs)
	if err != nil {
		return nil, err
	}
	if !arg.includeSnapshot && item != nil {
		item.SnapshotJson = "{}"
	}
	return item, nil
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

func (h *staffHandler) listRunEventsCall(ctx context.Context, principal *controlplanev1.Principal, arg runEventsArg) (*controlplanev1.ListRunEventsResponse, error) {
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

func (h *staffHandler) cancelRuntimeDeployTaskCall(ctx context.Context, principal *controlplanev1.Principal, arg runtimeDeployActionArg) (*controlplanev1.RuntimeDeployTaskActionResponse, error) {
	return callUnaryWithArg(ctx, principal, arg, buildCancelRuntimeDeployTaskRequest, h.cp.Service().CancelRuntimeDeployTask)
}

func (h *staffHandler) stopRuntimeDeployTaskCall(ctx context.Context, principal *controlplanev1.Principal, arg runtimeDeployActionArg) (*controlplanev1.RuntimeDeployTaskActionResponse, error) {
	return callUnaryWithArg(ctx, principal, arg, buildStopRuntimeDeployTaskRequest, h.cp.Service().StopRuntimeDeployTask)
}

func (h *staffHandler) listRuntimeErrorsCall(ctx context.Context, principal *controlplanev1.Principal, arg runtimeErrorsListArg) (*controlplanev1.ListRuntimeErrorsResponse, error) {
	return callUnaryWithArg(ctx, principal, arg, buildListRuntimeErrorsRequest, h.cp.Service().ListRuntimeErrors)
}

func (h *staffHandler) markRuntimeErrorViewedCall(ctx context.Context, principal *controlplanev1.Principal, runtimeErrorID string) (*controlplanev1.RuntimeError, error) {
	return callUnaryWithArg(ctx, principal, runtimeErrorID, buildMarkRuntimeErrorViewedRequest, h.cp.Service().MarkRuntimeErrorViewed)
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

func (h *staffHandler) getMissionControlDashboardCall(ctx context.Context, principal *controlplanev1.Principal, arg missionControlDashboardArg) (*controlplanev1.GetMissionControlSnapshotResponse, error) {
	return callUnaryWithArg(ctx, principal, arg, buildGetMissionControlDashboardRequest, h.cp.Service().GetMissionControlSnapshot)
}

func (h *staffHandler) getMissionControlEntityCall(ctx context.Context, principal *controlplanev1.Principal, arg missionControlEntityArg) (*controlplanev1.MissionControlEntityDetails, error) {
	return callUnaryWithArg(ctx, principal, arg, buildGetMissionControlEntityRequest, h.cp.Service().GetMissionControlEntity)
}

func (h *staffHandler) listMissionControlTimelineCall(ctx context.Context, principal *controlplanev1.Principal, arg missionControlTimelineArg) (*controlplanev1.ListMissionControlTimelineResponse, error) {
	return callUnaryWithArg(ctx, principal, arg, buildListMissionControlTimelineRequest, h.cp.Service().ListMissionControlTimeline)
}

func (h *staffHandler) getMissionControlCommandCall(ctx context.Context, principal *controlplanev1.Principal, commandID string) (*controlplanev1.MissionControlCommandState, error) {
	return callUnaryWithArg(ctx, principal, commandID, buildGetMissionControlCommandRequest, h.cp.Service().GetMissionControlCommand)
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
