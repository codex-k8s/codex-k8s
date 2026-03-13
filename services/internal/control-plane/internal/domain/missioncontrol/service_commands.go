package missioncontrol

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	missioncontrolrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/missioncontrol"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
	"github.com/google/uuid"
)

// SubmitCommand validates typed input, applies stale/policy guards and persists one command-ledger row.
func (s *Service) SubmitCommand(ctx context.Context, params SubmitCommandParams) (CommandAdmission, error) {
	if err := s.ensureCommandSubmissionAllowed(); err != nil {
		return CommandAdmission{}, err
	}
	params.ProjectID = strings.TrimSpace(params.ProjectID)
	params.ActorID = strings.TrimSpace(params.ActorID)
	params.CorrelationID = strings.TrimSpace(params.CorrelationID)
	params.BusinessIntentKey = strings.TrimSpace(params.BusinessIntentKey)

	requestedAt := params.RequestedAt
	if requestedAt.IsZero() {
		requestedAt = s.now()
	} else {
		requestedAt = requestedAt.UTC()
	}

	if params.ProjectID == "" {
		return CommandAdmission{}, errs.Validation{Field: "project_id", Msg: "is required"}
	}
	if params.ActorID == "" {
		return CommandAdmission{}, errs.Validation{Field: "actor_id", Msg: "is required"}
	}
	if params.CorrelationID == "" {
		return CommandAdmission{}, errs.Validation{Field: "correlation_id", Msg: "is required"}
	}
	if params.CommandKind == "" {
		return CommandAdmission{}, errs.Validation{Field: "command_kind", Msg: "is required"}
	}
	if params.BusinessIntentKey == "" {
		return CommandAdmission{}, errs.Validation{Field: "business_intent_key", Msg: "is required"}
	}

	normalizedPayload, entityRefs, err := normalizeCommandPayload(params.CommandKind, params.Payload)
	if err != nil {
		return CommandAdmission{}, err
	}

	targetRef, err := effectiveCommandTargetRef(params.CommandKind, params.TargetEntityRef, normalizedPayload)
	if err != nil {
		return CommandAdmission{}, err
	}

	targetEntity, err := s.resolveCommandTarget(ctx, params.ProjectID, params.CommandKind, targetRef, params.ExpectedProjectionVersion)
	if err != nil {
		return CommandAdmission{}, err
	}
	entityRefs = normalizeEventEntityRefs(targetEntity, entityRefs)
	if targetEntity != nil && targetEntity.ProjectionVersion != params.ExpectedProjectionVersion {
		return s.createBlockedCommand(ctx, params, normalizedPayload, targetEntity, entityRefs, enumtypes.MissionControlCommandFailureReasonProjectionStale, requestedAt)
	}

	if targetEntity != nil && targetEntity.SyncStatus == enumtypes.MissionControlSyncStatusDegraded && !allowedWhenDegraded(params.CommandKind) {
		return s.createBlockedCommand(ctx, params, normalizedPayload, targetEntity, entityRefs, enumtypes.MissionControlCommandFailureReasonPolicyDenied, requestedAt)
	}
	if params.CommandKind == enumtypes.MissionControlCommandKindRetrySync {
		if err := s.validateRetryTarget(ctx, params.ProjectID, normalizedPayload); err != nil {
			return CommandAdmission{}, err
		}
	}

	status := enumtypes.MissionControlCommandStatusAccepted
	approvalState := enumtypes.MissionControlApprovalStateNotRequired
	var approvalRequestID string
	var approvalRequestedAt *time.Time
	var approval *valuetypes.MissionControlApprovalSnapshot
	if normalizedPayload.StageNextStep != nil && normalizedPayload.StageNextStep.ApprovalRequirement == enumtypes.MissionControlApprovalRequirementOwnerReview {
		status = enumtypes.MissionControlCommandStatusPendingApproval
		approvalState = enumtypes.MissionControlApprovalStatePending
		approvalRequestID = uuid.NewString()
		approvalRequestedAt = &requestedAt
		approval = &valuetypes.MissionControlApprovalSnapshot{
			ApprovalState:     approvalState,
			ApprovalRequestID: approvalRequestID,
			RequestedAt:       approvalRequestedAt,
		}
	}

	payloadJSON, err := encodeCommandPayload(normalizedPayload)
	if err != nil {
		return CommandAdmission{}, err
	}
	resultJSON, err := encodeCommandResultPayload(valuetypes.MissionControlCommandResultPayload{
		EntityRefs: entityRefs,
		Approval:   approval,
	})
	if err != nil {
		return CommandAdmission{}, err
	}

	command, err := s.repository.CreateCommand(ctx, missioncontrolrepo.CreateCommandParams{
		ProjectID:           params.ProjectID,
		CommandKind:         params.CommandKind,
		TargetEntityID:      entityIDPointer(targetEntity),
		ActorID:             params.ActorID,
		BusinessIntentKey:   params.BusinessIntentKey,
		CorrelationID:       params.CorrelationID,
		Status:              status,
		ApprovalRequestID:   approvalRequestID,
		ApprovalState:       approvalState,
		ApprovalRequestedAt: approvalRequestedAt,
		PayloadJSON:         payloadJSON,
		ResultPayloadJSON:   resultJSON,
		RequestedAt:         requestedAt,
		UpdatedAt:           requestedAt,
	})
	if err != nil {
		return CommandAdmission{}, s.mapDuplicateIntentError(ctx, err, params.ProjectID, params.BusinessIntentKey, params.CorrelationID)
	}

	s.insertFlowEvent(ctx, params.CorrelationID, eventTypeMissionControlCommandAccepted, commandEventPayload{
		ProjectID:         command.ProjectID,
		CommandID:         command.ID,
		CommandKind:       command.CommandKind,
		Status:            command.Status,
		BusinessIntentKey: command.BusinessIntentKey,
		CorrelationID:     command.CorrelationID,
		EntityRefs:        entityRefs,
	})

	return CommandAdmission{
		Command:      command,
		TargetEntity: targetEntity,
		EntityRefs:   entityRefs,
		Approval:     approvalSnapshotFromCommand(command, ""),
	}, nil
}

