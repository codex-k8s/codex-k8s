package value

// DocsetLock mirrors docs/.docset-lock.json contract.
type DocsetLock struct {
	LockVersion int              `json:"lock_version"`
	Docset      DocsetLockDocset `json:"docset"`
	Files       []DocsetLockFile `json:"files"`
}

// DocsetLockDocset contains lock metadata for imported docset.
type DocsetLockDocset struct {
	ID             string   `json:"id"`
	Ref            string   `json:"ref"`
	Locale         string   `json:"locale"`
	SelectedGroups []string `json:"selected_groups"`
}

// DocsetLockFile tracks one imported file and source hash.
type DocsetLockFile struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	SourcePath string `json:"source_path"`
}
