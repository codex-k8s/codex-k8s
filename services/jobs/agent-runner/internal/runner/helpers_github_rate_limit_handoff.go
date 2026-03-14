package runner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	sharedgithubratelimit "github.com/codex-k8s/codex-k8s/libs/go/domain/githubratelimit"
	cpclient "github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/controlplane"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

var (
	githubRateLimitHeaderIntPatterns = map[string]*regexp.Regexp{
		"x-ratelimit-limit":     regexp.MustCompile(`(?mi)(?:^|[\s"'{,])x-ratelimit-limit["']?\s*[:=]\s*"?([0-9]+)"?`),
		"x-ratelimit-remaining": regexp.MustCompile(`(?mi)(?:^|[\s"'{,])x-ratelimit-remaining["']?\s*[:=]\s*"?([0-9]+)"?`),
		"x-ratelimit-used":      regexp.MustCompile(`(?mi)(?:^|[\s"'{,])x-ratelimit-used["']?\s*[:=]\s*"?([0-9]+)"?`),
		"x-ratelimit-reset":     regexp.MustCompile(`(?mi)(?:^|[\s"'{,])x-ratelimit-reset["']?\s*[:=]\s*"?([0-9]+)"?`),
		"retry-after":           regexp.MustCompile(`(?mi)(?:^|[\s"'{,])retry-after["']?\s*[:=]\s*"?([0-9]+)"?`),
		"x-github-request-id":   regexp.MustCompile(`(?mi)(?:^|[\s"'{,])x-github-request-id["']?\s*[:=]\s*"?([a-zA-Z0-9:_-]+)"?`),
		"x-ratelimit-resource":  regexp.MustCompile(`(?mi)(?:^|[\s"'{,])x-ratelimit-resource["']?\s*[:=]\s*"?([a-zA-Z0-9._-]+)"?`),
	}
	githubDocsURLPattern    = regexp.MustCompile(`https://docs\.github\.com/[^\s"'<>]+`)
	githubStatusCodePattern = regexp.MustCompile(`(?i)(?:status(?:\s+code)?|http(?:\s+status)?|response(?:\s+status)?|code)\D{0,8}(403|429)\b`)
)

type codexExecFailure struct {
	Output []byte
	Stderr string
	Err    error
}

func (e codexExecFailure) Error() string {
	if e.Err == nil {
		return "codex exec failed"
	}
	return e.Err.Error()
}

func (e codexExecFailure) Unwrap() error {
	return e.Err
}

type githubRateLimitWaitAcceptedError struct {
	WaitID          string
	WaitState       string
	WaitReason      string
	NextStepKind    string
	ResumeNotBefore *time.Time
	PersistErr      error
}

func (e githubRateLimitWaitAcceptedError) Error() string {
	message := fmt.Sprintf("github rate-limit handoff accepted: wait_id=%s state=%s reason=%s next_step=%s", e.WaitID, e.WaitState, e.WaitReason, e.NextStepKind)
	if e.ResumeNotBefore == nil || e.ResumeNotBefore.IsZero() {
		if e.PersistErr == nil {
			return message
		}
		return fmt.Sprintf("%s waiting_snapshot_persist_err=%v", message, e.PersistErr)
	}
	message = fmt.Sprintf("%s resume_not_before=%s", message, e.ResumeNotBefore.UTC().Format(time.RFC3339))
	if e.PersistErr == nil {
		return message
	}
	return fmt.Sprintf("%s waiting_snapshot_persist_err=%v", message, e.PersistErr)
}

func (e githubRateLimitWaitAcceptedError) Unwrap() error {
	return e.PersistErr
}

type githubRateLimitSignalCandidate struct {
	SignalID           string
	ProviderStatusCode int
	RequestFingerprint string
	StderrExcerpt      string
	MessageExcerpt     string
	Headers            cpclient.GitHubRateLimitHeaders
}