// QueueCommand transitions one accepted command into queued state.
func (s *Service) QueueCommand(ctx context.Context, params CommandQueueParams) (Command, error) {
	command, err := s.loadCommandForUpdate(ctx, params.ProjectID, params.CommandID)
	if err != nil {
		return Command{}, err
	}
	if commandAlreadyQueued(command.Status) {
		return command, nil
	}
	return s.transitionCommand(ctx, command, enumtypes.MissionControlCommandStatusQueued, transitionOptions{
		statusMessage: params.StatusMessage,
		updatedAt:     params.UpdatedAt,
	})
}

// MarkCommandPendingSync transitions one queued/accepted command into pending_sync.
func (s *Service) MarkCommandPendingSync(ctx context.Context, params CommandSyncProgressParams) (Command, error) {
	command, err := s.loadCommandForUpdate(ctx, params.ProjectID, params.CommandID)
	if err != nil {
		return Command{}, err
	}
	if duplicateDeliveryTransition(command, enumtypes.MissionControlCommandStatusPendingSync, params.ProviderDeliveryIDs, "") {
		return command, nil
	}
	return s.transitionCommand(ctx, command, enumtypes.MissionControlCommandStatusPendingSync, transitionOptions{
		statusMessage:       params.StatusMessage,
		providerDeliveryIDs: params.ProviderDeliveryIDs,
		updatedAt:           params.UpdatedAt,
	})
}

// MarkCommandReconciled closes one command as successfully reconciled.
func (s *Service) MarkCommandReconciled(ctx context.Context, params CommandReconcileParams) (Command, error) {
	command, err := s.loadCommandForUpdate(ctx, params.ProjectID, params.CommandID)
	if err != nil {
		return Command{}, err
	}
	if duplicateDeliveryTransition(command, enumtypes.MissionControlCommandStatusReconciled, params.ProviderDeliveryIDs, "") {
		return command, nil
	}
	return s.transitionCommand(ctx, command, enumtypes.MissionControlCommandStatusReconciled, transitionOptions{
		statusMessage:       params.StatusMessage,
		providerDeliveryIDs: params.ProviderDeliveryIDs,
		updatedAt:           params.UpdatedAt,
		reconciledAt:        params.ReconciledAt,
	})
}

