package app

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config defines environment-backed runtime settings for telegram-interaction-adapter.
type Config struct {
	HTTPAddr string `env:"CODEXK8S_HTTP_ADDR" envDefault:":8080"`

	PublicBaseURL string `env:"CODEXK8S_PUBLIC_BASE_URL"`

	TelegramBotToken string `env:"CODEXK8S_TELEGRAM_BOT_TOKEN"`
	TelegramChatID   string `env:"CODEXK8S_TELEGRAM_CHAT_ID"`

	TelegramRecipientBindingsJSON string `env:"CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_RECIPIENT_BINDINGS_JSON"`
	TelegramDeliveryBearerToken   string `env:"CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN"`
	TelegramWebhookSecret         string `env:"CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_WEBHOOK_SECRET"`
	TelegramStatePath             string `env:"CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_STATE_PATH" envDefault:"/var/lib/codex-k8s-telegram-interaction-adapter/state.json"`
	TelegramHTTPTimeout           string `env:"CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_HTTP_TIMEOUT" envDefault:"10s"`
	CallbackHTTPTimeout           string `env:"CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_CALLBACK_HTTP_TIMEOUT" envDefault:"10s"`
}

// LoadConfig parses and validates environment configuration.
func LoadConfig() (Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse telegram interaction adapter config from environment: %w", err)
	}
	return cfg, nil
}
