package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	userrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/user"
	"github.com/labstack/echo/v5"
)

func TestAuthenticatePrincipal_OAuth2ProxyHeaders_PersistGitHubLogin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("X-Auth-Request-Email", "user@example.com")
	req.Header.Set("X-Auth-Request-User", "ai-da-stas")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	users := &stubUserRepo{
		u: userrepo.User{
			ID:    "00000000-0000-0000-0000-000000000001",
			Email: "user@example.com",
		},
	}

	p, err := authenticatePrincipal(c, nil, users)
	if err != nil {
		t.Fatalf("authenticatePrincipal failed: %v", err)
	}
	if p.GitHubLogin != "ai-da-stas" {
		t.Fatalf("expected principal github login to be persisted, got %q", p.GitHubLogin)
	}
	if users.updateCalls != 1 {
		t.Fatalf("expected UpdateGitHubIdentity to be called once, got %d", users.updateCalls)
	}
	if users.updatedLogin != "ai-da-stas" {
		t.Fatalf("expected updated login to be %q, got %q", "ai-da-stas", users.updatedLogin)
	}
}

type stubUserRepo struct {
	u           userrepo.User
	updateCalls int
	updatedLogin string
}

func (s *stubUserRepo) EnsureOwner(_ context.Context, _ string) (userrepo.User, error) {
	return userrepo.User{}, nil
}

func (s *stubUserRepo) GetByID(_ context.Context, userID string) (userrepo.User, bool, error) {
	if userID == s.u.ID {
		return s.u, true, nil
	}
	return userrepo.User{}, false, nil
}

func (s *stubUserRepo) GetByEmail(_ context.Context, email string) (userrepo.User, bool, error) {
	if email == s.u.Email {
		return s.u, true, nil
	}
	return userrepo.User{}, false, nil
}

func (s *stubUserRepo) GetByGitHubLogin(_ context.Context, _ string) (userrepo.User, bool, error) {
	return userrepo.User{}, false, nil
}

func (s *stubUserRepo) UpdateGitHubIdentity(_ context.Context, userID string, githubUserID int64, githubLogin string) error {
	s.updateCalls++
	s.updatedLogin = githubLogin
	s.u.ID = userID
	s.u.GitHubUserID = githubUserID
	s.u.GitHubLogin = githubLogin
	return nil
}

func (s *stubUserRepo) CreateAllowedUser(_ context.Context, _ string, _ bool) (userrepo.User, error) {
	return userrepo.User{}, nil
}

func (s *stubUserRepo) List(_ context.Context, _ int) ([]userrepo.User, error) {
	return nil, nil
}

func (s *stubUserRepo) DeleteByID(_ context.Context, _ string) error {
	return nil
}
