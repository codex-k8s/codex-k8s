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
	"google.golang.org/protobuf/types/known/timestamppb"
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

// EvaluateRuntimeReuse asks control-plane whether one reusable namespace can skip runtime deploy/build.
func (c *Client) EvaluateRuntimeReuse(ctx context.Context, params workerdomain.EvaluateRuntimeReuseParams) (workerdomain.EvaluateRuntimeReuseResult, error) {
	resp, err := c.svc.EvaluateRuntimeReuse(ctx, &controlplanev1.EvaluateRuntimeReuseRequest{
		RunId:              strings.TrimSpace(params.RunID),
		ProjectId:          strings.TrimSpace(params.ProjectID),
		IssueNumber:        params.IssueNumber,
		AgentKey:           strings.TrimSpace(params.AgentKey),
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
		return workerdomain.EvaluateRuntimeReuseResult{}, err
	}
	return workerdomain.EvaluateRuntimeReuseResult{
		Reusable:          resp.GetReusable(),
		Namespace:         strings.TrimSpace(resp.GetNamespace()),
		TargetEnv:         strings.TrimSpace(resp.GetTargetEnv()),
		EffectiveBuildRef: strings.TrimSpace(resp.GetEffectiveBuildRef()),
		FingerprintHash:   strings.TrimSpace(resp.GetFingerprintHash()),
		Reason:            strings.TrimSpace(resp.GetReason()),
	}, nil
}

// ClaimNextInteractionDispatch reserves one due interaction delivery attempt.
func (c *Client) ClaimNextInteractionDispatch(ctx context.Context, pendingAttemptTimeout time.Duration) (workerdomain.InteractionDispatchClaim, bool, error) {
	resp, err := c.svc.ClaimNextInteractionDispatch(ctx, &controlplanev1.ClaimNextInteractionDispatchRequest{
		PendingAttemptTimeoutSeconds: int32(maxInt64(0, int64(pendingAttemptTimeout.Seconds()))),
	})
	if err != nil {
		return workerdomain.InteractionDispatchClaim{}, false, err
	}
	if !resp.GetFound() {
		return workerdomain.InteractionDispatchClaim{}, false, nil
	}

	var responseDeadlineAt *time.Time
	if resp.GetResponseDeadlineAt() != nil {
		value := resp.GetResponseDeadlineAt().AsTime().UTC()
		responseDeadlineAt = &value
	}

	return workerdomain.InteractionDispatchClaim{
		CorrelationID:      strings.TrimSpace(resp.GetCorrelationId()),
		InteractionID:      strings.TrimSpace(resp.GetInteractionId()),
		RunID:              strings.TrimSpace(resp.GetRunId()),
		InteractionKind:    strings.TrimSpace(resp.GetInteractionKind()),
		RecipientProvider:  strings.TrimSpace(resp.GetRecipientProvider()),
		RecipientRef:       strings.TrimSpace(resp.GetRecipientRef()),
		ResponseDeadlineAt: responseDeadlineAt,
		Attempt: workerdomain.InteractionDispatchAttempt{
			ID:          resp.GetAttemptId(),
			AttemptNo:   int(resp.GetAttemptNo()),
			DeliveryID:  strings.TrimSpace(resp.GetDeliveryId()),
			AdapterKind: strings.TrimSpace(resp.GetAdapterKind()),
		},
		RequestEnvelopeJSON: resp.GetRequestEnvelopeJson(),
	}, true, nil
}

// CompleteInteractionDispatch persists one dispatch outcome.
func (c *Client) CompleteInteractionDispatch(ctx context.Context, params workerdomain.CompleteInteractionDispatchParams) (workerdomain.CompleteInteractionDispatchResult, error) {
	resp, err := c.svc.CompleteInteractionDispatch(ctx, &controlplanev1.CompleteInteractionDispatchRequest{
		InteractionId:       strings.TrimSpace(params.InteractionID),
		DeliveryId:          strings.TrimSpace(params.DeliveryID),
		AdapterKind:         strings.TrimSpace(params.AdapterKind),
		Status:              strings.TrimSpace(params.Status),
		RequestEnvelopeJson: params.RequestEnvelopeJSON,
		AckPayloadJson:      params.AckPayloadJSON,
		AdapterDeliveryId:   optionalString(strings.TrimSpace(params.AdapterDeliveryID)),
		Retryable:           params.Retryable,
		LastErrorCode:       optionalString(strings.TrimSpace(params.LastErrorCode)),
		NextRetryAt:         optionalTimestamp(params.NextRetryAt),
		FinishedAt:          timestamppb.New(params.FinishedAt.UTC()),
	})
	if err != nil {
		return workerdomain.CompleteInteractionDispatchResult{}, err
	}
	return workerdomain.CompleteInteractionDispatchResult{
		InteractionID:    strings.TrimSpace(resp.GetInteractionId()),
		InteractionState: strings.TrimSpace(resp.GetInteractionState()),
		ResumeRequired:   resp.GetResumeRequired(),
	}, nil
}

// ExpireNextInteraction processes one due interaction expiry candidate.
func (c *Client) ExpireNextInteraction(ctx context.Context) (workerdomain.ExpireNextInteractionResult, error) {
	resp, err := c.svc.ExpireNextInteraction(ctx, &controlplanev1.ExpireNextInteractionRequest{})
	if err != nil {
		return workerdomain.ExpireNextInteractionResult{}, err
	}
	return workerdomain.ExpireNextInteractionResult{
		Found:            resp.GetFound(),
		InteractionID:    strings.TrimSpace(resp.GetInteractionId()),
		InteractionState: strings.TrimSpace(resp.GetInteractionState()),
		ResumeRequired:   resp.GetResumeRequired(),
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

func optionalTimestamp(value *time.Time) *timestamppb.Timestamp {
	if value == nil || value.IsZero() {
		return nil
	}
	return timestamppb.New(value.UTC())
}

func maxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
