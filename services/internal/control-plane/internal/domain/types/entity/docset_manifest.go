package entity

import "strings"

// DocsetManifest describes docset.manifest.json contract.
type DocsetManifest struct {
	ManifestVersion int                   `json:"manifest_version"`
	Docset          DocsetManifestDocset  `json:"docset"`
	Groups          []DocsetManifestGroup `json:"groups"`
	Items           []DocsetManifestItem  `json:"items"`
}

// DocsetManifestDocset contains docset identity block.
type DocsetManifestDocset struct {
	ID string `json:"id"`
}

// DocsetLocalizedText stores localized values for ru/en.
type DocsetLocalizedText struct {
	RU string `json:"ru"`
	EN string `json:"en"`
}

// ForLocale returns localized value by locale with EN fallback.
func (t DocsetLocalizedText) ForLocale(locale string) string {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "ru":
		return strings.TrimSpace(t.RU)
	default:
		return strings.TrimSpace(t.EN)
	}
}

// DocsetManifestGroup describes one importable group.
type DocsetManifestGroup struct {
	ID              string              `json:"id"`
	Title           DocsetLocalizedText `json:"title"`
	Description     DocsetLocalizedText `json:"description"`
	DefaultSelected bool                `json:"default_selected"`
	ItemIDs         []string            `json:"items"`
}

// DocsetManifestItem describes one importable document/item.
type DocsetManifestItem struct {
	ID          string              `json:"id"`
	Kind        string              `json:"kind"`
	Category    string              `json:"category"`
	Common      bool                `json:"common"`
	StackTags   []string            `json:"stack_tags"`
	ImportPath  string              `json:"import_path"`
	SourcePaths DocsetLocalizedText `json:"source_paths"`
	SHA256      DocsetLocalizedText `json:"sha256"`
	Title       DocsetLocalizedText `json:"title"`
	Description DocsetLocalizedText `json:"description"`
}
