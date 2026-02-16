package docset

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"strings"
)

type ImportPlanFile struct {
	SrcPath        string
	DstPath        string
	ExpectedSHA256 string
}

type ImportPlan struct {
	Files []ImportPlanFile
}

func BuildImportPlan(m Manifest, locale string, groupIDs []string) (ImportPlan, []string, error) {
	selected := normalizeIDs(groupIDs)
	if len(selected) == 0 {
		// Default behavior: pick default_selected groups.
		for _, g := range m.Groups {
			if g.DefaultSelected {
				// Epic requirement: examples group must not be imported by default.
				if strings.EqualFold(strings.TrimSpace(g.ID), "examples") {
					continue
				}
				selected = append(selected, g.ID)
			}
		}
		selected = normalizeIDs(selected)
	}

	groupsByID := make(map[string]ManifestGroup, len(m.Groups))
	for _, g := range m.Groups {
		groupsByID[g.ID] = g
	}
	for _, id := range selected {
		if _, ok := groupsByID[id]; !ok {
			return ImportPlan{}, nil, fmt.Errorf("unknown group id %q", id)
		}
	}

	seen := make(map[string]struct{})
	files := make([]ImportPlanFile, 0)
	for _, groupID := range selected {
		g := groupsByID[groupID]
		for _, item := range g.Items {
			dst := strings.TrimSpace(item.ImportPath)
			if err := validateImportPath(dst); err != nil {
				return ImportPlan{}, nil, fmt.Errorf("invalid import_path %q: %w", dst, err)
			}
			if _, ok := seen[dst]; ok {
				continue
			}
			seen[dst] = struct{}{}

			src := item.SourcePaths.ForLocale(locale)
			if err := validateSourcePath(src); err != nil {
				return ImportPlan{}, nil, fmt.Errorf("invalid source_path %q: %w", src, err)
			}
			wantSHA := strings.TrimSpace(item.SHA256.ForLocale(locale))
			files = append(files, ImportPlanFile{
				SrcPath:        src,
				DstPath:        dst,
				ExpectedSHA256: wantSHA,
			})
		}
	}

	return ImportPlan{Files: files}, selected, nil
}

func validateImportPath(p string) error {
	p = strings.TrimSpace(p)
	if p == "" {
		return fmt.Errorf("empty")
	}
	if strings.HasPrefix(p, "/") {
		return fmt.Errorf("must be relative")
	}
	clean := path.Clean(p)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("path traversal")
	}
	if clean != p {
		return fmt.Errorf("path must be normalized")
	}
	return nil
}

func validateSourcePath(p string) error {
	// Same rules as import path: we do not allow traversal.
	return validateImportPath(p)
}

func normalizeIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func SHA256Hex(blob []byte) string {
	sum := sha256.Sum256(blob)
	return hex.EncodeToString(sum[:])
}
