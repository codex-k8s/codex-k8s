package value

import (
	"time"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

// MissionControlEntityRef identifies one public entity reference used by commands and read-model responses.
type MissionControlEntityRef struct {
	EntityKind     enumtypes.MissionControlEntityKind `json:"entity_kind"`
	EntityPublicID string                             `json:"entity_public_id"`
}

// MissionControlRelationView exposes one relation via public entity refs only.
type MissionControlRelationView struct {
	RelationKind    enumtypes.MissionControlRelationKind       `json:"relation_kind"`
	SourceKind      enumtypes.MissionControlRelationSourceKind `json:"source_kind"`
	SourceEntityRef MissionControlEntityRef                    `json:"source_entity_ref"`
	TargetEntityRef MissionControlEntityRef                    `json:"target_entity_ref"`
}

// MissionControlDiscussionCreatePayload captures typed input for discussion.create.
type MissionControlDiscussionCreatePayload struct {
	Title           string                   `json:"title"`
	BodyMarkdown    string                   `json:"body_markdown,omitempty"`
	ParentEntityRef *MissionControlEntityRef `json:"parent_entity_ref,omitempty"`
}

// MissionControlWorkItemCreatePayload captures typed input for work_item.create.
type MissionControlWorkItemCreatePayload struct {
	Title             string                    `json:"title"`
	BodyMarkdown      string                    `json:"body_markdown,omitempty"`
	InitialLabels     []string                  `json:"initial_labels,omitempty"`
	RelatedEntityRefs []MissionControlEntityRef `json:"related_entity_refs,omitempty"`
}

// MissionControlDiscussionFormalizePayload captures typed input for discussion.formalize.
type MissionControlDiscussionFormalizePayload struct {
	SourceEntityRef MissionControlEntityRef `json:"source_entity_ref"`
	FormalizedKind  string                  `json:"formalized_kind"`
	Title           string                  `json:"title"`
	BodyMarkdown    string                  `json:"body_markdown,omitempty"`
}

// MissionControlStageNextStepExecutePayload captures typed input for stage.next_step.execute.
type MissionControlStageNextStepExecutePayload struct {
	ThreadKind          string                                      `json:"thread_kind"`
	ThreadNumber        int                                         `json:"thread_number"`
	TargetLabel         string                                      `json:"target_label"`
	RemovedLabels       []string                                    `json:"removed_labels,omitempty"`
	DisplayVariant      string                                      `json:"display_variant,omitempty"`
	ApprovalRequirement enumtypes.MissionControlApprovalRequirement `json:"approval_requirement"`
}

// MissionControlRetrySyncPayload captures typed input for command.retry_sync.
type MissionControlRetrySyncPayload struct {
	CommandID      string                                `json:"command_id"`
	RetryReason    string                                `json:"retry_reason,omitempty"`
	ExpectedStatus enumtypes.MissionControlCommandStatus `json:"expected_status,omitempty"`
}

// MissionControlCommandPayload is the closed domain union for Mission Control command payloads.
type MissionControlCommandPayload struct {
	DiscussionCreate    *MissionControlDiscussionCreatePayload     `json:"discussion_create,omitempty"`
	WorkItemCreate      *MissionControlWorkItemCreatePayload       `json:"work_item_create,omitempty"`
	DiscussionFormalize *MissionControlDiscussionFormalizePayload  `json:"discussion_formalize,omitempty"`
	StageNextStep       *MissionControlStageNextStepExecutePayload `json:"stage_next_step,omitempty"`
	RetrySync           *MissionControlRetrySyncPayload            `json:"retry_sync,omitempty"`
}

// MissionControlApprovalSnapshot captures approval metadata returned by command queries.
type MissionControlApprovalSnapshot struct {
	ApprovalState     enumtypes.MissionControlApprovalState `json:"approval_state"`
	ApprovalRequestID string                                `json:"approval_request_id,omitempty"`
	RequestedAt       *time.Time                            `json:"requested_at,omitempty"`
	DecidedAt         *time.Time                            `json:"decided_at,omitempty"`
	ApproverActorID   string                                `json:"approver_actor_id,omitempty"`
}

// MissionControlCommandResultPayload captures typed status metadata stored in result_payload.
type MissionControlCommandResultPayload struct {
	StatusMessage       string                          `json:"status_message,omitempty"`
	EntityRefs          []MissionControlEntityRef       `json:"entity_refs,omitempty"`
	Approval            *MissionControlApprovalSnapshot `json:"approval,omitempty"`
	ProviderDeliveryIDs []string                        `json:"provider_delivery_ids,omitempty"`
}

// MissionControlActiveSet contains one read-model snapshot slice for future transport adapters.
type MissionControlActiveSet struct {
	Entities  []entitytypes.MissionControlEntity `json:"entities"`
	Relations []MissionControlRelationView       `json:"relations"`
}

// MissionControlEntityDetails contains one entity details payload without transport concerns.
type MissionControlEntityDetails struct {
	Entity    entitytypes.MissionControlEntity          `json:"entity"`
	Relations []MissionControlRelationView              `json:"relations"`
	Timeline  []entitytypes.MissionControlTimelineEntry `json:"timeline"`
}

// MissionControlCommandAdmission describes the immediate result of command admission.
type MissionControlCommandAdmission struct {
	Command      entitytypes.MissionControlCommand `json:"command"`
	TargetEntity *entitytypes.MissionControlEntity `json:"target_entity,omitempty"`
	EntityRefs   []MissionControlEntityRef         `json:"entity_refs,omitempty"`
	Approval     *MissionControlApprovalSnapshot   `json:"approval,omitempty"`
}

// MissionControlCommandStatusView is the typed domain view used by future transport/status lookups.
type MissionControlCommandStatusView struct {
	Command             entitytypes.MissionControlCommand `json:"command"`
	EntityRefs          []MissionControlEntityRef         `json:"entity_refs,omitempty"`
	Approval            *MissionControlApprovalSnapshot   `json:"approval,omitempty"`
	StatusMessage       string                            `json:"status_message,omitempty"`
	ProviderDeliveryIDs []string                          `json:"provider_delivery_ids,omitempty"`
}
