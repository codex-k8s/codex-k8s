package joblauncher

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var nonDNSLabel = regexp.MustCompile(`[^a-z0-9-]`)

// JobState is a current Kubernetes Job execution state.
type JobState string

const (
	// JobStatePending indicates Job exists but has not started active Pods yet.
	JobStatePending JobState = "pending"
	// JobStateRunning indicates Job has active Pods.
	JobStateRunning JobState = "running"
	// JobStateSucceeded indicates Job reached complete condition.
	JobStateSucceeded JobState = "succeeded"
	// JobStateFailed indicates Job reached failed condition.
	JobStateFailed JobState = "failed"
	// JobStateNotFound indicates Job resource does not exist.
	JobStateNotFound JobState = "not_found"
)

// JobRef identifies Kubernetes Job object.
type JobRef struct {
	// Namespace is a Job namespace.
	Namespace string
	// Name is a Job resource name.
	Name string
}

// JobSpec defines minimal metadata for Job creation.
type JobSpec struct {
	// RunID uniquely identifies run.
	RunID string
	// CorrelationID links Job to flow.
	CorrelationID string
	// ProjectID stores effective project scope.
	ProjectID string
	// SlotNo stores slot number assigned to run.
	SlotNo int
}

// Config defines Job launcher runtime options.
type Config struct {
	// KubeconfigPath points to local kubeconfig for out-of-cluster execution.
	KubeconfigPath string
	// Namespace defines namespace for run Jobs.
	Namespace string
	// Image defines container image used by run Jobs.
	Image string
	// Command defines shell command executed by run Jobs.
	Command string
	// TTLSeconds controls ttlSecondsAfterFinished.
	TTLSeconds int32
	// BackoffLimit controls Job retries.
	BackoffLimit int32
	// ActiveDeadlineSeconds controls max execution duration.
	ActiveDeadlineSeconds int64
}

// Launcher creates and reconciles run Jobs in Kubernetes.
type Launcher struct {
	cfg    Config
	client kubernetes.Interface
}

// New creates launcher with auto-detected Kubernetes client configuration.
func New(cfg Config) (*Launcher, error) {
	restCfg, err := buildRESTConfig(cfg.KubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("build kubernetes rest config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("build kubernetes clientset: %w", err)
	}

	return NewForClient(cfg, clientset), nil
}

// NewForClient creates launcher over provided client implementation.
func NewForClient(cfg Config, client kubernetes.Interface) *Launcher {
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}
	if cfg.Image == "" {
		cfg.Image = "busybox:1.36"
	}
	if cfg.Command == "" {
		cfg.Command = "echo codex-k8s run && sleep 2"
	}
	if cfg.TTLSeconds <= 0 {
		cfg.TTLSeconds = 600
	}
	if cfg.ActiveDeadlineSeconds <= 0 {
		cfg.ActiveDeadlineSeconds = 900
	}

	return &Launcher{cfg: cfg, client: client}
}

// JobRef builds deterministic Job reference for run.
func (l *Launcher) JobRef(runID string) JobRef {
	return JobRef{
		Namespace: l.cfg.Namespace,
		Name:      BuildRunJobName(runID),
	}
}

// Launch creates Kubernetes Job or returns existing one when already present.
func (l *Launcher) Launch(ctx context.Context, spec JobSpec) (JobRef, error) {
	ref := l.JobRef(spec.RunID)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ref.Name,
			Namespace: ref.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "codex-k8s-run",
				"app.kubernetes.io/managed-by": "codex-k8s-worker",
				"codexk8s.io/run-id":           spec.RunID,
				"codexk8s.io/project-id":       sanitizeLabel(spec.ProjectID),
			},
			Annotations: map[string]string{
				"codexk8s.io/correlation-id": spec.CorrelationID,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &l.cfg.TTLSeconds,
			BackoffLimit:            &l.cfg.BackoffLimit,
			ActiveDeadlineSeconds:   &l.cfg.ActiveDeadlineSeconds,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name": "codex-k8s-run",
						"codexk8s.io/run-id":     spec.RunID,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "run",
							Image:   l.cfg.Image,
							Command: []string{"/bin/sh", "-c", l.cfg.Command},
							Env: []corev1.EnvVar{
								{Name: "CODEXK8S_RUN_ID", Value: spec.RunID},
								{Name: "CODEXK8S_CORRELATION_ID", Value: spec.CorrelationID},
								{Name: "CODEXK8S_PROJECT_ID", Value: spec.ProjectID},
								{Name: "CODEXK8S_SLOT_NO", Value: fmt.Sprintf("%d", spec.SlotNo)},
							},
						},
					},
				},
			},
		},
	}

	_, err := l.client.BatchV1().Jobs(ref.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return ref, nil
		}
		return JobRef{}, fmt.Errorf("create kubernetes job %s/%s: %w", ref.Namespace, ref.Name, err)
	}

	return ref, nil
}

