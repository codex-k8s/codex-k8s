package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/crypto/tokencrypt"
	platformtokenrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/platformtoken"
	repocfgrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/repocfg"
)

type syncGitHubTokensParams struct {
	PlatformTokenRaw string
	BotTokenRaw      string
	PlatformTokens   platformtokenrepo.Repository
	Repos            repocfgrepo.Repository
	TokenCrypt       *tokencrypt.Service
	Logger           *slog.Logger
}

func syncGitHubTokens(ctx context.Context, params syncGitHubTokensParams) error {
	if params.PlatformTokens == nil {
		return fmt.Errorf("platform tokens repository is required")
	}
	if params.Repos == nil {
		return fmt.Errorf("repositories repository is required")
	}
	if params.TokenCrypt == nil {
		return fmt.Errorf("token crypt service is required")
	}
	logger := params.Logger
	if logger == nil {
		logger = slog.Default()
	}

	current, found, err := params.PlatformTokens.Get(ctx)
	if err != nil {
		return fmt.Errorf("get platform github tokens: %w", err)
	}

	currentPlatformToken, err := decryptOptionalToken(params.TokenCrypt, current.PlatformTokenEncrypted)
	if err != nil {
		return fmt.Errorf("decrypt stored platform token: %w", err)
	}
	currentBotToken, err := decryptOptionalToken(params.TokenCrypt, current.BotTokenEncrypted)
	if err != nil {
		return fmt.Errorf("decrypt stored bot token: %w", err)
	}

	envPlatformToken := strings.TrimSpace(params.PlatformTokenRaw)
	envBotToken := strings.TrimSpace(params.BotTokenRaw)

	effectivePlatformToken := currentPlatformToken
	effectiveBotToken := currentBotToken
	platformUpdatedFromEnv := false
	botUpdatedFromEnv := false

	if envPlatformToken != "" && envPlatformToken != currentPlatformToken {
		effectivePlatformToken = envPlatformToken
		platformUpdatedFromEnv = true
	}
	if envBotToken != "" && envBotToken != currentBotToken {
		effectiveBotToken = envBotToken
		botUpdatedFromEnv = true
	}

	if !found || platformUpdatedFromEnv || botUpdatedFromEnv {
		platformEncrypted, err := encryptOptionalToken(params.TokenCrypt, effectivePlatformToken)
		if err != nil {
			return fmt.Errorf("encrypt platform token for upsert: %w", err)
		}
		botEncrypted, err := encryptOptionalToken(params.TokenCrypt, effectiveBotToken)
		if err != nil {
			return fmt.Errorf("encrypt bot token for upsert: %w", err)
		}

		if _, err := params.PlatformTokens.Upsert(ctx, platformtokenrepo.UpsertParams{
			PlatformTokenEncrypted: platformEncrypted,
			BotTokenEncrypted:      botEncrypted,
		}); err != nil {
			return fmt.Errorf("upsert platform github tokens: %w", err)
		}
	}

	if strings.TrimSpace(effectivePlatformToken) != "" {
		platformTokenEncrypted, err := params.TokenCrypt.EncryptString(effectivePlatformToken)
		if err != nil {
			return fmt.Errorf("encrypt platform token for repositories: %w", err)
		}

		affected, err := params.Repos.SetTokenEncryptedForAll(ctx, platformTokenEncrypted)
		if err != nil {
			return fmt.Errorf("sync repositories token_encrypted: %w", err)
		}
		logger.Info("repositories token_encrypted synced from platform token", "updated_rows", affected)
	}

	logger.Info(
		"platform github tokens sync completed",
		"platform_updated_from_env", platformUpdatedFromEnv,
		"bot_updated_from_env", botUpdatedFromEnv,
		"platform_token_present", strings.TrimSpace(effectivePlatformToken) != "",
		"bot_token_present", strings.TrimSpace(effectiveBotToken) != "",
	)
	return nil
}

func decryptOptionalToken(tokenCrypt *tokencrypt.Service, encrypted []byte) (string, error) {
	if len(encrypted) == 0 {
		return "", nil
	}
	value, err := tokenCrypt.DecryptString(encrypted)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func encryptOptionalToken(tokenCrypt *tokencrypt.Service, tokenRaw string) ([]byte, error) {
	trimmed := strings.TrimSpace(tokenRaw)
	if trimmed == "" {
		return nil, nil
	}
	return tokenCrypt.EncryptString(trimmed)
}
