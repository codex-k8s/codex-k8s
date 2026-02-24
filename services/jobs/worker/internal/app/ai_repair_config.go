package app

import "strings"

func resolveAIRepairNamespace(cfg Config) string {
	namespace := strings.TrimSpace(cfg.AIRepairNamespace)
	if namespace != "" {
		return namespace
	}
	namespace = strings.TrimSpace(cfg.ProductionNamespace)
	if namespace != "" {
		return namespace
	}
	return strings.TrimSpace(cfg.K8sNamespace)
}
