package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	controlplaneclient "github.com/codex-k8s/kodex/services/external/telegram-interaction-adapter/internal/controlplane"
	"github.com/codex-k8s/kodex/services/external/telegram-interaction-adapter/internal/service"
	httptransport "github.com/codex-k8s/kodex/services/external/telegram-interaction-adapter/internal/transport/http"
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
		return fmt.Errorf("parse KODEX_TELEGRAM_INTERACTION_ADAPTER_HTTP_TIMEOUT: %w", err)
	}
	if telegramHTTPTimeout <= 0 {
		return fmt.Errorf("KODEX_TELEGRAM_INTERACTION_ADAPTER_HTTP_TIMEOUT must be > 0")
	}

	sttTimeout, err := time.ParseDuration(cfg.TelegramSTTTimeout)
	if err != nil {
		return fmt.Errorf("parse KODEX_TELEGRAM_INTERACTION_ADAPTER_STT_TIMEOUT: %w", err)
	}
	if sttTimeout <= 0 {
		return fmt.Errorf("KODEX_TELEGRAM_INTERACTION_ADAPTER_STT_TIMEOUT must be > 0")
	}

	controlPlaneClient, err := controlplaneclient.Dial(appCtx, cfg.ControlPlaneGRPCTarget)
	if err != nil {
		return fmt.Errorf("dial control-plane grpc: %w", err)
	}
	defer func() { _ = controlPlaneClient.Close() }()

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

	var speechToText service.SpeechToText
	var audioConverter service.AudioConverter
	if strings.TrimSpace(cfg.OpenAIAPIKey) != "" {
		speechToText = service.NewOpenAISpeechToText(cfg.OpenAIAPIKey, cfg.TelegramSTTModel, sttTimeout, logger)
		audioConverter = service.FFmpegAudioConverter{}
	}

	adapterService, err := service.New(service.Config{
		PublicBaseURL:  cfg.PublicBaseURL,
		WebhookSecret:  cfg.TelegramWebhookSecret,
		DeliveryToken:  cfg.TelegramDeliveryBearerToken,
		Recipients:     recipientResolver,
		Bot:            botClient,
		CallbackSink:   service.NewControlPlaneCallbackSink(controlPlaneClient),
		AudioConverter: audioConverter,
		SpeechToText:   speechToText,
		Logger:         logger,
	})
	if err != nil {
		return fmt.Errorf("init telegram adapter service: %w", err)
	}

	if telegramWebhookSyncEnabled(cfg.Environment) {
		if err := adapterService.SyncWebhook(appCtx); err != nil {
			logger.Warn("telegram webhook sync failed", "err", err)
		}
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

func telegramWebhookSyncEnabled(environment string) bool {
	return !strings.EqualFold(strings.TrimSpace(environment), "ai")
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
		if err != nil {
			return fmt.Errorf("%s server failed: %w", component, err)
		}
		return nil
	}
}
