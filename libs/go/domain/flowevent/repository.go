package flowevent

import (
	"context"
	"encoding/json"
	"time"
)

// ActorType describes event source identity class.
type ActorType string

const (
	ActorTypeSystem ActorType = "system"
)

// ActorID is a concrete source identifier within ActorType.
type ActorID string

const (
	ActorIDGitHubWebhook ActorID = "github-webhook"
)

// EventType is a normalized lifecycle event name.
type EventType string

const (
	EventTypeWebhookReceived      EventType = "webhook.received"
	EventTypeWebhookDuplicate     EventType = "webhook.duplicate"
	EventTypeWebhookIgnored       EventType = "webhook.ignored"
	EventTypeRunStarted           EventType = "run.started"
	EventTypeRunSucceeded         EventType = "run.succeeded"
	EventTypeRunFailed            EventType = "run.failed"
	EventTypeRunFailedJobNotFound EventType = "run.failed.job_not_found"
	EventTypeRunFailedLaunchError EventType = "run.failed.launch_error"
)

// InsertParams defines a single flow event record.
type InsertParams struct {
	CorrelationID string
	ActorType     ActorType
	ActorID       ActorID
	EventType     EventType
	Payload       json.RawMessage
	CreatedAt     time.Time
}

// Repository persists flow events.
type Repository interface {
	Insert(ctx context.Context, params InsertParams) error
}
