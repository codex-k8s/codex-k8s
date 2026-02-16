package query

// RegistryImageListFilter keeps registry list filters.
type RegistryImageListFilter struct {
	Repository        string
	LimitRepositories int
	LimitTags         int
}

// RegistryImageDeleteParams identifies one tag to delete.
type RegistryImageDeleteParams struct {
	Repository string
	Tag        string
}

// RegistryImageCleanupFilter configures registry cleanup operation.
type RegistryImageCleanupFilter struct {
	RepositoryPrefix  string
	LimitRepositories int
	KeepTags          int
	DryRun            bool
}
