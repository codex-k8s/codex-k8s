package http

import (
	"fmt"
	"net/http"
	"strings"

	jwtlib "github.com/codex-k8s/codex-k8s/libs/go/auth/jwt"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/errs"
	userrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/user"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/staff"
	"github.com/labstack/echo/v5"
)

const (
	cookieAuthToken = "codexk8s_staff_jwt"
	ctxPrincipalKey = "codexk8s_principal"
)

type jwtVerifier interface {
	VerifyJWT(token string) (jwtlib.Claims, error)
}

func requireStaffAuth(verifier jwtVerifier, users userrepo.Repository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			p, err := authenticatePrincipal(c, verifier, users)
			if err != nil {
				return err
			}
			c.Set(ctxPrincipalKey, p)
			return next(c)
		}
	}
}

func getPrincipal(c *echo.Context) (staff.Principal, bool) {
	v := c.Get(ctxPrincipalKey)
	if v == nil {
		return staff.Principal{}, false
	}
	p, ok := v.(staff.Principal)
	return p, ok
}

func authenticatePrincipal(c *echo.Context, verifier jwtVerifier, users userrepo.Repository) (staff.Principal, error) {
	req := c.Request()

	// When running behind oauth2-proxy (dev/staging), accept identity from headers
	// and resolve platform access via the allowlist stored in the DB.
	// This keeps "registration disabled" semantics even if oauth2-proxy allows any GitHub user to authenticate.
	email := firstNonEmpty(
		req.Header.Get("X-Auth-Request-Email"),
		req.Header.Get("X-Forwarded-Email"),
	)
	login := firstNonEmpty(
		req.Header.Get("X-Auth-Request-User"),
		req.Header.Get("X-Forwarded-User"),
	)
	if email != "" {
		if users == nil {
			return staff.Principal{}, errs.Unauthorized{Msg: "staff auth misconfigured"}
		}
		u, ok, err := users.GetByEmail(req.Context(), email)
		if err != nil {
			return staff.Principal{}, fmt.Errorf("resolve staff user by email: %w", err)
		}
		if !ok {
			return staff.Principal{}, errs.Forbidden{Msg: "email is not allowed"}
		}
		if u.GitHubLogin == "" {
			u.GitHubLogin = login
		}
		return staff.Principal{
			UserID:          u.ID,
			Email:           u.Email,
			GitHubLogin:     u.GitHubLogin,
			IsPlatformAdmin: u.IsPlatformAdmin,
		}, nil
	}

	token := ""
	if authz := req.Header.Get("Authorization"); strings.HasPrefix(authz, "Bearer ") {
		token = strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
	}
	if token == "" {
		if ck, err := c.Cookie(cookieAuthToken); err == nil && ck != nil {
			token = ck.Value
		}
	}
	if token == "" {
		// For GET endpoints, surface unauthorized rather than method-level errors.
		if req.Method == http.MethodGet || req.Method == http.MethodHead {
			return staff.Principal{}, errs.Unauthorized{Msg: "missing auth token"}
		}
		return staff.Principal{}, errs.Unauthorized{Msg: "missing auth token"}
	}

	claims, err := verifier.VerifyJWT(token)
	if err != nil {
		return staff.Principal{}, errs.Unauthorized{Msg: "invalid auth token"}
	}

	return staff.Principal{
		UserID:          claims.Subject,
		Email:           claims.Email,
		GitHubLogin:     claims.GitHubLogin,
		IsPlatformAdmin: claims.IsAdmin,
	}, nil
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}
