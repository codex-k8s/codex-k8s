package docset

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Manifest struct {
	ManifestVersion int            `json:"manifest_version"`
	Docset          ManifestDocset  `json:"docset"`
	Groups          []ManifestGroup `json:"groups"`
	Items           []ManifestItem  `json:"items"`
}

type ManifestDocset struct {
	ID string `json:"id"`
}

type ManifestGroup struct {
	ID              string        `json:"id"`
	Title           LocalizedText `json:"title"`
	Description     LocalizedText `json:"description"`
	DefaultSelected bool          `json:"default_selected"`
	ItemIDs         []string      `json:"items"`
}

type LocalizedText struct {
	RU string `json:"ru"`
	EN string `json:"en"`
}

type ManifestItem struct {
	ID          string        `json:"id"`
	Kind        string        `json:"kind"`
	Category    string        `json:"category"`
	Common      bool          `json:"common"`
	StackTags   []string      `json:"stack_tags"`
	ImportPath  string        `json:"import_path"`
	SourcePaths LocalizedText `json:"source_paths"`
	SHA256      LocalizedText `json:"sha256"`
	Title       LocalizedText `json:"title"`
	Description LocalizedText `json:"description"`
}

func ParseManifest(blob []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(blob, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse docset manifest json: %w", err)
	}
	if m.ManifestVersion != 1 {
		return Manifest{}, fmt.Errorf("unsupported manifest_version=%d (expected 1)", m.ManifestVersion)
	}
	m.Docset.ID = strings.TrimSpace(m.Docset.ID)
	for i := range m.Groups {
		m.Groups[i].ID = strings.TrimSpace(m.Groups[i].ID)
	}
	for i := range m.Items {
		item := &m.Items[i]
		item.ID = strings.TrimSpace(item.ID)
		item.Kind = strings.TrimSpace(item.Kind)
		item.Category = strings.TrimSpace(item.Category)
		item.ImportPath = strings.TrimSpace(item.ImportPath)
		item.SourcePaths.RU = strings.TrimSpace(item.SourcePaths.RU)
		item.SourcePaths.EN = strings.TrimSpace(item.SourcePaths.EN)
		item.SHA256.RU = normalizeSHA256(item.SHA256.RU)
		item.SHA256.EN = normalizeSHA256(item.SHA256.EN)
	}
	return m, nil
}

func normalizeSHA256(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "sha256:")
	return strings.TrimSpace(raw)
}

func (t LocalizedText) ForLocale(locale string) string {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "ru":
		return strings.TrimSpace(t.RU)
	default:
		return strings.TrimSpace(t.EN)
	}
}
