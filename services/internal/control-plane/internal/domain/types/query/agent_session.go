package query

import (
	"encoding/json"
	"time"
)

// AgentSessionUpsertParams defines agent session snapshot persistence payload.
type AgentSessionUpsertParams struct {
	RunID              string
	CorrelationID      string
	ProjectID          string
	RepositoryFullName string
	IssueNumber        *int
	BranchName         string
	PRNumber           *int
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
	FinishedAt         *time.Time
}
