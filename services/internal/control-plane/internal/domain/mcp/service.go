package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	rundomain "github.com/codex-k8s/codex-k8s/libs/go/domain/run"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultTokenIssuer        = "codex-k8s/control-plane/mcp"
	defaultServerName         = "codex-k8s-control-plane-mcp"
	defaultInternalMCPBaseURL = "http://codex-k8s-control-plane:8081/mcp"
	minTokenTTL               = 24 * time.Hour
	defaultTokenTTL           = minTokenTTL
	maxTokenTTL               = 7 * 24 * time.Hour
)

// Config defines MCP domain behavior.
type Config struct {
	TokenSigningKey    string
	TokenIssuer        string
	ServerName         string
	PublicBaseURL      string
	InternalMCPBaseURL string
	DefaultTokenTTL    time.Duration
}

// GitHubClient defines GitHub operations used by MCP tools.
type GitHubClient interface {
	GetIssue(ctx context.Context, params GitHubGetIssueParams) (GitHubIssue, error)
	GetPullRequest(ctx context.Context, params GitHubGetPullRequestParams) (GitHubPullRequest, error)
	ListIssueComments(ctx context.Context, params GitHubListIssueCommentsParams) ([]GitHubIssueComment, error)
	ListIssueLabels(ctx context.Context, params GitHubListIssueLabelsParams) ([]GitHubLabel, error)
	ListBranches(ctx context.Context, params GitHubListBranchesParams) ([]GitHubBranch, error)
	EnsureBranch(ctx context.Context, params GitHubEnsureBranchParams) (GitHubBranch, error)
	UpsertPullRequest(ctx context.Context, params GitHubUpsertPullRequestParams) (GitHubPullRequest, error)
	CreateIssueComment(ctx context.Context, params GitHubCreateIssueCommentParams) (GitHubIssueComment, error)
	AddLabels(ctx context.Context, params GitHubMutateLabelsParams) ([]GitHubLabel, error)
	RemoveLabels(ctx context.Context, params GitHubMutateLabelsParams) ([]GitHubLabel, error)
}

// KubernetesClient defines Kubernetes operations used by MCP tools.
type KubernetesClient interface {
	ListPods(ctx context.Context, namespace string, limit int) ([]KubernetesPod, error)
	ListEvents(ctx context.Context, namespace string, limit int) ([]KubernetesEvent, error)
	ListResources(ctx context.Context, namespace string, kind KubernetesResourceKind, limit int) ([]KubernetesResourceRef, error)
	GetPodLogs(ctx context.Context, namespace string, pod string, container string, tailLines int64) (string, error)
	ExecPod(ctx context.Context, namespace string, pod string, container string, command []string) (KubernetesExecResult, error)
}

// Service provides MCP token handling, prompt context building and tool operations.
type Service struct {
	cfg Config

	runs       agentrunrepo.Repository
	flowEvents floweventrepo.Repository
	repos      repocfgrepo.Repository
	tokenCrypt *tokencrypt.Service
	github     GitHubClient
	kubernetes KubernetesClient

	toolCatalog []ToolCapability
	now         func() time.Time
}

// Dependencies wires infrastructure for MCP domain service.
type Dependencies struct {
	Runs       agentrunrepo.Repository
	FlowEvents floweventrepo.Repository
	Repos      repocfgrepo.Repository
	TokenCrypt *tokencrypt.Service
	GitHub     GitHubClient
	Kubernetes KubernetesClient
}

