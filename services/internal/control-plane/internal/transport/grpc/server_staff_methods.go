package grpc

import (
	"context"
	"encoding/json"
	"strings"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	configentryrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/configentry"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Server) ListProjects(ctx context.Context, req *controlplanev1.ListProjectsRequest) (*controlplanev1.ListProjectsResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	limit := clampLimit(req.Limit, 200)
	items, err := s.staff.ListProjects(ctx, p, limit)
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.Project, 0, len(items))
	for _, it := range items {
		out = append(out, &controlplanev1.Project{
			Id:   it.ID,
			Slug: it.Slug,
			Name: it.Name,
			Role: it.Role,
		})
	}
	return &controlplanev1.ListProjectsResponse{Items: out}, nil
}

func (s *Server) UpsertProject(ctx context.Context, req *controlplanev1.UpsertProjectRequest) (*controlplanev1.Project, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	item, err := s.staff.UpsertProject(ctx, p, strings.TrimSpace(req.Slug), strings.TrimSpace(req.Name))
	if err != nil {
		return nil, toStatus(err)
	}
	role := ""
	if p.IsPlatformAdmin {
		role = "admin"
	}
	return &controlplanev1.Project{Id: item.ID, Slug: item.Slug, Name: item.Name, Role: role}, nil
}

func (s *Server) GetProject(ctx context.Context, req *controlplanev1.GetProjectRequest) (*controlplanev1.Project, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	item, err := s.staff.GetProject(ctx, p, strings.TrimSpace(req.ProjectId))
	if err != nil {
		return nil, toStatus(err)
	}
	role := ""
	if p.IsPlatformAdmin {
		role = "admin"
	}
	return &controlplanev1.Project{Id: item.ID, Slug: item.Slug, Name: item.Name, Role: role}, nil
}

func (s *Server) DeleteProject(ctx context.Context, req *controlplanev1.DeleteProjectRequest) (*emptypb.Empty, error) {
	return s.delete1(ctx, req.GetPrincipal(), req.ProjectId, s.staff.DeleteProject)
}

func (s *Server) ListRuns(ctx context.Context, req *controlplanev1.ListRunsRequest) (*controlplanev1.ListRunsResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	limit := clampLimit(req.Limit, 200)
	items, err := s.staff.ListRuns(ctx, p, limit)
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.Run, 0, len(items))
	for _, r := range items {
		out = append(out, runToProto(r))
	}
	return &controlplanev1.ListRunsResponse{Items: out}, nil
}

