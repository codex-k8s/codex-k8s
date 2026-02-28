package servicescfg

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func loadMergedMap(path string, trail []string) (map[string]any, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve import path %q: %w", path, err)
	}
	for _, entry := range trail {
		if entry == absPath {
			chain := append(append([]string(nil), trail...), absPath)
			return nil, fmt.Errorf("imports cycle detected: %s", strings.Join(chain, " -> "))
		}
	}

	raw, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read services file %q: %w", absPath, err)
	}
	decoded, err := decodeYAMLToMap(raw)
	if err != nil {
		return nil, fmt.Errorf("decode services file %q: %w", absPath, err)
	}
	importPaths, err := extractImportPaths(decoded)
	if err != nil {
		return nil, fmt.Errorf("parse imports in %q: %w", absPath, err)
	}
	clearImports(decoded)

	merged := make(map[string]any)
	baseDir := filepath.Dir(absPath)
	for _, importPath := range importPaths {
		expanded, err := expandImportPaths(baseDir, importPath)
		if err != nil {
			return nil, fmt.Errorf("resolve import %q from %q: %w", importPath, absPath, err)
		}
		for _, candidate := range expanded {
			child, err := loadMergedMap(candidate, append(trail, absPath))
			if err != nil {
				return nil, err
			}
			merged = deepMergeMaps(merged, child)
		}
	}

	merged = deepMergeMaps(merged, decoded)
	return merged, nil
}

func expandImportPaths(baseDir string, pattern string) ([]string, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil, fmt.Errorf("empty import path")
	}

	resolved := pattern
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(baseDir, pattern)
	}
	matches, err := filepath.Glob(resolved)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no files matched import %q", pattern)
	}

	sort.Strings(matches)
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		stat, err := os.Stat(match)
		if err != nil {
			return nil, err
		}
		if stat.IsDir() {
			continue
		}
		out = append(out, match)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("import %q resolved only directories", pattern)
	}
	return out, nil
}

func decodeYAMLToMap(raw []byte) (map[string]any, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	merged := make(map[string]any)

	for {
		var doc map[string]any
		if err := decoder.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if len(doc) == 0 {
			continue
		}
		merged = deepMergeMaps(merged, doc)
	}
	return merged, nil
}

func extractImportPaths(doc map[string]any) ([]string, error) {
	spec, ok := doc["spec"].(map[string]any)
	if !ok {
		return nil, nil
	}
	rawImports, ok := spec["imports"]
	if !ok {
		return nil, nil
	}
	items, ok := rawImports.([]any)
	if !ok {
		return nil, fmt.Errorf("spec.imports must be a list")
	}

	var paths []string
	for _, item := range items {
		switch typed := item.(type) {
		case string:
			if strings.TrimSpace(typed) == "" {
				return nil, fmt.Errorf("spec.imports contains empty path")
			}
			paths = append(paths, typed)
		case map[string]any:
			path, _ := typed["path"].(string)
			path = strings.TrimSpace(path)
			if path == "" {
				return nil, fmt.Errorf("spec.imports item must define path")
			}
			paths = append(paths, path)
		default:
			return nil, fmt.Errorf("spec.imports item has unsupported type %T", item)
		}
	}
	return paths, nil
}

func clearImports(doc map[string]any) {
	spec, ok := doc["spec"].(map[string]any)
	if !ok {
		return
	}
	delete(spec, "imports")
	if len(spec) == 0 {
		delete(doc, "spec")
	}
}
