package runtimedeploy

import (
	"os"
	"regexp"
	"strings"
)

var (
	runtimeBuildRefAllowedPattern = regexp.MustCompile(`^[A-Za-z0-9._/@+-]+$`)
	gitCommitRefPattern           = regexp.MustCompile(`(?i)^[0-9a-f]{7,64}$`)
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

func resolveRuntimeBuildRef(candidates ...string) string {
	for _, candidate := range candidates {
		if normalized := normalizeRuntimeBuildRef(candidate); normalized != "" {
			return normalized
		}
	}
	return "main"
}

func normalizeRuntimeBuildCheckoutRef(buildRef string) string {
	resolved := resolveRuntimeBuildRef(buildRef)
	if strings.HasPrefix(resolved, "refs/") {
		return resolved
	}
	if gitCommitRefPattern.MatchString(resolved) {
		return resolved
	}
	return "origin/" + strings.TrimPrefix(resolved, "origin/")
}

func normalizeRuntimeBuildRef(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	trimmed = trimMatchingQuotes(trimmed)
	if trimmed == "" {
		return ""
	}
	if strings.ContainsAny(trimmed, ";|&`$(){}<>") {
		return ""
	}

	// Happy path: already a plain ref.
	if !strings.ContainsAny(trimmed, " \t\r\n") {
		return normalizeRuntimeBuildRefToken(trimmed)
	}

	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return ""
	}
	for idx, field := range fields {
		switch strings.TrimSpace(field) {
		case "-b", "--branch":
			if idx+1 < len(fields) {
				return normalizeRuntimeBuildRefToken(fields[idx+1])
			}
		}
	}
	for _, field := range fields {
		token := strings.TrimSpace(field)
		if token == "" {
			continue
		}
		switch token {
		case "git", "checkout", "switch", "--detach":
			continue
		}
		if strings.HasPrefix(token, "-") {
			continue
		}
		if normalized := normalizeRuntimeBuildRefToken(token); normalized != "" {
			return normalized
		}
	}
	return ""
}

func normalizeRuntimeBuildRefToken(token string) string {
	normalized := trimMatchingQuotes(strings.TrimSpace(token))
	if normalized == "" {
		return ""
	}
	normalized = strings.TrimPrefix(normalized, "refs/heads/")
	normalized = strings.TrimPrefix(normalized, "origin/")
	if normalized == "" {
		return ""
	}
	if normalized == "." || normalized == ".." || normalized == "/" {
		return ""
	}
	if strings.HasPrefix(normalized, "/") || strings.HasSuffix(normalized, "/") {
		return ""
	}
	if strings.Contains(normalized, "//") {
		return ""
	}
	if strings.HasPrefix(normalized, "-") {
		return ""
	}
	if strings.ContainsAny(normalized, " \t\r\n") {
		return ""
	}
	if !runtimeBuildRefAllowedPattern.MatchString(normalized) {
		return ""
	}
	return normalized
}

func trimMatchingQuotes(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < 2 {
		return trimmed
	}
	first := trimmed[0]
	last := trimmed[len(trimmed)-1]
	if (first == '\'' && last == '\'') || (first == '"' && last == '"') {
		return strings.TrimSpace(trimmed[1 : len(trimmed)-1])
	}
	return trimmed
}
