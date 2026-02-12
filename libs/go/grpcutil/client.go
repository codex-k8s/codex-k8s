package grpcutil

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// DialInsecureReady creates insecure gRPC client connection and waits until channel is ready.
func DialInsecureReady(ctx context.Context, target string) (*grpc.ClientConn, error) {
	if target == "" {
		return nil, fmt.Errorf("grpc target is required")
	}

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial grpc %q: %w", target, err)
	}

	if err := WaitForReady(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("wait for grpc ready %q: %w", target, err)
	}

	return conn, nil
}

// WaitForReady blocks until client connection reaches Ready or ctx is done.
func WaitForReady(ctx context.Context, conn *grpc.ClientConn) error {
	if conn.GetState() == connectivity.Idle {
		conn.Connect()
	}

	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !conn.WaitForStateChange(ctx, state) {
			return ctx.Err()
		}
	}
}
