package servicescfg

import (
	"fmt"
	"strings"
)

const (
	secretTemplateTokenKey      = "{key}"
	secretTemplateTokenSuffix   = "{suffix}"
	secretTemplateTokenEnv      = "{env}"
	secretTemplateTokenEnvUpper = "{env_upper}"
)

// SecretResolver resolves environment-scoped override key names.
type SecretResolver struct {
	envAliases   map[string]string
	keyOverrides map[string]map[string]string
	patterns     []resolvedSecretPattern
}

type resolvedSecretPattern struct {
	sourcePrefix     string
	excludePrefixes  []string
	excludeSuffixes  []string
	environments     map[string]struct{}
	overrideTemplate string
}

// DefaultSecretResolution returns fallback resolution policy used when services.yaml omits spec.secretResolution.
func DefaultSecretResolution() SecretResolution {
	return SecretResolution{
		EnvironmentAliases: map[string][]string{
			"production": []string{"prod"},
		},
		Patterns: []SecretOverridePattern{
			{
				SourcePrefix:     "CODEXK8S_",
				ExcludePrefixes:  []string{"CODEXK8S_AI_", "CODEXK8S_PRODUCTION_"},
				Environments:     []string{"production", "ai"},
				OverrideTemplate: "CODEXK8S_{env_upper}_{suffix}",
			},
		},
	}
}

// NewSecretResolver builds resolver from services stack and defaults.
func NewSecretResolver(stack *Stack) SecretResolver {
	if stack == nil {
		return newSecretResolverFromConfig(SecretResolution{})
	}
	return newSecretResolverFromConfig(stack.Spec.SecretResolution)
}

func newSecretResolverFromConfig(cfg SecretResolution) SecretResolver {
	merged := mergeSecretResolutionWithDefaults(cfg)
	aliases := buildEnvironmentAliasMap(merged.EnvironmentAliases)
	overrides := buildResolvedKeyOverrides(merged.KeyOverrides, aliases)
	patterns := buildResolvedPatterns(merged.Patterns, aliases)
	return SecretResolver{
		envAliases:   aliases,
		keyOverrides: overrides,
		patterns:     patterns,
	}
}

func mergeSecretResolutionWithDefaults(cfg SecretResolution) SecretResolution {
	base := DefaultSecretResolution()
	merged := SecretResolution{
		EnvironmentAliases: make(map[string][]string, len(base.EnvironmentAliases)+len(cfg.EnvironmentAliases)),
		KeyOverrides:       make([]SecretKeyOverrideRule, 0, len(cfg.KeyOverrides)),
		Patterns:           make([]SecretOverridePattern, 0, len(cfg.Patterns)+len(base.Patterns)),
	}

	for envName, aliases := range base.EnvironmentAliases {
		merged.EnvironmentAliases[envName] = append([]string(nil), aliases...)
	}
	for envName, aliases := range cfg.EnvironmentAliases {
		trimmedEnv := normalizeEnvName(envName)
		if trimmedEnv == "" {
			continue
		}
		existing := merged.EnvironmentAliases[trimmedEnv]
		seen := make(map[string]struct{}, len(existing))
		for _, value := range existing {
			seen[normalizeEnvName(value)] = struct{}{}
		}
		for _, alias := range aliases {
			trimmedAlias := normalizeEnvName(alias)
			if trimmedAlias == "" {
				continue
			}
			if _, ok := seen[trimmedAlias]; ok {
				continue
			}
			existing = append(existing, trimmedAlias)
			seen[trimmedAlias] = struct{}{}
		}
		merged.EnvironmentAliases[trimmedEnv] = existing
	}

	merged.KeyOverrides = append(merged.KeyOverrides, cfg.KeyOverrides...)
	merged.Patterns = append(merged.Patterns, cfg.Patterns...)
	merged.Patterns = append(merged.Patterns, base.Patterns...)
	return merged
}

