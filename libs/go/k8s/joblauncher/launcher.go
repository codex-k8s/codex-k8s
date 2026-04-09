package joblauncher

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/kodex/libs/go/domain/agent"
	webhookdomain "github.com/codex-k8s/kodex/libs/go/domain/webhook"
	"github.com/codex-k8s/kodex/libs/go/k8s/clientcfg"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var nonDNSLabel = regexp.MustCompile(`[^a-z0-9-]`)

const (
	runWorkloadAppName        = "kodex-run"
	runContainerName          = "run"
	aiRepairKeepaliveName     = "keepalive"
	aiRepairComponentLabelVal = "ai-repair"
	discussionComponentLabel  = "discussion"
	workerAppName             = "kodex"
	workerComponentLabel      = "worker"
	runRepoCacheVolumeName    = "repo-cache"
	runRepoCacheClaimName     = "kodex-repo-cache"
	runRepoCacheMountPath     = "/workspace"
)

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
	// JobImage overrides default run image for this specific run when set.
	JobImage string
	// RuntimeMode controls run profile in Kubernetes namespace.
	RuntimeMode agentdomain.RuntimeMode
	// Namespace is preferred namespace for this run.
	Namespace string
	// RuntimeTargetEnv is effective runtime target environment (ai/production).
	RuntimeTargetEnv string
	// RuntimeBuildRef is effective source ref/sha used for runtime execution.
	RuntimeBuildRef string
	// RuntimeAccessProfile is effective Kubernetes access profile for the run pod.
	RuntimeAccessProfile agentdomain.RuntimeAccessProfile
	// ControlPlaneGRPCTarget is control-plane gRPC endpoint for run callbacks.
	ControlPlaneGRPCTarget string
	// MCPBaseURL is control-plane MCP StreamableHTTP endpoint for run pod.
	MCPBaseURL string
	// MCPBearerToken is short-lived token bound to run and used for MCP auth.
	MCPBearerToken string
	// QualityGovernanceEnabled passes change-governance feature state into run pod env.
	QualityGovernanceEnabled bool
	// RepositoryFullName is repository slug in owner/name format.
	RepositoryFullName string
	// IssueNumber is issue number for deterministic branch policy.
	IssueNumber int64
	// TriggerKind defines run stage source (`run:*` catalog, e.g. `dev`, `vision`, `plan_revise`).
	TriggerKind string
	// TriggerLabel is original label that created this run.
	TriggerLabel string
	// DiscussionMode enables comment-only discussion flow without PR/push.
	DiscussionMode bool
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
	// PromptTemplateKind is effective prompt kind (`work`/`revise`/`discussion`).
	PromptTemplateKind string
	// PromptTemplateSource is effective prompt source (`repo_seed` for Day4 baseline).
	PromptTemplateSource string
	// PromptTemplateLocale is effective prompt locale.
	PromptTemplateLocale string
	// BaseBranch is base branch for PR flow.
	BaseBranch string
	// OpenAIAPIKey is passed to run pod for codex login.
	OpenAIAPIKey string
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
	// ServiceAccountName overrides pod service account for this run workload.
	ServiceAccountName string
}

// NamespaceSpec defines runtime namespace metadata.
type NamespaceSpec struct {
	// RunID identifies run owning namespace lifecycle.
	RunID string
	// ProjectID identifies project scope for namespace metadata.
	ProjectID string
	// IssueNumber identifies issue/pr thread for revise namespace reuse.
	IssueNumber int64
	// AgentKey identifies role for namespace ttl-by-role policy and revise reuse.
	AgentKey string
	// CorrelationID links namespace events to webhook flow.
	CorrelationID string
	// RuntimeMode controls whether namespace should be managed.
	RuntimeMode agentdomain.RuntimeMode
	// Namespace is target namespace name.
	Namespace string
	// AccessProfile defines Kubernetes access policy prepared for full-env runs.
	AccessProfile agentdomain.RuntimeAccessProfile
	// LeaseTTL keeps role-based namespace retention duration.
	LeaseTTL time.Duration
	// LeaseExpiresAt pins effective lease expiration timestamp when already resolved by caller.
	LeaseExpiresAt time.Time
}

