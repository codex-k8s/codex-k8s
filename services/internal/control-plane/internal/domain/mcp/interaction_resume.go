package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

const interactionResumeCorrelationPrefix = "interaction-resume:"

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

	scheduled, err := s.scheduleInteractionResume(ctx, run, interaction.ID)
	if err != nil {
		return false, err
	}

	waitCleared, err := s.clearInteractionWaitContext(ctx, session, interaction.ID, requireCurrentWait)
	if err != nil {
		return false, err
	}
	if waitCleared {
		s.auditInteractionWaitResumed(ctx, session, interaction.ID, string(resumePayload.RequestStatus))
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

func (s *Service) scheduleInteractionResume(ctx context.Context, run entitytypes.AgentRun, interactionID string) (bool, error) {
	runMeta, err := parseInteractionResumeRunPayload(run.RunPayload)
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
		RunPayload:    append(json.RawMessage(nil), run.RunPayload...),
		LearningMode:  runMeta.LearningMode,
	})
	if err != nil {
		return false, fmt.Errorf("create pending interaction resume run: %w", err)
	}

	return strings.TrimSpace(result.RunID) != "", nil
}

func parseInteractionResumeRunPayload(raw json.RawMessage) (interactionResumeRunPayload, error) {
	if len(raw) == 0 {
		return interactionResumeRunPayload{}, fmt.Errorf("run payload is empty")
	}

	var payload interactionResumeRunPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return interactionResumeRunPayload{}, fmt.Errorf("decode run payload for interaction resume: %w", err)
	}
	return payload, nil
}

func buildInteractionResumeCorrelationID(interactionID string) string {
	return interactionResumeCorrelationPrefix + strings.TrimSpace(interactionID)
}
