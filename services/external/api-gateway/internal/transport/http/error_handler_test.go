package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHTTPErrorHandler_GRPCStatusMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		err      error
		wantCode int
		wantBody string
	}{
		{
			name:     "invalid argument",
			err:      status.Error(codes.InvalidArgument, "project_id: is required"),
			wantCode: http.StatusBadRequest,
			wantBody: "invalid_argument",
		},
		{
			name:     "internal is sanitized",
			err:      status.Error(codes.Internal, "pq: relation users does not exist"),
			wantCode: http.StatusInternalServerError,
			wantBody: "internal",
		},
		{
			name:     "not found passthrough",
			err:      echo.ErrNotFound,
			wantCode: http.StatusNotFound,
		},
		{
			name:     "non-grpc fallback",
			err:      errors.New("boom"),
			wantCode: http.StatusInternalServerError,
			wantBody: "internal",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotCode, gotBody := runErrorHandler(tc.err)
			if gotCode != tc.wantCode {
				t.Fatalf("expected status %d, got %d", tc.wantCode, gotCode)
			}
			if tc.wantBody != "" && gotBody.Code != tc.wantBody {
				t.Fatalf("expected code %q, got %q", tc.wantBody, gotBody.Code)
			}
			if tc.name == "internal is sanitized" && gotBody.Message != "internal error" {
				t.Fatalf("expected sanitized message, got %q", gotBody.Message)
			}
		})
	}
}

func runErrorHandler(err error) (int, errorResponse) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/staff/projects", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	newHTTPErrorHandler(slog.New(slog.NewTextHandler(ioDiscard{}, nil)))(ctx, err)

	var resp errorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	return rec.Code, resp
}
