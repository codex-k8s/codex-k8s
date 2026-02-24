package worker

import (
	"strings"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
	querytypes "github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/types/query"
)

func isAIRepairRuntimePayload(payload querytypes.RunRuntimePayload) bool {
	if payload.Trigger == nil {
		return false
	}
	return isAIRepairTriggerKind(string(payload.Trigger.Kind))
}

func isAIRepairTriggerKind(triggerKind string) bool {
	return webhookdomain.NormalizeTriggerKind(strings.TrimSpace(triggerKind)) == webhookdomain.TriggerKindAIRepair
}

func (s *Service) resolveAIRepairNamespace(runNamespace string) string {
	namespace := sanitizeDNSLabelValue(runNamespace)
	if namespace != "" {
		return namespace
	}
	fallback := sanitizeDNSLabelValue(s.cfg.AIRepairNamespace)
	if fallback != "" {
		return fallback
	}
	return strings.TrimSpace(s.cfg.KubernetesNamespace)
}
