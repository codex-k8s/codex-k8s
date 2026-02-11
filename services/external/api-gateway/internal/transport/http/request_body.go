package http

import (
	"fmt"
	"io"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
)

func readRequestBody(body io.ReadCloser, maxBodyBytes int64) ([]byte, error) {
	defer func() { _ = body.Close() }()

	if maxBodyBytes <= 0 {
		maxBodyBytes = 1024 * 1024
	}

	limitedReader := io.LimitReader(body, maxBodyBytes+1)
	payload, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if int64(len(payload)) > maxBodyBytes {
		return nil, errs.Validation{
			Field: "body",
			Msg:   fmt.Sprintf("payload too large (max %d bytes)", maxBodyBytes),
		}
	}
	return payload, nil
}
