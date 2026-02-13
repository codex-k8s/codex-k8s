package kubernetes

import (
	"bytes"
	"context"
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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// Client provides Kubernetes operations for control-plane MCP domain.
type Client struct {
	clientset  kubernetes.Interface
	restConfig *rest.Config
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

	return NewForClient(restConfig, clientset), nil
}

// NewForClient creates Kubernetes MCP adapter for provided clientset.
func NewForClient(restConfig *rest.Config, clientset kubernetes.Interface) *Client {
	return &Client{
		clientset:  clientset,
		restConfig: rest.CopyConfig(restConfig),
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
