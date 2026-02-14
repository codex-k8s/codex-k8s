package kubernetes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/k8s/clientcfg"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1api "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// Client provides Kubernetes operations for control-plane MCP domain.
type Client struct {
	clientset  kubernetes.Interface
	restConfig *rest.Config
	dynamic    dynamic.Interface
	restMapper metav1api.RESTMapper
}

const (
	k8sKindDeployment              = "Deployment"
	k8sKindDaemonSet               = "DaemonSet"
	k8sKindStatefulSet             = "StatefulSet"
	k8sKindReplicaSet              = "ReplicaSet"
	k8sKindReplicationController   = "ReplicationController"
	k8sKindJob                     = "Job"
	k8sKindCronJob                 = "CronJob"
	k8sKindConfigMap               = "ConfigMap"
	k8sKindSecret                  = "Secret"
	k8sKindResourceQuota           = "ResourceQuota"
	k8sKindHorizontalPodAutoscaler = "HorizontalPodAutoscaler"
	k8sKindService                 = "Service"
	k8sKindEndpoints               = "Endpoints"
	k8sKindIngress                 = "Ingress"
	k8sKindIngressClass            = "IngressClass"
	k8sKindNetworkPolicy           = "NetworkPolicy"
	k8sKindPersistentVolumeClaim   = "PersistentVolumeClaim"
	k8sKindPersistentVolume        = "PersistentVolume"
	k8sKindStorageClass            = "StorageClass"

	runNamespaceManagedByLabel   = "codex-k8s.dev/managed-by"
	runNamespaceManagedByValue   = "codex-k8s-worker"
	runNamespacePurposeLabel     = "codex-k8s.dev/namespace-purpose"
	runNamespacePurposeValue     = "run"
	runNamespaceRuntimeModeLabel = "codex-k8s.dev/runtime-mode"
	runNamespaceRuntimeModeValue = "full-env"
)

var nonDNSLabel = regexp.MustCompile(`[^a-z0-9-]`)

