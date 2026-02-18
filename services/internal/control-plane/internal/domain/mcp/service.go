package mcp

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	agentsessionrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	mcpactionrequestrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/mcpactionrequest"
	platformtokenrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/platformtoken"
	projectdatabaserepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/projectdatabase"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
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
	TokenSigningKey              string
	TokenIssuer                  string
	ServerName                   string
	PublicBaseURL                string
	InternalMCPBaseURL           string
	RepositoryRoot               string
	ServicesConfigEnv            string
	DefaultTokenTTL              time.Duration
	DatabaseLifecycleAllowedEnvs []string
}

// GitHubClient defines GitHub operations used by MCP tools.
type GitHubClient interface {
	GetIssue(ctx context.Context, params GitHubGetIssueParams) (GitHubIssue, error)
	GetPullRequest(ctx context.Context, params GitHubGetPullRequestParams) (GitHubPullRequest, error)
	GetAuthenticatedUserLogin(ctx context.Context, token string) (string, error)
	ListIssueComments(ctx context.Context, params GitHubListIssueCommentsParams) ([]GitHubIssueComment, error)
	ListIssueLabels(ctx context.Context, params GitHubListIssueLabelsParams) ([]GitHubLabel, error)
	ListBranches(ctx context.Context, params GitHubListBranchesParams) ([]GitHubBranch, error)
	EnsureBranch(ctx context.Context, params GitHubEnsureBranchParams) (GitHubBranch, error)
	UpsertPullRequest(ctx context.Context, params GitHubUpsertPullRequestParams) (GitHubPullRequest, error)
	CreateIssueComment(ctx context.Context, params GitHubCreateIssueCommentParams) (GitHubIssueComment, error)
	AddLabels(ctx context.Context, params GitHubMutateLabelsParams) ([]GitHubLabel, error)
	RemoveLabels(ctx context.Context, params GitHubMutateLabelsParams) ([]GitHubLabel, error)
	UpsertRepositorySecret(ctx context.Context, params GitHubUpsertRepositorySecretParams) error
}

// KubernetesClient defines Kubernetes operations used by MCP tools.
type KubernetesClient interface {
	ListPods(ctx context.Context, namespace string, limit int) ([]KubernetesPod, error)
	ListEvents(ctx context.Context, namespace string, limit int) ([]KubernetesEvent, error)
	ListResources(ctx context.Context, namespace string, kind KubernetesResourceKind, limit int) ([]KubernetesResourceRef, error)
	GetPodLogs(ctx context.Context, namespace string, pod string, container string, tailLines int64) (string, error)
	ExecPod(ctx context.Context, namespace string, pod string, container string, command []string) (KubernetesExecResult, error)
	UpsertSecret(ctx context.Context, namespace string, secretName string, data map[string][]byte) error
}

// DatabaseClient defines database lifecycle operations used by MCP control tools.
type DatabaseClient interface {
	EnsureDatabase(ctx context.Context, databaseName string) (bool, error)
	DropDatabase(ctx context.Context, databaseName string) (bool, error)
	DatabaseExists(ctx context.Context, databaseName string) (bool, error)
}

// Service provides MCP token handling, prompt context building and tool operations.
type Service struct {
	cfg Config

	runs                         agentrunrepo.Repository
	flowEvents                   floweventrepo.Repository
	repos                        repocfgrepo.Repository
	platform                     platformtokenrepo.Repository
	actions                      mcpactionrequestrepo.Repository
	sessions                     agentsessionrepo.Repository
	projectDatabases             projectdatabaserepo.Repository
	tokenCrypt                   *tokencrypt.Service
	github                       GitHubClient
	kubernetes                   KubernetesClient
	database                     DatabaseClient
	databaseLifecycleAllowedEnvs map[string]struct{}

	toolCatalog []ToolCapability
	now         func() time.Time
}

// Dependencies wires infrastructure for MCP domain service.
type Dependencies struct {
	Runs             agentrunrepo.Repository
	FlowEvents       floweventrepo.Repository
	Repos            repocfgrepo.Repository
	Platform         platformtokenrepo.Repository
	Actions          mcpactionrequestrepo.Repository
	Sessions         agentsessionrepo.Repository
	ProjectDatabases projectdatabaserepo.Repository
	TokenCrypt       *tokencrypt.Service
	GitHub           GitHubClient
	Kubernetes       KubernetesClient
	Database         DatabaseClient
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
	cfg.RepositoryRoot = strings.TrimSpace(cfg.RepositoryRoot)
	if cfg.RepositoryRoot == "" {
		cfg.RepositoryRoot = "."
	}
	cfg.ServicesConfigEnv = strings.TrimSpace(cfg.ServicesConfigEnv)
	if cfg.ServicesConfigEnv == "" {
		cfg.ServicesConfigEnv = "production"
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
	if deps.Platform == nil {
		return nil, fmt.Errorf("platform token repository is required")
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
	if deps.Actions == nil {
		return nil, fmt.Errorf("mcp action requests repository is required")
	}
	if deps.Sessions == nil {
		return nil, fmt.Errorf("agent sessions repository is required")
	}
	if deps.ProjectDatabases == nil {
		return nil, fmt.Errorf("project database repository is required")
	}
	if deps.Database == nil {
		return nil, fmt.Errorf("database client is required")
	}

	databaseAllowedEnvs := normalizeDatabaseLifecycleAllowedEnvs(cfg.DatabaseLifecycleAllowedEnvs)

	catalog := DefaultToolCatalog()
	sort.Slice(catalog, func(i, j int) bool { return catalog[i].Name < catalog[j].Name })

	return &Service{
		cfg:                          cfg,
		runs:                         deps.Runs,
		flowEvents:                   deps.FlowEvents,
		repos:                        deps.Repos,
		platform:                     deps.Platform,
		actions:                      deps.Actions,
		sessions:                     deps.Sessions,
		projectDatabases:             deps.ProjectDatabases,
		tokenCrypt:                   deps.TokenCrypt,
		github:                       deps.GitHub,
		kubernetes:                   deps.Kubernetes,
		database:                     deps.Database,
		databaseLifecycleAllowedEnvs: databaseAllowedEnvs,
		toolCatalog:                  catalog,
		now:                          time.Now,
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
	normalized := strings.ToLower(strings.TrimSpace(status))
	return normalized == "pending" || normalized == "running"
}

func normalizeApprovalMode(mode entitytypes.MCPApprovalMode) entitytypes.MCPApprovalMode {
	switch mode {
	case entitytypes.MCPApprovalModeNone, entitytypes.MCPApprovalModeOwner, entitytypes.MCPApprovalModeDelegated:
		return mode
	default:
		return entitytypes.MCPApprovalModeOwner
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
