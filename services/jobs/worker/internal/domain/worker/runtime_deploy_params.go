package worker

import (
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

// buildPrepareRunEnvironmentParams extracts runtime deploy metadata from run payload and execution context.
func buildPrepareRunEnvironmentParams(claimed runqueuerepo.ClaimedRun, execution valuetypes.RunExecutionContext) PrepareRunEnvironmentParams {
	payload := parseRunRuntimePayload(claimed.RunPayload)

	params := PrepareRunEnvironmentParams{
		RunID:       strings.TrimSpace(claimed.RunID),
		RuntimeMode: strings.TrimSpace(string(execution.RuntimeMode)),
		SlotNo:      claimed.SlotNo,
	}

	if payload.Project != nil {
		params.ServicesYAMLPath = strings.TrimSpace(payload.Project.ServicesYAML)
	}
	if payload.Repository != nil {
		params.RepositoryFullName = strings.TrimSpace(payload.Repository.FullName)
	}
	if payload.Runtime != nil {
		params.TargetEnv = strings.TrimSpace(payload.Runtime.TargetEnv)
		params.BuildRef = strings.TrimSpace(payload.Runtime.BuildRef)
		params.DeployOnly = payload.Runtime.DeployOnly
		runtimeNamespace := sanitizeDNSLabelValue(payload.Runtime.Namespace)
		if runtimeNamespace != "" {
			params.Namespace = runtimeNamespace
		}
	}
	if params.TargetEnv == "" && execution.RuntimeMode == agentdomain.RuntimeModeFullEnv {
		params.TargetEnv = "ai"
	}

	return params
}