// NamespaceEnsureResult reports whether namespace was newly created or reused.
type NamespaceEnsureResult struct {
	Created        bool
	Reused         bool
	LeaseExpiresAt time.Time
}

// NamespaceReuseLookup resolves one active namespace lease for project/issue/agent tuple.
type NamespaceReuseLookup struct {
	ProjectID   string
	IssueNumber int64
	AgentKey    string
	Now         time.Time
}

// NamespaceReuseResult describes one active reusable namespace lease.
type NamespaceReuseResult struct {
	Namespace string
	ExpiresAt time.Time
}

// ManagedNamespaceState captures labels and annotations required for cleanup guardrails.
type ManagedNamespaceState struct {
	Namespace       string
	RunID           string
	ProjectID       string
	CorrelationID   string
	RuntimeMode     agentdomain.RuntimeMode
	RuntimeModeRaw  string
	CreatedAt       time.Time
	LeaseTTL        time.Duration
	LeaseExpiresAt  time.Time
	LeaseExpiresRaw string
}

// ManagedNamespaceListParams configures managed namespace listing for cleanup reconciliation.
type ManagedNamespaceListParams struct {
	// NamespacePrefix is the primary run namespace prefix configured for issue-scoped runtimes.
	// Cleanup also keeps known platform slot namespace prefixes in scope as an additional guardrail.
	NamespacePrefix string
}

// NamespaceWorkloadState summarizes active workload objects inside one managed namespace.
type NamespaceWorkloadState struct {
	ActivePods         []string
	ActiveJobs         []string
	ActiveCronJobs     []string
	ActiveDeployments  []string
	ActiveStatefulSets []string
	ActiveDaemonSets   []string
	ActiveReplicaSets  []string
}

// HasActiveWorkloads reports whether namespace still contains running workload objects.
func (s NamespaceWorkloadState) HasActiveWorkloads() bool {
	return len(s.ActivePods) > 0 ||
		len(s.ActiveJobs) > 0 ||
		len(s.ActiveCronJobs) > 0 ||
		len(s.ActiveDeployments) > 0 ||
		len(s.ActiveStatefulSets) > 0 ||
		len(s.ActiveDaemonSets) > 0 ||
		len(s.ActiveReplicaSets) > 0
}

