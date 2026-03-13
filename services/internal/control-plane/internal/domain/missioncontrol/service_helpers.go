package missioncontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	missioncontrolrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/missioncontrol"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

const (
	eventTypeMissionControlWarmupRequested    floweventdomain.EventType = "mission_control.warmup.summary_requested"
	eventTypeMissionControlEntityUpserted     floweventdomain.EventType = "mission_control.entity.upserted"
	eventTypeMissionControlEntityUpdated      floweventdomain.EventType = "mission_control.entity.projection_updated"
	eventTypeMissionControlRelationsReplaced  floweventdomain.EventType = "mission_control.relations.replaced"
	eventTypeMissionControlTimelineUpserted   floweventdomain.EventType = "mission_control.timeline.upserted"
	eventTypeMissionControlCommandAccepted    floweventdomain.EventType = "mission_control.command.accepted"
	eventTypeMissionControlCommandDeduped     floweventdomain.EventType = "mission_control.command.deduped"
	eventTypeMissionControlCommandQueued      floweventdomain.EventType = "mission_control.command.queued"
	eventTypeMissionControlCommandPendingSync floweventdomain.EventType = "mission_control.command.pending_sync"
	eventTypeMissionControlCommandReconciled  floweventdomain.EventType = "mission_control.command.reconciled"
	eventTypeMissionControlCommandFailed      floweventdomain.EventType = "mission_control.command.failed"
	eventTypeMissionControlCommandBlocked     floweventdomain.EventType = "mission_control.command.blocked"
	eventTypeMissionControlCommandCancelled   floweventdomain.EventType = "mission_control.command.cancelled"
)

type warmupEventPayload struct {
	ProjectID            string `json:"project_id"`
	RequestedBy          string `json:"requested_by"`
	CorrelationID        string `json:"correlation_id"`
	EntityCount          int64  `json:"entity_count"`
	RelationCount        int64  `json:"relation_count"`
	TimelineEntryCount   int64  `json:"timeline_entry_count"`
	CommandCount         int64  `json:"command_count"`
	MaxProjectionVersion int64  `json:"max_projection_version"`
}

type entityProjectionEventPayload struct {
	ProjectID         string                             `json:"project_id"`
	EntityKind        enumtypes.MissionControlEntityKind `json:"entity_kind"`
	EntityPublicID    string                             `json:"entity_public_id"`
	ProjectionVersion int64                              `json:"projection_version"`
}

type relationReplaceEventPayload struct {
	ProjectID      string `json:"project_id"`
	SourceEntityID int64  `json:"source_entity_id"`
	RelationCount  int    `json:"relation_count"`
}

type timelineEventPayload struct {
	ProjectID        string                                     `json:"project_id"`
	EntityID         int64                                      `json:"entity_id"`
	EntryExternalKey string                                     `json:"entry_external_key"`
	SourceKind       enumtypes.MissionControlTimelineSourceKind `json:"source_kind"`
}

type commandEventPayload struct {
	ProjectID         string                                       `json:"project_id"`
	CommandID         string                                       `json:"command_id"`
	CommandKind       enumtypes.MissionControlCommandKind          `json:"command_kind"`
	Status            enumtypes.MissionControlCommandStatus        `json:"status"`
	FailureReason     enumtypes.MissionControlCommandFailureReason `json:"failure_reason,omitempty"`
	BusinessIntentKey string                                       `json:"business_intent_key"`
	CorrelationID     string                                       `json:"correlation_id"`
	EntityRefs        []valuetypes.MissionControlEntityRef         `json:"entity_refs,omitempty"`
}

func (s *Service) capabilities() (valuetypes.MissionControlRolloutCapabilities, error) {
	return ResolveRolloutCapabilities(s.cfg.RolloutState)
}

func (s *Service) ensureDomainWriteAllowed() error {
	caps, err := s.capabilities()
	if err != nil {
		return err
	}
	if !caps.CanRunWarmup {
		return errs.FailedPrecondition{Msg: "mission control domain write path disabled"}
	}
	return nil
}

func (s *Service) ensureReadAllowed() error {
	caps, err := s.capabilities()
	if err != nil {
		return err
	}
	if !caps.CanServeSnapshot {
		return errs.FailedPrecondition{Msg: "mission control read path disabled"}
	}
	return nil
}