func buildEnvironmentAliasMap(input map[string][]string) map[string]string {
	out := map[string]string{
		"production": "production",
		"ai":         "ai",
	}
	for envName, aliases := range input {
		trimmedEnv := normalizeEnvName(envName)
		if trimmedEnv == "" {
			continue
		}
		out[trimmedEnv] = trimmedEnv
		for _, alias := range aliases {
			trimmedAlias := normalizeEnvName(alias)
			if trimmedAlias == "" {
				continue
			}
			out[trimmedAlias] = trimmedEnv
		}
	}
	return out
}

func buildResolvedKeyOverrides(rules []SecretKeyOverrideRule, aliases map[string]string) map[string]map[string]string {
	out := make(map[string]map[string]string)
	for _, rule := range rules {
		sourceKey := strings.TrimSpace(rule.SourceKey)
		if sourceKey == "" {
			continue
		}
		if _, ok := out[sourceKey]; !ok {
			out[sourceKey] = make(map[string]string)
		}
		for envName, overrideKey := range rule.OverrideKeys {
			canonical := canonicalEnvName(normalizeEnvName(envName), aliases)
			if canonical == "" {
				continue
			}
			trimmedOverride := strings.TrimSpace(overrideKey)
			if trimmedOverride == "" {
				continue
			}
			out[sourceKey][canonical] = trimmedOverride
		}
	}
	return out
}

func buildResolvedPatterns(patterns []SecretOverridePattern, aliases map[string]string) []resolvedSecretPattern {
	if len(patterns) == 0 {
		return nil
	}
	out := make([]resolvedSecretPattern, 0, len(patterns))
	for _, item := range patterns {
		template := strings.TrimSpace(item.OverrideTemplate)
		if template == "" {
			continue
		}

		resolved := resolvedSecretPattern{
			sourcePrefix:     strings.TrimSpace(item.SourcePrefix),
			overrideTemplate: template,
		}
		for _, value := range item.ExcludePrefixes {
			trimmed := strings.TrimSpace(value)
			if trimmed != "" {
				resolved.excludePrefixes = append(resolved.excludePrefixes, trimmed)
			}
		}
		for _, value := range item.ExcludeSuffixes {
			trimmed := strings.TrimSpace(value)
			if trimmed != "" {
				resolved.excludeSuffixes = append(resolved.excludeSuffixes, trimmed)
			}
		}
		if len(item.Environments) > 0 {
			resolved.environments = make(map[string]struct{}, len(item.Environments))
			for _, env := range item.Environments {
				canonical := canonicalEnvName(normalizeEnvName(env), aliases)
				if canonical == "" {
					continue
				}
				resolved.environments[canonical] = struct{}{}
			}
		}
		out = append(out, resolved)
	}
	return out
}

// CanonicalEnvironment resolves environment aliases (e.g. prod -> production).
func (r SecretResolver) CanonicalEnvironment(envName string) string {
	return canonicalEnvName(normalizeEnvName(envName), r.envAliases)
}