// MarkCommandFailed closes one command as failed with typed failure_reason.
func (s *Service) MarkCommandFailed(ctx context.Context, params CommandFailureParams) (Command, error) {
	return s.transitionFailedCommand(ctx, params)
}

// CancelCommand stops one in-flight command without mutating provider state further.
func (s *Service) CancelCommand(ctx context.Context, params CommandCancelParams) (Command, error) {
	return s.transitionSimpleCommand(ctx, params.ProjectID, params.CommandID, enumtypes.MissionControlCommandStatusCancelled, params.StatusMessage, params.UpdatedAt)
}

// ApplyApprovalDecision transitions one pending_approval command according to owner decision.
func (s *Service) ApplyApprovalDecision(ctx context.Context, params ApprovalDecisionParams) (Command, error) {
	command, err := s.loadCommandForUpdate(ctx, params.ProjectID, params.CommandID)
	if err != nil {
		return Command{}, err
	}
	if command.Status != enumtypes.MissionControlCommandStatusPendingApproval {
		return Command{}, errs.FailedPrecondition{Msg: "mission control command is not pending approval"}
	}
	if strings.TrimSpace(command.ApprovalRequestID) == "" {
		return Command{}, errs.FailedPrecondition{Msg: "mission control command has no approval request id"}
	}
	targetStatus, opts, err := approvalDecisionTransition(params, s.now())
	if err != nil {
		return Command{}, err
	}
	return s.transitionCommand(ctx, command, targetStatus, opts)
}

type transitionOptions struct {
	failureReason       enumtypes.MissionControlCommandFailureReason
	approvalState       enumtypes.MissionControlApprovalState
	approverActorID     string
	statusMessage       string
	providerDeliveryIDs []string
	updatedAt           time.Time
	reconciledAt        time.Time
	approvalDecidedAt   time.Time
}

func approvalDecisionTransition(params ApprovalDecisionParams, now time.Time) (enumtypes.MissionControlCommandStatus, transitionOptions, error) {
	decidedAt := zeroTransitionTime(params.UpdatedAt, now)
	base := transitionOptions{
		approvalState:     params.Decision,
		approverActorID:   params.ApproverActorID,
		statusMessage:     params.StatusMessage,
		updatedAt:         params.UpdatedAt,
		approvalDecidedAt: decidedAt,
	}

	switch params.Decision {
	case enumtypes.MissionControlApprovalStateApproved:
		return enumtypes.MissionControlCommandStatusQueued, base, nil
	case enumtypes.MissionControlApprovalStateDenied:
		base.failureReason = enumtypes.MissionControlCommandFailureReasonApprovalDenied
		return enumtypes.MissionControlCommandStatusBlocked, base, nil
	case enumtypes.MissionControlApprovalStateExpired:
		base.failureReason = enumtypes.MissionControlCommandFailureReasonApprovalExpired
		return enumtypes.MissionControlCommandStatusBlocked, base, nil
	default:
		return "", transitionOptions{}, errs.Validation{Field: "decision", Msg: "must be approved, denied or expired"}
	}
}

