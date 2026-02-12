package joblauncher

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	"github.com/codex-k8s/codex-k8s/libs/go/k8s/clientcfg"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	// RuntimeMode controls run profile in Kubernetes namespace.
	RuntimeMode agentdomain.RuntimeMode
	// Namespace is preferred namespace for this run.
	Namespace string
	// ControlPlaneGRPCTarget is control-plane gRPC endpoint for run callbacks.
	ControlPlaneGRPCTarget string
	// MCPBaseURL is control-plane MCP StreamableHTTP endpoint for run pod.
	MCPBaseURL string
	// MCPBearerToken is short-lived token bound to run and used for MCP auth.
	MCPBearerToken string
	// RepositoryFullName is repository slug in owner/name format.
	RepositoryFullName string
	// IssueNumber is issue number for deterministic branch policy.
	IssueNumber int64
	// TriggerKind defines run mode source (`dev`/`dev_revise`).
	TriggerKind string
	// TriggerLabel is original label that created this run.
	TriggerLabel string
	// TargetBranch overrides deterministic branch naming when already known.
	TargetBranch string
	// ExistingPRNumber preloads PR reference for revise flows when already known.
	ExistingPRNumber int
	// AgentKey is stable system-agent key used for session ownership.
	AgentKey string
	// AgentModel is effective model selected for this run.
	AgentModel string
	// AgentReasoningEffort is effective reasoning profile selected for this run.
	AgentReasoningEffort string
	// PromptTemplateKind is effective prompt kind (`work`/`review`).
	PromptTemplateKind string
	// PromptTemplateSource is effective prompt source (`repo_seed` for Day4 baseline).
	PromptTemplateSource string
	// PromptTemplateLocale is effective prompt locale.
	PromptTemplateLocale string
	// BaseBranch is base branch for PR flow.
	BaseBranch string
	// OpenAIAPIKey is passed to run pod for codex login.
	OpenAIAPIKey string
	// OpenAIAuthFile is optional auth.json content used by codex-cli without login.
	OpenAIAuthFile string
	// Context7APIKey enables Context7 docs lookups inside run pod when provided.
	Context7APIKey string
	// AgentDisplayName is human-readable agent name used for commit author.
	AgentDisplayName string
	// StateInReviewLabel is status label applied to PR when run waits owner review.
	StateInReviewLabel string
	// GitBotToken is passed to run pod for git transport operations.
	GitBotToken string
	// GitBotUsername is GitHub username used with token for git transport auth.
	GitBotUsername string
	// GitBotMail is git author email configured inside run pod.
	GitBotMail string
}

// NamespaceSpec defines runtime namespace metadata.
type NamespaceSpec struct {
	// RunID identifies run owning namespace lifecycle.
	RunID string
	// ProjectID identifies project scope for namespace metadata.
	ProjectID string
	// CorrelationID links namespace events to webhook flow.
	CorrelationID string
	// RuntimeMode controls whether namespace should be managed.
	RuntimeMode agentdomain.RuntimeMode
	// Namespace is target namespace name.
	Namespace string
}

// Config defines Job launcher runtime options.
type Config struct {
	// KubeconfigPath points to local kubeconfig for out-of-cluster execution.
	KubeconfigPath string
	// Namespace defines shared namespace for code-only runs.
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
	// RunServiceAccountName defines service account for full-env run jobs.
	RunServiceAccountName string
	// RunRoleName defines RBAC role name for full-env run jobs.
	RunRoleName string
	// RunRoleBindingName defines RBAC role binding name for full-env run jobs.
	RunRoleBindingName string
	// RunResourceQuotaName defines resource quota object name in runtime namespaces.
	RunResourceQuotaName string
	// RunLimitRangeName defines limit range object name in runtime namespaces.
	RunLimitRangeName string
	// RunCredentialsSecretName defines secret object with run credentials in runtime namespaces.
	RunCredentialsSecretName string
	// RunResourceQuotaPods defines max pod count in runtime namespace.
	RunResourceQuotaPods int64
	// RunResourceRequestsCPU defines requests.cpu hard quota in runtime namespace.
	RunResourceRequestsCPU string
	// RunResourceRequestsMemory defines requests.memory hard quota in runtime namespace.
	RunResourceRequestsMemory string
	// RunResourceLimitsCPU defines limits.cpu hard quota in runtime namespace.
	RunResourceLimitsCPU string
	// RunResourceLimitsMemory defines limits.memory hard quota in runtime namespace.
	RunResourceLimitsMemory string
	// RunDefaultRequestCPU defines default container CPU request in runtime namespace.
	RunDefaultRequestCPU string
	// RunDefaultRequestMemory defines default container memory request in runtime namespace.
	RunDefaultRequestMemory string
	// RunDefaultLimitCPU defines default container CPU limit in runtime namespace.
	RunDefaultLimitCPU string
	// RunDefaultLimitMemory defines default container memory limit in runtime namespace.
	RunDefaultLimitMemory string
}