func (s *Service) tryHandoffGitHubRateLimit(
	ctx context.Context,
	state codexState,
	result *runResult,
	runStartedAt time.Time,
	output []byte,
	stderr string,
	execErr error,
) error {
	if s == nil || s.cp == nil || result == nil {
		return nil
	}

	candidate, ok := detectGitHubRateLimitSignal(output, stderr, execErr)
	if !ok {
		return nil
	}

	s.refreshSessionResultFromDisk(result, state.sessionsDir)
	result.codexExecOutput = redactSensitiveOutput(
		trimCapturedOutput(formatCodexExecFailureOutput(output, stderr), maxCapturedCommandOutput),
		s.sensitiveValues(),
	)

	if _, err := s.persistSessionSnapshot(ctx, result, state, runStartedAt, runStatusRunning, nil); err != nil {
		return fmt.Errorf("persist running session snapshot before github rate-limit handoff: %w", err)
	}

	report, err := s.cp.ReportGitHubRateLimitSignal(ctx, cpclient.ReportGitHubRateLimitSignalParams{
		RunID:                  s.cfg.RunID,
		SignalID:               candidate.SignalID,
		CorrelationID:          s.cfg.CorrelationID,
		ContourKind:            githubRateLimitContourAgentBotToken,
		SignalOrigin:           "agent_runner",
		OperationClass:         githubRateLimitOperationAgentGitHubCall,
		ProviderStatusCode:     candidate.ProviderStatusCode,
		OccurredAt:             time.Now().UTC(),
		RequestFingerprint:     candidate.RequestFingerprint,
		StderrExcerpt:          candidate.StderrExcerpt,
		MessageExcerpt:         candidate.MessageExcerpt,
		Headers:                candidate.Headers,
		SessionSnapshotVersion: int64Ptr(result.snapshotVersion),
	})
	if err != nil {
		if grpcstatus.Code(err) == codes.FailedPrecondition {
			s.logger.Info("github rate-limit handoff rejected by control-plane; keeping original codex failure", "run_id", s.cfg.RunID, "err", err)
			return codexExecFailure{Output: output, Stderr: stderr, Err: execErr}
		}
		return fmt.Errorf("report github rate-limit signal: %w", err)
	}
	if strings.TrimSpace(report.RunnerAction) != sharedgithubratelimit.RunnerActionPersistSessionAndExitWait {
		return fmt.Errorf("unsupported github rate-limit runner action %q", strings.TrimSpace(report.RunnerAction))
	}

	s.refreshSessionResultFromDisk(result, state.sessionsDir)
	result.codexExecOutput = redactSensitiveOutput(
		trimCapturedOutput(formatCodexExecFailureOutput(output, stderr), maxCapturedCommandOutput),
		s.sensitiveValues(),
	)
	waitAcceptedErr := githubRateLimitWaitAcceptedError{
		WaitID:          strings.TrimSpace(report.WaitID),
		WaitState:       strings.TrimSpace(report.WaitState),
		WaitReason:      strings.TrimSpace(report.WaitReason),
		NextStepKind:    strings.TrimSpace(report.NextStepKind),
		ResumeNotBefore: report.ResumeNotBefore,
	}
	if _, err := s.persistSessionSnapshot(ctx, result, state, runStartedAt, runStatusWaitingBackpressure, nil); err != nil {
		s.logger.Warn(
			"github rate-limit handoff accepted but waiting snapshot persist failed; keeping accepted wait path to avoid terminal failed overwrite",
			"run_id", s.cfg.RunID,
			"wait_id", waitAcceptedErr.WaitID,
			"wait_state", waitAcceptedErr.WaitState,
			"err", err,
		)
		waitAcceptedErr.PersistErr = fmt.Errorf("persist waiting_backpressure session snapshot after github rate-limit handoff: %w", err)
		return waitAcceptedErr
	}

	return waitAcceptedErr
}