func (s *Service) transitionCommand(ctx context.Context, current Command, target enumtypes.MissionControlCommandStatus, opts transitionOptions) (Command, error) {
	if err := s.ensureDomainWriteAllowed(); err != nil {
		return Command{}, err
	}
	if !transitionStatusAllowed(current.Status, target) {
		return Command{}, errs.FailedPrecondition{Msg: "mission control command transition is not allowed"}
	}
	if (target == enumtypes.MissionControlCommandStatusFailed || target == enumtypes.MissionControlCommandStatusBlocked) && opts.failureReason == "" {
		return Command{}, errs.Validation{Field: "failure_reason", Msg: "is required for failed or blocked transitions"}
	}
	if current.Status == enumtypes.MissionControlCommandStatusPendingApproval {
		if current.ApprovalState != enumtypes.MissionControlApprovalStatePending {
			return Command{}, errs.FailedPrecondition{Msg: "mission control pending approval command is in unexpected approval state"}
		}
		if target == enumtypes.MissionControlCommandStatusQueued && opts.approvalState != enumtypes.MissionControlApprovalStateApproved {
			return Command{}, errs.FailedPrecondition{Msg: "mission control approval decision must be approved before queueing"}
		}
		if target == enumtypes.MissionControlCommandStatusBlocked &&
			opts.approvalState != enumtypes.MissionControlApprovalStateDenied &&
			opts.approvalState != enumtypes.MissionControlApprovalStateExpired {
			return Command{}, errs.FailedPrecondition{Msg: "mission control blocked approval transition must carry denied or expired decision"}
		}
	}

	resultPayload, err := decodeCommandResultPayload(current.ResultPayloadJSON)
	if err != nil {
		return Command{}, err
	}
	approval := resultPayload.Approval
	if approval == nil {
		approval = approvalSnapshotFromCommand(current, "")
	}
	if opts.approvalState != "" {
		if approval == nil {
			approval = &valuetypes.MissionControlApprovalSnapshot{}
		}
		approval.ApprovalState = opts.approvalState
		approval.ApprovalRequestID = current.ApprovalRequestID
		approval.RequestedAt = current.ApprovalRequestedAt
		if !opts.approvalDecidedAt.IsZero() {
			decidedAt := opts.approvalDecidedAt.UTC()
			approval.DecidedAt = &decidedAt
		}
		approval.ApproverActorID = opts.approverActorID
	}
	resultPayload = mergeCommandResultPayload(resultPayload, opts.statusMessage, approval, opts.providerDeliveryIDs)
	resultJSON, err := encodeCommandResultPayload(resultPayload)
	if err != nil {
		return Command{}, err
	}

	updateParams := missioncontrolrepo.UpdateCommandStatusParams{
		ProjectID: current.ProjectID,
		CommandID: current.ID,
		Status:    target,
		UpdatedAt: zeroTransitionTime(opts.updatedAt, s.now()),
		ResultPayloadPatch: missioncontrolrepo.OptionalJSONPatch{
			Set:   true,
			Value: resultJSON,
		},
	}
	if opts.failureReason != "" {
		updateParams.FailureReasonPatch = missioncontrolrepo.CommandFailureReasonPatch{
			Set:   true,
			Value: opts.failureReason,
		}
	}
	if normalizedDeliveries := normalizeProviderDeliveryIDs(opts.providerDeliveryIDs); len(normalizedDeliveries) > 0 {
		deliveryJSON, marshalErr := encodeStringArray(normalizedDeliveries)
		if marshalErr != nil {
			return Command{}, marshalErr
		}
		updateParams.ProviderDeliveriesPatch = missioncontrolrepo.OptionalJSONPatch{
			Set:   true,
			Value: deliveryJSON,
		}
	}
	if opts.approvalState != "" {
		updateParams.ApprovalStatePatch = missioncontrolrepo.CommandApprovalStatePatch{
			Set:   true,
			Value: opts.approvalState,
		}
		if !opts.approvalDecidedAt.IsZero() {
			decidedAt := opts.approvalDecidedAt.UTC()
			updateParams.ApprovalDecidedAtPatch = missioncontrolrepo.OptionalTimePatch{
				Set:   true,
				Value: &decidedAt,
			}
		}
	}
	if target == enumtypes.MissionControlCommandStatusReconciled {
		reconciledAt := zeroTransitionTime(opts.reconciledAt, s.now())
		updateParams.ReconciledAtPatch = missioncontrolrepo.OptionalTimePatch{
			Set:   true,
			Value: &reconciledAt,
		}
	}

	updated, found, err := s.repository.UpdateCommandStatus(ctx, updateParams)
	if err != nil {
		return Command{}, err
	}
	if !found {
		return Command{}, errs.NotFound{Msg: "mission control command not found"}
	}
	s.insertFlowEvent(ctx, updated.CorrelationID, commandEventTypeForStatus(updated.Status), commandEventPayload{
		ProjectID:         updated.ProjectID,
		CommandID:         updated.ID,
		CommandKind:       updated.CommandKind,
		Status:            updated.Status,
		FailureReason:     updated.FailureReason,
		BusinessIntentKey: updated.BusinessIntentKey,
		CorrelationID:     updated.CorrelationID,
		EntityRefs:        resultPayload.EntityRefs,
	})
	return updated, nil
}

