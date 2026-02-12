package controlplane

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/codex-k8s/codex-k8s/libs/go/grpcutil"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Client is an agent-runner wrapper over control-plane gRPC callbacks.
type Client struct {
	conn        *grpc.ClientConn
	svc         controlplanev1.ControlPlaneServiceClient
	bearerToken string
}

// SessionIdentity groups run identity fields persisted in agent session snapshots.
type SessionIdentity struct {
	RunID              string
	CorrelationID      string
	ProjectID          string
	RepositoryFullName string
	AgentKey           string
	IssueNumber        *int
	BranchName         string
	PRNumber           *int
	PRURL              string
}

// SessionTemplateContext captures template/model context used for this run.
type SessionTemplateContext struct {
	TriggerKind     string
	TemplateKind    string
	TemplateSource  string
	TemplateLocale  string
	Model           string
	ReasoningEffort string
}

// SessionRuntimeState captures session runtime status and codex snapshot files.
type SessionRuntimeState struct {
	Status           string
	SessionID        string
	SessionJSON      json.RawMessage
	CodexSessionPath string
	CodexSessionJSON json.RawMessage
	StartedAt        time.Time
	FinishedAt       *time.Time
}

// AgentSessionUpsertParams defines payload for session persistence callback.
type AgentSessionUpsertParams struct {
	Identity SessionIdentity
	Template SessionTemplateContext
	Runtime  SessionRuntimeState
}

