package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codex-k8s/kodex/libs/go/mcp/userinteraction"
	agentrunrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentrun"
	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/query"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

type interactionResumeRunPayload struct {
	Project struct {
		ID string `json:"id"`
	} `json:"project"`
	Agent struct {
		ID string `json:"id"`
	} `json:"agent"`
	LearningMode bool `json:"learning_mode"`
}

func (s *Service) finalizeInteractionResume(
	ctx context.Context,
	interaction entitytypes.InteractionRequest,
	response *entitytypes.InteractionResponseRecord,
	requireCurrentWait bool,
) (bool, error) {
	resumePayload := buildInteractionResumePayload(interaction, response)
	if resumePayload == nil {
		return false, nil
	}

	session, run, err := s.loadInteractionResumeRunContext(ctx, interaction.RunID)
	if err != nil {
		return false, err
	}
	sourceRunTerminal := isTerminalInteractionResumeSourceRun(run.Status)

	scheduled, err := s.scheduleInteractionResume(ctx, run, interaction.ID, resumePayload)
	if err != nil {
		return false, err
	}

	waitCleared, err := s.clearInteractionWaitContext(ctx, session, interaction.ID, requireCurrentWait && !sourceRunTerminal)
	if err != nil {
		return false, err
	}
	if waitCleared {
		s.auditInteractionWaitResumed(ctx, session, interaction.ID, string(resumePayload.RequestStatus))
	}
	if scheduled {
		observeInteractionResume(interaction, resumePayload)
	}

	return scheduled, nil
}

func (s *Service) loadInteractionResumeRunContext(ctx context.Context, runID string) (SessionContext, entitytypes.AgentRun, error) {
	if s.runs == nil {
		return SessionContext{}, entitytypes.AgentRun{}, fmt.Errorf("run repository is not configured")
	}

	run, found, err := s.runs.GetByID(ctx, strings.TrimSpace(runID))
	if err != nil {
		return SessionContext{}, entitytypes.AgentRun{}, fmt.Errorf("load run for interaction lifecycle: %w", err)
	}
	if !found {
		return SessionContext{}, entitytypes.AgentRun{}, fmt.Errorf("run not found for interaction lifecycle")
	}

	return SessionContext{
		RunID:         run.ID,
		CorrelationID: run.CorrelationID,
		ProjectID:     run.ProjectID,
	}, run, nil
}

func (s *Service) scheduleInteractionResume(
	ctx context.Context,
	run entitytypes.AgentRun,
	interactionID string,
	resumePayload *valuetypes.InteractionResumePayload,
) (bool, error) {
	if resumePayload == nil {
		return false, fmt.Errorf("interaction resume payload is required")
	}
	if isTerminalInteractionResumeSourceRun(run.Status) {
		return false, nil
	}

	runMeta, err := parseInteractionResumeRunPayload(run.RunPayload)
	if err != nil {
		return false, err
	}
	pendingRunPayload, err := buildInteractionResumePendingRunPayload(run.RunPayload, resumePayload)
	if err != nil {
		return false, err
	}

	agentID := strings.TrimSpace(runMeta.Agent.ID)
	if agentID == "" {
		return false, fmt.Errorf("run payload missing agent.id for interaction resume")
	}

	projectID := strings.TrimSpace(run.ProjectID)
	if projectID == "" {
		projectID = strings.TrimSpace(runMeta.Project.ID)
	}

	result, err := s.runs.CreatePendingIfAbsent(ctx, agentrunrepo.CreateParams{
		CorrelationID: buildInteractionResumeCorrelationID(interactionID),
		ProjectID:     projectID,
		AgentID:       agentID,
		RunPayload:    pendingRunPayload,
		LearningMode:  runMeta.LearningMode,
	})
	if err != nil {
		return false, fmt.Errorf("create pending interaction resume run: %w", err)
	}

	return strings.TrimSpace(result.RunID) != "", nil
}

func parseInteractionResumeRunPayload(raw json.RawMessage) (interactionResumeRunPayload, error) {
	return querytypes.DecodeRunPayloadInto[interactionResumeRunPayload](raw, "decode run payload for interaction resume")
}

func buildInteractionResumePendingRunPayload(
	raw json.RawMessage,
	resumePayload *valuetypes.InteractionResumePayload,
) (json.RawMessage, error) {
	if resumePayload == nil {
		return nil, fmt.Errorf("interaction resume payload is required")
	}

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("decode run payload for interaction resume persistence: %w", err)
	}

	encodedResumePayload, err := json.Marshal(resumePayload)
	if err != nil {
		return nil, fmt.Errorf("marshal interaction resume payload: %w", err)
	}
	if len(encodedResumePayload) > userinteraction.ResumePayloadMaxBytes {
		return nil, fmt.Errorf("interaction resume payload exceeds %d bytes", userinteraction.ResumePayloadMaxBytes)
	}
	envelope[userinteraction.ResumePayloadRunPayloadFieldName] = encodedResumePayload

	encodedRunPayload, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("marshal run payload with interaction resume payload: %w", err)
	}
	return json.RawMessage(encodedRunPayload), nil
}

func buildInteractionResumeCorrelationID(interactionID string) string {
	return userinteraction.ResumeCorrelationPrefix + strings.TrimSpace(interactionID)
}

func isTerminalInteractionResumeSourceRun(status string) bool {
	switch strings.TrimSpace(status) {
	case "succeeded", "failed", "canceled":
		return true
	default:
		return false
	}
}
