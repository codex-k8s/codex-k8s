package mcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type runTokenClaims struct {
	RunID         string `json:"run_id"`
	CorrelationID string `json:"correlation_id"`
	ProjectID     string `json:"project_id,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	RuntimeMode   string `json:"runtime_mode"`
	jwt.RegisteredClaims
}

func (s *Service) signRunToken(payload runTokenClaims) (string, error) {
	claimsRegistered := payload.RegisteredClaims
	if claimsRegistered.Issuer == "" {
		claimsRegistered.Issuer = s.cfg.TokenIssuer
	}
	if claimsRegistered.Subject == "" {
		claimsRegistered.Subject = "run:" + strings.TrimSpace(payload.RunID)
	}

	claims := runTokenClaims{
		RunID:            strings.TrimSpace(payload.RunID),
		CorrelationID:    strings.TrimSpace(payload.CorrelationID),
		ProjectID:        strings.TrimSpace(payload.ProjectID),
		Namespace:        strings.TrimSpace(payload.Namespace),
		RuntimeMode:      string(parseRuntimeMode(payload.RuntimeMode)),
		RegisteredClaims: claimsRegistered,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.TokenSigningKey))
	if err != nil {
		return "", fmt.Errorf("sign jwt token: %w", err)
	}
	return signed, nil
}

func (s *Service) parseRunToken(rawToken string) (SessionContext, error) {
	if strings.TrimSpace(rawToken) == "" {
		return SessionContext{}, fmt.Errorf("token is required")
	}

	parsed, err := jwt.ParseWithClaims(rawToken, &runTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected token signing method")
		}
		return []byte(s.cfg.TokenSigningKey), nil
	}, jwt.WithIssuer(s.cfg.TokenIssuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		return SessionContext{}, fmt.Errorf("parse token: %w", err)
	}
	claims, ok := parsed.Claims.(*runTokenClaims)
	if !ok {
		return SessionContext{}, fmt.Errorf("unexpected token claims")
	}
	if !parsed.Valid {
		return SessionContext{}, fmt.Errorf("token is invalid")
	}

	runID := strings.TrimSpace(claims.RunID)
	if runID == "" {
		return SessionContext{}, fmt.Errorf("token missing run_id")
	}
	correlationID := strings.TrimSpace(claims.CorrelationID)
	if correlationID == "" {
		return SessionContext{}, fmt.Errorf("token missing correlation_id")
	}

	expiresAt := time.Time{}
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.UTC()
	}
	if expiresAt.IsZero() {
		return SessionContext{}, fmt.Errorf("token missing expiration")
	}

	return SessionContext{
		RunID:         runID,
		CorrelationID: correlationID,
		ProjectID:     strings.TrimSpace(claims.ProjectID),
		Namespace:     strings.TrimSpace(claims.Namespace),
		RuntimeMode:   parseRuntimeMode(claims.RuntimeMode),
		ExpiresAt:     expiresAt,
	}, nil
}
