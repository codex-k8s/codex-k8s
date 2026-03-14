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
	"google.golang.org/protobuf/types/known/durationpb"
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
		InteractionId:          strings.TrimSpace(params.InteractionID),
		DeliveryId:             strings.TrimSpace(params.DeliveryID),
		AdapterKind:            strings.TrimSpace(params.AdapterKind),
		Status:                 strings.TrimSpace(params.Status),
		RequestEnvelopeJson:    params.RequestEnvelopeJSON,
		AckPayloadJson:         params.AckPayloadJSON,
		AdapterDeliveryId:      optionalString(strings.TrimSpace(params.AdapterDeliveryID)),
		ProviderMessageRefJson: params.ProviderMessageRefJSON,
		EditCapability:         optionalString(strings.TrimSpace(params.EditCapability)),
		Retryable:              params.Retryable,
		LastErrorCode:          optionalString(strings.TrimSpace(params.LastErrorCode)),
		NextRetryAt:            optionalTimestamp(params.NextRetryAt),
		CallbackTokenExpiresAt: optionalTimestamp(params.CallbackTokenExpiresAt),
		FinishedAt:             timestamppb.New(params.FinishedAt.UTC()),
	})
	if err != nil {
		return workerdomain.CompleteInteractionDispatchResult{}, err
	}
	return workerdomain.CompleteInteractionDispatchResult{
		InteractionID:       strings.TrimSpace(resp.GetInteractionId()),
		RunID:               strings.TrimSpace(resp.GetRunId()),
		InteractionState:    strings.TrimSpace(resp.GetInteractionState()),
		ResumeRequired:      resp.GetResumeRequired(),
		ResumeCorrelationID: strings.TrimSpace(resp.GetResumeCorrelationId()),
	}, nil
}

// ExpireNextInteraction processes one due interaction expiry candidate.
func (c *Client) ExpireNextInteraction(ctx context.Context) (workerdomain.ExpireNextInteractionResult, error) {
	resp, err := c.svc.ExpireNextInteraction(ctx, &controlplanev1.ExpireNextInteractionRequest{})
	if err != nil {
		return workerdomain.ExpireNextInteractionResult{}, err
	}
	return workerdomain.ExpireNextInteractionResult{
		Found:               resp.GetFound(),
		InteractionID:       strings.TrimSpace(resp.GetInteractionId()),
		RunID:               strings.TrimSpace(resp.GetRunId()),
		InteractionState:    strings.TrimSpace(resp.GetInteractionState()),
		ResumeRequired:      resp.GetResumeRequired(),
		ResumeCorrelationID: strings.TrimSpace(resp.GetResumeCorrelationId()),
	}, nil
}

// ProcessNextGitHubRateLimitWait claims and processes one due GitHub rate-limit wait.
func (c *Client) ProcessNextGitHubRateLimitWait(ctx context.Context, workerID string) (workerdomain.GitHubRateLimitProcessResult, bool, error) {
	resp, err := c.svc.ProcessNextGitHubRateLimitWait(ctx, &controlplanev1.ProcessNextGitHubRateLimitWaitRequest{
		WorkerId: strings.TrimSpace(workerID),
	})
	if err != nil {
		return workerdomain.GitHubRateLimitProcessResult{}, false, err
	}
	if !resp.GetFound() {
		return workerdomain.GitHubRateLimitProcessResult{}, false, nil
	}

	var resumeNotBefore *time.Time
	if resp.GetResumeNotBefore() != nil {
		value := resp.GetResumeNotBefore().AsTime().UTC()
		resumeNotBefore = &value
	}

	return workerdomain.GitHubRateLimitProcessResult{
		WaitID:                strings.TrimSpace(resp.GetWaitId()),
		RunID:                 strings.TrimSpace(resp.GetRunId()),
		State:                 strings.TrimSpace(resp.GetState()),
		ResolutionKind:        strings.TrimSpace(resp.GetResolutionKind()),
		AttemptNo:             int(resp.GetAttemptNo()),
		ManualActionKind:      strings.TrimSpace(resp.GetManualActionKind()),
		ResumeNotBefore:       resumeNotBefore,
		RequeuedCorrelationID: strings.TrimSpace(resp.GetRequeuedCorrelationId()),
	}, true, nil
}

// ListMissionControlWarmupProjects returns projects that require Mission Control backfill.
func (c *Client) ListMissionControlWarmupProjects(ctx context.Context, limit int) ([]workerdomain.MissionControlWarmupProject, error) {
	resp, err := c.svc.ListMissionControlWarmupProjects(ctx, &controlplanev1.ListMissionControlWarmupProjectsRequest{
		Limit: int32(maxInt64(0, int64(limit))),
	})
	if err != nil {
		return nil, err
	}
	items := make([]workerdomain.MissionControlWarmupProject, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		items = append(items, workerdomain.MissionControlWarmupProject{
			ProjectID:          strings.TrimSpace(item.GetProjectId()),
			ProjectName:        strings.TrimSpace(item.GetProjectName()),
			RepositoryFullName: strings.TrimSpace(item.GetRepositoryFullName()),
		})
	}
	return items, nil
}

