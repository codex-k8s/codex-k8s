package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	jwtlib "github.com/codex-k8s/codex-k8s/libs/go/auth/jwt"
	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/errs"
	userrepo "github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/repository/user"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const (
	jwtIssuer = "codex-k8s"
)

// Config defines staff authentication settings.
type Config struct {
	PublicBaseURL         string
	GitHubOAuthClientID   string
	GitHubOAuthClientSecret string
	JWTSigningKey         []byte
	JWTTTL                time.Duration
	CookieSecure          bool
}

// Service implements GitHub OAuth login and JWT issuance.
type Service struct {
	cfg     Config
	users   userrepo.Repository
	oauth   *oauth2.Config
	signer  *jwtlib.Signer
	verifier *jwtlib.Verifier
	now     func() time.Time
}

// NewService constructs staff auth service.
func NewService(cfg Config, users userrepo.Repository) (*Service, error) {
	if cfg.PublicBaseURL == "" {
		return nil, errors.New("public base url is required")
	}
	if cfg.GitHubOAuthClientID == "" || cfg.GitHubOAuthClientSecret == "" {
		return nil, errors.New("github oauth client id/secret are required")
	}
	if len(cfg.JWTSigningKey) == 0 {
		return nil, errors.New("jwt signing key is required")
	}
	if cfg.JWTTTL <= 0 {
		return nil, errors.New("jwt ttl must be > 0")
	}

	base, err := url.Parse(cfg.PublicBaseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("invalid public base url %q", cfg.PublicBaseURL)
	}

	callbackURL := strings.TrimRight(cfg.PublicBaseURL, "/") + "/api/v1/auth/github/callback"
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.GitHubOAuthClientID,
		ClientSecret: cfg.GitHubOAuthClientSecret,
		Endpoint:     github.Endpoint,
		RedirectURL:  callbackURL,
		Scopes:       []string{"read:user", "user:email"},
	}

	signer, err := jwtlib.NewSigner(jwtIssuer, cfg.JWTSigningKey, cfg.JWTTTL)
	if err != nil {
		return nil, err
	}
	verifier, err := jwtlib.NewVerifier(jwtIssuer, cfg.JWTSigningKey, 5*time.Second)
	if err != nil {
		return nil, err
	}

	return &Service{
		cfg:      cfg,
		users:    users,
		oauth:    oauthCfg,
		signer:   signer,
		verifier: verifier,
		now:      time.Now,
	}, nil
}

// BuildLoginURL returns GitHub OAuth authorize URL and the generated state value.
func (s *Service) BuildLoginURL() (authorizeURL string, state string, err error) {
	state, err = randomState()
	if err != nil {
		return "", "", err
	}
	return s.oauth.AuthCodeURL(state, oauth2.AccessTypeOnline), state, nil
}

// ExchangeAndIssueJWT exchanges OAuth code and returns a signed JWT token for an allowed user.
func (s *Service) ExchangeAndIssueJWT(ctx context.Context, code string) (jwtToken string, expiresAt time.Time, err error) {
	if code == "" {
		return "", time.Time{}, errs.Validation{Field: "code", Msg: "is required"}
	}

	tok, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("exchange oauth code: %w", err)
	}

	ghUser, ghEmail, err := fetchGitHubIdentity(ctx, tok)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("fetch github identity: %w", err)
	}
	if ghEmail == "" {
		return "", time.Time{}, errs.Forbidden{Msg: "github account must have a verified email"}
	}

	u, ok, err := s.users.GetByEmail(ctx, ghEmail)
	if err != nil {
		return "", time.Time{}, err
	}
	if !ok {
		return "", time.Time{}, errs.Forbidden{Msg: "email is not allowed"}
	}

	if err := s.users.UpdateGitHubIdentity(ctx, u.ID, ghUser.ID, ghUser.Login); err != nil {
		return "", time.Time{}, err
	}
	u.GitHubUserID = ghUser.ID
	u.GitHubLogin = ghUser.Login

	now := s.now().UTC()
	jwtToken, expiresAt, err = s.signer.Issue(u.ID, u.Email, u.GitHubLogin, u.IsPlatformAdmin, now)
	if err != nil {
		return "", time.Time{}, err
	}

	return jwtToken, expiresAt, nil
}

// VerifyJWT validates a JWT string and returns claims.
func (s *Service) VerifyJWT(token string) (jwtlib.Claims, error) {
	return s.verifier.Verify(token)
}

func randomState() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("random: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}