func (s *Service) transitionCommandByID(
	ctx context.Context,
	projectID string,
	commandID string,
	target enumtypes.MissionControlCommandStatus,
	opts transitionOptions,
) (Command, error) {
	command, err := s.loadCommandForUpdate(ctx, projectID, commandID)
	if err != nil {
		return Command{}, err
	}
	return s.transitionCommand(ctx, command, target, opts)
}

func (s *Service) transitionSimpleCommand(
	ctx context.Context,
	projectID string,
	commandID string,
	target enumtypes.MissionControlCommandStatus,
	statusMessage string,
	updatedAt time.Time,
) (Command, error) {
	return s.transitionCommandByID(ctx, projectID, commandID, target, transitionOptions{
		statusMessage: statusMessage,
		updatedAt:     updatedAt,
	})
}

func (s *Service) transitionFailedCommand(ctx context.Context, params CommandFailureParams) (Command, error) {
	command, err := s.loadCommandForUpdate(ctx, params.ProjectID, params.CommandID)
	if err != nil {
		return Command{}, err
	}
	if duplicateDeliveryTransition(command, enumtypes.MissionControlCommandStatusFailed, params.ProviderDeliveryIDs, params.FailureReason) {
		return command, nil
	}
	return s.transitionCommand(ctx, command, enumtypes.MissionControlCommandStatusFailed, transitionOptions{
		failureReason:       params.FailureReason,
		statusMessage:       params.StatusMessage,
		providerDeliveryIDs: params.ProviderDeliveryIDs,
		updatedAt:           params.UpdatedAt,
	})
}

func (s *Service) resolveCommandTarget(
	ctx context.Context,
	projectID string,
	commandKind enumtypes.MissionControlCommandKind,
	targetRef *valuetypes.MissionControlEntityRef,
	expectedProjectionVersion int64,
) (*Entity, error) {
	normalizedTargetRef := normalizeEntityRef(targetRef)
	if normalizedTargetRef == nil {
		if commandRequiresExistingTarget(commandKind) {
			return nil, errs.Validation{Field: "target_entity_ref", Msg: "is required for this command kind"}
		}
		return nil, nil
	}
	if normalizedTargetRef.EntityKind == "" || normalizedTargetRef.EntityPublicID == "" {
		return nil, errs.Validation{Field: "target_entity_ref", Msg: "must contain kind and public id"}
	}
	entity, found, err := s.repository.GetEntityByPublicID(ctx, projectID, normalizedTargetRef.EntityKind, normalizedTargetRef.EntityPublicID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errs.NotFound{Msg: "mission control entity not found"}
	}
	if expectedProjectionVersion <= 0 {
		return nil, errs.Validation{Field: "expected_projection_version", Msg: "is required for commands against an existing entity"}
	}
	if entity.ProjectionVersion != expectedProjectionVersion {
		return &entity, nil
	}
	return &entity, nil
}

func (s *Service) validateRetryTarget(ctx context.Context, projectID string, payload valuetypes.MissionControlCommandPayload) error {
	if payload.RetrySync == nil {
		return errs.Validation{Field: "payload", Msg: "retry_sync payload is required"}
	}
	target, found, err := s.repository.GetCommandByID(ctx, projectID, payload.RetrySync.CommandID)
	if err != nil {
		return err
	}
	if !found {
		return errs.NotFound{Msg: "mission control retry target command not found"}
	}
	if !retrySyncTargetStatusAllowed(target.Status) {
		return errs.FailedPrecondition{Msg: "mission control retry target status is not retryable"}
	}
	if payload.RetrySync.ExpectedStatus != "" && target.Status != payload.RetrySync.ExpectedStatus {
		return errs.FailedPrecondition{Msg: "mission control retry target status mismatch"}
	}
	return nil
}

