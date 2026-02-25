package worker

import (
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	runqueuerepo "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/repository/runqueue"
	valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"
)

// buildPrepareRunEnvironmentParams extracts runtime deploy metadata from run payload and execution context.
func buildPrepareRunEnvironmentParams(claimed runqueuerepo.ClaimedRun, execution valuetypes.RunExecutionContext) PrepareRunEnvironmentParams {
	return buildPrepareRunEnvironmentParamsBase(claimed.RunID, claimed.SlotNo, claimed.RunPayload, execution)
}

func buildPrepareRunEnvironmentParamsFromRunning(run runqueuerepo.RunningRun, execution valuetypes.RunExecutionContext) PrepareRunEnvironmentParams {
	return buildPrepareRunEnvironmentParamsBase(run.RunID, run.SlotNo, run.RunPayload, execution)
}

func buildPrepareRunEnvironmentParamsBase(runID string, slotNo int, runPayload []byte, execution valuetypes.RunExecutionContext) PrepareRunEnvironmentParams {
	payload := parseRunRuntimePayload(runPayload)

	params := PrepareRunEnvironmentParams{
		RunID:       strings.TrimSpace(runID),
		RuntimeMode: strings.TrimSpace(string(execution.RuntimeMode)),
		SlotNo:      slotNo,
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
