package staff

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
	querytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/query"
)

const (
	defaultRunRuntimeListLimit = 200
	defaultRunLogsTailLines    = 200
	maxRunLogsTailLines        = 2000
)

type runLogsSnapshot struct {
	Runtime runLogsRuntime `json:"runtime"`
}

type runLogsRuntime struct {
	CodexExecOutput string `json:"codex_exec_output"`
	GitPushOutput   string `json:"git_push_output"`
}

// ListRunJobs returns runtime jobs list with optional filters.
func (s *Service) ListRunJobs(ctx context.Context, principal Principal, filter querytypes.StaffRunListFilter) ([]entitytypes.StaffRun, error) {
	return s.listRuntimeRunsByScope(
		ctx,
		principal,
		filter,
		s.runs.ListJobsAll,
		s.runs.ListJobsForUser,
	)
}

// ListRunWaits returns wait queue list with optional filters.
func (s *Service) ListRunWaits(ctx context.Context, principal Principal, filter querytypes.StaffRunListFilter) ([]entitytypes.StaffRun, error) {
	return s.listRuntimeRunsByScope(
		ctx,
		principal,
		filter,
		s.runs.ListWaitsAll,
		s.runs.ListWaitsForUser,
	)
}

func (s *Service) listRuntimeRunsByScope(ctx context.Context, principal Principal, filter querytypes.StaffRunListFilter, listAllFn func(ctx context.Context, filter querytypes.StaffRunListFilter) ([]entitytypes.StaffRun, error), listForUserFn func(ctx context.Context, userID string, filter querytypes.StaffRunListFilter) ([]entitytypes.StaffRun, error)) ([]entitytypes.StaffRun, error) {
	normalizedFilter := normalizeRuntimeListFilter(filter)
	if principal.IsPlatformAdmin {
		return listAllFn(ctx, normalizedFilter)
	}
	return listForUserFn(ctx, principal.UserID, normalizedFilter)
}

// GetRunLogs returns run logs snapshot and tail lines.
func (s *Service) GetRunLogs(ctx context.Context, principal Principal, runID string, tailLines int) (entitytypes.StaffRunLogs, error) {
	normalizedRunID := strings.TrimSpace(runID)
	if normalizedRunID == "" {
		return entitytypes.StaffRunLogs{}, errs.Validation{Field: "run_id", Msg: "is required"}
	}
	_, _, err := s.resolveRunAccess(ctx, principal, normalizedRunID)
	if err != nil {
		return entitytypes.StaffRunLogs{}, err
	}

	item, ok, err := s.runs.GetLogsByRunID(ctx, normalizedRunID)
	if err != nil {
		return entitytypes.StaffRunLogs{}, err
	}
	if !ok {
		return entitytypes.StaffRunLogs{}, errs.Validation{Field: "run_id", Msg: "not found"}
	}

	item.TailLines = buildRunLogsTailLines(item.SnapshotJSON, normalizeTailLinesLimit(tailLines))
	return item, nil
}

func normalizeRuntimeListFilter(filter querytypes.StaffRunListFilter) querytypes.StaffRunListFilter {
	normalized := filter
	if normalized.Limit <= 0 {
		normalized.Limit = defaultRunRuntimeListLimit
	}
	normalized.TriggerKind = strings.TrimSpace(normalized.TriggerKind)
	normalized.Status = strings.TrimSpace(normalized.Status)
	normalized.AgentKey = strings.TrimSpace(normalized.AgentKey)
	normalized.WaitState = strings.TrimSpace(normalized.WaitState)
	return normalized
}

func normalizeTailLinesLimit(limit int) int {
	switch {
	case limit <= 0:
		return defaultRunLogsTailLines
	case limit > maxRunLogsTailLines:
		return maxRunLogsTailLines
	default:
		return limit
	}
}

func buildRunLogsTailLines(snapshotJSON []byte, limit int) []string {
	if len(snapshotJSON) == 0 || !json.Valid(snapshotJSON) {
		return []string{}
	}

	var snapshot runLogsSnapshot
	if err := json.Unmarshal(snapshotJSON, &snapshot); err != nil {
		return []string{}
	}

	lines := make([]string, 0, limit)
	lines = appendNonEmptyLines(lines, snapshot.Runtime.CodexExecOutput)
	lines = appendNonEmptyLines(lines, snapshot.Runtime.GitPushOutput)
	if len(lines) <= limit {
		return lines
	}
	return lines[len(lines)-limit:]
}

func appendNonEmptyLines(dst []string, raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return dst
	}
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		dst = append(dst, trimmed)
	}
	return dst
}