// AgentSessionSnapshot is latest persisted session snapshot for resume.
type AgentSessionSnapshot struct {
	RunID            string
	CorrelationID    string
	ProjectID        string
	RepositoryName   string
	AgentKey         string
	IssueNumber      int
	BranchName       string
	PRNumber         int
	PRURL            string
	TriggerKind      string
	TemplateKind     string
	TemplateSource   string
	TemplateLocale   string
	Model            string
	ReasoningEffort  string
	Status           string
	SessionID        string
	SessionJSON      json.RawMessage
	CodexSessionPath string
	CodexSessionJSON json.RawMessage
	StartedAt        time.Time
	FinishedAt       time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// LatestAgentSessionQuery describes latest-session lookup identity.
type LatestAgentSessionQuery struct {
	RepositoryFullName string
	BranchName         string
	AgentKey           string
}

// Dial creates control-plane gRPC client with run-bound bearer auth.
func Dial(ctx context.Context, target string, bearerToken string) (*Client, error) {
	conn, err := grpcutil.DialInsecureReady(ctx, strings.TrimSpace(target))
	if err != nil {
		return nil, fmt.Errorf("dial control-plane grpc: %w", err)
	}
	return &Client{
		conn:        conn,
		svc:         controlplanev1.NewControlPlaneServiceClient(conn),
		bearerToken: strings.TrimSpace(bearerToken),
	}, nil
}

// Close closes underlying gRPC connection.
func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// UpsertAgentSession stores or updates run session snapshot.
func (c *Client) UpsertAgentSession(ctx context.Context, params AgentSessionUpsertParams) error {
	identity := params.Identity
	template := params.Template
	runtime := params.Runtime

	request := &controlplanev1.UpsertAgentSessionRequest{
		RunId:               strings.TrimSpace(identity.RunID),
		CorrelationId:       strings.TrimSpace(identity.CorrelationID),
		ProjectId:           strings.TrimSpace(identity.ProjectID),
		RepositoryFullName:  strings.TrimSpace(identity.RepositoryFullName),
		AgentKey:            strings.TrimSpace(identity.AgentKey),
		IssueNumber:         intToOptional(identity.IssueNumber),
		BranchName:          strings.TrimSpace(identity.BranchName),
		PrNumber:            intToOptional(identity.PRNumber),
		PrUrl:               strings.TrimSpace(identity.PRURL),
		TriggerKind:         strings.TrimSpace(template.TriggerKind),
		TemplateKind:        strings.TrimSpace(template.TemplateKind),
		TemplateSource:      strings.TrimSpace(template.TemplateSource),
		TemplateLocale:      strings.TrimSpace(template.TemplateLocale),
		Model:               strings.TrimSpace(template.Model),
		ReasoningEffort:     strings.TrimSpace(template.ReasoningEffort),
		Status:              strings.TrimSpace(runtime.Status),
		SessionId:           strings.TrimSpace(runtime.SessionID),
		SessionJson:         []byte(runtime.SessionJSON),
		CodexCliSessionPath: strings.TrimSpace(runtime.CodexSessionPath),
		CodexCliSessionJson: []byte(runtime.CodexSessionJSON),
		StartedAt:           timestamppb.New(runtime.StartedAt.UTC()),
		FinishedAt:          optionalTimestamp(runtime.FinishedAt),
	}

	_, err := c.svc.UpsertAgentSession(c.withAuth(ctx), request)
	if err != nil {
		return fmt.Errorf("upsert agent session: %w", err)
	}
	return nil
}

// GetLatestAgentSession loads latest snapshot by repository/branch/agent key.
func (c *Client) GetLatestAgentSession(ctx context.Context, query LatestAgentSessionQuery) (AgentSessionSnapshot, bool, error) {
	resp, err := c.svc.GetLatestAgentSession(c.withAuth(ctx), &controlplanev1.GetLatestAgentSessionRequest{
		RepositoryFullName: strings.TrimSpace(query.RepositoryFullName),
		BranchName:         strings.TrimSpace(query.BranchName),
		AgentKey:           strings.TrimSpace(query.AgentKey),
	})
	if err != nil {
		return AgentSessionSnapshot{}, false, fmt.Errorf("get latest agent session: %w", err)
	}
	if !resp.GetFound() || resp.GetSession() == nil {
		return AgentSessionSnapshot{}, false, nil
	}

	snapshot := resp.GetSession()
	result := AgentSessionSnapshot{
		RunID:            strings.TrimSpace(snapshot.GetRunId()),
		CorrelationID:    strings.TrimSpace(snapshot.GetCorrelationId()),
		ProjectID:        strings.TrimSpace(snapshot.GetProjectId()),
		RepositoryName:   strings.TrimSpace(snapshot.GetRepositoryFullName()),
		AgentKey:         strings.TrimSpace(snapshot.GetAgentKey()),
		IssueNumber:      optionalToInt(snapshot.GetIssueNumber()),
		BranchName:       strings.TrimSpace(snapshot.GetBranchName()),
		PRNumber:         optionalToInt(snapshot.GetPrNumber()),
		PRURL:            strings.TrimSpace(snapshot.GetPrUrl()),
		TriggerKind:      strings.TrimSpace(snapshot.GetTriggerKind()),
		TemplateKind:     strings.TrimSpace(snapshot.GetTemplateKind()),
		TemplateSource:   strings.TrimSpace(snapshot.GetTemplateSource()),
		TemplateLocale:   strings.TrimSpace(snapshot.GetTemplateLocale()),
		Model:            strings.TrimSpace(snapshot.GetModel()),
		ReasoningEffort:  strings.TrimSpace(snapshot.GetReasoningEffort()),
		Status:           strings.TrimSpace(snapshot.GetStatus()),
		SessionID:        strings.TrimSpace(snapshot.GetSessionId()),
		SessionJSON:      json.RawMessage(snapshot.GetSessionJson()),
		CodexSessionPath: strings.TrimSpace(snapshot.GetCodexCliSessionPath()),
		CodexSessionJSON: json.RawMessage(snapshot.GetCodexCliSessionJson()),
		StartedAt:        timestampOrZero(snapshot.GetStartedAt()),
		FinishedAt:       timestampOrZero(snapshot.GetFinishedAt()),
		CreatedAt:        timestampOrZero(snapshot.GetCreatedAt()),
		UpdatedAt:        timestampOrZero(snapshot.GetUpdatedAt()),
	}
	return result, true, nil
}

// InsertRunFlowEvent persists one run-bound flow event.
func (c *Client) InsertRunFlowEvent(ctx context.Context, runID string, eventType flowevent.EventType, payload json.RawMessage) error {
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}

	_, err := c.svc.InsertRunFlowEvent(c.withAuth(ctx), &controlplanev1.InsertRunFlowEventRequest{
		RunId:       strings.TrimSpace(runID),
		EventType:   strings.TrimSpace(string(eventType)),
		PayloadJson: []byte(payload),
	})
	if err != nil {
		return fmt.Errorf("insert run flow event: %w", err)
	}
	return nil
}

func (c *Client) withAuth(ctx context.Context) context.Context {
	token := strings.TrimSpace(c.bearerToken)
	if token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

func intToOptional(value *int) *wrapperspb.Int32Value {
	if value == nil || *value <= 0 {
		return nil
	}
	return wrapperspb.Int32(int32(*value))
}

func optionalToInt(value *wrapperspb.Int32Value) int {
	if value == nil || value.Value <= 0 {
		return 0
	}
	return int(value.Value)
}

func optionalTimestamp(value *time.Time) *timestamppb.Timestamp {
	if value == nil {
		return nil
	}
	return timestamppb.New(value.UTC())
}

func timestampOrZero(value *timestamppb.Timestamp) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.AsTime().UTC()
}
