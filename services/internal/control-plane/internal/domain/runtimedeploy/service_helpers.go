package runtimedeploy

import (
	"os"
	"strings"
)

func sanitizeNameToken(value string, max int) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return ""
	}
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, ".", "-")
	normalized = imageTagSanitizer.ReplaceAllString(normalized, "-")
	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}
	normalized = strings.Trim(normalized, "-")
	if max > 0 && len(normalized) > max {
		normalized = strings.TrimRight(normalized[:max], "-")
	}
	return normalized
}

func sanitizeImageTag(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return ""
	}
	normalized = imageTagSanitizer.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, ".-")
	if normalized == "" {
		return ""
	}
	if len(normalized) > 120 {
		normalized = normalized[:120]
	}
	return normalized
}

func valueOr(values map[string]string, key string, fallback string) string {
	if values != nil {
		if value, ok := values[key]; ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func cloneStringMap(input map[string]string) map[string]string {
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
