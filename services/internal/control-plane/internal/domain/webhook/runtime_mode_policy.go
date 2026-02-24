package webhook

import (
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

const (
	runtimeModeSourceTriggerDefault     = "trigger_default"
	runtimeModeSourceServicesYAML       = "services_yaml"
	runtimeModeSourceServicesYAMLGlobal = "services_yaml_default"
	runtimeModeSourcePushMain           = "push_main"
)

// RuntimeModePolicy defines trigger->runtime mapping loaded from services.yaml.
type RuntimeModePolicy struct {
	Configured   bool
	Source       string
	DefaultMode  agentdomain.RuntimeMode
	TriggerModes map[webhookdomain.TriggerKind]agentdomain.RuntimeMode
}

// DefaultRuntimeModePolicy returns trigger-based legacy defaults.
func DefaultRuntimeModePolicy() RuntimeModePolicy {
	return RuntimeModePolicy{
		Configured:   false,
		DefaultMode:  agentdomain.RuntimeModeFullEnv,
		TriggerModes: map[webhookdomain.TriggerKind]agentdomain.RuntimeMode{},
	}
}

func (policy RuntimeModePolicy) withDefaults() RuntimeModePolicy {
	if policy.DefaultMode == "" {
		policy.DefaultMode = agentdomain.RuntimeModeFullEnv
	}
	if policy.TriggerModes == nil {
		policy.TriggerModes = map[webhookdomain.TriggerKind]agentdomain.RuntimeMode{}
	}
	return policy
}

func normalizeRuntimeTriggerKind(value webhookdomain.TriggerKind) webhookdomain.TriggerKind {
	return webhookdomain.NormalizeTriggerKind(strings.TrimSpace(string(value)))
}

func (policy RuntimeModePolicy) resolve(trigger *issueRunTrigger) (agentdomain.RuntimeMode, string) {
	if trigger == nil {
		return agentdomain.RuntimeModeCodeOnly, runtimeModeSourceTriggerDefault
	}
	triggerKind := normalizeRuntimeTriggerKind(trigger.Kind)
	if triggerKind == webhookdomain.TriggerKindAIRepair {
		policy = policy.withDefaults()
		if mode, ok := policy.TriggerModes[triggerKind]; ok && mode != "" {
			return mode, runtimeModeSourceServicesYAML
		}
		return agentdomain.RuntimeModeCodeOnly, runtimeModeSourceTriggerDefault
	}
	if !policy.Configured {
		return agentdomain.RuntimeModeFullEnv, runtimeModeSourceTriggerDefault
	}

	policy = policy.withDefaults()
	if mode, ok := policy.TriggerModes[triggerKind]; ok && mode != "" {
		return mode, runtimeModeSourceServicesYAML
	}
	return policy.DefaultMode, runtimeModeSourceServicesYAMLGlobal
}
