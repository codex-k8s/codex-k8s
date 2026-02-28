package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	agentcallbackdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/agentcallback"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	runstatusdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runstatus"
	runtimedeploydomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runtimedeploy"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/staff"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
	"github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/webhook"
	agentcallback "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/transport/agentcallback"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type webhookIngress interface {
	IngestGitHubWebhook(ctx context.Context, cmd webhook.IngestCommand) (webhook.IngestResult, error)
}

type mcpRunTokenService interface {
	IssueRunToken(ctx context.Context, params mcpdomain.IssueRunTokenParams) (mcpdomain.IssuedToken, error)
	VerifyRunToken(ctx context.Context, rawToken string) (mcpdomain.SessionContext, error)
	ListPendingApprovals(ctx context.Context, limit int) ([]mcpdomain.ApprovalListItem, error)
	ResolveApproval(ctx context.Context, params mcpdomain.ResolveApprovalParams) (mcpdomain.ResolveApprovalResult, error)
}

type agentCallbackService interface {
	UpsertAgentSession(ctx context.Context, params agentcallbackdomain.UpsertAgentSessionParams) error
	GetLatestAgentSession(ctx context.Context, query agentcallbackdomain.GetLatestAgentSessionQuery) (agentcallbackdomain.Session, bool, error)
	InsertRunFlowEvent(ctx context.Context, params agentcallbackdomain.InsertRunFlowEventParams) error
}

type runStatusService interface {
	UpsertRunStatusComment(ctx context.Context, params runstatusdomain.UpsertCommentParams) (runstatusdomain.UpsertCommentResult, error)
}

type runtimeDeployService interface {
	PrepareRunEnvironment(ctx context.Context, params runtimedeploydomain.PrepareParams) (runtimedeploydomain.PrepareResult, error)
}

type runtimeErrorRecorder interface {
	RecordBestEffort(ctx context.Context, params querytypes.RuntimeErrorRecordParams)
}

type codexAuthService interface {
	Get(ctx context.Context) ([]byte, bool, error)
	Upsert(ctx context.Context, authJSON []byte) error
}

// Dependencies wires domain services and repositories into the gRPC transport.
type Dependencies struct {
	Webhook        webhookIngress
	Staff          *staff.Service
	AgentCallbacks agentCallbackService
	RunStatus      runStatusService
	RuntimeDeploy  runtimeDeployService
	RuntimeErrors  runtimeErrorRecorder
	MCP            mcpRunTokenService
	CodexAuth      codexAuthService
	Logger         *slog.Logger
}

// Server implements ControlPlaneServiceServer.
type Server struct {
	controlplanev1.UnimplementedControlPlaneServiceServer

	webhook        webhookIngress
	staff          *staff.Service
	agentCallbacks agentCallbackService
	runStatus      runStatusService
	runtimeDeploy  runtimeDeployService
	runtimeErrors  runtimeErrorRecorder
	mcp            mcpRunTokenService
	codexAuth      codexAuthService
	logger         *slog.Logger
}

func NewServer(deps Dependencies) *Server {
	return &Server{
		webhook:        deps.Webhook,
		staff:          deps.Staff,
		agentCallbacks: deps.AgentCallbacks,
		runStatus:      deps.RunStatus,
		runtimeDeploy:  deps.RuntimeDeploy,
		runtimeErrors:  deps.RuntimeErrors,
		mcp:            deps.MCP,
		codexAuth:      deps.CodexAuth,
		logger:         deps.Logger,
	}
}

