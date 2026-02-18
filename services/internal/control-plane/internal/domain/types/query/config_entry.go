package query

import enumtypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/enum"

// ConfigEntryUpsertParams defines inputs for configuration entry upsert.
type ConfigEntryUpsertParams struct {
	Scope enumtypes.ConfigEntryScope
	Kind  enumtypes.ConfigEntryKind

	ProjectID    string
	RepositoryID string

	Key string

	// ValuePlain is used for variables.
	ValuePlain string
	// ValueEncrypted is used for secrets.
	ValueEncrypted []byte

	SyncTargets []string
	Mutability  enumtypes.ConfigEntryMutability
	IsDangerous bool

	CreatedByUserID string
	UpdatedByUserID string
}

type ConfigEntryListFilter struct {
	Scope        enumtypes.ConfigEntryScope
	ProjectID    string
	RepositoryID string
	Limit        int
}
