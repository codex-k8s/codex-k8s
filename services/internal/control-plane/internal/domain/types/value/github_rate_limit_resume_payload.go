package value

import (
	"encoding/json"
	"time"

	enumtypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/enum"
)

// AgentSessionRateLimitResumePayload is the deterministic JSON block prepended before runner resume.
type AgentSessionRateLimitResumePayload struct {
	WaitID                 string                                  `json:"wait_id"`
	WaitReason             enumtypes.AgentRunWaitReason            `json:"wait_reason"`
	ContourKind            enumtypes.GitHubRateLimitContourKind    `json:"contour_kind"`
	LimitKind              enumtypes.GitHubRateLimitLimitKind      `json:"limit_kind"`
	ResolutionKind         enumtypes.GitHubRateLimitResolutionKind `json:"resolution_kind"`
	RecoveredAt            time.Time                               `json:"recovered_at"`
	AttemptNo              int                                     `json:"attempt_no"`
	AffectedOperationClass enumtypes.GitHubRateLimitOperationClass `json:"affected_operation_class"`
	Guidance               string                                  `json:"guidance"`
}

// GitHubRateLimitResumePayloadBuildResult contains both typed and serialized resume payload.
type GitHubRateLimitResumePayloadBuildResult struct {
	Payload AgentSessionRateLimitResumePayload `json:"payload"`
	Raw     json.RawMessage                    `json:"raw"`
}
