package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/services/external/api-gateway/internal/domain/errs"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func newHTTPErrorHandler(logger *slog.Logger) func(c *echo.Context, err error) {
	return func(c *echo.Context, err error) {
		status := http.StatusInternalServerError
		resp := errorResponse{
			Code:    "internal",
			Message: "internal error",
		}

		var validation errs.Validation
		var unauthorized errs.Unauthorized
		var forbidden errs.Forbidden
		var conflict errs.Conflict
		var httpErr *echo.HTTPError

		switch {
		case errors.As(err, &validation):
			status = http.StatusBadRequest
			resp = errorResponse{
				Code:    "invalid_argument",
				Message: validation.Msg,
				Field:   validation.Field,
			}
		case errors.As(err, &unauthorized):
			status = http.StatusUnauthorized
			resp = errorResponse{
				Code:    "unauthorized",
				Message: unauthorized.Error(),
			}
		case errors.As(err, &forbidden):
			status = http.StatusForbidden
			resp = errorResponse{
				Code:    "forbidden",
				Message: forbidden.Error(),
			}
		case errors.As(err, &conflict):
			status = http.StatusConflict
			resp = errorResponse{
				Code:    "conflict",
				Message: conflict.Error(),
			}
		case errors.As(err, &httpErr):
			status = httpErr.Code
			resp = errorResponse{
				Code:    "invalid_argument",
				Message: http.StatusText(httpErr.Code),
			}
			if httpErr.Message != "" {
				resp.Message = httpErr.Message
			}
		}

		// Log only server-side failures.
		if status >= 500 {
			logger.Error("request failed",
				"status", status,
				"method", c.Request().Method,
				"path", c.Path(),
				"error", err.Error(),
			)
		}

		_ = c.JSON(status, resp)
	}
}
