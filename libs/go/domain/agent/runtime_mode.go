package agent

import "strings"

// RuntimeMode defines agent execution profile.
type RuntimeMode string

const (
	RuntimeModeFullEnv  RuntimeMode = "full-env"
	RuntimeModeCodeOnly RuntimeMode = "code-only"
)

// ParseRuntimeMode normalizes runtime mode string and falls back to code-only.
func ParseRuntimeMode(value string) RuntimeMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(RuntimeModeFullEnv):
		return RuntimeModeFullEnv
	default:
		return RuntimeModeCodeOnly
	}
}
