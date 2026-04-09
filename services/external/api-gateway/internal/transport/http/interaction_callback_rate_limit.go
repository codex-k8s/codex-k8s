package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/codex-k8s/kodex/libs/go/errs"
	"github.com/codex-k8s/kodex/services/external/api-gateway/internal/transport/http/models"
)

const (
	interactionCallbackIdentifierContextKey = "mcp_interaction_callback_identifier"
	interactionCallbackRateLimitRPS         = 5
	interactionCallbackRateLimitBurst       = 20
	interactionCallbackRateLimitExpiresIn   = 5 * time.Minute
	interactionCallbackBodyMaxBytes         = 64 * 1024
)

type interactionCallbackIdentifierEnvelope struct {
	InteractionID string `json:"interaction_id"`
}

func newInteractionCallbackRateLimitMiddleware(maxBodyBytes int64) echo.MiddlewareFunc {
	if maxBodyBytes <= 0 {
		maxBodyBytes = interactionCallbackBodyMaxBytes
	}

	limiter := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate:      interactionCallbackRateLimitRPS,
			Burst:     interactionCallbackRateLimitBurst,
			ExpiresIn: interactionCallbackRateLimitExpiresIn,
		}),
		IdentifierExtractor: func(c *echo.Context) (string, error) {
			if identifier, ok := c.Get(interactionCallbackIdentifierContextKey).(string); ok && strings.TrimSpace(identifier) != "" {
				return identifier, nil
			}
			return "ip:" + strings.TrimSpace(c.RealIP()), nil
		},
		ErrorHandler: func(c *echo.Context, err error) error {
			return errs.Validation{Field: "interaction_id", Msg: "cannot derive callback identifier"}
		},
		DenyHandler: func(c *echo.Context, identifier string, err error) error {
			return c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Code:    "resource_exhausted",
				Message: "rate limit exceeded",
			})
		},
	})

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		limited := limiter(next)
		return func(c *echo.Context) error {
			if err := cacheInteractionCallbackIdentifier(c, maxBodyBytes); err != nil {
				return err
			}
			return limited(c)
		}
	}
}

func cacheInteractionCallbackIdentifier(c *echo.Context, maxBodyBytes int64) error {
	payload, err := readRequestBody(c.Request().Body, maxBodyBytes)
	if err != nil {
		return err
	}

	c.Request().Body = http.NoBody
	if len(payload) > 0 {
		c.Request().Body = io.NopCloser(bytes.NewReader(payload))
	}

	var envelope interactionCallbackIdentifierEnvelope
	if err := json.Unmarshal(payload, &envelope); err == nil {
		if interactionID := strings.TrimSpace(envelope.InteractionID); interactionID != "" {
			c.Set(interactionCallbackIdentifierContextKey, "interaction:"+interactionID)
			return nil
		}
	}

	c.Set(interactionCallbackIdentifierContextKey, "ip:"+strings.TrimSpace(c.RealIP()))
	return nil
}