// NewClient creates Kubernetes MCP adapter with auto-detected REST config.
func NewClient(kubeconfigPath string) (*Client, error) {
	restConfig, err := clientcfg.BuildRESTConfig(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("build kubernetes rest config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("build kubernetes clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("build kubernetes dynamic client: %w", err)
	}

	mapper, err := buildRESTMapper(restConfig)
	if err != nil {
		return nil, fmt.Errorf("build kubernetes rest mapper: %w", err)
	}

	return NewForClients(restConfig, clientset, dynamicClient, mapper), nil
}

// NewForClient creates Kubernetes MCP adapter for provided clientset.
func NewForClient(restConfig *rest.Config, clientset kubernetes.Interface) *Client {
	dynamicClient, _ := dynamic.NewForConfig(restConfig)
	mapper, _ := buildRESTMapper(restConfig)
	return NewForClients(restConfig, clientset, dynamicClient, mapper)
}

// NewForClients creates Kubernetes adapter with explicit typed/dynamic clients.
func NewForClients(restConfig *rest.Config, clientset kubernetes.Interface, dynamicClient dynamic.Interface, mapper metav1api.RESTMapper) *Client {
	return &Client{
		clientset:  clientset,
		restConfig: rest.CopyConfig(restConfig),
		dynamic:    dynamicClient,
		restMapper: mapper,
	}
}

// ListPods lists pods in namespace with deterministic ordering.
func (c *Client) ListPods(ctx context.Context, namespace string, limit int) ([]mcpdomain.KubernetesPod, error) {
	items, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		Limit: int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	out := make([]mcpdomain.KubernetesPod, 0, len(items.Items))
	for _, item := range items.Items {
		pod := mcpdomain.KubernetesPod{
			Name:     item.Name,
			Phase:    string(item.Status.Phase),
			NodeName: item.Spec.NodeName,
		}
		if item.Status.StartTime != nil {
			pod.StartTime = item.Status.StartTime.UTC().Format(time.RFC3339)
		}
		out = append(out, pod)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// ListEvents lists namespace events with deterministic ordering.
func (c *Client) ListEvents(ctx context.Context, namespace string, limit int) ([]mcpdomain.KubernetesEvent, error) {
	items, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		Limit: int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	out := make([]mcpdomain.KubernetesEvent, 0, len(items.Items))
	for _, item := range items.Items {
		out = append(out, mcpdomain.KubernetesEvent{
			Type:      item.Type,
			Reason:    item.Reason,
			Message:   item.Message,
			Object:    formatInvolvedObject(item.InvolvedObject),
			Timestamp: eventTimestamp(item).Format(time.RFC3339),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Timestamp == out[j].Timestamp {
			return out[i].Object < out[j].Object
		}
		return out[i].Timestamp > out[j].Timestamp
	})
	return out, nil
}

// ListResources lists supported Kubernetes resources for one kind.
func (c *Client) ListResources(ctx context.Context, namespace string, kind mcpdomain.KubernetesResourceKind, limit int) ([]mcpdomain.KubernetesResourceRef, error) {
	operation := resourceListOperationForKind(kind)

	switch kind {
	case mcpdomain.KubernetesResourceKindDeployment:
		return c.listResourceRefs(ctx, limit, k8sKindDeployment, operation, listAsAny(c.clientset.AppsV1().Deployments(namespace).List), true)
	case mcpdomain.KubernetesResourceKindDaemonSet:
		return c.listResourceRefs(ctx, limit, k8sKindDaemonSet, operation, listAsAny(c.clientset.AppsV1().DaemonSets(namespace).List), true)
	case mcpdomain.KubernetesResourceKindStatefulSet:
		return c.listResourceRefs(ctx, limit, k8sKindStatefulSet, operation, listAsAny(c.clientset.AppsV1().StatefulSets(namespace).List), true)
	case mcpdomain.KubernetesResourceKindReplicaSet:
		return c.listResourceRefs(ctx, limit, k8sKindReplicaSet, operation, listAsAny(c.clientset.AppsV1().ReplicaSets(namespace).List), true)
	case mcpdomain.KubernetesResourceKindReplicationController:
		return c.listResourceRefs(ctx, limit, k8sKindReplicationController, operation, listAsAny(c.clientset.CoreV1().ReplicationControllers(namespace).List), true)
	case mcpdomain.KubernetesResourceKindJob:
		return c.listResourceRefs(ctx, limit, k8sKindJob, operation, listAsAny(c.clientset.BatchV1().Jobs(namespace).List), true)
	case mcpdomain.KubernetesResourceKindCronJob:
		return c.listResourceRefs(ctx, limit, k8sKindCronJob, operation, listAsAny(c.clientset.BatchV1().CronJobs(namespace).List), true)
	case mcpdomain.KubernetesResourceKindConfigMap:
		return c.listResourceRefs(ctx, limit, k8sKindConfigMap, operation, listAsAny(c.clientset.CoreV1().ConfigMaps(namespace).List), true)
	case mcpdomain.KubernetesResourceKindSecret:
		return c.listResourceRefs(ctx, limit, k8sKindSecret, operation, listAsAny(c.clientset.CoreV1().Secrets(namespace).List), true)
	case mcpdomain.KubernetesResourceKindResourceQuota:
		return c.listResourceRefs(ctx, limit, k8sKindResourceQuota, operation, listAsAny(c.clientset.CoreV1().ResourceQuotas(namespace).List), true)
	case mcpdomain.KubernetesResourceKindHPA:
		return c.listResourceRefs(ctx, limit, k8sKindHorizontalPodAutoscaler, operation, listAsAny(c.clientset.AutoscalingV2().HorizontalPodAutoscalers(namespace).List), true)
	case mcpdomain.KubernetesResourceKindService:
		return c.listResourceRefs(ctx, limit, k8sKindService, operation, listAsAny(c.clientset.CoreV1().Services(namespace).List), true)
	case mcpdomain.KubernetesResourceKindEndpoints:
		return c.listResourceRefs(ctx, limit, k8sKindEndpoints, operation, listAsAny(c.clientset.CoreV1().Endpoints(namespace).List), true)
	case mcpdomain.KubernetesResourceKindIngress:
		return c.listResourceRefs(ctx, limit, k8sKindIngress, operation, listAsAny(c.clientset.NetworkingV1().Ingresses(namespace).List), true)
	case mcpdomain.KubernetesResourceKindIngressClass:
		return c.listResourceRefs(ctx, limit, k8sKindIngressClass, operation, listAsAny(c.clientset.NetworkingV1().IngressClasses().List), false)
	case mcpdomain.KubernetesResourceKindNetworkPolicy:
		return c.listResourceRefs(ctx, limit, k8sKindNetworkPolicy, operation, listAsAny(c.clientset.NetworkingV1().NetworkPolicies(namespace).List), true)
	case mcpdomain.KubernetesResourceKindPVC:
		return c.listResourceRefs(ctx, limit, k8sKindPersistentVolumeClaim, operation, listAsAny(c.clientset.CoreV1().PersistentVolumeClaims(namespace).List), true)
	case mcpdomain.KubernetesResourceKindPV:
		return c.listResourceRefs(ctx, limit, k8sKindPersistentVolume, operation, listAsAny(c.clientset.CoreV1().PersistentVolumes().List), false)
	case mcpdomain.KubernetesResourceKindStorageClass:
		return c.listResourceRefs(ctx, limit, k8sKindStorageClass, operation, listAsAny(c.clientset.StorageV1().StorageClasses().List), false)
	default:
		return nil, fmt.Errorf("unsupported kubernetes resource kind %q", kind)
	}
}

// GetPodLogs returns pod logs from namespace.
func (c *Client) GetPodLogs(ctx context.Context, namespace string, pod string, container string, tailLines int64) (string, error) {
	options := &corev1.PodLogOptions{
		Container: strings.TrimSpace(container),
	}
	if tailLines > 0 {
		options.TailLines = &tailLines
	}

	stream, err := c.clientset.CoreV1().Pods(namespace).GetLogs(strings.TrimSpace(pod), options).Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("open logs stream: %w", err)
	}
	defer func() { _ = stream.Close() }()

	blob, err := io.ReadAll(stream)
	if err != nil {
		return "", fmt.Errorf("read logs stream: %w", err)
	}
	return string(blob), nil
}

// ExecPod executes command in pod container and returns stdout/stderr output.
func (c *Client) ExecPod(ctx context.Context, namespace string, pod string, container string, command []string) (mcpdomain.KubernetesExecResult, error) {
	options := &corev1.PodExecOptions{
		Container: strings.TrimSpace(container),
		Command:   command,
		Stdout:    true,
		Stderr:    true,
		Stdin:     false,
		TTY:       false,
	}

	request := c.clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(strings.TrimSpace(pod)).
		Namespace(namespace).
		SubResource("exec")
	request.VersionedParams(options, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(c.restConfig, "POST", request.URL())
	if err != nil {
		return mcpdomain.KubernetesExecResult{}, fmt.Errorf("build exec request: %w", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	}); err != nil {
		return mcpdomain.KubernetesExecResult{}, fmt.Errorf("stream pod exec: %w", err)
	}

	return mcpdomain.KubernetesExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}, nil
}

// UpsertSecret creates or updates one namespaced Kubernetes Secret with deterministic data keys.
func (c *Client) UpsertSecret(ctx context.Context, namespace string, secretName string, data map[string][]byte) error {
	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace == "" {
		return fmt.Errorf("kubernetes namespace is required")
	}
	targetSecretName := strings.TrimSpace(secretName)
	if targetSecretName == "" {
		return fmt.Errorf("kubernetes secret name is required")
	}
	if len(data) == 0 {
		return fmt.Errorf("kubernetes secret data is required")
	}

	existing, err := c.clientset.CoreV1().Secrets(targetNamespace).Get(ctx, targetSecretName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return fmt.Errorf("get kubernetes secret %s/%s: %w", targetNamespace, targetSecretName, err)
		}

		secretData := make(map[string][]byte, len(data))
		for key, value := range data {
			secretData[key] = append([]byte(nil), value...)
		}
		_, createErr := c.clientset.CoreV1().Secrets(targetNamespace).Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: targetSecretName,
			},
			Type: corev1.SecretTypeOpaque,
			Data: secretData,
		}, metav1.CreateOptions{})
		if createErr != nil {
			return fmt.Errorf("create kubernetes secret %s/%s: %w", targetNamespace, targetSecretName, createErr)
		}
		return nil
	}

	secretData := make(map[string][]byte, len(data))
	for key, value := range data {
		secretData[key] = append([]byte(nil), value...)
	}
	existing.Data = secretData
	if existing.Type == "" {
		existing.Type = corev1.SecretTypeOpaque
	}
	if _, err := c.clientset.CoreV1().Secrets(targetNamespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("update kubernetes secret %s/%s: %w", targetNamespace, targetSecretName, err)
	}
	return nil
}

