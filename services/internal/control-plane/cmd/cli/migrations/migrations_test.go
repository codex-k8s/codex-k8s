package migrations_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationVersionsAreUnique(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	versions := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".sql" {
			continue
		}
		version, _, ok := strings.Cut(name, "_")
		if !ok || strings.TrimSpace(version) == "" {
			t.Fatalf("migration file %q does not start with version prefix", name)
		}
		if prev, exists := versions[version]; exists {
			t.Fatalf("duplicate migration version %s: %s and %s", version, prev, name)
		}
		versions[version] = name
	}
}
