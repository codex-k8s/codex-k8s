package docset

import (
	"fmt"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

// BuildSafeSyncPlan compares current sha256 to locked sha and only updates when file has no local changes.
func BuildSafeSyncPlan(lock valuetypes.DocsetLock, newManifest entitytypes.DocsetManifest, locale string, currentSHAByPath map[string]string) (valuetypes.DocsetSyncPlan, error) {
	byPath := make(map[string]entitytypes.DocsetManifestItem, len(newManifest.Items))
	for _, item := range newManifest.Items {
		if item.ImportPath == "" {
			continue
		}
		byPath[item.ImportPath] = item
	}

	out := valuetypes.DocsetSyncPlan{
		Updates: make([]valuetypes.DocsetImportPlanFile, 0),
		Drift:   make([]valuetypes.DocsetSyncDecision, 0),
	}
	for _, f := range lock.Files {
		curSHA, ok := currentSHAByPath[f.Path]
		if !ok || curSHA == "" {
			out.Drift = append(out.Drift, valuetypes.DocsetSyncDecision{Path: f.Path, Action: "drift", Reason: "file missing"})
			continue
		}
		if curSHA != f.SHA256 {
			out.Drift = append(out.Drift, valuetypes.DocsetSyncDecision{Path: f.Path, Action: "drift", Reason: "local modifications detected"})
			continue
		}
		item, ok := byPath[f.Path]
		if !ok {
			out.Drift = append(out.Drift, valuetypes.DocsetSyncDecision{Path: f.Path, Action: "drift", Reason: "file not present in new manifest"})
			continue
		}
		newSHA := item.SHA256.ForLocale(locale)
		if newSHA == "" {
			out.Drift = append(out.Drift, valuetypes.DocsetSyncDecision{Path: f.Path, Action: "drift", Reason: "manifest missing sha256"})
			continue
		}
		if newSHA == f.SHA256 {
			continue
		}
		out.Updates = append(out.Updates, valuetypes.DocsetImportPlanFile{
			SrcPath:        item.SourcePaths.ForLocale(locale),
			DstPath:        f.Path,
			ExpectedSHA256: newSHA,
		})
	}

	return out, nil
}

func UpdateLockForSync(lock valuetypes.DocsetLock, newRef string, updatedFiles []valuetypes.DocsetLockFile) (valuetypes.DocsetLock, error) {
	if lock.LockVersion != 1 {
		return valuetypes.DocsetLock{}, fmt.Errorf("unsupported lock_version=%d", lock.LockVersion)
	}
	next := lock
	next.Docset.Ref = newRef

	updated := make(map[string]valuetypes.DocsetLockFile, len(updatedFiles))
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
