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

func TestSQLMigrationsDeclareGooseDirectives(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".sql" {
			continue
		}

		content, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}

		text := string(content)
		if !strings.Contains(text, "-- +goose Up") {
			t.Fatalf("migration %s must declare -- +goose Up", name)
		}
		if !strings.Contains(text, "-- +goose Down") {
			t.Fatalf("migration %s must declare -- +goose Down", name)
		}
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
		"ADD COLUMN IF NOT EXISTS lease_owner TEXT",
		"ADD COLUMN IF NOT EXISTS lease_until TIMESTAMPTZ",
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

func TestMissionControlGraphFoundationMigrationContainsWaveOneSchema(t *testing.T) {
	content, err := os.ReadFile("20260316103000_day32_mission_control_graph_foundation.sql")
	if err != nil {
		t.Fatalf("read mission control graph foundation migration: %v", err)
	}

	required := []string{
		"ADD COLUMN IF NOT EXISTS continuity_status TEXT NOT NULL DEFAULT 'complete'",
		"ADD COLUMN IF NOT EXISTS coverage_class TEXT NOT NULL DEFAULT 'open_primary'",
		"'run'",
		"'spawned_run'",
		"'produced_pull_request'",
		"CREATE TABLE IF NOT EXISTS mission_control_continuity_gaps",
		"CREATE TABLE IF NOT EXISTS mission_control_workspace_watermarks",
		"uq_mission_control_continuity_gaps_open_subject_kind",
		"idx_mission_control_workspace_watermarks_latest",
		"entity_kind = 'agent'",
		"active_state = 'archived'",
	}
	text := string(content)
	for _, item := range required {
		if !strings.Contains(text, item) {
			t.Fatalf("mission control graph foundation migration must contain %q", item)
		}
	}
}

func TestGitHubRateLimitFoundationMigrationContainsRequiredGuards(t *testing.T) {
	content, err := os.ReadFile("20260314110000_day30_github_rate_limit_wait_foundation.sql")
	if err != nil {
		t.Fatalf("read github rate-limit foundation migration: %v", err)
	}

	required := []string{
		"CREATE TABLE IF NOT EXISTS github_rate_limit_waits",
		"CREATE TABLE IF NOT EXISTS github_rate_limit_wait_evidence",
		"CREATE UNIQUE INDEX IF NOT EXISTS uq_github_rate_limit_waits_open_run_contour",
		"CREATE UNIQUE INDEX IF NOT EXISTS uq_github_rate_limit_waits_open_dominant_per_run",
		"CREATE INDEX IF NOT EXISTS idx_github_rate_limit_waits_resume_queue",
		"CONSTRAINT chk_github_rate_limit_waits_auto_resume_budget",
		"CONSTRAINT chk_github_rate_limit_waits_resolved_at",
		"CONSTRAINT chk_github_rate_limit_waits_manual_action_terminality",
		"CONSTRAINT chk_github_rate_limit_waits_resume_not_before",
		"waiting_backpressure",
		"github_rate_limit",
		"github_rate_limit_wait",
		"backpressure",
	}
	text := string(content)
	for _, item := range required {
		if !strings.Contains(text, item) {
			t.Fatalf("github rate-limit foundation migration must contain %q", item)
		}
	}
}
