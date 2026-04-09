package githubratelimit

import (
	"encoding/json"
	"fmt"
	"strings"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
	valuetypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/value"
)

// BuildAgentSessionResumePayload builds deterministic resume JSON for agent-runner.
func (s *Service) BuildAgentSessionResumePayload(params BuildResumePayloadParams) (ResumePayloadBuildResult, error) {
	if s == nil {
		return ResumePayloadBuildResult{}, fmt.Errorf("github rate-limit service is not configured")
	}
	if strings.TrimSpace(params.Wait.ID) == "" {
		return ResumePayloadBuildResult{}, fmt.Errorf("wait.id is required")
	}
	if params.Wait.ResumeActionKind != enumtypes.GitHubRateLimitResumeActionKindAgentSessionResume {
		return ResumePayloadBuildResult{}, fmt.Errorf("wait %s does not use agent_session_resume", params.Wait.ID)
	}
	if strings.TrimSpace(string(params.ResolutionKind)) == "" {
		return ResumePayloadBuildResult{}, fmt.Errorf("resolution_kind is required")
	}
	if params.AttemptNo <= 0 {
		return ResumePayloadBuildResult{}, fmt.Errorf("attempt_no must be > 0")
	}

	recoveredAt := params.RecoveredAt.UTC()
	if recoveredAt.IsZero() {
		recoveredAt = s.now()
	}

	guidance, err := renderMessageTemplate("resume_guidance_agent_session_resume", messageTemplateData{
		ContourKind:    string(params.Wait.ContourKind),
		LimitKind:      string(params.Wait.LimitKind),
		OperationClass: string(params.Wait.OperationClass),
		AttemptNo:      params.AttemptNo,
	})
	if err != nil {
		return ResumePayloadBuildResult{}, err
	}

	payload := valuetypes.AgentSessionRateLimitResumePayload{
		WaitID:                 params.Wait.ID,
		WaitReason:             enumtypes.AgentRunWaitReasonGitHubRateLimit,
		ContourKind:            params.Wait.ContourKind,
		LimitKind:              params.Wait.LimitKind,
		ResolutionKind:         params.ResolutionKind,
		RecoveredAt:            recoveredAt,
		AttemptNo:              params.AttemptNo,
		AffectedOperationClass: params.Wait.OperationClass,
		Guidance:               guidance,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return ResumePayloadBuildResult{}, fmt.Errorf("marshal github rate-limit resume payload: %w", err)
	}
	if len(raw) > rateLimitResumePayloadMaxBytes {
		return ResumePayloadBuildResult{}, fmt.Errorf("github rate-limit resume payload exceeds %d bytes", rateLimitResumePayloadMaxBytes)
	}

	return ResumePayloadBuildResult{
		Payload: payload,
		Raw:     append(json.RawMessage(nil), raw...),
	}, nil
}
