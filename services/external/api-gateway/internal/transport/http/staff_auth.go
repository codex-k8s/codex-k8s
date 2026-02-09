package http

import (
	"strings"

	jwtlib "github.com/codex-k8s/codex-k8s/libs/go/auth/jwt"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/errs"
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

func requireStaffAuth(verifier jwtVerifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			p, err := authenticatePrincipal(c, verifier)
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

func authenticatePrincipal(c *echo.Context, verifier jwtVerifier) (staff.Principal, error) {
	req := c.Request()

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