// Launcher creates and reconciles run Jobs in Kubernetes.
type Launcher struct {
	cfg    Config
	client kubernetes.Interface
}

// New creates launcher with auto-detected Kubernetes client configuration.
func New(cfg Config) (*Launcher, error) {
	restCfg, err := clientcfg.BuildRESTConfig(cfg.KubeconfigPath)
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
		cfg.Command = "/usr/local/bin/codex-k8s-agent-runner"
	}
	if cfg.TTLSeconds <= 0 {
		cfg.TTLSeconds = 600
	}
	if cfg.ActiveDeadlineSeconds <= 0 {
		cfg.ActiveDeadlineSeconds = 900
	}
	if cfg.RunServiceAccountName == "" {
		cfg.RunServiceAccountName = "codex-runner"
	}
	if cfg.RunRoleName == "" {
		cfg.RunRoleName = "codex-runner"
	}
	if cfg.RunRoleBindingName == "" {
		cfg.RunRoleBindingName = "codex-runner"
	}
	if cfg.RunResourceQuotaName == "" {
		cfg.RunResourceQuotaName = "codex-run-quota"
	}
	if cfg.RunLimitRangeName == "" {
		cfg.RunLimitRangeName = "codex-run-limits"
	}
	if cfg.RunCredentialsSecretName == "" {
		cfg.RunCredentialsSecretName = "codex-run-credentials"
	}
	if cfg.RunResourceQuotaPods <= 0 {
		cfg.RunResourceQuotaPods = 20
	}
	if strings.TrimSpace(cfg.RunResourceRequestsCPU) == "" {
		cfg.RunResourceRequestsCPU = "4"
	}
	if strings.TrimSpace(cfg.RunResourceRequestsMemory) == "" {
		cfg.RunResourceRequestsMemory = "8Gi"
	}
	if strings.TrimSpace(cfg.RunResourceLimitsCPU) == "" {
		cfg.RunResourceLimitsCPU = "8"
	}
	if strings.TrimSpace(cfg.RunResourceLimitsMemory) == "" {
		cfg.RunResourceLimitsMemory = "16Gi"
	}
	if strings.TrimSpace(cfg.RunDefaultRequestCPU) == "" {
		cfg.RunDefaultRequestCPU = "250m"
	}
	if strings.TrimSpace(cfg.RunDefaultRequestMemory) == "" {
		cfg.RunDefaultRequestMemory = "256Mi"
	}
	if strings.TrimSpace(cfg.RunDefaultLimitCPU) == "" {
		cfg.RunDefaultLimitCPU = "1"
	}
	if strings.TrimSpace(cfg.RunDefaultLimitMemory) == "" {
		cfg.RunDefaultLimitMemory = "1Gi"
	}

	return &Launcher{cfg: cfg, client: client}
}

// JobRef builds deterministic Job reference for run.
func (l *Launcher) JobRef(runID string, namespace string) JobRef {
	ns := strings.TrimSpace(namespace)
	if ns == "" {
		ns = l.cfg.Namespace
	}
	return JobRef{
		Namespace: ns,
		Name:      BuildRunJobName(runID),
	}
}

