package githubsignature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifySHA256(t *testing.T) {
	secret := []byte("top-secret")
	payload := []byte(`{"hello":"world"}`)

	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(payload)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if err := VerifySHA256(secret, payload, signature); err != nil {
		t.Fatalf("expected valid signature, got error: %v", err)
	}
}

func TestVerifySHA256_Invalid(t *testing.T) {
	secret := []byte("top-secret")
	payload := []byte(`{"hello":"world"}`)

	err := VerifySHA256(secret, payload, "sha256=deadbeef")
	if err == nil {
		t.Fatal("expected invalid signature error")
	}
}
