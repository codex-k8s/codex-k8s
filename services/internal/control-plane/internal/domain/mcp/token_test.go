package mcp

import (
	"context"
	"strings"
	"testing"
	"time"

	agentrunrepo "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/repository/agentrun"
	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
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
	token := mustSignInteractionCallbackTestToken(t, service, "interaction-1")

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
	token := mustSignInteractionCallbackTestToken(t, service, "interaction-1")

	_, err := service.VerifyInteractionCallbackToken(context.Background(), token, "interaction-2")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "token subject mismatch") {
		t.Fatalf("error = %q, want token subject mismatch", err)
	}
}

func TestIssueInteractionCallbackTokenKeepsPostDeadlineGrace(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 13, 18, 0, 0, 0, time.UTC)
	deadline := now.Add(30 * time.Minute)
	service := newTokenTestService("running")
	service.now = func() time.Time { return now }

	issued, err := service.issueInteractionCallbackToken(
		entitytypes.AgentRun{
			ID:            "run-1",
			CorrelationID: "corr-1",
			ProjectID:     "project-1",
		},
		entitytypes.InteractionRequest{
			ID:                 "interaction-1",
			ResponseDeadlineAt: &deadline,
		},
		"delivery-1",
	)
	if err != nil {
		t.Fatalf("issueInteractionCallbackToken returned error: %v", err)
	}
	if issued.KeyID != "kodex/test" {
		t.Fatalf("key_id = %q, want kodex/test", issued.KeyID)
	}

	session, err := service.parseRunToken(issued.Token)
	if err != nil {
		t.Fatalf("parseRunToken returned error: %v", err)
	}
	wantExpiry := deadline.Add(interactionCallbackTokenGraceTTL)
	if !session.ExpiresAt.Equal(wantExpiry) {
		t.Fatalf("expires_at = %s, want %s", session.ExpiresAt.Format(time.RFC3339), wantExpiry.Format(time.RFC3339))
	}
}

func newTokenTestService(runStatus string) *Service {
	return &Service{
		cfg: Config{
			TokenSigningKey: "test-signing-key",
			TokenIssuer:     "kodex/test",
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

func mustSignInteractionCallbackTestToken(t *testing.T, service *Service, interactionID string) string {
	t.Helper()

	issuedAt := time.Now().UTC().Add(-1 * time.Minute)
	token, err := service.signRunToken(runTokenClaims{
		RunID:         "run-1",
		CorrelationID: "corr-1",
		ProjectID:     "project-1",
		Scope:         interactionCallbackTokenScope,
		InteractionID: strings.TrimSpace(interactionID),
		AdapterKind:   interactionCallbackTokenAdapterKind,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    service.cfg.TokenIssuer,
			Subject:   interactionCallbackTokenSubjectPrefix + strings.TrimSpace(interactionID),
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
