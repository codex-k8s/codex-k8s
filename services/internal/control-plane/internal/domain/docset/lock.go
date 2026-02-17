package docset

import (
	"encoding/json"
	"fmt"
)

type Lock struct {
	LockVersion int `json:"lock_version"`
	Docset      LockDocset `json:"docset"`
	Files       []LockFile `json:"files"`
}

type LockDocset struct {
	ID            string   `json:"id"`
	Ref           string   `json:"ref"`
	Locale        string   `json:"locale"`
	SelectedGroups []string `json:"selected_groups"`
}

type LockFile struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	SourcePath string `json:"source_path"`
}

func NewLock(docsetID string, ref string, locale string, selectedGroups []string, files []LockFile) Lock {
	return Lock{
		LockVersion: 1,
		Docset: LockDocset{
			ID:             docsetID,
			Ref:            ref,
			Locale:         locale,
			SelectedGroups: append([]string(nil), selectedGroups...),
		},
		Files: append([]LockFile(nil), files...),
	}
}

func ParseLock(blob []byte) (Lock, error) {
	var lock Lock
	if err := json.Unmarshal(blob, &lock); err != nil {
		return Lock{}, fmt.Errorf("parse docset lock json: %w", err)
	}
	if lock.LockVersion != 1 {
		return Lock{}, fmt.Errorf("unsupported lock_version=%d (expected 1)", lock.LockVersion)
	}
	return lock, nil
}

func MarshalLock(lock Lock) ([]byte, error) {
	blob, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal docset lock json: %w", err)
	}
	blob = append(blob, '\n')
	return blob, nil
}

