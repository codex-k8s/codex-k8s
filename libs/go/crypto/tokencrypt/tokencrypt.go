package tokencrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// Service encrypts/decrypts small secret strings for storage in DB.
//
// Storage format: nonce || ciphertext (as raw bytes).
// Nonce length is defined by the underlying AEAD (AES-GCM => 12 bytes).
//
// This is intended for tokens (repo/agent), not for large blobs.
type Service struct {
	aead cipher.AEAD
	rand io.Reader
}

// NewService constructs a token encryption service using a hex-encoded 32-byte key.
// The key must be 64 hex characters (32 bytes).
func NewService(hexKey string) (*Service, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("decode hex key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: expected 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("init aes cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("init aes-gcm: %w", err)
	}

	return &Service{
		aead: aead,
		rand: rand.Reader,
	}, nil
}

// EncryptString encrypts plaintext and returns ciphertext bytes (nonce||ciphertext).
func (s *Service) EncryptString(plaintext string) ([]byte, error) {
	if s == nil || s.aead == nil {
		return nil, errors.New("tokencrypt service is not initialized")
	}
	if plaintext == "" {
		return nil, errors.New("plaintext is empty")
	}

	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(s.rand, nonce); err != nil {
		return nil, fmt.Errorf("read nonce: %w", err)
	}

	ciphertext := s.aead.Seal(nil, nonce, []byte(plaintext), nil)
	out := make([]byte, 0, len(nonce)+len(ciphertext))
	out = append(out, nonce...)
	out = append(out, ciphertext...)

	return out, nil
}

// DecryptString decrypts ciphertext bytes (nonce||ciphertext) and returns plaintext.
func (s *Service) DecryptString(ciphertext []byte) (string, error) {
	if s == nil || s.aead == nil {
		return "", errors.New("tokencrypt service is not initialized")
	}
	if len(ciphertext) == 0 {
		return "", errors.New("ciphertext is empty")
	}

	if len(ciphertext) < s.aead.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce := ciphertext[:s.aead.NonceSize()]
	enc := ciphertext[s.aead.NonceSize():]

	plain, err := s.aead.Open(nil, nonce, enc, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plain), nil
}
