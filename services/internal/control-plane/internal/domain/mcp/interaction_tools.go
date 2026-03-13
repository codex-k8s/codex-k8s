package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	interactionrequestrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/interactionrequest"
	enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"
)

func (s *Service) MCPUserNotify(ctx context.Context, session SessionContext, input UserNotifyInput) (UserNotifyResult, error) {
	tool, err := s.toolCapability(ToolMCPUserNotify)
	if err != nil {
		return UserNotifyResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, false)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return UserNotifyResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	normalizedInput, err := normalizeUserNotifyInput(input)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return UserNotifyResult{}, err
	}
	recipientProvider, recipientRef, err := resolveInteractionRecipient(runCtx)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return UserNotifyResult{}, err
	}

	item, err := s.interactions.Create(ctx, interactionrequestrepo.CreateParams{
		ProjectID:          strings.TrimSpace(runCtx.Session.ProjectID),
		RunID:              strings.TrimSpace(runCtx.Session.RunID),
		InteractionKind:    enumtypes.InteractionKindNotify,
		State:              enumtypes.InteractionStatePendingDispatch,
		ResolutionKind:     enumtypes.InteractionResolutionKindNone,
		RecipientProvider:  recipientProvider,
		RecipientRef:       recipientRef,
		RequestPayloadJSON: marshalRawJSON(normalizedInput),
		ContextLinksJSON:   marshalRawJSON(buildInteractionContextLinks(runCtx, s.cfg.PublicBaseURL)),
	})
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return UserNotifyResult{}, fmt.Errorf("create interaction request: %w", err)
	}

	s.auditInteractionRequestCreated(ctx, runCtx.Session, item, tool.Name)
	s.auditToolSucceeded(ctx, runCtx.Session, tool)

	return UserNotifyResult{
		Status:        interactionToolStatusAccepted,
		InteractionID: item.ID,
		DeliveryState: interactionDeliveryStateQueued,
		Message:       interactionToolStatusAccepted,
	}, nil
}

func (s *Service) MCPUserDecisionRequest(ctx context.Context, session SessionContext, input UserDecisionRequestInput) (UserDecisionRequestResult, error) {
	tool, err := s.toolCapability(ToolMCPUserDecisionRequest)
	if err != nil {
		return UserDecisionRequestResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, false)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return UserDecisionRequestResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	normalizedInput, err := normalizeUserDecisionRequestInput(input)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return UserDecisionRequestResult{}, err
	}
	existing, found, err := s.interactions.FindOpenDecisionByRunID(ctx, runCtx.Session.RunID)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return UserDecisionRequestResult{}, fmt.Errorf("find open decision interaction: %w", err)
	}
	if found {
		err := errs.FailedPrecondition{Msg: "run already has an open decision interaction"}
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return UserDecisionRequestResult{}, err
	}
	_ = existing

	recipientProvider, recipientRef, err := resolveInteractionRecipient(runCtx)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return UserDecisionRequestResult{}, err
	}

	expiresAt := s.now().UTC().Add(time.Duration(normalizedInput.ResponseTTLSeconds) * time.Second)
	item, err := s.interactions.Create(ctx, interactionrequestrepo.CreateParams{
		ProjectID:          strings.TrimSpace(runCtx.Session.ProjectID),
		RunID:              strings.TrimSpace(runCtx.Session.RunID),
		InteractionKind:    enumtypes.InteractionKindDecisionRequest,
		State:              enumtypes.InteractionStatePendingDispatch,
		ResolutionKind:     enumtypes.InteractionResolutionKindNone,
		RecipientProvider:  recipientProvider,
		RecipientRef:       recipientRef,
		RequestPayloadJSON: marshalRawJSON(normalizedInput),
		ContextLinksJSON:   marshalRawJSON(buildInteractionContextLinks(runCtx, s.cfg.PublicBaseURL)),
		ResponseDeadlineAt: &expiresAt,
	})
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return UserDecisionRequestResult{}, fmt.Errorf("create decision interaction request: %w", err)
	}
	if err := s.setRunWaitContext(
		ctx,
		runCtx.Session,
		waitStateMCP,
		true,
		enumtypes.AgentRunWaitReasonInteractionReply,
		enumtypes.AgentRunWaitTargetKindInteractionRequest,
		item.ID,
		&expiresAt,
	); err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return UserDecisionRequestResult{}, err
	}

	s.auditInteractionRequestCreated(ctx, runCtx.Session, item, tool.Name)
	s.auditInteractionWaitEntered(ctx, runCtx.Session, item.ID, expiresAt.Format(time.RFC3339Nano))
	s.auditToolSucceeded(ctx, runCtx.Session, tool)

	return UserDecisionRequestResult{
		Status:        interactionToolStatusPendingUserResponse,
		InteractionID: item.ID,
		WaitState:     string(waitStateMCP),
		WaitReason:    string(enumtypes.AgentRunWaitReasonInteractionReply),
		ExpiresAt:     expiresAt.Format(time.RFC3339Nano),
	}, nil
}

