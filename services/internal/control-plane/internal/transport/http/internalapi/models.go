package internalapi

import (
	"encoding/json"
	"time"
)

const sessionStatusRunning = "running"

type sessionUpsertRequest struct {
	RunID              string          `json:"run_id"`
	CorrelationID      string          `json:"correlation_id,omitempty"`
	ProjectID          string          `json:"project_id,omitempty"`
	RepositoryFullName string          `json:"repository_full_name"`
	IssueNumber        *int            `json:"issue_number,omitempty"`
	BranchName         string          `json:"branch_name"`
	PRNumber           *int            `json:"pr_number,omitempty"`
	PRURL              string          `json:"pr_url,omitempty"`
	TriggerKind        string          `json:"trigger_kind,omitempty"`
	TemplateKind       string          `json:"template_kind,omitempty"`
	TemplateSource     string          `json:"template_source,omitempty"`
	TemplateLocale     string          `json:"template_locale,omitempty"`
	Model              string          `json:"model,omitempty"`
	ReasoningEffort    string          `json:"reasoning_effort,omitempty"`
	Status             string          `json:"status,omitempty"`
	SessionID          string          `json:"session_id,omitempty"`
	SessionJSON        json.RawMessage `json:"session_json,omitempty"`
	CodexSessionPath   string          `json:"codex_cli_session_path,omitempty"`
	CodexSessionJSON   json.RawMessage `json:"codex_cli_session_json,omitempty"`
	StartedAt          *time.Time      `json:"started_at,omitempty"`
	FinishedAt         *time.Time      `json:"finished_at,omitempty"`
}

type latestSessionResponse struct {
	Found   bool                `json:"found"`
	Session *sessionSnapshotDTO `json:"session,omitempty"`
}

type sessionSnapshotDTO struct {
	RunID              string          `json:"run_id"`
	CorrelationID      string          `json:"correlation_id"`
	ProjectID          string          `json:"project_id,omitempty"`
	RepositoryFullName string          `json:"repository_full_name"`
	IssueNumber        int             `json:"issue_number,omitempty"`
	BranchName         string          `json:"branch_name"`
	PRNumber           int             `json:"pr_number,omitempty"`
	PRURL              string          `json:"pr_url,omitempty"`
	TriggerKind        string          `json:"trigger_kind,omitempty"`
	TemplateKind       string          `json:"template_kind,omitempty"`
	TemplateSource     string          `json:"template_source,omitempty"`
	TemplateLocale     string          `json:"template_locale,omitempty"`
	Model              string          `json:"model,omitempty"`
	ReasoningEffort    string          `json:"reasoning_effort,omitempty"`
	Status             string          `json:"status"`
	SessionID          string          `json:"session_id,omitempty"`
	SessionJSON        json.RawMessage `json:"session_json,omitempty"`
	CodexSessionPath   string          `json:"codex_cli_session_path,omitempty"`
	CodexSessionJSON   json.RawMessage `json:"codex_cli_session_json,omitempty"`
	StartedAt          time.Time       `json:"started_at"`
	FinishedAt         *time.Time      `json:"finished_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type eventRequest struct {
	RunID     string          `json:"run_id,omitempty"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type sessionUpsertResponse struct {
	OK    bool   `json:"ok"`
	RunID string `json:"run_id"`
}

type eventInsertResponse struct {
	OK        bool   `json:"ok"`
	EventType string `json:"event_type"`
}
