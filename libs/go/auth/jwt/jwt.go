package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// Claims defines codex-k8s staff JWT claims.
type Claims struct {
	Email       string `json:"email"`
	GitHubLogin string `json:"github_login,omitempty"`
	IsAdmin     bool   `json:"is_admin"`

	jwtv5.RegisteredClaims
}

// Signer issues signed JWT strings.
type Signer struct {
	issuer string
	key    []byte
	ttl    time.Duration
}

// NewSigner creates a signer for HS256 tokens.
func NewSigner(issuer string, key []byte, ttl time.Duration) (*Signer, error) {
	if issuer == "" {
		return nil, errors.New("issuer is required")
	}
	if len(key) == 0 {
		return nil, errors.New("signing key is required")
	}
	if ttl <= 0 {
		return nil, errors.New("ttl must be > 0")
	}
	return &Signer{issuer: issuer, key: key, ttl: ttl}, nil
}

// Issue creates a signed token for a subject.
func (s *Signer) Issue(subject string, email string, githubLogin string, isAdmin bool, now time.Time) (token string, expiresAt time.Time, err error) {
	if subject == "" {
		return "", time.Time{}, errors.New("subject is required")
	}
	if email == "" {
		return "", time.Time{}, errors.New("email is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}

	expiresAt = now.Add(s.ttl).UTC()
	claims := Claims{
		Email:       email,
		GitHubLogin: githubLogin,
		IsAdmin:     isAdmin,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   subject,
			IssuedAt:  jwtv5.NewNumericDate(now.UTC()),
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
		},
	}

	j := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	signed, err := j.SignedString(s.key)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

// Verifier validates and parses JWT strings.
type Verifier struct {
	issuer string
	key    []byte
	leeway time.Duration
}

// NewVerifier creates a verifier for HS256 tokens.
func NewVerifier(issuer string, key []byte, leeway time.Duration) (*Verifier, error) {
	if issuer == "" {
		return nil, errors.New("issuer is required")
	}
	if len(key) == 0 {
		return nil, errors.New("signing key is required")
	}
	if leeway < 0 {
		return nil, errors.New("leeway must be >= 0")
	}
	return &Verifier{issuer: issuer, key: key, leeway: leeway}, nil
}

// Verify parses and validates a token string and returns claims.
func (v *Verifier) Verify(tokenString string) (Claims, error) {
	if tokenString == "" {
		return Claims{}, errors.New("token is required")
	}

	claims := &Claims{}
	_, err := jwtv5.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwtv5.Token) (interface{}, error) {
			return v.key, nil
		},
		jwtv5.WithIssuer(v.issuer),
		jwtv5.WithValidMethods([]string{jwtv5.SigningMethodHS256.Alg()}),
		jwtv5.WithLeeway(v.leeway),
		jwtv5.WithExpirationRequired(),
	)
	if err != nil {
		return Claims{}, fmt.Errorf("parse/verify token: %w", err)
	}

	if claims.Subject == "" {
		return Claims{}, errors.New("token subject is required")
	}
	if claims.Email == "" {
		return Claims{}, errors.New("token email is required")
	}

	return *claims, nil
}
