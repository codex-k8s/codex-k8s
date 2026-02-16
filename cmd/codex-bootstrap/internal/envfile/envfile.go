package envfile

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Load reads KEY=VALUE variables from an env file.
func Load(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open env file %q: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("%s:%d: expected KEY=VALUE", path, lineNo)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("%s:%d: empty key", path, lineNo)
		}
		value = strings.TrimSpace(value)
		value = strings.Trim(value, "'")
		value = strings.Trim(value, "\"")
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan env file %q: %w", path, err)
	}
	return values, nil
}

// Upsert writes non-empty KEY=VALUE pairs into env file and preserves other lines.
func Upsert(path string, values map[string]string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("env file path is required")
	}
	if len(values) == 0 {
		return nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read env file %q: %w", path, err)
	}
	lines := strings.Split(string(raw), "\n")
	indexByKey := make(map[string]int, len(lines))
	for idx, line := range lines {
		key := parseLineKey(line)
		if key == "" {
			continue
		}
		indexByKey[key] = idx
	}

	keys := make([]string, 0, len(values))
	for key, value := range values {
		key = strings.TrimSpace(key)
		if key == "" || strings.TrimSpace(value) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := strings.TrimSpace(values[key])
		escaped := strings.ReplaceAll(value, "'", "'\\''")
		rendered := fmt.Sprintf("%s='%s'", key, escaped)
		if idx, ok := indexByKey[key]; ok {
			lines[idx] = rendered
			continue
		}
		lines = append(lines, rendered)
	}

	output := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(output), 0o600); err != nil {
		return fmt.Errorf("write env file %q: %w", path, err)
	}
	return nil
}

func parseLineKey(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return ""
	}
	if strings.HasPrefix(trimmed, "export ") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "export "))
	}
	key, _, ok := strings.Cut(trimmed, "=")
	if !ok {
		return ""
	}
	return strings.TrimSpace(key)
}
