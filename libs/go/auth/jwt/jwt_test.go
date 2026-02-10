package jwt

import (
	"testing"
	"time"
)

func TestJWT_IssueAndVerify(t *testing.T) {
	t.Parallel()

	key := []byte("test-secret")
	now := time.Now().UTC()

	signer, err := NewSigner("codex-k8s", key, 15*time.Minute)
	if err != nil {
		t.Fatalf("NewSigner: %v", err)
	}
	verifier, err := NewVerifier("codex-k8s", key, 0)
	if err != nil {
		t.Fatalf("NewVerifier: %v", err)
	}

	token, exp, err := signer.Issue("user-1", "owner@example.com", "ai-da-stas", true, true, now)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if exp.IsZero() {
		t.Fatalf("expected non-zero exp")
	}

	claims, err := verifier.Verify(token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("subject mismatch: %q", claims.Subject)
	}
	if claims.Email != "owner@example.com" {
		t.Fatalf("email mismatch: %q", claims.Email)
	}
	if claims.GitHubLogin != "ai-da-stas" {
		t.Fatalf("github_login mismatch: %q", claims.GitHubLogin)
	}
	if !claims.IsAdmin {
		t.Fatalf("expected is_admin=true")
	}
	if !claims.IsOwner {
		t.Fatalf("expected is_owner=true")
	}
}
