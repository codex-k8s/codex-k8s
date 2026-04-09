package servicescfg

import "testing"

func TestSecretResolver_DefaultPrefixPattern(t *testing.T) {
	t.Parallel()

	resolver := NewSecretResolver(nil)
	if got := resolver.CanonicalEnvironment("prod"); got != "production" {
		t.Fatalf("expected production alias for prod, got %q", got)
	}

	overrideKey, ok := resolver.ResolveOverrideKey("ai", "KODEX_OPENAI_API_KEY")
	if !ok {
		t.Fatalf("expected override key for KODEX_OPENAI_API_KEY")
	}
	if want := "KODEX_AI_OPENAI_API_KEY"; overrideKey != want {
		t.Fatalf("unexpected override key: got %q want %q", overrideKey, want)
	}

	if _, ok := resolver.ResolveOverrideKey("ai", "KODEX_AI_DOMAIN"); ok {
		t.Fatalf("did not expect override for already environment-scoped KODEX_AI_DOMAIN")
	}
}

func TestSecretResolver_KeyOverridesAndTemplatePattern(t *testing.T) {
	t.Parallel()

	stack := &Stack{
		Spec: Spec{
			SecretResolution: SecretResolution{
				KeyOverrides: []SecretKeyOverrideRule{
					{
						SourceKey: "KODEX_GITHUB_OAUTH_CLIENT_ID",
						OverrideKeys: map[string]string{
							"ai": "KODEX_GITHUB_OAUTH_CLIENT_ID_AI",
						},
					},
				},
				Patterns: []SecretOverridePattern{
					{
						SourcePrefix:     "KODEX_",
						ExcludeSuffixes:  []string{"_AI"},
						Environments:     []string{"ai"},
						OverrideTemplate: "{key}_{env_upper}",
					},
				},
			},
		},
	}

	resolver := NewSecretResolver(stack)
	oauthOverride, ok := resolver.ResolveOverrideKey("ai", "KODEX_GITHUB_OAUTH_CLIENT_ID")
	if !ok {
		t.Fatalf("expected oauth override key")
	}
	if want := "KODEX_GITHUB_OAUTH_CLIENT_ID_AI"; oauthOverride != want {
		t.Fatalf("unexpected oauth override key: got %q want %q", oauthOverride, want)
	}

	openAIOverride, ok := resolver.ResolveOverrideKey("ai", "KODEX_OPENAI_API_KEY")
	if !ok {
		t.Fatalf("expected pattern-based override key")
	}
	if want := "KODEX_OPENAI_API_KEY_AI"; openAIOverride != want {
		t.Fatalf("unexpected pattern-based override key: got %q want %q", openAIOverride, want)
	}

	values := map[string]string{
		"KODEX_GITHUB_OAUTH_CLIENT_ID":    "base",
		"KODEX_GITHUB_OAUTH_CLIENT_ID_AI": "ai",
	}
	value, sourceKey, ok := resolver.ResolveValueFromMap(values, "ai", "KODEX_GITHUB_OAUTH_CLIENT_ID")
	if !ok {
		t.Fatalf("expected resolved value")
	}
	if value != "ai" || sourceKey != "KODEX_GITHUB_OAUTH_CLIENT_ID_AI" {
		t.Fatalf("unexpected resolved value/source: value=%q source=%q", value, sourceKey)
	}
}

func TestValidateSecretResolution_DuplicateSourceKey(t *testing.T) {
	t.Parallel()

	err := validateSecretResolution(SecretResolution{
		KeyOverrides: []SecretKeyOverrideRule{
			{SourceKey: "KODEX_OPENAI_API_KEY", OverrideKeys: map[string]string{"ai": "KODEX_AI_OPENAI_API_KEY"}},
			{SourceKey: "KODEX_OPENAI_API_KEY", OverrideKeys: map[string]string{"production": "KODEX_PRODUCTION_OPENAI_API_KEY"}},
		},
	})
	if err == nil {
		t.Fatalf("expected validation error for duplicate sourceKey")
	}
}
