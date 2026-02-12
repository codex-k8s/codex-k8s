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
	ActorTypeAgent  ActorType = "agent"
)

// ActorID is a concrete source identifier within ActorType.
type ActorID string

const (
	ActorIDGitHubWebhook   ActorID = "github-webhook"
	ActorIDWorker          ActorID = "worker"
	ActorIDControlPlaneMCP ActorID = "control-plane-mcp"
	ActorIDAgentRunner     ActorID = "agent-runner"
)

// EventType is a normalized lifecycle event name.
type EventType string

const (
	EventTypeWebhookReceived           EventType = "webhook.received"
	EventTypeWebhookDuplicate          EventType = "webhook.duplicate"
	EventTypeWebhookIgnored            EventType = "webhook.ignored"
	EventTypeRunNamespacePrepared      EventType = "run.namespace.prepared"
	EventTypeRunNamespaceCleaned       EventType = "run.namespace.cleaned"
	EventTypeRunNamespaceCleanupFailed EventType = "run.namespace.cleanup_failed"
	EventTypeRunStarted                EventType = "run.started"
	EventTypeRunSucceeded              EventType = "run.succeeded"
	EventTypeRunFailed                 EventType = "run.failed"
	EventTypeRunFailedJobNotFound      EventType = "run.failed.job_not_found"
	EventTypeRunFailedLaunchError      EventType = "run.failed.launch_error"
	EventTypeRunFailedPrecondition     EventType = "run.failed.precondition"
	EventTypeRunMCPTokenIssued         EventType = "run.mcp.token.issued"
	EventTypeRunAgentStarted           EventType = "run.agent.started"
	EventTypeRunAgentSessionRestored   EventType = "run.agent.session.restored"
	EventTypeRunAgentSessionSaved      EventType = "run.agent.session.saved"
	EventTypeRunAgentResumeUsed        EventType = "run.agent.resume.used"
	EventTypeRunPRCreated              EventType = "run.pr.created"
	EventTypeRunPRUpdated              EventType = "run.pr.updated"
	EventTypeRunRevisePRNotFound       EventType = "run.revise.pr_not_found"
	EventTypePromptContextAssembled    EventType = "prompt.context.assembled"
	EventTypeMCPToolCalled             EventType = "mcp.tool.called"
	EventTypeMCPToolSucceeded          EventType = "mcp.tool.succeeded"
	EventTypeMCPToolFailed             EventType = "mcp.tool.failed"
	EventTypeMCPToolApprovalPending    EventType = "mcp.tool.approval_pending"
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
