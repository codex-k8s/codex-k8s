package http

import (
	"net/http/httptest"
	"testing"
)

func TestIsRealtimeOriginAllowed(t *testing.T) {
	t.Run("allow when origin header missing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "https://platform.codex-k8s.dev/api/v1/staff/realtime/ws", nil)
		req.Host = "platform.codex-k8s.dev"
		if !isRealtimeOriginAllowed(req) {
			t.Fatal("expected missing Origin header to be allowed")
		}
	})

	t.Run("allow same host", func(t *testing.T) {
		req := httptest.NewRequest("GET", "https://platform.codex-k8s.dev/api/v1/staff/realtime/ws", nil)
		req.Host = "platform.codex-k8s.dev"
		req.Header.Set("Origin", "https://platform.codex-k8s.dev")
		if !isRealtimeOriginAllowed(req) {
			t.Fatal("expected same host Origin to be allowed")
		}
	})

	t.Run("allow forwarded host", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://api-gateway:8080/api/v1/staff/realtime/ws", nil)
		req.Host = "api-gateway:8080"
		req.Header.Set("X-Forwarded-Host", "platform.codex-k8s.dev")
		req.Header.Set("Origin", "https://platform.codex-k8s.dev")
		if !isRealtimeOriginAllowed(req) {
			t.Fatal("expected Origin matching X-Forwarded-Host to be allowed")
		}
	})

	t.Run("deny different host", func(t *testing.T) {
		req := httptest.NewRequest("GET", "https://platform.codex-k8s.dev/api/v1/staff/realtime/ws", nil)
		req.Host = "platform.codex-k8s.dev"
		req.Header.Set("Origin", "https://evil.example.com")
		if isRealtimeOriginAllowed(req) {
			t.Fatal("expected different host Origin to be denied")
		}
	})

	t.Run("deny invalid origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "https://platform.codex-k8s.dev/api/v1/staff/realtime/ws", nil)
		req.Host = "platform.codex-k8s.dev"
		req.Header.Set("Origin", "://bad")
		if isRealtimeOriginAllowed(req) {
			t.Fatal("expected invalid Origin to be denied")
		}
	})
}
