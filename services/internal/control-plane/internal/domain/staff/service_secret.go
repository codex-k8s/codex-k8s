package staff

import (
	"fmt"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
)

// EncryptSecretValue encrypts plain secret value using platform token-crypto service.
func (s *Service) EncryptSecretValue(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errs.Validation{Field: "value_secret", Msg: "is required"}
	}
	enc, err := s.tokencrypt.EncryptString(value)
	if err != nil {
		return nil, fmt.Errorf("encrypt secret value: %w", err)
	}
	return enc, nil
}
