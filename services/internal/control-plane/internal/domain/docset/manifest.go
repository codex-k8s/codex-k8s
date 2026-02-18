package docset

import (
	"encoding/json"
	"fmt"
	"strings"

	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

func ParseManifest(blob []byte) (entitytypes.DocsetManifest, error) {
	var m entitytypes.DocsetManifest
	if err := json.Unmarshal(blob, &m); err != nil {
		return entitytypes.DocsetManifest{}, fmt.Errorf("parse docset manifest json: %w", err)
	}
	if m.ManifestVersion != 1 {
		return entitytypes.DocsetManifest{}, fmt.Errorf("unsupported manifest_version=%d (expected 1)", m.ManifestVersion)
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
