package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codex-k8s/codex-k8s/services/external/telegram-interaction-adapter/internal/service"
	httptransport "github.com/codex-k8s/codex-k8s/services/external/telegram-interaction-adapter/internal/transport/http"
)

// Run starts telegram-interaction-adapter and blocks until shutdown.
func Run() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	appCtx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	telegramHTTPTimeout, err := time.ParseDuration(cfg.TelegramHTTPTimeout)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_HTTP_TIMEOUT: %w", err)
	}
	if telegramHTTPTimeout <= 0 {
		return fmt.Errorf("CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_HTTP_TIMEOUT must be > 0")
	}
	callbackHTTPTimeout, err := time.ParseDuration(cfg.CallbackHTTPTimeout)
	if err != nil {
		return fmt.Errorf("parse CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_CALLBACK_HTTP_TIMEOUT: %w", err)
	}
	if callbackHTTPTimeout <= 0 {
		return fmt.Errorf("CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_CALLBACK_HTTP_TIMEOUT must be > 0")
	}

	sessionStore, err := service.NewFileSessionStore(
		cfg.TelegramStatePath,
		cfg.TelegramWebhookSecret+":"+cfg.TelegramDeliveryBearerToken,
		logger,
	)
	if err != nil {
		return fmt.Errorf("init telegram adapter session store: %w", err)
	}

	recipientResolver, err := service.NewRecipientResolver(cfg.TelegramChatID, cfg.TelegramRecipientBindingsJSON)
	if err != nil {
		return fmt.Errorf("init telegram recipient resolver: %w", err)
	}

	botClient, err := service.NewTelegramBotClient(service.TelegramBotClientConfig{
		Token:   cfg.TelegramBotToken,
		Timeout: telegramHTTPTimeout,
		Logger:  logger,
	})
	if err != nil {
		return fmt.Errorf("init telegram bot client: %w", err)
	}

	adapterService, err := service.New(service.Config{
		PublicBaseURL:  cfg.PublicBaseURL,
		WebhookSecret:  cfg.TelegramWebhookSecret,
		SessionStore:   sessionStore,
		Recipients:     recipientResolver,
		Bot:            botClient,
		CallbackClient: &http.Client{Timeout: callbackHTTPTimeout},
		DeliveryToken:  cfg.TelegramDeliveryBearerToken,
		Logger:         logger,
	})
	if err != nil {
		return fmt.Errorf("init telegram adapter service: %w", err)
	}

	if err := adapterService.SyncWebhook(appCtx); err != nil {
		logger.Warn("telegram webhook sync failed", "err", err)
	}

	server, err := httptransport.NewServer(httptransport.ServerConfig{
		HTTPAddr: cfg.HTTPAddr,
		Service:  adapterService,
		Logger:   logger,
	})
	if err != nil {
		return fmt.Errorf("init telegram adapter http server: %w", err)
	}

	ctx, stop := signal.NotifyContext(appCtx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("telegram-interaction-adapter started", "addr", cfg.HTTPAddr)
		serverErr <- server.Start()
	}()

	return waitForServerLifecycle(ctx, appCtx, logger, serverErr, "telegram-interaction-adapter", server.Shutdown)
}

func waitForServerLifecycle(ctx context.Context, appCtx context.Context, logger *slog.Logger, serverErr <-chan error, component string, shutdown func(context.Context) error) error {
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(appCtx, 15*time.Second)
		defer cancel()
		logger.Info("shutting down service", "component", component)
		if err := shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown %s: %w", component, err)
		}
		return nil
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("%s server failed: %w", component, err)
		}
		return nil
	}
}
