package manifesttpl

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"
)

// Render executes Kubernetes manifest template with shared helper functions.
func Render(name string, raw []byte, vars map[string]string) ([]byte, error) {
	tpl, err := template.New(name).Funcs(funcMap(vars)).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse manifest template %q: %w", name, err)
	}

	var out bytes.Buffer
	if err := tpl.Execute(&out, map[string]any{
		"Vars": vars,
	}); err != nil {
		return nil, fmt.Errorf("execute manifest template %q: %w", name, err)
	}

	return out.Bytes(), nil
}

func funcMap(vars map[string]string) template.FuncMap {
	return template.FuncMap{
		"default": func(value any, fallback any) any {
			if isEmpty(value) {
				return fallback
			}
			return value
		},
		"envOr": func(key string, fallback string) string {
			normalized := strings.TrimSpace(key)
			if normalized == "" {
				return fallback
			}
			if value, ok := vars[normalized]; ok && strings.TrimSpace(value) != "" {
				return value
			}
			if value := os.Getenv(normalized); strings.TrimSpace(value) != "" {
				return value
			}
			return fallback
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
}

func isEmpty(value any) bool {
	if value == nil {
		return true
	}

	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) == ""
	case []byte:
		return len(typed) == 0
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return rv.Len() == 0
	case reflect.Interface, reflect.Pointer:
		return rv.IsNil()
	}
	return false
}
