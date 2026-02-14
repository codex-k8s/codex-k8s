package servicescfg

import "strings"

// ResolveTriggerRuntimeMode resolves runtime mode for trigger kind using webhookRuntime policy.
func ResolveTriggerRuntimeMode(stack *Stack, triggerKind string) RuntimeMode {
	defaultMode := RuntimeModeFullEnv
	if stack == nil {
		return defaultMode
	}

	if stack.Spec.WebhookRuntime.DefaultMode != "" {
		defaultMode = stack.Spec.WebhookRuntime.DefaultMode
	}

	key := normalizeTriggerModeKey(triggerKind)
	if key == "" {
		return defaultMode
	}

	mode, ok := stack.Spec.WebhookRuntime.TriggerModes[key]
	if !ok || mode == "" {
		return defaultMode
	}
	return mode
}

func normalizeTriggerModeKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
