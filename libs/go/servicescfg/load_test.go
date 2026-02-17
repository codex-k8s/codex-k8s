package servicescfg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_WithImportsComponentsAndTemplates(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	baseFile := filepath.Join(tmpDir, "base.yaml")
	rootFile := filepath.Join(tmpDir, "services.yaml")

	base := `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-production"
  components:
    - name: go-default
      serviceDefaults:
        codeUpdateStrategy: hot-reload
        deployGroup: internal
`
	root := `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  imports:
    - path: base.yaml
  services:
    - name: control-plane
      use: [go-default]
      codeUpdateStrategy: restart
    - name: worker
`

	writeFile(t, baseFile, base)
	writeFile(t, rootFile, root)

	result, err := Load(rootFile, LoadOptions{Env: "production"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if got, want := result.Context.Namespace, "demo-production"; got != want {
		t.Fatalf("unexpected namespace: got %q want %q", got, want)
	}

	if len(result.Stack.Spec.Services) != 2 {
		t.Fatalf("unexpected services count: %d", len(result.Stack.Spec.Services))
	}

	controlPlane := result.Stack.Spec.Services[0]
	if got, want := controlPlane.CodeUpdateStrategy, CodeUpdateStrategyRestart; got != want {
		t.Fatalf("unexpected control-plane strategy: got %q want %q", got, want)
	}
	if got, want := controlPlane.DeployGroup, "internal"; got != want {
		t.Fatalf("unexpected control-plane deployGroup: got %q want %q", got, want)
	}

	worker := result.Stack.Spec.Services[1]
	if got, want := worker.CodeUpdateStrategy, CodeUpdateStrategyRebuild; got != want {
		t.Fatalf("unexpected worker strategy: got %q want %q", got, want)
	}
}

func TestLoad_ImportCycle(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	first := filepath.Join(tmpDir, "first.yaml")
	second := filepath.Join(tmpDir, "second.yaml")

	writeFile(t, first, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  imports:
    - path: second.yaml
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-production"
`)
	writeFile(t, second, `
spec:
  imports:
    - path: first.yaml
`)

	_, err := Load(first, LoadOptions{Env: "production"})
	if err == nil {
		t.Fatalf("expected cycle error")
	}
	if !strings.Contains(err.Error(), "imports cycle detected") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoad_UnknownComponentReference(t *testing.T) {
	t.Parallel()

	assertLoadErrorContains(t, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-production"
  services:
    - name: api
      use: [unknown-component]
`, "unknown component")
}

func TestLoad_CodexK8sRequiresProductionTemplate(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "services.yaml")
	writeFile(t, path, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: codex-k8s
spec:
  environments:
    production:
      namespaceTemplate: "hardcoded-namespace"
`)

	_, err := Load(path, LoadOptions{Env: "production"})
	if err == nil {
		t.Fatalf("expected codex-k8s production template validation error")
	}
	if !strings.Contains(err.Error(), "codex-k8s requires production namespace template") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveEnvironment_Inheritance(t *testing.T) {
	t.Parallel()

	stack := &Stack{
		Spec: Spec{
			Environments: map[string]Environment{
				"production": {NamespaceTemplate: "{{ .Project }}-production", DomainTemplate: "demo.example.com", ImagePullPolicy: "Always"},
				"ai":         {From: "production"},
			},
		},
	}

	resolved, err := ResolveEnvironment(stack, "ai")
	if err != nil {
		t.Fatalf("resolve environment: %v", err)
	}
	if got, want := resolved.NamespaceTemplate, "{{ .Project }}-production"; got != want {
		t.Fatalf("unexpected namespaceTemplate: got %q want %q", got, want)
	}
	if got, want := resolved.ImagePullPolicy, "Always"; got != want {
		t.Fatalf("unexpected imagePullPolicy: got %q want %q", got, want)
	}
	if got, want := resolved.DomainTemplate, "demo.example.com"; got != want {
		t.Fatalf("unexpected domainTemplate: got %q want %q", got, want)
	}
}

func TestLoadFromYAML_RendersDomainTemplate(t *testing.T) {
	t.Parallel()

	raw := []byte(strings.TrimSpace(`
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
      domainTemplate: "{{ .Namespace }}.example.com"
`))

	result, err := LoadFromYAML(raw, LoadOptions{Env: "production"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got, want := result.Context.Namespace, "demo-prod"; got != want {
		t.Fatalf("unexpected namespace: got %q want %q", got, want)
	}

	envCfg, err := ResolveEnvironment(result.Stack, "production")
	if err != nil {
		t.Fatalf("resolve environment: %v", err)
	}
	if got, want := strings.TrimSpace(envCfg.DomainTemplate), "demo-prod.example.com"; got != want {
		t.Fatalf("unexpected domainTemplate: got %q want %q", got, want)
	}
}

func TestLoad_WebhookRuntimeModes(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "services.yaml")
	writeFile(t, path, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-production"
  webhookRuntime:
    defaultMode: full-env
    triggerModes:
      self_improve: code-only
      dev: full-env
`)

	result, err := Load(path, LoadOptions{Env: "production"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if got, want := ResolveTriggerRuntimeMode(result.Stack, "self_improve"), RuntimeModeCodeOnly; got != want {
		t.Fatalf("unexpected runtime mode for self_improve: got %q want %q", got, want)
	}
	if got, want := ResolveTriggerRuntimeMode(result.Stack, "dev"), RuntimeModeFullEnv; got != want {
		t.Fatalf("unexpected runtime mode for dev: got %q want %q", got, want)
	}
	if got, want := ResolveTriggerRuntimeMode(result.Stack, "unknown"), RuntimeModeFullEnv; got != want {
		t.Fatalf("unexpected runtime mode for unknown trigger: got %q want %q", got, want)
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertLoadErrorContains(t *testing.T, content string, wantSubstring string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "services.yaml")
	writeFile(t, path, content)

	_, err := Load(path, LoadOptions{Env: "production"})
	if err == nil {
		t.Fatalf("expected load error with substring %q", wantSubstring)
	}
	if !strings.Contains(err.Error(), wantSubstring) {
		t.Fatalf("unexpected error: %v", err)
	}
}
