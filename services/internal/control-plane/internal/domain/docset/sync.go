package docset

import "fmt"

type SyncDecision struct {
	Path   string
	Action string // update|drift|missing
	Reason string
}

type SyncPlan struct {
	Updates []ImportPlanFile
	Drift   []SyncDecision
}

// BuildSafeSyncPlan compares current sha256 to locked sha and only updates when file has no local changes.
func BuildSafeSyncPlan(lock Lock, newManifest Manifest, locale string, currentSHAByPath map[string]string) (SyncPlan, error) {
	byPath := make(map[string]ManifestItem, len(newManifest.Items))
	for _, item := range newManifest.Items {
		if item.ImportPath == "" {
			continue
		}
		byPath[item.ImportPath] = item
	}

	out := SyncPlan{Updates: make([]ImportPlanFile, 0), Drift: make([]SyncDecision, 0)}
	for _, f := range lock.Files {
		curSHA, ok := currentSHAByPath[f.Path]
		if !ok || curSHA == "" {
			out.Drift = append(out.Drift, SyncDecision{Path: f.Path, Action: "drift", Reason: "file missing"})
			continue
		}
		if curSHA != f.SHA256 {
			out.Drift = append(out.Drift, SyncDecision{Path: f.Path, Action: "drift", Reason: "local modifications detected"})
			continue
		}
		item, ok := byPath[f.Path]
		if !ok {
			out.Drift = append(out.Drift, SyncDecision{Path: f.Path, Action: "drift", Reason: "file not present in new manifest"})
			continue
		}
		newSHA := item.SHA256.ForLocale(locale)
		if newSHA == "" {
			out.Drift = append(out.Drift, SyncDecision{Path: f.Path, Action: "drift", Reason: "manifest missing sha256"})
			continue
		}
		if newSHA == f.SHA256 {
			continue
		}
		out.Updates = append(out.Updates, ImportPlanFile{
			SrcPath:        item.SourcePaths.ForLocale(locale),
			DstPath:        f.Path,
			ExpectedSHA256: newSHA,
		})
	}

	return out, nil
}

func UpdateLockForSync(lock Lock, newRef string, updatedFiles []LockFile) (Lock, error) {
	if lock.LockVersion != 1 {
		return Lock{}, fmt.Errorf("unsupported lock_version=%d", lock.LockVersion)
	}
	next := lock
	next.Docset.Ref = newRef

	updated := make(map[string]LockFile, len(updatedFiles))
	for _, f := range updatedFiles {
		updated[f.Path] = f
	}
	for i := range next.Files {
		if u, ok := updated[next.Files[i].Path]; ok {
			next.Files[i] = u
		}
	}
	return next, nil
}
