package mcp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

const secretSyncDerivationVersion = "v1"

type secretSyncDeterministicParams struct {
	ProjectID            string
	Repository           string
	Environment          string
	GitHubSecretName     string
	KubernetesNamespace  string
	KubernetesSecretName string
	KubernetesSecretKey  string
}

type secretSyncIdempotencyParams struct {
	ExplicitKey          string
	ProjectID            string
	Repository           string
	Environment          string
	GitHubSecretName     string
	KubernetesNamespace  string
	KubernetesSecretName string
	KubernetesSecretKey  string
	Policy               SecretSyncPolicy
	SecretValue          string
}

func normalizeSecretSyncPolicy(value SecretSyncPolicy) (SecretSyncPolicy, error) {
	normalized := SecretSyncPolicy(strings.ToLower(strings.TrimSpace(string(value))))
	if normalized == "" {
		return SecretSyncPolicyDeterministic, nil
	}
	switch normalized {
	case SecretSyncPolicyDeterministic, SecretSyncPolicyRandom, SecretSyncPolicyProvided:
		return normalized, nil
	default:
		return "", fmt.Errorf("policy is invalid")
	}
}

func normalizeSecretSyncRepository(value string) (string, error) {
	repository := strings.TrimSpace(value)
	if repository == "" {
		return "", nil
	}
	owner, name := splitRepoFullName(repository)
	if owner == "" || name == "" {
		return "", fmt.Errorf("repository must be in owner/name format")
	}
	return owner + "/" + name, nil
}

func deriveDeterministicSecretValue(seed string, params secretSyncDeterministicParams) (string, error) {
	key := strings.TrimSpace(seed)
	if key == "" {
		return "", fmt.Errorf("secret derivation seed is not configured")
	}
	material := strings.Join([]string{
		secretSyncDerivationVersion,
		strings.TrimSpace(params.ProjectID),
		strings.TrimSpace(params.Repository),
		normalizeEnvName(params.Environment),
		strings.TrimSpace(params.GitHubSecretName),
		strings.TrimSpace(params.KubernetesNamespace),
		strings.TrimSpace(params.KubernetesSecretName),
		normalizeKubernetesSecretDataKey(params.KubernetesSecretKey),
	}, "\n")
	return hmacSHA256Base64(key, material), nil
}

func deriveSecretSyncIdempotencyKey(seed string, params secretSyncIdempotencyParams) (string, error) {
	explicit, err := normalizeSecretSyncIdempotencyKey(params.ExplicitKey)
	if err != nil {
		return "", err
	}
	if explicit != "" {
		return explicit, nil
	}

	key := strings.TrimSpace(seed)
	if key == "" {
		return "", fmt.Errorf("idempotency derivation seed is not configured")
	}

	policy, err := normalizeSecretSyncPolicy(params.Policy)
	if err != nil {
		return "", err
	}
	secretFingerprint := "generated"
	if strings.TrimSpace(params.SecretValue) != "" {
		secretFingerprint = hmacSHA256Base64(key, "secret:"+params.SecretValue)
	}
	material := strings.Join([]string{
		secretSyncDerivationVersion,
		strings.TrimSpace(params.ProjectID),
		strings.TrimSpace(params.Repository),
		normalizeEnvName(params.Environment),
		strings.TrimSpace(params.GitHubSecretName),
		strings.TrimSpace(params.KubernetesNamespace),
		strings.TrimSpace(params.KubernetesSecretName),
		normalizeKubernetesSecretDataKey(params.KubernetesSecretKey),
		string(policy),
		secretFingerprint,
	}, "\n")
	return hmacSHA256Base64(key, material), nil
}

func normalizeSecretSyncIdempotencyKey(value string) (string, error) {
	key := strings.TrimSpace(value)
	if key == "" {
		return "", nil
	}
	if len(key) > 128 {
		return "", fmt.Errorf("idempotency_key is too long")
	}
	for _, r := range key {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '.' || r == '_' || r == '-' || r == ':' || r == '/' {
			continue
		}
		return "", fmt.Errorf("idempotency_key contains invalid characters")
	}
	return strings.ToLower(key), nil
}

func hmacSHA256Base64(key string, value string) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
