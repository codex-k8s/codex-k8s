package servicescfg

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads and parses services configuration YAML.
func Load(path string) (*Config, string, error) {
	if path == "" {
		return nil, "", fmt.Errorf("services config path is empty")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, "", fmt.Errorf("resolve services config path: %w", err)
	}
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return nil, "", fmt.Errorf("read services config %q: %w", absPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, "", fmt.Errorf("parse services config %q: %w", absPath, err)
	}
	if cfg.Project == "" {
		return nil, "", fmt.Errorf("services config %q: project is required", absPath)
	}

	return &cfg, filepath.Dir(absPath), nil
}
