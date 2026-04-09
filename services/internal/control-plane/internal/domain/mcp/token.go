package mcp

import (
	"fmt"
	"strings"
	"time"

	entitytypes "github.com/codex-k8s/kodex/services/internal/control-plane/internal/domain/types/entity"
	"github.com/golang-jwt/jwt/v5"
)

const (
	interactionCallbackTokenGraceTTL      = minTokenTTL
	interactionCallbackTokenSubjectPrefix = "mcp-interaction-callback:"
	interactionCallbackTokenScope         = "interaction_callback"
	interactionCallbackTokenAdapterKind   = "telegram"
)

type runTokenClaims struct {
	RunID         string `json:"run_id"`
	CorrelationID string `json:"correlation_id"`
	ProjectID     string `json:"project_id,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	RuntimeMode   string `json:"runtime_mode"`
	Scope         string `json:"scope,omitempty"`
	InteractionID string `json:"interaction_id,omitempty"`
	DeliveryID    string `json:"delivery_id,omitempty"`
	AdapterKind   string `json:"adapter_kind,omitempty"`
	jwt.RegisteredClaims
}

type interactionCallbackToken struct {
	Token     string
	KeyID     string
	ExpiresAt time.Time
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
		Scope:            strings.TrimSpace(payload.Scope),
		InteractionID:    strings.TrimSpace(payload.InteractionID),
		DeliveryID:       strings.TrimSpace(payload.DeliveryID),
		AdapterKind:      strings.TrimSpace(payload.AdapterKind),
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
	}, jwt.WithIssuer(s.cfg.TokenIssuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}), jwt.WithTimeFunc(s.now))
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
		TokenSubject:  strings.TrimSpace(claims.Subject),
		ExpiresAt:     expiresAt,
	}, nil
}

func (s *Service) issueInteractionCallbackToken(run entitytypes.AgentRun, interaction entitytypes.InteractionRequest, deliveryID string) (interactionCallbackToken, error) {
	now := s.now().UTC()
	expiresAt := now.Add(interactionCallbackTokenGraceTTL)
	if interaction.ResponseDeadlineAt != nil {
		deadline := interaction.ResponseDeadlineAt.UTC().Add(interactionCallbackTokenGraceTTL)
		if deadline.After(expiresAt) {
			expiresAt = deadline
		}
	}

	token, err := s.signRunToken(runTokenClaims{
		RunID:         strings.TrimSpace(run.ID),
		CorrelationID: strings.TrimSpace(run.CorrelationID),
		ProjectID:     strings.TrimSpace(run.ProjectID),
		Scope:         interactionCallbackTokenScope,
		InteractionID: strings.TrimSpace(interaction.ID),
		DeliveryID:    strings.TrimSpace(deliveryID),
		AdapterKind:   interactionCallbackTokenAdapterKind,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   interactionCallbackTokenSubjectPrefix + strings.TrimSpace(interaction.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})
	if err != nil {
		return interactionCallbackToken{}, err
	}
	return interactionCallbackToken{
		Token:     token,
		KeyID:     s.interactionCallbackTokenKeyID(),
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) interactionCallbackTokenKeyID() string {
	keyID := strings.TrimSpace(s.cfg.TokenIssuer)
	if keyID == "" {
		return defaultTokenIssuer
	}
	return keyID
}
