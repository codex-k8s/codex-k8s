package systemsettings

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	ChangeChannel                 = "codex_system_settings"
	GitHubRateLimitWaitEnabledKey = "github_rate_limit_wait_enabled"
	defaultReloadTimeout          = 5 * time.Second
	defaultReconnectDelay         = 2 * time.Second
)

// ReloadLoopConfig controls LISTEN/NOTIFY-backed cache refresh behavior.
type ReloadLoopConfig struct {
	DSN            string
	ListenQuery    string
	ReloadTimeout  time.Duration
	ReconnectDelay time.Duration
}

// StartReloadLoop keeps one LISTEN connection open and invokes reload on startup and each notification.
func StartReloadLoop(ctx context.Context, cfg ReloadLoopConfig, logger *slog.Logger, reload func(context.Context) error) error {
	cfg = normalizeReloadLoopConfig(cfg)
	if strings.TrimSpace(cfg.DSN) == "" {
		return fmt.Errorf("reload loop dsn is required")
	}
	if strings.TrimSpace(cfg.ListenQuery) == "" {
		return fmt.Errorf("reload loop listen query is required")
	}
	if reload == nil {
		return fmt.Errorf("reload callback is required")
	}

	if err := runReload(ctx, cfg.ReloadTimeout, reload); err != nil {
		return err
	}

	go func() {
		for {
			if ctx.Err() != nil {
				return
			}

			err := waitForReloadSignal(ctx, cfg, reload)
			if ctx.Err() != nil {
				return
			}
			if logger != nil && err != nil {
				logger.Warn(
					"system settings reload loop reconnecting after notification listener error",
					"channel", ChangeChannel,
					"err", err,
				)
			}

			timer := time.NewTimer(cfg.ReconnectDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}()

	return nil
}

func waitForReloadSignal(ctx context.Context, cfg ReloadLoopConfig, reload func(context.Context) error) error {
	conn, err := pgx.Connect(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("connect notification listener: %w", err)
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, cfg.ListenQuery); err != nil {
		return fmt.Errorf("listen for system settings changes: %w", err)
	}

	if err := runReload(ctx, cfg.ReloadTimeout, reload); err != nil {
		return err
	}

	for {
		if _, err := conn.WaitForNotification(ctx); err != nil {
			return fmt.Errorf("wait for system settings notification: %w", err)
		}
		if err := runReload(ctx, cfg.ReloadTimeout, reload); err != nil {
			return err
		}
	}
}

func runReload(ctx context.Context, timeout time.Duration, reload func(context.Context) error) error {
	reloadCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := reload(reloadCtx); err != nil {
		return fmt.Errorf("reload system settings snapshot: %w", err)
	}
	return nil
}

func normalizeReloadLoopConfig(cfg ReloadLoopConfig) ReloadLoopConfig {
	if cfg.ReloadTimeout <= 0 {
		cfg.ReloadTimeout = defaultReloadTimeout
	}
	if cfg.ReconnectDelay <= 0 {
		cfg.ReconnectDelay = defaultReconnectDelay
	}
	return cfg
}