// DeleteManagedRunNamespace deletes a full-env run namespace when it is marked as worker-managed.
func (c *Client) DeleteManagedRunNamespace(ctx context.Context, namespace string) (bool, error) {
	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace == "" {
		return false, fmt.Errorf("namespace is required")
	}

	ns, err := c.clientset.CoreV1().Namespaces().Get(ctx, targetNamespace, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("get namespace %s: %w", targetNamespace, err)
	}

	if strings.TrimSpace(ns.Labels[runNamespaceManagedByLabel]) != runNamespaceManagedByValue {
		return false, fmt.Errorf("namespace %s is not managed by codex-k8s-worker", targetNamespace)
	}
	if strings.TrimSpace(ns.Labels[runNamespacePurposeLabel]) != runNamespacePurposeValue {
		return false, fmt.Errorf("namespace %s is not a run namespace", targetNamespace)
	}
	if strings.TrimSpace(ns.Labels[runNamespaceRuntimeModeLabel]) != runNamespaceRuntimeModeValue {
		return false, fmt.Errorf("namespace %s is not full-env runtime namespace", targetNamespace)
	}

	if err := c.clientset.CoreV1().Namespaces().Delete(ctx, targetNamespace, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("delete namespace %s: %w", targetNamespace, err)
	}

	return true, nil
}