func (s *Server) IngestGitHubWebhook(ctx context.Context, req *controlplanev1.IngestGitHubWebhookRequest) (*controlplanev1.IngestGitHubWebhookResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	res, err := s.webhook.IngestGitHubWebhook(ctx, webhook.IngestCommand{
		CorrelationID: strings.TrimSpace(req.CorrelationId),
		EventType:     strings.TrimSpace(req.EventType),
		DeliveryID:    strings.TrimSpace(req.DeliveryId),
		ReceivedAt:    tsToTime(req.ReceivedAt),
		Payload:       req.PayloadJson,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &controlplanev1.IngestGitHubWebhookResponse{
		CorrelationId: res.CorrelationID,
		RunId:         res.RunID,
		Status:        string(res.Status),
		Duplicate:     res.Duplicate,
	}, nil
}

func (s *Server) ResolveStaffByEmail(ctx context.Context, req *controlplanev1.ResolveStaffByEmailRequest) (*controlplanev1.ResolveStaffByEmailResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	u, err := s.staff.ResolveStaffByEmail(ctx, querytypes.StaffResolveByEmailParams{
		Email:       strings.TrimSpace(req.Email),
		GitHubLogin: strings.TrimSpace(req.GetGithubLogin()),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.ResolveStaffByEmailResponse{Principal: userToPrincipal(u)}, nil
}

func (s *Server) AuthorizeOAuthUser(ctx context.Context, req *controlplanev1.AuthorizeOAuthUserRequest) (*controlplanev1.AuthorizeOAuthUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	u, err := s.staff.AuthorizeOAuthUser(ctx, querytypes.StaffAuthorizeOAuthUserParams{
		Email:        strings.TrimSpace(req.Email),
		GitHubUserID: req.GetGithubUserId(),
		GitHubLogin:  strings.TrimSpace(req.GetGithubLogin()),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &controlplanev1.AuthorizeOAuthUserResponse{Principal: userToPrincipal(u)}, nil
}

func (s *Server) authenticateRunToken(ctx context.Context) (mcpdomain.SessionContext, error) {
	if s.mcp == nil {
		return mcpdomain.SessionContext{}, status.Error(codes.FailedPrecondition, "mcp service is not configured")
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return mcpdomain.SessionContext{}, status.Error(codes.Unauthenticated, "missing bearer token")
	}

	rawToken := bearerTokenFromMetadata(md)
	if rawToken == "" {
		return mcpdomain.SessionContext{}, status.Error(codes.Unauthenticated, "missing bearer token")
	}

	runSession, err := s.mcp.VerifyRunToken(ctx, rawToken)
	if err != nil {
		return mcpdomain.SessionContext{}, status.Error(codes.Unauthenticated, "invalid bearer token")
	}
	return runSession, nil
}

func bearerTokenFromMetadata(md metadata.MD) string {
	for _, value := range md.Get("authorization") {
		token := agentcallback.ParseBearerToken(value)
		if token != "" {
			return token
		}
	}
	return ""
}

func intPtrFromOptional(value *wrapperspb.Int32Value) *int {
	if value == nil || value.Value <= 0 {
		return nil
	}
	v := int(value.Value)
	return &v
}

func intToOptional(value int32) *wrapperspb.Int32Value {
	if value <= 0 {
		return nil
	}
	return wrapperspb.Int32(value)
}

func optionalTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	value := ts.AsTime().UTC()
	return &value
}

type delete1Fn func(context.Context, staff.Principal, string) error
type delete2Fn func(context.Context, staff.Principal, string, string) error

func (s *Server) delete1(ctx context.Context, principal *controlplanev1.Principal, id string, fn delete1Fn) (*emptypb.Empty, error) {
	p, err := requirePrincipal(principal)
	if err != nil {
		return nil, err
	}
	if err := fn(ctx, p, strings.TrimSpace(id)); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) delete2(ctx context.Context, principal *controlplanev1.Principal, id1 string, id2 string, fn delete2Fn) (*emptypb.Empty, error) {
	p, err := requirePrincipal(principal)
	if err != nil {
		return nil, err
	}
	if err := fn(ctx, p, strings.TrimSpace(id1), strings.TrimSpace(id2)); err != nil {
		return nil, toStatus(err)
	}
	return &emptypb.Empty{}, nil
}

func requirePrincipal(p *controlplanev1.Principal) (staff.Principal, error) {
	if p == nil || strings.TrimSpace(p.UserId) == "" {
		return staff.Principal{}, status.Error(codes.Unauthenticated, "not authenticated")
	}
	return staff.Principal{
		UserID:          strings.TrimSpace(p.UserId),
		Email:           strings.TrimSpace(p.Email),
		GitHubLogin:     strings.TrimSpace(p.GithubLogin),
		IsPlatformAdmin: p.IsPlatformAdmin,
		IsPlatformOwner: p.IsPlatformOwner,
	}, nil
}

func userToPrincipal(u entitytypes.User) *controlplanev1.Principal {
	return &controlplanev1.Principal{
		UserId:          u.ID,
		Email:           u.Email,
		GithubLogin:     u.GitHubLogin,
		IsPlatformAdmin: u.IsPlatformAdmin,
		IsPlatformOwner: u.IsPlatformOwner,
	}
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}

	// Preserve cancellation semantics for callers that implement retry logic based on gRPC codes.
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, context.Canceled.Error())
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, context.DeadlineExceeded.Error())
	}
	if errors.Is(err, runtimedeploydomain.ErrTaskCanceled) {
		msg := strings.TrimSpace(err.Error())
		if msg == "" {
			msg = runtimedeploydomain.ErrTaskCanceled.Error()
		}
		return status.Error(codes.Canceled, msg)
	}
	// Pass through existing gRPC statuses unchanged.
	if _, ok := status.FromError(err); ok {
		return err
	}

	var v errs.Validation
	var u errs.Unauthorized
	var f errs.Forbidden
	var c errs.Conflict

	switch {
	case errors.As(err, &v):
		msg := v.Msg
		if v.Field != "" {
			msg = fmt.Sprintf("%s: %s", v.Field, v.Msg)
		}
		return status.Error(codes.InvalidArgument, msg)
	case errors.As(err, &u):
		return status.Error(codes.Unauthenticated, u.Error())
	case errors.As(err, &f):
		return status.Error(codes.PermissionDenied, f.Error())
	case errors.As(err, &c):
		return status.Error(codes.AlreadyExists, c.Error())
	default:
		// Prefer explicit error text for debugging. Keep it bounded to avoid huge status messages
		// when upstream includes logs in errors.
		msg := strings.TrimSpace(err.Error())
		if msg == "" {
			msg = "internal error"
		}
		const maxLen = 1024
		if len(msg) > maxLen {
			msg = msg[:maxLen] + "..."
		}
		return status.Error(codes.Internal, msg)
	}
}

func clampLimit(n int32, def int) int {
	if n <= 0 {
		return def
	}
	if n > 1000 {
		return 1000
	}
	return int(n)
}

func tsToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime().UTC()
}

