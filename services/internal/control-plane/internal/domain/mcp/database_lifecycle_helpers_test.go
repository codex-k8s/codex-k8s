package mcp

import (
	"encoding/json"
	"testing"
)

func TestNormalizeDatabaseLifecycleAllowedEnvsDefaults(t *testing.T) {
	allowed := normalizeDatabaseLifecycleAllowedEnvs(nil)
	for _, env := range []string{"dev", "production", "prod"} {
		if _, ok := allowed[env]; !ok {
			t.Fatalf("expected default env %q in allowlist", env)
		}
	}
}

func TestNormalizeDatabaseLifecycleAllowedEnvsCustom(t *testing.T) {
	allowed := normalizeDatabaseLifecycleAllowedEnvs([]string{"  qa  ", "production", "QA", ""})
	if len(allowed) != 2 {
		t.Fatalf("expected 2 unique envs, got %d", len(allowed))
	}
	if _, ok := allowed["qa"]; !ok {
		t.Fatalf("expected normalized env qa")
	}
	if _, ok := allowed["production"]; !ok {
		t.Fatalf("expected env production")
	}
}

func TestNormalizeDatabaseLifecycleName(t *testing.T) {
	valid := []string{"db_main", "db-qa-01", "_internal"}
	for _, item := range valid {
		if _, err := normalizeDatabaseLifecycleName(item); err != nil {
			t.Fatalf("expected valid database name %q: %v", item, err)
		}
	}
	invalid := []string{"", "9db", "db name", "db$name"}
	for _, item := range invalid {
		if _, err := normalizeDatabaseLifecycleName(item); err == nil {
			t.Fatalf("expected invalid database name %q", item)
		}
	}
}

func TestDecodeDatabaseLifecyclePayload(t *testing.T) {
	raw := mustMarshalJSON(t, databaseLifecyclePayload{
		ProjectID:    "0eb03a0f-89cc-4fbe-b944-19949bf32e0e",
		Environment:  "PRODUCTION",
		Action:       DatabaseLifecycleActionDescribe,
		DatabaseName: "codex_stage",
	})
	payload, err := decodeDatabaseLifecyclePayload(raw)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Environment != "production" {
		t.Fatalf("expected normalized environment, got %q", payload.Environment)
	}
	if payload.Action != DatabaseLifecycleActionDescribe {
		t.Fatalf("expected describe action, got %q", payload.Action)
	}
}

func TestDecodeDatabaseLifecyclePayloadDeleteRequiresConfirm(t *testing.T) {
	raw := mustMarshalJSON(t, databaseLifecyclePayload{
		ProjectID:     "0eb03a0f-89cc-4fbe-b944-19949bf32e0e",
		Environment:   "dev",
		Action:        DatabaseLifecycleActionDelete,
		DatabaseName:  "codex_dev",
		ConfirmDelete: false,
	})
	if _, err := decodeDatabaseLifecyclePayload(raw); err == nil {
		t.Fatalf("expected confirm_delete validation error")
	}
}

func mustMarshalJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return raw
}