// FindManagedRunNamespaceByRunID resolves one managed full-env run namespace by run id label.
func (c *Client) FindManagedRunNamespaceByRunID(ctx context.Context, runID string) (string, bool, error) {
	targetRunID := sanitizeRunLabelValue(runID)
	if targetRunID == "" {
		return "", false, fmt.Errorf("run id is required")
	}

	items, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("codex-k8s.dev/run-id=%s", targetRunID),
	})
	if err != nil {
		return "", false, fmt.Errorf("list namespaces by run id %s: %w", targetRunID, err)
	}

	names := make([]string, 0, len(items.Items))
	for _, item := range items.Items {
		if strings.TrimSpace(item.Labels[runNamespaceManagedByLabel]) != runNamespaceManagedByValue {
			continue
		}
		if strings.TrimSpace(item.Labels[runNamespacePurposeLabel]) != runNamespacePurposeValue {
			continue
		}
		if strings.TrimSpace(item.Labels[runNamespaceRuntimeModeLabel]) != runNamespaceRuntimeModeValue {
			continue
		}
		names = append(names, strings.TrimSpace(item.Name))
	}
	if len(names) == 0 {
		return "", false, nil
	}
	sort.Strings(names)
	return names[0], true, nil
}

// NamespaceExists reports whether namespace exists.
func (c *Client) NamespaceExists(ctx context.Context, namespace string) (bool, error) {
	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace == "" {
		return false, fmt.Errorf("namespace is required")
	}

	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, targetNamespace, metav1.GetOptions{})
	if err == nil {
		return true, nil
	}
	if k8serrors.IsNotFound(err) {
		return false, nil
	}
	return false, fmt.Errorf("get namespace %s: %w", targetNamespace, err)
}