func (s *Server) ListRunJobs(ctx context.Context, req *controlplanev1.ListRunJobsRequest) (*controlplanev1.ListRunJobsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	items, err := s.staff.ListRunJobs(ctx, p, querytypes.StaffRunListFilter{
		Limit:       clampLimit(req.GetLimit(), 200),
		TriggerKind: optionalProtoString(req.TriggerKind),
		Status:      optionalProtoString(req.Status),
		AgentKey:    optionalProtoString(req.AgentKey),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.Run, 0, len(items))
	for _, item := range items {
		out = append(out, runToProto(item))
	}
	return &controlplanev1.ListRunJobsResponse{Items: out}, nil
}

func (s *Server) ListRunWaits(ctx context.Context, req *controlplanev1.ListRunWaitsRequest) (*controlplanev1.ListRunWaitsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	items, err := s.staff.ListRunWaits(ctx, p, querytypes.StaffRunListFilter{
		Limit:       clampLimit(req.GetLimit(), 200),
		TriggerKind: optionalProtoString(req.TriggerKind),
		Status:      optionalProtoString(req.Status),
		AgentKey:    optionalProtoString(req.AgentKey),
		WaitState:   optionalProtoString(req.WaitState),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.Run, 0, len(items))
	for _, item := range items {
		out = append(out, runToProto(item))
	}
	return &controlplanev1.ListRunWaitsResponse{Items: out}, nil
}

func (s *Server) GetRun(ctx context.Context, req *controlplanev1.GetRunRequest) (*controlplanev1.Run, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	r, err := s.staff.GetRun(ctx, p, strings.TrimSpace(req.RunId))
	if err != nil {
		return nil, toStatus(err)
	}
	return runToProto(r), nil
}

func (s *Server) GetRunLogs(ctx context.Context, req *controlplanev1.GetRunLogsRequest) (*controlplanev1.RunLogs, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	item, err := s.staff.GetRunLogs(ctx, p, strings.TrimSpace(req.RunId), int(req.GetTailLines()))
	if err != nil {
		return nil, toStatus(err)
	}

	snapshotJSON := strings.TrimSpace(string(item.SnapshotJSON))
	if snapshotJSON == "" {
		snapshotJSON = "{}"
	}
	out := &controlplanev1.RunLogs{
		RunId:        item.RunID,
		Status:       item.Status,
		SnapshotJson: snapshotJSON,
		TailLines:    item.TailLines,
	}
	if item.UpdatedAt != nil {
		out.UpdatedAt = timestamppb.New(item.UpdatedAt.UTC())
	}
	return out, nil
}

func (s *Server) ListPendingApprovals(ctx context.Context, req *controlplanev1.ListPendingApprovalsRequest) (*controlplanev1.ListPendingApprovalsResponse, error) {
	if s.mcp == nil {
		return nil, status.Error(codes.FailedPrecondition, "mcp service is not configured")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	if !p.IsPlatformAdmin {
		return nil, status.Error(codes.PermissionDenied, "platform admin required")
	}

	limit := clampLimit(req.GetLimit(), 200)
	items, err := s.mcp.ListPendingApprovals(ctx, limit)
	if err != nil {
		return nil, toStatus(err)
	}

	out := make([]*controlplanev1.ApprovalRequest, 0, len(items))
	for _, item := range items {
		out = append(out, approvalToProto(item))
	}
	return &controlplanev1.ListPendingApprovalsResponse{Items: out}, nil
}

func (s *Server) ResolveApprovalDecision(ctx context.Context, req *controlplanev1.ResolveApprovalDecisionRequest) (*controlplanev1.ResolveApprovalDecisionResponse, error) {
	if s.mcp == nil {
		return nil, status.Error(codes.FailedPrecondition, "mcp service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	if !p.IsPlatformAdmin {
		return nil, status.Error(codes.PermissionDenied, "platform admin required")
	}

	if req.GetApprovalRequestId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "approval_request_id is required")
	}
	decision := strings.TrimSpace(req.GetDecision())
	if decision == "" {
		return nil, status.Error(codes.InvalidArgument, "decision is required")
	}

	item, err := s.mcp.ResolveApproval(ctx, mcpdomain.ResolveApprovalParams{
		RequestID: req.GetApprovalRequestId(),
		Decision:  mcpdomain.ApprovalDecision(decision),
		ActorID:   approvalActorID(p),
		Reason:    strings.TrimSpace(req.GetReason()),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &controlplanev1.ResolveApprovalDecisionResponse{
		Id:            item.ID,
		CorrelationId: item.CorrelationID,
		RunId:         stringPtrOrNil(item.RunID),
		ToolName:      item.ToolName,
		Action:        item.Action,
		ApprovalState: item.ApprovalState,
	}, nil
}

func (s *Server) ListRunEvents(ctx context.Context, req *controlplanev1.ListRunEventsRequest) (*controlplanev1.ListRunEventsResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	limit := clampLimit(req.Limit, 500)
	items, err := s.staff.ListRunFlowEvents(ctx, p, strings.TrimSpace(req.RunId), limit)
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.FlowEvent, 0, len(items))
	for _, e := range items {
		out = append(out, &controlplanev1.FlowEvent{
			CorrelationId: e.CorrelationID,
			EventType:     e.EventType,
			CreatedAt:     timestamppb.New(e.CreatedAt.UTC()),
			PayloadJson:   string(e.PayloadJSON),
		})
	}
	return &controlplanev1.ListRunEventsResponse{Items: out}, nil
}

func (s *Server) ListRunLearningFeedback(ctx context.Context, req *controlplanev1.ListRunLearningFeedbackRequest) (*controlplanev1.ListRunLearningFeedbackResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	limit := clampLimit(req.Limit, 200)
	items, err := s.staff.ListRunLearningFeedback(ctx, p, strings.TrimSpace(req.RunId), limit)
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.LearningFeedback, 0, len(items))
	for _, f := range items {
		out = append(out, &controlplanev1.LearningFeedback{
			Id:           f.ID,
			RunId:        f.RunID,
			RepositoryId: stringPtrOrNil(f.RepositoryID),
			PrNumber:     int32PtrOrNil(int32(f.PRNumber)),
			FilePath:     stringPtrOrNil(f.FilePath),
			Line:         int32PtrOrNil(int32(f.Line)),
			Kind:         f.Kind,
			Explanation:  f.Explanation,
			CreatedAt:    timestamppb.New(f.CreatedAt.UTC()),
		})
	}
	return &controlplanev1.ListRunLearningFeedbackResponse{Items: out}, nil
}

func (s *Server) DeleteRunNamespace(ctx context.Context, req *controlplanev1.DeleteRunNamespaceRequest) (*controlplanev1.DeleteRunNamespaceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}

	result, err := s.staff.DeleteRunNamespace(ctx, p, strings.TrimSpace(req.GetRunId()))
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.DeleteRunNamespaceResponse{
		Ok:             true,
		RunId:          result.RunID,
		Namespace:      result.Namespace,
		Deleted:        result.Deleted,
		AlreadyDeleted: result.AlreadyDeleted,
		CommentUrl:     stringPtrOrNil(result.CommentURL),
	}, nil
}

func (s *Server) ListUsers(ctx context.Context, req *controlplanev1.ListUsersRequest) (*controlplanev1.ListUsersResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	limit := clampLimit(req.Limit, 200)
	items, err := s.staff.ListUsers(ctx, p, limit)
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.User, 0, len(items))
	for _, u := range items {
		out = append(out, &controlplanev1.User{
			Id:              u.ID,
			Email:           u.Email,
			GithubUserId:    int64PtrOrNil(u.GitHubUserID),
			GithubLogin:     stringPtrOrNil(u.GitHubLogin),
			IsPlatformAdmin: u.IsPlatformAdmin,
			IsPlatformOwner: u.IsPlatformOwner,
		})
	}
	return &controlplanev1.ListUsersResponse{Items: out}, nil
}

