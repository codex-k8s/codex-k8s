package service

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed messages_en.tmpl
var messageBundleEN []byte

//go:embed messages_ru.tmpl
var messageBundleRU []byte

type messageRenderer struct {
	bundles map[string]*template.Template
}

func newMessageRenderer() (*messageRenderer, error) {
	renderer := &messageRenderer{
		bundles: map[string]*template.Template{},
	}
	for locale, payload := range map[string][]byte{
		"en": messageBundleEN,
		"ru": messageBundleRU,
	} {
		tmpl, err := template.New(locale).Parse(string(payload))
		if err != nil {
			return nil, fmt.Errorf("parse %s message bundle: %w", locale, err)
		}
		renderer.bundles[locale] = tmpl
	}
	return renderer, nil
}

func (r *messageRenderer) Render(locale string, key string, data any) string {
	selectedLocale := "en"
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(locale)), "ru") {
		selectedLocale = "ru"
	}
	tmpl := r.bundles[selectedLocale]
	if tmpl == nil {
		tmpl = r.bundles["en"]
	}
	if tmpl == nil {
		return ""
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, key, data); err != nil {
		if selectedLocale != "en" && r.bundles["en"] != nil {
			buf.Reset()
			if err := r.bundles["en"].ExecuteTemplate(&buf, key, data); err != nil {
				return ""
			}
			return strings.TrimSpace(buf.String())
		}
		return ""
	}
	return strings.TrimSpace(buf.String())
}