// JobExists reports whether one Kubernetes Job exists.
func (c *Client) JobExists(ctx context.Context, namespace string, jobName string) (bool, error) {
	targetNamespace := strings.TrimSpace(namespace)
	targetJobName := strings.TrimSpace(jobName)
	if targetNamespace == "" {
		return false, fmt.Errorf("namespace is required")
	}
	if targetJobName == "" {
		return false, fmt.Errorf("job name is required")
	}

	_, err := c.clientset.BatchV1().Jobs(targetNamespace).Get(ctx, targetJobName, metav1.GetOptions{})
	if err == nil {
		return true, nil
	}
	if k8serrors.IsNotFound(err) {
		return false, nil
	}
	return false, fmt.Errorf("get job %s/%s: %w", targetNamespace, targetJobName, err)
}

// AppliedResourceRef identifies one applied resource object.
type AppliedResourceRef struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

// UpsertConfigMap creates or updates one namespaced ConfigMap.
func (c *Client) UpsertConfigMap(ctx context.Context, namespace string, name string, data map[string]string) error {
	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace == "" {
		return fmt.Errorf("kubernetes namespace is required")
	}
	targetName := strings.TrimSpace(name)
	if targetName == "" {
		return fmt.Errorf("kubernetes configmap name is required")
	}
	if len(data) == 0 {
		return fmt.Errorf("kubernetes configmap data is required")
	}

	existing, err := c.clientset.CoreV1().ConfigMaps(targetNamespace).Get(ctx, targetName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return fmt.Errorf("get kubernetes configmap %s/%s: %w", targetNamespace, targetName, err)
		}
		copiedData := make(map[string]string, len(data))
		for key, value := range data {
			copiedData[key] = value
		}
		_, createErr := c.clientset.CoreV1().ConfigMaps(targetNamespace).Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: targetName},
			Data:       copiedData,
		}, metav1.CreateOptions{})
		if createErr != nil {
			return fmt.Errorf("create kubernetes configmap %s/%s: %w", targetNamespace, targetName, createErr)
		}
		return nil
	}

	copiedData := make(map[string]string, len(data))
	for key, value := range data {
		copiedData[key] = value
	}
	existing.Data = copiedData
	if _, err := c.clientset.CoreV1().ConfigMaps(targetNamespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("update kubernetes configmap %s/%s: %w", targetNamespace, targetName, err)
	}
	return nil
}

// GetSecretData returns one namespaced secret data map when secret exists.
func (c *Client) GetSecretData(ctx context.Context, namespace string, name string) (map[string][]byte, bool, error) {
	targetNamespace := strings.TrimSpace(namespace)
	targetName := strings.TrimSpace(name)
	if targetNamespace == "" || targetName == "" {
		return nil, false, fmt.Errorf("secret namespace and name are required")
	}

	secret, err := c.clientset.CoreV1().Secrets(targetNamespace).Get(ctx, targetName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get kubernetes secret %s/%s: %w", targetNamespace, targetName, err)
	}

	out := make(map[string][]byte, len(secret.Data))
	for key, value := range secret.Data {
		out[key] = append([]byte(nil), value...)
	}
	return out, true, nil
}

