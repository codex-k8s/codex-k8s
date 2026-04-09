package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// RecipientResolver resolves opaque platform recipient refs into Telegram chat ids.
type RecipientResolver struct {
	defaultChatID *int64
	mappings      map[string]int64
	allowedChats  map[int64]struct{}
}

// NewRecipientResolver parses default and per-login chat bindings.
func NewRecipientResolver(defaultChatID string, bindingsJSON string) (*RecipientResolver, error) {
	resolver := &RecipientResolver{
		mappings:     map[string]int64{},
		allowedChats: map[int64]struct{}{},
	}

	if value := strings.TrimSpace(defaultChatID); value != "" {
		parsed, err := parseTelegramChatID(value)
		if err != nil {
			return nil, fmt.Errorf("parse KODEX_TELEGRAM_CHAT_ID: %w", err)
		}
		resolver.defaultChatID = &parsed
		resolver.allowedChats[parsed] = struct{}{}
	}

	if strings.TrimSpace(bindingsJSON) == "" {
		return resolver, nil
	}

	items := map[string]json.RawMessage{}
	if err := json.Unmarshal([]byte(bindingsJSON), &items); err != nil {
		return nil, fmt.Errorf("parse KODEX_TELEGRAM_INTERACTION_ADAPTER_RECIPIENT_BINDINGS_JSON: %w", err)
	}
	for login, raw := range items {
		normalizedLogin := strings.TrimSpace(login)
		if normalizedLogin == "" {
			continue
		}
		chatID, err := parseTelegramChatID(strings.Trim(string(raw), "\""))
		if err != nil {
			var numeric int64
			if unmarshalErr := json.Unmarshal(raw, &numeric); unmarshalErr != nil {
				return nil, fmt.Errorf("parse telegram binding for %s: %w", normalizedLogin, err)
			}
			chatID = numeric
		}
		resolver.mappings[normalizedLogin] = chatID
		resolver.allowedChats[chatID] = struct{}{}
	}

	return resolver, nil
}

// Resolve converts opaque platform recipient refs into a Telegram chat id.
func (r *RecipientResolver) Resolve(recipientRef string) (int64, error) {
	normalized := strings.TrimSpace(recipientRef)
	switch {
	case strings.HasPrefix(normalized, recipientRefPrefixGitHubLogin):
		login := strings.TrimSpace(strings.TrimPrefix(normalized, recipientRefPrefixGitHubLogin))
		if chatID, ok := r.mappings[login]; ok {
			return chatID, nil
		}
		if r.defaultChatID != nil {
			return *r.defaultChatID, nil
		}
		return 0, fmt.Errorf("telegram recipient mapping for github login %q is not configured", login)
	case strings.HasPrefix(normalized, recipientRefPrefixChatID):
		return parseTelegramChatID(strings.TrimPrefix(normalized, recipientRefPrefixChatID))
	default:
		if normalized == "" {
			return 0, fmt.Errorf("recipient_ref is required")
		}
		return parseTelegramChatID(normalized)
	}
}

// IsAllowedChat reports whether webhook update chat is allowed for this adapter.
func (r *RecipientResolver) IsAllowedChat(chatID int64) bool {
	if len(r.allowedChats) == 0 {
		return true
	}
	_, ok := r.allowedChats[chatID]
	return ok
}

func parseTelegramChatID(value string) (int64, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid telegram chat id %q: %w", value, err)
	}
	return parsed, nil
}