// ResolveOverrideKey returns override key for source key in target environment.
func (r SecretResolver) ResolveOverrideKey(envName string, sourceKey string) (string, bool) {
	canonicalEnv := r.CanonicalEnvironment(envName)
	sourceKey = strings.TrimSpace(sourceKey)
	if canonicalEnv == "" || sourceKey == "" {
		return "", false
	}

	if byEnv, ok := r.keyOverrides[sourceKey]; ok {
		if key := strings.TrimSpace(byEnv[canonicalEnv]); key != "" {
			return key, true
		}
	}

	for _, pattern := range r.patterns {
		if pattern.sourcePrefix != "" && !strings.HasPrefix(sourceKey, pattern.sourcePrefix) {
			continue
		}
		if len(pattern.environments) > 0 {
			if _, ok := pattern.environments[canonicalEnv]; !ok {
				continue
			}
		}
		skip := false
		for _, value := range pattern.excludePrefixes {
			if strings.HasPrefix(sourceKey, value) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		for _, value := range pattern.excludeSuffixes {
			if strings.HasSuffix(sourceKey, value) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		suffix := sourceKey
		if pattern.sourcePrefix != "" {
			suffix = strings.TrimPrefix(sourceKey, pattern.sourcePrefix)
		}
		key := renderSecretOverrideTemplate(pattern.overrideTemplate, sourceKey, suffix, canonicalEnv)
		if key == "" || key == sourceKey {
			continue
		}
		return key, true
	}
	return "", false
}

// ResolveValueFromMap resolves value by deterministic chain: env override -> base key.
func (r SecretResolver) ResolveValueFromMap(values map[string]string, envName string, sourceKey string) (string, string, bool) {
	if len(values) == 0 {
		return "", "", false
	}
	if overrideKey, ok := r.ResolveOverrideKey(envName, sourceKey); ok {
		if value := strings.TrimSpace(values[overrideKey]); value != "" {
			return value, overrideKey, true
		}
	}
	sourceKey = strings.TrimSpace(sourceKey)
	if sourceKey == "" {
		return "", "", false
	}
	if value := strings.TrimSpace(values[sourceKey]); value != "" {
		return value, sourceKey, true
	}
	return "", "", false
}

func renderSecretOverrideTemplate(tmpl string, key string, suffix string, envName string) string {
	tmpl = strings.TrimSpace(tmpl)
	if tmpl == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		secretTemplateTokenKey, key,
		secretTemplateTokenSuffix, suffix,
		secretTemplateTokenEnv, envName,
		secretTemplateTokenEnvUpper, strings.ToUpper(envName),
	)
	return strings.TrimSpace(replacer.Replace(tmpl))
}

func normalizeEnvName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func canonicalEnvName(value string, aliases map[string]string) string {
	if value == "" {
		return ""
	}
	if aliases == nil {
		return value
	}
	if canonical, ok := aliases[value]; ok {
		return canonical
	}
	return value
}

func validateSecretResolution(cfg SecretResolution) error {
	for envName, aliases := range cfg.EnvironmentAliases {
		if normalizeEnvName(envName) == "" {
			return fmt.Errorf("spec.secretResolution.environmentAliases contains empty environment key")
		}
		for idx, alias := range aliases {
			if normalizeEnvName(alias) == "" {
				return fmt.Errorf("spec.secretResolution.environmentAliases[%q][%d] is empty", envName, idx)
			}
		}
	}

	seenSources := make(map[string]struct{}, len(cfg.KeyOverrides))
	for idx, rule := range cfg.KeyOverrides {
		sourceKey := strings.TrimSpace(rule.SourceKey)
		if sourceKey == "" {
			return fmt.Errorf("spec.secretResolution.keyOverrides[%d].sourceKey is required", idx)
		}
		if _, exists := seenSources[sourceKey]; exists {
			return fmt.Errorf("spec.secretResolution.keyOverrides has duplicate sourceKey %q", sourceKey)
		}
		seenSources[sourceKey] = struct{}{}
		for envName, overrideKey := range rule.OverrideKeys {
			if normalizeEnvName(envName) == "" {
				return fmt.Errorf("spec.secretResolution.keyOverrides[%d].overrideKeys contains empty env key", idx)
			}
			if strings.TrimSpace(overrideKey) == "" {
				return fmt.Errorf("spec.secretResolution.keyOverrides[%d].overrideKeys[%q] is empty", idx, envName)
			}
		}
	}

	for idx, pattern := range cfg.Patterns {
		if strings.TrimSpace(pattern.OverrideTemplate) == "" {
			return fmt.Errorf("spec.secretResolution.patterns[%d].overrideTemplate is required", idx)
		}
		for envIdx, envName := range pattern.Environments {
			if normalizeEnvName(envName) == "" {
				return fmt.Errorf("spec.secretResolution.patterns[%d].environments[%d] is empty", idx, envIdx)
			}
		}
	}

	return nil
}
