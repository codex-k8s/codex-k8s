package controlplane

import (
	"context"
	"fmt"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Client is a small api-gateway wrapper over the internal control-plane gRPC API.
// It converts gRPC status codes into platform domain errors (errs.*) suitable for HTTP mapping.
type Client struct {
	conn *grpc.ClientConn
	svc  controlplanev1.ControlPlaneServiceClient
}

func Dial(ctx context.Context, target string) (*Client, error) {
	if target == "" {
		return nil, fmt.Errorf("control-plane grpc target is required")
	}

	// grpc.DialContext/WithBlock are deprecated; grpc.NewClient creates a virtual connection,
	// then we optionally wait until it reports Ready (bounded by ctx) to reduce transient 500s on startup.
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial control-plane grpc %q: %w", target, err)
	}

	if err := waitForReady(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("wait for control-plane grpc ready %q: %w", target, err)
	}

	return &Client{
		conn: conn,
		svc:  controlplanev1.NewControlPlaneServiceClient(conn),
	}, nil
}

func waitForReady(ctx context.Context, conn *grpc.ClientConn) error {
	// Start connecting if the channel is idle.
	if conn.GetState() == connectivity.Idle {
		conn.Connect()
	}

	for {
		s := conn.GetState()
		if s == connectivity.Ready {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !conn.WaitForStateChange(ctx, s) {
			return ctx.Err()
		}
	}
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) ResolveStaffByEmail(ctx context.Context, email string, githubLogin string) (*controlplanev1.Principal, error) {
	resp, err := c.svc.ResolveStaffByEmail(ctx, &controlplanev1.ResolveStaffByEmailRequest{
		Email:       email,
		GithubLogin: githubLogin,
	})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp.GetPrincipal(), nil
}

func (c *Client) AuthorizeOAuthUser(ctx context.Context, email string, githubUserID int64, githubLogin string) (*controlplanev1.Principal, error) {
	resp, err := c.svc.AuthorizeOAuthUser(ctx, &controlplanev1.AuthorizeOAuthUserRequest{
		Email:        email,
		GithubUserId: githubUserID,
		GithubLogin:  githubLogin,
	})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp.GetPrincipal(), nil
}

func (c *Client) IngestGitHubWebhook(ctx context.Context, correlationID string, eventType string, deliveryID string, receivedAt time.Time, payloadJSON []byte) (*controlplanev1.IngestGitHubWebhookResponse, error) {
	resp, err := c.svc.IngestGitHubWebhook(ctx, &controlplanev1.IngestGitHubWebhookRequest{
		CorrelationId: correlationID,
		EventType:     eventType,
		DeliveryId:    deliveryID,
		ReceivedAt:    timestamppb.New(receivedAt.UTC()),
		PayloadJson:   payloadJSON,
	})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) ListProjects(ctx context.Context, p *controlplanev1.Principal, limit int32) (*controlplanev1.ListProjectsResponse, error) {
	resp, err := c.svc.ListProjects(ctx, &controlplanev1.ListProjectsRequest{Principal: p, Limit: limit})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) UpsertProject(ctx context.Context, p *controlplanev1.Principal, slug string, name string) (*controlplanev1.Project, error) {
	resp, err := c.svc.UpsertProject(ctx, &controlplanev1.UpsertProjectRequest{Principal: p, Slug: slug, Name: name})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) GetProject(ctx context.Context, p *controlplanev1.Principal, projectID string) (*controlplanev1.Project, error) {
	resp, err := c.svc.GetProject(ctx, &controlplanev1.GetProjectRequest{Principal: p, ProjectId: projectID})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) DeleteProject(ctx context.Context, p *controlplanev1.Principal, projectID string) error {
	_, err := c.svc.DeleteProject(ctx, &controlplanev1.DeleteProjectRequest{Principal: p, ProjectId: projectID})
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (c *Client) ListRuns(ctx context.Context, p *controlplanev1.Principal, limit int32) (*controlplanev1.ListRunsResponse, error) {
	resp, err := c.svc.ListRuns(ctx, &controlplanev1.ListRunsRequest{Principal: p, Limit: limit})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) GetRun(ctx context.Context, p *controlplanev1.Principal, runID string) (*controlplanev1.Run, error) {
	resp, err := c.svc.GetRun(ctx, &controlplanev1.GetRunRequest{Principal: p, RunId: runID})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) ListRunEvents(ctx context.Context, p *controlplanev1.Principal, runID string, limit int32) (*controlplanev1.ListRunEventsResponse, error) {
	resp, err := c.svc.ListRunEvents(ctx, &controlplanev1.ListRunEventsRequest{Principal: p, RunId: runID, Limit: limit})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) ListRunLearningFeedback(ctx context.Context, p *controlplanev1.Principal, runID string, limit int32) (*controlplanev1.ListRunLearningFeedbackResponse, error) {
	resp, err := c.svc.ListRunLearningFeedback(ctx, &controlplanev1.ListRunLearningFeedbackRequest{Principal: p, RunId: runID, Limit: limit})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) ListUsers(ctx context.Context, p *controlplanev1.Principal, limit int32) (*controlplanev1.ListUsersResponse, error) {
	resp, err := c.svc.ListUsers(ctx, &controlplanev1.ListUsersRequest{Principal: p, Limit: limit})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) CreateUser(ctx context.Context, p *controlplanev1.Principal, email string, isPlatformAdmin bool) (*controlplanev1.User, error) {
	resp, err := c.svc.CreateUser(ctx, &controlplanev1.CreateUserRequest{Principal: p, Email: email, IsPlatformAdmin: isPlatformAdmin})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) DeleteUser(ctx context.Context, p *controlplanev1.Principal, userID string) error {
	_, err := c.svc.DeleteUser(ctx, &controlplanev1.DeleteUserRequest{Principal: p, UserId: userID})
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (c *Client) ListProjectMembers(ctx context.Context, p *controlplanev1.Principal, projectID string, limit int32) (*controlplanev1.ListProjectMembersResponse, error) {
	resp, err := c.svc.ListProjectMembers(ctx, &controlplanev1.ListProjectMembersRequest{Principal: p, ProjectId: projectID, Limit: limit})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) UpsertProjectMember(ctx context.Context, p *controlplanev1.Principal, projectID string, userID string, email string, role string) error {
	_, err := c.svc.UpsertProjectMember(ctx, &controlplanev1.UpsertProjectMemberRequest{
		Principal: p,
		ProjectId: projectID,
		UserId:    userID,
		Email:     email,
		Role:      role,
	})
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (c *Client) DeleteProjectMember(ctx context.Context, p *controlplanev1.Principal, projectID string, userID string) error {
	_, err := c.svc.DeleteProjectMember(ctx, &controlplanev1.DeleteProjectMemberRequest{Principal: p, ProjectId: projectID, UserId: userID})
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (c *Client) SetProjectMemberLearningModeOverride(ctx context.Context, p *controlplanev1.Principal, projectID string, userID string, enabled *bool) error {
	var w *wrapperspb.BoolValue
	if enabled != nil {
		w = wrapperspb.Bool(*enabled)
	}
	_, err := c.svc.SetProjectMemberLearningModeOverride(ctx, &controlplanev1.SetProjectMemberLearningModeOverrideRequest{
		Principal: p,
		ProjectId: projectID,
		UserId:    userID,
		Enabled:   w,
	})
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (c *Client) ListProjectRepositories(ctx context.Context, p *controlplanev1.Principal, projectID string, limit int32) (*controlplanev1.ListProjectRepositoriesResponse, error) {
	resp, err := c.svc.ListProjectRepositories(ctx, &controlplanev1.ListProjectRepositoriesRequest{Principal: p, ProjectId: projectID, Limit: limit})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) UpsertProjectRepository(ctx context.Context, p *controlplanev1.Principal, projectID string, provider string, owner string, name string, token string, servicesYAMLPath string) (*controlplanev1.RepositoryBinding, error) {
	resp, err := c.svc.UpsertProjectRepository(ctx, &controlplanev1.UpsertProjectRepositoryRequest{
		Principal:        p,
		ProjectId:        projectID,
		Provider:         provider,
		Owner:            owner,
		Name:             name,
		Token:            token,
		ServicesYamlPath: servicesYAMLPath,
	})
	if err != nil {
		return nil, toDomainError(err)
	}
	return resp, nil
}

func (c *Client) DeleteProjectRepository(ctx context.Context, p *controlplanev1.Principal, projectID string, repositoryID string) error {
	_, err := c.svc.DeleteProjectRepository(ctx, &controlplanev1.DeleteProjectRepositoryRequest{Principal: p, ProjectId: projectID, RepositoryId: repositoryID})
	if err != nil {
		return toDomainError(err)
	}
	return nil
}

func (c *Client) Empty() *emptypb.Empty { return &emptypb.Empty{} }

func toDomainError(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case codes.InvalidArgument:
		return errs.Validation{Msg: st.Message()}
	case codes.NotFound:
		return errs.Validation{Msg: st.Message()}
	case codes.Unauthenticated:
		return errs.Unauthorized{Msg: st.Message()}
	case codes.PermissionDenied:
		return errs.Forbidden{Msg: st.Message()}
	case codes.AlreadyExists:
		return errs.Conflict{Msg: st.Message()}
	default:
		return err
	}
}
