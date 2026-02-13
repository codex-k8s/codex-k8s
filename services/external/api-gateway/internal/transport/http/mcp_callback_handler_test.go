package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/codex-k8s/codex-k8s/libs/go/errs"
)

func TestResolveMCPCallbackToken(t *testing.T) {
	tests := []struct {
		name          string
		callbackToken string
		authHeader    string
		want          string
	}{
		{
			name:          "header token has priority",
			callbackToken: "token-1",
			authHeader:    "Bearer token-2",
			want:          "token-1",
		},
		{
			name:          "bearer token fallback",
			callbackToken: "",
			authHeader:    "Bearer token-2",
			want:          "token-2",
		},
		{
			name:          "bearer token fallback case insensitive",
			callbackToken: "",
			authHeader:    "bearer token-3",
			want:          "token-3",
		},
		{
			name:          "missing token",
			callbackToken: "",
			authHeader:    "",
			want:          "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := resolveMCPCallbackToken(tt.callbackToken, tt.authHeader)
			if got != tt.want {
				t.Fatalf("resolveMCPCallbackToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsMCPDecisionAllowed(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "approved", value: "approved", want: true},
		{name: "applied", value: "applied", want: true},
		{name: "trimmed", value: " denied ", want: true},
		{name: "invalid", value: "retry", want: false},
		{name: "empty", value: "", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := isMCPDecisionAllowed(tt.value); got != tt.want {
				t.Fatalf("isMCPDecisionAllowed(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestMCPCallbackHandlerRejectsWhenNotConfigured(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/mcp/approver/callback", strings.NewReader(`{"approval_request_id":1,"decision":"approved"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	h := &mcpCallbackHandler{}
	err := h.CallbackApprover(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var unauthorized errs.Unauthorized
	if !errors.As(err, &unauthorized) {
		t.Fatalf("expected errs.Unauthorized, got %T", err)
	}
}
