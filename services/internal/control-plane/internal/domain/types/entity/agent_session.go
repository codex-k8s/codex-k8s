package entity

import (
	"encoding/json"
	"time"
)

// AgentSession stores resumable codex-cli snapshot and run-level execution metadata.
type AgentSession struct {
	ID                 int64
	RunID              string
	CorrelationID      string
	ProjectID          string
	RepositoryFullName string
	IssueNumber        int
	BranchName         string
	PRNumber           int
	PRURL              string
	TriggerKind        string
	TemplateKind       string
	TemplateSource     string
	TemplateLocale     string
	Model              string
	ReasoningEffort    string
	Status             string
	SessionID          string
	SessionJSON        json.RawMessage
	CodexSessionPath   string
	CodexSessionJSON   json.RawMessage
	StartedAt          time.Time
	FinishedAt         time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
