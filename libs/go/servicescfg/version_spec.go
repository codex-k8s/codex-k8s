package servicescfg

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// VersionSpec defines one version entry in spec.versions.
//
// Backward-compatible forms in YAML:
//   - version-as-string: `api-gateway: "0.1.0"`
//   - extended object:
//     api-gateway:
//     value: "0.1.0"
//     bumpOn:
//   - services/external/api-gateway
type VersionSpec struct {
	Value  string   `yaml:"value,omitempty"`
	BumpOn []string `yaml:"bumpOn,omitempty"`
}

type versionSpecAlias struct {
	Value  string   `yaml:"value,omitempty"`
	BumpOn []string `yaml:"bumpOn,omitempty"`
}

// UnmarshalYAML supports both scalar and object forms for spec.versions entries.
func (s *VersionSpec) UnmarshalYAML(node *yaml.Node) error {
	if s == nil {
		return fmt.Errorf("version spec target is nil")
	}

	switch node.Kind {
	case yaml.ScalarNode:
		var value string
		if err := node.Decode(&value); err != nil {
			return err
		}
		s.Value = strings.TrimSpace(value)
		s.BumpOn = nil
		return nil
	case yaml.MappingNode:
		var decoded versionSpecAlias
		if err := node.Decode(&decoded); err != nil {
			return err
		}
		s.Value = strings.TrimSpace(decoded.Value)
		s.BumpOn = normalizeVersionBumpPaths(decoded.BumpOn)
		return nil
	default:
		return fmt.Errorf("version spec must be string or object")
	}
}

// MarshalYAML keeps scalar output for simple entries and object output when bumpOn is configured.
func (s VersionSpec) MarshalYAML() (any, error) {
	value := strings.TrimSpace(s.Value)
	bumpOn := normalizeVersionBumpPaths(s.BumpOn)
	if len(bumpOn) == 0 {
		return value, nil
	}
	return versionSpecAlias{
		Value:  value,
		BumpOn: bumpOn,
	}, nil
}

func normalizeVersions(input map[string]VersionSpec) (map[string]VersionSpec, error) {
	if len(input) == 0 {
		return nil, nil
	}

	normalized := make(map[string]VersionSpec, len(input))
	for rawKey, rawSpec := range input {
		key := strings.TrimSpace(rawKey)
		if key == "" {
			return nil, fmt.Errorf("spec.versions contains empty key")
		}
		if _, exists := normalized[key]; exists {
			return nil, fmt.Errorf("spec.versions contains duplicate key %q", key)
		}

		spec := VersionSpec{
			Value:  strings.TrimSpace(rawSpec.Value),
			BumpOn: normalizeVersionBumpPaths(rawSpec.BumpOn),
		}
		if spec.Value == "" {
			return nil, fmt.Errorf("spec.versions[%q].value is required", key)
		}
		normalized[key] = spec
	}
	return normalized, nil
}

func versionValues(input map[string]VersionSpec) map[string]string {
	if len(input) == 0 {
		return nil
	}
	values := make(map[string]string, len(input))
	for key, spec := range input {
		values[key] = strings.TrimSpace(spec.Value)
	}
	return values
}

func normalizeVersionBumpPaths(paths []string) []string {
	return NormalizeRepositoryRelativePaths(paths)
}

// NormalizeRepositoryRelativePaths trims, deduplicates and sorts repo-relative paths.
func NormalizeRepositoryRelativePaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, raw := range paths {
		normalized := NormalizeRepositoryRelativePath(raw)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

// NormalizeRepositoryRelativePath normalizes one path and rejects unsafe values.
func NormalizeRepositoryRelativePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	cleaned := filepath.ToSlash(filepath.Clean(path))
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "." || cleaned == "/" {
		return ""
	}
	cleaned = strings.TrimPrefix(cleaned, "./")
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "" || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	return cleaned
}
