package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	agentsessionrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/staff"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/webhook"
	agentcallback "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/transport/agentcallback"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type webhookIngress interface {
	IngestGitHubWebhook(ctx context.Context, cmd webhook.IngestCommand) (webhook.IngestResult, error)
}

type mcpRunTokenService interface {
	IssueRunToken(ctx context.Context, params mcpdomain.IssueRunTokenParams) (mcpdomain.IssuedToken, error)
	VerifyRunToken(ctx context.Context, rawToken string) (mcpdomain.SessionContext, error)
}

// Dependencies wires domain services and repositories into the gRPC transport.
type Dependencies struct {
	Webhook    webhookIngress
	Staff      *staff.Service
	Users      userrepo.Repository
	Sessions   agentsessionrepo.Repository
	FlowEvents floweventrepo.Repository
	MCP        mcpRunTokenService
	Logger     *slog.Logger
}

// Server implements ControlPlaneServiceServer.
type Server struct {
	controlplanev1.UnimplementedControlPlaneServiceServer

	webhook    webhookIngress
	staff      *staff.Service
	users      userrepo.Repository
	sessions   agentsessionrepo.Repository
	flowEvents floweventrepo.Repository
	mcp        mcpRunTokenService
	logger     *slog.Logger
}

func NewServer(deps Dependencies) *Server {
	return &Server{
		webhook:    deps.Webhook,
		staff:      deps.Staff,
		users:      deps.Users,
		sessions:   deps.Sessions,
		flowEvents: deps.FlowEvents,
		mcp:        deps.MCP,
		logger:     deps.Logger,
	}
}

func (s *Server) IngestGitHubWebhook(ctx context.Context, req *controlplanev1.IngestGitHubWebhookRequest) (*controlplanev1.IngestGitHubWebhookResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	res, err := s.webhook.IngestGitHubWebhook(ctx, webhook.IngestCommand{
		CorrelationID: strings.TrimSpace(req.CorrelationId),
		EventType:     strings.TrimSpace(req.EventType),
		DeliveryID:    strings.TrimSpace(req.DeliveryId),
		ReceivedAt:    tsToTime(req.ReceivedAt),
		Payload:       req.PayloadJson,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &controlplanev1.IngestGitHubWebhookResponse{
		CorrelationId: res.CorrelationID,
		RunId:         res.RunID,
		Status:        string(res.Status),
		Duplicate:     res.Duplicate,
	}, nil
}

func (s *Server) ResolveStaffByEmail(ctx context.Context, req *controlplanev1.ResolveStaffByEmailRequest) (*controlplanev1.ResolveStaffByEmailResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	email := strings.TrimSpace(req.Email)
	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	u, ok, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, toStatus(err)
	}
	if !ok {
		return nil, toStatus(errs.Forbidden{Msg: "email is not allowed"})
	}

	login := strings.TrimSpace(req.GetGithubLogin())
	if login != "" && !strings.EqualFold(u.GitHubLogin, login) {
		if err := s.users.UpdateGitHubIdentity(ctx, u.ID, u.GitHubUserID, login); err != nil {
			return nil, toStatus(err)
		}
		u.GitHubLogin = login
	}

	return &controlplanev1.ResolveStaffByEmailResponse{Principal: userToPrincipal(u)}, nil
}

func (s *Server) AuthorizeOAuthUser(ctx context.Context, req *controlplanev1.AuthorizeOAuthUserRequest) (*controlplanev1.AuthorizeOAuthUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	email := strings.TrimSpace(req.Email)
	if email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	login := strings.TrimSpace(req.GithubLogin)
	if login == "" {
		return nil, status.Error(codes.InvalidArgument, "github_login is required")
	}
	if req.GithubUserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "github_user_id is required")
	}

	u, ok, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, toStatus(err)
	}
	if !ok {
		return nil, toStatus(errs.Forbidden{Msg: "email is not allowed"})
	}
	if err := s.users.UpdateGitHubIdentity(ctx, u.ID, req.GithubUserId, login); err != nil {
		return nil, toStatus(err)
	}
	u.GitHubUserID = req.GithubUserId
	u.GitHubLogin = login

	return &controlplanev1.AuthorizeOAuthUserResponse{Principal: userToPrincipal(u)}, nil
}

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
		out = append(out, &controlplanev1.RepositoryBinding{
			Id:               r.ID,
			ProjectId:        r.ProjectID,
			Provider:         r.Provider,
			ExternalId:       r.ExternalID,
			Owner:            r.Owner,
			Name:             r.Name,
			ServicesYamlPath: r.ServicesYAMLPath,
		})
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
	)
	if err != nil {
		return nil, toStatus(err)
	}
	return &controlplanev1.RepositoryBinding{
		Id:               item.ID,
		ProjectId:        item.ProjectID,
		Provider:         item.Provider,
		ExternalId:       item.ExternalID,
		Owner:            item.Owner,
		Name:             item.Name,
		ServicesYamlPath: item.ServicesYAMLPath,
	}, nil
}

