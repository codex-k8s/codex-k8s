package workerpresence

import (
	"context"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/codex-k8s/codex-k8s/libs/go/k8s/clientcfg"
	"github.com/codex-k8s/codex-k8s/services/jobs/worker/internal/domain/worker"
)

const defaultWorkerLabelSelector = "app.kubernetes.io/name=codex-k8s,app.kubernetes.io/component=worker"

// Adapter lists active worker pods from Kubernetes.
type Adapter struct {
	client        kubernetes.Interface
	namespace     string
	labelSelector string
}

// NewAdapter creates worker presence adapter with auto-detected Kubernetes client configuration.
func NewAdapter(kubeconfigPath string, namespace string) (*Adapter, error) {
	restCfg, err := clientcfg.BuildRESTConfig(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("build kubernetes rest config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("build kubernetes clientset: %w", err)
	}

	return NewAdapterForClient(namespace, clientset), nil
}

// NewAdapterForClient creates worker presence adapter over provided client implementation.
func NewAdapterForClient(namespace string, client kubernetes.Interface) *Adapter {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		namespace = "default"
	}

	return &Adapter{
		client:        client,
		namespace:     namespace,
		labelSelector: defaultWorkerLabelSelector,
	}
}

// ListActiveWorkerIDs returns active worker pod names in deterministic order.
func (a *Adapter) ListActiveWorkerIDs(ctx context.Context) ([]string, error) {
	pods, err := a.client.CoreV1().Pods(a.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: a.labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("list worker pods: %w", err)
	}

	active := make([]string, 0, len(pods.Items))
	seen := make(map[string]struct{}, len(pods.Items))
	for _, pod := range pods.Items {
		if !isActiveWorkerPod(pod) {
			continue
		}
		if _, ok := seen[pod.Name]; ok {
			continue
		}
		seen[pod.Name] = struct{}{}
		active = append(active, pod.Name)
	}

	sort.Strings(active)
	return active, nil
}

func isActiveWorkerPod(pod corev1.Pod) bool {
	name := strings.TrimSpace(pod.Name)
	if name == "" {
		return false
	}
	if pod.DeletionTimestamp != nil {
		return false
	}

	switch pod.Status.Phase {
	case corev1.PodFailed, corev1.PodSucceeded:
		return false
	default:
		return true
	}
}

var _ worker.WorkerPresenceChecker = (*Adapter)(nil)