func (s *Server) CreateUser(ctx context.Context, req *controlplanev1.CreateUserRequest) (*controlplanev1.User, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	u, err := s.staff.CreateAllowedUser(ctx, p, strings.TrimSpace(req.Email), req.IsPlatformAdmin)
	if err != nil {
		return nil, toStatus(err)
	}
	return &controlplanev1.User{
		Id:              u.ID,
		Email:           u.Email,
		GithubUserId:    int64PtrOrNil(u.GitHubUserID),
		GithubLogin:     stringPtrOrNil(u.GitHubLogin),
		IsPlatformAdmin: u.IsPlatformAdmin,
		IsPlatformOwner: u.IsPlatformOwner,
	}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *controlplanev1.DeleteUserRequest) (*emptypb.Empty, error) {
	return s.delete1(ctx, req.GetPrincipal(), req.UserId, s.staff.DeleteUser)
}

func (s *Server) ListProjectMembers(ctx context.Context, req *controlplanev1.ListProjectMembersRequest) (*controlplanev1.ListProjectMembersResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	limit := clampLimit(req.Limit, 200)
	items, err := s.staff.ListProjectMembers(ctx, p, strings.TrimSpace(req.ProjectId), limit)
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.ProjectMember, 0, len(items))
	for _, m := range items {
		var override *wrapperspb.BoolValue
		if m.LearningModeOverride != nil {
			override = wrapperspb.Bool(*m.LearningModeOverride)
		}
		out = append(out, &controlplanev1.ProjectMember{
			ProjectId:            m.ProjectID,
			UserId:               m.UserID,
			Email:                m.Email,
			Role:                 m.Role,
			LearningModeOverride: override,
		})
	}
	return &controlplanev1.ListProjectMembersResponse{Items: out}, nil
}

func (s *Server) UpsertProjectMember(ctx context.Context, req *controlplanev1.UpsertProjectMemberRequest) (*emptypb.Empty, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}

	projectID := strings.TrimSpace(req.GetProjectId())
	userID := strings.TrimSpace(req.GetUserId())
	email := strings.TrimSpace(req.GetEmail())
	role := strings.TrimSpace(req.GetRole())

	if email != "" {
		if err := s.staff.UpsertProjectMemberByEmail(ctx, p, projectID, email, role); err != nil {
			return nil, toStatus(err)
		}
		return &emptypb.Empty{}, nil
	}

	if err := s.staff.UpsertProjectMember(ctx, p, projectID, userID, role); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteProjectMember(ctx context.Context, req *controlplanev1.DeleteProjectMemberRequest) (*emptypb.Empty, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	if err := s.staff.DeleteProjectMember(ctx, p, strings.TrimSpace(req.ProjectId), strings.TrimSpace(req.UserId)); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) SetProjectMemberLearningModeOverride(ctx context.Context, req *controlplanev1.SetProjectMemberLearningModeOverrideRequest) (*emptypb.Empty, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	var enabled *bool
	if req.Enabled != nil {
		v := req.Enabled.Value
		enabled = &v
	}
	if err := s.staff.SetProjectMemberLearningModeOverride(ctx, p, strings.TrimSpace(req.ProjectId), strings.TrimSpace(req.UserId), enabled); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListProjectRepositories(ctx context.Context, req *controlplanev1.ListProjectRepositoriesRequest) (*controlplanev1.ListProjectRepositoriesResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	limit := clampLimit(req.Limit, 200)
	items, err := s.staff.ListProjectRepositories(ctx, p, strings.TrimSpace(req.ProjectId), limit)
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.RepositoryBinding, 0, len(items))
	for _, r := range items {
		out = append(out, repositoryBindingToProto(r))
	}
	return &controlplanev1.ListProjectRepositoriesResponse{Items: out}, nil
}

