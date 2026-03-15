package service

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	ffmpegSampleRate = "16000"
	ffmpegChannels   = "1"
	ffmpegFormat     = "mp3"
)

// AudioPayload stores one in-memory audio file prepared for STT.
type AudioPayload struct {
	Content     []byte
	ContentType string
	FileName    string
}

// AudioConverter normalizes Telegram voice payloads into an STT-compatible format.
type AudioConverter interface {
	Convert(context.Context, AudioPayload) (AudioPayload, error)
}

// FFmpegAudioConverter converts Telegram voice payloads with ffmpeg when needed.
type FFmpegAudioConverter struct{}

// Convert normalizes audio to MP3/mono/16k when the input is not already compatible.
func (c FFmpegAudioConverter) Convert(ctx context.Context, payload AudioPayload) (AudioPayload, error) {
	if len(payload.Content) == 0 {
		return AudioPayload{}, fmt.Errorf("empty audio content")
	}
	if isOpenAICompatibleAudio(payload.ContentType, payload.FileName) {
		return payload, nil
	}

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-nostdin",
		"-y",
		"-i", "pipe:0",
		"-ac", ffmpegChannels,
		"-ar", ffmpegSampleRate,
		"-f", ffmpegFormat,
		"pipe:1",
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdin = bytes.NewReader(payload.Content)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText != "" {
			return AudioPayload{}, fmt.Errorf("ffmpeg failed: %w: %s", err, errText)
		}
		return AudioPayload{}, fmt.Errorf("ffmpeg failed: %w", err)
	}

	output := stdout.Bytes()
	if len(output) == 0 {
		return AudioPayload{}, fmt.Errorf("empty transcoded audio")
	}

	return AudioPayload{
		Content:     output,
		ContentType: "audio/mpeg",
		FileName:    normalizeAudioFilename(payload.FileName),
	}, nil
}

func normalizeAudioFilename(fileName string) string {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return "voice.mp3"
	}
	lower := strings.ToLower(fileName)
	if strings.HasSuffix(lower, ".mp3") {
		return fileName
	}
	if ext := filepath.Ext(fileName); ext != "" {
		return strings.TrimSuffix(fileName, ext) + ".mp3"
	}
	return fileName + ".mp3"
}

func isOpenAICompatibleAudio(contentType string, fileName string) bool {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "audio/mpeg", "audio/mp3", "audio/mp4", "audio/mp4a-latm", "audio/x-m4a", "audio/m4a", "audio/wav", "audio/x-wav", "audio/webm":
		return true
	}

	lowerName := strings.ToLower(strings.TrimSpace(fileName))
	return strings.HasSuffix(lowerName, ".mp3") ||
		strings.HasSuffix(lowerName, ".mpeg") ||
		strings.HasSuffix(lowerName, ".mp4") ||
		strings.HasSuffix(lowerName, ".m4a") ||
		strings.HasSuffix(lowerName, ".wav") ||
		strings.HasSuffix(lowerName, ".webm")
}
