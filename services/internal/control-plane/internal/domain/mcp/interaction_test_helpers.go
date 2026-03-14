package mcp

import (
	"context"
	"time"

	interactionrequestrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/interactionrequest"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type interactionTestRepository struct {
	ensureBindingParams interactionrequestrepo.EnsureChannelBindingParams
	upsertHandleParams  interactionrequestrepo.UpsertCallbackHandlesParams
}

func (r *interactionTestRepository) Create(context.Context, interactionrequestrepo.CreateParams) (interactionrequestrepo.Request, error) {
	return entitytypes.InteractionRequest{}, nil
}

func (r *interactionTestRepository) GetByID(context.Context, string) (interactionrequestrepo.Request, bool, error) {
	return entitytypes.InteractionRequest{}, false, nil
}

func (r *interactionTestRepository) FindOpenDecisionByRunID(context.Context, string) (interactionrequestrepo.Request, bool, error) {
	return entitytypes.InteractionRequest{}, false, nil
}

func (r *interactionTestRepository) EnsureChannelBinding(_ context.Context, params interactionrequestrepo.EnsureChannelBindingParams) (interactionrequestrepo.ChannelBinding, error) {
	r.ensureBindingParams = params
	return entitytypes.InteractionChannelBinding{
		ID:                11,
		InteractionID:     params.InteractionID,
		AdapterKind:       params.AdapterKind,
		RecipientRef:      params.RecipientRef,
		EditCapability:    enumtypes.InteractionEditCapabilityEditable,
		ContinuationState: enumtypes.InteractionContinuationStatePendingPrimaryDelivery,
		CreatedAt:         time.Date(2026, time.March, 13, 16, 0, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2026, time.March, 13, 16, 0, 0, 0, time.UTC),
	}, nil
}

func (r *interactionTestRepository) UpsertCallbackHandles(_ context.Context, params interactionrequestrepo.UpsertCallbackHandlesParams) ([]interactionrequestrepo.CallbackHandle, error) {
	r.upsertHandleParams = params
	items := make([]interactionrequestrepo.CallbackHandle, 0, len(params.Items))
	for idx, item := range params.Items {
		items = append(items, entitytypes.InteractionCallbackHandle{
			ID:                 int64(idx + 1),
			InteractionID:      params.InteractionID,
			ChannelBindingID:   params.ChannelBindingID,
			HandleHash:         item.HandleHash,
			HandleKind:         item.HandleKind,
			OptionID:           item.OptionID,
			State:              enumtypes.InteractionCallbackHandleStateOpen,
			ResponseDeadlineAt: item.ResponseDeadlineAt,
			GraceExpiresAt:     item.GraceExpiresAt,
			CreatedAt:          time.Date(2026, time.March, 13, 16, 0, 0, 0, time.UTC),
		})
	}
	return items, nil
}

func (r *interactionTestRepository) ClaimNextDispatch(context.Context, interactionrequestrepo.ClaimDispatchParams) (interactionrequestrepo.DispatchClaim, bool, error) {
	return querytypes.InteractionDispatchClaim{}, false, nil
}

func (r *interactionTestRepository) CompleteDispatch(context.Context, interactionrequestrepo.CompleteDispatchParams) (interactionrequestrepo.CompleteDispatchResult, error) {
	return querytypes.InteractionDispatchCompleteResult{}, nil
}

func (r *interactionTestRepository) UpdateDispatchBinding(context.Context, interactionrequestrepo.UpdateDispatchBindingParams) (interactionrequestrepo.ChannelBinding, error) {
	return entitytypes.InteractionChannelBinding{}, nil
}

func (r *interactionTestRepository) ExpireNextDue(context.Context, interactionrequestrepo.ExpireDueParams) (interactionrequestrepo.ExpireDueResult, bool, error) {
	return querytypes.InteractionExpireDueResult{}, false, nil
}

func (r *interactionTestRepository) CreateDeliveryAttempt(context.Context, interactionrequestrepo.CreateDeliveryAttemptParams) (interactionrequestrepo.DeliveryAttempt, error) {
	return entitytypes.InteractionDeliveryAttempt{}, nil
}

func (r *interactionTestRepository) ApplyCallback(context.Context, interactionrequestrepo.ApplyCallbackParams) (interactionrequestrepo.ApplyCallbackResult, error) {
	return querytypes.InteractionCallbackApplyResult{}, nil
}
