package service

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
)

// SpeechToText transcribes audio into plain text.
type SpeechToText interface {
	Transcribe(context.Context, AudioPayload, string) (string, error)
}

// OpenAISpeechToText uses OpenAI Audio Transcriptions API.
type OpenAISpeechToText struct {
	client  openai.Client
	model   string
	timeout time.Duration
	logger  *slog.Logger
}

// NewOpenAISpeechToText builds the default OpenAI-backed STT implementation.
func NewOpenAISpeechToText(apiKey string, model string, timeout time.Duration, logger *slog.Logger) *OpenAISpeechToText {
	if logger == nil {
		logger = slog.Default()
	}
	return &OpenAISpeechToText{
		client:  openai.NewClient(option.WithAPIKey(strings.TrimSpace(apiKey))),
		model:   strings.TrimSpace(model),
		timeout: timeout,
		logger:  logger,
	}
}

// Transcribe converts audio into text using the configured OpenAI model.
func (s *OpenAISpeechToText) Transcribe(ctx context.Context, payload AudioPayload, language string) (string, error) {
	if len(payload.Content) == 0 {
		return "", fmt.Errorf("empty audio content")
	}

	timeout := s.timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	params := openai.AudioTranscriptionNewParams{
		File:  openai.File(bytes.NewReader(payload.Content), payload.FileName, payload.ContentType),
		Model: openai.AudioModel(strings.TrimSpace(s.model)),
	}
	if language = strings.TrimSpace(language); language != "" {
		params.Language = param.NewOpt(language)
	}

	result, err := s.client.Audio.Transcriptions.New(requestCtx, params)
	if err != nil {
		s.logger.Error("openai speech-to-text failed", "err", err)
		return "", err
	}
	if result == nil || strings.TrimSpace(result.Text) == "" {
		return "", fmt.Errorf("empty speech-to-text response")
	}
	return strings.TrimSpace(result.Text), nil
}
