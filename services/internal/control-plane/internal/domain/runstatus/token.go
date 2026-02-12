package runstatus

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

func (s *Service) signDeleteToken(payload deleteTokenPayload) (string, error) {
	if strings.TrimSpace(payload.RunID) == "" || strings.TrimSpace(payload.Namespace) == "" {
		return "", errDeleteTokenInvalid
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal delete token payload: %w", err)
	}
	payloadPart := base64.RawURLEncoding.EncodeToString(rawPayload)
	mac := hmac.New(sha256.New, []byte(s.cfg.TokenSigningKey))
	_, _ = mac.Write([]byte(payloadPart))
	signaturePart := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payloadPart + "." + signaturePart, nil
}

func (s *Service) verifyDeleteToken(rawToken string) (deleteTokenPayload, error) {
	token := strings.TrimSpace(rawToken)
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return deleteTokenPayload{}, errDeleteTokenInvalid
	}

	mac := hmac.New(sha256.New, []byte(s.cfg.TokenSigningKey))
	_, _ = mac.Write([]byte(parts[0]))
	expectedSignature := mac.Sum(nil)
	actualSignature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return deleteTokenPayload{}, errDeleteTokenInvalid
	}
	if !hmac.Equal(actualSignature, expectedSignature) {
		return deleteTokenPayload{}, errDeleteTokenInvalid
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return deleteTokenPayload{}, errDeleteTokenInvalid
	}
	var payload deleteTokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return deleteTokenPayload{}, errDeleteTokenInvalid
	}
	if strings.TrimSpace(payload.RunID) == "" || strings.TrimSpace(payload.Namespace) == "" {
		return deleteTokenPayload{}, errDeleteTokenInvalid
	}
	return payload, nil
}
