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

func TestMissionControlFoundationMigrationKeepsTenantScopedForeignKeys(t *testing.T) {
	content, err := os.ReadFile("20260313133000_day29_mission_control_foundation.sql")
	if err != nil {
		t.Fatalf("read mission control foundation migration: %v", err)
	}

	required := []string{
		"CONSTRAINT uq_mission_control_entities_project_row",
		"FOREIGN KEY (project_id, source_entity_id)",
		"FOREIGN KEY (project_id, target_entity_id)",
		"FOREIGN KEY (project_id, entity_id)",
		"FOREIGN KEY (project_id, command_id)",
		"CONSTRAINT uq_mission_control_commands_project_row",
	}
	text := string(content)
	for _, item := range required {
		if !strings.Contains(text, item) {
			t.Fatalf("mission control foundation migration must contain %q", item)
		}
	}
}

func TestMissionControlCommandLeaseMigrationAddsClaimColumns(t *testing.T) {
	content, err := os.ReadFile("20260314090000_day29_mission_control_command_leases.sql")
	if err != nil {
		t.Fatalf("read mission control command lease migration: %v", err)
	}

	required := []string{
		"ADD COLUMN lease_owner TEXT",
		"ADD COLUMN lease_until TIMESTAMPTZ",
		"idx_mission_control_commands_claimable",
		"WHERE status IN ('accepted', 'queued')",
	}
	text := string(content)
	for _, item := range required {
		if !strings.Contains(text, item) {
			t.Fatalf("mission control command lease migration must contain %q", item)
		}
	}
}
