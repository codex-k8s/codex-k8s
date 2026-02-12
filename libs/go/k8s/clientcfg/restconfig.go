package clientcfg

import (
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// BuildRESTConfig resolves Kubernetes REST config from explicit kubeconfig,
// in-cluster config, or default kubeconfig rules.
func BuildRESTConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("build kubeconfig from path %q: %w", kubeconfigPath, err)
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err == nil {
		return cfg, nil
	}

	fallbackCfg, fallbackErr := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if fallbackErr != nil {
		return nil, fmt.Errorf("build in-cluster config: %w; fallback config: %v", err, fallbackErr)
	}

	return fallbackCfg, nil
}
