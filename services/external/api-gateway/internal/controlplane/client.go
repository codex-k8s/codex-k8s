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
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (c *Client) Service() controlplanev1.ControlPlaneServiceClient {
	if c == nil {
		return nil
	}
	return c.svc
}

func (c *Client) ResolveStaffByEmail(ctx context.Context, email string, githubLogin string) (*controlplanev1.Principal, error) {
	resp, err := c.svc.ResolveStaffByEmail(ctx, &controlplanev1.ResolveStaffByEmailRequest{
		Email:       email,
		GithubLogin: githubLogin,
	})
	if err != nil {
		return nil, ToDomainError(err)
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
		return nil, ToDomainError(err)
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
		return nil, ToDomainError(err)
	}
	return resp, nil
}

func ToDomainError(err error) error {
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