func (s *Service) SubmitInteractionCallback(ctx context.Context, params SubmitInteractionCallbackParams) (SubmitInteractionCallbackResult, error) {
	if s.interactions == nil {
		return SubmitInteractionCallbackResult{}, fmt.Errorf("interaction repository is not configured")
	}

	normalized := params
	normalized.InteractionID = strings.TrimSpace(normalized.InteractionID)
	normalized.DeliveryID = strings.TrimSpace(normalized.DeliveryID)
	normalized.AdapterEventID = strings.TrimSpace(normalized.AdapterEventID)
	normalized.DeliveryStatus = strings.TrimSpace(normalized.DeliveryStatus)
	normalized.SelectedOptionID = strings.TrimSpace(normalized.SelectedOptionID)
	normalized.FreeText = strings.TrimSpace(normalized.FreeText)
	normalized.ResponderRef = strings.TrimSpace(normalized.ResponderRef)
	if normalized.InteractionID == "" {
		return SubmitInteractionCallbackResult{}, errs.Validation{Field: "interaction_id", Msg: "is required"}
	}
	if normalized.AdapterEventID == "" {
		return SubmitInteractionCallbackResult{}, errs.Validation{Field: "adapter_event_id", Msg: "is required"}
	}
	if normalized.CallbackKind == "" {
		return SubmitInteractionCallbackResult{}, errs.Validation{Field: "callback_kind", Msg: "is required"}
	}
	if len(normalized.NormalizedPayloadJSON) == 0 {
		normalized.NormalizedPayloadJSON = buildInteractionCallbackNormalizedPayload(normalized)
	}
	if len(normalized.RawPayloadJSON) == 0 {
		normalized.RawPayloadJSON = normalized.NormalizedPayloadJSON
	}

	result, err := s.interactions.ApplyCallback(ctx, normalized)
	if err != nil {
		return SubmitInteractionCallbackResult{}, err
	}

	run, found, err := s.runs.GetByID(ctx, result.Interaction.RunID)
	if err != nil {
		return SubmitInteractionCallbackResult{}, fmt.Errorf("load run for interaction callback audit: %w", err)
	}
	if !found {
		return SubmitInteractionCallbackResult{}, fmt.Errorf("run not found for interaction callback")
	}
	session := SessionContext{
		RunID:         run.ID,
		CorrelationID: run.CorrelationID,
		ProjectID:     run.ProjectID,
	}

	s.auditInteractionCallbackReceived(ctx, session, result.Interaction.ID, normalized.AdapterEventID, normalized.CallbackKind)
	if result.Classification == enumtypes.InteractionCallbackResultClassificationAccepted {
		s.auditInteractionResponseAccepted(ctx, session, result.Interaction.ID, result.ResponseRecord)
	} else {
		s.auditInteractionResponseRejected(ctx, session, result.Interaction.ID, result.Classification)
	}

	resumePayload := buildInteractionResumePayload(result.Interaction, result.ResponseRecord)
	if result.ResumeRequired {
		if err := s.setRunWaitContext(ctx, session, waitStateNone, false, "", "", "", nil); err != nil {
			return SubmitInteractionCallbackResult{}, err
		}
		requestStatus := ""
		if resumePayload != nil {
			requestStatus = string(resumePayload.RequestStatus)
		}
		s.auditInteractionWaitResumed(ctx, session, result.Interaction.ID, requestStatus)
	}

	return SubmitInteractionCallbackResult{
		Accepted:            result.Classification == enumtypes.InteractionCallbackResultClassificationAccepted,
		Classification:      result.Classification,
		InteractionState:    string(result.Interaction.State),
		ResumeRequired:      result.ResumeRequired,
		EffectiveResponseID: result.EffectiveResponseID,
		ResumePayload:       resumePayload,
	}, nil
}