func (s *Server) UpsertProjectRepository(ctx context.Context, req *controlplanev1.UpsertProjectRepositoryRequest) (*controlplanev1.RepositoryBinding, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	item, err := s.staff.UpsertProjectRepository(
		ctx,
		p,
		strings.TrimSpace(req.ProjectId),
		strings.TrimSpace(req.Provider),
		strings.TrimSpace(req.Owner),
		strings.TrimSpace(req.Name),
		req.Token,
		strings.TrimSpace(req.ServicesYamlPath),
		strings.TrimSpace(req.GetAlias()),
		strings.TrimSpace(req.GetRole()),
		strings.TrimSpace(req.GetDefaultRef()),
		optionalProtoString(req.DocsRootPath),
	)
	if err != nil {
		return nil, toStatus(err)
	}
	return repositoryBindingToProto(item), nil
}

func (s *Server) DeleteProjectRepository(ctx context.Context, req *controlplanev1.DeleteProjectRepositoryRequest) (*emptypb.Empty, error) {
	return s.delete2(ctx, req.GetPrincipal(), req.ProjectId, req.RepositoryId, s.staff.DeleteProjectRepository)
}

func (s *Server) UpsertRepositoryBotParams(ctx context.Context, req *controlplanev1.UpsertRepositoryBotParamsRequest) (*emptypb.Empty, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	var botToken *string
	if req.BotToken != nil {
		v := strings.TrimSpace(*req.BotToken)
		botToken = &v
	}
	var botUsername *string
	if req.BotUsername != nil {
		v := strings.TrimSpace(*req.BotUsername)
		botUsername = &v
	}
	var botEmail *string
	if req.BotEmail != nil {
		v := strings.TrimSpace(*req.BotEmail)
		botEmail = &v
	}
	if err := s.staff.UpsertRepositoryBotParams(ctx, p, strings.TrimSpace(req.ProjectId), strings.TrimSpace(req.RepositoryId), botToken, botUsername, botEmail); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) RunRepositoryPreflight(ctx context.Context, req *controlplanev1.RunRepositoryPreflightRequest) (*controlplanev1.RunRepositoryPreflightResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	report, err := s.staff.RunRepositoryPreflight(ctx, p, strings.TrimSpace(req.ProjectId), strings.TrimSpace(req.RepositoryId))
	if err != nil {
		return nil, toStatus(err)
	}
	checks := make([]*controlplanev1.PreflightCheckResult, 0, len(report.Checks))
	for _, c := range report.Checks {
		checks = append(checks, &controlplanev1.PreflightCheckResult{
			Name:    c.Name,
			Status:  c.Status,
			Details: stringPtrOrNil(strings.TrimSpace(c.Details)),
		})
	}
	encoded, _ := json.Marshal(report)
	finished := timestamppb.New(report.FinishedAt.UTC())
	return &controlplanev1.RunRepositoryPreflightResponse{
		RepositoryId: strings.TrimSpace(req.RepositoryId),
		Status:       report.Status,
		Checks:       checks,
		ReportJson:   string(encoded),
		FinishedAt:   finished,
	}, nil
}

func (s *Server) GetProjectGitHubTokens(ctx context.Context, req *controlplanev1.GetProjectGitHubTokensRequest) (*controlplanev1.ProjectGitHubTokens, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	item, ok, err := s.staff.GetProjectGitHubTokens(ctx, p, strings.TrimSpace(req.ProjectId))
	if err != nil {
		return nil, toStatus(err)
	}
	if !ok {
		return &controlplanev1.ProjectGitHubTokens{ProjectId: strings.TrimSpace(req.ProjectId)}, nil
	}
	return &controlplanev1.ProjectGitHubTokens{
		ProjectId:        item.ProjectID,
		HasPlatformToken: item.HasPlatformToken,
		HasBotToken:      item.HasBotToken,
		BotUsername:      stringPtrOrNil(strings.TrimSpace(item.BotUsername)),
		BotEmail:         stringPtrOrNil(strings.TrimSpace(item.BotEmail)),
	}, nil
}

