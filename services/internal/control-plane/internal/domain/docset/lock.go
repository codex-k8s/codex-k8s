package docset

import (
	"encoding/json"
	"fmt"

	valuetypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/value"
)

func NewLock(docsetID string, ref string, locale string, selectedGroups []string, files []valuetypes.DocsetLockFile) valuetypes.DocsetLock {
	return valuetypes.DocsetLock{
		LockVersion: 1,
		Docset: valuetypes.DocsetLockDocset{
			ID:             docsetID,
			Ref:            ref,
			Locale:         locale,
			SelectedGroups: append([]string(nil), selectedGroups...),
		},
		Files: append([]valuetypes.DocsetLockFile(nil), files...),
	}
}

func ParseLock(blob []byte) (valuetypes.DocsetLock, error) {
	var lock valuetypes.DocsetLock
	if err := json.Unmarshal(blob, &lock); err != nil {
		return valuetypes.DocsetLock{}, fmt.Errorf("parse docset lock json: %w", err)
	}
	if lock.LockVersion != 1 {
		return valuetypes.DocsetLock{}, fmt.Errorf("unsupported lock_version=%d (expected 1)", lock.LockVersion)
	}
	return lock, nil
}

func MarshalLock(lock valuetypes.DocsetLock) ([]byte, error) {
	blob, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal docset lock json: %w", err)
	}
	blob = append(blob, '\n')
	return blob, nil
}
