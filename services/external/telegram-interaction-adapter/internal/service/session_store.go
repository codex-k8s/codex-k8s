package service

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SessionStore persists adapter-local callback state.
type SessionStore interface {
	Upsert(SessionRecord) error
	GetByHandle(handle string) (SessionRecord, bool)
	GetByInteractionID(interactionID string) (SessionRecord, bool)
	GetByReply(chatID int64, messageID string) (SessionRecord, bool)
	GetSingleOpenByChat(chatID int64) (SessionRecord, bool)
	CleanupExpired(now time.Time) error
}

// SessionRecord stores callback handles and callback endpoint data for one interaction.
type SessionRecord struct {
	InteractionID           string               `json:"interaction_id"`
	DeliveryID              string               `json:"delivery_id"`
	RecipientRef            string               `json:"recipient_ref"`
	Locale                  string               `json:"locale"`
	CallbackURL             string               `json:"callback_url"`
	CallbackBearerToken     string               `json:"callback_bearer_token"`
	ChatID                  int64                `json:"chat_id"`
	PrimaryMessageID        string               `json:"primary_message_id"`
	ProviderMessageRef      ProviderMessageRef   `json:"provider_message_ref"`
	OptionHandleHashes      map[string]time.Time `json:"option_handle_hashes"`
	FreeTextHandle          string               `json:"-"`
	FreeTextHandleHash      string               `json:"free_text_handle_hash,omitempty"`
	EncryptedFreeTextHandle string               `json:"encrypted_free_text_handle,omitempty"`
	FreeTextExpiresAt       *time.Time           `json:"free_text_expires_at,omitempty"`
	ExpiresAt               time.Time            `json:"expires_at"`
	UpdatedAt               time.Time            `json:"updated_at"`
}

type fileSessionStore struct {
	mu       sync.RWMutex
	path     string
	secret   string
	logger   *slog.Logger
	sessions map[string]SessionRecord
}

type persistedSessionStore struct {
	Sessions map[string]SessionRecord `json:"sessions"`
}