// Details returns deterministic workload diagnostics for logs and audit payloads.
func (s NamespaceWorkloadState) Details() []string {
	details := make([]string, 0, 7)
	details = appendWorkloadDetails(details, "pods", s.ActivePods)
	details = appendWorkloadDetails(details, "jobs", s.ActiveJobs)
	details = appendWorkloadDetails(details, "cronjobs", s.ActiveCronJobs)
	details = appendWorkloadDetails(details, "deployments", s.ActiveDeployments)
	details = appendWorkloadDetails(details, "statefulsets", s.ActiveStatefulSets)
	details = appendWorkloadDetails(details, "daemonsets", s.ActiveDaemonSets)
	details = appendWorkloadDetails(details, "replicasets", s.ActiveReplicaSets)
	return details
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
	// RunReadOnlyServiceAccountName defines service account for production read-only run jobs.
	RunReadOnlyServiceAccountName string
	// RunReadOnlyRoleName defines read-only RBAC role name for production read-only run jobs.
	RunReadOnlyRoleName string
	// RunReadOnlyRoleBindingName defines read-only RBAC role binding name for production read-only run jobs.
	RunReadOnlyRoleBindingName string
	// RunResourceQuotaName defines resource quota object name in runtime namespaces.
	RunResourceQuotaName string
	// RunLimitRangeName defines limit range object name in runtime namespaces.
	RunLimitRangeName string
	// RunCredentialsSecretName defines secret object with run credentials in runtime namespaces.
	RunCredentialsSecretName string
	// RunResourceQuotaPods defines max pod count in runtime namespace.
	RunResourceQuotaPods int64
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
		cfg.Command = "/usr/local/bin/kodex-agent-runner"
	}
	if cfg.TTLSeconds <= 0 {
		cfg.TTLSeconds = 600
	}
	if cfg.ActiveDeadlineSeconds <= 0 {
		cfg.ActiveDeadlineSeconds = 18000
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
	if cfg.RunReadOnlyServiceAccountName == "" {
		cfg.RunReadOnlyServiceAccountName = "codex-runner-readonly"
	}
	if cfg.RunReadOnlyRoleName == "" {
		cfg.RunReadOnlyRoleName = "codex-runner-readonly"
	}
	if cfg.RunReadOnlyRoleBindingName == "" {
		cfg.RunReadOnlyRoleBindingName = "codex-runner-readonly"
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

	return &Launcher{cfg: cfg, client: client}
}

func appendWorkloadDetails(dst []string, kind string, names []string) []string {
	if len(names) == 0 {
		return dst
	}
	return append(dst, kind+"="+strings.Join(names, ","))
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

// ListWorkerPodNames returns worker pod names currently visible in one namespace.
func (l *Launcher) ListWorkerPodNames(ctx context.Context, namespace string) ([]string, error) {
	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace == "" {
		targetNamespace = l.cfg.Namespace
	}

	pods, err := l.client.CoreV1().Pods(targetNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/component=%s", workerAppName, workerComponentLabel),
	})
	if err != nil {
		return nil, fmt.Errorf("list worker pods in namespace %s: %w", targetNamespace, err)
	}

	result := make([]string, 0, len(pods.Items))
	for _, item := range pods.Items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		result = append(result, name)
	}
	sort.Strings(result)
	return result, nil
}

// Launch creates Kubernetes Job or returns existing one when already present.
func (l *Launcher) Launch(ctx context.Context, spec JobSpec) (JobRef, error) {
	ref := l.JobRef(spec.RunID, spec.Namespace)
	jobImage := strings.TrimSpace(spec.JobImage)
	if jobImage == "" {
		jobImage = l.cfg.Image
	}

	if isAIRepairTriggerKind(spec.TriggerKind) {
		return l.launchAIRepairPod(ctx, ref, spec, jobImage)
	}
	if spec.DiscussionMode {
		return l.launchRunPod(ctx, ref, spec, jobImage, discussionComponentLabel, false)
	}

	container := buildRunContainer(spec, jobImage, l.cfg.Command)
	if shouldMountRepoCache(spec) {
		container = withRepoCacheVolumeMount(container)
	}

	podSpec := corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyNever,
		Containers:    []corev1.Container{container},
	}
	if shouldMountRepoCache(spec) {
		podSpec.Volumes = append(podSpec.Volumes, repoCacheVolume())
	}

	serviceAccountName := strings.TrimSpace(spec.ServiceAccountName)
	if serviceAccountName == "" && spec.RuntimeMode == agentdomain.RuntimeModeFullEnv {
		serviceAccountName = l.cfg.RunServiceAccountName
	}
	if serviceAccountName != "" {
		podSpec.ServiceAccountName = serviceAccountName
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ref.Name,
			Namespace: ref.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       runWorkloadAppName,
				"app.kubernetes.io/managed-by": "kodex-worker",
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
						"app.kubernetes.io/name": runWorkloadAppName,
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

func (l *Launcher) launchAIRepairPod(ctx context.Context, ref JobRef, spec JobSpec, jobImage string) (JobRef, error) {
	return l.launchRunPod(ctx, ref, spec, jobImage, aiRepairComponentLabelVal, true)
}

func (l *Launcher) launchRunPod(ctx context.Context, ref JobRef, spec JobSpec, jobImage string, componentLabel string, withKeepalive bool) (JobRef, error) {
	runContainer := buildRunContainer(spec, jobImage, l.cfg.Command)
	containers := []corev1.Container{runContainer}
	if shouldMountRepoCache(spec) {
		runContainer = withRepoCacheVolumeMount(runContainer)
		containers[0] = runContainer
	}
	if withKeepalive {
		keepaliveContainer := corev1.Container{
			Name:    aiRepairKeepaliveName,
			Image:   jobImage,
			Command: []string{"/bin/sh", "-c", "trap : TERM INT; while true; do sleep 3600; done"},
		}
		if shouldMountRepoCache(spec) {
			keepaliveContainer = withRepoCacheVolumeMount(keepaliveContainer)
		}
		containers = append(containers, keepaliveContainer)
	}

	podSpec := corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyNever,
		Containers:    containers,
	}
	if shouldMountRepoCache(spec) {
		podSpec.Volumes = append(podSpec.Volumes, repoCacheVolume())
	}
	serviceAccountName := strings.TrimSpace(spec.ServiceAccountName)
	if serviceAccountName != "" {
		podSpec.ServiceAccountName = serviceAccountName
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ref.Name,
			Namespace: ref.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       runWorkloadAppName,
				"app.kubernetes.io/component":  strings.TrimSpace(componentLabel),
				"app.kubernetes.io/managed-by": "kodex-worker",
				metadataLabelRunID:             spec.RunID,
				metadataLabelProjectID:         sanitizeLabel(spec.ProjectID),
			},
			Annotations: map[string]string{
				metadataAnnotationCorrelationID: spec.CorrelationID,
			},
		},
		Spec: podSpec,
	}

	if _, err := l.client.CoreV1().Pods(ref.Namespace).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return ref, nil
		}
		return JobRef{}, fmt.Errorf("create kubernetes pod %s/%s: %w", ref.Namespace, ref.Name, err)
	}

	return ref, nil
}

