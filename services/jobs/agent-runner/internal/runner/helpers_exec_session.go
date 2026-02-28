package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	toolGapNotFoundQuotedPattern  = regexp.MustCompile(`['"]([a-zA-Z0-9._-]+)['"]:\s+executable file not found`)
	toolGapCommandNotFoundPattern = regexp.MustCompile(`(?m)(?:^|:\s)([a-zA-Z0-9._-]+):\s+command not found$`)
	toolGapMissingCommandPattern  = regexp.MustCompile(`(?m)missing (?:required )?command[:\s]+([a-zA-Z0-9._-]+)$`)
)

type sessionFileCandidate struct {
	path    string
	modTime time.Time
}

type codexSessionIdentity struct {
	SessionID      string `json:"session_id"`
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	ThreadID       string `json:"thread_id"`
}

func readJSONFileOrNil(path string) json.RawMessage {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	if !json.Valid(bytes) {
		return nil
	}
	return json.RawMessage(bytes)
}

func latestSessionFile(sessionsDir string) string {
	files := make([]sessionFileCandidate, 0, 4)

	_ = filepath.WalkDir(sessionsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(d.Name())) != ".json" {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		files = append(files, sessionFileCandidate{path: path, modTime: info.ModTime()})
		return nil
	})
	if len(files) == 0 {
		return ""
	}
	sort.Slice(files, func(i, j int) bool { return files[i].modTime.After(files[j].modTime) })
	return files[0].path
}

func extractSessionIDFromFile(path string) string {
	bytes, err := os.ReadFile(path)
	if err != nil || !json.Valid(bytes) {
		return ""
	}

	var payload codexSessionIdentity
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return ""
	}

	for _, value := range []string{payload.SessionID, payload.ID, payload.ConversationID, payload.ThreadID} {
		stringValue := strings.TrimSpace(value)
		if stringValue != "" {
			return stringValue
		}
	}
	return ""
}

func runCommandQuiet(ctx context.Context, dir string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	if strings.TrimSpace(dir) != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

func runCommandWithInput(ctx context.Context, input []byte, stdout io.Writer, stderr io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func runCommandCaptureOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdoutBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return stdoutBuffer.Bytes(), nil
}

func runCommandCaptureCombinedOutput(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if strings.TrimSpace(dir) != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	return trimCapturedOutput(string(output), maxCapturedCommandOutput), err
}

func parseCodexReportOutput(output []byte) (codexReport, json.RawMessage, error) {
	trimmedOutput := strings.TrimSpace(string(output))
	if trimmedOutput == "" {
		return codexReport{}, nil, fmt.Errorf("empty codex output")
	}

	tryDecode := func(raw []byte) (codexReport, bool) {
		if !json.Valid(raw) {
			return codexReport{}, false
		}
		var report codexReport
		if err := json.Unmarshal(raw, &report); err != nil {
			return codexReport{}, false
		}
		return report, true
	}

	if report, ok := tryDecode([]byte(trimmedOutput)); ok {
		return report, json.RawMessage(trimmedOutput), nil
	}

	lines := strings.Split(trimmedOutput, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if report, ok := tryDecode([]byte(line)); ok {
			return report, json.RawMessage(line), nil
		}
	}

	return codexReport{}, nil, fmt.Errorf("failed to parse codex structured output")
}

func trimCapturedOutput(raw string, maxBytes int) string {
	trimmed := strings.TrimSpace(raw)
	if maxBytes <= 0 || len(trimmed) <= maxBytes {
		return trimmed
	}
	if maxBytes < len("...(truncated)") {
		return trimmed[:maxBytes]
	}
	cutoff := maxBytes - len("...(truncated)")
	return trimmed[:cutoff] + "...(truncated)"
}

func buildSessionLogJSON(result runResult, status string) json.RawMessage {
	report := result.report
	report.ActionItems = normalizeStringList(report.ActionItems)
	report.EvidenceRefs = normalizeStringList(report.EvidenceRefs)
	report.ToolGaps = normalizeStringList(result.toolGaps)
	payload := sessionLogSnapshot{
		Version: sessionLogVersionV1,
		Status:  strings.TrimSpace(status),
		Report:  report,
		Runtime: sessionRuntimeLogFields{
			TargetBranch:     strings.TrimSpace(result.targetBranch),
			CodexExecOutput:  strings.TrimSpace(result.codexExecOutput),
			GitPushOutput:    strings.TrimSpace(result.gitPushOutput),
			ExistingPRNumber: result.existingPRNumber,
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return json.RawMessage(raw)
}

func normalizeStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		lower := strings.ToLower(item)
		if _, exists := seen[lower]; exists {
			continue
		}
		seen[lower] = struct{}{}
		normalized = append(normalized, item)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func detectToolGaps(report codexReport, outputs ...string) []string {
	candidates := make([]string, 0, len(report.ToolGaps)+4)
	candidates = append(candidates, report.ToolGaps...)
	for _, output := range outputs {
		trimmed := strings.TrimSpace(output)
		if trimmed == "" {
			continue
		}
		candidates = append(candidates, extractToolGapCandidates(trimmed)...)
	}
	return normalizeStringList(candidates)
}

func extractToolGapCandidates(output string) []string {
	candidates := make([]string, 0, 4)

	for _, match := range toolGapNotFoundQuotedPattern.FindAllStringSubmatch(output, -1) {
		if len(match) >= 2 {
			candidates = append(candidates, strings.TrimSpace(match[1]))
		}
	}
	for _, match := range toolGapCommandNotFoundPattern.FindAllStringSubmatch(output, -1) {
		if len(match) >= 2 {
			candidates = append(candidates, strings.TrimSpace(match[1]))
		}
	}
	for _, match := range toolGapMissingCommandPattern.FindAllStringSubmatch(strings.ToLower(output), -1) {
		if len(match) >= 2 {
			candidates = append(candidates, strings.TrimSpace(match[1]))
		}
	}

	return candidates
}