func (s *Service) ensureCommandSubmissionAllowed() error {
	caps, err := s.capabilities()
	if err != nil {
		return err
	}
	if !caps.CanSubmitCoreCommand {
		return errs.FailedPrecondition{Msg: "mission control command submission disabled"}
	}
	return nil
}

func (s *Service) insertFlowEvent(ctx context.Context, correlationID string, eventType floweventdomain.EventType, payload any) {
	if s.flowEvents == nil || strings.TrimSpace(correlationID) == "" {
		return
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_ = s.flowEvents.Insert(ctx, floweventrepo.InsertParams{
		CorrelationID: strings.TrimSpace(correlationID),
		ActorType:     floweventdomain.ActorTypeSystem,
		ActorID:       floweventdomain.ActorIDControlPlane,
		EventType:     eventType,
		Payload:       rawPayload,
		CreatedAt:     s.now(),
	})
}

func normalizeTimelineLimit(limit int, fallback int) int {
	if limit > 0 {
		return limit
	}
	if fallback > 0 {
		return fallback
	}
	return 50
}

func normalizeStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeEntityRef(ref *valuetypes.MissionControlEntityRef) *valuetypes.MissionControlEntityRef {
	if ref == nil {
		return nil
	}
	normalized := *ref
	normalized.EntityPublicID = strings.TrimSpace(normalized.EntityPublicID)
	if normalized.EntityKind == "" && normalized.EntityPublicID == "" {
		return nil
	}
	return &normalized
}

func effectiveCommandTargetRef(
	commandKind enumtypes.MissionControlCommandKind,
	targetRef *valuetypes.MissionControlEntityRef,
	payload valuetypes.MissionControlCommandPayload,
) (*valuetypes.MissionControlEntityRef, error) {
	normalizedTarget := normalizeEntityRef(targetRef)
	if commandKind != enumtypes.MissionControlCommandKindDiscussionFormalize {
		return normalizedTarget, nil
	}
	if payload.DiscussionFormalize == nil {
		return nil, errs.Validation{Field: "payload", Msg: "discussion.formalize payload is required"}
	}
	sourceRef := normalizeEntityRef(&payload.DiscussionFormalize.SourceEntityRef)
	if sourceRef == nil {
		return nil, errs.Validation{Field: "payload.source_entity_ref", Msg: "must contain kind and public id"}
	}
	if normalizedTarget == nil {
		return sourceRef, nil
	}
	if normalizedTarget.EntityKind != sourceRef.EntityKind || normalizedTarget.EntityPublicID != sourceRef.EntityPublicID {
		return nil, errs.Validation{Field: "target_entity_ref", Msg: "must match payload.source_entity_ref for discussion.formalize"}
	}
	return sourceRef, nil
}

func normalizeEntityRefs(refs []valuetypes.MissionControlEntityRef) []valuetypes.MissionControlEntityRef {
	if len(refs) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(refs))
	out := make([]valuetypes.MissionControlEntityRef, 0, len(refs))
	for _, ref := range refs {
		normalized := normalizeEntityRef(&ref)
		if normalized == nil || normalized.EntityKind == "" || normalized.EntityPublicID == "" {
			continue
		}
		key := string(normalized.EntityKind) + "/" + normalized.EntityPublicID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, *normalized)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func entityRefFromEntity(entity missioncontrolrepo.Entity) valuetypes.MissionControlEntityRef {
	return valuetypes.MissionControlEntityRef{
		EntityKind:     entity.EntityKind,
		EntityPublicID: entity.EntityExternalKey,
	}
}

func approvalSnapshotFromCommand(command missioncontrolrepo.Command, approverActorID string) *valuetypes.MissionControlApprovalSnapshot {
	if command.ApprovalState == enumtypes.MissionControlApprovalStateNotRequired &&
		command.ApprovalRequestID == "" &&
		command.ApprovalRequestedAt == nil &&
		command.ApprovalDecidedAt == nil &&
		strings.TrimSpace(approverActorID) == "" {
		return nil
	}
	return &valuetypes.MissionControlApprovalSnapshot{
		ApprovalState:     command.ApprovalState,
		ApprovalRequestID: command.ApprovalRequestID,
		RequestedAt:       command.ApprovalRequestedAt,
		DecidedAt:         command.ApprovalDecidedAt,
		ApproverActorID:   strings.TrimSpace(approverActorID),
	}
}

func mergeCommandResultPayload(
	current valuetypes.MissionControlCommandResultPayload,
	statusMessage string,
	approval *valuetypes.MissionControlApprovalSnapshot,
	providerDeliveryIDs []string,
) valuetypes.MissionControlCommandResultPayload {
	if trimmed := strings.TrimSpace(statusMessage); trimmed != "" {
		current.StatusMessage = trimmed
	}
	if approval != nil {
		current.Approval = approval
	}
	if normalized := normalizeStringSlice(providerDeliveryIDs); len(normalized) > 0 {
		current.ProviderDeliveryIDs = normalized
	}
	current.EntityRefs = normalizeEntityRefs(current.EntityRefs)
	return current
}

func decodeCommandResultPayload(raw []byte) (valuetypes.MissionControlCommandResultPayload, error) {
	if len(raw) == 0 {
		return valuetypes.MissionControlCommandResultPayload{}, nil
	}
	var payload valuetypes.MissionControlCommandResultPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return valuetypes.MissionControlCommandResultPayload{}, err
	}
	payload.EntityRefs = normalizeEntityRefs(payload.EntityRefs)
	payload.ProviderDeliveryIDs = normalizeStringSlice(payload.ProviderDeliveryIDs)
	return payload, nil
}