func buildRunContainer(spec JobSpec, image string, command string) corev1.Container {
	return corev1.Container{
		Name:    runContainerName,
		Image:   image,
		Command: []string{"/bin/sh", "-c", command},
		Env:     buildRunContainerEnv(spec),
	}
}

func buildRunContainerEnv(spec JobSpec) []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "KODEX_RUN_ID", Value: spec.RunID},
		{Name: "KODEX_CORRELATION_ID", Value: spec.CorrelationID},
		{Name: "KODEX_PROJECT_ID", Value: spec.ProjectID},
		{Name: "KODEX_SLOT_NO", Value: fmt.Sprintf("%d", spec.SlotNo)},
		{Name: "KODEX_RUNTIME_MODE", Value: string(spec.RuntimeMode)},
		{Name: "KODEX_RUNTIME_TARGET_ENV", Value: strings.TrimSpace(spec.RuntimeTargetEnv)},
		{Name: "KODEX_RUNTIME_BUILD_REF", Value: strings.TrimSpace(spec.RuntimeBuildRef)},
		{Name: "KODEX_RUNTIME_ACCESS_PROFILE", Value: strings.TrimSpace(string(spec.RuntimeAccessProfile))},
		{Name: "KODEX_CONTROL_PLANE_GRPC_TARGET", Value: strings.TrimSpace(spec.ControlPlaneGRPCTarget)},
		{Name: "KODEX_MCP_BASE_URL", Value: strings.TrimSpace(spec.MCPBaseURL)},
		{Name: "KODEX_MCP_BEARER_TOKEN", Value: strings.TrimSpace(spec.MCPBearerToken)},
		{Name: "KODEX_QUALITY_GOVERNANCE_ENABLED", Value: fmt.Sprintf("%t", spec.QualityGovernanceEnabled)},
		{Name: "KODEX_REPOSITORY_FULL_NAME", Value: strings.TrimSpace(spec.RepositoryFullName)},
		{Name: "KODEX_ISSUE_NUMBER", Value: fmt.Sprintf("%d", spec.IssueNumber)},
		{Name: "KODEX_RUN_TRIGGER_KIND", Value: strings.TrimSpace(spec.TriggerKind)},
		{Name: "KODEX_RUN_TRIGGER_LABEL", Value: strings.TrimSpace(spec.TriggerLabel)},
		{Name: "KODEX_DISCUSSION_MODE", Value: fmt.Sprintf("%t", spec.DiscussionMode)},
		{Name: "KODEX_RUN_TARGET_BRANCH", Value: strings.TrimSpace(spec.TargetBranch)},
		{Name: "KODEX_EXISTING_PR_NUMBER", Value: fmt.Sprintf("%d", spec.ExistingPRNumber)},
		{Name: "KODEX_AGENT_KEY", Value: strings.TrimSpace(spec.AgentKey)},
		{Name: "KODEX_AGENT_MODEL", Value: strings.TrimSpace(spec.AgentModel)},
		{Name: "KODEX_AGENT_REASONING_EFFORT", Value: strings.TrimSpace(spec.AgentReasoningEffort)},
		{Name: "KODEX_PROMPT_TEMPLATE_KIND", Value: strings.TrimSpace(spec.PromptTemplateKind)},
		{Name: "KODEX_PROMPT_TEMPLATE_SOURCE", Value: strings.TrimSpace(spec.PromptTemplateSource)},
		{Name: "KODEX_PROMPT_TEMPLATE_LOCALE", Value: strings.TrimSpace(spec.PromptTemplateLocale)},
		{Name: "KODEX_STATE_IN_REVIEW_LABEL", Value: strings.TrimSpace(spec.StateInReviewLabel)},
		{Name: "KODEX_AGENT_BASE_BRANCH", Value: strings.TrimSpace(spec.BaseBranch)},
		{Name: "KODEX_OPENAI_API_KEY", Value: strings.TrimSpace(spec.OpenAIAPIKey)},
		{Name: "KODEX_CONTEXT7_API_KEY", Value: strings.TrimSpace(spec.Context7APIKey)},
		{Name: "KODEX_AGENT_DISPLAY_NAME", Value: strings.TrimSpace(spec.AgentDisplayName)},
		{Name: "KODEX_GIT_BOT_TOKEN", Value: strings.TrimSpace(spec.GitBotToken)},
		{Name: "KODEX_GIT_BOT_USERNAME", Value: strings.TrimSpace(spec.GitBotUsername)},
		{Name: "KODEX_GIT_BOT_MAIL", Value: strings.TrimSpace(spec.GitBotMail)},
	}
}

