package query

// DocsetGroup is one UI-visible docset group entry.
type DocsetGroup struct {
	ID              string
	Title           string
	Description     string
	DefaultSelected bool
}

// DocsetImportResult contains GitHub PR metadata for docset import.
type DocsetImportResult struct {
	RepositoryFullName string
	PRNumber           int
	PRURL              string
	Branch             string
	FilesTotal         int
}

// DocsetSyncResult contains GitHub PR metadata for docset safe sync.
type DocsetSyncResult struct {
	RepositoryFullName string
	PRNumber           int
	PRURL              string
	Branch             string
	FilesUpdated       int
	FilesDrift         int
}
