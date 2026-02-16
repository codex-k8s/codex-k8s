package cli

import "strings"

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

func applyEnvironmentOverrides(values map[string]string, envName string, keys []string) {
	trimmedEnv := strings.ToLower(strings.TrimSpace(envName))
	if trimmedEnv == "" || len(keys) == 0 || len(values) == 0 {
		return
	}

	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		overrideKey := environmentOverrideKey(trimmedEnv, key)
		if overrideKey == "" {
			continue
		}
		if overrideValue := strings.TrimSpace(values[overrideKey]); overrideValue != "" {
			values[key] = overrideValue
		}
	}
}

func environmentOverrideKey(envName string, key string) string {
	if !strings.HasPrefix(key, "CODEXK8S_") {
		return ""
	}
	if strings.HasPrefix(key, "CODEXK8S_AI_") || strings.HasPrefix(key, "CODEXK8S_PRODUCTION_") {
		return ""
	}

	suffix := strings.TrimPrefix(key, "CODEXK8S_")
	if suffix == "" {
		return ""
	}

	switch strings.ToLower(strings.TrimSpace(envName)) {
	case githubEnvironmentProduction:
		return "CODEXK8S_PRODUCTION_" + suffix
	case githubEnvironmentAI:
		return "CODEXK8S_AI_" + suffix
	default:
		return ""
	}
}
