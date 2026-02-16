package docset

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Manifest struct {
	ManifestVersion int            `json:"manifest_version"`
	ID              string         `json:"id"`
	Groups          []ManifestGroup `json:"groups"`
}

type ManifestGroup struct {
	ID              string             `json:"id"`
	Title           LocalizedText      `json:"title"`
	Description     LocalizedText      `json:"description"`
	DefaultSelected bool               `json:"default_selected"`
	Items           []ManifestGroupItem `json:"items"`
}

type LocalizedText struct {
	RU string `json:"ru"`
	EN string `json:"en"`
}

type ManifestGroupItem struct {
	ImportPath   string        `json:"import_path"`
	SourcePaths  LocalizedText `json:"source_paths"`
	SHA256       LocalizedText `json:"sha256"`
	Category     string        `json:"category"`
	Common       bool          `json:"common"`
	StackTags    []string      `json:"stack_tags"`
	Title        LocalizedText `json:"title"`
	Description  LocalizedText `json:"description"`
}

func ParseManifest(blob []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(blob, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse docset manifest json: %w", err)
	}
	if m.ManifestVersion != 1 {
		return Manifest{}, fmt.Errorf("unsupported manifest_version=%d (expected 1)", m.ManifestVersion)
	}
	for i := range m.Groups {
		m.Groups[i].ID = strings.TrimSpace(m.Groups[i].ID)
		for j := range m.Groups[i].Items {
			m.Groups[i].Items[j].ImportPath = strings.TrimSpace(m.Groups[i].Items[j].ImportPath)
		}
	}
	return m, nil
}

func (t LocalizedText) ForLocale(locale string) string {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "ru":
		return strings.TrimSpace(t.RU)
	default:
		return strings.TrimSpace(t.EN)
	}
}

