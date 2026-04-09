package controlplane

import (
	"context"
	"fmt"
	"strings"

	"github.com/codex-k8s/kodex/libs/go/grpcutil"
	controlplanev1 "github.com/codex-k8s/kodex/proto/gen/go/kodex/controlplane/v1"
	"google.golang.org/grpc"
)

// Client wraps control-plane gRPC calls used by telegram-interaction-adapter.
type Client struct {
	conn *grpc.ClientConn
	svc  controlplanev1.ControlPlaneServiceClient
}

// Dial creates a ready control-plane gRPC client.
func Dial(ctx context.Context, target string) (*Client, error) {
	target = strings.TrimSpace(target)
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

// Close closes the underlying gRPC connection.
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// SubmitAdapterInteractionCallback forwards one normalized adapter callback to control-plane.
func (c *Client) SubmitAdapterInteractionCallback(
	ctx context.Context,
	req *controlplanev1.SubmitInteractionCallbackRequest,
) (*controlplanev1.SubmitInteractionCallbackResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("submit adapter interaction callback request is required")
	}
	return c.svc.SubmitAdapterInteractionCallback(ctx, req)
}
