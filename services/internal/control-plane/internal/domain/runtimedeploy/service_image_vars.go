package runtimedeploy

import (
	"sort"
	"strings"
	"unicode"

	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
)

func applyStackImageVars(vars map[string]string, stack *servicescfg.Stack) {
	if vars == nil || stack == nil || len(stack.Spec.Images) == 0 {
		return
	}

	names := make([]string, 0, len(stack.Spec.Images))
	for name := range stack.Spec.Images {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		image := stack.Spec.Images[name]
		envKey := imageEnvVar(name)
		if envKey == "" {
			continue
		}
		imageRef := resolveStackImageRef(image)
		if imageRef == "" {
			continue
		}
		vars[envKey] = imageRef
	}
}

func resolveStackImageRef(image servicescfg.Image) string {
	local := strings.TrimSpace(image.Local)
	if local != "" {
		return local
	}
	repository := strings.TrimSpace(image.Repository)
	if repository == "" {
		return ""
	}
	tag := strings.TrimSpace(image.TagTemplate)
	if tag == "" {
		tag = "latest"
	}
	return repository + ":" + tag
}

func imageEnvVar(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	normalized := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return unicode.ToUpper(r)
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		default:
			return '_'
		}
	}, trimmed)
	normalized = strings.Trim(normalized, "_")
	if normalized == "" {
		return ""
	}
	return "CODEXK8S_" + normalized + "_IMAGE"
}
