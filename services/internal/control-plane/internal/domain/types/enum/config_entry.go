package enum

// ConfigEntryScope defines where a config entry is applied.
type ConfigEntryScope string

const (
	ConfigEntryScopePlatform   ConfigEntryScope = "platform"
	ConfigEntryScopeProject    ConfigEntryScope = "project"
	ConfigEntryScopeRepository ConfigEntryScope = "repository"
)

// ConfigEntryKind defines config entry value category.
type ConfigEntryKind string

const (
	ConfigEntryKindVariable ConfigEntryKind = "variable"
	ConfigEntryKindSecret   ConfigEntryKind = "secret"
)

// ConfigEntryMutability defines synchronization behavior for existing keys.
type ConfigEntryMutability string

const (
	ConfigEntryMutabilityStartupRequired ConfigEntryMutability = "startup_required"
	ConfigEntryMutabilityRuntimeMutable  ConfigEntryMutability = "runtime_mutable"
)
