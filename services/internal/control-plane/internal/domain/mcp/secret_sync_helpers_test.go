package mcp

import "testing"

func TestNormalizeSecretSyncPolicy(t *testing.T) {
	t.Run("default deterministic", func(t *testing.T) {
		policy, err := normalizeSecretSyncPolicy("")
		if err != nil {
			t.Fatalf("normalizeSecretSyncPolicy returned error: %v", err)
		}
		if policy != SecretSyncPolicyDeterministic {
			t.Fatalf("expected default policy %q, got %q", SecretSyncPolicyDeterministic, policy)
		}
	})

	t.Run("invalid policy", func(t *testing.T) {
		if _, err := normalizeSecretSyncPolicy("unknown"); err == nil {
			t.Fatal("expected invalid policy error")
		}
	})
}

func TestDeriveDeterministicSecretValueStable(t *testing.T) {
	seed := "test-seed"
	params := secretSyncDeterministicParams{
		ProjectID:            "project-1",
		Repository:           "codex-k8s/codex-k8s",
		Environment:          "production",
		GitHubSecretName:     "APP_TOKEN",
		KubernetesNamespace:  "codex-k8s-prod",
		KubernetesSecretName: "app-secrets",
		KubernetesSecretKey:  "token",
	}

	first, err := deriveDeterministicSecretValue(seed, params)
	if err != nil {
		t.Fatalf("deriveDeterministicSecretValue returned error: %v", err)
	}
	second, err := deriveDeterministicSecretValue(seed, params)
	if err != nil {
		t.Fatalf("deriveDeterministicSecretValue returned error: %v", err)
	}
	if first != second {
		t.Fatalf("expected deterministic value, got %q and %q", first, second)
	}

	changed := params
	changed.Environment = "prod"
	third, err := deriveDeterministicSecretValue(seed, changed)
	if err != nil {
		t.Fatalf("deriveDeterministicSecretValue returned error: %v", err)
	}
	if first == third {
		t.Fatal("expected different value when material changes")
	}
}

func TestDeriveSecretSyncIdempotencyKey(t *testing.T) {
	seed := "test-seed"
	params := secretSyncIdempotencyParams{
		ProjectID:            "project-1",
		Repository:           "codex-k8s/codex-k8s",
		Environment:          "production",
		GitHubSecretName:     "APP_TOKEN",
		KubernetesNamespace:  "codex-k8s-prod",
		KubernetesSecretName: "app-secrets",
		KubernetesSecretKey:  "token",
		Policy:               SecretSyncPolicyDeterministic,
		SecretValue:          "secret-value",
	}

	t.Run("explicit key", func(t *testing.T) {
		withExplicit := params
		withExplicit.ExplicitKey = "RUN-1"
		key, err := deriveSecretSyncIdempotencyKey(seed, withExplicit)
		if err != nil {
			t.Fatalf("deriveSecretSyncIdempotencyKey returned error: %v", err)
		}
		if key != "run-1" {
			t.Fatalf("expected normalized explicit key run-1, got %q", key)
		}
	})

	t.Run("deterministic key", func(t *testing.T) {
		first, err := deriveSecretSyncIdempotencyKey(seed, params)
		if err != nil {
			t.Fatalf("deriveSecretSyncIdempotencyKey returned error: %v", err)
		}
		second, err := deriveSecretSyncIdempotencyKey(seed, params)
		if err != nil {
			t.Fatalf("deriveSecretSyncIdempotencyKey returned error: %v", err)
		}
		if first != second {
			t.Fatalf("expected deterministic key, got %q and %q", first, second)
		}

		changed := params
		changed.SecretValue = "another-secret"
		third, err := deriveSecretSyncIdempotencyKey(seed, changed)
		if err != nil {
			t.Fatalf("deriveSecretSyncIdempotencyKey returned error: %v", err)
		}
		if first == third {
			t.Fatal("expected different idempotency key when secret value changes")
		}
	})
}
