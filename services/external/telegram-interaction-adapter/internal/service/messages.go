package service

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

//go:embed messages_en.json
var messageBundleEN []byte

//go:embed messages_ru.json
var messageBundleRU []byte

type messageRenderer struct {
	bundles map[string]map[string]*template.Template
}

func newMessageRenderer() (*messageRenderer, error) {
	renderer := &messageRenderer{
		bundles: map[string]map[string]*template.Template{},
	}
	for locale, payload := range map[string][]byte{
		"en": messageBundleEN,
		"ru": messageBundleRU,
	} {
		templates := map[string]string{}
		if err := json.Unmarshal(payload, &templates); err != nil {
			return nil, fmt.Errorf("unmarshal %s message bundle: %w", locale, err)
		}
		renderer.bundles[locale] = map[string]*template.Template{}
		for key, value := range templates {
			tmpl, err := template.New(locale + ":" + key).Parse(value)
			if err != nil {
				return nil, fmt.Errorf("parse %s template %s: %w", locale, key, err)
			}
			renderer.bundles[locale][key] = tmpl
		}
	}
	return renderer, nil
}

func (r *messageRenderer) Render(locale string, key string, data any) string {
	selectedLocale := "en"
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(locale)), "ru") {
		selectedLocale = "ru"
	}
	bundle := r.bundles[selectedLocale]
	tmpl, ok := bundle[key]
	if !ok {
		tmpl = r.bundles["en"][key]
	}
	if tmpl == nil {
		return ""
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return strings.TrimSpace(buf.String())
}
