package value

// DocsetImportPlanFile describes one file operation during docset import.
type DocsetImportPlanFile struct {
	SrcPath        string
	DstPath        string
	ExpectedSHA256 string
}

// DocsetImportPlan is a set of files selected for import.
type DocsetImportPlan struct {
	Files []DocsetImportPlanFile
}

// DocsetSyncDecision describes one drift decision during safe sync.
type DocsetSyncDecision struct {
	Path   string
	Action string // update|drift|missing
	Reason string
}

// DocsetSyncPlan contains update and drift actions for sync.
type DocsetSyncPlan struct {
	Updates []DocsetImportPlanFile
	Drift   []DocsetSyncDecision
}