func (s *Server) UpsertProjectGitHubTokens(ctx context.Context, req *controlplanev1.UpsertProjectGitHubTokensRequest) (*emptypb.Empty, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	var platformToken *string
	if req.PlatformToken != nil {
		v := strings.TrimSpace(*req.PlatformToken)
		platformToken = &v
	}
	var botToken *string
	if req.BotToken != nil {
		v := strings.TrimSpace(*req.BotToken)
		botToken = &v
	}
	var botUsername *string
	if req.BotUsername != nil {
		v := strings.TrimSpace(*req.BotUsername)
		botUsername = &v
	}
	var botEmail *string
	if req.BotEmail != nil {
		v := strings.TrimSpace(*req.BotEmail)
		botEmail = &v
	}
	if err := s.staff.UpsertProjectGitHubTokens(ctx, p, strings.TrimSpace(req.ProjectId), platformToken, botToken, botUsername, botEmail); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) TransitionIssueStageLabel(ctx context.Context, req *controlplanev1.TransitionIssueStageLabelRequest) (*controlplanev1.TransitionIssueStageLabelResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}

	result, err := s.staff.TransitionIssueStageLabel(ctx, p, querytypes.IssueStageLabelTransitionParams{
		RepositoryFullName: strings.TrimSpace(req.RepositoryFullName),
		IssueNumber:        int(req.GetIssueNumber()),
		TargetLabel:        strings.TrimSpace(req.TargetLabel),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.TransitionIssueStageLabelResponse{
		RepositoryFullName: result.RepositoryFullName,
		IssueNumber:        int32(result.IssueNumber),
		IssueUrl:           stringPtrOrNil(strings.TrimSpace(result.IssueURL)),
		RemovedLabels:      result.RemovedLabels,
		AddedLabels:        result.AddedLabels,
		FinalLabels:        result.FinalLabels,
	}, nil
}

func (s *Server) ListConfigEntries(ctx context.Context, req *controlplanev1.ListConfigEntriesRequest) (*controlplanev1.ListConfigEntriesResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	scope := strings.TrimSpace(req.Scope)
	projectID := strings.TrimSpace(req.GetProjectId())
	repositoryID := strings.TrimSpace(req.GetRepositoryId())
	limit := clampLimit(req.Limit, 200)
	items, err := s.staff.ListConfigEntries(ctx, p, scope, projectID, repositoryID, limit)
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.ConfigEntry, 0, len(items))
	for _, item := range items {
		out = append(out, configEntryToProto(item))
	}
	return &controlplanev1.ListConfigEntriesResponse{Items: out}, nil
}

func (s *Server) UpsertConfigEntry(ctx context.Context, req *controlplanev1.UpsertConfigEntryRequest) (*controlplanev1.ConfigEntry, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	valueSecret := strings.TrimSpace(req.GetValueSecret())
	var encrypted []byte
	if valueSecret != "" {
		encrypted, err = s.staff.EncryptSecretValue(valueSecret)
		if err != nil {
			return nil, toStatus(err)
		}
	}
	item, err := s.staff.UpsertConfigEntry(ctx, p, querytypes.ConfigEntryUpsertParams{
		Scope:           enumtypes.ConfigEntryScope(strings.TrimSpace(req.Scope)),
		Kind:            enumtypes.ConfigEntryKind(strings.TrimSpace(req.Kind)),
		ProjectID:       strings.TrimSpace(req.GetProjectId()),
		RepositoryID:    strings.TrimSpace(req.GetRepositoryId()),
		Key:             strings.TrimSpace(req.Key),
		ValuePlain:      strings.TrimSpace(req.GetValuePlain()),
		ValueEncrypted:  encrypted,
		SyncTargets:     req.SyncTargets,
		Mutability:      enumtypes.ConfigEntryMutability(strings.TrimSpace(req.Mutability)),
		IsDangerous:     req.IsDangerous,
		CreatedByUserID: p.UserID,
	}, req.DangerousConfirmed)
	if err != nil {
		return nil, toStatus(err)
	}
	return configEntryToProto(item), nil
}

