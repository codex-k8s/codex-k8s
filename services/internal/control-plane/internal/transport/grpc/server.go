package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	staffrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/staffrun"
	userrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/user"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/staff"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/webhook"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type webhookIngress interface {
	IngestGitHubWebhook(ctx context.Context, cmd webhook.IngestCommand) (webhook.IngestResult, error)
}

// Dependencies wires domain services and repositories into the gRPC transport.
type Dependencies struct {
	Webhook webhookIngress
	Staff   *staff.Service
	Users   userrepo.Repository
	Logger  *slog.Logger
}

// Server implements ControlPlaneServiceServer.
type Server struct {
	controlplanev1.UnimplementedControlPlaneServiceServer

	webhook webhookIngress
	staff   *staff.Service
	users   userrepo.Repository
	logger  *slog.Logger
}

func NewServer(deps Dependencies) *Server {
	return &Server{
		webhook: deps.Webhook,
		staff:   deps.Staff,
		users:   deps.Users,
		logger:  deps.Logger,
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
		Status:        res.Status,
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

	login := strings.TrimSpace(req.GithubLogin)
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
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, &controlplanev1.Project{
			Id:   asString(m["id"]),
			Slug: asString(m["slug"]),
			Name: asString(m["name"]),
			Role: asString(m["role"]),
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
			RepositoryId: f.RepositoryID,
			PrNumber:     int32(f.PRNumber),
			FilePath:     f.FilePath,
			Line:         int32(f.Line),
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
			GithubUserId:    u.GitHubUserID,
			GithubLogin:     u.GitHubLogin,
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
		GithubUserId:    u.GitHubUserID,
		GithubLogin:     u.GitHubLogin,
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

	projectID := strings.TrimSpace(req.ProjectId)
	userID := strings.TrimSpace(req.UserId)
	email := strings.TrimSpace(req.Email)
	role := strings.TrimSpace(req.Role)

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
		ProjectId:     r.ProjectID,
		ProjectSlug:   r.ProjectSlug,
		ProjectName:   r.ProjectName,
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

func asString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprintf("%v", v)
	}
}
