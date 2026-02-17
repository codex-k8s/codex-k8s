package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	floweventdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/flowevent"
	cpclient "github.com/codex-k8s/codex-k8s/services/jobs/agent-runner/internal/controlplane"
)

type codexAuthMode string

const (
	codexAuthModeUnknown codexAuthMode = ""
	codexAuthModeChatGPT codexAuthMode = "chatgpt"
	codexAuthModeAPIKey  codexAuthMode = "apikey"
)

const (
	codexAuthCheckModel           = "gpt-5.3-codex"
	codexAuthCheckReasoningEffort = "low"
	codexAuthCheckPrompt          = "пинг: ответь одним словом 'понг'. не используй инструменты."

	codexAuthFileName = "auth.json"
)

var (
	ansiColorRegexp          = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	deviceAuthURLRegexp      = regexp.MustCompile(`https?://[^\s]+`)
	deviceAuthUserCodeRegexp = regexp.MustCompile(`[A-Z0-9]{4,}-[A-Z0-9]{4,}`)
)

type codexAuthStatus struct {
	LoggedIn bool
	Mode     codexAuthMode
	Raw      string
}

type codexDeviceAuthPayload struct {
	VerificationURL string `json:"verification_url"`
	UserCode        string `json:"user_code"`
}

type codexAuthSynchronizedPayload struct {
	Source string `json:"source"`
}

func (s *Service) ensureCodexReady(ctx context.Context, state codexState) error {
	desiredMode := s.desiredCodexAuthMode()

	if desiredMode == codexAuthModeChatGPT {
		// Ensure ChatGPT device auth is valid by doing a minimal call to gpt-5.3.
		if err := s.writeCodexConfig(state.codexDir, codexAuthCheckModel, codexAuthCheckReasoningEffort); err != nil {
			return err
		}
		if err := s.ensureCodexAuthenticated(ctx, state); err != nil {
			return err
		}
	}

	// Restore actual run model config.
	if err := s.writeCodexConfig(state.codexDir, s.cfg.AgentModel, s.cfg.AgentReasoningEffort); err != nil {
		return err
	}

	// Best-effort: persist refreshed tokens after auth flow before executing the main prompt.
	if err := s.syncCodexAuthToControlPlane(ctx, state, "ready"); err != nil {
		s.logger.Warn("sync codex auth failed", "err", err)
	}
	return nil
}

