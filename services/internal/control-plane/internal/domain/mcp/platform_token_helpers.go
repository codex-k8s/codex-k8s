package mcp

import (
	"context"
	"fmt"
	"strings"
)

func (s *Service) loadBotToken(ctx context.Context) (string, error) {
	item, found, err := s.platform.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("get platform github tokens: %w", err)
	}
	if !found || len(item.BotTokenEncrypted) == 0 {
		return "", fmt.Errorf("bot token is not configured")
	}

	token, err := s.tokenCrypt.DecryptString(item.BotTokenEncrypted)
	if err != nil {
		return "", fmt.Errorf("decrypt bot token: %w", err)
	}
	if strings.TrimSpace(token) == "" {
		return "", fmt.Errorf("bot token is not configured")
	}
	return token, nil
}
