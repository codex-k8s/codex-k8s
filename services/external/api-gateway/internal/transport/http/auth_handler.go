package http

import (
	"context"
	"net/http"
	"time"

	jwtlib "github.com/codex-k8s/codex-k8s/libs/go/auth/jwt"
	"github.com/codex-k8s/codex-k8s/libs/go/errs"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/transport/http/casters"
	"github.com/labstack/echo/v5"
)

const (
	cookieOAuthState = "codexk8s_oauth_state"
)

type authService interface {
	BuildLoginURL() (authorizeURL string, state string, err error)
	ExchangeAndIssueJWT(ctx context.Context, code string) (jwtToken string, expiresAt time.Time, err error)
	VerifyJWT(token string) (jwtlib.Claims, error)
}

// authHandler implements OAuth login/callback and basic session endpoints.
type authHandler struct {
	svc          authService
	cookieSecure bool
}

func newAuthHandler(svc authService, cookieSecure bool) *authHandler {
	return &authHandler{
		svc:          svc,
		cookieSecure: cookieSecure,
	}
}

func (h *authHandler) LoginGitHub(c *echo.Context) error {
	authorizeURL, state, err := h.svc.BuildLoginURL()
	if err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:     cookieOAuthState,
		Value:    state,
		Path:     "/api/v1/auth/github/callback",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300,
	})

	return c.Redirect(http.StatusFound, authorizeURL)
}

func (h *authHandler) CallbackGitHub(c *echo.Context) error {
	qState := c.QueryParam("state")
	qCode := c.QueryParam("code")
	if qState == "" {
		return errs.Validation{Field: "state", Msg: "is required"}
	}
	if qCode == "" {
		return errs.Validation{Field: "code", Msg: "is required"}
	}

	stateCookie, err := c.Cookie(cookieOAuthState)
	if err != nil || stateCookie == nil || stateCookie.Value == "" {
		return errs.Unauthorized{Msg: "missing oauth state"}
	}
	if stateCookie.Value != qState {
		return errs.Unauthorized{Msg: "oauth state mismatch"}
	}

	// Clear state cookie on callback.
	c.SetCookie(&http.Cookie{
		Name:     cookieOAuthState,
		Value:    "",
		Path:     "/api/v1/auth/github/callback",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	jwtToken, expiresAt, err := h.svc.ExchangeAndIssueJWT(c.Request().Context(), qCode)
	if err != nil {
		return err
	}

	ttl := int(time.Until(expiresAt).Seconds())
	if ttl < 1 {
		ttl = 1
	}
	c.SetCookie(&http.Cookie{
		Name:     cookieAuthToken,
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   ttl,
	})

	return c.Redirect(http.StatusFound, "/")
}

func (h *authHandler) Logout(c *echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:     cookieAuthToken,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	// When running behind oauth2-proxy (production/dev), the browser session is primarily backed
	// by oauth2-proxy cookies. If we don't clear them, users may appear "still logged in".
	//
	// oauth2-proxy defaults:
	// - main session cookie: `_oauth2_proxy`
	// - csrf cookie: `_oauth2_proxy_csrf`
	for _, name := range []string{"_oauth2_proxy", "_oauth2_proxy_csrf"} {
		c.SetCookie(&http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   h.cookieSecure,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *authHandler) Me(c *echo.Context) error {
	p, ok := getPrincipal(c)
	if !ok {
		return errs.Unauthorized{Msg: "not authenticated"}
	}
	return c.JSON(http.StatusOK, casters.Me(p))
}
