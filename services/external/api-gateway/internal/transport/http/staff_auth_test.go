package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/labstack/echo/v5"
)

func TestAuthenticatePrincipal_OAuth2ProxyHeaders_DelegatesToControlPlane(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("X-Auth-Request-Email", "user@example.com")
	req.Header.Set("X-Auth-Request-User", "ai-da-stas")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	calls := 0
	var gotEmail string
	var gotLogin string
	resolver := func(_ context.Context, email string, githubLogin string) (*controlplanev1.Principal, error) {
		calls++
		gotEmail = email
		gotLogin = githubLogin
		return &controlplanev1.Principal{
			UserId:          "00000000-0000-0000-0000-000000000001",
			Email:           email,
			GithubLogin:     githubLogin,
			IsPlatformAdmin: true,
			IsPlatformOwner: false,
		}, nil
	}

	p, err := authenticatePrincipal(c, nil, resolver)
	if err != nil {
		t.Fatalf("authenticatePrincipal failed: %v", err)
	}
	if p.GetGithubLogin() != "ai-da-stas" {
		t.Fatalf("expected principal github login %q, got %q", "ai-da-stas", p.GetGithubLogin())
	}
	if calls != 1 {
		t.Fatalf("expected resolver to be called once, got %d", calls)
	}
	if gotEmail != "user@example.com" {
		t.Fatalf("expected resolver email %q, got %q", "user@example.com", gotEmail)
	}
	if gotLogin != "ai-da-stas" {
		t.Fatalf("expected resolver login %q, got %q", "ai-da-stas", gotLogin)
	}
}
