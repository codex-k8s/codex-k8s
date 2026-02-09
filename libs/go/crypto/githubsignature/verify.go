package githubsignature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

const prefixSHA256 = "sha256="

var (
	// ErrMissingSignature indicates that X-Hub-Signature-256 header is absent.
	ErrMissingSignature = errors.New("missing github signature")
	// ErrInvalidFormat indicates that signature header does not use sha256= prefix.
	ErrInvalidFormat = errors.New("invalid github signature format")
	// ErrInvalidHex indicates that signature value is not valid hexadecimal.
	ErrInvalidHex = errors.New("invalid github signature hex")
	// ErrInvalidSignature indicates that computed HMAC does not match the header.
	ErrInvalidSignature = errors.New("invalid github signature")
)

// VerifySHA256 validates X-Hub-Signature-256 against the raw payload.
func VerifySHA256(secret, payload []byte, signatureHeader string) error {
	if len(signatureHeader) == 0 {
		return ErrMissingSignature
	}

	if !strings.HasPrefix(signatureHeader, prefixSHA256) {
		return ErrInvalidFormat
	}

	gotHex := strings.TrimPrefix(signatureHeader, prefixSHA256)
	gotSig, err := hex.DecodeString(gotHex)
	if err != nil {
		return ErrInvalidHex
	}

	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(payload)
	wantSig := mac.Sum(nil)

	if !hmac.Equal(wantSig, gotSig) {
		return ErrInvalidSignature
	}

	return nil
}
