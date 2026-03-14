package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileSessionStore_EncryptsFreeTextHandleAndLoads(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "state.json")
	store, err := NewFileSessionStore(path, "webhook-secret:delivery-token", nil)
	if err != nil {
		t.Fatalf("NewFileSessionStore() error = %v", err)
	}

	now := time.Now().UTC()
	record := SessionRecord{
		InteractionID:       "interaction-1",
		DeliveryID:          "delivery-1",
		RecipientRef:        "github_login:tester",
		Locale:              "ru",
		CallbackURL:         "http://codex-k8s/api/v1/mcp/interactions/callback",
		CallbackBearerToken: "callback-token",
		ChatID:              101,
		PrimaryMessageID:    "55",
		ProviderMessageRef: ProviderMessageRef{
			ChatRef:   "101",
			MessageID: "55",
		},
		OptionHandleHashes: map[string]time.Time{
			hashInteractionHandle("option-handle"): now.Add(time.Hour),
		},
		FreeTextHandle:    "free-text-handle",
		FreeTextExpiresAt: timePtr(now.Add(time.Hour)),
		ExpiresAt:         now.Add(2 * time.Hour),
	}
	if err := store.Upsert(record); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	rawState, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	stateText := string(rawState)
	if strings.Contains(stateText, "free-text-handle") {
		t.Fatal("session store leaked raw free text handle")
	}
	if strings.Contains(stateText, "option-handle") {
		t.Fatal("session store leaked raw option handle")
	}

	reloaded, err := NewFileSessionStore(path, "webhook-secret:delivery-token", nil)
	if err != nil {
		t.Fatalf("reload NewFileSessionStore() error = %v", err)
	}
	loadedRecord, ok := reloaded.GetSingleOpenByChat(101)
	if !ok {
		t.Fatal("GetSingleOpenByChat() did not find persisted session")
	}
	if loadedRecord.FreeTextHandle != "free-text-handle" {
		t.Fatalf("FreeTextHandle = %q, want %q", loadedRecord.FreeTextHandle, "free-text-handle")
	}
	if _, ok := reloaded.GetByHandle("option-handle"); !ok {
		t.Fatal("GetByHandle() did not resolve hashed option handle")
	}
}

func timePtr(value time.Time) *time.Time {
	return &value
}
