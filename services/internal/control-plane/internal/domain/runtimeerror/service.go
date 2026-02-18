package runtimeerror

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	runtimeerrorrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/runtimeerror"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

const (
	defaultSource     = "unknown"
	defaultLevel      = "error"
	defaultMessage    = "runtime error"
	maxSourceLength   = 160
	maxMessageLength  = 4000
	maxStackLength    = 16000
	maxContextLength  = 255
	maxIDStringLength = 64
)

// Service records runtime errors into persistent journal.
type Service struct {
	repo   runtimeerrorrepo.Repository
	logger *slog.Logger
}

// NewService constructs runtime error recorder service.
func NewService(repo runtimeerrorrepo.Repository, logger *slog.Logger) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("runtime error repository is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{repo: repo, logger: logger}, nil
}

// Record persists one runtime error with normalized payload.
func (s *Service) Record(ctx context.Context, params querytypes.RuntimeErrorRecordParams) (entitytypes.RuntimeError, error) {
	prepared := normalizeRecordParams(params)
	item, err := s.repo.Insert(ctx, prepared)
	if err != nil {
		return entitytypes.RuntimeError{}, fmt.Errorf("insert runtime error: %w", err)
	}
	return item, nil
}

// RecordBestEffort persists runtime error and only logs failures.
func (s *Service) RecordBestEffort(ctx context.Context, params querytypes.RuntimeErrorRecordParams) {
	if s == nil {
		return
	}
	if _, err := s.Record(ctx, params); err != nil {
		s.logger.Warn("record runtime error failed", "source", strings.TrimSpace(params.Source), "err", err)
	}
}

func normalizeRecordParams(params querytypes.RuntimeErrorRecordParams) querytypes.RuntimeErrorRecordParams {
	normalized := params
	normalized.Source = normalizeLimitedText(params.Source, defaultSource, maxSourceLength)
	normalized.Level = normalizeLevel(params.Level)
	normalized.Message = normalizeLimitedText(params.Message, defaultMessage, maxMessageLength)
	normalized.StackTrace = normalizeLimitedText(params.StackTrace, "", maxStackLength)
	normalized.CorrelationID = normalizeLimitedText(params.CorrelationID, "", maxContextLength)
	normalized.Namespace = normalizeLimitedText(params.Namespace, "", maxContextLength)
	normalized.JobName = normalizeLimitedText(params.JobName, "", maxContextLength)
	normalized.RunID = normalizeLimitedText(params.RunID, "", maxIDStringLength)
	normalized.ProjectID = normalizeLimitedText(params.ProjectID, "", maxIDStringLength)
	normalized.DetailsJSON = normalizeDetailsJSON(params.DetailsJSON)
	return normalized
}

func normalizeLevel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "critical", "fatal":
		return "critical"
	case "warning", "warn":
		return "warning"
	case "error", "err":
		return "error"
	default:
		return defaultLevel
	}
}

func normalizeLimitedText(raw string, fallback string, maxLength int) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = fallback
	}
	if maxLength > 0 && len(value) > maxLength {
		value = value[:maxLength]
	}
	return value
}

func normalizeDetailsJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return json.RawMessage(`{}`)
	}
	if json.Valid(raw) {
		return json.RawMessage(trimmed)
	}
	safe, _ := json.Marshal(map[string]string{"raw": trimmed})
	return json.RawMessage(safe)
}