// RunMissionControlWarmup executes one Mission Control warmup/backfill cycle.
func (c *Client) RunMissionControlWarmup(ctx context.Context, projectID string, requestedBy string, correlationID string, forceRebuild bool) (workerdomain.MissionControlWarmupResult, error) {
	resp, err := c.svc.RunMissionControlWarmup(ctx, &controlplanev1.RunMissionControlWarmupRequest{
		ProjectId:     strings.TrimSpace(projectID),
		RequestedBy:   strings.TrimSpace(requestedBy),
		CorrelationId: strings.TrimSpace(correlationID),
		ForceRebuild:  forceRebuild,
	})
	if err != nil {
		return workerdomain.MissionControlWarmupResult{}, err
	}
	return workerdomain.MissionControlWarmupResult{
		ProjectID:            strings.TrimSpace(resp.GetProjectId()),
		EntityCount:          resp.GetEntityCount(),
		RelationCount:        resp.GetRelationCount(),
		TimelineEntryCount:   resp.GetTimelineEntryCount(),
		CommandCount:         resp.GetCommandCount(),
		MaxProjectionVersion: resp.GetMaxProjectionVersion(),
		BackfilledEntities:   int(resp.GetBackfilledEntities()),
		BackfilledRelations:  int(resp.GetBackfilledRelations()),
		BackfilledTimelines:  int(resp.GetBackfilledTimelines()),
	}, nil
}

// ClaimMissionControlPendingCommands returns Mission Control commands leased for this worker.
func (c *Client) ClaimMissionControlPendingCommands(ctx context.Context, workerID string, leaseTTL time.Duration, limit int) ([]workerdomain.MissionControlPendingCommand, error) {
	resp, err := c.svc.ClaimMissionControlPendingCommands(ctx, &controlplanev1.ClaimMissionControlPendingCommandsRequest{
		Limit:    int32(maxInt64(0, int64(limit))),
		WorkerId: strings.TrimSpace(workerID),
		LeaseTtl: durationpb.New(leaseTTL),
	})
	if err != nil {
		return nil, err
	}
	items := make([]workerdomain.MissionControlPendingCommand, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		command := workerdomain.MissionControlPendingCommand{
			ProjectID:            strings.TrimSpace(item.GetProjectId()),
			CommandID:            strings.TrimSpace(item.GetCommandId()),
			CommandKind:          strings.TrimSpace(item.GetCommandKind()),
			EffectiveCommandKind: strings.TrimSpace(item.GetEffectiveCommandKind()),
			Status:               strings.TrimSpace(item.GetStatus()),
			CorrelationID:        strings.TrimSpace(item.GetCorrelationId()),
			BusinessIntentKey:    strings.TrimSpace(item.GetBusinessIntentKey()),
			RepositoryFullName:   strings.TrimSpace(item.GetRepositoryFullName()),
			RetryTargetCommandID: strings.TrimSpace(item.GetRetryTargetCommandId()),
			RequestedAt:          tsOrZero(item.GetRequestedAt()),
			UpdatedAt:            tsOrZero(item.GetUpdatedAt()),
		}
		if item.GetStageNextStep() != nil {
			command.StageNextStep = &workerdomain.MissionControlStageNextStepPayload{
				ThreadKind:  strings.TrimSpace(item.GetStageNextStep().GetThreadKind()),
				ThreadNo:    int(item.GetStageNextStep().GetThreadNumber()),
				TargetLabel: strings.TrimSpace(item.GetStageNextStep().GetTargetLabel()),
			}
		}
		items = append(items, command)
	}
	return items, nil
}

// QueueMissionControlCommand marks one Mission Control command as queued.
func (c *Client) QueueMissionControlCommand(ctx context.Context, params workerdomain.MissionControlQueueCommandParams) (workerdomain.MissionControlCommandState, error) {
	resp, err := c.svc.QueueMissionControlCommand(ctx, &controlplanev1.QueueMissionControlCommandRequest{
		ProjectId:     strings.TrimSpace(params.ProjectID),
		CommandId:     strings.TrimSpace(params.CommandID),
		StatusMessage: optionalString(strings.TrimSpace(params.StatusMessage)),
		UpdatedAt:     optionalTimestamp(timePointer(params.UpdatedAt)),
	})
	if err != nil {
		return workerdomain.MissionControlCommandState{}, err
	}
	return missionControlCommandStateFromProto(resp), nil
}

// MarkMissionControlCommandPendingSync marks one Mission Control command as pending_sync.
func (c *Client) MarkMissionControlCommandPendingSync(ctx context.Context, params workerdomain.MissionControlPendingSyncParams) (workerdomain.MissionControlCommandState, error) {
	resp, err := c.svc.MarkMissionControlCommandPendingSync(ctx, &controlplanev1.MarkMissionControlCommandPendingSyncRequest{
		ProjectId:           strings.TrimSpace(params.ProjectID),
		CommandId:           strings.TrimSpace(params.CommandID),
		ProviderDeliveryIds: params.ProviderDeliveryIDs,
		StatusMessage:       optionalString(strings.TrimSpace(params.StatusMessage)),
		UpdatedAt:           optionalTimestamp(timePointer(params.UpdatedAt)),
	})
	if err != nil {
		return workerdomain.MissionControlCommandState{}, err
	}
	return missionControlCommandStateFromProto(resp), nil
}

