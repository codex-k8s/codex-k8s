package worker

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	webhookdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/webhook"
)

const (
	defaultRunNamespacePrefix = "codex-issue"
	runNamespaceFallback      = "codex-run"
)

var nonDNSLabel = regexp.MustCompile(`[^a-z0-9-]`)

// runRuntimePayload keeps only fields that influence worker runtime decisions.
type runRuntimePayload struct {
	Trigger *runRuntimeTrigger `json:"trigger"`
	Issue   *runRuntimeIssue   `json:"issue"`
}

// runRuntimeTrigger captures normalized trigger kind from webhook payload.
type runRuntimeTrigger struct {
	Kind webhookdomain.TriggerKind `json:"kind"`
}

// runRuntimeIssue captures optional issue metadata used in namespace naming.
type runRuntimeIssue struct {
	Number int64 `json:"number"`
}

// runExecutionContext contains resolved execution mode and namespace metadata for one run.
type runExecutionContext struct {
	RuntimeMode agentdomain.RuntimeMode
	Namespace   string
	IssueNumber int64
}

// resolveRunExecutionContext derives runtime mode and namespace strategy from run payload.
func resolveRunExecutionContext(runID string, projectID string, runPayload json.RawMessage, namespacePrefix string) runExecutionContext {
	meta := parseRunRuntimePayload(runPayload)
	mode := resolveRuntimeMode(meta)
	context := runExecutionContext{
		RuntimeMode: mode,
		IssueNumber: resolveIssueNumber(meta),
	}

	if mode == agentdomain.RuntimeModeFullEnv {
		context.Namespace = buildRunNamespace(namespacePrefix, projectID, runID, context.IssueNumber)
	}
	return context
}

// parseRunRuntimePayload parses only fields required for runtime routing and tolerates malformed payloads.
func parseRunRuntimePayload(raw json.RawMessage) runRuntimePayload {
	if len(raw) == 0 {
		return runRuntimePayload{}
	}
	var payload runRuntimePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return runRuntimePayload{}
	}
	return payload
}

// resolveRuntimeMode maps trigger kind to execution profile with code-only fallback.
func resolveRuntimeMode(payload runRuntimePayload) agentdomain.RuntimeMode {
	if payload.Trigger == nil {
		return agentdomain.RuntimeModeCodeOnly
	}
	switch payload.Trigger.Kind {
	case webhookdomain.TriggerKindDev, webhookdomain.TriggerKindDevRevise:
		return agentdomain.RuntimeModeFullEnv
	default:
		return agentdomain.RuntimeModeCodeOnly
	}
}

// resolveIssueNumber returns positive issue number or zero when not provided.
func resolveIssueNumber(payload runRuntimePayload) int64 {
	if payload.Issue == nil {
		return 0
	}
	if payload.Issue.Number <= 0 {
		return 0
	}
	return payload.Issue.Number
}

// buildRunNamespace composes deterministic DNS-safe namespace name for full-env runs.
func buildRunNamespace(prefix string, projectID string, runID string, issueNumber int64) string {
	basePrefix := sanitizeDNSLabelValue(prefix)
	if basePrefix == "" {
		basePrefix = defaultRunNamespacePrefix
	}

	projectPart := compactIdentifier(projectID, 12)
	if projectPart == "" {
		projectPart = "project"
	}

	runPart := compactIdentifier(runID, 12)
	if runPart == "" {
		runPart = "run"
	}

	var candidate string
	if issueNumber > 0 {
		candidate = fmt.Sprintf(
			"%s-%s-i%s-r%s",
			basePrefix,
			projectPart,
			strconv.FormatInt(issueNumber, 10),
			runPart,
		)
	} else {
		candidate = fmt.Sprintf("%s-run-%s", basePrefix, runPart)
	}

	candidate = sanitizeDNSLabelValue(candidate)
	if candidate == "" {
		return runNamespaceFallback
	}
	if len(candidate) <= 63 {
		return candidate
	}
	candidate = strings.TrimRight(candidate[:63], "-")
	if candidate == "" {
		return runNamespaceFallback
	}
	return candidate
}

// compactIdentifier strips non-essential separators and truncates identifier to fixed length.
func compactIdentifier(value string, max int) string {
	if max <= 0 {
		return ""
	}
	clean := strings.ToLower(strings.TrimSpace(value))
	if clean == "" {
		return ""
	}
	clean = strings.ReplaceAll(clean, "_", "")
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, ".", "")
	clean = nonDNSLabel.ReplaceAllString(clean, "")
	if len(clean) > max {
		return clean[:max]
	}
	return clean
}

// sanitizeDNSLabelValue converts arbitrary text into Kubernetes DNS label format.
func sanitizeDNSLabelValue(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return ""
	}
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, ".", "-")
	normalized = nonDNSLabel.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}
	return normalized
}
