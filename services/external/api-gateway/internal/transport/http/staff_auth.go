package http

import (
	"context"
	"net/http"
	"strings"

	jwtlib "github.com/codex-k8s/codex-k8s/libs/go/auth/jwt"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	controlplanev1 "github.com/codex-k8s/codex-k8s/proto/gen/go/codexk8s/controlplane/v1"
	"github.com/labstack/echo/v5"
)

const (
	cookieAuthToken = "codexk8s_staff_jwt"
	ctxPrincipalKey = "codexk8s_principal"
)

type jwtVerifier interface {
	VerifyJWT(token string) (jwtlib.Claims, error)
}

// requireStaffAuth authenticates staff requests either via:
// - oauth2-proxy injected headers (production/dev): resolve allowlist and identity in control-plane; or
// - JWT: verify locally and attach claims as principal.
func requireStaffAuth(verifier jwtVerifier, resolver func(ctx context.Context, email string, githubLogin string) (*controlplanev1.Principal, error)) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			p, err := authenticatePrincipal(c, verifier, resolver)
			if err != nil {
				return err
			}
			c.Set(ctxPrincipalKey, p)
			return next(c)
		}
	}
}

func getPrincipal(c *echo.Context) (*controlplanev1.Principal, bool) {
	v := c.Get(ctxPrincipalKey)
	if v == nil {
		return nil, false
	}
	p, ok := v.(*controlplanev1.Principal)
	return p, ok
}

func authenticatePrincipal(c *echo.Context, verifier jwtVerifier, resolver func(ctx context.Context, email string, githubLogin string) (*controlplanev1.Principal, error)) (*controlplanev1.Principal, error) {
	req := c.Request()

	// When running behind oauth2-proxy (dev/production), accept identity from headers
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
		if resolver == nil {
			return nil, errs.Unauthorized{Msg: "staff auth misconfigured"}
		}
		p, err := resolver(req.Context(), email, login)
		if err != nil {
			return nil, err
		}
		if p == nil || strings.TrimSpace(p.UserId) == "" {
			return nil, errs.Unauthorized{Msg: "staff auth misconfigured"}
		}
		return p, nil
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
			return nil, errs.Unauthorized{Msg: "missing auth token"}
		}
		return nil, errs.Unauthorized{Msg: "missing auth token"}
	}

	claims, err := verifier.VerifyJWT(token)
	if err != nil {
		return nil, errs.Unauthorized{Msg: "invalid auth token"}
	}

	return &controlplanev1.Principal{
		UserId:          claims.Subject,
		Email:           claims.Email,
		GithubLogin:     claims.GitHubLogin,
		IsPlatformAdmin: claims.IsAdmin,
		IsPlatformOwner: claims.IsOwner,
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
