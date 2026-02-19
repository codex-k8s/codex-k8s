package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	cpclient "github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/controlplane"
)

// ExitError allows caller to map runner failures to process exit code.
type ExitError struct {
	ExitCode int
	Err      error
}

func (e ExitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("runner failed with exit code %d", e.ExitCode)
	}
	return e.Err.Error()
}

func (e ExitError) Unwrap() error {
	return e.Err
}

// PromptConfig defines per-run prompt rendering and model settings.
type PromptConfig struct {
	TriggerKind          string
	TriggerLabel         string
	PromptTemplateKind   string
	PromptTemplateSource string
	PromptTemplateLocale string
	StateInReviewLabel   string
	AgentModel           string
	AgentReasoningEffort string
	AgentBaseBranch      string
	AgentDisplayName     string
}

// GitBotConfig defines git transport credentials for runner pod.
type GitBotConfig struct {
	GitBotToken    string
	GitBotUsername string
	GitBotMail     string
}

// OpenAIConfig defines codex-cli authentication inputs.
type OpenAIConfig struct {
	OpenAIAPIKey string
}

// Config defines runtime parameters for one agent-runner job.
type Config struct {
	RunID              string
	CorrelationID      string
	ProjectID          string
	RepositoryFullName string
	AgentKey           string
	IssueNumber        int64
	RunTargetBranch    string
	ExistingPRNumber   int
	RuntimeMode        string

	PromptConfig

	ControlPlaneGRPCTarget string
	MCPBaseURL             string
	MCPBearerToken         string
	RunAccessKey           string

	GitBotConfig
	OpenAIConfig
}

// ControlPlaneCallbacks defines required control-plane callbacks for runner lifecycle.
type ControlPlaneCallbacks interface {
	UpsertAgentSession(ctx context.Context, params cpclient.AgentSessionUpsertParams) error
	GetLatestAgentSession(ctx context.Context, query cpclient.LatestAgentSessionQuery) (cpclient.AgentSessionSnapshot, bool, error)
	InsertRunFlowEvent(ctx context.Context, runID string, eventType floweventdomain.EventType, payload json.RawMessage) error
	GetCodexAuth(ctx context.Context) ([]byte, bool, error)
	UpsertCodexAuth(ctx context.Context, authJSON []byte) error
	UpsertRunStatusComment(ctx context.Context, params cpclient.UpsertRunStatusCommentParams) error
}

// Service runs one codex-driven development/revise cycle.
type Service struct {
	cfg    Config
	cp     ControlPlaneCallbacks
	logger *slog.Logger
}

// NewService creates runner service.
func NewService(cfg Config, cp ControlPlaneCallbacks, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{cfg: cfg, cp: cp, logger: logger}
}

type runResult struct {
	targetBranch        string
	triggerKind         string
	templateKind        string
	restoredSessionPath string
	sessionFilePath     string
	sessionID           string
	existingPRNumber    int
	prNumber            int
	prURL               string
	report              codexReport
	codexExecOutput     string
	gitPushOutput       string
	toolGaps            []string
}

type restoredSession struct {
	restoredSessionPath string
	existingPRNumber    int
	prNotFound          bool
}

type codexState struct {
	homeDir      string
	codexDir     string
	sessionsDir  string
	workspaceDir string
	repoDir      string
}

type codexReport struct {
	Summary         string   `json:"summary"`
	Branch          string   `json:"branch"`
	PRNumber        int      `json:"pr_number"`
	PRURL           string   `json:"pr_url"`
	SessionID       string   `json:"session_id"`
	Model           string   `json:"model"`
	ReasoningEffort string   `json:"reasoning_effort"`
	Diagnosis       string   `json:"diagnosis,omitempty"`
	ActionItems     []string `json:"action_items,omitempty"`
	EvidenceRefs    []string `json:"evidence_refs,omitempty"`
	ToolGaps        []string `json:"tool_gaps,omitempty"`
}

type promptTaskTemplateData struct {
	BaseBranch   string
	PromptLocale string
}

type promptEnvelopeTemplateData struct {
	RepositoryFullName string
	RunID              string
	IssueNumber        int64
	AgentKey           string
	RuntimeMode        string
	IsFullEnv          bool
	TargetBranch       string
	BaseBranch         string
	TriggerKind        string
	IsReviseTrigger    bool
	HasExistingPR      bool
	ExistingPRNumber   int
	TriggerLabel       string
	StateInReviewLabel string
	HasContext7        bool
	HasRunAccessKey    bool
	RunAccessKey       string
	PromptLocale       string
	TaskBody           string
}

type codexConfigTemplateData struct {
	Model           string
	ReasoningEffort string
	MCPBaseURL      string
	HasContext7     bool
}

type kubectlKubeconfigTemplateData struct {
	CACertificatePath string
	ServiceHost       string
	ServicePort       string
	ClusterName       string
	Namespace         string
	UserName          string
	ContextName       string
	Token             string
}

type sessionLogSnapshot struct {
	Version string                  `json:"version"`
	Status  string                  `json:"status"`
	Report  codexReport             `json:"report"`
	Runtime sessionRuntimeLogFields `json:"runtime"`
}

type sessionRuntimeLogFields struct {
	TargetBranch     string `json:"target_branch"`
	CodexExecOutput  string `json:"codex_exec_output,omitempty"`
	GitPushOutput    string `json:"git_push_output,omitempty"`
	ExistingPRNumber int    `json:"existing_pr_number,omitempty"`
}

type selfImproveDiagnosisReadyPayload struct {
	Diagnosis    string   `json:"diagnosis"`
	ActionItems  []string `json:"action_items"`
	EvidenceRefs []string `json:"evidence_refs"`
	ToolGaps     []string `json:"tool_gaps,omitempty"`
}

type toolchainGapDetectedPayload struct {
	ToolGaps             []string `json:"tool_gaps"`
	Sources              []string `json:"sources"`
	SuggestedUpdatePaths []string `json:"suggested_update_paths"`
}