// Launch creates Kubernetes Job or returns existing one when already present.
func (l *Launcher) Launch(ctx context.Context, spec JobSpec) (JobRef, error) {
	ref := l.JobRef(spec.RunID, spec.Namespace)
	podSpec := corev1.PodSpec{
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
					{Name: "CODEXK8S_RUNTIME_MODE", Value: string(spec.RuntimeMode)},
					{Name: "CODEXK8S_CONTROL_PLANE_GRPC_TARGET", Value: strings.TrimSpace(spec.ControlPlaneGRPCTarget)},
					{Name: "CODEXK8S_MCP_BASE_URL", Value: strings.TrimSpace(spec.MCPBaseURL)},
					{Name: "CODEXK8S_MCP_BEARER_TOKEN", Value: strings.TrimSpace(spec.MCPBearerToken)},
					{Name: "CODEXK8S_REPOSITORY_FULL_NAME", Value: strings.TrimSpace(spec.RepositoryFullName)},
					{Name: "CODEXK8S_ISSUE_NUMBER", Value: fmt.Sprintf("%d", spec.IssueNumber)},
					{Name: "CODEXK8S_RUN_TRIGGER_KIND", Value: strings.TrimSpace(spec.TriggerKind)},
					{Name: "CODEXK8S_RUN_TRIGGER_LABEL", Value: strings.TrimSpace(spec.TriggerLabel)},
					{Name: "CODEXK8S_RUN_TARGET_BRANCH", Value: strings.TrimSpace(spec.TargetBranch)},
					{Name: "CODEXK8S_EXISTING_PR_NUMBER", Value: fmt.Sprintf("%d", spec.ExistingPRNumber)},
					{Name: "CODEXK8S_AGENT_KEY", Value: strings.TrimSpace(spec.AgentKey)},
					{Name: "CODEXK8S_AGENT_MODEL", Value: strings.TrimSpace(spec.AgentModel)},
					{Name: "CODEXK8S_AGENT_REASONING_EFFORT", Value: strings.TrimSpace(spec.AgentReasoningEffort)},
					{Name: "CODEXK8S_PROMPT_TEMPLATE_KIND", Value: strings.TrimSpace(spec.PromptTemplateKind)},
					{Name: "CODEXK8S_PROMPT_TEMPLATE_SOURCE", Value: strings.TrimSpace(spec.PromptTemplateSource)},
					{Name: "CODEXK8S_PROMPT_TEMPLATE_LOCALE", Value: strings.TrimSpace(spec.PromptTemplateLocale)},
					{Name: "CODEXK8S_STATE_IN_REVIEW_LABEL", Value: strings.TrimSpace(spec.StateInReviewLabel)},
					{Name: "CODEXK8S_AGENT_BASE_BRANCH", Value: strings.TrimSpace(spec.BaseBranch)},
					{Name: "CODEXK8S_OPENAI_API_KEY", Value: strings.TrimSpace(spec.OpenAIAPIKey)},
					{Name: "CODEXK8S_OPENAI_AUTH_FILE", Value: strings.TrimSpace(spec.OpenAIAuthFile)},
					{Name: "CODEXK8S_CONTEXT7_API_KEY", Value: strings.TrimSpace(spec.Context7APIKey)},
					{Name: "CODEXK8S_AGENT_DISPLAY_NAME", Value: strings.TrimSpace(spec.AgentDisplayName)},
					{Name: "CODEXK8S_GIT_BOT_TOKEN", Value: strings.TrimSpace(spec.GitBotToken)},
					{Name: "CODEXK8S_GIT_BOT_USERNAME", Value: strings.TrimSpace(spec.GitBotUsername)},
					{Name: "CODEXK8S_GIT_BOT_MAIL", Value: strings.TrimSpace(spec.GitBotMail)},
				},
			},
		},
	}
	if spec.RuntimeMode == agentdomain.RuntimeModeFullEnv {
		podSpec.ServiceAccountName = l.cfg.RunServiceAccountName
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ref.Name,
			Namespace: ref.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "codex-k8s-run",
				"app.kubernetes.io/managed-by": "codex-k8s-worker",
				metadataLabelRunID:             spec.RunID,
				metadataLabelProjectID:         sanitizeLabel(spec.ProjectID),
			},
			Annotations: map[string]string{
				metadataAnnotationCorrelationID: spec.CorrelationID,
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
						metadataLabelRunID:       spec.RunID,
					},
				},
				Spec: podSpec,
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

// buildRESTConfig resolves Kubernetes REST config from explicit kubeconfig, in-cluster env, or default kubeconfig.
// sanitizeLabel converts arbitrary string to Kubernetes label-safe value.
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

// hasTerminalWaitingReason marks waiting container reasons that should fail run reconciliation early.
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
