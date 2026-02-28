package worker

import valuetypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/value"

func applyPreparedNamespace(execution valuetypes.RunExecutionContext, namespace string) valuetypes.RunExecutionContext {
	resolvedNamespace := sanitizeDNSLabelValue(namespace)
	if resolvedNamespace == "" {
		return execution
	}
	execution.Namespace = resolvedNamespace
	return execution
}
