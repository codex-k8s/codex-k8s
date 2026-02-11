package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
)

const statusClientClosedRequest = 499

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func newHTTPErrorHandler(logger *slog.Logger) func(c *echo.Context, err error) {
	return func(c *echo.Context, err error) {
		statusCode := http.StatusInternalServerError
		resp := errorResponse{
			Code:    "internal",
			Message: "internal error",
		}
		logAsError := true
		logAsWarn := false

		// Echo may return plain sentinel errors for routing/method mismatches.
		if errors.Is(err, echo.ErrNotFound) {
			_ = c.NoContent(http.StatusNotFound)
			return
		}
		if errors.Is(err, echo.ErrMethodNotAllowed) {
			_ = c.NoContent(http.StatusMethodNotAllowed)
			return
		}

		var validation errs.Validation
		var unauthorized errs.Unauthorized
		var forbidden errs.Forbidden
		var conflict errs.Conflict
		var httpErr *echo.HTTPError
		grpcErr := grpcStatus(err)

		switch {
		case errors.Is(err, context.Canceled):
			statusCode = statusClientClosedRequest
			resp = errorResponse{
				Code:    "canceled",
				Message: "request canceled",
			}
			logAsError = false
		case errors.Is(err, context.DeadlineExceeded):
			statusCode = http.StatusGatewayTimeout
			resp = errorResponse{
				Code:    "deadline_exceeded",
				Message: "request deadline exceeded",
			}
			logAsError = false
			logAsWarn = true
		case errors.As(err, &validation):
			statusCode = http.StatusBadRequest
			resp = errorResponse{
				Code:    "invalid_argument",
				Message: validation.Msg,
				Field:   validation.Field,
			}
			logAsError = false
		case errors.As(err, &unauthorized):
			statusCode = http.StatusUnauthorized
			resp = errorResponse{
				Code:    "unauthorized",
				Message: unauthorized.Error(),
			}
			logAsError = false
		case errors.As(err, &forbidden):
			statusCode = http.StatusForbidden
			resp = errorResponse{
				Code:    "forbidden",
				Message: forbidden.Error(),
			}
			logAsError = false
		case errors.As(err, &conflict):
			statusCode = http.StatusConflict
			resp = errorResponse{
				Code:    "conflict",
				Message: conflict.Error(),
			}
			logAsError = false
		case grpcErr != nil:
			statusCode, resp, logAsError, logAsWarn = mapGRPCStatus(grpcErr)
		case errors.As(err, &httpErr):
			statusCode = httpErr.Code
			resp = errorResponse{
				Code:    "invalid_argument",
				Message: http.StatusText(httpErr.Code),
			}
			if statusCode < 500 && httpErr.Message != "" {
				resp.Message = httpErr.Message
			}
			logAsError = statusCode >= 500
		}

		if logAsError {
			logger.Error("request failed",
				"status", statusCode,
				"method", c.Request().Method,
				"path", c.Path(),
				"error", err.Error(),
			)
		}
		if !logAsError && logAsWarn {
			logger.Warn("request degraded",
				"status", statusCode,
				"method", c.Request().Method,
				"path", c.Path(),
				"error", err.Error(),
			)
		}

		_ = c.JSON(statusCode, resp)
	}
}

func grpcStatus(err error) *status.Status {
	st, ok := status.FromError(err)
	if !ok {
		return nil
	}
	return st
}

func mapGRPCStatus(st *status.Status) (int, errorResponse, bool, bool) {
	switch st.Code() {
	case codes.InvalidArgument:
		return http.StatusBadRequest, errorResponse{
			Code:    "invalid_argument",
			Message: defaultMessage(st.Message(), "invalid request"),
		}, false, false
	case codes.Unauthenticated:
		return http.StatusUnauthorized, errorResponse{
			Code:    "unauthorized",
			Message: defaultMessage(st.Message(), "not authenticated"),
		}, false, false
	case codes.PermissionDenied:
		return http.StatusForbidden, errorResponse{
			Code:    "forbidden",
			Message: defaultMessage(st.Message(), "forbidden"),
		}, false, false
	case codes.NotFound:
		return http.StatusNotFound, errorResponse{
			Code:    "not_found",
			Message: defaultMessage(st.Message(), "resource not found"),
		}, false, false
	case codes.AlreadyExists:
		return http.StatusConflict, errorResponse{
			Code:    "conflict",
			Message: defaultMessage(st.Message(), "resource already exists"),
		}, false, false
	case codes.FailedPrecondition:
		return http.StatusConflict, errorResponse{
			Code:    "failed_precondition",
			Message: defaultMessage(st.Message(), "operation precondition failed"),
		}, false, false
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests, errorResponse{
			Code:    "resource_exhausted",
			Message: "rate limit exceeded",
		}, false, false
	case codes.Canceled:
		return statusClientClosedRequest, errorResponse{
			Code:    "canceled",
			Message: "request canceled",
		}, false, false
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, errorResponse{
			Code:    "deadline_exceeded",
			Message: "request deadline exceeded",
		}, false, true
	case codes.Unavailable:
		return http.StatusServiceUnavailable, errorResponse{
			Code:    "unavailable",
			Message: "service unavailable",
		}, false, true
	default:
		return http.StatusInternalServerError, errorResponse{
			Code:    "internal",
			Message: "internal error",
		}, true, false
	}
}

func defaultMessage(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
