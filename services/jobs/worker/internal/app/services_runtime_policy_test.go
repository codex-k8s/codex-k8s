package app

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadWebhookRuntimeNamespaceTTLPolicy_DefaultFallback(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	defaultTTL, byRole := loadWebhookRuntimeNamespaceTTLPolicy(Config{}, logger)
	if defaultTTL != 24*time.Hour {
		t.Fatalf("expected default ttl 24h, got %s", defaultTTL)
	}
	if len(byRole) != 0 {
		t.Fatalf("expected empty role ttl map, got %d entries", len(byRole))
	}
}

func TestLoadWebhookRuntimeNamespaceTTLPolicy_FromServicesConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "services.yaml")
	writePolicyFixture(t, path, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  webhookRuntime:
    defaultNamespaceTTL: 36h
    namespaceTTLByRole:
      dev: 12h
      qa: 2h
`)

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	defaultTTL, byRole := loadWebhookRuntimeNamespaceTTLPolicy(Config{
		ServicesConfigPath: path,
		ServicesConfigEnv:  "production",
	}, logger)

	if defaultTTL != 36*time.Hour {
		t.Fatalf("expected default ttl 36h, got %s", defaultTTL)
	}
	if got, want := byRole["dev"], 12*time.Hour; got != want {
		t.Fatalf("unexpected dev ttl: got %s want %s", got, want)
	}
	if got, want := byRole["qa"], 2*time.Hour; got != want {
		t.Fatalf("unexpected qa ttl: got %s want %s", got, want)
	}
}

func writePolicyFixture(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
}