// NewService creates MCP domain service.
func NewService(cfg Config, deps Dependencies) (*Service, error) {
	cfg.TokenSigningKey = strings.TrimSpace(cfg.TokenSigningKey)
	if cfg.TokenSigningKey == "" {
		return nil, fmt.Errorf("mcp token signing key is required")
	}
	cfg.TokenIssuer = strings.TrimSpace(cfg.TokenIssuer)
	if cfg.TokenIssuer == "" {
		cfg.TokenIssuer = defaultTokenIssuer
	}
	cfg.ServerName = strings.TrimSpace(cfg.ServerName)
	if cfg.ServerName == "" {
		cfg.ServerName = defaultServerName
	}
	cfg.InternalMCPBaseURL = strings.TrimSpace(cfg.InternalMCPBaseURL)
	if cfg.InternalMCPBaseURL == "" {
		cfg.InternalMCPBaseURL = defaultInternalMCPBaseURL
	}
	if cfg.DefaultTokenTTL <= 0 {
		cfg.DefaultTokenTTL = defaultTokenTTL
	}
	if cfg.DefaultTokenTTL < minTokenTTL {
		cfg.DefaultTokenTTL = minTokenTTL
	}
	if cfg.DefaultTokenTTL > maxTokenTTL {
		cfg.DefaultTokenTTL = maxTokenTTL
	}
	if deps.Runs == nil {
		return nil, fmt.Errorf("runs repository is required")
	}
	if deps.Repos == nil {
		return nil, fmt.Errorf("repositories repository is required")
	}
	if deps.TokenCrypt == nil {
		return nil, fmt.Errorf("token crypto service is required")
	}
	if deps.GitHub == nil {
		return nil, fmt.Errorf("github client is required")
	}
	if deps.Kubernetes == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}

	catalog := DefaultToolCatalog()
	sort.Slice(catalog, func(i, j int) bool { return catalog[i].Name < catalog[j].Name })

	return &Service{
		cfg:         cfg,
		runs:        deps.Runs,
		flowEvents:  deps.FlowEvents,
		repos:       deps.Repos,
		tokenCrypt:  deps.TokenCrypt,
		github:      deps.GitHub,
		kubernetes:  deps.Kubernetes,
		toolCatalog: catalog,
		now:         time.Now,
	}, nil
}

// ToolCatalog returns deterministic MCP tool catalog snapshot.
func (s *Service) ToolCatalog() []ToolCapability {
	out := make([]ToolCapability, len(s.toolCatalog))
	copy(out, s.toolCatalog)
	return out
}

// IssueRunToken mints one short-lived MCP token bound to a specific run.
func (s *Service) IssueRunToken(ctx context.Context, params IssueRunTokenParams) (IssuedToken, error) {
	runID := strings.TrimSpace(params.RunID)
	if runID == "" {
		return IssuedToken{}, fmt.Errorf("run_id is required")
	}
	runtimeMode := normalizeRuntimeMode(params.RuntimeMode)
	namespace := strings.TrimSpace(params.Namespace)
	if runtimeMode == agentdomain.RuntimeModeFullEnv && namespace == "" {
		return IssuedToken{}, fmt.Errorf("namespace is required for full-env")
	}

	run, ok, err := s.runs.GetByID(ctx, runID)
	if err != nil {
		return IssuedToken{}, fmt.Errorf("get run for token issue: %w", err)
	}
	if !ok {
		return IssuedToken{}, fmt.Errorf("run not found")
	}
	if !isRunActive(run.Status) {
		return IssuedToken{}, fmt.Errorf("run status %q is not active", run.Status)
	}

	ttl := params.TTL
	if ttl <= 0 {
		ttl = s.cfg.DefaultTokenTTL
	}
	if ttl < minTokenTTL {
		ttl = minTokenTTL
	}
	if ttl > maxTokenTTL {
		ttl = maxTokenTTL
	}

	now := s.now().UTC()
	expiresAt := now.Add(ttl)
	token, err := s.signRunToken(runTokenClaims{
		RunID:            run.ID,
		CorrelationID:    run.CorrelationID,
		ProjectID:        run.ProjectID,
		Namespace:        namespace,
		RuntimeMode:      string(runtimeMode),
		RegisteredClaims: runTokenRegisteredClaims(now, expiresAt, s.cfg.TokenIssuer, run.ID),
	})
	if err != nil {
		return IssuedToken{}, fmt.Errorf("sign run token: %w", err)
	}

	s.auditRunTokenIssued(ctx, run, namespace, runtimeMode, expiresAt)

	return IssuedToken{Token: token, ExpiresAt: expiresAt}, nil
}

