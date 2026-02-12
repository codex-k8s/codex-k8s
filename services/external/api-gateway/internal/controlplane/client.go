package controlplane

import (
	"context"
	"fmt"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/cast"
	"github.com/codex-k8s/codex-k8s/libs/go/grpcutil"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Client is a small api-gateway wrapper over the internal control-plane gRPC API.
type Client struct {
	conn *grpc.ClientConn
	svc  controlplanev1.ControlPlaneServiceClient
}

func Dial(ctx context.Context, target string) (*Client, error) {
	if target == "" {
		return nil, fmt.Errorf("control-plane grpc target is required")
	}

	conn, err := grpcutil.DialInsecureReady(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("dial control-plane grpc: %w", err)
	}

	return &Client{
		conn: conn,
		svc:  controlplanev1.NewControlPlaneServiceClient(conn),
	}, nil
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
		GithubLogin: cast.TrimmedStringPtr(githubLogin),
	})
	if err != nil {
		return nil, err
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
		return nil, err
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
		return nil, err
	}
	return resp, nil
}
