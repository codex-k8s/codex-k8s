package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codex-k8s/codex-k8s/libs/go/servicescfg"
)

func loadSecretResolver(configPath string, values map[string]string) (servicescfg.SecretResolver, error) {
	path := strings.TrimSpace(configPath)
	if path == "" {
		return servicescfg.NewSecretResolver(nil), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return servicescfg.NewSecretResolver(nil), fmt.Errorf("resolve config path %q: %w", path, err)
	}
	if _, err := os.Stat(absPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return servicescfg.NewSecretResolver(nil), nil
		}
		return servicescfg.NewSecretResolver(nil), fmt.Errorf("stat config path %q: %w", absPath, err)
	}

	envName := strings.TrimSpace(values["CODEXK8S_SERVICES_CONFIG_ENV"])
	if envName == "" {
		envName = githubEnvironmentProduction
	}
	loaded, err := servicescfg.Load(absPath, servicescfg.LoadOptions{
		Env:  envName,
		Vars: values,
	})
	if err != nil && !strings.EqualFold(envName, githubEnvironmentProduction) {
		loaded, err = servicescfg.Load(absPath, servicescfg.LoadOptions{
			Env:  githubEnvironmentProduction,
			Vars: values,
		})
	}
	if err != nil {
		return servicescfg.NewSecretResolver(nil), fmt.Errorf("load services config %q: %w", absPath, err)
	}
	return servicescfg.NewSecretResolver(loaded.Stack), nil
}
