package grpc

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	agentcallbackdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/agentcallback"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	runstatusdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runstatus"
	runtimedeploydomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runtimedeploy"
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

func (s *Server) ListRuntimeDeployTasks(ctx context.Context, req *controlplanev1.ListRuntimeDeployTasksRequest) (*controlplanev1.ListRuntimeDeployTasksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	items, err := s.staff.ListRuntimeDeployTasks(ctx, p, clampLimit(req.GetLimit(), 200), optionalProtoString(req.Status), optionalProtoString(req.TargetEnv))
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.RuntimeDeployTask, 0, len(items))
	for _, item := range items {
		out = append(out, runtimeDeployTaskToProto(item))
	}
	return &controlplanev1.ListRuntimeDeployTasksResponse{Items: out}, nil
}

func (s *Server) GetRuntimeDeployTask(ctx context.Context, req *controlplanev1.GetRuntimeDeployTaskRequest) (*controlplanev1.RuntimeDeployTask, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	item, err := s.staff.GetRuntimeDeployTask(ctx, p, strings.TrimSpace(req.GetRunId()))
	if err != nil {
		return nil, toStatus(err)
	}
	return runtimeDeployTaskToProto(item), nil
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
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	item, err := s.staff.MarkRuntimeErrorViewed(ctx, p, strings.TrimSpace(req.GetRuntimeErrorId()))
	if err != nil {
		return nil, toStatus(err)
	}
	return runtimeErrorToProto(item), nil
}

func (s *Server) ListRegistryImages(ctx context.Context, req *controlplanev1.ListRegistryImagesRequest) (*controlplanev1.ListRegistryImagesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	items, err := s.staff.ListRegistryImages(ctx, p, querytypes.RegistryImageListFilter{
		Repository:        optionalProtoString(req.Repository),
		LimitRepositories: clampLimit(req.GetLimitRepositories(), 100),
		LimitTags:         clampLimit(req.GetLimitTags(), 50),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*controlplanev1.RegistryImageRepository, 0, len(items))
	for _, item := range items {
		out = append(out, registryImageRepositoryToProto(item))
	}
	return &controlplanev1.ListRegistryImagesResponse{Items: out}, nil
}

func (s *Server) DeleteRegistryImageTag(ctx context.Context, req *controlplanev1.DeleteRegistryImageTagRequest) (*controlplanev1.RegistryImageDeleteResult, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	item, err := s.staff.DeleteRegistryImageTag(ctx, p, querytypes.RegistryImageDeleteParams{
		Repository: strings.TrimSpace(req.GetRepository()),
		Tag:        strings.TrimSpace(req.GetTag()),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return registryImageDeleteToProto(item), nil
}

func (s *Server) CleanupRegistryImages(ctx context.Context, req *controlplanev1.CleanupRegistryImagesRequest) (*controlplanev1.CleanupRegistryImagesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	p, err := requirePrincipal(req.GetPrincipal())
	if err != nil {
		return nil, err
	}
	item, err := s.staff.CleanupRegistryImages(ctx, p, querytypes.RegistryImageCleanupFilter{
		RepositoryPrefix:  optionalProtoString(req.RepositoryPrefix),
		LimitRepositories: clampLimit(req.GetLimitRepositories(), 100),
		KeepTags:          int(req.GetKeepTags()),
		DryRun:            req.GetDryRun(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	deleted := make([]*controlplanev1.RegistryImageDeleteResult, 0, len(item.Deleted))
	for _, deleteItem := range item.Deleted {
		deleted = append(deleted, registryImageDeleteToProto(deleteItem))
	}
	skipped := make([]*controlplanev1.RegistryImageDeleteResult, 0, len(item.Skipped))
	for _, skipItem := range item.Skipped {
		skipped = append(skipped, registryImageDeleteToProto(skipItem))
	}
	return &controlplanev1.CleanupRegistryImagesResponse{
		RepositoriesScanned: int32(item.RepositoriesScanned),
		TagsDeleted:         int32(item.TagsDeleted),
		TagsSkipped:         int32(item.TagsSkipped),
		Deleted:             deleted,
		Skipped:             skipped,
	}, nil
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

	if err := s.agentCallbacks.UpsertAgentSession(ctx, agentcallbackdomain.UpsertAgentSessionParams{
		RunID:              runID,
		CorrelationID:      correlationID,
		ProjectID:          projectID,
		RepositoryFullName: repositoryFullName,
		AgentKey:           agentKey,
		IssueNumber:        intPtrFromOptional(req.GetIssueNumber()),
		BranchName:         branchName,
		PRNumber:           intPtrFromOptional(req.GetPrNumber()),
		PRURL:              strings.TrimSpace(req.GetPrUrl()),
		TriggerKind:        strings.TrimSpace(req.GetTriggerKind()),
		TemplateKind:       strings.TrimSpace(req.GetTemplateKind()),
		TemplateSource:     strings.TrimSpace(req.GetTemplateSource()),
		TemplateLocale:     strings.TrimSpace(req.GetTemplateLocale()),
		Model:              strings.TrimSpace(req.GetModel()),
		ReasoningEffort:    strings.TrimSpace(req.GetReasoningEffort()),
		Status:             statusValue,
		SessionID:          strings.TrimSpace(req.GetSessionId()),
		SessionJSON:        json.RawMessage(req.GetSessionJson()),
		CodexSessionPath:   strings.TrimSpace(req.GetCodexCliSessionPath()),
		CodexSessionJSON:   json.RawMessage(req.GetCodexCliSessionJson()),
		StartedAt:          startedAt,
		FinishedAt:         optionalTime(req.GetFinishedAt()),
	}); err != nil {
		s.logger.Error("upsert agent session via grpc failed", "run_id", runID, "err", err)
		return nil, status.Error(codes.Internal, "failed to persist agent session")
	}

	return &controlplanev1.UpsertAgentSessionResponse{Ok: true, RunId: runID}, nil
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
		s.logger.Error("get latest agent session via grpc failed", "repository_full_name", repositoryFullName, "branch_name", branchName, "agent_key", agentKey, "err", err)
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
		StartedAt:           timestamppb.New(item.StartedAt.UTC()),
		CreatedAt:           timestamppb.New(item.CreatedAt.UTC()),
		UpdatedAt:           timestamppb.New(item.UpdatedAt.UTC()),
	}
	if !item.FinishedAt.IsZero() {
		snapshot.FinishedAt = timestamppb.New(item.FinishedAt.UTC())
	}

	return &controlplanev1.GetLatestAgentSessionResponse{
		Found:   true,
		Session: snapshot,
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
		s.logger.Error("insert run flow event via grpc failed", "run_id", runID, "event_type", eventType, "err", err)
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
		s.logger.Error("upsert run status comment via grpc failed", "run_id", runID, "phase", phase, "err", err)
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
		s.logger.Error("get codex auth via grpc failed", "err", err)
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
		s.logger.Error("upsert codex auth via grpc failed", "err", err)
		return nil, status.Error(codes.Internal, "failed to persist codex auth")
	}
	return &controlplanev1.UpsertCodexAuthResponse{Ok: true}, nil
}

const sessionStatusRunning = "running"