// MarkMissionControlCommandReconciled marks one Mission Control command as reconciled.
func (c *Client) MarkMissionControlCommandReconciled(ctx context.Context, params workerdomain.MissionControlReconciledParams) (workerdomain.MissionControlCommandState, error) {
	resp, err := c.svc.MarkMissionControlCommandReconciled(ctx, &controlplanev1.MarkMissionControlCommandReconciledRequest{
		ProjectId:           strings.TrimSpace(params.ProjectID),
		CommandId:           strings.TrimSpace(params.CommandID),
		ProviderDeliveryIds: params.ProviderDeliveryIDs,
		StatusMessage:       optionalString(strings.TrimSpace(params.StatusMessage)),
		UpdatedAt:           optionalTimestamp(timePointer(params.UpdatedAt)),
		ReconciledAt:        optionalTimestamp(timePointer(params.ReconciledAt)),
	})
	if err != nil {
		return workerdomain.MissionControlCommandState{}, err
	}
	return missionControlCommandStateFromProto(resp), nil
}

// MarkMissionControlCommandFailed marks one Mission Control command as failed.
func (c *Client) MarkMissionControlCommandFailed(ctx context.Context, params workerdomain.MissionControlFailedParams) (workerdomain.MissionControlCommandState, error) {
	resp, err := c.svc.MarkMissionControlCommandFailed(ctx, &controlplanev1.MarkMissionControlCommandFailedRequest{
		ProjectId:           strings.TrimSpace(params.ProjectID),
		CommandId:           strings.TrimSpace(params.CommandID),
		FailureReason:       strings.TrimSpace(params.FailureReason),
		ProviderDeliveryIds: params.ProviderDeliveryIDs,
		StatusMessage:       optionalString(strings.TrimSpace(params.StatusMessage)),
		UpdatedAt:           optionalTimestamp(timePointer(params.UpdatedAt)),
	})
	if err != nil {
		return workerdomain.MissionControlCommandState{}, err
	}
	return missionControlCommandStateFromProto(resp), nil
}

// ExecuteNextStepAction applies one idempotent label transition on GitHub.
func (c *Client) ExecuteNextStepAction(ctx context.Context, params workerdomain.NextStepExecuteParams) error {
	req := &controlplanev1.NextStepActionRequest{
		Principal: &controlplanev1.Principal{
			UserId:          "system:mission-control-worker",
			IsPlatformAdmin: true,
		},
		RepositoryFullName: strings.TrimSpace(params.RepositoryFullName),
		TargetLabel:        strings.TrimSpace(params.TargetLabel),
	}
	switch strings.TrimSpace(strings.ToLower(params.ThreadKind)) {
	case "pull_request", "pr":
		req.ActionKind = "pull_request_label_add"
		req.PullRequestNumber = optionalInt32(params.ThreadNo)
	default:
		req.ActionKind = "issue_stage_transition"
		req.IssueNumber = optionalInt32(params.ThreadNo)
	}
	_, err := c.svc.ExecuteNextStepAction(ctx, req)
	if err != nil {
		return err
	}
	return nil
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

func optionalInt32(value int) *int32 {
	if value <= 0 {
		return nil
	}
	result := int32(value)
	return &result
}

func timePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	resolved := value.UTC()
	return &resolved
}

func tsOrZero(value *timestamppb.Timestamp) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.AsTime().UTC()
}

func missionControlCommandStateFromProto(resp *controlplanev1.MissionControlCommandState) workerdomain.MissionControlCommandState {
	if resp == nil {
		return workerdomain.MissionControlCommandState{}
	}
	var reconciledAt *time.Time
	if resp.GetReconciledAt() != nil {
		value := resp.GetReconciledAt().AsTime().UTC()
		reconciledAt = &value
	}
	return workerdomain.MissionControlCommandState{
		ProjectID:           strings.TrimSpace(resp.GetProjectId()),
		CommandID:           strings.TrimSpace(resp.GetCommandId()),
		CommandKind:         strings.TrimSpace(resp.GetCommandKind()),
		Status:              strings.TrimSpace(resp.GetStatus()),
		FailureReason:       strings.TrimSpace(resp.GetFailureReason()),
		CorrelationID:       strings.TrimSpace(resp.GetCorrelationId()),
		ProviderDeliveryIDs: resp.GetProviderDeliveryIds(),
		StatusMessage:       strings.TrimSpace(resp.GetStatusMessage()),
		UpdatedAt:           tsOrZero(resp.GetUpdatedAt()),
		ReconciledAt:        reconciledAt,
	}
}

func maxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
