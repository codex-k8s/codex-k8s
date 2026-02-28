package servicescfg

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

func deepMergeMaps(base map[string]any, overlay map[string]any) map[string]any {
	out := cloneMapAny(base)
	for key, value := range overlay {
		existing, ok := out[key]
		if !ok {
			out[key] = cloneAny(value)
			continue
		}
		out[key] = deepMergeValue(existing, value)
	}
	return out
}

func deepMergeValue(base any, overlay any) any {
	baseMap, baseIsMap := base.(map[string]any)
	overlayMap, overlayIsMap := overlay.(map[string]any)
	if baseIsMap && overlayIsMap {
		return deepMergeMaps(baseMap, overlayMap)
	}
	return cloneAny(overlay)
}

func cloneAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMapAny(typed)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, cloneAny(item))
		}
		return out
	default:
		return typed
	}
}

func cloneMapAny(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = cloneAny(value)
	}
	return out
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func renderTemplate(name string, raw []byte, ctx ResolvedContext) ([]byte, error) {
	funcs := template.FuncMap{
		"default": func(value string, def string) string {
			if strings.TrimSpace(value) == "" {
				return def
			}
			return value
		},
		"envOr": func(key string, def string) string {
			if value, ok := ctx.Vars[key]; ok && value != "" {
				return value
			}
			if value := os.Getenv(key); value != "" {
				return value
			}
			return def
		},
		"ternary": func(cond bool, left any, right any) any {
			if cond {
				return left
			}
			return right
		},
		"join":       strings.Join,
		"trimPrefix": strings.TrimPrefix,
		"toLower":    strings.ToLower,
	}

	tmpl, err := template.New(name).Funcs(funcs).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse template %q: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return nil, fmt.Errorf("execute template %q: %w", name, err)
	}
	return buf.Bytes(), nil
}