func (s *Service) createBlockedCommand(
	ctx context.Context,
	params SubmitCommandParams,
	payload valuetypes.MissionControlCommandPayload,
	targetEntity *Entity,
	entityRefs []valuetypes.MissionControlEntityRef,
	failureReason enumtypes.MissionControlCommandFailureReason,
	requestedAt time.Time,
) (CommandAdmission, error) {
	payloadJSON, err := encodeCommandPayload(payload)
	if err != nil {
		return CommandAdmission{}, err
	}
	resultJSON, err := encodeCommandResultPayload(valuetypes.MissionControlCommandResultPayload{
		EntityRefs: entityRefs,
	})
	if err != nil {
		return CommandAdmission{}, err
	}
	command, err := s.repository.CreateCommand(ctx, missioncontrolrepo.CreateCommandParams{
		ProjectID:         params.ProjectID,
		CommandKind:       params.CommandKind,
		TargetEntityID:    entityIDPointer(targetEntity),
		ActorID:           params.ActorID,
		BusinessIntentKey: params.BusinessIntentKey,
		CorrelationID:     params.CorrelationID,
		Status:            enumtypes.MissionControlCommandStatusBlocked,
		FailureReason:     failureReason,
		ApprovalState:     enumtypes.MissionControlApprovalStateNotRequired,
		PayloadJSON:       payloadJSON,
		ResultPayloadJSON: resultJSON,
		RequestedAt:       requestedAt,
		UpdatedAt:         requestedAt,
	})
	if err != nil {
		return CommandAdmission{}, s.mapDuplicateIntentError(ctx, err, params.ProjectID, params.BusinessIntentKey, params.CorrelationID)
	}
	s.insertFlowEvent(ctx, params.CorrelationID, eventTypeMissionControlCommandBlocked, commandEventPayload{
		ProjectID:         command.ProjectID,
		CommandID:         command.ID,
		CommandKind:       command.CommandKind,
		Status:            command.Status,
		FailureReason:     command.FailureReason,
		BusinessIntentKey: command.BusinessIntentKey,
		CorrelationID:     command.CorrelationID,
		EntityRefs:        entityRefs,
	})
	return CommandAdmission{
		Command:      command,
		TargetEntity: targetEntity,
		EntityRefs:   entityRefs,
	}, nil
}

func (s *Service) mapDuplicateIntentError(ctx context.Context, err error, projectID string, businessIntentKey string, correlationID string) error {
	var duplicate missioncontrolrepo.DuplicateBusinessIntent
	if !errors.As(err, &duplicate) {
		return err
	}
	existing, found, lookupErr := s.repository.GetCommandByBusinessIntent(ctx, projectID, businessIntentKey)
	if lookupErr != nil {
		return lookupErr
	}
	if !found {
		return err
	}
	s.insertFlowEvent(ctx, correlationID, eventTypeMissionControlCommandDeduped, commandEventPayload{
		ProjectID:         existing.ProjectID,
		CommandID:         existing.ID,
		CommandKind:       existing.CommandKind,
		Status:            existing.Status,
		FailureReason:     existing.FailureReason,
		BusinessIntentKey: existing.BusinessIntentKey,
		CorrelationID:     existing.CorrelationID,
	})
	return DuplicateIntentError{
		ProjectID:         projectID,
		BusinessIntentKey: businessIntentKey,
		ExistingCommand:   existing,
	}
}

func (s *Service) loadCommandForUpdate(ctx context.Context, projectID string, commandID string) (Command, error) {
	if err := s.ensureDomainWriteAllowed(); err != nil {
		return Command{}, err
	}
	command, found, err := s.repository.GetCommandByID(ctx, projectID, commandID)
	if err != nil {
		return Command{}, err
	}
	if !found {
		return Command{}, errs.NotFound{Msg: "mission control command not found"}
	}
	return command, nil
}

func entityIDPointer(entity *Entity) *int64 {
	if entity == nil {
		return nil
	}
	return &entity.ID
}

func zeroTransitionTime(value time.Time, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback.UTC()
	}
	return value.UTC()
}

func encodeStringArray(values []string) ([]byte, error) {
	raw, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