func (s *Server) DeleteProjectRepository(ctx context.Context, req *controlplanev1.DeleteProjectRepositoryRequest) (*emptypb.Empty, error) {
	return s.delete2(ctx, req.GetPrincipal(), req.ProjectId, req.RepositoryId, s.staff.DeleteProjectRepository)
}

func (s *Server) IssueRunMCPToken(ctx context.Context, req *controlplanev1.IssueRunMCPTokenRequest) (*controlplanev1.IssueRunMCPTokenResponse, error) {
	if s.mcp == nil {
		return nil, status.Error(codes.FailedPrecondition, "mcp service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		return nil, status.Error(codes.InvalidArgument, "run_id is required")
	}
	ttl := time.Duration(req.GetTtlSeconds()) * time.Second

	issuedToken, err := s.mcp.IssueRunToken(ctx, mcpdomain.IssueRunTokenParams{
		RunID:       runID,
		Namespace:   strings.TrimSpace(req.GetNamespace()),
		RuntimeMode: parseRuntimeMode(req.GetRuntimeMode()),
		TTL:         ttl,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.IssueRunMCPTokenResponse{
		Token:     issuedToken.Token,
		ExpiresAt: timestamppb.New(issuedToken.ExpiresAt.UTC()),
	}, nil
}

func (s *Server) UpsertAgentSession(ctx context.Context, req *controlplanev1.UpsertAgentSessionRequest) (*controlplanev1.UpsertAgentSessionResponse, error) {
	if s.sessions == nil {
		return nil, status.Error(codes.FailedPrecondition, "agent session repository is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	runSession, err := s.authenticateRunToken(ctx)
	if err != nil {
		return nil, err
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		runID = runSession.RunID
	}
	if runID == "" {
		return nil, status.Error(codes.InvalidArgument, "run_id is required")
	}
	if runID != runSession.RunID {
		return nil, status.Error(codes.PermissionDenied, "run_id mismatch with token")
	}

	repositoryFullName := strings.TrimSpace(req.GetRepositoryFullName())
	if repositoryFullName == "" {
		return nil, status.Error(codes.InvalidArgument, "repository_full_name is required")
	}
	branchName := strings.TrimSpace(req.GetBranchName())
	if branchName == "" {
		return nil, status.Error(codes.InvalidArgument, "branch_name is required")
	}
	agentKey := strings.TrimSpace(req.GetAgentKey())
	if agentKey == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_key is required")
	}

	correlationID := strings.TrimSpace(req.GetCorrelationId())
	if correlationID == "" {
		correlationID = runSession.CorrelationID
	}
	if correlationID == "" {
		return nil, status.Error(codes.InvalidArgument, "correlation_id is required")
	}

	projectID := strings.TrimSpace(req.GetProjectId())
	if projectID == "" {
		projectID = runSession.ProjectID
	}

	statusValue := strings.TrimSpace(req.GetStatus())
	if statusValue == "" {
		statusValue = sessionStatusRunning
	}

	startedAt := time.Now().UTC()
	if req.GetStartedAt() != nil {
		startedAt = req.GetStartedAt().AsTime().UTC()
	}

	if err := s.sessions.Upsert(ctx, agentsessionrepo.UpsertParams{
		RunID:              runID,
		CorrelationID:      correlationID,
		ProjectID:          projectID,
		RepositoryFullName: repositoryFullName,
		AgentKey:           agentKey,
		IssueNumber:        intPtrFromOptional(req.GetIssueNumber()),
		BranchName:         branchName,
		PRNumber:           intPtrFromOptional(req.GetPrNumber()),
		PRURL:              strings.TrimSpace(req.GetPrUrl()),
		TriggerKind:        strings.TrimSpace(req.GetTriggerKind()),
		TemplateKind:       strings.TrimSpace(req.GetTemplateKind()),
		TemplateSource:     strings.TrimSpace(req.GetTemplateSource()),
		TemplateLocale:     strings.TrimSpace(req.GetTemplateLocale()),
		Model:              strings.TrimSpace(req.GetModel()),
		ReasoningEffort:    strings.TrimSpace(req.GetReasoningEffort()),
		Status:             statusValue,
		SessionID:          strings.TrimSpace(req.GetSessionId()),
		SessionJSON:        json.RawMessage(req.GetSessionJson()),
		CodexSessionPath:   strings.TrimSpace(req.GetCodexCliSessionPath()),
		CodexSessionJSON:   json.RawMessage(req.GetCodexCliSessionJson()),
		StartedAt:          startedAt,
		FinishedAt:         optionalTime(req.GetFinishedAt()),
	}); err != nil {
		s.logger.Error("upsert agent session via grpc failed", "run_id", runID, "err", err)
		return nil, status.Error(codes.Internal, "failed to persist agent session")
	}

	return &controlplanev1.UpsertAgentSessionResponse{Ok: true, RunId: runID}, nil
}

func (s *Server) GetLatestAgentSession(ctx context.Context, req *controlplanev1.GetLatestAgentSessionRequest) (*controlplanev1.GetLatestAgentSessionResponse, error) {
	if s.sessions == nil {
		return nil, status.Error(codes.FailedPrecondition, "agent session repository is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if _, err := s.authenticateRunToken(ctx); err != nil {
		return nil, err
	}

	repositoryFullName := strings.TrimSpace(req.GetRepositoryFullName())
	branchName := strings.TrimSpace(req.GetBranchName())
	agentKey := strings.TrimSpace(req.GetAgentKey())
	if repositoryFullName == "" || branchName == "" || agentKey == "" {
		return nil, status.Error(codes.InvalidArgument, "repository_full_name, branch_name and agent_key are required")
	}

	item, found, err := s.sessions.GetLatestByRepositoryBranchAndAgent(ctx, repositoryFullName, branchName, agentKey)
	if err != nil {
		s.logger.Error("get latest agent session via grpc failed", "repository_full_name", repositoryFullName, "branch_name", branchName, "agent_key", agentKey, "err", err)
		return nil, status.Error(codes.Internal, "failed to load latest agent session")
	}
	if !found {
		return &controlplanev1.GetLatestAgentSessionResponse{Found: false}, nil
	}

	snapshot := &controlplanev1.AgentSessionSnapshot{
		RunId:               item.RunID,
		CorrelationId:       item.CorrelationID,
		ProjectId:           item.ProjectID,
		RepositoryFullName:  item.RepositoryFullName,
		AgentKey:            item.AgentKey,
		IssueNumber:         intToOptional(int32(item.IssueNumber)),
		BranchName:          item.BranchName,
		PrNumber:            intToOptional(int32(item.PRNumber)),
		PrUrl:               item.PRURL,
		TriggerKind:         item.TriggerKind,
		TemplateKind:        item.TemplateKind,
		TemplateSource:      item.TemplateSource,
		TemplateLocale:      item.TemplateLocale,
		Model:               item.Model,
		ReasoningEffort:     item.ReasoningEffort,
		Status:              item.Status,
		SessionId:           item.SessionID,
		SessionJson:         []byte(item.SessionJSON),
		CodexCliSessionPath: item.CodexSessionPath,
		CodexCliSessionJson: []byte(item.CodexSessionJSON),
		StartedAt:           timestamppb.New(item.StartedAt.UTC()),
		CreatedAt:           timestamppb.New(item.CreatedAt.UTC()),
		UpdatedAt:           timestamppb.New(item.UpdatedAt.UTC()),
	}
	if !item.FinishedAt.IsZero() {
		snapshot.FinishedAt = timestamppb.New(item.FinishedAt.UTC())
	}

	return &controlplanev1.GetLatestAgentSessionResponse{
		Found:   true,
		Session: snapshot,
	}, nil
}

func (s *Server) InsertRunFlowEvent(ctx context.Context, req *controlplanev1.InsertRunFlowEventRequest) (*controlplanev1.InsertRunFlowEventResponse, error) {
	if s.flowEvents == nil {
		return nil, status.Error(codes.FailedPrecondition, "flow event repository is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	runSession, err := s.authenticateRunToken(ctx)
	if err != nil {
		return nil, err
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		runID = runSession.RunID
	}
	if runID != runSession.RunID {
		return nil, status.Error(codes.PermissionDenied, "run_id mismatch with token")
	}

	eventType, err := agentcallback.ParseEventType(req.GetEventType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	payload := req.GetPayloadJson()
	if len(payload) == 0 {
		payload = []byte(`{}`)
	}

	if err := s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: runSession.CorrelationID,
		ActorType:     floweventdomain.ActorTypeAgent,
		ActorID:       floweventdomain.ActorIDAgentRunner,
		EventType:     eventType,
		Payload:       json.RawMessage(payload),
		CreatedAt:     time.Now().UTC(),
	}); err != nil {
		s.logger.Error("insert run flow event via grpc failed", "run_id", runID, "event_type", eventType, "err", err)
		return nil, status.Error(codes.Internal, "failed to persist flow event")
	}

	return &controlplanev1.InsertRunFlowEventResponse{Ok: true, EventType: string(eventType)}, nil
}

const sessionStatusRunning = "running"

func (s *Server) authenticateRunToken(ctx context.Context) (mcpdomain.SessionContext, error) {
	if s.mcp == nil {
		return mcpdomain.SessionContext{}, status.Error(codes.FailedPrecondition, "mcp service is not configured")
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return mcpdomain.SessionContext{}, status.Error(codes.Unauthenticated, "missing bearer token")
	}

	rawToken := bearerTokenFromMetadata(md)
	if rawToken == "" {
		return mcpdomain.SessionContext{}, status.Error(codes.Unauthenticated, "missing bearer token")
	}

	runSession, err := s.mcp.VerifyRunToken(ctx, rawToken)
	if err != nil {
		return mcpdomain.SessionContext{}, status.Error(codes.Unauthenticated, "invalid bearer token")
	}
	return runSession, nil
}

func bearerTokenFromMetadata(md metadata.MD) string {
	for _, value := range md.Get("authorization") {
		token := agentcallback.ParseBearerToken(value)
		if token != "" {
			return token
		}
	}
	return ""
}

func intPtrFromOptional(value *wrapperspb.Int32Value) *int {
	if value == nil || value.Value <= 0 {
		return nil
	}
	v := int(value.Value)
	return &v
}

func intToOptional(value int32) *wrapperspb.Int32Value {
	if value <= 0 {
		return nil
	}
	return wrapperspb.Int32(value)
}

func optionalTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	value := ts.AsTime().UTC()
	return &value
}

type delete1Fn func(context.Context, staff.Principal, string) error
type delete2Fn func(context.Context, staff.Principal, string, string) error

func (s *Server) delete1(ctx context.Context, principal *controlplanev1.Principal, id string, fn delete1Fn) (*emptypb.Empty, error) {
	p, err := requirePrincipal(principal)
	if err != nil {
		return nil, err
	}
	if err := fn(ctx, p, strings.TrimSpace(id)); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) delete2(ctx context.Context, principal *controlplanev1.Principal, id1 string, id2 string, fn delete2Fn) (*emptypb.Empty, error) {
	p, err := requirePrincipal(principal)
	if err != nil {
		return nil, err
	}
	if err := fn(ctx, p, strings.TrimSpace(id1), strings.TrimSpace(id2)); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func requirePrincipal(p *controlplanev1.Principal) (staff.Principal, error) {
	if p == nil || strings.TrimSpace(p.UserId) == "" {
		return staff.Principal{}, status.Error(codes.Unauthenticated, "not authenticated")
	}
	return staff.Principal{
		UserID:          strings.TrimSpace(p.UserId),
		Email:           strings.TrimSpace(p.Email),
		GitHubLogin:     strings.TrimSpace(p.GithubLogin),
		IsPlatformAdmin: p.IsPlatformAdmin,
		IsPlatformOwner: p.IsPlatformOwner,
	}, nil
}

func userToPrincipal(u userrepo.User) *controlplanev1.Principal {
	return &controlplanev1.Principal{
		UserId:          u.ID,
		Email:           u.Email,
		GithubLogin:     u.GitHubLogin,
		IsPlatformAdmin: u.IsPlatformAdmin,
		IsPlatformOwner: u.IsPlatformOwner,
	}
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}

	var v errs.Validation
	var u errs.Unauthorized
	var f errs.Forbidden
	var c errs.Conflict

	switch {
	case errors.As(err, &v):
		msg := v.Msg
		if v.Field != "" {
			msg = fmt.Sprintf("%s: %s", v.Field, v.Msg)
		}
		return status.Error(codes.InvalidArgument, msg)
	case errors.As(err, &u):
		return status.Error(codes.Unauthenticated, u.Error())
	case errors.As(err, &f):
		return status.Error(codes.PermissionDenied, f.Error())
	case errors.As(err, &c):
		return status.Error(codes.AlreadyExists, c.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func clampLimit(n int32, def int) int {
	if n <= 0 {
		return def
	}
	if n > 1000 {
		return 1000
	}
	return int(n)
}

func tsToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime().UTC()
}

func runToProto(r staffrunrepo.Run) *controlplanev1.Run {
	out := &controlplanev1.Run{
		Id:            r.ID,
		CorrelationId: r.CorrelationID,
		ProjectId:     stringPtrOrNil(r.ProjectID),
		ProjectSlug:   stringPtrOrNil(r.ProjectSlug),
		ProjectName:   stringPtrOrNil(r.ProjectName),
		Status:        r.Status,
		CreatedAt:     timestamppb.New(r.CreatedAt.UTC()),
	}
	if r.StartedAt != nil {
		out.StartedAt = timestamppb.New(r.StartedAt.UTC())
	}
	if r.FinishedAt != nil {
		out.FinishedAt = timestamppb.New(r.FinishedAt.UTC())
	}
	return out
}

func stringPtrOrNil(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func int32PtrOrNil(value int32) *int32 {
	if value <= 0 {
		return nil
	}
	return &value
}

func int64PtrOrNil(value int64) *int64 {
	if value <= 0 {
		return nil
	}
	return &value
}

func parseRuntimeMode(value string) agentdomain.RuntimeMode {
	if strings.EqualFold(strings.TrimSpace(value), string(agentdomain.RuntimeModeFullEnv)) {
		return agentdomain.RuntimeModeFullEnv
	}
	return agentdomain.RuntimeModeCodeOnly
}