// NewFileSessionStore creates a JSON file-backed session store.
func NewFileSessionStore(path string, secret string, logger *slog.Logger) (SessionStore, error) {
	store := &fileSessionStore{
		path:     strings.TrimSpace(path),
		secret:   strings.TrimSpace(secret),
		logger:   logger,
		sessions: map[string]SessionRecord{},
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *fileSessionStore) Upsert(record SessionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record.InteractionID = strings.TrimSpace(record.InteractionID)
	record.DeliveryID = strings.TrimSpace(record.DeliveryID)
	record.RecipientRef = strings.TrimSpace(record.RecipientRef)
	record.Locale = strings.TrimSpace(record.Locale)
	record.CallbackURL = strings.TrimSpace(record.CallbackURL)
	record.CallbackBearerToken = strings.TrimSpace(record.CallbackBearerToken)
	record.PrimaryMessageID = strings.TrimSpace(record.PrimaryMessageID)
	record.ProviderMessageRef.ChatRef = strings.TrimSpace(record.ProviderMessageRef.ChatRef)
	record.ProviderMessageRef.MessageID = strings.TrimSpace(record.ProviderMessageRef.MessageID)
	record.ProviderMessageRef.InlineMessageID = strings.TrimSpace(record.ProviderMessageRef.InlineMessageID)
	if record.OptionHandleHashes == nil {
		record.OptionHandleHashes = map[string]time.Time{}
	}
	if strings.TrimSpace(record.FreeTextHandle) != "" {
		record.FreeTextHandleHash = hashInteractionHandle(record.FreeTextHandle)
		encrypted, err := encryptSessionValue(s.secret, record.FreeTextHandle)
		if err != nil {
			return fmt.Errorf("encrypt free text handle: %w", err)
		}
		record.EncryptedFreeTextHandle = encrypted
	}
	record.UpdatedAt = time.Now().UTC()
	s.sessions[strings.TrimSpace(record.InteractionID)] = record
	return s.saveLocked()
}

func (s *fileSessionStore) GetByHandle(handle string) (SessionRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, session := range s.sessions {
		hashedHandle := hashInteractionHandle(handle)
		if _, ok := session.OptionHandleHashes[hashedHandle]; ok {
			return session, true
		}
		if strings.TrimSpace(session.FreeTextHandleHash) == hashedHandle {
			return session, true
		}
	}
	return SessionRecord{}, false
}

func (s *fileSessionStore) GetByInteractionID(interactionID string) (SessionRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.sessions[strings.TrimSpace(interactionID)]
	return record, ok
}

func (s *fileSessionStore) GetByReply(chatID int64, messageID string) (SessionRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	normalizedMessageID := strings.TrimSpace(messageID)
	for _, session := range s.sessions {
		if session.ChatID != chatID {
			continue
		}
		if strings.TrimSpace(session.PrimaryMessageID) == normalizedMessageID && strings.TrimSpace(session.FreeTextHandleHash) != "" {
			return session, true
		}
	}
	return SessionRecord{}, false
}

func (s *fileSessionStore) GetSingleOpenByChat(chatID int64) (SessionRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var match SessionRecord
	found := false
	now := time.Now().UTC()
	for _, session := range s.sessions {
		if session.ChatID != chatID || strings.TrimSpace(session.FreeTextHandleHash) == "" || session.ExpiresAt.Before(now) {
			continue
		}
		if found {
			return SessionRecord{}, false
		}
		match = session
		found = true
	}
	return match, found
}

func (s *fileSessionStore) CleanupExpired(now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	changed := false
	for interactionID, session := range s.sessions {
		if session.ExpiresAt.Before(now.UTC()) {
			delete(s.sessions, interactionID)
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return s.saveLocked()
}

func (s *fileSessionStore) load() error {
	if s.path == "" {
		return nil
	}
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read session store %q: %w", s.path, err)
	}
	if len(raw) == 0 {
		return nil
	}

	state := persistedSessionStore{}
	if err := json.Unmarshal(raw, &state); err != nil {
		return fmt.Errorf("unmarshal session store %q: %w", s.path, err)
	}
	if state.Sessions != nil {
		for interactionID, session := range state.Sessions {
			if strings.TrimSpace(session.EncryptedFreeTextHandle) != "" {
				freeTextHandle, err := decryptSessionValue(s.secret, session.EncryptedFreeTextHandle)
				if err != nil {
					return fmt.Errorf("decrypt free text handle for %s: %w", interactionID, err)
				}
				session.FreeTextHandle = freeTextHandle
				if strings.TrimSpace(session.FreeTextHandleHash) == "" {
					session.FreeTextHandleHash = hashInteractionHandle(freeTextHandle)
				}
			}
			if session.OptionHandleHashes == nil {
				session.OptionHandleHashes = map[string]time.Time{}
			}
			state.Sessions[interactionID] = session
		}
		s.sessions = state.Sessions
	}
	return nil
}

func (s *fileSessionStore) saveLocked() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create session store dir: %w", err)
	}
	payload, err := json.MarshalIndent(persistedSessionStore{Sessions: s.sessions}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session store: %w", err)
	}
	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o600); err != nil {
		return fmt.Errorf("write temp session store: %w", err)
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("rename temp session store: %w", err)
	}
	return nil
}

func hashInteractionHandle(handle string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(handle)))
	return hex.EncodeToString(sum[:])
}

func encryptSessionValue(secret string, value string) (string, error) {
	block, err := aes.NewCipher(sessionEncryptionKey(secret))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := cryptorand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(value), nil)
	payload := append(nonce, ciphertext...)
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decryptSessionValue(secret string, encoded string) (string, error) {
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(sessionEncryptionKey(secret))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < gcm.NonceSize() {
		return "", fmt.Errorf("encrypted session payload is too short")
	}
	nonce := payload[:gcm.NonceSize()]
	ciphertext := payload[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func sessionEncryptionKey(secret string) []byte {
	sum := sha256.Sum256([]byte(strings.TrimSpace(secret)))
	return sum[:]
}
