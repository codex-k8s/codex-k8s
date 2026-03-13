package agent

import "strings"

// RuntimeAccessProfile defines Kubernetes access profile for one runtime session.
type RuntimeAccessProfile string

const (
	RuntimeAccessProfileCandidate          RuntimeAccessProfile = "candidate"
	RuntimeAccessProfileProductionReadOnly RuntimeAccessProfile = "production-readonly"
)

// ParseRuntimeAccessProfile normalizes access profile string and falls back to candidate.
func ParseRuntimeAccessProfile(value string) RuntimeAccessProfile {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(RuntimeAccessProfileProductionReadOnly):
		return RuntimeAccessProfileProductionReadOnly
	default:
		return RuntimeAccessProfileCandidate
	}
}
