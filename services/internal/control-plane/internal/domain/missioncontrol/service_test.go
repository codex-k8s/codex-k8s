package missioncontrol

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	floweventrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/flowevent"
	missioncontrolrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/missioncontrol"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

func TestSubmitCommandPendingApprovalForStageNextStep(t *testing.T) {
	t.Parallel()

	svc, repo, events, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	_, err := repo.UpsertEntity(context.Background(), missioncontrolrepo.UpsertEntityParams{
		ProjectID:         "proj-1",
		EntityKind:        enumtypes.MissionControlEntityKindDiscussion,
		EntityExternalKey: "DISC-1",
		Title:             "Discussion",
		ProjectionVersion: 3,
		ProjectedAt:       now,
	})
	if err != nil {
		t.Fatalf("seed entity: %v", err)
	}

	result, err := svc.SubmitCommand(context.Background(), SubmitCommandParams{
		ProjectID:                 "proj-1",
		ActorID:                   "owner",
		CorrelationID:             "corr-1",
		CommandKind:               enumtypes.MissionControlCommandKindStageNextStep,
		TargetEntityRef:           &valuetypes.MissionControlEntityRef{EntityKind: enumtypes.MissionControlEntityKindDiscussion, EntityPublicID: "DISC-1"},
		BusinessIntentKey:         "intent-1",
		ExpectedProjectionVersion: 3,
		Payload: valuetypes.MissionControlCommandPayload{
			StageNextStep: &valuetypes.MissionControlStageNextStepExecutePayload{
				ThreadKind:          "issue",
				ThreadNumber:        370,
				TargetLabel:         "run:qa",
				ApprovalRequirement: enumtypes.MissionControlApprovalRequirementOwnerReview,
			},
		},
		RequestedAt: now,
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if got, want := result.Command.Status, enumtypes.MissionControlCommandStatusPendingApproval; got != want {
		t.Fatalf("command status = %s, want %s", got, want)
	}
	if got, want := result.Command.ApprovalState, enumtypes.MissionControlApprovalStatePending; got != want {
		t.Fatalf("approval state = %s, want %s", got, want)
	}
	if result.Command.ApprovalRequestID == "" {
		t.Fatal("approval request id must be generated")
	}
	if len(result.EntityRefs) != 1 || result.EntityRefs[0].EntityPublicID != "DISC-1" {
		t.Fatalf("unexpected entity refs: %+v", result.EntityRefs)
	}
	if len(events.items) != 1 || events.items[0].EventType != eventTypeMissionControlCommandAccepted {
		t.Fatalf("unexpected events: %+v", events.items)
	}
}

func TestSubmitCommandBlocksOnStaleProjectionVersion(t *testing.T) {
	t.Parallel()

	svc, repo, events, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	_, err := repo.UpsertEntity(context.Background(), missioncontrolrepo.UpsertEntityParams{
		ProjectID:         "proj-1",
		EntityKind:        enumtypes.MissionControlEntityKindDiscussion,
		EntityExternalKey: "DISC-1",
		Title:             "Discussion",
		ProjectionVersion: 7,
		ProjectedAt:       now,
	})
	if err != nil {
		t.Fatalf("seed entity: %v", err)
	}

	result, err := svc.SubmitCommand(context.Background(), SubmitCommandParams{
		ProjectID:                 "proj-1",
		ActorID:                   "owner",
		CorrelationID:             "corr-stale",
		CommandKind:               enumtypes.MissionControlCommandKindDiscussionFormalize,
		TargetEntityRef:           &valuetypes.MissionControlEntityRef{EntityKind: enumtypes.MissionControlEntityKindDiscussion, EntityPublicID: "DISC-1"},
		BusinessIntentKey:         "intent-stale",
		ExpectedProjectionVersion: 6,
		Payload: valuetypes.MissionControlCommandPayload{
			DiscussionFormalize: &valuetypes.MissionControlDiscussionFormalizePayload{
				SourceEntityRef: valuetypes.MissionControlEntityRef{EntityKind: enumtypes.MissionControlEntityKindDiscussion, EntityPublicID: "DISC-1"},
				FormalizedKind:  "work_item",
				Title:           "Task from discussion",
			},
		},
		RequestedAt: now,
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if got, want := result.Command.Status, enumtypes.MissionControlCommandStatusBlocked; got != want {
		t.Fatalf("command status = %s, want %s", got, want)
	}
	if got, want := result.Command.FailureReason, enumtypes.MissionControlCommandFailureReasonProjectionStale; got != want {
		t.Fatalf("failure reason = %s, want %s", got, want)
	}
	if len(events.items) != 1 || events.items[0].EventType != eventTypeMissionControlCommandBlocked {
		t.Fatalf("unexpected events: %+v", events.items)
	}
}

func TestSubmitCommandFormalizeUsesPayloadSourceAsEffectiveTarget(t *testing.T) {
	t.Parallel()

	svc, repo, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	sourceDiscussion, err := repo.UpsertEntity(context.Background(), missioncontrolrepo.UpsertEntityParams{
		ProjectID:         "proj-1",
		EntityKind:        enumtypes.MissionControlEntityKindDiscussion,
		EntityExternalKey: "DISC-1",
		Title:             "Discussion",
		ProjectionVersion: 5,
		ProjectedAt:       now,
	})
	if err != nil {
		t.Fatalf("seed entity: %v", err)
	}

	result, err := svc.SubmitCommand(context.Background(), SubmitCommandParams{
		ProjectID:                 "proj-1",
		ActorID:                   "owner",
		CorrelationID:             "corr-formalize",
		CommandKind:               enumtypes.MissionControlCommandKindDiscussionFormalize,
		BusinessIntentKey:         "intent-formalize",
		ExpectedProjectionVersion: 5,
		Payload: valuetypes.MissionControlCommandPayload{
			DiscussionFormalize: &valuetypes.MissionControlDiscussionFormalizePayload{
				SourceEntityRef: valuetypes.MissionControlEntityRef{
					EntityKind:     enumtypes.MissionControlEntityKindDiscussion,
					EntityPublicID: "DISC-1",
				},
				FormalizedKind: "work_item",
				Title:          "Task from discussion",
			},
		},
		RequestedAt: now,
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if got, want := result.TargetEntity.ID, sourceDiscussion.ID; got != want {
		t.Fatalf("target entity id = %d, want %d", got, want)
	}
	if got, want := result.EntityRefs[0].EntityPublicID, "DISC-1"; got != want {
		t.Fatalf("entity ref public id = %s, want %s", got, want)
	}
}

func TestSubmitCommandRejectsFormalizeTargetMismatch(t *testing.T) {
	t.Parallel()

	svc, repo, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	_, err := repo.UpsertEntity(context.Background(), missioncontrolrepo.UpsertEntityParams{
		ProjectID:         "proj-1",
		EntityKind:        enumtypes.MissionControlEntityKindDiscussion,
		EntityExternalKey: "DISC-1",
		Title:             "Discussion",
		ProjectionVersion: 5,
		ProjectedAt:       now,
	})
	if err != nil {
		t.Fatalf("seed source entity: %v", err)
	}
	_, err = repo.UpsertEntity(context.Background(), missioncontrolrepo.UpsertEntityParams{
		ProjectID:         "proj-1",
		EntityKind:        enumtypes.MissionControlEntityKindDiscussion,
		EntityExternalKey: "DISC-2",
		Title:             "Other discussion",
		ProjectionVersion: 3,
		ProjectedAt:       now,
	})
	if err != nil {
		t.Fatalf("seed target entity: %v", err)
	}

	_, err = svc.SubmitCommand(context.Background(), SubmitCommandParams{
		ProjectID:                 "proj-1",
		ActorID:                   "owner",
		CorrelationID:             "corr-formalize-mismatch",
		CommandKind:               enumtypes.MissionControlCommandKindDiscussionFormalize,
		TargetEntityRef:           &valuetypes.MissionControlEntityRef{EntityKind: enumtypes.MissionControlEntityKindDiscussion, EntityPublicID: "DISC-2"},
		BusinessIntentKey:         "intent-formalize-mismatch",
		ExpectedProjectionVersion: 5,
		Payload: valuetypes.MissionControlCommandPayload{
			DiscussionFormalize: &valuetypes.MissionControlDiscussionFormalizePayload{
				SourceEntityRef: valuetypes.MissionControlEntityRef{
					EntityKind:     enumtypes.MissionControlEntityKindDiscussion,
					EntityPublicID: "DISC-1",
				},
				FormalizedKind: "work_item",
				Title:          "Task from discussion",
			},
		},
		RequestedAt: now,
	})
	var validation errs.Validation
	if !errors.As(err, &validation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if got, want := validation.Field, "target_entity_ref"; got != want {
		t.Fatalf("validation field = %s, want %s", got, want)
	}
}

func TestQueueCommandIsIdempotentAfterPendingSync(t *testing.T) {
	t.Parallel()

	svc, repo, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	command := seedCommandForTransitionTest(t, repo, "proj-1", now)
	if _, err := svc.MarkCommandPendingSync(context.Background(), CommandSyncProgressParams{
		ProjectID:           "proj-1",
		CommandID:           command.ID,
		ProviderDeliveryIDs: []string{"delivery-1"},
		UpdatedAt:           now.Add(time.Minute),
	}); err != nil {
		t.Fatalf("MarkCommandPendingSync() error = %v", err)
	}

	got, err := svc.QueueCommand(context.Background(), CommandQueueParams{
		ProjectID: "proj-1",
		CommandID: command.ID,
		UpdatedAt: now.Add(2 * time.Minute),
	})
	if err != nil {
		t.Fatalf("QueueCommand() error = %v", err)
	}
	if got.Status != enumtypes.MissionControlCommandStatusPendingSync {
		t.Fatalf("status = %s, want pending_sync", got.Status)
	}
}

func TestMarkCommandReconciledIsIdempotentForDuplicateDelivery(t *testing.T) {
	t.Parallel()

	svc, repo, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	command := seedCommandForTransitionTest(t, repo, "proj-1", now)
	if _, err := svc.MarkCommandPendingSync(context.Background(), CommandSyncProgressParams{
		ProjectID:           "proj-1",
		CommandID:           command.ID,
		ProviderDeliveryIDs: []string{"delivery-1"},
		UpdatedAt:           now.Add(time.Minute),
	}); err != nil {
		t.Fatalf("MarkCommandPendingSync() error = %v", err)
	}
	if _, err := svc.MarkCommandReconciled(context.Background(), CommandReconcileParams{
		ProjectID:           "proj-1",
		CommandID:           command.ID,
		ProviderDeliveryIDs: []string{"delivery-1"},
		UpdatedAt:           now.Add(2 * time.Minute),
		ReconciledAt:        now.Add(2 * time.Minute),
	}); err != nil {
		t.Fatalf("MarkCommandReconciled() error = %v", err)
	}

	got, err := svc.MarkCommandReconciled(context.Background(), CommandReconcileParams{
		ProjectID:           "proj-1",
		CommandID:           command.ID,
		ProviderDeliveryIDs: []string{"delivery-1"},
		UpdatedAt:           now.Add(3 * time.Minute),
		ReconciledAt:        now.Add(3 * time.Minute),
	})
	if err != nil {
		t.Fatalf("duplicate MarkCommandReconciled() error = %v", err)
	}
	if got.Status != enumtypes.MissionControlCommandStatusReconciled {
		t.Fatalf("status = %s, want reconciled", got.Status)
	}
}

func TestMarkCommandFailedIsIdempotentForDuplicateDelivery(t *testing.T) {
	t.Parallel()

	svc, repo, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	command := seedCommandForTransitionTest(t, repo, "proj-1", now)
	if _, err := svc.MarkCommandFailed(context.Background(), CommandFailureParams{
		ProjectID:           "proj-1",
		CommandID:           command.ID,
		FailureReason:       enumtypes.MissionControlCommandFailureReasonProviderError,
		ProviderDeliveryIDs: []string{"delivery-1"},
		UpdatedAt:           now.Add(time.Minute),
	}); err != nil {
		t.Fatalf("MarkCommandFailed() error = %v", err)
	}

	got, err := svc.MarkCommandFailed(context.Background(), CommandFailureParams{
		ProjectID:           "proj-1",
		CommandID:           command.ID,
		FailureReason:       enumtypes.MissionControlCommandFailureReasonProviderError,
		ProviderDeliveryIDs: []string{"delivery-1"},
		UpdatedAt:           now.Add(2 * time.Minute),
	})
	if err != nil {
		t.Fatalf("duplicate MarkCommandFailed() error = %v", err)
	}
	if got.Status != enumtypes.MissionControlCommandStatusFailed {
		t.Fatalf("status = %s, want failed", got.Status)
	}
}

func TestSubmitCommandDedupesBusinessIntent(t *testing.T) {
	t.Parallel()

	svc, _, events, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})

	params := SubmitCommandParams{
		ProjectID:         "proj-1",
		ActorID:           "owner",
		CorrelationID:     "corr-dup-1",
		CommandKind:       enumtypes.MissionControlCommandKindDiscussionCreate,
		BusinessIntentKey: "intent-dup",
		Payload: valuetypes.MissionControlCommandPayload{
			DiscussionCreate: &valuetypes.MissionControlDiscussionCreatePayload{
				Title: "New discussion",
			},
		},
		RequestedAt: now,
	}
	first, err := svc.SubmitCommand(context.Background(), params)
	if err != nil {
		t.Fatalf("first SubmitCommand() error = %v", err)
	}

	params.CorrelationID = "corr-dup-2"
	_, err = svc.SubmitCommand(context.Background(), params)
	if err == nil {
		t.Fatal("expected duplicate intent error, got nil")
	}
	var duplicateErr DuplicateIntentError
	if !errors.As(err, &duplicateErr) {
		t.Fatalf("expected DuplicateIntentError, got %T", err)
	}
	if duplicateErr.ExistingCommand.ID != first.Command.ID {
		t.Fatalf("existing command id = %s, want %s", duplicateErr.ExistingCommand.ID, first.Command.ID)
	}
	if len(events.items) != 2 || events.items[1].EventType != eventTypeMissionControlCommandDeduped {
		t.Fatalf("unexpected events: %+v", events.items)
	}
}

func TestCommandLifecycleTransitions(t *testing.T) {
	t.Parallel()

	svc, _, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})

	admission, err := svc.SubmitCommand(context.Background(), SubmitCommandParams{
		ProjectID:         "proj-1",
		ActorID:           "owner",
		CorrelationID:     "corr-life",
		CommandKind:       enumtypes.MissionControlCommandKindDiscussionCreate,
		BusinessIntentKey: "intent-life",
		Payload: valuetypes.MissionControlCommandPayload{
			DiscussionCreate: &valuetypes.MissionControlDiscussionCreatePayload{
				Title: "Life cycle discussion",
			},
		},
		RequestedAt: now,
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}

	queued, err := svc.QueueCommand(context.Background(), CommandQueueParams{
		ProjectID: "proj-1",
		CommandID: admission.Command.ID,
		UpdatedAt: now.Add(1 * time.Minute),
	})
	if err != nil {
		t.Fatalf("QueueCommand() error = %v", err)
	}
	if got, want := queued.Status, enumtypes.MissionControlCommandStatusQueued; got != want {
		t.Fatalf("queued status = %s, want %s", got, want)
	}

	pendingSync, err := svc.MarkCommandPendingSync(context.Background(), CommandSyncProgressParams{
		ProjectID:           "proj-1",
		CommandID:           admission.Command.ID,
		ProviderDeliveryIDs: []string{"delivery-1"},
		UpdatedAt:           now.Add(2 * time.Minute),
	})
	if err != nil {
		t.Fatalf("MarkCommandPendingSync() error = %v", err)
	}
	if got, want := pendingSync.Status, enumtypes.MissionControlCommandStatusPendingSync; got != want {
		t.Fatalf("pending_sync status = %s, want %s", got, want)
	}

	reconciled, err := svc.MarkCommandReconciled(context.Background(), CommandReconcileParams{
		ProjectID:           "proj-1",
		CommandID:           admission.Command.ID,
		ProviderDeliveryIDs: []string{"delivery-1"},
		ReconciledAt:        now.Add(3 * time.Minute),
		UpdatedAt:           now.Add(3 * time.Minute),
	})
	if err != nil {
		t.Fatalf("MarkCommandReconciled() error = %v", err)
	}
	if got, want := reconciled.Status, enumtypes.MissionControlCommandStatusReconciled; got != want {
		t.Fatalf("reconciled status = %s, want %s", got, want)
	}
	if reconciled.ReconciledAt == nil {
		t.Fatal("reconciled_at must be set")
	}

	queuedAgain, err := svc.QueueCommand(context.Background(), CommandQueueParams{
		ProjectID: "proj-1",
		CommandID: admission.Command.ID,
		UpdatedAt: now.Add(4 * time.Minute),
	})
	if err != nil {
		t.Fatalf("QueueCommand() after reconcile error = %v", err)
	}
	if got, want := queuedAgain.Status, enumtypes.MissionControlCommandStatusReconciled; got != want {
		t.Fatalf("queue after reconcile status = %s, want %s", got, want)
	}
}

func TestApplyApprovalDecisionQueuesPendingCommand(t *testing.T) {
	t.Parallel()

	svc, repo, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	_, err := repo.UpsertEntity(context.Background(), missioncontrolrepo.UpsertEntityParams{
		ProjectID:         "proj-1",
		EntityKind:        enumtypes.MissionControlEntityKindDiscussion,
		EntityExternalKey: "DISC-1",
		Title:             "Discussion",
		ProjectionVersion: 3,
		ProjectedAt:       now,
	})
	if err != nil {
		t.Fatalf("seed entity: %v", err)
	}

	admission, err := svc.SubmitCommand(context.Background(), SubmitCommandParams{
		ProjectID:                 "proj-1",
		ActorID:                   "owner",
		CorrelationID:             "corr-approval",
		CommandKind:               enumtypes.MissionControlCommandKindStageNextStep,
		TargetEntityRef:           &valuetypes.MissionControlEntityRef{EntityKind: enumtypes.MissionControlEntityKindDiscussion, EntityPublicID: "DISC-1"},
		BusinessIntentKey:         "intent-approval",
		ExpectedProjectionVersion: 3,
		Payload: valuetypes.MissionControlCommandPayload{
			StageNextStep: &valuetypes.MissionControlStageNextStepExecutePayload{
				ThreadKind:          "issue",
				ThreadNumber:        370,
				TargetLabel:         "run:qa",
				ApprovalRequirement: enumtypes.MissionControlApprovalRequirementOwnerReview,
			},
		},
		RequestedAt: now,
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}

	approved, err := svc.ApplyApprovalDecision(context.Background(), ApprovalDecisionParams{
		ProjectID:       "proj-1",
		CommandID:       admission.Command.ID,
		Decision:        enumtypes.MissionControlApprovalStateApproved,
		ApproverActorID: "owner",
		StatusMessage:   "approved",
		UpdatedAt:       now.Add(1 * time.Minute),
	})
	if err != nil {
		t.Fatalf("ApplyApprovalDecision() error = %v", err)
	}
	if got, want := approved.Status, enumtypes.MissionControlCommandStatusQueued; got != want {
		t.Fatalf("approved status = %s, want %s", got, want)
	}
	if got, want := approved.ApprovalState, enumtypes.MissionControlApprovalStateApproved; got != want {
		t.Fatalf("approval state = %s, want %s", got, want)
	}
}

func TestSubmitCommandRetrySyncRejectsNonRetryableStatus(t *testing.T) {
	t.Parallel()

	svc, _, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})

	accepted, err := svc.SubmitCommand(context.Background(), SubmitCommandParams{
		ProjectID:         "proj-1",
		ActorID:           "owner",
		CorrelationID:     "corr-source-command",
		CommandKind:       enumtypes.MissionControlCommandKindDiscussionCreate,
		BusinessIntentKey: "intent-source-command",
		Payload: valuetypes.MissionControlCommandPayload{
			DiscussionCreate: &valuetypes.MissionControlDiscussionCreatePayload{Title: "Source command"},
		},
		RequestedAt: now,
	})
	if err != nil {
		t.Fatalf("seed source command: %v", err)
	}

	_, err = svc.SubmitCommand(context.Background(), SubmitCommandParams{
		ProjectID:         "proj-1",
		ActorID:           "owner",
		CorrelationID:     "corr-retry",
		CommandKind:       enumtypes.MissionControlCommandKindRetrySync,
		BusinessIntentKey: "intent-retry",
		Payload: valuetypes.MissionControlCommandPayload{
			RetrySync: &valuetypes.MissionControlRetrySyncPayload{
				CommandID: accepted.Command.ID,
			},
		},
		RequestedAt: now.Add(1 * time.Minute),
	})
	var precondition errs.FailedPrecondition
	if !errors.As(err, &precondition) {
		t.Fatalf("expected failed precondition, got %v", err)
	}
}

func TestListActiveSetAndEntityDetails(t *testing.T) {
	t.Parallel()

	svc, repo, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	discussion, err := repo.UpsertEntity(context.Background(), missioncontrolrepo.UpsertEntityParams{
		ProjectID:         "proj-1",
		EntityKind:        enumtypes.MissionControlEntityKindDiscussion,
		EntityExternalKey: "DISC-1",
		Title:             "Discussion",
		ProjectionVersion: 2,
		ProjectedAt:       now,
	})
	if err != nil {
		t.Fatalf("seed discussion: %v", err)
	}
	workItem, err := repo.UpsertEntity(context.Background(), missioncontrolrepo.UpsertEntityParams{
		ProjectID:         "proj-1",
		EntityKind:        enumtypes.MissionControlEntityKindWorkItem,
		EntityExternalKey: "TASK-1",
		Title:             "Task",
		ProjectionVersion: 4,
		ProjectedAt:       now.Add(1 * time.Minute),
	})
	if err != nil {
		t.Fatalf("seed work item: %v", err)
	}
	if err := repo.ReplaceRelationsForSource(context.Background(), missioncontrolrepo.ReplaceRelationsParams{
		ProjectID:      "proj-1",
		SourceEntityID: discussion.ID,
		Relations: []missioncontrolrepo.RelationSeed{
			{
				TargetEntityID: workItem.ID,
				RelationKind:   enumtypes.MissionControlRelationKindFormalizedFrom,
				SourceKind:     enumtypes.MissionControlRelationSourceKindPlatform,
			},
		},
	}); err != nil {
		t.Fatalf("seed relations: %v", err)
	}
	if _, err := repo.UpsertTimelineEntry(context.Background(), missioncontrolrepo.UpsertTimelineEntryParams{
		ProjectID:        "proj-1",
		EntityID:         discussion.ID,
		SourceKind:       enumtypes.MissionControlTimelineSourceKindPlatform,
		EntryExternalKey: "timeline-1",
		Summary:          "created",
		OccurredAt:       now.Add(2 * time.Minute),
	}); err != nil {
		t.Fatalf("seed timeline: %v", err)
	}

	activeSet, err := svc.ListActiveSet(context.Background(), ActiveSetQuery{ProjectID: "proj-1"})
	if err != nil {
		t.Fatalf("ListActiveSet() error = %v", err)
	}
	if got, want := len(activeSet.Entities), 2; got != want {
		t.Fatalf("entity count = %d, want %d", got, want)
	}
	if got, want := len(activeSet.Relations), 1; got != want {
		t.Fatalf("relation count = %d, want %d", got, want)
	}
	if got, want := activeSet.Relations[0].SourceEntityRef.EntityPublicID, "DISC-1"; got != want {
		t.Fatalf("active-set source public id = %s, want %s", got, want)
	}
	if got, want := activeSet.Relations[0].TargetEntityRef.EntityPublicID, "TASK-1"; got != want {
		t.Fatalf("active-set target public id = %s, want %s", got, want)
	}

	details, err := svc.GetEntityDetails(context.Background(), EntityDetailsQuery{
		ProjectID:      "proj-1",
		EntityKind:     enumtypes.MissionControlEntityKindDiscussion,
		EntityPublicID: "DISC-1",
	})
	if err != nil {
		t.Fatalf("GetEntityDetails() error = %v", err)
	}
	if got, want := details.Entity.EntityExternalKey, "DISC-1"; got != want {
		t.Fatalf("entity public id = %s, want %s", got, want)
	}
	if got, want := len(details.Relations), 1; got != want {
		t.Fatalf("details relation count = %d, want %d", got, want)
	}
	if got, want := details.Relations[0].SourceEntityRef.EntityPublicID, "DISC-1"; got != want {
		t.Fatalf("details source public id = %s, want %s", got, want)
	}
	if got, want := details.Relations[0].TargetEntityRef.EntityPublicID, "TASK-1"; got != want {
		t.Fatalf("details target public id = %s, want %s", got, want)
	}
	if got, want := len(details.Timeline), 1; got != want {
		t.Fatalf("details timeline count = %d, want %d", got, want)
	}
}

func TestReadPathQueriesStayAvailableWithoutWriteSideFlags(t *testing.T) {
	t.Parallel()

	svc, _, _, _ := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})

	_, err := svc.ListActiveSet(context.Background(), ActiveSetQuery{ProjectID: "proj-1"})
	if err != nil {
		t.Fatalf("ListActiveSet() error = %v", err)
	}
}

type workspaceGraphSeedOptions struct {
	SkipPullRequest          bool
	PullRequestCoverageClass enumtypes.MissionControlCoverageClass
}

func seedWorkspaceProjectionGraph(
	t *testing.T,
	repo *inMemoryRepository,
	now time.Time,
	opts workspaceGraphSeedOptions,
) (Entity, Entity, Entity) {
	t.Helper()

	if opts.SkipPullRequest {
		opts.PullRequestCoverageClass = ""
	}
	if !opts.SkipPullRequest || opts.PullRequestCoverageClass != "" {
		if opts.PullRequestCoverageClass == "" {
			opts.PullRequestCoverageClass = enumtypes.MissionControlCoverageClassOpenPrimary
		}
	}

	workItemPayload := mustMarshalPayload(t, valuetypes.MissionControlWorkItemProjectionPayload{
		RepositoryFullName: "repo",
		IssueNumber:        372,
		StageLabel:         "run:design",
		Labels:             []string{"run:design"},
	})
	workItemEntity, err := repo.UpsertEntity(context.Background(), missioncontrolrepo.UpsertEntityParams{
		ProjectID:         "proj-1",
		EntityKind:        enumtypes.MissionControlEntityKindWorkItem,
		EntityExternalKey: "repo#372",
		ProviderKind:      enumtypes.MissionControlProviderKindGitHub,
		Title:             "Issue 372",
		ActiveState:       enumtypes.MissionControlActiveStateWorking,
		SyncStatus:        enumtypes.MissionControlSyncStatusSynced,
		CoverageClass:     enumtypes.MissionControlCoverageClassOpenPrimary,
		ProjectionVersion: 11,
		CardPayloadJSON:   workItemPayload,
		DetailPayloadJSON: workItemPayload,
		ProjectedAt:       now,
		StaleAfter:        timePointerForTest(now.Add(24 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("seed work item entity: %v", err)
	}

	runPayload := mustMarshalPayload(t, valuetypes.MissionControlRunProjectionPayload{
		RunID:              "run-18",
		RepositoryFullName: "repo",
		StageLabel:         "run:design",
		IssueRef:           "repo#372",
		LastStatus:         "succeeded",
	})
	runEntity, err := repo.UpsertEntity(context.Background(), missioncontrolrepo.UpsertEntityParams{
		ProjectID:         "proj-1",
		EntityKind:        enumtypes.MissionControlEntityKindRun,
		EntityExternalKey: "run-18",
		ProviderKind:      enumtypes.MissionControlProviderKindPlatform,
		Title:             "Run 18",
		ActiveState:       enumtypes.MissionControlActiveStateWorking,
		SyncStatus:        enumtypes.MissionControlSyncStatusSynced,
		CoverageClass:     enumtypes.MissionControlCoverageClassOpenPrimary,
		ProjectionVersion: 12,
		CardPayloadJSON:   runPayload,
		DetailPayloadJSON: runPayload,
		ProjectedAt:       now.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("seed run entity: %v", err)
	}

	pullRequestEntity := Entity{}
	if !opts.SkipPullRequest || opts.PullRequestCoverageClass != "" {
		pullRequestPayload := mustMarshalPayload(t, valuetypes.MissionControlPullRequestProjectionPayload{
			RepositoryFullName: "repo",
			PullRequestNumber:  18,
			LinkedIssueRefs:    []string{"repo#372"},
		})
		pullRequestEntity, err = repo.UpsertEntity(context.Background(), missioncontrolrepo.UpsertEntityParams{
			ProjectID:         "proj-1",
			EntityKind:        enumtypes.MissionControlEntityKindPullRequest,
			EntityExternalKey: "repo/pull/18",
			ProviderKind:      enumtypes.MissionControlProviderKindGitHub,
			Title:             "PR 18",
			ActiveState:       enumtypes.MissionControlActiveStateReview,
			SyncStatus:        enumtypes.MissionControlSyncStatusSynced,
			CoverageClass:     opts.PullRequestCoverageClass,
			ProjectionVersion: 13,
			CardPayloadJSON:   pullRequestPayload,
			DetailPayloadJSON: pullRequestPayload,
			ProjectedAt:       now.Add(2 * time.Minute),
			StaleAfter:        timePointerForTest(now.Add(24 * time.Hour)),
		})
		if err != nil {
			t.Fatalf("seed pull request entity: %v", err)
		}
	}

	relations := []missioncontrolrepo.RelationSeed{{
		TargetEntityID: runEntity.ID,
		RelationKind:   enumtypes.MissionControlRelationKindSpawnedRun,
		SourceKind:     enumtypes.MissionControlRelationSourceKindPlatform,
	}}
	if pullRequestEntity.ID > 0 {
		if err := repo.ReplaceRelationsForSource(context.Background(), missioncontrolrepo.ReplaceRelationsParams{
			ProjectID:      "proj-1",
			SourceEntityID: runEntity.ID,
			Relations: []missioncontrolrepo.RelationSeed{{
				TargetEntityID: pullRequestEntity.ID,
				RelationKind:   enumtypes.MissionControlRelationKindProducedPullRequest,
				SourceKind:     enumtypes.MissionControlRelationSourceKindPlatform,
			}},
		}); err != nil {
			t.Fatalf("seed run relations: %v", err)
		}
	}
	if err := repo.ReplaceRelationsForSource(context.Background(), missioncontrolrepo.ReplaceRelationsParams{
		ProjectID:      "proj-1",
		SourceEntityID: workItemEntity.ID,
		Relations:      relations,
	}); err != nil {
		t.Fatalf("seed work item relations: %v", err)
	}

	return workItemEntity, runEntity, pullRequestEntity
}

func timePointerForTest(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	copied := value.UTC()
	return &copied
}

func TestRefreshWorkspaceProjectionCreatesMissingFollowUpGapAndWatermarks(t *testing.T) {
	t.Parallel()

	svc, repo, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	_, runEntity, pullRequestEntity := seedWorkspaceProjectionGraph(t, repo, now, workspaceGraphSeedOptions{})

	summary, err := svc.RefreshWorkspaceProjection(context.Background(), WorkspaceRefreshParams{
		ProjectID:     "proj-1",
		CorrelationID: "corr-refresh",
		ObservedAt:    now,
	})
	if err != nil {
		t.Fatalf("RefreshWorkspaceProjection() error = %v", err)
	}
	if summary.ReadyForReconcile {
		t.Fatal("expected reconcile gate to stay closed while missing follow-up issue gap is open")
	}
	if got, want := summary.MissingFollowUpIssueGapCount, 1; got != want {
		t.Fatalf("missing follow-up gap count = %d, want %d", got, want)
	}

	gaps, err := repo.ListContinuityGaps(context.Background(), missioncontrolrepo.ContinuityGapListFilter{
		ProjectID: "proj-1",
		Statuses:  []enumtypes.MissionControlGapStatus{enumtypes.MissionControlGapStatusOpen},
	})
	if err != nil {
		t.Fatalf("ListContinuityGaps() error = %v", err)
	}
	if got, want := len(gaps), 1; got != want {
		t.Fatalf("open gap count = %d, want %d", got, want)
	}
	if got, want := gaps[0].GapKind, enumtypes.MissionControlGapKindMissingFollowUpIssue; got != want {
		t.Fatalf("gap kind = %s, want %s", got, want)
	}
	if got, want := gaps[0].SubjectEntityID, pullRequestEntity.ID; got != want {
		t.Fatalf("gap subject entity id = %d, want %d", got, want)
	}

	watermarks, err := repo.ListLatestWorkspaceWatermarks(context.Background(), "proj-1")
	if err != nil {
		t.Fatalf("ListLatestWorkspaceWatermarks() error = %v", err)
	}
	if got, want := len(watermarks), 4; got != want {
		t.Fatalf("watermark count = %d, want %d", got, want)
	}
	launchPolicyFound := false
	for _, watermark := range watermarks {
		if watermark.WatermarkKind != enumtypes.MissionControlWorkspaceWatermarkKindLaunchPolicy {
			continue
		}
		launchPolicyFound = true
		if got, want := watermark.Status, enumtypes.MissionControlWorkspaceWatermarkStatusDegraded; got != want {
			t.Fatalf("launch policy watermark status = %s, want %s", got, want)
		}
	}
	if !launchPolicyFound {
		t.Fatal("expected launch policy watermark to be recorded")
	}
	if _, ok := repo.entitiesByKey[entityKey("proj-1", enumtypes.MissionControlEntityKindRun, runEntity.EntityExternalKey)]; !ok {
		t.Fatal("expected run entity to remain in projection after refresh")
	}
}

func TestGetWorkspaceIncludesSecondaryRecentClosedContext(t *testing.T) {
	t.Parallel()

	svc, repo, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	workItemEntity, _, _ := seedWorkspaceProjectionGraph(t, repo, now, workspaceGraphSeedOptions{
		PullRequestCoverageClass: enumtypes.MissionControlCoverageClassRecentClosedContext,
	})
	if _, err := svc.RefreshWorkspaceProjection(context.Background(), WorkspaceRefreshParams{
		ProjectID:  "proj-1",
		ObservedAt: now,
	}); err != nil {
		t.Fatalf("RefreshWorkspaceProjection() error = %v", err)
	}

	workspace, err := svc.GetWorkspace(context.Background(), WorkspaceQuery{
		ProjectID:   "proj-1",
		StatePreset: enumtypes.MissionControlWorkspaceStatePresetWorking,
		RootLimit:   10,
	})
	if err != nil {
		t.Fatalf("GetWorkspace() error = %v", err)
	}
	if got, want := workspace.Summary.RootCount, 1; got != want {
		t.Fatalf("root count = %d, want %d", got, want)
	}
	visibilityByPublicID := make(map[string]enumtypes.MissionControlWorkspaceVisibilityTier, len(workspace.Nodes))
	for _, node := range workspace.Nodes {
		visibilityByPublicID[node.NodeRef.EntityPublicID] = node.VisibilityTier
	}
	if got, want := visibilityByPublicID[workItemEntity.EntityExternalKey], enumtypes.MissionControlWorkspaceVisibilityTierPrimary; got != want {
		t.Fatalf("work item visibility = %s, want %s", got, want)
	}
	if got, want := visibilityByPublicID["repo/pull/18"], enumtypes.MissionControlWorkspaceVisibilityTierSecondaryDimmed; got != want {
		t.Fatalf("pull request visibility = %s, want %s", got, want)
	}
}

func TestPreviewLaunchResolvesFollowUpGapWithoutCreatingCommands(t *testing.T) {
	t.Parallel()

	svc, repo, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	_, _, pullRequestEntity := seedWorkspaceProjectionGraph(t, repo, now, workspaceGraphSeedOptions{})
	if _, err := svc.RefreshWorkspaceProjection(context.Background(), WorkspaceRefreshParams{
		ProjectID:  "proj-1",
		ObservedAt: now,
	}); err != nil {
		t.Fatalf("RefreshWorkspaceProjection() error = %v", err)
	}

	preview, err := svc.PreviewLaunch(context.Background(), LaunchPreviewParams{
		ProjectID:                 "proj-1",
		NodeKind:                  enumtypes.MissionControlEntityKindPullRequest,
		NodePublicID:              pullRequestEntity.EntityExternalKey,
		ThreadKind:                "issue",
		ThreadNumber:              543,
		TargetLabel:               "run:plan",
		ExpectedProjectionVersion: pullRequestEntity.ProjectionVersion,
	})
	if err != nil {
		t.Fatalf("PreviewLaunch() error = %v", err)
	}
	if got, want := preview.BlockingReason, ""; got != want {
		t.Fatalf("blocking reason = %q, want empty", got)
	}
	if got, want := len(preview.ContinuityEffect.ResolvedGapIDs), 1; got != want {
		t.Fatalf("resolved gap count = %d, want %d", got, want)
	}
	if got, want := len(preview.ContinuityEffect.RemainingGapIDs), 0; got != want {
		t.Fatalf("remaining gap count = %d, want %d", got, want)
	}
	if got, want := len(repo.commandsByID), 0; got != want {
		t.Fatalf("preview must stay read-only, commands count = %d, want %d", got, want)
	}
}

func TestSubmitCommandBlocksStageNextStepWhenMissingPullRequestGapRemains(t *testing.T) {
	t.Parallel()

	svc, repo, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	workItemEntity, _, _ := seedWorkspaceProjectionGraph(t, repo, now, workspaceGraphSeedOptions{
		SkipPullRequest: true,
	})
	if _, err := svc.RefreshWorkspaceProjection(context.Background(), WorkspaceRefreshParams{
		ProjectID:  "proj-1",
		ObservedAt: now,
	}); err != nil {
		t.Fatalf("RefreshWorkspaceProjection() error = %v", err)
	}

	admission, err := svc.SubmitCommand(context.Background(), SubmitCommandParams{
		ProjectID:                 "proj-1",
		ActorID:                   "owner",
		CorrelationID:             "corr-stage-preview-blocked",
		CommandKind:               enumtypes.MissionControlCommandKindStageNextStep,
		TargetEntityRef:           &valuetypes.MissionControlEntityRef{EntityKind: enumtypes.MissionControlEntityKindWorkItem, EntityPublicID: workItemEntity.EntityExternalKey},
		BusinessIntentKey:         "intent-stage-preview-blocked",
		ExpectedProjectionVersion: workItemEntity.ProjectionVersion,
		Payload: valuetypes.MissionControlCommandPayload{
			StageNextStep: &valuetypes.MissionControlStageNextStepExecutePayload{
				ThreadKind:          "issue",
				ThreadNumber:        543,
				TargetLabel:         "run:plan",
				ApprovalRequirement: enumtypes.MissionControlApprovalRequirementNone,
			},
		},
		RequestedAt: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	if got, want := admission.Command.Status, enumtypes.MissionControlCommandStatusBlocked; got != want {
		t.Fatalf("command status = %s, want %s", got, want)
	}
	if got, want := admission.Command.FailureReason, enumtypes.MissionControlCommandFailureReasonPolicyDenied; got != want {
		t.Fatalf("failure reason = %s, want %s", got, want)
	}
}

type inMemoryRepository struct {
	nextEntityID    int64
	nextRelationID  int64
	nextTimelineID  int64
	nextCommandID   int64
	nextWatermarkID int64

	entitiesByKey          map[string]Entity
	relationsByID          map[int64]Relation
	timelineByCompositeKey map[string]TimelineEntry
	continuityGapsByKey    map[string]missioncontrolrepo.ContinuityGap
	workspaceWatermarks    []missioncontrolrepo.WorkspaceWatermark
	commandsByID           map[string]Command
	commandIDByIntent      map[string]string
}

func newInMemoryRepository() *inMemoryRepository {
	return &inMemoryRepository{
		entitiesByKey:          make(map[string]Entity),
		relationsByID:          make(map[int64]Relation),
		timelineByCompositeKey: make(map[string]TimelineEntry),
		continuityGapsByKey:    make(map[string]missioncontrolrepo.ContinuityGap),
		commandsByID:           make(map[string]Command),
		commandIDByIntent:      make(map[string]string),
	}
}

func (r *inMemoryRepository) UpsertEntity(_ context.Context, params missioncontrolrepo.UpsertEntityParams) (Entity, error) {
	key := entityKey(params.ProjectID, params.EntityKind, params.EntityExternalKey)
	existing, found := r.entitiesByKey[key]
	if !found {
		r.nextEntityID++
		existing = Entity{
			ID:        r.nextEntityID,
			ProjectID: strings.TrimSpace(params.ProjectID),
			CreatedAt: nowOr(params.ProjectedAt),
		}
	}
	existing.EntityKind = params.EntityKind
	existing.EntityExternalKey = strings.TrimSpace(params.EntityExternalKey)
	existing.ProviderKind = params.ProviderKind
	existing.ProviderURL = strings.TrimSpace(params.ProviderURL)
	existing.Title = strings.TrimSpace(params.Title)
	existing.ActiveState = params.ActiveState
	existing.SyncStatus = params.SyncStatus
	existing.ContinuityStatus = params.ContinuityStatus
	existing.CoverageClass = params.CoverageClass
	existing.ProjectionVersion = params.ProjectionVersion
	if existing.ProjectionVersion <= 0 {
		existing.ProjectionVersion = 1
	}
	existing.CardPayloadJSON = params.CardPayloadJSON
	existing.DetailPayloadJSON = params.DetailPayloadJSON
	existing.LastTimelineAt = params.LastTimelineAt
	existing.ProviderUpdatedAt = params.ProviderUpdatedAt
	existing.ProjectedAt = nowOr(params.ProjectedAt)
	existing.StaleAfter = params.StaleAfter
	existing.UpdatedAt = existing.ProjectedAt
	r.entitiesByKey[key] = existing
	return existing, nil
}

func (r *inMemoryRepository) UpdateEntityProjection(_ context.Context, params missioncontrolrepo.UpdateEntityParams) (Entity, error) {
	key := entityKey(params.ProjectID, params.EntityKind, params.EntityExternalKey)
	existing, found := r.entitiesByKey[key]
	if !found {
		return Entity{}, errs.NotFound{Msg: "mission control entity not found"}
	}
	if existing.ProjectionVersion != params.ExpectedProjectionVersion {
		return Entity{}, missioncontrolrepo.ProjectionVersionConflict{
			ProjectID:                 params.ProjectID,
			EntityKind:                params.EntityKind,
			EntityExternalKey:         params.EntityExternalKey,
			ExpectedProjectionVersion: params.ExpectedProjectionVersion,
			ActualProjectionVersion:   existing.ProjectionVersion,
		}
	}
	existing.ProviderURL = strings.TrimSpace(params.ProviderURL)
	existing.Title = strings.TrimSpace(params.Title)
	existing.ActiveState = params.ActiveState
	existing.SyncStatus = params.SyncStatus
	existing.ContinuityStatus = params.ContinuityStatus
	existing.CoverageClass = params.CoverageClass
	existing.CardPayloadJSON = params.CardPayloadJSON
	existing.DetailPayloadJSON = params.DetailPayloadJSON
	existing.LastTimelineAt = params.LastTimelineAt
	existing.ProviderUpdatedAt = params.ProviderUpdatedAt
	existing.ProjectedAt = nowOr(params.ProjectedAt)
	existing.StaleAfter = params.StaleAfter
	existing.ProjectionVersion++
	existing.UpdatedAt = existing.ProjectedAt
	r.entitiesByKey[key] = existing
	return existing, nil
}

func (r *inMemoryRepository) GetEntityByPublicID(_ context.Context, projectID string, entityKind enumtypes.MissionControlEntityKind, entityExternalKey string) (Entity, bool, error) {
	entity, found := r.entitiesByKey[entityKey(projectID, entityKind, entityExternalKey)]
	return entity, found, nil
}

func (r *inMemoryRepository) GetEntityByID(_ context.Context, projectID string, entityID int64) (Entity, bool, error) {
	projectID = strings.TrimSpace(projectID)
	for _, entity := range r.entitiesByKey {
		if entity.ProjectID == projectID && entity.ID == entityID {
			return entity, true, nil
		}
	}
	return Entity{}, false, nil
}

func (r *inMemoryRepository) ListEntities(_ context.Context, filter missioncontrolrepo.EntityListFilter) ([]Entity, error) {
	items := make([]Entity, 0, len(r.entitiesByKey))
	for _, entity := range r.entitiesByKey {
		if entity.ProjectID != strings.TrimSpace(filter.ProjectID) {
			continue
		}
		if len(filter.ActiveStates) > 0 && !containsActiveState(filter.ActiveStates, entity.ActiveState) {
			continue
		}
		if len(filter.SyncStatuses) > 0 && !containsSyncStatus(filter.SyncStatuses, entity.SyncStatus) {
			continue
		}
		items = append(items, entity)
	}
	sort.Slice(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		leftTimeline := time.Time{}
		rightTimeline := time.Time{}
		if left.LastTimelineAt != nil {
			leftTimeline = *left.LastTimelineAt
		}
		if right.LastTimelineAt != nil {
			rightTimeline = *right.LastTimelineAt
		}
		if !leftTimeline.Equal(rightTimeline) {
			return leftTimeline.After(rightTimeline)
		}
		if !left.ProjectedAt.Equal(right.ProjectedAt) {
			return left.ProjectedAt.After(right.ProjectedAt)
		}
		return left.ID > right.ID
	})
	if filter.Limit > 0 && len(items) > filter.Limit {
		items = items[:filter.Limit]
	}
	return items, nil
}

func (r *inMemoryRepository) ReplaceRelationsForSource(_ context.Context, params missioncontrolrepo.ReplaceRelationsParams) error {
	for id, relation := range r.relationsByID {
		if relation.ProjectID == params.ProjectID && relation.SourceEntityID == params.SourceEntityID {
			delete(r.relationsByID, id)
		}
	}
	for _, relation := range params.Relations {
		r.nextRelationID++
		r.relationsByID[r.nextRelationID] = Relation{
			ID:             r.nextRelationID,
			ProjectID:      params.ProjectID,
			SourceEntityID: params.SourceEntityID,
			RelationKind:   relation.RelationKind,
			TargetEntityID: relation.TargetEntityID,
			SourceKind:     relation.SourceKind,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
	}
	return nil
}

func (r *inMemoryRepository) ListRelationsForEntity(_ context.Context, projectID string, entityID int64) ([]Relation, error) {
	items := make([]Relation, 0)
	for _, relation := range r.relationsByID {
		if relation.ProjectID != strings.TrimSpace(projectID) {
			continue
		}
		if relation.SourceEntityID == entityID || relation.TargetEntityID == entityID {
			items = append(items, relation)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID > items[j].ID })
	return items, nil
}

func (r *inMemoryRepository) UpsertTimelineEntry(_ context.Context, params missioncontrolrepo.UpsertTimelineEntryParams) (TimelineEntry, error) {
	key := fmt.Sprintf("%s/%s/%s", params.ProjectID, params.SourceKind, params.EntryExternalKey)
	entry, found := r.timelineByCompositeKey[key]
	if !found {
		r.nextTimelineID++
		entry.ID = r.nextTimelineID
		entry.CreatedAt = nowOr(params.OccurredAt)
	}
	entry.ProjectID = params.ProjectID
	entry.EntityID = params.EntityID
	entry.SourceKind = params.SourceKind
	entry.EntryExternalKey = params.EntryExternalKey
	entry.CommandID = params.CommandID
	entry.Summary = params.Summary
	entry.BodyMarkdown = params.BodyMarkdown
	entry.PayloadJSON = params.PayloadJSON
	entry.OccurredAt = nowOr(params.OccurredAt)
	entry.ProviderURL = params.ProviderURL
	entry.IsReadOnly = params.IsReadOnly
	r.timelineByCompositeKey[key] = entry
	return entry, nil
}

func (r *inMemoryRepository) ListTimelineEntries(_ context.Context, filter missioncontrolrepo.TimelineListFilter) ([]TimelineEntry, error) {
	items := make([]TimelineEntry, 0)
	for _, entry := range r.timelineByCompositeKey {
		if entry.ProjectID == strings.TrimSpace(filter.ProjectID) && entry.EntityID == filter.EntityID {
			items = append(items, entry)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if !items[i].OccurredAt.Equal(items[j].OccurredAt) {
			return items[i].OccurredAt.After(items[j].OccurredAt)
		}
		return items[i].ID > items[j].ID
	})
	if filter.Limit > 0 && len(items) > filter.Limit {
		items = items[:filter.Limit]
	}
	return items, nil
}

func (r *inMemoryRepository) ListContinuityGaps(_ context.Context, filter missioncontrolrepo.ContinuityGapListFilter) ([]missioncontrolrepo.ContinuityGap, error) {
	items := make([]missioncontrolrepo.ContinuityGap, 0, len(r.continuityGapsByKey))
	subjectFilter := make(map[int64]struct{}, len(filter.SubjectEntityIDs))
	for _, subjectEntityID := range filter.SubjectEntityIDs {
		subjectFilter[subjectEntityID] = struct{}{}
	}
	for _, gap := range r.continuityGapsByKey {
		if gap.ProjectID != strings.TrimSpace(filter.ProjectID) {
			continue
		}
		if len(subjectFilter) > 0 {
			if _, ok := subjectFilter[gap.SubjectEntityID]; !ok {
				continue
			}
		}
		if len(filter.Statuses) > 0 {
			match := false
			for _, status := range filter.Statuses {
				if gap.Status == status {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		items = append(items, gap)
	}
	sort.Slice(items, func(i, j int) bool {
		if !items[i].DetectedAt.Equal(items[j].DetectedAt) {
			return items[i].DetectedAt.After(items[j].DetectedAt)
		}
		return items[i].ID > items[j].ID
	})
	return items, nil
}

func (r *inMemoryRepository) SyncContinuityGaps(_ context.Context, params missioncontrolrepo.SyncContinuityGapsParams) error {
	projectID := strings.TrimSpace(params.ProjectID)
	desired := make(map[string]missioncontrolrepo.ContinuityGap, len(params.DesiredOpen))
	for _, seed := range params.DesiredOpen {
		key := continuityGapKey(projectID, seed.SubjectEntityID, seed.GapKind)
		current, found := r.continuityGapsByKey[key]
		if !found {
			current = missioncontrolrepo.ContinuityGap{
				ID:              int64(len(r.continuityGapsByKey) + len(desired) + 1),
				ProjectID:       projectID,
				SubjectEntityID: seed.SubjectEntityID,
			}
		}
		current.GapKind = seed.GapKind
		current.Severity = seed.Severity
		current.Status = enumtypes.MissionControlGapStatusOpen
		current.ExpectedEntityKind = seed.ExpectedEntityKind
		current.ExpectedStageLabel = seed.ExpectedStageLabel
		current.ResolutionHint = seed.ResolutionHint
		current.PayloadJSON = cloneBytes(seed.PayloadJSON)
		current.DetectedAt = nowOr(seed.DetectedAt)
		current.ResolvedAt = nil
		current.UpdatedAt = nowOr(seed.DetectedAt)
		desired[key] = current
	}
	for key, gap := range r.continuityGapsByKey {
		if gap.ProjectID != projectID {
			continue
		}
		if _, ok := desired[key]; ok {
			continue
		}
		if gap.Status == enumtypes.MissionControlGapStatusOpen {
			gap.Status = enumtypes.MissionControlGapStatusResolved
			resolvedAt := nowOr(params.ResolvedAt)
			gap.ResolvedAt = &resolvedAt
			gap.UpdatedAt = resolvedAt
			r.continuityGapsByKey[key] = gap
		}
	}
	for key, gap := range desired {
		r.continuityGapsByKey[key] = gap
	}
	return nil
}

func (r *inMemoryRepository) CreateWorkspaceWatermark(_ context.Context, params missioncontrolrepo.CreateWorkspaceWatermarkParams) (missioncontrolrepo.WorkspaceWatermark, error) {
	r.nextWatermarkID++
	watermark := missioncontrolrepo.WorkspaceWatermark{
		ID:              r.nextWatermarkID,
		ProjectID:       strings.TrimSpace(params.ProjectID),
		WatermarkKind:   params.WatermarkKind,
		Status:          params.Status,
		Summary:         strings.TrimSpace(params.Summary),
		WindowStartedAt: params.WindowStartedAt,
		WindowEndedAt:   params.WindowEndedAt,
		ObservedAt:      nowOr(params.ObservedAt),
		PayloadJSON:     cloneBytes(params.PayloadJSON),
		CreatedAt:       nowOr(params.ObservedAt),
	}
	r.workspaceWatermarks = append(r.workspaceWatermarks, watermark)
	return watermark, nil
}

func (r *inMemoryRepository) ListLatestWorkspaceWatermarks(_ context.Context, projectID string) ([]missioncontrolrepo.WorkspaceWatermark, error) {
	projectID = strings.TrimSpace(projectID)
	latestByKind := make(map[enumtypes.MissionControlWorkspaceWatermarkKind]missioncontrolrepo.WorkspaceWatermark)
	for _, watermark := range r.workspaceWatermarks {
		if watermark.ProjectID != projectID {
			continue
		}
		current, found := latestByKind[watermark.WatermarkKind]
		if !found || watermark.ObservedAt.After(current.ObservedAt) || (watermark.ObservedAt.Equal(current.ObservedAt) && watermark.ID > current.ID) {
			latestByKind[watermark.WatermarkKind] = watermark
		}
	}
	items := make([]missioncontrolrepo.WorkspaceWatermark, 0, len(latestByKind))
	for _, watermark := range latestByKind {
		items = append(items, watermark)
	}
	sort.Slice(items, func(i, j int) bool {
		if !items[i].ObservedAt.Equal(items[j].ObservedAt) {
			return items[i].ObservedAt.After(items[j].ObservedAt)
		}
		return items[i].ID > items[j].ID
	})
	return items, nil
}

func (r *inMemoryRepository) CreateCommand(_ context.Context, params missioncontrolrepo.CreateCommandParams) (Command, error) {
	intentKey := commandIntentKey(params.ProjectID, params.BusinessIntentKey)
	if _, exists := r.commandIDByIntent[intentKey]; exists {
		return Command{}, missioncontrolrepo.DuplicateBusinessIntent{
			ProjectID:         params.ProjectID,
			BusinessIntentKey: params.BusinessIntentKey,
		}
	}
	r.nextCommandID++
	commandID := fmt.Sprintf("cmd-%d", r.nextCommandID)
	command := Command{
		ID:                  commandID,
		ProjectID:           params.ProjectID,
		CommandKind:         params.CommandKind,
		TargetEntityID:      params.TargetEntityID,
		ActorID:             params.ActorID,
		BusinessIntentKey:   params.BusinessIntentKey,
		CorrelationID:       params.CorrelationID,
		Status:              params.Status,
		FailureReason:       params.FailureReason,
		ApprovalRequestID:   params.ApprovalRequestID,
		ApprovalState:       params.ApprovalState,
		ApprovalRequestedAt: params.ApprovalRequestedAt,
		ApprovalDecidedAt:   params.ApprovalDecidedAt,
		PayloadJSON:         params.PayloadJSON,
		ResultPayloadJSON:   params.ResultPayloadJSON,
		ProviderDeliveries:  cloneBytes(params.ProviderDeliveries),
		RequestedAt:         nowOr(params.RequestedAt),
		UpdatedAt:           nowOr(params.UpdatedAt),
		ReconciledAt:        params.ReconciledAt,
	}
	r.commandsByID[commandID] = command
	r.commandIDByIntent[intentKey] = commandID
	return command, nil
}

func (r *inMemoryRepository) GetCommandByID(_ context.Context, projectID string, commandID string) (Command, bool, error) {
	command, found := r.commandsByID[strings.TrimSpace(commandID)]
	if !found || command.ProjectID != strings.TrimSpace(projectID) {
		return Command{}, false, nil
	}
	return command, true, nil
}

func (r *inMemoryRepository) GetCommandByBusinessIntent(_ context.Context, projectID string, businessIntentKey string) (Command, bool, error) {
	commandID, found := r.commandIDByIntent[commandIntentKey(projectID, businessIntentKey)]
	if !found {
		return Command{}, false, nil
	}
	command := r.commandsByID[commandID]
	return command, true, nil
}

func (r *inMemoryRepository) ListCommands(_ context.Context, filter missioncontrolrepo.CommandListFilter) ([]Command, error) {
	items := make([]Command, 0, len(r.commandsByID))
	for _, command := range r.commandsByID {
		if command.ProjectID != strings.TrimSpace(filter.ProjectID) {
			continue
		}
		if len(filter.Statuses) > 0 && !containsCommandStatus(filter.Statuses, command.Status) {
			continue
		}
		items = append(items, command)
	}
	sort.Slice(items, func(i, j int) bool {
		if !items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].UpdatedAt.After(items[j].UpdatedAt)
		}
		return items[i].ID > items[j].ID
	})
	if filter.Limit > 0 && len(items) > filter.Limit {
		items = items[:filter.Limit]
	}
	return items, nil
}

func (r *inMemoryRepository) ListCommandsAll(_ context.Context, filter missioncontrolrepo.GlobalCommandListFilter) ([]Command, error) {
	items := make([]Command, 0, len(r.commandsByID))
	for _, command := range r.commandsByID {
		if len(filter.Statuses) > 0 && !containsCommandStatus(filter.Statuses, command.Status) {
			continue
		}
		items = append(items, command)
	}
	sort.Slice(items, func(i, j int) bool {
		if !items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].UpdatedAt.After(items[j].UpdatedAt)
		}
		return items[i].ID > items[j].ID
	})
	if filter.Limit > 0 && len(items) > filter.Limit {
		items = items[:filter.Limit]
	}
	return items, nil
}

func (r *inMemoryRepository) ClaimCommandsAll(_ context.Context, params missioncontrolrepo.ClaimCommandParams) ([]Command, error) {
	now := time.Now().UTC()
	items := make([]Command, 0, len(r.commandsByID))
	for _, command := range r.commandsByID {
		if len(params.Statuses) > 0 && !containsCommandStatus(params.Statuses, command.Status) {
			continue
		}
		if command.LeaseUntil != nil && command.LeaseUntil.After(now) {
			continue
		}
		items = append(items, command)
	}
	sort.Slice(items, func(i, j int) bool {
		if !items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
			return items[i].UpdatedAt.After(items[j].UpdatedAt)
		}
		return items[i].ID > items[j].ID
	})
	if params.Limit > 0 && len(items) > params.Limit {
		items = items[:params.Limit]
	}
	claimed := make([]Command, 0, len(items))
	for _, command := range items {
		command.LeaseOwner = strings.TrimSpace(params.WorkerID)
		if params.LeaseTTL > 0 {
			leaseUntil := now.Add(params.LeaseTTL)
			command.LeaseUntil = &leaseUntil
		}
		r.commandsByID[command.ID] = command
		claimed = append(claimed, command)
	}
	return claimed, nil
}

func (r *inMemoryRepository) UpdateCommandStatus(_ context.Context, params missioncontrolrepo.UpdateCommandStatusParams) (Command, bool, error) {
	command, found := r.commandsByID[strings.TrimSpace(params.CommandID)]
	if !found || command.ProjectID != strings.TrimSpace(params.ProjectID) {
		return Command{}, false, nil
	}
	command.Status = params.Status
	if params.FailureReasonPatch.Set {
		command.FailureReason = params.FailureReasonPatch.Value
	}
	if params.ApprovalRequestIDPatch.Set {
		command.ApprovalRequestID = params.ApprovalRequestIDPatch.Value
	}
	if params.ApprovalStatePatch.Set {
		command.ApprovalState = params.ApprovalStatePatch.Value
	}
	if params.ApprovalRequestedAtPatch.Set {
		command.ApprovalRequestedAt = params.ApprovalRequestedAtPatch.Value
	}
	if params.ApprovalDecidedAtPatch.Set {
		command.ApprovalDecidedAt = params.ApprovalDecidedAtPatch.Value
	}
	if params.ResultPayloadPatch.Set {
		command.ResultPayloadJSON = cloneBytes(params.ResultPayloadPatch.Value)
	}
	if params.ProviderDeliveriesPatch.Set {
		command.ProviderDeliveries = cloneBytes(params.ProviderDeliveriesPatch.Value)
	}
	if params.LeaseOwnerPatch.Set {
		command.LeaseOwner = params.LeaseOwnerPatch.Value
	}
	if params.LeaseUntilPatch.Set {
		command.LeaseUntil = params.LeaseUntilPatch.Value
	}
	command.UpdatedAt = nowOr(params.UpdatedAt)
	if params.ReconciledAtPatch.Set {
		command.ReconciledAt = params.ReconciledAtPatch.Value
	}
	r.commandsByID[command.ID] = command
	return command, true, nil
}

func (r *inMemoryRepository) GetWarmupSummary(_ context.Context, projectID string) (WarmupSummary, error) {
	projectID = strings.TrimSpace(projectID)
	summary := WarmupSummary{ProjectID: projectID}
	for _, entity := range r.entitiesByKey {
		if entity.ProjectID != projectID {
			continue
		}
		summary.EntityCount++
		if entity.ProjectionVersion > summary.MaxProjectionVersion {
			summary.MaxProjectionVersion = entity.ProjectionVersion
		}
		if entity.EntityKind == enumtypes.MissionControlEntityKindRun {
			summary.RunEntityCount++
		}
		if entity.EntityKind == enumtypes.MissionControlEntityKindAgent {
			summary.LegacyAgentCount++
		}
	}
	for _, relation := range r.relationsByID {
		if relation.ProjectID == projectID {
			summary.RelationCount++
		}
	}
	for _, entry := range r.timelineByCompositeKey {
		if entry.ProjectID == projectID {
			summary.TimelineEntryCount++
		}
	}
	for _, command := range r.commandsByID {
		if command.ProjectID == projectID {
			summary.CommandCount++
		}
	}
	for _, gap := range r.continuityGapsByKey {
		if gap.ProjectID != projectID {
			continue
		}
		summary.ContinuityGapCount++
		if gap.Status == enumtypes.MissionControlGapStatusOpen {
			summary.OpenContinuityGapCount++
		}
		if gap.Severity == enumtypes.MissionControlGapSeverityBlocking {
			summary.BlockingGapCount++
		}
		switch gap.GapKind {
		case enumtypes.MissionControlGapKindMissingPullRequest:
			summary.MissingPullRequestGapCount++
		case enumtypes.MissionControlGapKindMissingFollowUpIssue:
			summary.MissingFollowUpIssueGapCount++
		}
	}
	for _, watermark := range r.workspaceWatermarks {
		if watermark.ProjectID == projectID {
			summary.WatermarkCount++
		}
	}
	return summary, nil
}

type flowEventRecorder struct {
	items []floweventrepo.InsertParams
}

func (r *flowEventRecorder) Insert(_ context.Context, params floweventrepo.InsertParams) error {
	r.items = append(r.items, params)
	return nil
}

func newTestService(t *testing.T, rolloutState valuetypes.MissionControlRolloutState) (*Service, *inMemoryRepository, *flowEventRecorder, time.Time) {
	t.Helper()

	repo := newInMemoryRepository()
	events := &flowEventRecorder{}
	service, err := NewService(Config{
		RolloutState: rolloutState,
	}, Dependencies{
		Repository: repo,
		FlowEvents: events,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	now := time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	return service, repo, events, now
}

func seedCommandForTransitionTest(t *testing.T, repo *inMemoryRepository, projectID string, now time.Time) Command {
	t.Helper()

	command, err := repo.CreateCommand(context.Background(), missioncontrolrepo.CreateCommandParams{
		ProjectID:         projectID,
		CommandKind:       enumtypes.MissionControlCommandKindStageNextStep,
		ActorID:           "owner",
		BusinessIntentKey: "intent-transition",
		CorrelationID:     "corr-transition",
		Status:            enumtypes.MissionControlCommandStatusAccepted,
		ApprovalState:     enumtypes.MissionControlApprovalStateNotRequired,
		PayloadJSON: mustMarshalPayload(t, valuetypes.MissionControlCommandPayload{
			StageNextStep: &valuetypes.MissionControlStageNextStepExecutePayload{
				ThreadKind:          "issue",
				ThreadNumber:        371,
				TargetLabel:         "run:qa",
				ApprovalRequirement: enumtypes.MissionControlApprovalRequirementNone,
			},
		}),
		ResultPayloadJSON: mustMarshalPayload(t, valuetypes.MissionControlCommandResultPayload{}),
		RequestedAt:       now,
		UpdatedAt:         now,
	})
	if err != nil {
		t.Fatalf("CreateCommand() error = %v", err)
	}
	return command
}

func mustMarshalPayload(t *testing.T, value any) []byte {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return raw
}

func entityKey(projectID string, entityKind enumtypes.MissionControlEntityKind, entityExternalKey string) string {
	return strings.TrimSpace(projectID) + "/" + string(entityKind) + "/" + strings.TrimSpace(entityExternalKey)
}

func commandIntentKey(projectID string, businessIntentKey string) string {
	return strings.TrimSpace(projectID) + "/" + strings.TrimSpace(businessIntentKey)
}

func continuityGapKey(projectID string, subjectEntityID int64, gapKind enumtypes.MissionControlGapKind) string {
	return fmt.Sprintf("%s/%d/%s", strings.TrimSpace(projectID), subjectEntityID, string(gapKind))
}

func cloneBytes(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	cloned := make([]byte, len(value))
	copy(cloned, value)
	return cloned
}

func nowOr(value time.Time) time.Time {
	if value.IsZero() {
		return time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC)
	}
	return value.UTC()
}

func containsActiveState(items []enumtypes.MissionControlActiveState, target enumtypes.MissionControlActiveState) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func containsSyncStatus(items []enumtypes.MissionControlSyncStatus, target enumtypes.MissionControlSyncStatus) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func containsCommandStatus(items []enumtypes.MissionControlCommandStatus, target enumtypes.MissionControlCommandStatus) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func TestGetCommandStatusDecodesResultPayload(t *testing.T) {
	t.Parallel()

	svc, _, _, now := newTestService(t, valuetypes.MissionControlRolloutState{
		SchemaReady: true,
		DomainReady: true,
	})
	admission, err := svc.SubmitCommand(context.Background(), SubmitCommandParams{
		ProjectID:         "proj-1",
		ActorID:           "owner",
		CorrelationID:     "corr-status",
		CommandKind:       enumtypes.MissionControlCommandKindDiscussionCreate,
		BusinessIntentKey: "intent-status",
		Payload: valuetypes.MissionControlCommandPayload{
			DiscussionCreate: &valuetypes.MissionControlDiscussionCreatePayload{Title: "Status command"},
		},
		RequestedAt: now,
	})
	if err != nil {
		t.Fatalf("SubmitCommand() error = %v", err)
	}
	queued, err := svc.QueueCommand(context.Background(), CommandQueueParams{
		ProjectID:     "proj-1",
		CommandID:     admission.Command.ID,
		StatusMessage: "queued",
		UpdatedAt:     now.Add(1 * time.Minute),
	})
	if err != nil {
		t.Fatalf("QueueCommand() error = %v", err)
	}
	view, err := svc.GetCommandStatus(context.Background(), "proj-1", queued.ID)
	if err != nil {
		t.Fatalf("GetCommandStatus() error = %v", err)
	}
	if got, want := view.StatusMessage, "queued"; got != want {
		t.Fatalf("status message = %q, want %q", got, want)
	}
}