// Status returns current Job state by Job status fields.
func (l *Launcher) Status(ctx context.Context, ref JobRef) (JobState, error) {
	job, err := l.client.BatchV1().Jobs(ref.Namespace).Get(ctx, ref.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return JobStateNotFound, nil
		}
		return "", fmt.Errorf("get kubernetes job %s/%s: %w", ref.Namespace, ref.Name, err)
	}

	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			return JobStateSucceeded, nil
		}
		if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			return JobStateFailed, nil
		}
	}

	// Some failures (e.g. ImagePullBackOff) don't immediately surface as JobFailed
	// and can keep a run stuck in "pending" forever unless we inspect Pod state.
	if job.Status.Succeeded > 0 {
		return JobStateSucceeded, nil
	}
	if job.Status.Failed > 0 {
		return JobStateFailed, nil
	}
	if job.Status.Active > 0 {
		return JobStateRunning, nil
	}

	pods, err := l.client.CoreV1().Pods(ref.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", ref.Name),
	})
	if err == nil {
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodFailed {
				return JobStateFailed, nil
			}
			if hasTerminalWaitingReason(pod.Status.InitContainerStatuses) || hasTerminalWaitingReason(pod.Status.ContainerStatuses) {
				return JobStateFailed, nil
			}
		}
	}

	return JobStatePending, nil
}

// BuildRunJobName returns deterministic DNS-compatible Job name.
func BuildRunJobName(runID string) string {
	normalized := strings.ToLower(strings.ReplaceAll(runID, "_", "-"))
	normalized = strings.ReplaceAll(normalized, ".", "-")
	normalized = nonDNSLabel.ReplaceAllString(normalized, "")
	normalized = strings.Trim(normalized, "-")
	if normalized == "" {
		normalized = "run"
	}

	name := "codex-k8s-run-" + normalized
	if len(name) > 63 {
		name = name[:63]
	}
	name = strings.TrimRight(name, "-")
	if name == "" {
		return "codex-k8s-run"
	}
	return name
}

func buildRESTConfig(kubeconfigPath string) (*rest.Config, error) {
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

func sanitizeLabel(value string) string {
	if value == "" {
		return "unknown"
	}
	normalized := strings.ToLower(value)
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = nonDNSLabel.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	if normalized == "" {
		return "unknown"
	}
	if len(normalized) > 63 {
		normalized = normalized[:63]
		normalized = strings.TrimRight(normalized, "-")
	}
	if normalized == "" {
		return "unknown"
	}
	return normalized
}

func hasTerminalWaitingReason(statuses []corev1.ContainerStatus) bool {
	for _, cs := range statuses {
		if cs.State.Waiting == nil {
			continue
		}
		reason := cs.State.Waiting.Reason
		if reason == "" {
			continue
		}

		switch reason {
		case "ErrImagePull",
			"ImagePullBackOff",
			"InvalidImageName",
			"CreateContainerConfigError",
			"CreateContainerError",
			"RunContainerError",
			"CrashLoopBackOff":
			return true
		}

		// Generic backoff reasons are almost always terminal in the context of a Job pod.
		if strings.Contains(reason, "BackOff") {
			return true
		}
	}
	return false
}