func detectGitHubRateLimitSignal(output []byte, stderr string, execErr error) (githubRateLimitSignalCandidate, bool) {
	combined := strings.Join(nonEmptyStrings(
		string(output),
		stderr,
		errorString(execErr),
	), "\n")
	lowerCombined := strings.ToLower(strings.TrimSpace(combined))
	if lowerCombined == "" {
		return githubRateLimitSignalCandidate{}, false
	}

	statusCode, hasStatusCode := extractGitHubStatusCode(combined)
	if !looksLikeGitHubRateLimit(lowerCombined, hasStatusCode) {
		return githubRateLimitSignalCandidate{}, false
	}
	if !hasStatusCode {
		statusCode = 403
	}

	headers := cpclient.GitHubRateLimitHeaders{
		RateLimitLimit:     extractGitHubHeaderInt(combined, "x-ratelimit-limit"),
		RateLimitRemaining: extractGitHubHeaderInt(combined, "x-ratelimit-remaining"),
		RateLimitUsed:      extractGitHubHeaderInt(combined, "x-ratelimit-used"),
		RateLimitResource:  extractGitHubHeaderString(combined, "x-ratelimit-resource"),
		RetryAfterSeconds:  extractGitHubHeaderInt(combined, "retry-after"),
		GitHubRequestID:    extractGitHubHeaderString(combined, "x-github-request-id"),
		DocumentationURL:   extractGitHubDocumentationURL(combined),
	}
	if resetAt := extractGitHubRateLimitResetAt(combined); resetAt != nil {
		headers.RateLimitResetAt = resetAt
	}

	stderrExcerpt := trimCapturedOutput(strings.TrimSpace(stderr), sharedgithubratelimit.SignalExcerptMaxBytes)
	messageExcerpt := trimCapturedOutput(selectGitHubRateLimitExcerpt(output, stderr, execErr), sharedgithubratelimit.SignalExcerptMaxBytes)
	requestFingerprint := buildGitHubRateLimitRequestFingerprint(headers, statusCode, messageExcerpt)
	signalID := buildGitHubRateLimitSignalID(headers, statusCode, stderrExcerpt, messageExcerpt)

	return githubRateLimitSignalCandidate{
		SignalID:           signalID,
		ProviderStatusCode: statusCode,
		RequestFingerprint: requestFingerprint,
		StderrExcerpt:      stderrExcerpt,
		MessageExcerpt:     messageExcerpt,
		Headers:            headers,
	}, true
}

func looksLikeGitHubRateLimit(lowerCombined string, hasStatusCode bool) bool {
	if strings.Contains(lowerCombined, "api rate limit exceeded") ||
		strings.Contains(lowerCombined, "secondary rate limit") ||
		strings.Contains(lowerCombined, "secondary-rate-limits") ||
		strings.Contains(lowerCombined, "abuse detection") {
		return true
	}
	if strings.Contains(lowerCombined, "rate limit") &&
		(hasStatusCode ||
			strings.Contains(lowerCombined, "retry-after") ||
			strings.Contains(lowerCombined, "x-ratelimit-") ||
			strings.Contains(lowerCombined, "docs.github.com") ||
			strings.Contains(lowerCombined, "x-github-request-id")) {
		return true
	}
	return false
}

func selectGitHubRateLimitExcerpt(output []byte, stderr string, execErr error) string {
	for _, source := range []string{strings.TrimSpace(stderr), strings.TrimSpace(string(output)), strings.TrimSpace(errorString(execErr))} {
		if source == "" {
			continue
		}
		for _, line := range strings.Split(source, "\n") {
			trimmedLine := strings.TrimSpace(line)
			lowerLine := strings.ToLower(trimmedLine)
			if trimmedLine == "" {
				continue
			}
			if strings.Contains(lowerLine, "rate limit") || strings.Contains(lowerLine, "retry-after") || strings.Contains(lowerLine, "x-ratelimit-") {
				return trimmedLine
			}
		}
		if source != "" {
			return source
		}
	}
	return ""
}

func extractGitHubStatusCode(input string) (int, bool) {
	match := githubStatusCodePattern.FindStringSubmatch(input)
	if len(match) < 2 {
		return 0, false
	}
	value, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, false
	}
	if value != 403 && value != 429 {
		return 0, false
	}
	return value, true
}