func parseRuntimeErrorListState(raw string) querytypes.RuntimeErrorListState {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(querytypes.RuntimeErrorListStateAll):
		return querytypes.RuntimeErrorListStateAll
	case string(querytypes.RuntimeErrorListStateViewed):
		return querytypes.RuntimeErrorListStateViewed
	default:
		return querytypes.RuntimeErrorListStateActive
	}
}

func runToProto(r entitytypes.StaffRun) *controlplanev1.Run {
	out := &controlplanev1.Run{
		Id:              r.ID,
		CorrelationId:   r.CorrelationID,
		ProjectId:       stringPtrOrNil(r.ProjectID),
		ProjectSlug:     stringPtrOrNil(r.ProjectSlug),
		ProjectName:     stringPtrOrNil(r.ProjectName),
		IssueNumber:     int32PtrOrNil(int32(r.IssueNumber)),
		IssueUrl:        stringPtrOrNil(r.IssueURL),
		PrNumber:        int32PtrOrNil(int32(r.PRNumber)),
		PrUrl:           stringPtrOrNil(r.PRURL),
		TriggerKind:     stringPtrOrNil(r.TriggerKind),
		TriggerLabel:    stringPtrOrNil(r.TriggerLabel),
		AgentKey:        stringPtrOrNil(r.AgentKey),
		JobName:         stringPtrOrNil(r.JobName),
		JobNamespace:    stringPtrOrNil(r.JobNamespace),
		Namespace:       stringPtrOrNil(r.Namespace),
		JobExists:       r.JobExists,
		NamespaceExists: r.NamespaceExists,
		WaitState:       stringPtrOrNil(r.WaitState),
		WaitReason:      stringPtrOrNil(r.WaitReason),
		Status:          r.Status,
		CreatedAt:       timestamppb.New(r.CreatedAt.UTC()),
	}
	if r.StartedAt != nil {
		out.StartedAt = timestamppb.New(r.StartedAt.UTC())
	}
	if r.FinishedAt != nil {
		out.FinishedAt = timestamppb.New(r.FinishedAt.UTC())
	}
	if r.WaitSince != nil {
		out.WaitSince = timestamppb.New(r.WaitSince.UTC())
	}
	if r.LastHeartbeatAt != nil {
		out.LastHeartbeatAt = timestamppb.New(r.LastHeartbeatAt.UTC())
	}
	return out
}

