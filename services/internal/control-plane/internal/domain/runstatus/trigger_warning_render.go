package runstatus

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:embed i18n/trigger_warning.yaml
var triggerWarningI18nFS embed.FS

const (
	triggerWarningI18nPath = "i18n/trigger_warning.yaml"
)

type triggerWarningRenderParams struct {
	Locale            string
	ThreadKind        string
	ReasonCode        TriggerWarningReasonCode
	ConflictingLabels []string
	SuggestedLabels   []string
}

type triggerWarningTemplateContext struct {
	ReasonCode        TriggerWarningReasonCode
	ReasonText        string
	HintText          string
	IsPullRequest     bool
	ConflictingLabels []string
	SuggestedLabels   []string
}

type triggerWarningLocaleStrings struct {
	Template     string                      `yaml:"template"`
	Reasons      map[string]triggerWarningUI `yaml:"reasons"`
	SenderPrefix triggerWarningUI            `yaml:"sender_prefix"`
	Default      triggerWarningUI            `yaml:"default"`
}

type triggerWarningUI struct {
	Reason string `yaml:"reason"`
	Hint   string `yaml:"hint"`
}

type triggerWarningI18nFile struct {
	Locales map[string]triggerWarningLocaleStrings `yaml:"locales"`
}

type triggerWarningLocaleCatalog struct {
	Template     *template.Template
	Reasons      map[TriggerWarningReasonCode]triggerWarningUI
	SenderPrefix triggerWarningUI
	Default      triggerWarningUI
}

var (
	triggerWarningCatalogOnce sync.Once
	triggerWarningCatalogData map[string]triggerWarningLocaleCatalog
	triggerWarningCatalogErr  error
)

func renderTriggerWarningCommentBody(params triggerWarningRenderParams) (string, error) {
	locale := normalizeLocale(params.Locale, localeEN)
	reasonCode := TriggerWarningReasonCode(strings.TrimSpace(string(params.ReasonCode)))
	if reasonCode == "" {
		return "", fmt.Errorf("reason code is required")
	}
	threadKind := normalizeCommentTargetKind(params.ThreadKind)
	if threadKind == "" {
		return "", fmt.Errorf("thread kind is required")
	}

	catalog, err := loadTriggerWarningCatalog()
	if err != nil {
		return "", err
	}
	localeCatalog, ok := catalog[locale]
	if !ok {
		localeCatalog, ok = catalog[localeEN]
	}
	if !ok {
		return "", fmt.Errorf("trigger warning i18n locale catalog is empty")
	}

	entry, ok := localeCatalog.Reasons[reasonCode]
	if !ok {
		if strings.HasPrefix(string(reasonCode), "sender_") {
			entry = localeCatalog.SenderPrefix
		} else {
			entry = localeCatalog.Default
		}
	}

	var out bytes.Buffer
	if err := localeCatalog.Template.Execute(&out, triggerWarningTemplateContext{
		ReasonCode:        reasonCode,
		ReasonText:        strings.TrimSpace(entry.Reason),
		HintText:          strings.TrimSpace(entry.Hint),
		IsPullRequest:     threadKind == commentTargetKindPullRequest,
		ConflictingLabels: normalizeConflictLabels(params.ConflictingLabels),
		SuggestedLabels:   normalizeConflictLabels(params.SuggestedLabels),
	}); err != nil {
		return "", fmt.Errorf("render trigger warning template: %w", err)
	}
	return strings.TrimSpace(out.String()) + "\n", nil
}

func loadTriggerWarningCatalog() (map[string]triggerWarningLocaleCatalog, error) {
	triggerWarningCatalogOnce.Do(func() {
		raw, err := triggerWarningI18nFS.ReadFile(triggerWarningI18nPath)
		if err != nil {
			triggerWarningCatalogErr = fmt.Errorf("read trigger warning i18n file: %w", err)
			return
		}

		var parsed triggerWarningI18nFile
		if err := yaml.Unmarshal(raw, &parsed); err != nil {
			triggerWarningCatalogErr = fmt.Errorf("decode trigger warning i18n yaml: %w", err)
			return
		}

		catalog := make(map[string]triggerWarningLocaleCatalog, len(parsed.Locales))
		for locale, localeData := range parsed.Locales {
			normalizedLocale := normalizeLocale(locale, localeEN)
			templateBody := strings.TrimSpace(localeData.Template)
			if templateBody == "" {
				triggerWarningCatalogErr = fmt.Errorf("trigger warning i18n template is empty for locale %q", normalizedLocale)
				return
			}
			tpl, err := template.New("trigger-warning-" + normalizedLocale).Option("missingkey=error").Parse(templateBody)
			if err != nil {
				triggerWarningCatalogErr = fmt.Errorf("parse trigger warning template for locale %q: %w", normalizedLocale, err)
				return
			}

			reasons := make(map[TriggerWarningReasonCode]triggerWarningUI, len(localeData.Reasons))
			for rawCode, value := range localeData.Reasons {
				code := TriggerWarningReasonCode(strings.TrimSpace(rawCode))
				if code == "" {
					continue
				}
				reasons[code] = triggerWarningUI{
					Reason: strings.TrimSpace(value.Reason),
					Hint:   strings.TrimSpace(value.Hint),
				}
			}

			catalog[normalizedLocale] = triggerWarningLocaleCatalog{
				Template: tpl,
				Reasons:  reasons,
				SenderPrefix: triggerWarningUI{
					Reason: strings.TrimSpace(localeData.SenderPrefix.Reason),
					Hint:   strings.TrimSpace(localeData.SenderPrefix.Hint),
				},
				Default: triggerWarningUI{
					Reason: strings.TrimSpace(localeData.Default.Reason),
					Hint:   strings.TrimSpace(localeData.Default.Hint),
				},
			}
		}

		triggerWarningCatalogData = catalog
	})
	if triggerWarningCatalogErr != nil {
		return nil, triggerWarningCatalogErr
	}
	return triggerWarningCatalogData, nil
}
