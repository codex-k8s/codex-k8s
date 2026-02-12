package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/codex-k8s/codex-k8s/libs/go/k8s/clientcfg"
	mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"
	corev1 "k8s.io/api/core/v1"
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
