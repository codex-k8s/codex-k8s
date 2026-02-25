package entity

import "time"

// PromptTemplateKeyItem is one row in staff template-key list.
type PromptTemplateKeyItem struct {
	TemplateKey   string
	Scope         string
	ProjectID     string
	Role          string
	Kind          string
	Locale        string
	ActiveVersion int
	UpdatedAt     time.Time
}

// PromptTemplateVersion stores one persisted template version row.
type PromptTemplateVersion struct {
	TemplateKey       string
	Version           int
	Status            string
	Source            string
	Checksum          string
	BodyMarkdown      string
	ChangeReason      string
	SupersedesVersion *int
	UpdatedBy         string
	UpdatedAt         time.Time
	ActivatedAt       *time.Time
}

// PromptTemplateSeedSyncItem describes one dry-run/apply sync action.
type PromptTemplateSeedSyncItem struct {
	TemplateKey string
	Action      string
	Checksum    string
	Reason      string
}

// PromptTemplateSeedSyncResult aggregates seed sync counters and actions.
type PromptTemplateSeedSyncResult struct {
	CreatedCount int
	UpdatedCount int
	SkippedCount int
	Items        []PromptTemplateSeedSyncItem
}

// PromptTemplateAuditEvent stores one flow_events row projected to prompt-template audit.
type PromptTemplateAuditEvent struct {
	ID            int64
	CorrelationID string
	ProjectID     string
	TemplateKey   string
	Version       *int
	ActorID       string
	EventType     string
	PayloadJSON   string
	CreatedAt     time.Time
}
