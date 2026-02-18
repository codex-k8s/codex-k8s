package cli

import (
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
)

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func applyEnvironmentOverrides(values map[string]string, envName string, keys []string, resolver servicescfg.SecretResolver) {
	trimmedEnv := strings.ToLower(strings.TrimSpace(envName))
	if trimmedEnv == "" || len(keys) == 0 || len(values) == 0 {
		return
	}

	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		overrideKey, ok := resolver.ResolveOverrideKey(trimmedEnv, key)
		if !ok {
			continue
		}
		if overrideValue := strings.TrimSpace(values[overrideKey]); overrideValue != "" {
			values[key] = overrideValue
		}
	}
}
