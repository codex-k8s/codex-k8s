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
	ActorTypeHuman  ActorType = "human"
)

// ActorID is a concrete source identifier within ActorType.
type ActorID string

const (
	ActorIDGitHubWebhook   ActorID = "github-webhook"
	ActorIDWorker          ActorID = "worker"
	ActorIDControlPlane    ActorID = "control-plane"
	ActorIDControlPlaneMCP ActorID = "control-plane-mcp"
	ActorIDAgentRunner     ActorID = "agent-runner"
)

// EventType is a normalized lifecycle event name.
type EventType string

const (
	EventTypeWebhookReceived  EventType = "webhook.received"
	EventTypeWebhookDuplicate EventType = "webhook.duplicate"
	EventTypeWebhookIgnored   EventType = "webhook.ignored"
)

const (
	EventTypeRunNamespacePrepared       EventType = "run.namespace.prepared"
	EventTypeRunNamespaceTTLScheduled   EventType = "run.namespace.ttl_scheduled"
	EventTypeRunNamespaceTTLExtended    EventType = "run.namespace.ttl_extended"
	EventTypeRunNamespaceCleaned        EventType = "run.namespace.cleaned"
	EventTypeRunNamespaceCleanupFailed  EventType = "run.namespace.cleanup_failed"
	EventTypeRunNamespaceCleanupSkipped EventType = "run.namespace.cleanup_skipped"
	EventTypeRunStarted                 EventType = "run.started"
	EventTypeRunSucceeded               EventType = "run.succeeded"
	EventTypeRunCanceled                EventType = "run.canceled"
	EventTypeRunFailed                  EventType = "run.failed"
	EventTypeRunFailedJobNotFound       EventType = "run.failed.job_not_found"
	EventTypeRunFailedLaunchError       EventType = "run.failed.launch_error"
	EventTypeRunFailedPrecondition      EventType = "run.failed.precondition"
)

const (
	EventTypeRunMCPTokenIssued            EventType = "run.mcp.token.issued"
	EventTypeRunAgentStarted              EventType = "run.agent.started"
	EventTypeRunAgentStatusReported       EventType = "run.agent.status_reported"
	EventTypeRunAgentSessionRestored      EventType = "run.agent.session.restored"
	EventTypeRunAgentSessionSaved         EventType = "run.agent.session.saved"
	EventTypeRunAgentResumeUsed           EventType = "run.agent.resume.used"
	EventTypeRunCodexAuthRequired         EventType = "run.codex.auth.required"
	EventTypeRunCodexAuthSynchronized     EventType = "run.codex.auth.synchronized"
	EventTypeRunPRCreated                 EventType = "run.pr.created"
	EventTypeRunPRUpdated                 EventType = "run.pr.updated"
	EventTypeRunRevisePRNotFound          EventType = "run.revise.pr_not_found"
	EventTypeRunSelfImproveDiagnosisReady EventType = "run.self_improve.diagnosis_ready"
	EventTypeRunToolchainGapDetected      EventType = "run.toolchain.gap_detected"
	EventTypeRunStatusCommentUpserted     EventType = "run.status.comment.upserted"
	EventTypeRunTriggerConflictComment    EventType = "run.trigger.conflict.comment"
	EventTypeRunNamespaceDeleteByStaff    EventType = "run.namespace.delete_by_staff"
	EventTypeRunNamespaceDeleteBySystem   EventType = "run.namespace.delete_by_system"
)

const (
	EventTypePromptContextAssembled EventType = "prompt.context.assembled"
	EventTypeMCPToolCalled          EventType = "mcp.tool.called"
	EventTypeMCPToolSucceeded       EventType = "mcp.tool.succeeded"
	EventTypeMCPToolFailed          EventType = "mcp.tool.failed"
	EventTypeMCPToolApprovalPending EventType = "mcp.tool.approval_pending"
	EventTypeApprovalRequested      EventType = "approval.requested"
	EventTypeApprovalApproved       EventType = "approval.approved"
	EventTypeApprovalDenied         EventType = "approval.denied"
	EventTypeApprovalExpired        EventType = "approval.expired"
	EventTypeApprovalFailed         EventType = "approval.failed"
	EventTypeApprovalApplied        EventType = "approval.applied"
	EventTypeRunWaitPaused          EventType = "run.wait.paused"
	EventTypeRunWaitResumed         EventType = "run.wait.resumed"
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
