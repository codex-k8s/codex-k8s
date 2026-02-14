package envfile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Load reads KEY=VALUE env file preserving unknown keys and supporting comments.
func Load(path string) (map[string]string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve env file path: %w", err)
	}
	file, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("open env file %q: %w", absPath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	out := make(map[string]string, 64)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := parseEnvLine(line)
		if !ok {
			continue
		}
		out[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan env file %q: %w", absPath, err)
	}
	return out, nil
}

// Save writes deterministic sorted KEY=VALUE env file.
func Save(path string, vars map[string]string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve env file path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return fmt.Errorf("mkdir env file parent: %w", err)
	}
	lines := make([]string, 0, len(vars))
	for key, value := range vars {
		lines = append(lines, fmt.Sprintf("%s=%q", key, value))
	}
	sortStrings(lines)
	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}
	if err := os.WriteFile(absPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write env file %q: %w", absPath, err)
	}
	return nil
}

func parseEnvLine(line string) (string, string, bool) {
	idx := strings.IndexRune(line, '=')
	if idx <= 0 {
		return "", "", false
	}
	key := strings.TrimSpace(line[:idx])
	if key == "" {
		return "", "", false
	}
	rawValue := strings.TrimSpace(line[idx+1:])
	value := strings.Trim(rawValue, `"'`)
	return key, os.ExpandEnv(value), true
}

func sortStrings(values []string) {
	if len(values) < 2 {
		return
	}
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[i] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}
