package main

import (
	"context"
	"time"

	kubernetesclient "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/clients/kubernetes"
	runtimedeploydomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/runtimedeploy"
)

type runtimeDeployKubernetesAdapter struct {
	client *kubernetesclient.Client
}

func (a runtimeDeployKubernetesAdapter) EnsureNamespace(ctx context.Context, namespace string) error {
	return a.client.EnsureNamespace(ctx, namespace)
}

func (a runtimeDeployKubernetesAdapter) UpsertSecret(ctx context.Context, namespace string, secretName string, data map[string][]byte) error {
	return a.client.UpsertSecret(ctx, namespace, secretName, data)
}

func (a runtimeDeployKubernetesAdapter) UpsertTLSSecret(ctx context.Context, namespace string, secretName string, data map[string][]byte) error {
	return a.client.UpsertTLSSecret(ctx, namespace, secretName, data)
}

func (a runtimeDeployKubernetesAdapter) UpsertConfigMap(ctx context.Context, namespace string, name string, data map[string]string) error {
	return a.client.UpsertConfigMap(ctx, namespace, name, data)
}

func (a runtimeDeployKubernetesAdapter) GetSecretData(ctx context.Context, namespace string, name string) (map[string][]byte, bool, error) {
	return a.client.GetSecretData(ctx, namespace, name)
}

func (a runtimeDeployKubernetesAdapter) DeleteJobIfExists(ctx context.Context, namespace string, name string) error {
	return a.client.DeleteJobIfExists(ctx, namespace, name)
}

func (a runtimeDeployKubernetesAdapter) WaitForJobComplete(ctx context.Context, namespace string, name string, timeout time.Duration) error {
	return a.client.WaitForJobComplete(ctx, namespace, name, timeout)
}

func (a runtimeDeployKubernetesAdapter) GetJobLogs(ctx context.Context, namespace string, name string, tailLines int64) (string, error) {
	return a.client.GetJobLogs(ctx, namespace, name, tailLines)
}

func (a runtimeDeployKubernetesAdapter) WaitForDeploymentReady(ctx context.Context, namespace string, name string, timeout time.Duration) error {
	return a.client.WaitForDeploymentReady(ctx, namespace, name, timeout)
}

func (a runtimeDeployKubernetesAdapter) WaitForStatefulSetReady(ctx context.Context, namespace string, name string, timeout time.Duration) error {
	return a.client.WaitForStatefulSetReady(ctx, namespace, name, timeout)
}

func (a runtimeDeployKubernetesAdapter) WaitForDaemonSetReady(ctx context.Context, namespace string, name string, timeout time.Duration) error {
	return a.client.WaitForDaemonSetReady(ctx, namespace, name, timeout)
}

func (a runtimeDeployKubernetesAdapter) ApplyManifest(ctx context.Context, manifest []byte, namespaceOverride string, fieldManager string) ([]runtimedeploydomain.AppliedResourceRef, error) {
	refs, err := a.client.ApplyManifest(ctx, manifest, namespaceOverride, fieldManager)
	if err != nil {
		return nil, err
	}
	out := make([]runtimedeploydomain.AppliedResourceRef, 0, len(refs))
	for _, ref := range refs {
		out = append(out, runtimedeploydomain.AppliedResourceRef{
			APIVersion: ref.APIVersion,
			Kind:       ref.Kind,
			Namespace:  ref.Namespace,
			Name:       ref.Name,
		})
	}
	return out, nil
}