func encodeCommandPayload(payload valuetypes.MissionControlCommandPayload) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal mission control command payload: %w", err)
	}
	return raw, nil
}

func encodeCommandResultPayload(payload valuetypes.MissionControlCommandResultPayload) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal mission control command result payload: %w", err)
	}
	return raw, nil
}

func allowedWhenDegraded(kind enumtypes.MissionControlCommandKind) bool {
	return kind == enumtypes.MissionControlCommandKindRetrySync
}

func retrySyncTargetStatusAllowed(status enumtypes.MissionControlCommandStatus) bool {
	return status == enumtypes.MissionControlCommandStatusFailed
}

func commandRequiresExistingTarget(kind enumtypes.MissionControlCommandKind) bool {
	switch kind {
	case enumtypes.MissionControlCommandKindDiscussionFormalize, enumtypes.MissionControlCommandKindStageNextStep:
		return true
	default:
		return false
	}
}

func normalizeCommandPayload(kind enumtypes.MissionControlCommandKind, payload valuetypes.MissionControlCommandPayload) (valuetypes.MissionControlCommandPayload, []valuetypes.MissionControlEntityRef, error) {
	if selectedPayloadCount(payload) != 1 {
		return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload", Msg: "must contain exactly one command payload variant"}
	}

	switch kind {
	case enumtypes.MissionControlCommandKindDiscussionCreate:
		if payload.DiscussionCreate == nil {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload", Msg: "discussion.create payload is required"}
		}
		normalized := *payload.DiscussionCreate
		normalized.Title = strings.TrimSpace(normalized.Title)
		normalized.BodyMarkdown = strings.TrimSpace(normalized.BodyMarkdown)
		normalized.ParentEntityRef = normalizeEntityRef(normalized.ParentEntityRef)
		if normalized.Title == "" {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload.title", Msg: "is required"}
		}
		refList := make([]valuetypes.MissionControlEntityRef, 0, 1)
		if normalized.ParentEntityRef != nil {
			if normalized.ParentEntityRef.EntityKind == "" || normalized.ParentEntityRef.EntityPublicID == "" {
				return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload.parent_entity_ref", Msg: "must contain kind and public id"}
			}
			refList = append(refList, *normalized.ParentEntityRef)
		}
		return valuetypes.MissionControlCommandPayload{
			DiscussionCreate: &normalized,
		}, normalizeEntityRefs(refList), nil
	case enumtypes.MissionControlCommandKindWorkItemCreate:
		if payload.WorkItemCreate == nil {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload", Msg: "work_item.create payload is required"}
		}
		normalized := *payload.WorkItemCreate
		normalized.Title = strings.TrimSpace(normalized.Title)
		normalized.BodyMarkdown = strings.TrimSpace(normalized.BodyMarkdown)
		normalized.InitialLabels = normalizeStringSlice(normalized.InitialLabels)
		normalized.RelatedEntityRefs = normalizeEntityRefs(normalized.RelatedEntityRefs)
		if normalized.Title == "" {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload.title", Msg: "is required"}
		}
		return valuetypes.MissionControlCommandPayload{
			WorkItemCreate: &normalized,
		}, normalized.RelatedEntityRefs, nil
	case enumtypes.MissionControlCommandKindDiscussionFormalize:
		if payload.DiscussionFormalize == nil {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload", Msg: "discussion.formalize payload is required"}
		}
		normalized := *payload.DiscussionFormalize
		normalizedSourceRef := normalizeEntityRef(&normalized.SourceEntityRef)
		if normalizedSourceRef == nil {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload.source_entity_ref", Msg: "must contain kind and public id"}
		}
		normalized.SourceEntityRef = *normalizedSourceRef
		normalized.FormalizedKind = strings.TrimSpace(normalized.FormalizedKind)
		normalized.Title = strings.TrimSpace(normalized.Title)
		normalized.BodyMarkdown = strings.TrimSpace(normalized.BodyMarkdown)
		if normalized.FormalizedKind == "" {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload.formalized_kind", Msg: "is required"}
		}
		if normalized.Title == "" {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload.title", Msg: "is required"}
		}
		refList := []valuetypes.MissionControlEntityRef{normalized.SourceEntityRef}
		return valuetypes.MissionControlCommandPayload{
			DiscussionFormalize: &normalized,
		}, normalizeEntityRefs(refList), nil
	case enumtypes.MissionControlCommandKindStageNextStep:
		if payload.StageNextStep == nil {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload", Msg: "stage.next_step.execute payload is required"}
		}
		normalized := *payload.StageNextStep
		normalized.ThreadKind = strings.TrimSpace(normalized.ThreadKind)
		normalized.TargetLabel = strings.TrimSpace(normalized.TargetLabel)
		normalized.RemovedLabels = normalizeStringSlice(normalized.RemovedLabels)
		normalized.DisplayVariant = strings.TrimSpace(normalized.DisplayVariant)
		if normalized.ApprovalRequirement == "" {
			normalized.ApprovalRequirement = enumtypes.MissionControlApprovalRequirementNone
		}
		if normalized.ThreadKind == "" {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload.thread_kind", Msg: "is required"}
		}
		if normalized.ThreadNumber <= 0 {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload.thread_number", Msg: "must be positive"}
		}
		if normalized.TargetLabel == "" {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload.target_label", Msg: "is required"}
		}
		switch normalized.ApprovalRequirement {
		case enumtypes.MissionControlApprovalRequirementNone, enumtypes.MissionControlApprovalRequirementOwnerReview:
		default:
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload.approval_requirement", Msg: "must be a known mission control approval requirement"}
		}
		return valuetypes.MissionControlCommandPayload{
			StageNextStep: &normalized,
		}, nil, nil
	case enumtypes.MissionControlCommandKindRetrySync:
		if payload.RetrySync == nil {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload", Msg: "command.retry_sync payload is required"}
		}
		normalized := *payload.RetrySync
		normalized.CommandID = strings.TrimSpace(normalized.CommandID)
		normalized.RetryReason = strings.TrimSpace(normalized.RetryReason)
		if normalized.CommandID == "" {
			return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "payload.command_id", Msg: "is required"}
		}
		return valuetypes.MissionControlCommandPayload{
			RetrySync: &normalized,
		}, nil, nil
	default:
		return valuetypes.MissionControlCommandPayload{}, nil, errs.Validation{Field: "command_kind", Msg: "must be a known mission control command kind"}
	}
}

