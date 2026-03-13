package worker

import (
	"context"
	"time"
)

const noopInteractionAdapterKind = "noop"

type noopInteractionLifecycleClient struct{}

func (noopInteractionLifecycleClient) ClaimNextInteractionDispatch(context.Context, time.Duration) (InteractionDispatchClaim, bool, error) {
	return InteractionDispatchClaim{}, false, nil
}

func (noopInteractionLifecycleClient) CompleteInteractionDispatch(context.Context, CompleteInteractionDispatchParams) (CompleteInteractionDispatchResult, error) {
	return CompleteInteractionDispatchResult{}, nil
}

func (noopInteractionLifecycleClient) ExpireNextInteraction(context.Context) (ExpireNextInteractionResult, error) {
	return ExpireNextInteractionResult{}, nil
}

type noopInteractionDispatcher struct{}

func (noopInteractionDispatcher) Dispatch(context.Context, InteractionDispatchClaim) (InteractionDispatchAck, error) {
	return InteractionDispatchAck{
		AdapterKind:    noopInteractionAdapterKind,
		AckPayloadJSON: []byte(`{"accepted":true,"adapter_kind":"noop"}`),
	}, nil
}
