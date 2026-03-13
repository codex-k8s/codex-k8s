package mcp

import (
	"context"
	"strings"
	"testing"
	"time"

	agentrunrepo "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/repository/agentrun"
	"github.com/golang-jwt/jwt/v5"
)

func TestVerifyRunTokenRejectsInactiveRun(t *testing.T) {
	t.Parallel()

	service := newTokenTestService("completed")
	token := mustSignTokenTestRunToken(t, service, "run:run-1")

	_, err := service.VerifyRunToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), `run status "completed" is not active`) {
		t.Fatalf("error = %q, want inactive-run message", err)
	}
}

func TestVerifyInteractionCallbackTokenAllowsInactiveRun(t *testing.T) {
	t.Parallel()

	service := newTokenTestService("completed")
	token := mustSignTokenTestRunToken(t, service, interactionCallbackTokenSubjectPrefix+"interaction-1")

	session, err := service.VerifyInteractionCallbackToken(context.Background(), token, "interaction-1")
	if err != nil {
		t.Fatalf("VerifyInteractionCallbackToken returned error: %v", err)
	}
	if session.RunID != "run-1" {
		t.Fatalf("run_id = %q, want run-1", session.RunID)
	}
	if session.TokenSubject != interactionCallbackTokenSubjectPrefix+"interaction-1" {
		t.Fatalf("token_subject = %q, want interaction callback subject", session.TokenSubject)
	}
}

func TestVerifyInteractionCallbackTokenRejectsSubjectMismatch(t *testing.T) {
	t.Parallel()

	service := newTokenTestService("completed")
	token := mustSignTokenTestRunToken(t, service, interactionCallbackTokenSubjectPrefix+"interaction-1")

	_, err := service.VerifyInteractionCallbackToken(context.Background(), token, "interaction-2")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "token subject mismatch") {
		t.Fatalf("error = %q, want token subject mismatch", err)
	}
}

func newTokenTestService(runStatus string) *Service {
	return &Service{
		cfg: Config{
			TokenSigningKey: "test-signing-key",
			TokenIssuer:     "codex-k8s/test",
		},
		runs: &interactionTestRunsRepository{
			byID: map[string]agentrunrepo.Run{
				"run-1": {
					ID:            "run-1",
					CorrelationID: "corr-1",
					ProjectID:     "project-1",
					Status:        runStatus,
				},
			},
		},
	}
}

func mustSignTokenTestRunToken(t *testing.T, service *Service, subject string) string {
	t.Helper()

	issuedAt := time.Now().UTC().Add(-1 * time.Minute)
	token, err := service.signRunToken(runTokenClaims{
		RunID:         "run-1",
		CorrelationID: "corr-1",
		ProjectID:     "project-1",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    service.cfg.TokenIssuer,
			Subject:   strings.TrimSpace(subject),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(5 * time.Minute)),
		},
	})
	if err != nil {
		t.Fatalf("signRunToken returned error: %v", err)
	}
	return token
}