func runtimeErrorToProto(item entitytypes.RuntimeError) *controlplanev1.RuntimeError {
	detailsJSON := strings.TrimSpace(string(item.DetailsJSON))
	if detailsJSON == "" {
		detailsJSON = "{}"
	}
	out := &controlplanev1.RuntimeError{
		Id:            item.ID,
		Source:        item.Source,
		Level:         item.Level,
		Message:       item.Message,
		DetailsJson:   detailsJSON,
		StackTrace:    stringPtrOrNil(item.StackTrace),
		CorrelationId: stringPtrOrNil(item.CorrelationID),
		RunId:         stringPtrOrNil(item.RunID),
		ProjectId:     stringPtrOrNil(item.ProjectID),
		Namespace:     stringPtrOrNil(item.Namespace),
		JobName:       stringPtrOrNil(item.JobName),
		ViewedBy:      stringPtrOrNil(item.ViewedBy),
		CreatedAt:     timestamppb.New(item.CreatedAt.UTC()),
	}
	if item.ViewedAt != nil {
		out.ViewedAt = timestamppb.New(item.ViewedAt.UTC())
	}
	return out
}

func runtimeDeployTaskToProto(item entitytypes.RuntimeDeployTask) *controlplanev1.RuntimeDeployTask {
	out := &controlplanev1.RuntimeDeployTask{
		RunId:              item.RunID,
		RuntimeMode:        item.RuntimeMode,
		Namespace:          item.Namespace,
		TargetEnv:          item.TargetEnv,
		SlotNo:             int32(item.SlotNo),
		RepositoryFullName: item.RepositoryFullName,
		ServicesYamlPath:   item.ServicesYAMLPath,
		BuildRef:           item.BuildRef,
		DeployOnly:         item.DeployOnly,
		Status:             string(item.Status),
		LeaseOwner:         stringPtrOrNil(item.LeaseOwner),
		Attempts:           int32(item.Attempts),
		LastError:          stringPtrOrNil(item.LastError),
		ResultNamespace:    stringPtrOrNil(item.ResultNamespace),
		ResultTargetEnv:    stringPtrOrNil(item.ResultTargetEnv),
		CreatedAt:          timestamppb.New(item.CreatedAt.UTC()),
		UpdatedAt:          timestamppb.New(item.UpdatedAt.UTC()),
		Logs:               runtimeDeployLogsToProto(item.Logs),
	}
	if !item.LeaseUntil.IsZero() {
		out.LeaseUntil = timestamppb.New(item.LeaseUntil.UTC())
	}
	if !item.StartedAt.IsZero() {
		out.StartedAt = timestamppb.New(item.StartedAt.UTC())
	}
	if !item.FinishedAt.IsZero() {
		out.FinishedAt = timestamppb.New(item.FinishedAt.UTC())
	}
	return out
}

func runtimeDeployLogsToProto(items []entitytypes.RuntimeDeployTaskLogEntry) []*controlplanev1.RuntimeDeployTaskLog {
	if len(items) == 0 {
		return []*controlplanev1.RuntimeDeployTaskLog{}
	}
	out := make([]*controlplanev1.RuntimeDeployTaskLog, 0, len(items))
	for _, item := range items {
		logItem := &controlplanev1.RuntimeDeployTaskLog{
			Stage:   strings.TrimSpace(item.Stage),
			Level:   strings.TrimSpace(item.Level),
			Message: strings.TrimSpace(item.Message),
		}
		if !item.CreatedAt.IsZero() {
			logItem.CreatedAt = timestamppb.New(item.CreatedAt.UTC())
		}
		out = append(out, logItem)
	}
	return out
}

