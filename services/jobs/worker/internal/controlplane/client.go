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

// PrepareRunEnvironment asks control-plane to build images and deploy stack for run runtime target.
func (c *Client) PrepareRunEnvironment(ctx context.Context, params workerdomain.PrepareRunEnvironmentParams) (workerdomain.PrepareRunEnvironmentResult, error) {
	resp, err := c.svc.PrepareRunEnvironment(ctx, &controlplanev1.PrepareRunEnvironmentRequest{
		RunId:              strings.TrimSpace(params.RunID),
		RuntimeMode:        strings.TrimSpace(params.RuntimeMode),
		Namespace:          strings.TrimSpace(params.Namespace),
		TargetEnv:          strings.TrimSpace(params.TargetEnv),
		SlotNo:             int32(params.SlotNo),
		RepositoryFullName: strings.TrimSpace(params.RepositoryFullName),
		ServicesYamlPath:   strings.TrimSpace(params.ServicesYAMLPath),
		BuildRef:           strings.TrimSpace(params.BuildRef),
		DeployOnly:         params.DeployOnly,
	})
	if err != nil {
		return workerdomain.PrepareRunEnvironmentResult{}, err
	}
	return workerdomain.PrepareRunEnvironmentResult{
		Namespace: strings.TrimSpace(resp.GetNamespace()),
		TargetEnv: strings.TrimSpace(resp.GetTargetEnv()),
	}, nil
}

// UpsertRunStatusComment updates one run status comment in issue thread.
func (c *Client) UpsertRunStatusComment(ctx context.Context, params workerdomain.RunStatusCommentParams) (workerdomain.RunStatusCommentResult, error) {
	resp, err := c.svc.UpsertRunStatusComment(ctx, &controlplanev1.UpsertRunStatusCommentRequest{
		RunId:           strings.TrimSpace(params.RunID),
		Phase:           strings.TrimSpace(string(params.Phase)),
		JobName:         optionalString(strings.TrimSpace(params.JobName)),
		JobNamespace:    optionalString(strings.TrimSpace(params.JobNamespace)),
		RuntimeMode:     optionalString(strings.TrimSpace(params.RuntimeMode)),
		Namespace:       optionalString(strings.TrimSpace(params.Namespace)),
		TriggerKind:     optionalString(strings.TrimSpace(params.TriggerKind)),
		PromptLocale:    optionalString(strings.TrimSpace(params.PromptLocale)),
		Model:           optionalString(strings.TrimSpace(params.Model)),
		ReasoningEffort: optionalString(strings.TrimSpace(params.ReasoningEffort)),
		RunStatus:       optionalString(strings.TrimSpace(params.RunStatus)),
		Deleted:         params.Deleted,
		AlreadyDeleted:  params.AlreadyDeleted,
	})
	if err != nil {
		return workerdomain.RunStatusCommentResult{}, err
	}
	return workerdomain.RunStatusCommentResult{
		CommentID:  resp.GetCommentId(),
		CommentURL: strings.TrimSpace(resp.GetCommentUrl()),
	}, nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
