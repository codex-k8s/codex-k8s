package mcp

type controlAction string

const (
	controlActionSecretSyncEnv    controlAction = "secret_sync_env"
	controlActionDatabaseCreate   controlAction = "database_create"
	controlActionDatabaseDelete   controlAction = "database_delete"
	controlActionDatabaseDescribe controlAction = "database_describe"
	controlActionOwnerFeedback    controlAction = "owner_feedback_request"
	controlActionSecretDefaultKey               = "value"
)

type waitState string

const (
	waitStateNone waitState = ""
	waitStateMCP  waitState = "mcp"
)

const (
	triggerLabelRunDev       = "run:dev"
	triggerLabelRunDevRevise = "run:dev:revise"
	triggerLabelRunOps       = "run:ops"
	triggerLabelRunSelfPatch = "run:self-improve"
)

const (
	agentKeyDev = "dev"
	agentKeySRE = "sre"
)

const (
	controlToolMessageApplied               = "applied"
	controlToolMessageDryRun                = "dry_run"
	controlToolMessageApprovalRequired      = "approval_required"
	controlToolMessageIdempotentReplay      = "idempotent_replay"
	controlToolMessageDescribed             = "described"
	controlToolMessageDeleteConfirmRequired = "delete_confirmation_required"
)
