package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mymmrac/telego"
)

// BotClient abstracts Telegram Bot API operations needed by the adapter service.
type BotClient interface {
	Ready() bool
	SendMessage(context.Context, SendMessageRequest) (SentMessage, error)
	EditMessageKeyboard(context.Context, EditMessageKeyboardRequest) error
	AnswerCallbackQuery(context.Context, AnswerCallbackQueryRequest) error
	SetWebhook(context.Context, SetWebhookRequest) error
}

// TelegramBotClientConfig configures Bot API transport.
type TelegramBotClientConfig struct {
	Token   string
	Timeout time.Duration
	Logger  *slog.Logger
}

// SendMessageRequest holds one Telegram message send request.
type SendMessageRequest struct {
	ChatID        int64
	Text          string
	ActionLabel   string
	ActionURL     string
	InlineOptions []InlineOption
}

// InlineOption describes one inline keyboard callback button.
type InlineOption struct {
	Label        string
	CallbackData string
}

// SentMessage stores minimal message identifiers used by the adapter.
type SentMessage struct {
	ChatID    int64
	MessageID int
	SentAt    time.Time
}

// EditMessageKeyboardRequest removes or replaces inline keyboard for one message.
type EditMessageKeyboardRequest struct {
	ChatID    int64
	MessageID int
}

// AnswerCallbackQueryRequest acknowledges Telegram callback query.
type AnswerCallbackQueryRequest struct {
	QueryID string
	Text    string
}

// SetWebhookRequest updates Telegram webhook configuration.
type SetWebhookRequest struct {
	URL         string
	SecretToken string
}

type telegramBotClient struct {
	bot    *telego.Bot
	logger *slog.Logger
}

// NewTelegramBotClient builds the default telego-backed Bot API client.
func NewTelegramBotClient(cfg TelegramBotClientConfig) (BotClient, error) {
	token := strings.TrimSpace(cfg.Token)
	if token == "" {
		return &telegramBotClient{logger: cfg.Logger}, nil
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	bot, err := telego.NewBot(token, telego.WithHTTPClient(&http.Client{Timeout: timeout}))
	if err != nil {
		return nil, fmt.Errorf("create telego bot: %w", err)
	}
	return &telegramBotClient{
		bot:    bot,
		logger: cfg.Logger,
	}, nil
}

func (c *telegramBotClient) Ready() bool {
	return c.bot != nil
}

func (c *telegramBotClient) SendMessage(ctx context.Context, req SendMessageRequest) (SentMessage, error) {
	if c.bot == nil {
		return SentMessage{}, fmt.Errorf("telegram bot token is not configured")
	}

	params := &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: req.ChatID},
		Text:   req.Text,
	}
	if len(req.InlineOptions) > 0 {
		rows := make([][]telego.InlineKeyboardButton, 0, len(req.InlineOptions))
		for _, option := range req.InlineOptions {
			rows = append(rows, []telego.InlineKeyboardButton{{
				Text:         option.Label,
				CallbackData: option.CallbackData,
			}})
		}
		params.ReplyMarkup = &telego.InlineKeyboardMarkup{InlineKeyboard: rows}
	} else if strings.TrimSpace(req.ActionURL) != "" && strings.TrimSpace(req.ActionLabel) != "" {
		params.ReplyMarkup = &telego.InlineKeyboardMarkup{
			InlineKeyboard: [][]telego.InlineKeyboardButton{{
				{
					Text: req.ActionLabel,
					URL:  req.ActionURL,
				},
			}},
		}
	}

	message, err := c.bot.SendMessage(ctx, params)
	if err != nil {
		return SentMessage{}, err
	}
	return SentMessage{
		ChatID:    req.ChatID,
		MessageID: message.MessageID,
		SentAt:    time.Now().UTC(),
	}, nil
}

func (c *telegramBotClient) EditMessageKeyboard(ctx context.Context, req EditMessageKeyboardRequest) error {
	if c.bot == nil {
		return fmt.Errorf("telegram bot token is not configured")
	}
	_, err := c.bot.EditMessageReplyMarkup(ctx, &telego.EditMessageReplyMarkupParams{
		ChatID:    telego.ChatID{ID: req.ChatID},
		MessageID: req.MessageID,
		ReplyMarkup: &telego.InlineKeyboardMarkup{
			InlineKeyboard: [][]telego.InlineKeyboardButton{},
		},
	})
	return err
}

func (c *telegramBotClient) AnswerCallbackQuery(ctx context.Context, req AnswerCallbackQueryRequest) error {
	if c.bot == nil {
		return fmt.Errorf("telegram bot token is not configured")
	}
	return c.bot.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
		CallbackQueryID: req.QueryID,
		Text:            req.Text,
	})
}

func (c *telegramBotClient) SetWebhook(ctx context.Context, req SetWebhookRequest) error {
	if c.bot == nil {
		return nil
	}
	return c.bot.SetWebhook(ctx, &telego.SetWebhookParams{
		URL:         req.URL,
		SecretToken: req.SecretToken,
		AllowedUpdates: []string{
			"callback_query",
			"message",
		},
	})
}

func normalizeTelegramProviderMessageRef(chatID int64, messageID int, sentAt time.Time) *ProviderMessageRef {
	sentAtValue := sentAt.UTC()
	return &ProviderMessageRef{
		ChatRef:   strconv.FormatInt(chatID, 10),
		MessageID: strconv.Itoa(messageID),
		SentAt:    &sentAtValue,
	}
}
