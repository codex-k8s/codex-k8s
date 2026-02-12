package controlplane

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/grpcutil"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	workerdomain "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/worker"
	"google.golang.org/grpc"
)

// Client is a worker-side wrapper over control-plane gRPC.
type Client struct {
	conn *grpc.ClientConn
	svc  controlplanev1.ControlPlaneServiceClient
}

// Dial creates control-plane gRPC client.
func Dial(ctx context.Context, target string) (*Client, error) {
	conn, err := grpcutil.DialInsecureReady(ctx, strings.TrimSpace(target))
	if err != nil {
		return nil, fmt.Errorf("dial control-plane grpc: %w", err)
	}
	return &Client{
		conn: conn,
		svc:  controlplanev1.NewControlPlaneServiceClient(conn),
	}, nil
}

// Close closes underlying gRPC connection.
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// IssueRunMCPToken requests short-lived run-bound MCP token from control-plane.
func (c *Client) IssueRunMCPToken(ctx context.Context, params workerdomain.IssueMCPTokenParams) (workerdomain.IssuedMCPToken, error) {
	resp, err := c.svc.IssueRunMCPToken(ctx, &controlplanev1.IssueRunMCPTokenRequest{
		RunId:       strings.TrimSpace(params.RunID),
		Namespace:   strings.TrimSpace(params.Namespace),
		RuntimeMode: strings.TrimSpace(string(params.RuntimeMode)),
	})
	if err != nil {
		return workerdomain.IssuedMCPToken{}, err
	}

	token := strings.TrimSpace(resp.GetToken())
	if token == "" {
		return workerdomain.IssuedMCPToken{}, fmt.Errorf("control-plane returned empty mcp token")
	}
	expiresAt := time.Time{}
	if resp.GetExpiresAt() != nil {
		expiresAt = resp.GetExpiresAt().AsTime().UTC()
	}

	return workerdomain.IssuedMCPToken{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