func repositoryBindingToProto(item repocfgrepo.RepositoryBinding) *controlplanev1.RepositoryBinding {
	botUsername := strings.TrimSpace(item.BotUsername)
	botEmail := strings.TrimSpace(item.BotEmail)
	preflightUpdatedAt := strings.TrimSpace(item.PreflightUpdatedAt)
	return &controlplanev1.RepositoryBinding{
		Id:                 item.ID,
		ProjectId:          item.ProjectID,
		Alias:              item.Alias,
		Role:               item.Role,
		DefaultRef:         item.DefaultRef,
		Provider:           item.Provider,
		ExternalId:         item.ExternalID,
		Owner:              item.Owner,
		Name:               item.Name,
		ServicesYamlPath:   item.ServicesYAMLPath,
		DocsRootPath:       stringPtrOrNil(strings.TrimSpace(item.DocsRootPath)),
		BotUsername:        stringPtrOrNil(botUsername),
		BotEmail:           stringPtrOrNil(botEmail),
		PreflightUpdatedAt: stringPtrOrNil(preflightUpdatedAt),
	}
}

func configEntryToProto(item configentryrepo.ConfigEntry) *controlplanev1.ConfigEntry {
	return &controlplanev1.ConfigEntry{
		Id:           item.ID,
		Scope:        string(item.Scope),
		Kind:         string(item.Kind),
		ProjectId:    stringPtrOrNil(strings.TrimSpace(item.ProjectID)),
		RepositoryId: stringPtrOrNil(strings.TrimSpace(item.RepositoryID)),
		Key:          item.Key,
		Value:        stringPtrOrNil(strings.TrimSpace(item.Value)),
		SyncTargets:  item.SyncTargets,
		Mutability:   string(item.Mutability),
		IsDangerous:  item.IsDangerous,
		UpdatedAt:    stringPtrOrNil(strings.TrimSpace(item.UpdatedAt)),
	}
}

func (s *Server) DeleteConfigEntry(ctx context.Context, req *controlplanev1.DeleteConfigEntryRequest) (*emptypb.Empty, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	if err := s.staff.DeleteConfigEntry(ctx, p, strings.TrimSpace(req.ConfigEntryId)); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListDocsetGroups(ctx context.Context, req *controlplanev1.ListDocsetGroupsRequest) (*controlplanev1.ListDocsetGroupsResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	groups, err := s.staff.ListDocsetGroups(ctx, p, strings.TrimSpace(req.DocsetRef), strings.TrimSpace(req.Locale))
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.DocsetGroup, 0, len(groups))
	for _, g := range groups {
		out = append(out, &controlplanev1.DocsetGroup{
			Id:              g.ID,
			Title:           g.Title,
			Description:     g.Description,
			DefaultSelected: g.DefaultSelected,
		})
	}
	return &controlplanev1.ListDocsetGroupsResponse{Groups: out}, nil
}

func (s *Server) ImportDocset(ctx context.Context, req *controlplanev1.ImportDocsetRequest) (*controlplanev1.ImportDocsetResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	res, err := s.staff.ImportDocset(ctx, p, strings.TrimSpace(req.ProjectId), strings.TrimSpace(req.RepositoryId), strings.TrimSpace(req.DocsetRef), strings.TrimSpace(req.Locale), req.GroupIds)
	if err != nil {
		return nil, toStatus(err)
	}
	return &controlplanev1.ImportDocsetResponse{
		RepositoryFullName: res.RepositoryFullName,
		PrNumber:           int32(res.PRNumber),
		PrUrl:              res.PRURL,
		Branch:             res.Branch,
		FilesTotal:         int32(res.FilesTotal),
	}, nil
}

func (s *Server) SyncDocset(ctx context.Context, req *controlplanev1.SyncDocsetRequest) (*controlplanev1.SyncDocsetResponse, error) {
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	res, err := s.staff.SyncDocset(ctx, p, strings.TrimSpace(req.ProjectId), strings.TrimSpace(req.RepositoryId), strings.TrimSpace(req.DocsetRef))
	if err != nil {
		return nil, toStatus(err)
	}
	return &controlplanev1.SyncDocsetResponse{
		RepositoryFullName: res.RepositoryFullName,
		PrNumber:           int32(res.PRNumber),
		PrUrl:              res.PRURL,
		Branch:             res.Branch,
		FilesUpdated:       int32(res.FilesUpdated),
		FilesDrift:         int32(res.FilesDrift),
	}, nil
}