// DeleteJobIfExists deletes namespaced Job if present.
func (c *Client) DeleteJobIfExists(ctx context.Context, namespace string, name string) error {
	targetNamespace := strings.TrimSpace(namespace)
	targetName := strings.TrimSpace(name)
	if targetNamespace == "" || targetName == "" {
		return fmt.Errorf("job namespace and name are required")
	}
	if err := c.clientset.BatchV1().Jobs(targetNamespace).Delete(ctx, targetName, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("delete job %s/%s: %w", targetNamespace, targetName, err)
	}
	return nil
}

// WaitForJobComplete waits until one Job reports Complete condition.
func (c *Client) WaitForJobComplete(ctx context.Context, namespace string, name string, timeout time.Duration) error {
	targetNamespace := strings.TrimSpace(namespace)
	targetName := strings.TrimSpace(name)
	if targetNamespace == "" || targetName == "" {
		return fmt.Errorf("job namespace and name are required")
	}
	if timeout <= 0 {
		timeout = 20 * time.Minute
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("wait job %s/%s complete: %w", targetNamespace, targetName, waitCtx.Err())
		case <-ticker.C:
			job, err := c.clientset.BatchV1().Jobs(targetNamespace).Get(waitCtx, targetName, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return fmt.Errorf("get job %s/%s: %w", targetNamespace, targetName, err)
			}
			for _, condition := range job.Status.Conditions {
				if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
					return nil
				}
				if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
					return fmt.Errorf("job %s/%s failed", targetNamespace, targetName)
				}
			}
		}
	}
}

// WaitForDeploymentReady waits until deployment reports Available condition.
func (c *Client) WaitForDeploymentReady(ctx context.Context, namespace string, name string, timeout time.Duration) error {
	targetNamespace := strings.TrimSpace(namespace)
	targetName := strings.TrimSpace(name)
	if targetNamespace == "" || targetName == "" {
		return fmt.Errorf("deployment namespace and name are required")
	}
	if timeout <= 0 {
		timeout = 20 * time.Minute
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("wait deployment %s/%s ready: %w", targetNamespace, targetName, waitCtx.Err())
		case <-ticker.C:
			deployment, err := c.clientset.AppsV1().Deployments(targetNamespace).Get(waitCtx, targetName, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return fmt.Errorf("get deployment %s/%s: %w", targetNamespace, targetName, err)
			}
			if isDeploymentReady(deployment) {
				return nil
			}
		}
	}
}

// WaitForStatefulSetReady waits until statefulset has all replicas ready.
func (c *Client) WaitForStatefulSetReady(ctx context.Context, namespace string, name string, timeout time.Duration) error {
	targetNamespace := strings.TrimSpace(namespace)
	targetName := strings.TrimSpace(name)
	if targetNamespace == "" || targetName == "" {
		return fmt.Errorf("statefulset namespace and name are required")
	}
	if timeout <= 0 {
		timeout = 20 * time.Minute
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("wait statefulset %s/%s ready: %w", targetNamespace, targetName, waitCtx.Err())
		case <-ticker.C:
			statefulSet, err := c.clientset.AppsV1().StatefulSets(targetNamespace).Get(waitCtx, targetName, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return fmt.Errorf("get statefulset %s/%s: %w", targetNamespace, targetName, err)
			}
			specReplicas := int32(1)
			if statefulSet.Spec.Replicas != nil {
				specReplicas = *statefulSet.Spec.Replicas
			}
			if statefulSet.Status.ReadyReplicas >= specReplicas {
				return nil
			}
		}
	}
}

