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

func readCommandOutputFileOrNil(path string) []byte {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	if len(bytes) == 0 {
		return nil
	}
	return bytes
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

func runCommandCaptureOutputWithStderr(ctx context.Context, dir string, name string, args ...string) ([]byte, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if strings.TrimSpace(dir) != "" {
		cmd.Dir = dir
	}
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuffer)
	err := cmd.Run()
	return stdoutBuffer.Bytes(), trimCapturedOutput(stderrBuffer.String(), maxCapturedCommandOutput), err
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
		report, err := decodeCodexReport(raw)
		if err != nil {
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

type rawCodexReport struct {
	Status          string          `json:"status"`
	Summary         json.RawMessage `json:"summary"`
	Branch          string          `json:"branch"`
	PRNumber        int             `json:"pr_number"`
	PRURL           string          `json:"pr_url"`
	SessionID       string          `json:"session_id"`
	Model           string          `json:"model"`
	ReasoningEffort string          `json:"reasoning_effort"`
	Diagnosis       string          `json:"diagnosis"`
	ActionItems     []string        `json:"action_items"`
	EvidenceRefs    []string        `json:"evidence_refs"`
	ToolGaps        []string        `json:"tool_gaps"`
	PullRequest     *rawCodexPR     `json:"pr"`
}

type rawCodexPR struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

func decodeCodexReport(raw []byte) (codexReport, error) {
	var payload rawCodexReport
	if err := json.Unmarshal(raw, &payload); err != nil {
		return codexReport{}, err
	}

	summary, err := decodeCodexSummary(payload.Summary)
	if err != nil {
		return codexReport{}, err
	}

	report := codexReport{
		Summary:         summary,
		Branch:          strings.TrimSpace(payload.Branch),
		PRNumber:        payload.PRNumber,
		PRURL:           strings.TrimSpace(payload.PRURL),
		SessionID:       strings.TrimSpace(payload.SessionID),
		Model:           strings.TrimSpace(payload.Model),
		ReasoningEffort: strings.TrimSpace(payload.ReasoningEffort),
		Diagnosis:       strings.TrimSpace(payload.Diagnosis),
		ActionItems:     normalizeStringList(payload.ActionItems),
		EvidenceRefs:    normalizeStringList(payload.EvidenceRefs),
		ToolGaps:        normalizeStringList(payload.ToolGaps),
	}
	if payload.PullRequest != nil {
		if report.PRNumber <= 0 {
			report.PRNumber = payload.PullRequest.Number
		}
		if report.PRURL == "" {
			report.PRURL = strings.TrimSpace(payload.PullRequest.URL)
		}
	}
	return report, nil
}

func decodeCodexSummary(raw json.RawMessage) (string, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return "", nil
	}

	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return strings.TrimSpace(single), nil
	}

	var items []string
	if err := json.Unmarshal(raw, &items); err == nil {
		normalized := normalizeStringList(items)
		if len(normalized) == 0 {
			return "", nil
		}
		return strings.Join(normalized, "\n"), nil
	}

	return "", fmt.Errorf("decode codex summary: unsupported payload %s", trimmed)
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