// VerifyRunToken validates MCP bearer token and resolves bound session context.
func (s *Service) VerifyRunToken(ctx context.Context, rawToken string) (SessionContext, error) {
	claims, err := s.parseRunToken(strings.TrimSpace(rawToken))
	if err != nil {
		return SessionContext{}, err
	}

	run, ok, err := s.runs.GetByID(ctx, claims.RunID)
	if err != nil {
		return SessionContext{}, fmt.Errorf("get run for token verify: %w", err)
	}
	if !ok {
		return SessionContext{}, fmt.Errorf("run not found")
	}
	if !isRunActive(run.Status) {
		return SessionContext{}, fmt.Errorf("run status %q is not active", run.Status)
	}
	if run.CorrelationID != claims.CorrelationID {
		return SessionContext{}, fmt.Errorf("correlation mismatch")
	}
	if run.ProjectID != "" && claims.ProjectID != "" && run.ProjectID != claims.ProjectID {
		return SessionContext{}, fmt.Errorf("project mismatch")
	}

	return SessionContext{
		RunID:         claims.RunID,
		CorrelationID: claims.CorrelationID,
		ProjectID:     claims.ProjectID,
		Namespace:     claims.Namespace,
		RuntimeMode:   parseRuntimeMode(string(claims.RuntimeMode)),
		ExpiresAt:     claims.ExpiresAt,
	}, nil
}

func normalizeRuntimeMode(value agentdomain.RuntimeMode) agentdomain.RuntimeMode {
	if strings.EqualFold(strings.TrimSpace(string(value)), string(agentdomain.RuntimeModeFullEnv)) {
		return agentdomain.RuntimeModeFullEnv
	}
	return agentdomain.RuntimeModeCodeOnly
}

func parseRuntimeMode(value string) agentdomain.RuntimeMode {
	if strings.EqualFold(strings.TrimSpace(value), string(agentdomain.RuntimeModeFullEnv)) {
		return agentdomain.RuntimeModeFullEnv
	}
	return agentdomain.RuntimeModeCodeOnly
}

func isRunActive(status string) bool {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case string(rundomain.StatusPending), string(rundomain.StatusRunning):
		return true
	default:
		return false
	}
}

func (s *Service) auditRunTokenIssued(ctx context.Context, run agentrunrepo.Run, namespace string, runtimeMode agentdomain.RuntimeMode, expiresAt time.Time) {
	if s.flowEvents == nil {
		return
	}
	payload := encodeRunTokenIssuedEventPayload(runTokenIssuedEventPayload{
		RunID:       run.ID,
		ProjectID:   run.ProjectID,
		Namespace:   namespace,
		RuntimeMode: runtimeMode,
		ExpiresAt:   expiresAt.UTC().Format(time.RFC3339Nano),
	})
	_ = s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: run.CorrelationID,
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorIDControlPlaneMCP,
		EventType:     floweventdomain.EventTypeRunMCPTokenIssued,
		Payload:       payload,
		CreatedAt:     s.now().UTC(),
	})
}

type resolvedRunContext struct {
	Session    SessionContext
	Run        agentrunrepo.Run
	Repository repocfgrepo.RepositoryBinding
	Token      string
	Payload    runPayload
}

type runPayload struct {
	Project    runPayloadProject    `json:"project"`
	Repository runPayloadRepository `json:"repository"`
	Issue      *runPayloadIssue     `json:"issue,omitempty"`
	Trigger    *runPayloadTrigger   `json:"trigger,omitempty"`
}

type runPayloadProject struct {
	ID           string `json:"id"`
	RepositoryID string `json:"repository_id"`
	ServicesYAML string `json:"services_yaml"`
}

type runPayloadRepository struct {
	FullName string `json:"full_name"`
	Name     string `json:"name"`
}

type runPayloadIssue struct {
	Number  int64  `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
}

type runPayloadTrigger struct {
	Label string `json:"label"`
	Kind  string `json:"kind"`
}

func parseRunPayload(raw json.RawMessage) (runPayload, error) {
	if len(raw) == 0 {
		return runPayload{}, fmt.Errorf("run payload is empty")
	}
	var payload runPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return runPayload{}, fmt.Errorf("decode run payload: %w", err)
	}
	return payload, nil
}

func splitRepoFullName(fullName string) (owner string, name string) {
	parts := strings.Split(strings.TrimSpace(fullName), "/")
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func clampLimit(value int, def int, max int) int {
	if value <= 0 {
		return def
	}
	if value > max {
		return max
	}
	return value
}

func runTokenRegisteredClaims(issuedAt time.Time, expiresAt time.Time, issuer string, runID string) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   "run:" + strings.TrimSpace(runID),
		IssuedAt:  jwt.NewNumericDate(issuedAt.UTC()),
		NotBefore: jwt.NewNumericDate(issuedAt.UTC()),
		ExpiresAt: jwt.NewNumericDate(expiresAt.UTC()),
	}
}
