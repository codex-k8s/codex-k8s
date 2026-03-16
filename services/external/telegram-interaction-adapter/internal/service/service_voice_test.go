package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/mymmrac/telego"
)

type testBotClient struct {
	sendRequests []SendMessageRequest
	downloadErr  error
}

func (b *testBotClient) Ready() bool { return true }
func (b *testBotClient) SendMessage(_ context.Context, req SendMessageRequest) (SentMessage, error) {
	b.sendRequests = append(b.sendRequests, req)
	return SentMessage{ChatID: req.ChatID, MessageID: len(b.sendRequests), SentAt: time.Now().UTC()}, nil
}
func (b *testBotClient) EditMessageKeyboard(context.Context, EditMessageKeyboardRequest) error { return nil }
func (b *testBotClient) AnswerCallbackQuery(context.Context, AnswerCallbackQueryRequest) error  { return nil }
func (b *testBotClient) DownloadFile(context.Context, string) (DownloadedFile, error) {
	return DownloadedFile{}, b.downloadErr
}
func (b *testBotClient) SetWebhook(context.Context, SetWebhookRequest) error { return nil }

type testCallbackSink struct{}

func (testCallbackSink) Submit(context.Context, CallbackEnvelope) (CallbackOutcome, error) {
	return CallbackOutcome{}, nil
}

type testAudioConverter struct{}

func (testAudioConverter) Convert(_ context.Context, payload AudioPayload) (AudioPayload, error) {
	return AudioPayload{
		Content:     payload.Content,
		ContentType: "audio/mpeg",
		FileName:    "voice.mp3",
	}, nil
}

type testSpeechToText struct{}

func (testSpeechToText) Transcribe(context.Context, AudioPayload, string) (string, error) {
	return "", errors.New("stt unavailable")
}

func TestHandleWebhook_AcksVoiceTranscriptionFailuresWithoutRetry(t *testing.T) {
	t.Parallel()

	bot := &testBotClient{downloadErr: errors.New("stt unavailable")}
	recipients, err := NewRecipientResolver("989530970", "")
	if err != nil {
		t.Fatalf("NewRecipientResolver() error = %v", err)
	}
	svc, err := New(Config{
		Recipients:     recipients,
		Bot:            bot,
		CallbackSink:   testCallbackSink{},
		AudioConverter: testAudioConverter{},
		SpeechToText:   testSpeechToText{},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	update := telego.Update{
		UpdateID: 77,
		Message: &telego.Message{
			MessageID: 55,
			Chat:      telego.Chat{ID: 989530970},
			From:      &telego.User{ID: 1, LanguageCode: "ru"},
			Voice:     &telego.Voice{FileID: "voice-file"},
		},
	}
	raw, err := json.Marshal(update)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	if err := svc.HandleWebhook(context.Background(), raw); err != nil {
		t.Fatalf("HandleWebhook() error = %v, want nil", err)
	}
	if len(bot.sendRequests) != 1 {
		t.Fatalf("sendRequests len = %d, want 1", len(bot.sendRequests))
	}
	if got, want := bot.sendRequests[0].Text, "🎙️ Не удалось распознать голосовой ответ. Попробуйте еще раз или отправьте текст."; got != want {
		t.Fatalf("notification text = %q, want %q", got, want)
	}
}
