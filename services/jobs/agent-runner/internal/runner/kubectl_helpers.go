package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	k8sServiceAccountRoot     = "/var/run/secrets/kubernetes.io/serviceaccount"
	k8sServiceAccountToken    = k8sServiceAccountRoot + "/token"
	k8sServiceAccountCA       = k8sServiceAccountRoot + "/ca.crt"
	k8sServiceAccountNS       = k8sServiceAccountRoot + "/namespace"
	k8sEnvServiceHost         = "KUBERNETES_SERVICE_HOST"
	k8sEnvServicePort         = "KUBERNETES_SERVICE_PORT"
	kubectlClusterName        = "incluster"
	kubectlContextName        = "incluster"
	kubectlUserName           = "run-sa"
	kubectlKubeconfigFileName = "config"
)

func (s *Service) configureKubectlAccess(homeDir string) (func(), error) {
	if normalizeRuntimeMode(s.cfg.RuntimeMode) != runtimeModeFullEnv {
		return func() {}, nil
	}

	serviceHost := strings.TrimSpace(os.Getenv(k8sEnvServiceHost))
	servicePort := strings.TrimSpace(os.Getenv(k8sEnvServicePort))
	if serviceHost == "" || servicePort == "" {
		return nil, fmt.Errorf("full-env requires Kubernetes in-cluster endpoint variables")
	}

	tokenBytes, err := os.ReadFile(k8sServiceAccountToken)
	if err != nil {
		return nil, fmt.Errorf("read serviceaccount token: %w", err)
	}
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return nil, fmt.Errorf("serviceaccount token is empty")
	}

	namespaceBytes, err := os.ReadFile(k8sServiceAccountNS)
	if err != nil {
		return nil, fmt.Errorf("read serviceaccount namespace: %w", err)
	}
	namespace := strings.TrimSpace(string(namespaceBytes))
	if namespace == "" {
		return nil, fmt.Errorf("serviceaccount namespace is empty")
	}

	kubeDir := filepath.Join(homeDir, ".kube")
	if err := os.MkdirAll(kubeDir, 0o700); err != nil {
		return nil, fmt.Errorf("create kube dir: %w", err)
	}
	kubeconfigPath := filepath.Join(kubeDir, kubectlKubeconfigFileName)
	kubeconfigContent, err := renderTemplate(templateNameKubeconfig, kubectlKubeconfigTemplateData{
		CACertificatePath: k8sServiceAccountCA,
		ServiceHost:       serviceHost,
		ServicePort:       servicePort,
		ClusterName:       kubectlClusterName,
		Namespace:         namespace,
		UserName:          kubectlUserName,
		ContextName:       kubectlContextName,
		Token:             token,
	})
	if err != nil {
		return nil, fmt.Errorf("render kubeconfig template: %w", err)
	}

	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfigContent), 0o600); err != nil {
		return nil, fmt.Errorf("write kubeconfig: %w", err)
	}

	previousKubeconfig, hadKubeconfig := os.LookupEnv(envKubeconfig)
	if err := os.Setenv(envKubeconfig, kubeconfigPath); err != nil {
		return nil, fmt.Errorf("set %s: %w", envKubeconfig, err)
	}

	cleanup := func() {
		restoreEnvVariable(envKubeconfig, previousKubeconfig, hadKubeconfig)
		_ = os.Remove(kubeconfigPath)
	}
	return cleanup, nil
}