func shouldMountRepoCache(spec JobSpec) bool {
	if isAIRepairTriggerKind(spec.TriggerKind) {
		return true
	}
	return spec.RuntimeMode == agentdomain.RuntimeModeFullEnv
}

func withRepoCacheVolumeMount(container corev1.Container) corev1.Container {
	container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
		Name:      runRepoCacheVolumeName,
		MountPath: runRepoCacheMountPath,
	})
	return container
}

func repoCacheVolume() corev1.Volume {
	return corev1.Volume{
		Name: runRepoCacheVolumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: runRepoCacheClaimName,
			},
		},
	}
}

// Status returns current Job state by Job status fields.
func (l *Launcher) Status(ctx context.Context, ref JobRef) (JobState, error) {
	job, err := l.client.BatchV1().Jobs(ref.Namespace).Get(ctx, ref.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return l.statusByPod(ctx, ref)
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

func (l *Launcher) statusByPod(ctx context.Context, ref JobRef) (JobState, error) {
	pod, err := l.client.CoreV1().Pods(ref.Namespace).Get(ctx, ref.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return JobStateNotFound, nil
		}
		return "", fmt.Errorf("get kubernetes pod %s/%s: %w", ref.Namespace, ref.Name, err)
	}

	if runStatus, ok := findContainerStatusByName(pod.Status.ContainerStatuses, runContainerName); ok {
		if runStatus.State.Terminated != nil {
			if runStatus.State.Terminated.ExitCode == 0 {
				return JobStateSucceeded, nil
			}
			return JobStateFailed, nil
		}
		if hasTerminalWaitingReason([]corev1.ContainerStatus{runStatus}) {
			return JobStateFailed, nil
		}
		if runStatus.State.Running != nil {
			return JobStateRunning, nil
		}
		if runStatus.State.Waiting != nil {
			return JobStatePending, nil
		}
	}

	if hasTerminalWaitingReason(pod.Status.InitContainerStatuses) || hasTerminalWaitingReason(pod.Status.ContainerStatuses) {
		return JobStateFailed, nil
	}

	switch pod.Status.Phase {
	case corev1.PodSucceeded:
		return JobStateSucceeded, nil
	case corev1.PodFailed:
		return JobStateFailed, nil
	case corev1.PodRunning:
		return JobStateRunning, nil
	case corev1.PodPending:
		return JobStatePending, nil
	default:
		return JobStatePending, nil
	}
}

func findContainerStatusByName(statuses []corev1.ContainerStatus, name string) (corev1.ContainerStatus, bool) {
	for _, item := range statuses {
		if strings.TrimSpace(item.Name) == name {
			return item, true
		}
	}
	return corev1.ContainerStatus{}, false
}

// FindRunJobRefByRunID resolves run Kubernetes Job reference by run id label across namespaces.
func (l *Launcher) FindRunJobRefByRunID(ctx context.Context, runID string) (JobRef, bool, error) {
	targetRunID := strings.TrimSpace(runID)
	if targetRunID == "" {
		return JobRef{}, false, fmt.Errorf("run id is required")
	}

	selector := fmt.Sprintf("%s=%s,app.kubernetes.io/name=%s", metadataLabelRunID, targetRunID, runWorkloadAppName)
	jobs, err := l.client.BatchV1().Jobs(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return JobRef{}, false, fmt.Errorf("list kubernetes jobs by run id %s: %w", targetRunID, err)
	}
	expectedName := BuildRunJobName(targetRunID)
	candidates := make([]JobRef, 0, len(jobs.Items))
	for _, item := range jobs.Items {
		if strings.TrimSpace(item.Name) == expectedName {
			candidates = append(candidates, JobRef{
				Namespace: strings.TrimSpace(item.Namespace),
				Name:      strings.TrimSpace(item.Name),
			})
		}
	}
	if len(candidates) == 0 {
		for _, item := range jobs.Items {
			candidates = append(candidates, JobRef{
				Namespace: strings.TrimSpace(item.Namespace),
				Name:      strings.TrimSpace(item.Name),
			})
		}
	}

	if len(candidates) == 0 {
		pods, podErr := l.client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
			LabelSelector: selector,
		})
		if podErr != nil {
			return JobRef{}, false, fmt.Errorf("list kubernetes pods by run id %s: %w", targetRunID, podErr)
		}
		for _, item := range pods.Items {
			name := strings.TrimSpace(item.Name)
			if name == expectedName {
				candidates = append(candidates, JobRef{
					Namespace: strings.TrimSpace(item.Namespace),
					Name:      name,
				})
			}
		}
		if len(candidates) == 0 {
			for _, item := range pods.Items {
				candidates = append(candidates, JobRef{
					Namespace: strings.TrimSpace(item.Namespace),
					Name:      strings.TrimSpace(item.Name),
				})
			}
		}
	}
	if len(candidates) == 0 {
		return JobRef{}, false, nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		left := candidates[i].Namespace + "/" + candidates[i].Name
		right := candidates[j].Namespace + "/" + candidates[j].Name
		return left < right
	})

	return candidates[0], true, nil
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

	name := "kodex-run-" + normalized
	if len(name) > 63 {
		name = name[:63]
	}
	name = strings.TrimRight(name, "-")
	if name == "" {
		return "kodex-run"
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

func isAIRepairTriggerKind(value string) bool {
	return webhookdomain.NormalizeTriggerKind(strings.TrimSpace(value)) == webhookdomain.TriggerKindAIRepair
}
