package interactionrequest

import (
	"context"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

type (
	Request                     = entitytypes.InteractionRequest
	DeliveryAttempt             = entitytypes.InteractionDeliveryAttempt
	CallbackEvent               = entitytypes.InteractionCallbackEvent
	ResponseRecord              = entitytypes.InteractionResponseRecord
	CreateParams                = querytypes.InteractionRequestCreateParams
	CreateDeliveryAttemptParams = querytypes.InteractionDeliveryAttemptCreateParams
	ClaimDispatchParams         = querytypes.InteractionDispatchClaimParams
	DispatchClaim               = querytypes.InteractionDispatchClaim
	CompleteDispatchParams      = querytypes.InteractionDispatchCompleteParams
	CompleteDispatchResult      = querytypes.InteractionDispatchCompleteResult
	ExpireDueParams             = querytypes.InteractionExpireDueParams
	ExpireDueResult             = querytypes.InteractionExpireDueResult
	ApplyCallbackParams         = querytypes.InteractionCallbackApplyParams
	ApplyCallbackResult         = querytypes.InteractionCallbackApplyResult
)

// Repository persists interaction aggregate, delivery attempts and callback evidence.
type Repository interface {
	// Create inserts one interaction aggregate.
	Create(ctx context.Context, params CreateParams) (Request, error)
	// GetByID returns one interaction aggregate by id.
	GetByID(ctx context.Context, interactionID string) (Request, bool, error)
	// FindOpenDecisionByRunID returns open decision interaction for one run when present.
	FindOpenDecisionByRunID(ctx context.Context, runID string) (Request, bool, error)
	// ClaimNextDispatch reserves or reclaims one due dispatch attempt.
	ClaimNextDispatch(ctx context.Context, params ClaimDispatchParams) (DispatchClaim, bool, error)
	// CompleteDispatch persists one dispatch attempt outcome and applies aggregate state mutation when needed.
	CompleteDispatch(ctx context.Context, params CompleteDispatchParams) (CompleteDispatchResult, error)
	// ExpireNextDue marks one deadline-expired decision interaction terminal.
	ExpireNextDue(ctx context.Context, params ExpireDueParams) (ExpireDueResult, bool, error)
	// CreateDeliveryAttempt appends one dispatch-attempt ledger row.
	CreateDeliveryAttempt(ctx context.Context, params CreateDeliveryAttemptParams) (DeliveryAttempt, error)
	// ApplyCallback persists callback evidence, optional typed response and terminal aggregate state transition.
	ApplyCallback(ctx context.Context, params ApplyCallbackParams) (ApplyCallbackResult, error)
}
