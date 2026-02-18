package entity

import enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"

// ConfigEntry is a persisted configuration entry (variable or secret).
//
// Secret values are never returned in plaintext; for secrets Value is always empty.
type ConfigEntry struct {
	ID string

	Scope enumtypes.ConfigEntryScope
	Kind  enumtypes.ConfigEntryKind

	ProjectID    string
	RepositoryID string

	Key string

	// Value is returned only for variables (kind=variable).
	Value string

	SyncTargets []string
	Mutability  enumtypes.ConfigEntryMutability
	IsDangerous bool

	UpdatedAt string
}
