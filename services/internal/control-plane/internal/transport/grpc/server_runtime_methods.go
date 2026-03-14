package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	agentcallbackdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/agentcallback"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	agentsessionrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
	runstatusdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runstatus"
	runtimedeploydomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runtimedeploy"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/staff"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	agentcallback "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/transport/agentcallback"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) IssueRunMCPToken(ctx context.Context, req *controlplanev1.IssueRunMCPTokenRequest) (*controlplanev1.IssueRunMCPTokenResponse, error) {
	if s.mcp == nil {
		return nil, status.Error(codes.FailedPrecondition, "mcp service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		return nil, status.Error(codes.InvalidArgument, "run_id is required")
	}
	ttl := time.Duration(req.GetTtlSeconds()) * time.Second

	issuedToken, err := s.mcp.IssueRunToken(ctx, mcpdomain.IssueRunTokenParams{
		RunID:       runID,
		Namespace:   strings.TrimSpace(req.GetNamespace()),
		RuntimeMode: parseRuntimeMode(req.GetRuntimeMode()),
		TTL:         ttl,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.IssueRunMCPTokenResponse{
		Token:     issuedToken.Token,
		ExpiresAt: timestamppb.New(issuedToken.ExpiresAt.UTC()),
	}, nil
}

func (s *Server) PrepareRunEnvironment(ctx context.Context, req *controlplanev1.PrepareRunEnvironmentRequest) (*controlplanev1.PrepareRunEnvironmentResponse, error) {
	if s.runtimeDeploy == nil {
		return nil, status.Error(codes.FailedPrecondition, "runtime deploy service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		return nil, status.Error(codes.InvalidArgument, "run_id is required")
	}

	prepared, err := s.runtimeDeploy.PrepareRunEnvironment(ctx, runtimedeploydomain.PrepareParams{
		RunID:              runID,
		RuntimeMode:        strings.TrimSpace(req.GetRuntimeMode()),
		Namespace:          strings.TrimSpace(req.GetNamespace()),
		TargetEnv:          strings.TrimSpace(req.GetTargetEnv()),
		SlotNo:             int(req.GetSlotNo()),
		RepositoryFullName: strings.TrimSpace(req.GetRepositoryFullName()),
		ServicesYAMLPath:   strings.TrimSpace(req.GetServicesYamlPath()),
		BuildRef:           strings.TrimSpace(req.GetBuildRef()),
		DeployOnly:         req.GetDeployOnly(),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.PrepareRunEnvironmentResponse{
		Ok:        true,
		RunId:     runID,
		Namespace: strings.TrimSpace(prepared.Namespace),
		TargetEnv: strings.TrimSpace(prepared.TargetEnv),
	}, nil
}

func (s *Server) EvaluateRuntimeReuse(ctx context.Context, req *controlplanev1.EvaluateRuntimeReuseRequest) (*controlplanev1.EvaluateRuntimeReuseResponse, error) {
	if s.runtimeDeploy == nil {
		return nil, status.Error(codes.FailedPrecondition, "runtime deploy service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		return nil, status.Error(codes.InvalidArgument, "run_id is required")
	}

	evaluated, err := s.runtimeDeploy.EvaluateRuntimeReuse(ctx, runtimedeploydomain.EvaluateReuseParams{
		RunID:              runID,
		ProjectID:          strings.TrimSpace(req.GetProjectId()),
		IssueNumber:        req.GetIssueNumber(),
		AgentKey:           strings.TrimSpace(req.GetAgentKey()),
		RuntimeMode:        strings.TrimSpace(req.GetRuntimeMode()),
		Namespace:          strings.TrimSpace(req.GetNamespace()),
		TargetEnv:          strings.TrimSpace(req.GetTargetEnv()),
		SlotNo:             int(req.GetSlotNo()),
		RepositoryFullName: strings.TrimSpace(req.GetRepositoryFullName()),
		ServicesYAMLPath:   strings.TrimSpace(req.GetServicesYamlPath()),
		BuildRef:           strings.TrimSpace(req.GetBuildRef()),
		DeployOnly:         req.GetDeployOnly(),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.EvaluateRuntimeReuseResponse{
		Reusable:          evaluated.Reusable,
		Namespace:         strings.TrimSpace(evaluated.Namespace),
		TargetEnv:         strings.TrimSpace(evaluated.TargetEnv),
		EffectiveBuildRef: strings.TrimSpace(evaluated.EffectiveBuildRef),
		FingerprintHash:   strings.TrimSpace(evaluated.FingerprintHash),
		Reason:            strings.TrimSpace(evaluated.Reason),
	}, nil
}

func (s *Server) ListRuntimeDeployTasks(ctx context.Context, req *controlplanev1.ListRuntimeDeployTasksRequest) (*controlplanev1.ListRuntimeDeployTasksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	page := clampPage(req.GetPage())
	pageSize := clampLimit(req.GetPageSize(), 20)
	items, totalCount, err := s.staff.ListRuntimeDeployTasks(
		ctx,
		p,
		page,
		pageSize,
		optionalProtoString(req.Status),
		optionalProtoString(req.TargetEnv),
	)
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.RuntimeDeployTask, 0, len(items))
	for _, item := range items {
		out = append(out, runtimeDeployTaskToProto(item))
	}
	return &controlplanev1.ListRuntimeDeployTasksResponse{
		Items:      out,
		TotalCount: int32(totalCount),
		Page:       int32(page),
		PageSize:   int32(pageSize),
	}, nil
}

func (s *Server) GetRuntimeDeployTask(ctx context.Context, req *controlplanev1.GetRuntimeDeployTaskRequest) (*controlplanev1.RuntimeDeployTask, error) {
	return requestStaffEntity(ctx, req, req.GetRunId(), s.staff.GetRuntimeDeployTask, runtimeDeployTaskToProto)
}

type runtimeDeployTaskActionRequest interface {
	GetPrincipal() *controlplanev1.Principal
	GetRunId() string
	GetReason() string
}

func (s *Server) requestRuntimeDeployTaskAction(
	ctx context.Context,
	req runtimeDeployTaskActionRequest,
	action runtimedeploydomain.TaskAction,
) (*controlplanev1.RuntimeDeployTaskActionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	result, err := s.staff.RequestRuntimeDeployTaskAction(
		ctx,
		p,
		strings.TrimSpace(req.GetRunId()),
		action,
		strings.TrimSpace(req.GetReason()),
	)
	if err != nil {
		return nil, toStatus(err)
	}
	return runtimeDeployTaskActionToProto(result), nil
}

func (s *Server) CancelRuntimeDeployTask(ctx context.Context, req *controlplanev1.CancelRuntimeDeployTaskRequest) (*controlplanev1.RuntimeDeployTaskActionResponse, error) {
	return s.requestRuntimeDeployTaskAction(ctx, req, runtimedeploydomain.TaskActionCancel)
}

func (s *Server) StopRuntimeDeployTask(ctx context.Context, req *controlplanev1.StopRuntimeDeployTaskRequest) (*controlplanev1.RuntimeDeployTaskActionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if !req.GetForce() {
		return nil, status.Error(codes.InvalidArgument, "force must be true for stop")
	}
	return s.requestRuntimeDeployTaskAction(ctx, req, runtimedeploydomain.TaskActionStop)
}

// TODO(codex-k8s#81): This RPC is temporarily unused after staff UI removed
// "platform error" alerts. Decide later whether to keep, repurpose, or remove it.
func (s *Server) ListRuntimeErrors(ctx context.Context, req *controlplanev1.ListRuntimeErrorsRequest) (*controlplanev1.ListRuntimeErrorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	items, err := s.staff.ListRuntimeErrors(ctx, p, querytypes.RuntimeErrorListFilter{
		Limit:         clampLimit(req.GetLimit(), 100),
		State:         parseRuntimeErrorListState(optionalProtoString(req.State)),
		Level:         optionalProtoString(req.Level),
		Source:        optionalProtoString(req.Source),
		RunID:         optionalProtoString(req.RunId),
		CorrelationID: optionalProtoString(req.CorrelationId),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.RuntimeError, 0, len(items))
	for _, item := range items {
		out = append(out, runtimeErrorToProto(item))
	}
	return &controlplanev1.ListRuntimeErrorsResponse{Items: out}, nil
}

// TODO(codex-k8s#81): This RPC is temporarily unused after staff UI removed
// "platform error" alerts. Decide later whether to keep, repurpose, or remove it.
func (s *Server) MarkRuntimeErrorViewed(ctx context.Context, req *controlplanev1.MarkRuntimeErrorViewedRequest) (*controlplanev1.RuntimeError, error) {
	return requestStaffEntity(ctx, req, req.GetRuntimeErrorId(), s.staff.MarkRuntimeErrorViewed, runtimeErrorToProto)
}

func mapStaffEntity[Item any, Out any](fetch func() (Item, error), cast func(Item) Out) (Out, error) {
	item, err := fetch()
	if err != nil {
		var zero Out
		return zero, toStatus(err)
	}
	return cast(item), nil
}

func requestStaffEntity[Req principalRequest, Item any, Out any](
	ctx context.Context,
	req Req,
	rawID string,
	fetch func(context.Context, staff.Principal, string) (Item, error),
	cast func(Item) Out,
) (Out, error) {
	return withRequestPrincipal(req, func(principal staff.Principal) (Out, error) {
		return mapStaffEntity(
			func() (Item, error) {
				return fetch(ctx, principal, strings.TrimSpace(rawID))
			},
			cast,
		)
	})
}

func (s *Server) UpsertAgentSession(ctx context.Context, req *controlplanev1.UpsertAgentSessionRequest) (*controlplanev1.UpsertAgentSessionResponse, error) {
	if s.agentCallbacks == nil {
		return nil, status.Error(codes.FailedPrecondition, "agent callback service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	runSession, err := s.authenticateRunToken(ctx)
	if err != nil {
		return nil, err
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		runID = runSession.RunID
	}
	if runID == "" {
		return nil, status.Error(codes.InvalidArgument, "run_id is required")
	}
	if runID != runSession.RunID {
		return nil, status.Error(codes.PermissionDenied, "run_id mismatch with token")
	}

	repositoryFullName := strings.TrimSpace(req.GetRepositoryFullName())
	if repositoryFullName == "" {
		return nil, status.Error(codes.InvalidArgument, "repository_full_name is required")
	}
	branchName := strings.TrimSpace(req.GetBranchName())
	if branchName == "" {
		return nil, status.Error(codes.InvalidArgument, "branch_name is required")
	}
	agentKey := strings.TrimSpace(req.GetAgentKey())
	if agentKey == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_key is required")
	}
	if req.GetSnapshotVersion() < 0 {
		return nil, status.Error(codes.InvalidArgument, "snapshot_version must be >= 0")
	}

	correlationID := strings.TrimSpace(req.GetCorrelationId())
	if correlationID == "" {
		correlationID = runSession.CorrelationID
	}
	if correlationID == "" {
		return nil, status.Error(codes.InvalidArgument, "correlation_id is required")
	}

	projectID := strings.TrimSpace(req.GetProjectId())
	if projectID == "" {
		projectID = runSession.ProjectID
	}

	statusValue := strings.TrimSpace(req.GetStatus())
	if statusValue == "" {
		statusValue = sessionStatusRunning
	}

	startedAt := time.Now().UTC()
	if req.GetStartedAt() != nil {
		startedAt = req.GetStartedAt().AsTime().UTC()
	}

	result, err := s.agentCallbacks.UpsertAgentSession(ctx, agentcallbackdomain.UpsertAgentSessionParams{
		RunID:                   runID,
		CorrelationID:           correlationID,
		ProjectID:               projectID,
		RepositoryFullName:      repositoryFullName,
		AgentKey:                agentKey,
		IssueNumber:             intPtrFromOptional(req.GetIssueNumber()),
		BranchName:              branchName,
		PRNumber:                intPtrFromOptional(req.GetPrNumber()),
		PRURL:                   strings.TrimSpace(req.GetPrUrl()),
		TriggerKind:             strings.TrimSpace(req.GetTriggerKind()),
		TemplateKind:            strings.TrimSpace(req.GetTemplateKind()),
		TemplateSource:          strings.TrimSpace(req.GetTemplateSource()),
		TemplateLocale:          strings.TrimSpace(req.GetTemplateLocale()),
		Model:                   strings.TrimSpace(req.GetModel()),
		ReasoningEffort:         strings.TrimSpace(req.GetReasoningEffort()),
		Status:                  statusValue,
		SessionID:               strings.TrimSpace(req.GetSessionId()),
		SessionJSON:             json.RawMessage(req.GetSessionJson()),
		CodexSessionPath:        strings.TrimSpace(req.GetCodexCliSessionPath()),
		CodexSessionJSON:        json.RawMessage(req.GetCodexCliSessionJson()),
		ExpectedSnapshotVersion: req.GetSnapshotVersion(),
		SnapshotChecksum:        strings.TrimSpace(req.GetSnapshotChecksum()),
		StartedAt:               startedAt,
		FinishedAt:              optionalTime(req.GetFinishedAt()),
	})
	if err != nil {
		var conflict agentsessionrepo.SnapshotVersionConflict
		if errors.As(err, &conflict) {
			return nil, agentSessionSnapshotVersionConflictStatus(conflict)
		}
		return nil, status.Error(codes.Internal, "failed to persist agent session")
	}

	return &controlplanev1.UpsertAgentSessionResponse{
		Ok:               true,
		RunId:            runID,
		SnapshotVersion:  result.SnapshotVersion,
		SnapshotChecksum: stringPtrOrNil(result.SnapshotChecksum),
	}, nil
}

func (s *Server) GetLatestAgentSession(ctx context.Context, req *controlplanev1.GetLatestAgentSessionRequest) (*controlplanev1.GetLatestAgentSessionResponse, error) {
	if s.agentCallbacks == nil {
		return nil, status.Error(codes.FailedPrecondition, "agent callback service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if _, err := s.authenticateRunToken(ctx); err != nil {
		return nil, err
	}

	repositoryFullName := strings.TrimSpace(req.GetRepositoryFullName())
	branchName := strings.TrimSpace(req.GetBranchName())
	agentKey := strings.TrimSpace(req.GetAgentKey())
	if repositoryFullName == "" || branchName == "" || agentKey == "" {
		return nil, status.Error(codes.InvalidArgument, "repository_full_name, branch_name and agent_key are required")
	}

	item, found, err := s.agentCallbacks.GetLatestAgentSession(ctx, agentcallbackdomain.GetLatestAgentSessionQuery{
		RepositoryFullName: repositoryFullName,
		BranchName:         branchName,
		AgentKey:           agentKey,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to load latest agent session")
	}
	if !found {
		return &controlplanev1.GetLatestAgentSessionResponse{Found: false}, nil
	}

	snapshot := &controlplanev1.AgentSessionSnapshot{
		RunId:               item.RunID,
		CorrelationId:       item.CorrelationID,
		ProjectId:           stringPtrOrNil(item.ProjectID),
		RepositoryFullName:  item.RepositoryFullName,
		AgentKey:            item.AgentKey,
		IssueNumber:         intToOptional(int32(item.IssueNumber)),
		BranchName:          item.BranchName,
		PrNumber:            intToOptional(int32(item.PRNumber)),
		PrUrl:               stringPtrOrNil(item.PRURL),
		TriggerKind:         stringPtrOrNil(item.TriggerKind),
		TemplateKind:        stringPtrOrNil(item.TemplateKind),
		TemplateSource:      stringPtrOrNil(item.TemplateSource),
		TemplateLocale:      stringPtrOrNil(item.TemplateLocale),
		Model:               stringPtrOrNil(item.Model),
		ReasoningEffort:     stringPtrOrNil(item.ReasoningEffort),
		Status:              stringPtrOrNil(item.Status),
		SessionId:           stringPtrOrNil(item.SessionID),
		SessionJson:         bytesOrNil(item.SessionJSON),
		CodexCliSessionPath: stringPtrOrNil(item.CodexSessionPath),
		CodexCliSessionJson: bytesOrNil(item.CodexSessionJSON),
		SnapshotVersion:     item.SnapshotVersion,
		SnapshotChecksum:    stringPtrOrNil(item.SnapshotChecksum),
		StartedAt:           timestamppb.New(item.StartedAt.UTC()),
		CreatedAt:           timestamppb.New(item.CreatedAt.UTC()),
		UpdatedAt:           timestamppb.New(item.UpdatedAt.UTC()),
	}
	if !item.FinishedAt.IsZero() {
		snapshot.FinishedAt = timestamppb.New(item.FinishedAt.UTC())
	}
	if !item.SnapshotUpdatedAt.IsZero() {
		snapshot.SnapshotUpdatedAt = timestamppb.New(item.SnapshotUpdatedAt.UTC())
	}

	return &controlplanev1.GetLatestAgentSessionResponse{
		Found:   true,
		Session: snapshot,
	}, nil
}

func (s *Server) GetRunInteractionResumePayload(ctx context.Context, req *controlplanev1.GetRunInteractionResumePayloadRequest) (*controlplanev1.GetRunInteractionResumePayloadResponse, error) {
	return executeRunScopedPayloadSpec(ctx, req, s.loadRunScopedPayload, interactionResumePayloadSpec(s))
}

func buildRunScopedPayloadResponse[T any](payload json.RawMessage, found bool, err error, build func(bool, json.RawMessage) T) (T, error) {
	if err != nil {
		var zero T
		return zero, err
	}
	return build(found, payload), nil
}

type runScopedPayloadSpec[T any] struct {
	label string
	load  func(context.Context, string) (json.RawMessage, bool, error)
	build func(bool, json.RawMessage) T
}

func executeRunScopedPayloadSpec[T any](
	ctx context.Context,
	req any,
	loadRunScoped func(context.Context, any, string, func(context.Context, string) (json.RawMessage, bool, error)) (json.RawMessage, bool, error),
	spec runScopedPayloadSpec[T],
) (T, error) {
	payload, found, err := loadRunScoped(ctx, req, spec.label, spec.load)
	return buildRunScopedPayloadResponse(payload, found, err, spec.build)
}

func interactionResumePayloadSpec(s *Server) runScopedPayloadSpec[*controlplanev1.GetRunInteractionResumePayloadResponse] {
	return runScopedPayloadSpec[*controlplanev1.GetRunInteractionResumePayloadResponse]{
		label: "interaction resume payload",
		load:  s.agentCallbacks.GetRunInteractionResumePayload,
		build: interactionResumePayloadResponseBuilder,
	}
}

var interactionResumePayloadResponseBuilder = func(found bool, payload json.RawMessage) *controlplanev1.GetRunInteractionResumePayloadResponse {
	return &controlplanev1.GetRunInteractionResumePayloadResponse{Found: found, PayloadJson: payload}
}

func (s *Server) loadRunScopedPayload(
	ctx context.Context,
	req any,
	payloadLabel string,
	load func(context.Context, string) (json.RawMessage, bool, error),
) (json.RawMessage, bool, error) {
	if s.agentCallbacks == nil {
		return nil, false, status.Error(codes.FailedPrecondition, "agent callback service is not configured")
	}
	if req == nil {
		return nil, false, status.Error(codes.InvalidArgument, "request is required")
	}

	runSession, err := s.authenticateRunToken(ctx)
	if err != nil {
		return nil, false, err
	}

	payload, found, err := load(ctx, runSession.RunID)
	if err != nil {
		return nil, false, status.Errorf(codes.Internal, "failed to load %s", payloadLabel)
	}
	return payload, found, nil
}

func (s *Server) LookupRunPullRequest(ctx context.Context, req *controlplanev1.LookupRunPullRequestRequest) (*controlplanev1.LookupRunPullRequestResponse, error) {
	if s.agentCallbacks == nil {
		return nil, status.Error(codes.FailedPrecondition, "agent callback service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	runSession, err := s.authenticateRunToken(ctx)
	if err != nil {
		return nil, err
	}

	projectID := strings.TrimSpace(req.GetProjectId())
	if projectID == "" {
		projectID = strings.TrimSpace(runSession.ProjectID)
	}
	repositoryFullName := strings.TrimSpace(req.GetRepositoryFullName())
	headBranch := strings.TrimSpace(req.GetHeadBranch())
	pullRequestNumber := intFromOptional(req.GetPrNumber())
	if projectID == "" || repositoryFullName == "" || (pullRequestNumber <= 0 && headBranch == "") {
		return nil, status.Error(codes.InvalidArgument, "project_id, repository_full_name and one of pr_number/head_branch are required")
	}

	item, found, err := s.agentCallbacks.LookupPullRequest(ctx, agentcallbackdomain.LookupPullRequestQuery{
		ProjectID:          projectID,
		RepositoryFullName: repositoryFullName,
		PullRequestNumber:  pullRequestNumber,
		HeadBranch:         headBranch,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to lookup pull request")
	}
	if !found {
		return &controlplanev1.LookupRunPullRequestResponse{Found: false}, nil
	}

	return &controlplanev1.LookupRunPullRequestResponse{
		Found:      true,
		PrNumber:   int32(item.Number),
		PrUrl:      item.URL,
		PrState:    stringPtrOrNil(item.State),
		HeadBranch: stringPtrOrNil(item.Head),
		BaseBranch: stringPtrOrNil(item.Base),
	}, nil
}

func (s *Server) InsertRunFlowEvent(ctx context.Context, req *controlplanev1.InsertRunFlowEventRequest) (*controlplanev1.InsertRunFlowEventResponse, error) {
	if s.agentCallbacks == nil {
		return nil, status.Error(codes.FailedPrecondition, "agent callback service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	runSession, err := s.authenticateRunToken(ctx)
	if err != nil {
		return nil, err
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		runID = runSession.RunID
	}
	if runID != runSession.RunID {
		return nil, status.Error(codes.PermissionDenied, "run_id mismatch with token")
	}

	eventType, err := agentcallback.ParseEventType(req.GetEventType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.agentCallbacks.InsertRunFlowEvent(ctx, agentcallbackdomain.InsertRunFlowEventParams{
		CorrelationID: runSession.CorrelationID,
		EventType:     eventType,
		Payload:       json.RawMessage(req.GetPayloadJson()),
		CreatedAt:     time.Now().UTC(),
	}); err != nil {
		return nil, status.Error(codes.Internal, "failed to persist flow event")
	}

	return &controlplanev1.InsertRunFlowEventResponse{Ok: true, EventType: string(eventType)}, nil
}

func (s *Server) UpsertRunStatusComment(ctx context.Context, req *controlplanev1.UpsertRunStatusCommentRequest) (*controlplanev1.UpsertRunStatusCommentResponse, error) {
	if s.runStatus == nil {
		return nil, status.Error(codes.FailedPrecondition, "run status service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	runID := strings.TrimSpace(req.GetRunId())
	if runID == "" {
		return nil, status.Error(codes.InvalidArgument, "run_id is required")
	}

	phase, err := parseRunStatusPhase(req.GetPhase())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.runStatus.UpsertRunStatusComment(ctx, runstatusdomain.UpsertCommentParams{
		RunID:                    runID,
		Phase:                    phase,
		JobName:                  strings.TrimSpace(req.GetJobName()),
		JobNamespace:             strings.TrimSpace(req.GetJobNamespace()),
		RuntimeMode:              strings.TrimSpace(req.GetRuntimeMode()),
		Namespace:                strings.TrimSpace(req.GetNamespace()),
		TriggerKind:              strings.TrimSpace(req.GetTriggerKind()),
		PromptLocale:             strings.TrimSpace(req.GetPromptLocale()),
		Model:                    strings.TrimSpace(req.GetModel()),
		ReasoningEffort:          strings.TrimSpace(req.GetReasoningEffort()),
		RunStatus:                strings.TrimSpace(req.GetRunStatus()),
		CodexAuthVerificationURL: strings.TrimSpace(req.GetCodexAuthVerificationUrl()),
		CodexAuthUserCode:        strings.TrimSpace(req.GetCodexAuthUserCode()),
		Deleted:                  req.GetDeleted(),
		AlreadyDeleted:           req.GetAlreadyDeleted(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to upsert run status comment")
	}

	return &controlplanev1.UpsertRunStatusCommentResponse{
		Ok:         true,
		RunId:      runID,
		CommentId:  result.CommentID,
		CommentUrl: stringPtrOrNil(result.CommentURL),
	}, nil
}

func (s *Server) GetCodexAuth(ctx context.Context, req *controlplanev1.GetCodexAuthRequest) (*controlplanev1.GetCodexAuthResponse, error) {
	if s.codexAuth == nil {
		return nil, status.Error(codes.FailedPrecondition, "codex auth service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if _, err := s.authenticateRunToken(ctx); err != nil {
		return nil, err
	}

	authJSON, found, err := s.codexAuth.Get(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to load codex auth")
	}
	return &controlplanev1.GetCodexAuthResponse{
		Found:    found,
		AuthJson: authJSON,
	}, nil
}

func (s *Server) UpsertCodexAuth(ctx context.Context, req *controlplanev1.UpsertCodexAuthRequest) (*controlplanev1.UpsertCodexAuthResponse, error) {
	if s.codexAuth == nil {
		return nil, status.Error(codes.FailedPrecondition, "codex auth service is not configured")
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if _, err := s.authenticateRunToken(ctx); err != nil {
		return nil, err
	}

	if err := s.codexAuth.Upsert(ctx, req.GetAuthJson()); err != nil {
		return nil, status.Error(codes.Internal, "failed to persist codex auth")
	}
	return &controlplanev1.UpsertCodexAuthResponse{Ok: true}, nil
}

const sessionStatusRunning = "running"
