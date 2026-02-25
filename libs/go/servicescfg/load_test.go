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
    defaultNamespaceTTL: 24h
    namespaceTTLByRole:
      dev: 24h
      QA: 30m
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

	if got, want := result.Stack.Spec.WebhookRuntime.DefaultNamespaceTTL, "24h0m0s"; got != want {
		t.Fatalf("unexpected default namespace ttl: got %q want %q", got, want)
	}
	if got, want := result.Stack.Spec.WebhookRuntime.NamespaceTTLByRole["dev"], "24h0m0s"; got != want {
		t.Fatalf("unexpected dev namespace ttl: got %q want %q", got, want)
	}
	if got, want := result.Stack.Spec.WebhookRuntime.NamespaceTTLByRole["qa"], "30m0s"; got != want {
		t.Fatalf("unexpected qa namespace ttl: got %q want %q", got, want)
	}
	if _, ok := result.Stack.Spec.WebhookRuntime.NamespaceTTLByRole["QA"]; ok {
		t.Fatalf("expected namespace ttl role keys to be normalized to lower-case")
	}
}

func TestLoad_WebhookRuntimeNamespaceTTLValidation(t *testing.T) {
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
  webhookRuntime:
    defaultMode: full-env
    defaultNamespaceTTL: zero
`, "defaultNamespaceTTL")

	assertLoadErrorContains(t, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-production"
  webhookRuntime:
    namespaceTTLByRole:
      dev: bogus
`, "namespaceTTLByRole")
}