func (s *Service) ensureCodexAuthenticated(ctx context.Context, state codexState) error {
	desiredMode := s.desiredCodexAuthMode()

	switch desiredMode {
	case codexAuthModeAPIKey:
		apiKey := strings.TrimSpace(s.cfg.OpenAIAPIKey)
		if apiKey == "" {
			return fmt.Errorf("CODEXK8S_OPENAI_API_KEY is required for Codex API-key auth")
		}
		_ = runCommandQuiet(ctx, "", "codex", "logout")
		if err := runCommandWithInput(ctx, []byte(apiKey), io.Discard, io.Discard, "codex", "login", "--with-api-key"); err != nil {
			return fmt.Errorf("codex login --with-api-key failed: %w", err)
		}
		return nil
	case codexAuthModeChatGPT:
		break
	default:
		return fmt.Errorf("unsupported codex auth mode: %q", desiredMode)
	}

	// Try to hydrate auth.json from control-plane to avoid repeated device auth per run.
	if err := s.restoreCodexAuthFromControlPlane(ctx, state); err != nil {
		s.logger.Warn("restore codex auth skipped", "err", err)
	}

	status, err := readCodexAuthStatus(ctx)
	if err != nil {
		return err
	}

	if status.LoggedIn && (status.Mode == codexAuthModeChatGPT || status.Mode == codexAuthModeUnknown) {
		if err := s.codexAuthPing(ctx, state); err == nil {
			return nil
		}
		// Logged in, but tokens may be stale. Force re-auth.
		s.logger.Warn("codex auth ping failed; forcing device auth", "status", status.Raw)
	}

	// Ensure we start from a clean state.
	_ = runCommandQuiet(ctx, "", "codex", "logout")

	_ = s.cp.UpsertRunStatusComment(ctx, cpclient.UpsertRunStatusCommentParams{
		RunID: s.cfg.RunID,
		Phase: "auth_required",
	})

	deviceAuth := newDeviceAuthCapture()
	stdoutWriter := newLineTeeWriter(os.Stdout, deviceAuth.observeLine)
	stderrWriter := newLineTeeWriter(os.Stderr, deviceAuth.observeLine)

	deviceAuth.setOnReady(func(url string, code string) {
		_ = s.emitEvent(ctx, floweventdomain.EventTypeRunCodexAuthRequired, codexDeviceAuthPayload{
			VerificationURL: url,
			UserCode:        code,
		})
	})

	cmd := exec.CommandContext(ctx, "codex", "login", "--device-auth")
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("codex login --device-auth failed: %w", err)
	}

	// Verify auth with a minimal request to gpt-5.3.
	if err := s.codexAuthPing(ctx, state); err != nil {
		return err
	}

	if err := s.syncCodexAuthToControlPlane(ctx, state, "device_auth"); err != nil {
		return err
	}

	// Restore status comment back to started to avoid leaving "auth required" marker in GitHub thread.
	_ = s.cp.UpsertRunStatusComment(ctx, cpclient.UpsertRunStatusCommentParams{
		RunID: s.cfg.RunID,
		Phase: "started",
	})

	return nil
}

func (s *Service) codexAuthPing(ctx context.Context, state codexState) error {
	pingDir := filepath.Join(state.workspaceDir, "codex-auth-check")
	if err := os.MkdirAll(pingDir, 0o755); err != nil {
		return fmt.Errorf("create codex auth check dir: %w", err)
	}

	checkCtx, cancel := context.WithTimeout(ctx, 75*time.Second)
	defer cancel()

	out, err := runCommandCaptureCombinedOutput(checkCtx, "", "codex", "exec", "--cd", pingDir, codexAuthCheckPrompt)
	if err != nil {
		return fmt.Errorf("codex auth ping failed: %w: %s", err, out)
	}
	return nil
}

func (s *Service) restoreCodexAuthFromControlPlane(ctx context.Context, state codexState) error {
	authJSON, found, err := s.cp.GetCodexAuth(ctx)
	if err != nil {
		return err
	}
	if !found || len(authJSON) == 0 {
		return nil
	}

	authPath := filepath.Join(state.codexDir, codexAuthFileName)
	if err := os.MkdirAll(filepath.Dir(authPath), 0o755); err != nil {
		return fmt.Errorf("create codex auth dir: %w", err)
	}
	if err := os.WriteFile(authPath, authJSON, 0o600); err != nil {
		return fmt.Errorf("write codex auth file: %w", err)
	}
	return nil
}

func (s *Service) syncCodexAuthToControlPlane(ctx context.Context, state codexState, source string) error {
	authPath := filepath.Join(state.codexDir, codexAuthFileName)
	raw, err := os.ReadFile(authPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read codex auth file: %w", err)
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil
	}
	if err := s.cp.UpsertCodexAuth(ctx, trimmed); err != nil {
		return err
	}
	_ = s.emitEvent(ctx, floweventdomain.EventTypeRunCodexAuthSynchronized, codexAuthSynchronizedPayload{
		Source: strings.TrimSpace(source),
	})
	return nil
}

