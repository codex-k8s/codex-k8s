package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	agentsessionrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentsession"
	mcpactionrequestrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/mcpactionrequest"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

func (s *Service) MCPSecretSyncEnv(ctx context.Context, session SessionContext, input SecretSyncEnvInput) (SecretSyncEnvResult, error) {
	tool, err := s.toolCapability(ToolMCPSecretSyncEnv)
	if err != nil {
		return SecretSyncEnvResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, true)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return SecretSyncEnvResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	environment := normalizeEnvName(input.Environment)
	if environment == "" {
		err := fmt.Errorf("environment is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SecretSyncEnvResult{}, err
	}
	githubSecretName := strings.TrimSpace(input.GitHubSecretName)
	if githubSecretName == "" {
		err := fmt.Errorf("github_secret_name is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SecretSyncEnvResult{}, err
	}
	kubernetesNamespace := normalizeSecretTargetNamespace(runCtx.Session, input.KubernetesNamespace)
	if kubernetesNamespace == "" {
		err := fmt.Errorf("kubernetes_namespace is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SecretSyncEnvResult{}, err
	}
	kubernetesSecretName := strings.TrimSpace(input.KubernetesSecretName)
	if kubernetesSecretName == "" {
		err := fmt.Errorf("kubernetes_secret_name is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SecretSyncEnvResult{}, err
	}
	kubernetesSecretKey := normalizeKubernetesSecretDataKey(input.KubernetesSecretKey)

	secretValue := strings.TrimSpace(input.SecretValue)
	if secretValue == "" {
		secretValue, err = newGeneratedSecretValue()
		if err != nil {
			s.auditToolFailed(ctx, runCtx.Session, tool, err)
			return SecretSyncEnvResult{}, err
		}
	}
	encryptedSecret, err := s.tokenCrypt.EncryptString(secretValue)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SecretSyncEnvResult{}, fmt.Errorf("encrypt secret value: %w", err)
	}

	targetRef := marshalRawJSON(approvalTargetRef{
		Environment:          environment,
		GitHubSecretName:     githubSecretName,
		KubernetesNamespace:  kubernetesNamespace,
		KubernetesSecretName: kubernetesSecretName,
		KubernetesSecretKey:  kubernetesSecretKey,
	})
	payload := marshalRawJSON(secretSyncPayload{
		Environment:          environment,
		GitHubSecretName:     githubSecretName,
		KubernetesNamespace:  kubernetesNamespace,
		KubernetesSecretName: kubernetesSecretName,
		KubernetesSecretKey:  kubernetesSecretKey,
		SecretValueEncrypted: encryptedValueBase64(encryptedSecret),
	})

	if input.DryRun {
		s.auditToolSucceeded(ctx, runCtx.Session, tool)
		return SecretSyncEnvResult{
			Status:        ToolExecutionStatusOK,
			ApprovalState: string(entitytypes.MCPApprovalModeNone),
			Environment:   environment,
			GitHubSecret:  githubSecretName,
			KubernetesRef: kubernetesNamespace + "/" + kubernetesSecretName + "#" + kubernetesSecretKey,
			DryRun:        true,
			Message:       controlToolMessageDryRun,
		}, nil
	}

	approvalMode := normalizeApprovalMode(resolveControlApprovalMode(tool.Name, runCtx))
	if approvalMode == entitytypes.MCPApprovalModeNone {
		if err := s.applySecretSync(ctx, runCtx, payload); err != nil {
			s.auditToolFailed(ctx, runCtx.Session, tool, err)
			return SecretSyncEnvResult{}, err
		}
		item, err := s.createAppliedActionRequest(
			ctx,
			runCtx,
			tool,
			string(controlActionSecretSyncEnv),
			targetRef,
			payload,
			"store applied control action",
		)
		if err != nil {
			s.auditToolFailed(ctx, runCtx.Session, tool, err)
			return SecretSyncEnvResult{}, err
		}
		s.auditApprovalApplied(ctx, runCtx.Session, item, string(floweventdomain.ActorIDControlPlaneMCP))
		s.auditToolSucceeded(ctx, runCtx.Session, tool)
		return SecretSyncEnvResult{
			Status:        ToolExecutionStatusOK,
			RequestID:     item.ID,
			ApprovalState: string(item.ApprovalState),
			Environment:   environment,
			GitHubSecret:  githubSecretName,
			KubernetesRef: kubernetesNamespace + "/" + kubernetesSecretName + "#" + kubernetesSecretKey,
			Message:       controlToolMessageApplied,
		}, nil
	}

	request, created, err := s.ensurePendingApprovalRequest(ctx, runCtx, tool, string(controlActionSecretSyncEnv), targetRef, approvalMode, payload)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return SecretSyncEnvResult{}, err
	}
	s.auditToolApprovalPending(ctx, runCtx.Session, tool, controlToolMessageApprovalRequired)
	if created {
		s.auditApprovalRequested(ctx, runCtx.Session, request, tool)
	}
	return SecretSyncEnvResult{
		Status:        ToolExecutionStatusApprovalRequired,
		RequestID:     request.ID,
		ApprovalState: string(request.ApprovalState),
		Environment:   environment,
		GitHubSecret:  githubSecretName,
		KubernetesRef: kubernetesNamespace + "/" + kubernetesSecretName + "#" + kubernetesSecretKey,
		Message:       controlToolMessageApprovalRequired,
	}, nil
}

func (s *Service) MCPDatabaseLifecycle(ctx context.Context, session SessionContext, input DatabaseLifecycleInput) (DatabaseLifecycleResult, error) {
	tool, err := s.toolCapability(ToolMCPDatabaseLifecycle)
	if err != nil {
		return DatabaseLifecycleResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, false)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return DatabaseLifecycleResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	environment := normalizeEnvName(input.Environment)
	if environment == "" {
		err := fmt.Errorf("environment is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return DatabaseLifecycleResult{}, err
	}
	action := DatabaseLifecycleAction(strings.ToLower(strings.TrimSpace(string(input.Action))))
	switch action {
	case DatabaseLifecycleActionCreate, DatabaseLifecycleActionDelete:
	default:
		err := fmt.Errorf("action is invalid")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return DatabaseLifecycleResult{}, err
	}
	databaseName := strings.TrimSpace(input.DatabaseName)
	if databaseName == "" {
		err := fmt.Errorf("database_name is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return DatabaseLifecycleResult{}, err
	}

	targetRef := marshalRawJSON(approvalTargetRef{
		Environment:  environment,
		DatabaseName: databaseName,
	})
	payload := marshalRawJSON(databaseLifecyclePayload{
		Environment:  environment,
		Action:       action,
		DatabaseName: databaseName,
	})

	if input.DryRun {
		s.auditToolSucceeded(ctx, runCtx.Session, tool)
		return DatabaseLifecycleResult{
			Status:        ToolExecutionStatusOK,
			ApprovalState: string(entitytypes.MCPApprovalModeNone),
			Environment:   environment,
			Action:        string(action),
			DatabaseName:  databaseName,
			DryRun:        true,
			Message:       controlToolMessageDryRun,
		}, nil
	}

	approvalMode := normalizeApprovalMode(resolveControlApprovalMode(tool.Name, runCtx))
	if approvalMode == entitytypes.MCPApprovalModeNone {
		if _, err := s.applyDatabaseLifecycle(ctx, payload); err != nil {
			s.auditToolFailed(ctx, runCtx.Session, tool, err)
			return DatabaseLifecycleResult{}, err
		}
		item, err := s.createAppliedActionRequest(
			ctx,
			runCtx,
			tool,
			databaseActionName(action),
			targetRef,
			payload,
			"store applied control action",
		)
		if err != nil {
			s.auditToolFailed(ctx, runCtx.Session, tool, err)
			return DatabaseLifecycleResult{}, err
		}
		s.auditApprovalApplied(ctx, runCtx.Session, item, string(floweventdomain.ActorIDControlPlaneMCP))
		s.auditToolSucceeded(ctx, runCtx.Session, tool)
		return DatabaseLifecycleResult{
			Status:        ToolExecutionStatusOK,
			RequestID:     item.ID,
			ApprovalState: string(item.ApprovalState),
			Environment:   environment,
			Action:        string(action),
			DatabaseName:  databaseName,
			Applied:       true,
			Message:       controlToolMessageApplied,
		}, nil
	}

	request, created, err := s.ensurePendingApprovalRequest(ctx, runCtx, tool, databaseActionName(action), targetRef, approvalMode, payload)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return DatabaseLifecycleResult{}, err
	}
	s.auditToolApprovalPending(ctx, runCtx.Session, tool, controlToolMessageApprovalRequired)
	if created {
		s.auditApprovalRequested(ctx, runCtx.Session, request, tool)
	}
	return DatabaseLifecycleResult{
		Status:        ToolExecutionStatusApprovalRequired,
		RequestID:     request.ID,
		ApprovalState: string(request.ApprovalState),
		Environment:   environment,
		Action:        string(action),
		DatabaseName:  databaseName,
		Message:       controlToolMessageApprovalRequired,
	}, nil
}

func (s *Service) MCPOwnerFeedbackRequest(ctx context.Context, session SessionContext, input OwnerFeedbackRequestInput) (OwnerFeedbackRequestResult, error) {
	tool, err := s.toolCapability(ToolMCPOwnerFeedbackRequest)
	if err != nil {
		return OwnerFeedbackRequestResult{}, err
	}

	runCtx, err := s.resolveRunContext(ctx, session, false)
	if err != nil {
		s.auditToolFailed(ctx, session, tool, err)
		return OwnerFeedbackRequestResult{}, err
	}
	s.auditToolCalled(ctx, runCtx.Session, tool)

	question := strings.TrimSpace(input.Question)
	if question == "" {
		err := fmt.Errorf("question is required")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return OwnerFeedbackRequestResult{}, err
	}
	options := normalizeOptions(input.Options)
	if len(options) < 2 || len(options) > 5 {
		err := fmt.Errorf("options count must be between 2 and 5")
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return OwnerFeedbackRequestResult{}, err
	}

	targetRef := marshalRawJSON(approvalTargetRef{Environment: "owner-feedback"})
	payload := marshalRawJSON(ownerFeedbackPayload{
		Question:    question,
		Options:     options,
		AllowCustom: input.AllowCustom,
	})

	if input.DryRun {
		s.auditToolSucceeded(ctx, runCtx.Session, tool)
		return OwnerFeedbackRequestResult{
			Status:        ToolExecutionStatusOK,
			ApprovalState: string(entitytypes.MCPApprovalModeNone),
			Question:      question,
			Options:       options,
			DryRun:        true,
			Message:       controlToolMessageDryRun,
		}, nil
	}

	approvalMode := normalizeApprovalMode(resolveControlApprovalMode(tool.Name, runCtx))
	if approvalMode == entitytypes.MCPApprovalModeNone {
		item, err := s.createAppliedActionRequest(
			ctx,
			runCtx,
			tool,
			string(controlActionOwnerFeedback),
			targetRef,
			payload,
			"store owner feedback action",
		)
		if err != nil {
			s.auditToolFailed(ctx, runCtx.Session, tool, err)
			return OwnerFeedbackRequestResult{}, err
		}
		s.auditApprovalApplied(ctx, runCtx.Session, item, string(floweventdomain.ActorIDControlPlaneMCP))
		s.auditToolSucceeded(ctx, runCtx.Session, tool)
		return OwnerFeedbackRequestResult{
			Status:        ToolExecutionStatusOK,
			RequestID:     item.ID,
			ApprovalState: string(item.ApprovalState),
			Question:      question,
			Options:       options,
			Message:       controlToolMessageApplied,
		}, nil
	}

	request, created, err := s.ensurePendingApprovalRequest(ctx, runCtx, tool, string(controlActionOwnerFeedback), targetRef, approvalMode, payload)
	if err != nil {
		s.auditToolFailed(ctx, runCtx.Session, tool, err)
		return OwnerFeedbackRequestResult{}, err
	}
	s.auditToolApprovalPending(ctx, runCtx.Session, tool, controlToolMessageApprovalRequired)
	if created {
		s.auditApprovalRequested(ctx, runCtx.Session, request, tool)
	}
	return OwnerFeedbackRequestResult{
		Status:        ToolExecutionStatusApprovalRequired,
		RequestID:     request.ID,
		ApprovalState: string(request.ApprovalState),
		Question:      question,
		Options:       options,
		Message:       controlToolMessageApprovalRequired,
	}, nil
}

func (s *Service) ListPendingApprovals(ctx context.Context, limit int) ([]ApprovalListItem, error) {
	items, err := s.actions.ListPending(ctx, clampLimit(limit, 100, 500))
	if err != nil {
		return nil, fmt.Errorf("list pending approvals: %w", err)
	}

	out := make([]ApprovalListItem, 0, len(items))
	for _, item := range items {
		out = append(out, ApprovalListItem{
			ID:            item.ID,
			CorrelationID: item.CorrelationID,
			RunID:         item.RunID,
			ProjectID:     item.ProjectID,
			ProjectSlug:   item.ProjectSlug,
			ProjectName:   item.ProjectName,
			IssueNumber:   item.IssueNumber,
			PRNumber:      item.PRNumber,
			TriggerLabel:  item.TriggerLabel,
			ToolName:      item.ToolName,
			Action:        item.Action,
			ApprovalMode:  string(item.ApprovalMode),
			RequestedBy:   item.RequestedBy,
			CreatedAt:     item.CreatedAt.UTC(),
		})
	}
	return out, nil
}

func (s *Service) ResolveApproval(ctx context.Context, params ResolveApprovalParams) (ResolveApprovalResult, error) {
	if params.RequestID <= 0 {
		return ResolveApprovalResult{}, fmt.Errorf("request_id is required")
	}
	decision, err := normalizeApprovalDecision(params.Decision)
	if err != nil {
		return ResolveApprovalResult{}, err
	}
	actorID := strings.TrimSpace(params.ActorID)
	if actorID == "" {
		return ResolveApprovalResult{}, fmt.Errorf("actor_id is required")
	}
	reason := strings.TrimSpace(params.Reason)

	item, ok, err := s.actions.GetByID(ctx, params.RequestID)
	if err != nil {
		return ResolveApprovalResult{}, fmt.Errorf("get approval request: %w", err)
	}
	if !ok {
		return ResolveApprovalResult{}, fmt.Errorf("approval request not found")
	}
	if item.ApprovalState != entitytypes.MCPApprovalStateRequested {
		return ResolveApprovalResult{
			ID:            item.ID,
			CorrelationID: item.CorrelationID,
			RunID:         item.RunID,
			ToolName:      item.ToolName,
			Action:        item.Action,
			ApprovalState: string(item.ApprovalState),
		}, nil
	}

	state := decisionToApprovalState(decision)
	decisionPayload := marshalRawJSON(approvalDecisionPayload{
		Decision:  string(decision),
		ActorID:   actorID,
		Reason:    reason,
		DecidedAt: nowRFC3339Nano(s.now()),
	})
	updated, ok, err := s.actions.UpdateState(ctx, mcpactionrequestrepo.UpdateStateParams{
		ID:            item.ID,
		ApprovalState: state,
		AppliedBy:     actorID,
		Payload:       decisionPayload,
	})
	if err != nil {
		return ResolveApprovalResult{}, fmt.Errorf("update approval state: %w", err)
	}
	if !ok {
		return ResolveApprovalResult{}, fmt.Errorf("approval request not found")
	}

	approvalSession := SessionContext{RunID: updated.RunID, CorrelationID: updated.CorrelationID}
	switch state {
	case entitytypes.MCPApprovalStateApproved:
		s.auditApprovalApproved(ctx, approvalSession, updated, actorID, reason)
		applied, applyErr := s.applyApprovedControlAction(ctx, updated, actorID)
		if applyErr != nil {
			failedPayload := marshalRawJSON(approvalDecisionPayload{
				Decision:  string(ApprovalDecisionFailed),
				ActorID:   string(floweventdomain.ActorIDControlPlaneMCP),
				Error:     applyErr.Error(),
				DecidedAt: nowRFC3339Nano(s.now()),
			})
			failed, ok, err := s.actions.UpdateState(ctx, mcpactionrequestrepo.UpdateStateParams{
				ID:            updated.ID,
				ApprovalState: entitytypes.MCPApprovalStateFailed,
				AppliedBy:     string(floweventdomain.ActorIDControlPlaneMCP),
				Payload:       failedPayload,
			})
			if err != nil {
				return ResolveApprovalResult{}, fmt.Errorf("mark approval failed: %w", err)
			}
			if !ok {
				return ResolveApprovalResult{}, fmt.Errorf("mark approval failed: request not found")
			}
			s.auditApprovalFailed(ctx, approvalSession, failed, string(floweventdomain.ActorIDControlPlaneMCP), applyErr.Error())
			if clearErr := s.setRunWaitState(ctx, approvalSession, waitStateNone, false); clearErr != nil {
				return ResolveApprovalResult{}, clearErr
			}
			return ResolveApprovalResult{
				ID:            failed.ID,
				CorrelationID: failed.CorrelationID,
				RunID:         failed.RunID,
				ToolName:      failed.ToolName,
				Action:        failed.Action,
				ApprovalState: string(failed.ApprovalState),
			}, nil
		}
		updated = applied
		if clearErr := s.setRunWaitState(ctx, approvalSession, waitStateNone, false); clearErr != nil {
			return ResolveApprovalResult{}, clearErr
		}
	case entitytypes.MCPApprovalStateDenied:
		s.auditApprovalDenied(ctx, approvalSession, updated, actorID, reason)
		if clearErr := s.setRunWaitState(ctx, approvalSession, waitStateNone, false); clearErr != nil {
			return ResolveApprovalResult{}, clearErr
		}
	case entitytypes.MCPApprovalStateExpired:
		s.auditApprovalExpired(ctx, approvalSession, updated, actorID, reason)
		if clearErr := s.setRunWaitState(ctx, approvalSession, waitStateNone, false); clearErr != nil {
			return ResolveApprovalResult{}, clearErr
		}
	case entitytypes.MCPApprovalStateFailed:
		s.auditApprovalFailed(ctx, approvalSession, updated, actorID, reason)
		if clearErr := s.setRunWaitState(ctx, approvalSession, waitStateNone, false); clearErr != nil {
			return ResolveApprovalResult{}, clearErr
		}
	}

	return ResolveApprovalResult{
		ID:            updated.ID,
		CorrelationID: updated.CorrelationID,
		RunID:         updated.RunID,
		ToolName:      updated.ToolName,
		Action:        updated.Action,
		ApprovalState: string(updated.ApprovalState),
	}, nil
}

func (s *Service) ensurePendingApprovalRequest(
	ctx context.Context,
	runCtx resolvedRunContext,
	tool ToolCapability,
	action string,
	targetRef json.RawMessage,
	approvalMode entitytypes.MCPApprovalMode,
	payload json.RawMessage,
) (entitytypes.MCPActionRequest, bool, error) {
	existing, found, err := s.actions.FindPendingBySignature(
		ctx,
		runCtx.Session.RunID,
		string(tool.Name),
		action,
		targetRef,
	)
	if err != nil {
		return entitytypes.MCPActionRequest{}, false, fmt.Errorf("find pending approval request: %w", err)
	}
	if found {
		return existing, false, nil
	}

	item, err := s.actions.Create(ctx, mcpactionrequestrepo.CreateParams{
		CorrelationID: runCtx.Session.CorrelationID,
		RunID:         runCtx.Session.RunID,
		ToolName:      string(tool.Name),
		Action:        action,
		TargetRef:     targetRef,
		ApprovalMode:  approvalMode,
		ApprovalState: entitytypes.MCPApprovalStateRequested,
		RequestedBy:   requestActorID(runCtx),
		Payload:       payload,
	})
	if err != nil {
		return entitytypes.MCPActionRequest{}, false, fmt.Errorf("create approval request: %w", err)
	}
	if err := s.setRunWaitState(ctx, runCtx.Session, waitStateMCP, true); err != nil {
		return entitytypes.MCPActionRequest{}, false, err
	}

	return item, true, nil
}

func (s *Service) createAppliedActionRequest(
	ctx context.Context,
	runCtx resolvedRunContext,
	tool ToolCapability,
	action string,
	targetRef json.RawMessage,
	payload json.RawMessage,
	errPrefix string,
) (entitytypes.MCPActionRequest, error) {
	item, err := s.actions.Create(ctx, mcpactionrequestrepo.CreateParams{
		CorrelationID: runCtx.Session.CorrelationID,
		RunID:         runCtx.Session.RunID,
		ToolName:      string(tool.Name),
		Action:        action,
		TargetRef:     targetRef,
		ApprovalMode:  entitytypes.MCPApprovalModeNone,
		ApprovalState: entitytypes.MCPApprovalStateApplied,
		RequestedBy:   requestActorID(runCtx),
		AppliedBy:     string(floweventdomain.ActorIDControlPlaneMCP),
		Payload:       payload,
	})
	if err != nil {
		return entitytypes.MCPActionRequest{}, fmt.Errorf("%s: %w", errPrefix, err)
	}
	return item, nil
}

func (s *Service) applyApprovedControlAction(
	ctx context.Context,
	item entitytypes.MCPActionRequest,
	actorID string,
) (entitytypes.MCPActionRequest, error) {
	session := SessionContext{
		RunID:         item.RunID,
		CorrelationID: item.CorrelationID,
		ProjectID:     item.ProjectID,
	}
	runCtx, err := s.resolveRunContext(ctx, session, item.ToolName == string(ToolMCPSecretSyncEnv))
	if err != nil {
		return entitytypes.MCPActionRequest{}, err
	}

	switch ToolName(item.ToolName) {
	case ToolMCPSecretSyncEnv:
		if err := s.applySecretSync(ctx, runCtx, item.Payload); err != nil {
			return entitytypes.MCPActionRequest{}, err
		}
	case ToolMCPDatabaseLifecycle:
		if _, err := s.applyDatabaseLifecycle(ctx, item.Payload); err != nil {
			return entitytypes.MCPActionRequest{}, err
		}
	case ToolMCPOwnerFeedbackRequest:
		// This tool stores operator decision in payload and does not have external side effects.
	default:
		return entitytypes.MCPActionRequest{}, fmt.Errorf("unsupported tool %q", item.ToolName)
	}

	appliedPayload := marshalRawJSON(approvalAppliedPayload{
		AppliedAt: nowRFC3339Nano(s.now()),
		AppliedBy: actorID,
	})
	updated, ok, err := s.actions.UpdateState(ctx, mcpactionrequestrepo.UpdateStateParams{
		ID:            item.ID,
		ApprovalState: entitytypes.MCPApprovalStateApplied,
		AppliedBy:     actorID,
		Payload:       appliedPayload,
	})
	if err != nil {
		return entitytypes.MCPActionRequest{}, fmt.Errorf("mark approval applied: %w", err)
	}
	if !ok {
		return entitytypes.MCPActionRequest{}, fmt.Errorf("mark approval applied: request not found")
	}
	s.auditApprovalApplied(ctx, runCtx.Session, updated, actorID)
	return updated, nil
}

func (s *Service) applySecretSync(ctx context.Context, runCtx resolvedRunContext, payloadRaw json.RawMessage) error {
	payload, err := decodeSecretSyncPayload(payloadRaw)
	if err != nil {
		return err
	}
	encryptedValue, err := decodeEncryptedValueBase64(payload.SecretValueEncrypted)
	if err != nil {
		return err
	}
	secretValue, err := s.tokenCrypt.DecryptString(encryptedValue)
	if err != nil {
		return fmt.Errorf("decrypt secret value: %w", err)
	}
	if strings.TrimSpace(secretValue) == "" {
		return fmt.Errorf("secret value is empty")
	}

	if err := s.github.UpsertRepositorySecret(ctx, GitHubUpsertRepositorySecretParams{
		Token:      runCtx.Token,
		Owner:      runCtx.Repository.Owner,
		Repository: runCtx.Repository.Name,
		SecretName: payload.GitHubSecretName,
		Value:      secretValue,
	}); err != nil {
		return fmt.Errorf("sync github secret: %w", err)
	}

	if err := s.kubernetes.UpsertSecret(ctx, payload.KubernetesNamespace, payload.KubernetesSecretName, map[string][]byte{
		payload.KubernetesSecretKey: []byte(secretValue),
	}); err != nil {
		return fmt.Errorf("sync kubernetes secret: %w", err)
	}
	return nil
}

func (s *Service) applyDatabaseLifecycle(ctx context.Context, payloadRaw json.RawMessage) (bool, error) {
	payload, err := decodeDatabaseLifecyclePayload(payloadRaw)
	if err != nil {
		return false, err
	}

	switch payload.Action {
	case DatabaseLifecycleActionCreate:
		created, err := s.database.EnsureDatabase(ctx, payload.DatabaseName)
		if err != nil {
			return false, fmt.Errorf("create database: %w", err)
		}
		return created, nil
	case DatabaseLifecycleActionDelete:
		deleted, err := s.database.DropDatabase(ctx, payload.DatabaseName)
		if err != nil {
			return false, fmt.Errorf("delete database: %w", err)
		}
		return deleted, nil
	default:
		return false, fmt.Errorf("unsupported database action %q", payload.Action)
	}
}

func decodeSecretSyncPayload(raw json.RawMessage) (secretSyncPayload, error) {
	var payload secretSyncPayload
	if len(raw) == 0 {
		return payload, fmt.Errorf("secret sync payload is empty")
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return payload, fmt.Errorf("decode secret sync payload: %w", err)
	}
	payload.Environment = normalizeEnvName(payload.Environment)
	payload.GitHubSecretName = strings.TrimSpace(payload.GitHubSecretName)
	payload.KubernetesNamespace = strings.TrimSpace(payload.KubernetesNamespace)
	payload.KubernetesSecretName = strings.TrimSpace(payload.KubernetesSecretName)
	payload.KubernetesSecretKey = normalizeKubernetesSecretDataKey(payload.KubernetesSecretKey)

	if payload.Environment == "" {
		return payload, fmt.Errorf("secret sync payload environment is required")
	}
	if payload.GitHubSecretName == "" {
		return payload, fmt.Errorf("secret sync payload github_secret_name is required")
	}
	if payload.KubernetesNamespace == "" {
		return payload, fmt.Errorf("secret sync payload kubernetes_namespace is required")
	}
	if payload.KubernetesSecretName == "" {
		return payload, fmt.Errorf("secret sync payload kubernetes_secret_name is required")
	}
	if strings.TrimSpace(payload.SecretValueEncrypted) == "" {
		return payload, fmt.Errorf("secret sync payload secret value is missing")
	}
	return payload, nil
}

func decodeDatabaseLifecyclePayload(raw json.RawMessage) (databaseLifecyclePayload, error) {
	var payload databaseLifecyclePayload
	if len(raw) == 0 {
		return payload, fmt.Errorf("database lifecycle payload is empty")
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return payload, fmt.Errorf("decode database lifecycle payload: %w", err)
	}
	payload.Environment = normalizeEnvName(payload.Environment)
	payload.Action = DatabaseLifecycleAction(strings.ToLower(strings.TrimSpace(string(payload.Action))))
	payload.DatabaseName = strings.TrimSpace(payload.DatabaseName)

	if payload.Environment == "" {
		return payload, fmt.Errorf("database lifecycle payload environment is required")
	}
	switch payload.Action {
	case DatabaseLifecycleActionCreate, DatabaseLifecycleActionDelete:
	default:
		return payload, fmt.Errorf("database lifecycle payload action is invalid")
	}
	if payload.DatabaseName == "" {
		return payload, fmt.Errorf("database lifecycle payload database_name is required")
	}
	return payload, nil
}

func decisionToApprovalState(decision ApprovalDecision) entitytypes.MCPApprovalState {
	switch decision {
	case ApprovalDecisionApproved:
		return entitytypes.MCPApprovalStateApproved
	case ApprovalDecisionDenied:
		return entitytypes.MCPApprovalStateDenied
	case ApprovalDecisionExpired:
		return entitytypes.MCPApprovalStateExpired
	case ApprovalDecisionFailed:
		return entitytypes.MCPApprovalStateFailed
	default:
		return entitytypes.MCPApprovalStateFailed
	}
}

func normalizeApprovalDecision(value ApprovalDecision) (ApprovalDecision, error) {
	decision := ApprovalDecision(strings.ToLower(strings.TrimSpace(string(value))))
	switch decision {
	case ApprovalDecisionApproved, ApprovalDecisionDenied, ApprovalDecisionExpired, ApprovalDecisionFailed:
		return decision, nil
	default:
		return "", fmt.Errorf("decision is invalid")
	}
}

func databaseActionName(action DatabaseLifecycleAction) string {
	switch action {
	case DatabaseLifecycleActionCreate:
		return string(controlActionDatabaseCreate)
	case DatabaseLifecycleActionDelete:
		return string(controlActionDatabaseDelete)
	default:
		return string(action)
	}
}

func (s *Service) setRunWaitState(ctx context.Context, session SessionContext, state waitState, timeoutGuardDisabled bool) error {
	if strings.TrimSpace(session.RunID) == "" || s.sessions == nil {
		return nil
	}
	var lastHeartbeatAt *time.Time
	now := s.now().UTC()
	if state == waitStateMCP {
		lastHeartbeatAt = &now
	}

	_, err := s.sessions.SetWaitStateByRunID(ctx, agentsessionrepo.SetWaitStateParams{
		RunID:                session.RunID,
		WaitState:            string(state),
		TimeoutGuardDisabled: timeoutGuardDisabled,
		LastHeartbeatAt:      lastHeartbeatAt,
	})
	if err != nil {
		return fmt.Errorf("set run wait state: %w", err)
	}

	switch state {
	case waitStateMCP:
		s.auditRunWaitPaused(ctx, session, runWaitPayload{
			RunID:                session.RunID,
			WaitState:            string(state),
			TimeoutGuardDisabled: timeoutGuardDisabled,
		})
	default:
		s.auditRunWaitResumed(ctx, session, runWaitPayload{
			RunID:                session.RunID,
			TimeoutGuardDisabled: timeoutGuardDisabled,
		})
	}
	return nil
}