func TestLoad_SecretResolutionContract(t *testing.T) {
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
      namespaceTemplate: "{{ .Project }}-prod"
  secretResolution:
    environmentAliases:
      production: [prod]
    keyOverrides:
      - sourceKey: CODEXK8S_GITHUB_OAUTH_CLIENT_ID
        overrideKeys:
          ai: CODEXK8S_GITHUB_OAUTH_CLIENT_ID_AI
    patterns:
      - sourcePrefix: CODEXK8S_
        excludeSuffixes: [_AI]
        environments: [ai]
        overrideTemplate: "{key}_{env_upper}"
`)

	result, err := Load(path, LoadOptions{Env: "production"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	resolver := NewSecretResolver(result.Stack)
	overrideKey, ok := resolver.ResolveOverrideKey("ai", "CODEXK8S_GITHUB_OAUTH_CLIENT_ID")
	if !ok {
		t.Fatalf("expected explicit key override")
	}
	if got, want := overrideKey, "CODEXK8S_GITHUB_OAUTH_CLIENT_ID_AI"; got != want {
		t.Fatalf("unexpected override key: got %q want %q", got, want)
	}
}

func TestLoad_ServiceScopeNormalization(t *testing.T) {
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
      namespaceTemplate: "{{ .Project }}-prod"
  services:
    - name: singleton
      scope: infrastructure-singleton
    - name: defaulted
`)

	result, err := Load(path, LoadOptions{Env: "production"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got, want := result.Stack.Spec.Services[0].Scope, ServiceScopeInfrastructureSingleton; got != want {
		t.Fatalf("unexpected singleton scope: got %q want %q", got, want)
	}
	if got, want := result.Stack.Spec.Services[1].Scope, ServiceScopeEnvironment; got != want {
		t.Fatalf("unexpected default scope: got %q want %q", got, want)
	}
}

func TestLoad_ServiceScopeValidation(t *testing.T) {
	t.Parallel()

	assertLoadErrorContains(t, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  services:
    - name: invalid
      scope: cluster
`, "serviceScope")
}

func TestLoad_ProjectDocsNormalizationAndRoles(t *testing.T) {
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
      namespaceTemplate: "{{ .Project }}-prod"
  projectDocs:
    - path: ./docs/../docs/README.md
      repository: Core-Docs
      description: "  Main handbook  "
      roles: [DEV, qa, DEV]
`)

	result, err := Load(path, LoadOptions{Env: "production"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(result.Stack.Spec.ProjectDocs) != 1 {
		t.Fatalf("projectDocs len=%d, want 1", len(result.Stack.Spec.ProjectDocs))
	}
	item := result.Stack.Spec.ProjectDocs[0]
	if got, want := item.Repository, "core-docs"; got != want {
		t.Fatalf("projectDocs[0].repository=%q, want %q", got, want)
	}
	if got, want := item.Path, "docs/README.md"; got != want {
		t.Fatalf("projectDocs[0].path=%q, want %q", got, want)
	}
	if got, want := item.Description, "Main handbook"; got != want {
		t.Fatalf("projectDocs[0].description=%q, want %q", got, want)
	}
	if got, want := strings.Join(item.Roles, ","), "dev,qa"; got != want {
		t.Fatalf("projectDocs[0].roles=%q, want %q", got, want)
	}
}

func TestLoad_ProjectDocsValidation(t *testing.T) {
	t.Parallel()

	assertLoadErrorContains(t, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  projectDocs:
    - path: ../outside.md
`, "projectDocs[0].path")

	assertLoadErrorContains(t, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  projectDocs:
    - path: docs/README.md
      repository: docs
    - path: ./docs/README.md
      repository: docs
`, "duplicate spec.projectDocs entry")

	path := filepath.Join(t.TempDir(), "services-allow-duplicate-paths.yaml")
	writeFile(t, path, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  projectDocs:
    - repository: docs
      path: docs/README.md
    - repository: service-api
      path: docs/README.md
`)
	if _, err := Load(path, LoadOptions{Env: "production"}); err != nil {
		t.Fatalf("expected duplicate path in different repositories to be valid: %v", err)
	}

	assertLoadErrorContains(t, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  projectDocs:
    - path: docs/README.md
      repository: "@invalid"
`, "projectDocs[0].repository")
}

func TestLoad_RoleDocTemplatesNormalization(t *testing.T) {
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
      namespaceTemplate: "{{ .Project }}-prod"
  roleDocTemplates:
    PM:
      - path: ./docs/templates/../templates/prd.md
        repository: Core-Docs
        description: "  PRD template  "
      - path: docs/templates/nfr.md
        repository: core-docs
`)

	result, err := Load(path, LoadOptions{Env: "production"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	templates, ok := result.Stack.Spec.RoleDocTemplates["pm"]
	if !ok {
		t.Fatalf("expected normalized role key pm")
	}
	if len(templates) != 2 {
		t.Fatalf("roleDocTemplates[pm] len=%d, want 2", len(templates))
	}
	if got, want := templates[0].Repository, "core-docs"; got != want {
		t.Fatalf("roleDocTemplates[pm][0].repository=%q, want %q", got, want)
	}
	if got, want := templates[0].Path, "docs/templates/nfr.md"; got != want {
		t.Fatalf("roleDocTemplates[pm][0].path=%q, want %q", got, want)
	}
	if got, want := templates[1].Path, "docs/templates/prd.md"; got != want {
		t.Fatalf("roleDocTemplates[pm][1].path=%q, want %q", got, want)
	}
	if got, want := templates[1].Description, "PRD template"; got != want {
		t.Fatalf("roleDocTemplates[pm][1].description=%q, want %q", got, want)
	}
}

func TestLoad_RoleDocTemplatesValidation(t *testing.T) {
	t.Parallel()

	assertLoadErrorContains(t, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  roleDocTemplates:
    pm: []
`, "spec.roleDocTemplates[\"pm\"] must contain at least one template")

	assertLoadErrorContains(t, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  roleDocTemplates:
    pm:
      - path: docs/templates/prd.md
      - path: ./docs/templates/prd.md
`, "duplicate spec.roleDocTemplates entry for role \"pm\" path \"docs/templates/prd.md\"")

	assertLoadErrorContains(t, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  roleDocTemplates:
    pm:
      - path: ../outside.md
`, "spec.roleDocTemplates[\"pm\"][0].path")

	assertLoadErrorContains(t, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  roleDocTemplates:
    pm:
      - path: docs/templates/prd.md
        repository: "@invalid"
`, "spec.roleDocTemplates[\"pm\"][0].repository")
}

func TestLoadFromYAML_SchemaValidationErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		raw         string
		errContains string
	}{
		{
			name:        "services_not_array",
			errContains: "services schema validation failed",
			raw: `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  services: 123
`,
		},
		{
			name:        "version_scalar_not_allowed",
			errContains: "cannot unmarshal",
			raw: `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  versions:
    worker: "0.1.0"
`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := LoadFromYAML([]byte(strings.TrimSpace(tc.raw)), LoadOptions{Env: "production"})
			if err == nil {
				t.Fatalf("expected schema validation error")
			}
			if !strings.Contains(err.Error(), tc.errContains) {
				t.Fatalf("expected error containing %q, got: %v", tc.errContains, err)
			}
		})
	}
}

func TestLoadFromYAML_UnknownFieldsAreForwardCompatible(t *testing.T) {
	t.Parallel()

	result, err := LoadFromYAML([]byte(strings.TrimSpace(`
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
  unknownRootMeta: true
unknownRootField: "allowed"
spec:
  unknownSpecField:
    enabled: true
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
      unknownEnvironmentField: "ok"
  versions:
    worker:
      value: "0.1.0"
      unknownVersionField: "future"
  services:
    - name: worker
      unknownServiceField: true
`)), LoadOptions{Env: "production"})
	if err != nil {
		t.Fatalf("expected unknown fields to be ignored for forward compatibility, got: %v", err)
	}

	if got, want := result.Stack.Metadata.Name, "demo"; got != want {
		t.Fatalf("unexpected metadata.name: got %q want %q", got, want)
	}
	if got, want := len(result.Stack.Spec.Services), 1; got != want {
		t.Fatalf("unexpected services count: got %d want %d", got, want)
	}
	if got, want := result.Stack.Spec.Services[0].Name, "worker"; got != want {
		t.Fatalf("unexpected service name: got %q want %q", got, want)
	}
	if got, want := result.Stack.Spec.Versions["worker"].Value, "0.1.0"; got != want {
		t.Fatalf("unexpected version value: got %q want %q", got, want)
	}
}

func TestLoad_VersionsRenderTagTemplates(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "services.yaml")
	writeFile(t, path, `
apiVersion: codex-k8s.dev/v1alpha1
kind: ServiceStack
metadata:
  name: demo
spec:
  versions:
    api-gateway:
      value: "0.2.1"
      bumpOn:
        - ./services/external/api-gateway
    worker:
      value: "0.4.0"
  environments:
    production:
      namespaceTemplate: "{{ .Project }}-prod"
  images:
    api-gateway:
      type: build
      repository: registry.local/demo/api-gateway
      tagTemplate: '{{ .Env }}-{{ index .Versions "api-gateway" }}'
    worker:
      type: build
      repository: registry.local/demo/worker
      tagTemplate: '{{ .Env }}-{{ index .Versions "worker" }}'
`)

	result, err := Load(path, LoadOptions{Env: "production"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if got, want := result.Context.Versions["api-gateway"], "0.2.1"; got != want {
		t.Fatalf("unexpected api-gateway version in context: got %q want %q", got, want)
	}
	if got, want := result.Context.Versions["worker"], "0.4.0"; got != want {
		t.Fatalf("unexpected worker version in context: got %q want %q", got, want)
	}

	apiVersion := result.Stack.Spec.Versions["api-gateway"]
	if got, want := apiVersion.Value, "0.2.1"; got != want {
		t.Fatalf("unexpected api-gateway version value: got %q want %q", got, want)
	}
	if got, want := strings.Join(apiVersion.BumpOn, ","), "services/external/api-gateway"; got != want {
		t.Fatalf("unexpected api-gateway bumpOn: got %q want %q", got, want)
	}

	image := result.Stack.Spec.Images["api-gateway"]
	if got, want := image.TagTemplate, "production-0.2.1"; got != want {
		t.Fatalf("unexpected api-gateway tagTemplate: got %q want %q", got, want)
	}
	workerImage := result.Stack.Spec.Images["worker"]
	if got, want := workerImage.TagTemplate, "production-0.4.0"; got != want {
		t.Fatalf("unexpected worker tagTemplate: got %q want %q", got, want)
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
