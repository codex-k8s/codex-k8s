package servicescfg

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
)

const (
	templateNameNamespace = "namespace"
	maxSHAForTemplate     = 12
)

var definePattern = regexp.MustCompile(`\{\{\s*define\s+"([^"]+)"\s*\}\}`)

// RenderContext represents data shared for services/deploy template rendering.
type RenderContext struct {
	Env        string
	Namespace  string
	Project    string
	ProjectDir string
	Slot       int
	Now        time.Time
	EnvMap     map[string]string
	Vars       map[string]string
}

type partialTemplate struct {
	path string
	body string
}

// Renderer renders templates with declared partials.
type Renderer struct {
	rootDir             string
	partials            []partialTemplate
	partialDefineToFile map[string]string
	ctx                 RenderContext
}

// NewRenderer creates renderer from config + root dir.
func NewRenderer(cfg *Config, rootDir string, ctx RenderContext) (*Renderer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("services config is nil")
	}
	if rootDir == "" {
		return nil, fmt.Errorf("renderer root directory is empty")
	}

	partialPaths, err := resolvePartials(rootDir, cfg.Templates.Partials)
	if err != nil {
		return nil, err
	}

	partials := make([]partialTemplate, 0, len(partialPaths))
	defineToFile := make(map[string]string, 16)
	for _, path := range partialPaths {
		bodyBytes, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read partial template %q: %w", path, err)
		}
		body := string(bodyBytes)
		for _, name := range extractDefinedTemplates(body) {
			if prevPath, exists := defineToFile[name]; exists {
				return nil, fmt.Errorf("partial template define conflict: %q in %q and %q", name, prevPath, path)
			}
			defineToFile[name] = path
		}
		partials = append(partials, partialTemplate{
			path: path,
			body: body,
		})
	}

	if ctx.Now.IsZero() {
		ctx.Now = time.Now().UTC()
	}
	if ctx.EnvMap == nil {
		ctx.EnvMap = map[string]string{}
	}
	if ctx.Vars == nil {
		ctx.Vars = map[string]string{}
	}

	return &Renderer{
		rootDir:             rootDir,
		partials:            partials,
		partialDefineToFile: defineToFile,
		ctx:                 ctx,
	}, nil
}

// RenderFile renders template file from project root.
func (r *Renderer) RenderFile(relativePath string, data any) ([]byte, error) {
	path := filepath.Join(r.rootDir, relativePath)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template file %q: %w", path, err)
	}
	rendered, err := r.Render(relativePath, raw, data)
	if err != nil {
		return nil, err
	}
	return rendered, nil
}

// Render renders raw template content with configured partials.
func (r *Renderer) Render(name string, raw []byte, data any) ([]byte, error) {
	expanded := expandEnvReferences(string(raw), r.ctx.EnvMap)
	rawDefines := extractDefinedTemplates(expanded)
	for _, defined := range rawDefines {
		if sourcePath, exists := r.partialDefineToFile[defined]; exists {
			return nil, fmt.Errorf("template define conflict for %q between render target %q and partial %q", defined, name, sourcePath)
		}
	}
	if hasDuplicate(rawDefines) {
		return nil, fmt.Errorf("template define conflict inside %q", name)
	}

	var rootTemplate *template.Template
	funcMap := r.buildFuncMap(&rootTemplate)
	rootTemplate = template.New(name).Option("missingkey=error").Funcs(funcMap)

	for _, part := range r.partials {
		if _, err := rootTemplate.Parse(part.body); err != nil {
			return nil, fmt.Errorf("parse partial template %q: %w", part.path, err)
		}
	}
	if _, err := rootTemplate.Parse(expanded); err != nil {
		return nil, fmt.Errorf("parse template %q: %w", name, err)
	}

	var out bytes.Buffer
	if err := rootTemplate.Execute(&out, data); err != nil {
		return nil, fmt.Errorf("render template %q: %w", name, err)
	}
	return out.Bytes(), nil
}

func (r *Renderer) buildFuncMap(rootTemplate **template.Template) template.FuncMap {
	return template.FuncMap{
		"default":    funcDefault,
		"toLower":    strings.ToLower,
		"slug":       funcSlug,
		"truncSHA":   funcTruncSHA,
		"envOr":      funcEnvOr(r.ctx.EnvMap),
		"ternary":    funcTernary,
		"now":        func() time.Time { return r.ctx.Now },
		"join":       strings.Join,
		"indent":     funcIndent,
		"trimPrefix": strings.TrimPrefix,
		"include": func(templateName string, data any) (string, error) {
			if rootTemplate == nil || *rootTemplate == nil {
				return "", fmt.Errorf("include called before template initialization")
			}
			var out bytes.Buffer
			if err := (*rootTemplate).ExecuteTemplate(&out, templateName, data); err != nil {
				return "", err
			}
			return out.String(), nil
		},
	}
}

func resolvePartials(rootDir string, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, nil
	}
	collected := make(map[string]struct{}, len(patterns))
	for _, pattern := range patterns {
		trimmed := strings.TrimSpace(pattern)
		if trimmed == "" {
			continue
		}
		glob := trimmed
		if !filepath.IsAbs(glob) {
			glob = filepath.Join(rootDir, glob)
		}
		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, fmt.Errorf("resolve partial glob %q: %w", pattern, err)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("partial glob %q did not match files", pattern)
		}
		for _, match := range matches {
			if stat, statErr := os.Stat(match); statErr != nil || stat.IsDir() {
				continue
			}
			collected[match] = struct{}{}
		}
	}
	paths := make([]string, 0, len(collected))
	for path := range collected {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

func extractDefinedTemplates(raw string) []string {
	matches := definePattern.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return nil
	}
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		name := strings.TrimSpace(match[1])
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func hasDuplicate(items []string) bool {
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if _, exists := seen[item]; exists {
			return true
		}
		seen[item] = struct{}{}
	}
	return false
}

func expandEnvReferences(input string, envMap map[string]string) string {
	return os.Expand(input, func(key string) string {
		if value, ok := envMap[key]; ok {
			return value
		}
		return "${" + key + "}"
	})
}

func funcDefault(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func funcSlug(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = strings.ReplaceAll(normalized, "_", "-")
	return normalized
}

func funcTruncSHA(value string) string {
	if len(value) <= maxSHAForTemplate {
		return value
	}
	return value[:maxSHAForTemplate]
}

func funcEnvOr(envMap map[string]string) func(string, string) string {
	return func(key string, fallback string) string {
		if value, ok := envMap[key]; ok && strings.TrimSpace(value) != "" {
			return value
		}
		return fallback
	}
}

func funcTernary(condition bool, left any, right any) any {
	if condition {
		return left
	}
	return right
}

func funcIndent(spaces int, value string) string {
	if value == "" {
		return ""
	}
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(strings.TrimRight(value, "\n"), "\n")
	for idx := range lines {
		lines[idx] = prefix + lines[idx]
	}
	return strings.Join(lines, "\n")
}

// ResolveNamespace resolves namespace pattern using renderer rules.
func (r *Renderer) ResolveNamespace(rawPattern string) (string, error) {
	rendered, err := r.Render(templateNameNamespace, []byte(rawPattern), r.ctx)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(rendered)), nil
}