// WaitForDaemonSetReady waits until daemonset reports desired number scheduled as ready.
func (c *Client) WaitForDaemonSetReady(ctx context.Context, namespace string, name string, timeout time.Duration) error {
	targetNamespace := strings.TrimSpace(namespace)
	targetName := strings.TrimSpace(name)
	if targetNamespace == "" || targetName == "" {
		return fmt.Errorf("daemonset namespace and name are required")
	}
	if timeout <= 0 {
		timeout = 20 * time.Minute
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("wait daemonset %s/%s ready: %w", targetNamespace, targetName, waitCtx.Err())
		case <-ticker.C:
			daemonSet, err := c.clientset.AppsV1().DaemonSets(targetNamespace).Get(waitCtx, targetName, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					continue
				}
				return fmt.Errorf("get daemonset %s/%s: %w", targetNamespace, targetName, err)
			}
			if daemonSet.Status.DesiredNumberScheduled > 0 &&
				daemonSet.Status.NumberReady >= daemonSet.Status.DesiredNumberScheduled {
				return nil
			}
		}
	}
}

// ApplyManifest applies YAML manifest documents via server-side apply.
func (c *Client) ApplyManifest(ctx context.Context, manifest []byte, namespaceOverride string, fieldManager string) ([]AppliedResourceRef, error) {
	if len(bytes.TrimSpace(manifest)) == 0 {
		return nil, nil
	}
	if c.dynamic == nil {
		return nil, fmt.Errorf("dynamic kubernetes client is not configured")
	}
	if c.restMapper == nil {
		return nil, fmt.Errorf("kubernetes rest mapper is not configured")
	}
	manager := strings.TrimSpace(fieldManager)
	if manager == "" {
		manager = "codex-k8s-control-plane"
	}
	overrideNamespace := strings.TrimSpace(namespaceOverride)

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 4096)
	applied := make([]AppliedResourceRef, 0, 8)

	for {
		var objectMap map[string]any
		if err := decoder.Decode(&objectMap); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decode manifest yaml: %w", err)
		}
		if len(objectMap) == 0 {
			continue
		}

		obj := &unstructured.Unstructured{Object: objectMap}
		gvk := obj.GroupVersionKind()
		if gvk.Empty() {
			return nil, fmt.Errorf("manifest object has empty apiVersion/kind")
		}
		targetName := strings.TrimSpace(obj.GetName())
		if targetName == "" {
			return nil, fmt.Errorf("manifest object %s/%s has empty metadata.name", gvk.GroupVersion().String(), gvk.Kind)
		}

		mapping, err := c.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return nil, fmt.Errorf("resolve rest mapping for %s: %w", gvk.String(), err)
		}

		resource := c.dynamic.Resource(mapping.Resource)
		var resourceInterface dynamic.ResourceInterface
		targetNamespace := strings.TrimSpace(obj.GetNamespace())
		if mapping.Scope.Name() == metav1api.RESTScopeNameNamespace {
			if overrideNamespace != "" {
				targetNamespace = overrideNamespace
			}
			if targetNamespace == "" {
				return nil, fmt.Errorf("manifest object %s/%s is namespaced but namespace is empty", gvk.String(), targetName)
			}
			obj.SetNamespace(targetNamespace)
			resourceInterface = resource.Namespace(targetNamespace)
		} else {
			resourceInterface = resource
		}

		payload, err := json.Marshal(obj.Object)
		if err != nil {
			return nil, fmt.Errorf("marshal manifest object %s/%s: %w", gvk.String(), targetName, err)
		}
		force := true
		if _, err := resourceInterface.Patch(ctx, targetName, types.ApplyPatchType, payload, metav1.PatchOptions{
			FieldManager: manager,
			Force:        &force,
		}); err != nil {
			return nil, fmt.Errorf("apply manifest object %s/%s: %w", gvk.String(), targetName, err)
		}

		applied = append(applied, AppliedResourceRef{
			APIVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
			Namespace:  targetNamespace,
			Name:       targetName,
		})
	}

	return applied, nil
}

func isDeploymentReady(deployment *appsv1.Deployment) bool {
	if deployment == nil {
		return false
	}
	if deployment.Status.ObservedGeneration < deployment.Generation {
		return false
	}
	replicas := int32(1)
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	}
	if deployment.Status.UpdatedReplicas < replicas {
		return false
	}
	if deployment.Status.AvailableReplicas < replicas {
		return false
	}
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentAvailable && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func buildRESTMapper(restConfig *rest.Config) (metav1api.RESTMapper, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, err
	}
	return restmapper.NewDiscoveryRESTMapper(groupResources), nil
}

