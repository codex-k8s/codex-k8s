package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	githubCredentialURLPattern = regexp.MustCompile(`https://[^/\s:@]+:[^@\s/]+@github\.com`)
	authorizationBearerPattern = regexp.MustCompile(`(?i)(authorization:\s*bearer\s+)[^\s"'` + "`" + `]+`)
	githubPATPattern           = regexp.MustCompile(`github_pat_[A-Za-z0-9_]+`)
	unsafePathComponentPattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
)

func (s *Service) configureGitAuthEnvironment(workspaceDir string) (func(), error) {
	scriptPath := filepath.Join(workspaceDir, fmt.Sprintf("git-askpass-%s.sh", sanitizePathComponent(s.cfg.RunID)))
	scriptContent := `#!/bin/sh
case "$1" in
  *Username*) printf '%s\n' "$CODEXK8S_GIT_BOT_USERNAME" ;;
  *Password*) printf '%s\n' "$CODEXK8S_GIT_BOT_TOKEN" ;;
  *) printf '\n' ;;
esac
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o700); err != nil {
		return nil, fmt.Errorf("write git askpass script: %w", err)
	}

	previousAskPass, hadAskPass := os.LookupEnv(envGitAskPass)
	previousPrompt, hadPrompt := os.LookupEnv(envGitTerminalPrompt)
	previousAskPassRequire, hadAskPassRequire := os.LookupEnv(envGitAskPassRequire)

	if err := os.Setenv(envGitAskPass, scriptPath); err != nil {
		return nil, fmt.Errorf("set %s: %w", envGitAskPass, err)
	}
	if err := os.Setenv(envGitTerminalPrompt, "0"); err != nil {
		return nil, fmt.Errorf("set %s: %w", envGitTerminalPrompt, err)
	}
	if err := os.Setenv(envGitAskPassRequire, gitAskPassRequireForce); err != nil {
		return nil, fmt.Errorf("set %s: %w", envGitAskPassRequire, err)
	}

	cleanup := func() {
		restoreEnvVariable(envGitAskPass, previousAskPass, hadAskPass)
		restoreEnvVariable(envGitTerminalPrompt, previousPrompt, hadPrompt)
		restoreEnvVariable(envGitAskPassRequire, previousAskPassRequire, hadAskPassRequire)
		_ = os.Remove(scriptPath)
	}
	return cleanup, nil
}

func restoreEnvVariable(name string, value string, hadValue bool) {
	if hadValue {
		_ = os.Setenv(name, value)
		return
	}
	_ = os.Unsetenv(name)
}

func sanitizePathComponent(value string) string {
	normalized := unsafePathComponentPattern.ReplaceAllString(strings.TrimSpace(value), "-")
	normalized = strings.Trim(normalized, "-")
	if normalized == "" {
		return "run"
	}
	return normalized
}

func (s *Service) sensitiveValues() []string {
	candidates := []string{
		strings.TrimSpace(s.cfg.GitBotToken),
		strings.TrimSpace(s.cfg.OpenAIAPIKey),
		strings.TrimSpace(s.cfg.MCPBearerToken),
	}
	seen := make(map[string]struct{}, len(candidates))
	secrets := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}
		secrets = append(secrets, candidate)
	}
	return secrets
}

func redactSensitiveOutput(raw string, secrets []string) string {
	redacted := strings.TrimSpace(raw)
	if redacted == "" {
		return ""
	}

	for _, secret := range secrets {
		trimmedSecret := strings.TrimSpace(secret)
		if trimmedSecret == "" {
			continue
		}
		redacted = strings.ReplaceAll(redacted, trimmedSecret, redactedSecretValue)
	}

	redacted = githubCredentialURLPattern.ReplaceAllString(redacted, "https://"+redactedSecretValue+":"+redactedSecretValue+"@github.com")
	redacted = authorizationBearerPattern.ReplaceAllString(redacted, "${1}"+redactedSecretValue)
	redacted = githubPATPattern.ReplaceAllString(redacted, redactedSecretValue)
	return redacted
}
