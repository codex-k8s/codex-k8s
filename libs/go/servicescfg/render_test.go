package servicescfg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRenderer_Render_WithIncludeAndPartial(t *testing.T) {
	tmpDir := t.TempDir()
	partDir := filepath.Join(tmpDir, "deploy", "_tpl")
	if err := os.MkdirAll(partDir, 0o755); err != nil {
		t.Fatalf("mkdir partial dir: %v", err)
	}
	partialPath := filepath.Join(partDir, "common.gohtml")
	if err := os.WriteFile(partialPath, []byte(`{{ define "labels.common" }}app.kubernetes.io/name: codex-k8s{{ end }}`), 0o644); err != nil {
		t.Fatalf("write partial: %v", err)
	}

	cfg := &Config{
		Project: "codex-k8s",
		Templates: TemplatesConfig{
			Partials: []string{"deploy/_tpl/*.gohtml"},
		},
	}
	renderer, err := NewRenderer(cfg, tmpDir, RenderContext{
		Now:    time.Date(2026, time.February, 14, 12, 0, 0, 0, time.UTC),
		EnvMap: map[string]string{"CODEXK8S_STAGING_NAMESPACE": "codex-k8s-ai-staging"},
	})
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	out, err := renderer.Render("manifest", []byte("metadata:\n  namespace: ${CODEXK8S_STAGING_NAMESPACE}\n  labels:\n{{ include \"labels.common\" . | indent 4 }}\n"), map[string]string{})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	rendered := string(out)
	if !strings.Contains(rendered, "namespace: codex-k8s-ai-staging") {
		t.Fatalf("expected rendered namespace, got %q", rendered)
	}
	if !strings.Contains(rendered, "app.kubernetes.io/name: codex-k8s") {
		t.Fatalf("expected partial include content, got %q", rendered)
	}
}

func TestRenderer_PartialDefineConflict(t *testing.T) {
	tmpDir := t.TempDir()
	partDir := filepath.Join(tmpDir, "deploy", "_tpl")
	if err := os.MkdirAll(partDir, 0o755); err != nil {
		t.Fatalf("mkdir partial dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partDir, "a.gohtml"), []byte(`{{ define "dup.name" }}A{{ end }}`), 0o644); err != nil {
		t.Fatalf("write partial a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partDir, "b.gohtml"), []byte(`{{ define "dup.name" }}B{{ end }}`), 0o644); err != nil {
		t.Fatalf("write partial b: %v", err)
	}

	cfg := &Config{
		Project: "codex-k8s",
		Templates: TemplatesConfig{
			Partials: []string{"deploy/_tpl/*.gohtml"},
		},
	}
	_, err := NewRenderer(cfg, tmpDir, RenderContext{})
	if err == nil {
		t.Fatal("expected duplicate partial define error, got nil")
	}
	if !strings.Contains(err.Error(), "partial template define conflict") {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestRenderer_MissingPartialGlob(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		Project: "codex-k8s",
		Templates: TemplatesConfig{
			Partials: []string{"deploy/_tpl/*.gohtml"},
		},
	}
	_, err := NewRenderer(cfg, tmpDir, RenderContext{})
	if err == nil {
		t.Fatal("expected missing glob error, got nil")
	}
	if !strings.Contains(err.Error(), "did not match files") {
		t.Fatalf("expected missing glob error, got %v", err)
	}
}