func listAsAny[T any](fn func(context.Context, metav1.ListOptions) (*T, error)) func(context.Context, metav1.ListOptions) (any, error) {
	return func(ctx context.Context, options metav1.ListOptions) (any, error) {
		return fn(ctx, options)
	}
}

func (c *Client) listResourceRefs(
	ctx context.Context,
	limit int,
	kind string,
	operation resourceListOperation,
	listFn func(context.Context, metav1.ListOptions) (any, error),
	includeNamespace bool,
) ([]mcpdomain.KubernetesResourceRef, error) {
	list, err := listFn(ctx, metav1.ListOptions{
		Limit: int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	refs, err := resourceRefsFromList(kind, list, includeNamespace)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	sortResourceRefs(refs)
	return refs, nil
}

func resourceRefsFromList(kind string, list any, includeNamespace bool) ([]mcpdomain.KubernetesResourceRef, error) {
	value := reflect.ValueOf(list)
	if !value.IsValid() {
		return nil, fmt.Errorf("kubernetes list response is invalid")
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, fmt.Errorf("kubernetes list response is nil")
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("kubernetes list response must be struct, got %s", value.Kind())
	}

	itemsField := value.FieldByName("Items")
	if !itemsField.IsValid() || itemsField.Kind() != reflect.Slice {
		return nil, fmt.Errorf("kubernetes list response does not expose Items")
	}

	out := make([]mcpdomain.KubernetesResourceRef, 0, itemsField.Len())
	for i := 0; i < itemsField.Len(); i++ {
		item := itemsField.Index(i)
		var obj metav1.Object

		if item.Kind() == reflect.Pointer {
			if item.IsNil() {
				continue
			}
			if casted, ok := item.Interface().(metav1.Object); ok {
				obj = casted
			}
		} else if item.CanAddr() {
			if casted, ok := item.Addr().Interface().(metav1.Object); ok {
				obj = casted
			}
		}
		if obj == nil {
			continue
		}

		ref := mcpdomain.KubernetesResourceRef{
			Kind: kind,
			Name: strings.TrimSpace(obj.GetName()),
		}
		if includeNamespace {
			ref.Namespace = strings.TrimSpace(obj.GetNamespace())
		}
		out = append(out, ref)
	}
	return out, nil
}

func sortResourceRefs(items []mcpdomain.KubernetesResourceRef) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Kind == items[j].Kind {
			if items[i].Namespace == items[j].Namespace {
				return items[i].Name < items[j].Name
			}
			return items[i].Namespace < items[j].Namespace
		}
		return items[i].Kind < items[j].Kind
	})
}

func sanitizeRunLabelValue(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return ""
	}
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = nonDNSLabel.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	if normalized == "" {
		return ""
	}
	if len(normalized) > 63 {
		normalized = strings.TrimRight(normalized[:63], "-")
	}
	return normalized
}

func formatInvolvedObject(ref corev1.ObjectReference) string {
	kind := strings.TrimSpace(ref.Kind)
	name := strings.TrimSpace(ref.Name)
	if kind == "" && name == "" {
		return ""
	}
	if kind == "" {
		return name
	}
	if name == "" {
		return kind
	}
	return kind + "/" + name
}

func eventTimestamp(event corev1.Event) time.Time {
	if !event.EventTime.IsZero() {
		return event.EventTime.UTC()
	}
	if !event.LastTimestamp.IsZero() {
		return event.LastTimestamp.UTC()
	}
	if !event.FirstTimestamp.IsZero() {
		return event.FirstTimestamp.UTC()
	}
	if !event.CreationTimestamp.IsZero() {
		return event.CreationTimestamp.UTC()
	}
	return time.Time{}
}
