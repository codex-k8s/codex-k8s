package entity

import "time"

// RegistryImageTag describes one registry tag in one repository.
type RegistryImageTag struct {
	Tag             string
	Digest          string
	CreatedAt       *time.Time
	ConfigSizeBytes int64
}

// RegistryImageRepository groups tags by repository.
type RegistryImageRepository struct {
	Repository string
	TagCount   int
	Tags       []RegistryImageTag
}

// RegistryImageDeleteResult describes one delete operation result.
type RegistryImageDeleteResult struct {
	Repository string
	Tag        string
	Digest     string
	Deleted    bool
}

// RegistryImageCleanupResult describes bulk cleanup execution result.
type RegistryImageCleanupResult struct {
	RepositoriesScanned int
	TagsDeleted         int
	TagsSkipped         int
	Deleted             []RegistryImageDeleteResult
	Skipped             []RegistryImageDeleteResult
}