func extractGitHubHeaderInt(input string, headerName string) *int {
	pattern, ok := githubRateLimitHeaderIntPatterns[headerName]
	if !ok {
		return nil
	}
	match := pattern.FindStringSubmatch(input)
	if len(match) < 2 {
		return nil
	}
	value, err := strconv.Atoi(match[1])
	if err != nil {
		return nil
	}
	return &value
}

func extractGitHubHeaderString(input string, headerName string) string {
	pattern, ok := githubRateLimitHeaderIntPatterns[headerName]
	if !ok {
		return ""
	}
	match := pattern.FindStringSubmatch(input)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func extractGitHubRateLimitResetAt(input string) *time.Time {
	epochSeconds := extractGitHubHeaderInt(input, "x-ratelimit-reset")
	if epochSeconds == nil || *epochSeconds <= 0 {
		return nil
	}
	resetAt := time.Unix(int64(*epochSeconds), 0).UTC()
	return &resetAt
}

func extractGitHubDocumentationURL(input string) string {
	match := githubDocsURLPattern.FindString(input)
	if strings.TrimSpace(match) == "" {
		return ""
	}
	return strings.TrimSpace(match)
}

func buildGitHubRateLimitRequestFingerprint(headers cpclient.GitHubRateLimitHeaders, statusCode int, messageExcerpt string) string {
	if strings.TrimSpace(headers.GitHubRequestID) != "" {
		return strings.TrimSpace(headers.GitHubRequestID)
	}
	if strings.TrimSpace(messageExcerpt) == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(strings.Join([]string{
		strconv.Itoa(statusCode),
		strings.TrimSpace(headers.RateLimitResource),
		strings.TrimSpace(messageExcerpt),
	}, "|")))
	return "ghrlfp-" + hex.EncodeToString(sum[:8])
}

func buildGitHubRateLimitSignalID(headers cpclient.GitHubRateLimitHeaders, statusCode int, stderrExcerpt string, messageExcerpt string) string {
	parts := []string{
		strconv.Itoa(statusCode),
		strings.TrimSpace(headers.GitHubRequestID),
		strings.TrimSpace(headers.DocumentationURL),
		strings.TrimSpace(headers.RateLimitResource),
		strings.TrimSpace(messageExcerpt),
		strings.TrimSpace(stderrExcerpt),
	}
	if headers.RetryAfterSeconds != nil {
		parts = append(parts, strconv.Itoa(*headers.RetryAfterSeconds))
	}
	if headers.RateLimitResetAt != nil && !headers.RateLimitResetAt.IsZero() {
		parts = append(parts, strconv.FormatInt(headers.RateLimitResetAt.UTC().Unix(), 10))
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return "ghrlsig-" + hex.EncodeToString(sum[:12])
}

func formatCodexExecFailureOutput(output []byte, stderr string) string {
	parts := make([]string, 0, 2)
	if trimmedOutput := strings.TrimSpace(string(output)); trimmedOutput != "" {
		parts = append(parts, "stdout:\n"+trimmedOutput)
	}
	if trimmedStderr := strings.TrimSpace(stderr); trimmedStderr != "" {
		parts = append(parts, "stderr:\n"+trimmedStderr)
	}
	return strings.Join(parts, "\n\n")
}

func (s *Service) refreshSessionResultFromDisk(result *runResult, sessionsDir string) {
	if result == nil {
		return
	}
	result.sessionFilePath = latestSessionFile(sessionsDir)
	if strings.TrimSpace(result.sessionFilePath) != "" {
		if sessionID := extractSessionIDFromFile(result.sessionFilePath); strings.TrimSpace(sessionID) != "" {
			result.sessionID = sessionID
		}
	}
}

func nonEmptyStrings(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func int64Ptr(value int64) *int64 {
	if value <= 0 {
		return nil
	}
	result := value
	return &result
}

func isGitHubRateLimitWaitAccepted(err error) bool {
	var waitErr githubRateLimitWaitAcceptedError
	return errors.As(err, &waitErr)
}
