package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const defaultGitHubWebhookEvents = "push,pull_request,issues,issue_comment,pull_request_review,pull_request_review_comment"

func loadEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(strings.TrimSpace(path))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("invalid env-file line %q", line)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		value = strings.TrimSpace(value)
		value = strings.Trim(value, "'")
		value = strings.Trim(value, "\"")
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func applyRuntimeDeployEnvDefaults(values map[string]string) {
	setEnvDefault(values, "KODEX_PRODUCTION_NAMESPACE", "kodex-prod")
	setEnvDefault(values, "KODEX_PRODUCTION_DOMAIN", "platform.kodex.works")
	setEnvDefault(values, "KODEX_INTERNAL_REGISTRY_SERVICE", "kodex-registry")
	setEnvDefault(values, "KODEX_INTERNAL_REGISTRY_PORT", "5000")
	setEnvDefault(values, "KODEX_INTERNAL_REGISTRY_STORAGE_SIZE", "20Gi")
	setEnvDefault(values, "KODEX_INTERNAL_REGISTRY_HOST", "127.0.0.1:"+strings.TrimSpace(values["KODEX_INTERNAL_REGISTRY_PORT"]))
	setEnvDefault(values, "KODEX_RUNTIME_DEPLOY_FIELD_MANAGER", "kodex-control-plane")
	setEnvDefault(values, "KODEX_GITHUB_WEBHOOK_EVENTS", defaultGitHubWebhookEvents)
	setEnvDefault(values, "KODEX_GITHUB_WEBHOOK_URL", resolveWebhookURL(values))
}

func setEnvDefault(values map[string]string, key string, fallback string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	if strings.TrimSpace(values[key]) != "" {
		return
	}
	values[key] = fallback
}

func resolveWebhookURL(values map[string]string) string {
	if explicit := strings.TrimSpace(values["KODEX_GITHUB_WEBHOOK_URL"]); explicit != "" {
		return explicit
	}
	domain := strings.TrimSpace(values["KODEX_PRODUCTION_DOMAIN"])
	if domain == "" {
		return ""
	}
	return "https://" + domain + "/api/v1/webhooks/github"
}
