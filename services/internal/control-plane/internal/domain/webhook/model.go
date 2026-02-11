package webhook

import (
	"encoding/json"
	"time"
)

// IngestCommand is a normalized webhook command accepted by the domain service.
type IngestCommand struct {
	// CorrelationID is a deduplication key shared across flow records.
	CorrelationID string
	// EventType is a provider event name from request headers.
	EventType string
	// DeliveryID is a provider delivery identifier from webhook headers.
	DeliveryID string
	// ReceivedAt is the time when ingress received the webhook.
	ReceivedAt time.Time
	// Payload is raw webhook JSON body.
	Payload json.RawMessage
}

// IngestResult is a transport-facing outcome of webhook ingestion.
type IngestResult struct {
	// CorrelationID mirrors the request correlation id.
	CorrelationID string `json:"correlation_id"`
	// RunID references an agent run linked to this webhook.
	RunID string `json:"run_id,omitempty"`
	// Status is accepted, duplicate, or ignored.
	Status string `json:"status"`
	// Duplicate is true when webhook was already processed before.
	Duplicate bool `json:"duplicate"`
}

// TriggerLabels defines active issue labels that create development runs.
type TriggerLabels struct {
	RunDev       string
	RunDevRevise string
}

func defaultTriggerLabels() TriggerLabels {
	return TriggerLabels{
		RunDev:       "run:dev",
		RunDevRevise: "run:dev:revise",
	}
}

type issueRunTrigger struct {
	Label string
	Kind  string
}

// githubWebhookEnvelope is a local transport DTO for fields used by the domain.
// It is intentionally minimal and independent from provider SDK structs.
type githubWebhookEnvelope struct {
	Action       string                   `json:"action"`
	Installation githubInstallationRecord `json:"installation"`
	Repository   githubRepositoryRecord   `json:"repository"`
	Issue        githubIssueRecord        `json:"issue"`
	Label        githubLabelRecord        `json:"label"`
	Sender       githubActorRecord        `json:"sender"`
}

type githubInstallationRecord struct {
	ID int64 `json:"id"`
}

type githubRepositoryRecord struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
	Name     string `json:"name"`
	Private  bool   `json:"private"`
}

type githubIssueRecord struct {
	ID          int64                 `json:"id"`
	Number      int64                 `json:"number"`
	Title       string                `json:"title"`
	HTMLURL     string                `json:"html_url"`
	State       string                `json:"state"`
	User        githubActorRecord     `json:"user"`
	PullRequest *githubPullRequestRef `json:"pull_request"`
}

type githubLabelRecord struct {
	Name string `json:"name"`
}

type githubActorRecord struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

type githubPullRequestRef struct {
	URL     string `json:"url"`
	HTMLURL string `json:"html_url"`
}
