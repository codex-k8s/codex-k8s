package mcp

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	"github.com/codex-k8s/codex-k8s/libs/go/postgres"
	entitytypes "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/types/entity"
)

type approvalTargetRef struct {
	ProjectID            string `json:"project_id,omitempty"`
	Repository           string `json:"repository,omitempty"`
	Environment          string `json:"environment,omitempty"`
	GitHubSecretName     string `json:"github_secret_name,omitempty"`
	KubernetesNamespace  string `json:"kubernetes_namespace,omitempty"`
	KubernetesSecretName string `json:"kubernetes_secret_name,omitempty"`
	KubernetesSecretKey  string `json:"kubernetes_secret_key,omitempty"`
	Policy               string `json:"policy,omitempty"`
	IdempotencyKey       string `json:"idempotency_key,omitempty"`
	DatabaseName         string `json:"database_name,omitempty"`
}

type secretSyncPayload struct {
	ProjectID            string           `json:"project_id,omitempty"`
	Repository           string           `json:"repository,omitempty"`
	Environment          string           `json:"environment"`
	GitHubSecretName     string           `json:"github_secret_name"`
	KubernetesNamespace  string           `json:"kubernetes_namespace"`
	KubernetesSecretName string           `json:"kubernetes_secret_name"`
	KubernetesSecretKey  string           `json:"kubernetes_secret_key"`
	Policy               SecretSyncPolicy `json:"policy,omitempty"`
	IdempotencyKey       string           `json:"idempotency_key,omitempty"`
	SecretValueEncrypted string           `json:"secret_value_encrypted"`
}

type databaseLifecyclePayload struct {
	ProjectID     string                  `json:"project_id"`
	Environment   string                  `json:"environment"`
	Action        DatabaseLifecycleAction `json:"action"`
	DatabaseName  string                  `json:"database_name"`
	ConfirmDelete bool                    `json:"confirm_delete,omitempty"`
}

type ownerFeedbackPayload struct {
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	AllowCustom bool     `json:"allow_custom"`
}

type approvalDecisionPayload struct {
	Decision  string `json:"decision"`
	ActorID   string `json:"actor_id,omitempty"`
	Reason    string `json:"reason,omitempty"`
	DecidedAt string `json:"decided_at,omitempty"`
	Error     string `json:"error,omitempty"`
}

type approvalAppliedPayload struct {
	AppliedAt string `json:"applied_at"`
	AppliedBy string `json:"applied_by,omitempty"`
}

type runWaitPayload struct {
	RunID                string `json:"run_id"`
	WaitState            string `json:"wait_state,omitempty"`
	TimeoutGuardDisabled bool   `json:"timeout_guard_disabled"`
}

var defaultDatabaseLifecycleAllowedEnvs = []string{"dev", "staging", "prod"}

func resolveControlApprovalMode(tool ToolName, runCtx resolvedRunContext) entitytypes.MCPApprovalMode {
	triggerLabel := ""
	if runCtx.Payload.Trigger != nil {
		triggerLabel = strings.ToLower(strings.TrimSpace(runCtx.Payload.Trigger.Label))
	}
	agentKey := ""
	if runCtx.Payload.Agent != nil {
		agentKey = strings.ToLower(strings.TrimSpace(runCtx.Payload.Agent.Key))
	}

	runtimeMode := normalizeRuntimeMode(runCtx.Session.RuntimeMode)
	if runtimeMode != agentdomain.RuntimeModeFullEnv {
		return entitytypes.MCPApprovalModeOwner
	}

	switch tool {
	case ToolMCPSecretSyncEnv:
		if triggerLabel == triggerLabelRunDevRevise || triggerLabel == triggerLabelRunSelfPatch {
			return entitytypes.MCPApprovalModeDelegated
		}
		if triggerLabel == triggerLabelRunOps && agentKey == agentKeySRE {
			return entitytypes.MCPApprovalModeDelegated
		}
		return entitytypes.MCPApprovalModeOwner
	case ToolMCPDatabaseLifecycle:
		if triggerLabel == triggerLabelRunOps && (agentKey == agentKeySRE || agentKey == agentKeyDev) {
			return entitytypes.MCPApprovalModeDelegated
		}
		return entitytypes.MCPApprovalModeOwner
	case ToolMCPOwnerFeedbackRequest:
		return entitytypes.MCPApprovalModeOwner
	default:
		return entitytypes.MCPApprovalModeOwner
	}
}

func normalizeEnvName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeDatabaseLifecycleAllowedEnvs(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		env := normalizeEnvName(value)
		if env == "" {
			continue
		}
		out[env] = struct{}{}
	}
	if len(out) > 0 {
		return out
	}
	out = make(map[string]struct{}, len(defaultDatabaseLifecycleAllowedEnvs))
	for _, env := range defaultDatabaseLifecycleAllowedEnvs {
		out[env] = struct{}{}
	}
	return out
}

func isDatabaseLifecycleEnvironmentAllowed(allowed map[string]struct{}, env string) bool {
	if len(allowed) == 0 {
		return true
	}
	_, ok := allowed[normalizeEnvName(env)]
	return ok
}

func listDatabaseLifecycleAllowedEnvs(allowed map[string]struct{}) []string {
	if len(allowed) == 0 {
		return nil
	}
	out := make([]string, 0, len(allowed))
	for env := range allowed {
		out = append(out, env)
	}
	slices.Sort(out)
	return out
}

func normalizeDatabaseLifecycleName(value string) (string, error) {
	return postgres.NormalizeDatabaseName(value)
}

func normalizeKubernetesSecretDataKey(value string) string {
	key := strings.TrimSpace(value)
	if key == "" {
		return controlActionSecretDefaultKey
	}
	return key
}

func normalizeSecretTargetNamespace(session SessionContext, explicitNamespace string) string {
	namespace := strings.TrimSpace(explicitNamespace)
	if namespace != "" {
		return namespace
	}
	return strings.TrimSpace(session.Namespace)
}

func requestActorID(runCtx resolvedRunContext) string {
	if runCtx.Payload.Agent != nil {
		key := strings.TrimSpace(runCtx.Payload.Agent.Key)
		if key != "" {
			return "agent:" + key
		}
	}
	return "agent:unknown"
}

func normalizeOptions(values []string) []string {
	return normalizeDistinctStrings(values)
}

func normalizeDistinctStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func marshalRawJSON(value any) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return raw
}

func encryptedValueBase64(encrypted []byte) string {
	if len(encrypted) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(encrypted)
}

func decodeEncryptedValueBase64(encoded string) ([]byte, error) {
	value := strings.TrimSpace(encoded)
	if value == "" {
		return nil, fmt.Errorf("secret encrypted payload is empty")
	}
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode encrypted secret payload: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("secret encrypted payload is empty")
	}
	return data, nil
}

func newGeneratedSecretValue() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate secret value: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func nowRFC3339Nano(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