func selectedPayloadCount(payload valuetypes.MissionControlCommandPayload) int {
	count := 0
	if payload.DiscussionCreate != nil {
		count++
	}
	if payload.WorkItemCreate != nil {
		count++
	}
	if payload.DiscussionFormalize != nil {
		count++
	}
	if payload.StageNextStep != nil {
		count++
	}
	if payload.RetrySync != nil {
		count++
	}
	return count
}

func transitionStatusAllowed(current enumtypes.MissionControlCommandStatus, target enumtypes.MissionControlCommandStatus) bool {
	switch current {
	case enumtypes.MissionControlCommandStatusAccepted:
		return target == enumtypes.MissionControlCommandStatusQueued ||
			target == enumtypes.MissionControlCommandStatusPendingSync ||
			target == enumtypes.MissionControlCommandStatusFailed ||
			target == enumtypes.MissionControlCommandStatusBlocked ||
			target == enumtypes.MissionControlCommandStatusCancelled
	case enumtypes.MissionControlCommandStatusPendingApproval:
		return target == enumtypes.MissionControlCommandStatusQueued ||
			target == enumtypes.MissionControlCommandStatusBlocked ||
			target == enumtypes.MissionControlCommandStatusCancelled
	case enumtypes.MissionControlCommandStatusQueued:
		return target == enumtypes.MissionControlCommandStatusPendingSync ||
			target == enumtypes.MissionControlCommandStatusReconciled ||
			target == enumtypes.MissionControlCommandStatusFailed ||
			target == enumtypes.MissionControlCommandStatusBlocked ||
			target == enumtypes.MissionControlCommandStatusCancelled
	case enumtypes.MissionControlCommandStatusPendingSync:
		return target == enumtypes.MissionControlCommandStatusReconciled ||
			target == enumtypes.MissionControlCommandStatusFailed ||
			target == enumtypes.MissionControlCommandStatusBlocked ||
			target == enumtypes.MissionControlCommandStatusCancelled
	default:
		return false
	}
}

