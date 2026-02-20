package worker

import (
	"encoding/json"
	"strings"
	"time"

	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

type namespaceLeaseContext struct {
	AgentKey    string
	IssueNumber int64
	TriggerKind webhookdomain.TriggerKind
	IsRevise    bool
}

func normalizeNamespaceTTLByRole(input map[string]time.Duration) map[string]time.Duration {
	if len(input) == 0 {
		return map[string]time.Duration{}
	}

	result := make(map[string]time.Duration, len(input))
	for rawRole, ttl := range input {
		role := strings.ToLower(strings.TrimSpace(rawRole))
		if role == "" {
			continue
		}
		if ttl <= 0 {
			continue
		}
		result[role] = ttl
	}
	return result
}

func resolveNamespaceLeaseContext(runPayload json.RawMessage) namespaceLeaseContext {
	runAgent := parseRunAgentPayload(runPayload)
	runtimePayload := parseRunRuntimePayload(runPayload)

	agentKey := strings.ToLower(strings.TrimSpace(runAgent.agentKey))
	issueNumber := runAgent.issueNumber
	if issueNumber <= 0 {
		issueNumber = resolveIssueNumber(runtimePayload)
	}

	triggerKind := webhookdomain.NormalizeTriggerKind(runAgent.triggerKind)
	if strings.TrimSpace(string(triggerKind)) == "" && runtimePayload.Trigger != nil {
		triggerKind = webhookdomain.NormalizeTriggerKind(string(runtimePayload.Trigger.Kind))
	}

	return namespaceLeaseContext{
		AgentKey:    agentKey,
		IssueNumber: issueNumber,
		TriggerKind: triggerKind,
		IsRevise:    webhookdomain.IsReviseTriggerKind(triggerKind),
	}
}

func (s *Service) resolveNamespaceTTL(agentKey string) time.Duration {
	role := strings.ToLower(strings.TrimSpace(agentKey))
	if role != "" {
		if ttl, ok := s.cfg.NamespaceTTLByRole[role]; ok && ttl > 0 {
			return ttl
		}
	}
	return s.cfg.DefaultNamespaceTTL
}