func readCodexAuthStatus(ctx context.Context) (codexAuthStatus, error) {
	out, err := runCommandCaptureCombinedOutput(ctx, "", "codex", "login", "status")
	trimmed := strings.TrimSpace(stripANSI(out))
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// "codex login status" typically exits non-zero when not logged in.
			return codexAuthStatus{LoggedIn: false, Mode: parseCodexAuthMode(trimmed), Raw: trimmed}, nil
		}
		return codexAuthStatus{}, fmt.Errorf("codex login status failed: %w", err)
	}

	if trimmed == "" {
		return codexAuthStatus{LoggedIn: false, Mode: codexAuthModeUnknown, Raw: ""}, nil
	}
	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "not logged in") || strings.Contains(lower, "logged out") {
		return codexAuthStatus{LoggedIn: false, Mode: parseCodexAuthMode(trimmed), Raw: trimmed}, nil
	}
	return codexAuthStatus{LoggedIn: true, Mode: parseCodexAuthMode(trimmed), Raw: trimmed}, nil
}

func parseCodexAuthMode(raw string) codexAuthMode {
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "chatgpt"):
		return codexAuthModeChatGPT
	case strings.Contains(lower, "api key") || strings.Contains(lower, "apikey"):
		return codexAuthModeAPIKey
	default:
		return codexAuthModeUnknown
	}
}

func (s *Service) desiredCodexAuthMode() codexAuthMode {
	// Subscription-backed models typically require ChatGPT auth.
	if isGPT53Model(s.cfg.AgentModel) {
		return codexAuthModeChatGPT
	}
	if strings.TrimSpace(s.cfg.OpenAIAPIKey) != "" {
		return codexAuthModeAPIKey
	}
	return codexAuthModeChatGPT
}

func isGPT53Model(model string) bool {
	normalized := strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(normalized, "gpt-5.3-")
}

func stripANSI(raw string) string {
	return ansiColorRegexp.ReplaceAllString(raw, "")
}

type lineTeeWriter struct {
	dst    io.Writer
	onLine func(line string)

	mu  sync.Mutex
	buf []byte
}

func newLineTeeWriter(dst io.Writer, onLine func(line string)) *lineTeeWriter {
	if dst == nil {
		dst = io.Discard
	}
	if onLine == nil {
		onLine = func(string) {}
	}
	return &lineTeeWriter{
		dst:    dst,
		onLine: onLine,
		buf:    make([]byte, 0, 1024),
	}
}

func (w *lineTeeWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	n, err := w.dst.Write(p)
	w.buf = append(w.buf, p...)

	lines := make([]string, 0, 4)
	for {
		idx := bytes.IndexByte(w.buf, '\n')
		if idx < 0 {
			break
		}
		line := string(w.buf[:idx])
		w.buf = w.buf[idx+1:]
		lines = append(lines, line)
	}
	w.mu.Unlock()

	for _, line := range lines {
		w.onLine(line)
	}

	return n, err
}

type deviceAuthCapture struct {
	mu       sync.Mutex
	url      string
	code     string
	onReady  func(url string, code string)
	emitted  bool
}

func newDeviceAuthCapture() *deviceAuthCapture {
	return &deviceAuthCapture{}
}

func (c *deviceAuthCapture) setOnReady(fn func(url string, code string)) {
	c.mu.Lock()
	c.onReady = fn
	c.mu.Unlock()
}

func (c *deviceAuthCapture) observeLine(line string) {
	stripped := strings.TrimSpace(stripANSI(line))
	if stripped == "" {
		return
	}

	var toEmit func(url string, code string)
	var url string
	var code string

	c.mu.Lock()
	if c.url == "" {
		if match := deviceAuthURLRegexp.FindString(stripped); match != "" {
			c.url = strings.TrimSpace(match)
		}
	}
	if c.code == "" {
		if match := deviceAuthUserCodeRegexp.FindString(stripped); match != "" {
			c.code = strings.TrimSpace(match)
		}
	}
	if !c.emitted && c.url != "" && c.code != "" && c.onReady != nil {
		c.emitted = true
		toEmit = c.onReady
		url = c.url
		code = c.code
	}
	c.mu.Unlock()

	if toEmit != nil {
		toEmit(url, code)
	}
}