func commandEventTypeForStatus(status enumtypes.MissionControlCommandStatus) floweventdomain.EventType {
	switch status {
	case enumtypes.MissionControlCommandStatusQueued:
		return eventTypeMissionControlCommandQueued
	case enumtypes.MissionControlCommandStatusPendingSync:
		return eventTypeMissionControlCommandPendingSync
	case enumtypes.MissionControlCommandStatusReconciled:
		return eventTypeMissionControlCommandReconciled
	case enumtypes.MissionControlCommandStatusFailed:
		return eventTypeMissionControlCommandFailed
	case enumtypes.MissionControlCommandStatusBlocked:
		return eventTypeMissionControlCommandBlocked
	case enumtypes.MissionControlCommandStatusCancelled:
		return eventTypeMissionControlCommandCancelled
	default:
		return eventTypeMissionControlCommandAccepted
	}
}

func normalizeEventEntityRefs(target *missioncontrolrepo.Entity, refs []valuetypes.MissionControlEntityRef) []valuetypes.MissionControlEntityRef {
	out := make([]valuetypes.MissionControlEntityRef, 0, len(refs)+1)
	out = append(out, refs...)
	if target != nil {
		out = append(out, entityRefFromEntity(*target))
	}
	return normalizeEntityRefs(out)
}

func normalizeProviderDeliveryIDs(values []string) []string {
	return normalizeStringSlice(values)
}

func commandAlreadyQueued(status enumtypes.MissionControlCommandStatus) bool {
	return status == enumtypes.MissionControlCommandStatusQueued ||
		status == enumtypes.MissionControlCommandStatusPendingSync ||
		status == enumtypes.MissionControlCommandStatusReconciled
}

func duplicateDeliveryTransition(
	command missioncontrolrepo.Command,
	target enumtypes.MissionControlCommandStatus,
	providerDeliveryIDs []string,
	failureReason enumtypes.MissionControlCommandFailureReason,
) bool {
	switch target {
	case enumtypes.MissionControlCommandStatusPendingSync, enumtypes.MissionControlCommandStatusReconciled:
		if command.Status != target && !(target == enumtypes.MissionControlCommandStatusPendingSync && command.Status == enumtypes.MissionControlCommandStatusReconciled) {
			return false
		}
		return commandContainsProviderDeliveries(command, providerDeliveryIDs)
	case enumtypes.MissionControlCommandStatusFailed:
		if command.Status != enumtypes.MissionControlCommandStatusFailed {
			return false
		}
		if failureReason != "" && command.FailureReason != failureReason {
			return false
		}
		return commandContainsProviderDeliveries(command, providerDeliveryIDs)
	default:
		return false
	}
}

func commandContainsProviderDeliveries(command missioncontrolrepo.Command, providerDeliveryIDs []string) bool {
	requested := normalizeProviderDeliveryIDs(providerDeliveryIDs)
	if len(requested) == 0 {
		return true
	}
	current, err := decodeCommandResultPayload(command.ResultPayloadJSON)
	if err != nil {
		return false
	}
	if len(current.ProviderDeliveryIDs) == 0 {
		return false
	}
	currentSet := make(map[string]struct{}, len(current.ProviderDeliveryIDs))
	for _, deliveryID := range current.ProviderDeliveryIDs {
		currentSet[deliveryID] = struct{}{}
	}
	for _, deliveryID := range requested {
		if _, ok := currentSet[deliveryID]; !ok {
			return false
		}
	}
	return true
}