func registryImageRepositoryToProto(item entitytypes.RegistryImageRepository) *controlplanev1.RegistryImageRepository {
	out := &controlplanev1.RegistryImageRepository{
		Repository: item.Repository,
		TagCount:   int32(item.TagCount),
		Tags:       make([]*controlplanev1.RegistryImageTag, 0, len(item.Tags)),
	}
	for _, tag := range item.Tags {
		out.Tags = append(out.Tags, registryImageTagToProto(tag))
	}
	return out
}

func registryImageTagToProto(item entitytypes.RegistryImageTag) *controlplanev1.RegistryImageTag {
	out := &controlplanev1.RegistryImageTag{
		Tag:             item.Tag,
		Digest:          item.Digest,
		ConfigSizeBytes: item.ConfigSizeBytes,
	}
	if item.CreatedAt != nil {
		out.CreatedAt = timestamppb.New(item.CreatedAt.UTC())
	}
	return out
}

func registryImageDeleteToProto(item entitytypes.RegistryImageDeleteResult) *controlplanev1.RegistryImageDeleteResult {
	return &controlplanev1.RegistryImageDeleteResult{
		Repository: item.Repository,
		Tag:        item.Tag,
		Digest:     item.Digest,
		Deleted:    item.Deleted,
	}
}

func approvalToProto(item mcpdomain.ApprovalListItem) *controlplanev1.ApprovalRequest {
	return &controlplanev1.ApprovalRequest{
		Id:            item.ID,
		CorrelationId: item.CorrelationID,
		RunId:         stringPtrOrNil(item.RunID),
		ProjectId:     stringPtrOrNil(item.ProjectID),
		ProjectSlug:   stringPtrOrNil(item.ProjectSlug),
		ProjectName:   stringPtrOrNil(item.ProjectName),
		IssueNumber:   intToOptional(int32(item.IssueNumber)),
		PrNumber:      intToOptional(int32(item.PRNumber)),
		TriggerLabel:  stringPtrOrNil(item.TriggerLabel),
		ToolName:      item.ToolName,
		Action:        item.Action,
		ApprovalMode:  item.ApprovalMode,
		RequestedBy:   item.RequestedBy,
		CreatedAt:     timestamppb.New(item.CreatedAt.UTC()),
	}
}

func approvalActorID(principal staff.Principal) string {
	if value := strings.TrimSpace(principal.GitHubLogin); value != "" {
		return "staff:" + value
	}
	if value := strings.TrimSpace(principal.Email); value != "" {
		return "staff:" + value
	}
	return "staff:" + strings.TrimSpace(principal.UserID)
}

func stringPtrOrNil(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func int32PtrOrNil(value int32) *int32 {
	if value <= 0 {
		return nil
	}
	return &value
}

func int64PtrOrNil(value int64) *int64 {
	if value <= 0 {
		return nil
	}
	return &value
}

func optionalProtoString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func bytesOrNil(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	return value
}

func parseRuntimeMode(value string) agentdomain.RuntimeMode {
	if strings.EqualFold(strings.TrimSpace(value), string(agentdomain.RuntimeModeFullEnv)) {
		return agentdomain.RuntimeModeFullEnv
	}
	return agentdomain.RuntimeModeCodeOnly
}

func parseRunStatusPhase(value string) (runstatusdomain.Phase, error) {
	switch strings.TrimSpace(value) {
	case string(runstatusdomain.PhaseCreated):
		return runstatusdomain.PhaseCreated, nil
	case string(runstatusdomain.PhasePreparingRuntime):
		return runstatusdomain.PhasePreparingRuntime, nil
	case string(runstatusdomain.PhaseStarted):
		return runstatusdomain.PhaseStarted, nil
	case string(runstatusdomain.PhaseAuthRequired):
		return runstatusdomain.PhaseAuthRequired, nil
	case string(runstatusdomain.PhaseAuthResolved):
		return runstatusdomain.PhaseAuthResolved, nil
	case string(runstatusdomain.PhaseFinished):
		return runstatusdomain.PhaseFinished, nil
	case string(runstatusdomain.PhaseNamespaceDeleted):
		return runstatusdomain.PhaseNamespaceDeleted, nil
	default:
		return "", fmt.Errorf("unsupported phase %q", value)
	}
}
