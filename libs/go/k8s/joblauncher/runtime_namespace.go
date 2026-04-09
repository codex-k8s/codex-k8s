package joblauncher

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/kodex/libs/go/domain/agent"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	runNamespaceManagedByLabel      = metadataLabelManagedBy
	runNamespacePurposeLabel        = metadataLabelNamespacePurpose
	runNamespaceRuntimeModeLabel    = metadataLabelRuntimeMode
	runNamespaceProjectIDLabel      = metadataLabelProjectID
	runNamespaceRunIDLabel          = metadataLabelRunID
	runNamespaceIssueNumberLabel    = metadataLabelIssueNumber
	runNamespaceAgentKeyLabel       = metadataLabelAgentKey
	runNamespaceCorrelationAnnotKey = metadataAnnotationCorrelationID
	runNamespaceLeaseTTLAnnotKey    = metadataAnnotationNamespaceTTL
	runNamespaceLeaseExpAnnotKey    = metadataAnnotationNamespaceExp
	runNamespaceLeaseUpdAnnotKey    = metadataAnnotationNamespaceUpd

	runNamespaceManagedByValue = "kodex-worker"
	runNamespacePurposeValue   = "run"

	runNamespaceCleanupSlotPrefix = "kodex-dev-"
)

// EnsureNamespace prepares baseline runtime resources for managed run execution.
func (l *Launcher) EnsureNamespace(ctx context.Context, spec NamespaceSpec) (NamespaceEnsureResult, error) {
	switch spec.RuntimeMode {
	case agentdomain.RuntimeModeFullEnv, agentdomain.RuntimeModeCodeOnly:
	default:
		return NamespaceEnsureResult{}, nil
	}
	namespace := strings.TrimSpace(spec.Namespace)
	if namespace == "" {
		return NamespaceEnsureResult{}, fmt.Errorf("runtime namespace is required for managed run")
	}

	ensureResult, err := l.ensureNamespaceObject(ctx, spec)
	if err != nil {
		return NamespaceEnsureResult{}, fmt.Errorf("ensure namespace %s: %w", namespace, err)
	}
	if spec.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		return ensureResult, nil
	}
	if _, err := l.EnsureAccessProfile(ctx, namespace, spec.AccessProfile); err != nil {
		return NamespaceEnsureResult{}, fmt.Errorf("ensure access profile in namespace %s: %w", namespace, err)
	}
	if err := l.ensureResourceQuota(ctx, namespace); err != nil {
		return NamespaceEnsureResult{}, fmt.Errorf("ensure resource quota in namespace %s: %w", namespace, err)
	}
	if err := l.ensureLimitRange(ctx, namespace); err != nil {
		return NamespaceEnsureResult{}, fmt.Errorf("ensure limit range in namespace %s: %w", namespace, err)
	}
	return ensureResult, nil
}

// CleanupNamespace removes managed runtime namespace after run completion.
func (l *Launcher) CleanupNamespace(ctx context.Context, spec NamespaceSpec) error {
	switch spec.RuntimeMode {
	case agentdomain.RuntimeModeFullEnv, agentdomain.RuntimeModeCodeOnly:
	default:
		return nil
	}

	namespace := strings.TrimSpace(spec.Namespace)
	if namespace == "" {
		return nil
	}
	_, err := l.DeleteManagedNamespace(ctx, namespace)
	if err != nil {
		return err
	}
	return nil
}

// FindReusableNamespace resolves one active namespace lease for project/issue/agent tuple.
func (l *Launcher) FindReusableNamespace(ctx context.Context, lookup NamespaceReuseLookup) (NamespaceReuseResult, bool, error) {
	projectID := sanitizeLabel(strings.TrimSpace(lookup.ProjectID))
	if projectID == "" || projectID == "unknown" {
		return NamespaceReuseResult{}, false, fmt.Errorf("project id is required")
	}
	if lookup.IssueNumber <= 0 {
		return NamespaceReuseResult{}, false, fmt.Errorf("issue number must be positive")
	}
	agentKey := sanitizeLabel(strings.TrimSpace(lookup.AgentKey))
	if agentKey == "" || agentKey == "unknown" {
		return NamespaceReuseResult{}, false, fmt.Errorf("agent key is required")
	}

	now := lookup.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	items, err := l.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf(
			"%s=%s,%s=%s,%s=%s,%s=%s,%s=%d,%s=%s",
			runNamespaceManagedByLabel,
			runNamespaceManagedByValue,
			runNamespacePurposeLabel,
			runNamespacePurposeValue,
			runNamespaceRuntimeModeLabel,
			string(agentdomain.RuntimeModeFullEnv),
			runNamespaceProjectIDLabel,
			projectID,
			runNamespaceIssueNumberLabel,
			lookup.IssueNumber,
			runNamespaceAgentKeyLabel,
			agentKey,
		),
	})
	if err != nil {
		return NamespaceReuseResult{}, false, fmt.Errorf("list reusable run namespaces: %w", err)
	}

	best := NamespaceReuseResult{}
	found := false
	for _, item := range items.Items {
		if item.DeletionTimestamp != nil {
			continue
		}
		expiresAt, ok := parseNamespaceLeaseExpiresAt(item.Annotations)
		if !ok || !expiresAt.After(now) {
			continue
		}
		candidate := NamespaceReuseResult{
			Namespace: strings.TrimSpace(item.Name),
			ExpiresAt: expiresAt,
		}
		if !found || candidate.ExpiresAt.After(best.ExpiresAt) {
			best = candidate
			found = true
		}
	}
	if !found {
		return NamespaceReuseResult{}, false, nil
	}
	return best, true, nil
}

// ListManagedRunNamespaces returns worker-managed runtime namespaces with deterministic ordering.
// Cleanup keeps the configured issue-run prefix and known platform slot prefixes in scope.
func (l *Launcher) ListManagedRunNamespaces(ctx context.Context, params ManagedNamespaceListParams) ([]ManagedNamespaceState, error) {
	prefix := strings.TrimSpace(params.NamespacePrefix)
	if prefix != "" {
		prefix = sanitizeLabel(prefix)
		if prefix == "unknown" {
			prefix = ""
		}
	}
	items, err := l.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf(
			"%s=%s,%s=%s",
			runNamespaceManagedByLabel,
			runNamespaceManagedByValue,
			runNamespacePurposeLabel,
			runNamespacePurposeValue,
		),
	})
	if err != nil {
		return nil, fmt.Errorf("list managed run namespaces: %w", err)
	}
	sort.Slice(items.Items, func(i, j int) bool {
		return items.Items[i].Name < items.Items[j].Name
	})

	states := make([]ManagedNamespaceState, 0, len(items.Items))
	for _, item := range items.Items {
		namespace := strings.TrimSpace(item.Name)
		if namespace == "" || namespace == strings.TrimSpace(l.cfg.Namespace) {
			continue
		}
		if item.DeletionTimestamp != nil {
			continue
		}
		if !managedRunNamespaceMatchesCleanupScope(namespace, prefix) {
			continue
		}
		runtimeModeRaw := strings.TrimSpace(item.Labels[runNamespaceRuntimeModeLabel])
		if runtimeModeRaw == "" {
			continue
		}
		expiresAt, _ := parseNamespaceLeaseExpiresAt(item.Annotations)
		states = append(states, ManagedNamespaceState{
			Namespace:       namespace,
			RunID:           strings.TrimSpace(item.Labels[runNamespaceRunIDLabel]),
			ProjectID:       strings.TrimSpace(item.Labels[runNamespaceProjectIDLabel]),
			CorrelationID:   strings.TrimSpace(item.Annotations[runNamespaceCorrelationAnnotKey]),
			RuntimeMode:     agentdomain.ParseRuntimeMode(runtimeModeRaw),
			RuntimeModeRaw:  runtimeModeRaw,
			CreatedAt:       item.CreationTimestamp.Time.UTC(),
			LeaseTTL:        parseNamespaceLeaseTTL(item.Annotations),
			LeaseExpiresAt:  expiresAt,
			LeaseExpiresRaw: strings.TrimSpace(item.Annotations[runNamespaceLeaseExpAnnotKey]),
		})
	}
	return states, nil
}

// InspectNamespaceWorkloads returns active workload objects inside one managed namespace.
func (l *Launcher) InspectNamespaceWorkloads(ctx context.Context, namespace string) (NamespaceWorkloadState, error) {
	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace == "" {
		return NamespaceWorkloadState{}, fmt.Errorf("namespace is required")
	}

	pods, err := l.client.CoreV1().Pods(targetNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return NamespaceWorkloadState{}, fmt.Errorf("list pods in namespace %s: %w", targetNamespace, err)
	}
	jobs, err := l.client.BatchV1().Jobs(targetNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return NamespaceWorkloadState{}, fmt.Errorf("list jobs in namespace %s: %w", targetNamespace, err)
	}
	cronJobs, err := l.client.BatchV1().CronJobs(targetNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return NamespaceWorkloadState{}, fmt.Errorf("list cronjobs in namespace %s: %w", targetNamespace, err)
	}
	deployments, err := l.client.AppsV1().Deployments(targetNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return NamespaceWorkloadState{}, fmt.Errorf("list deployments in namespace %s: %w", targetNamespace, err)
	}
	statefulSets, err := l.client.AppsV1().StatefulSets(targetNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return NamespaceWorkloadState{}, fmt.Errorf("list statefulsets in namespace %s: %w", targetNamespace, err)
	}
	daemonSets, err := l.client.AppsV1().DaemonSets(targetNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return NamespaceWorkloadState{}, fmt.Errorf("list daemonsets in namespace %s: %w", targetNamespace, err)
	}
	replicaSets, err := l.client.AppsV1().ReplicaSets(targetNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return NamespaceWorkloadState{}, fmt.Errorf("list replicasets in namespace %s: %w", targetNamespace, err)
	}

	return NamespaceWorkloadState{
		ActivePods:         activePodNames(pods.Items),
		ActiveJobs:         activeJobNames(jobs.Items),
		ActiveCronJobs:     activeCronJobNames(cronJobs.Items),
		ActiveDeployments:  activeReplicatedWorkloadNames(deployments.Items),
		ActiveStatefulSets: activeReplicatedWorkloadNames(statefulSets.Items),
		ActiveDaemonSets:   activeDaemonSetNames(daemonSets.Items),
		ActiveReplicaSets:  activeReplicatedWorkloadNames(replicaSets.Items),
	}, nil
}

// DeleteManagedNamespace removes one worker-managed run namespace.
func (l *Launcher) DeleteManagedNamespace(ctx context.Context, namespace string) (bool, error) {
	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace == "" {
		return false, nil
	}
	if targetNamespace == strings.TrimSpace(l.cfg.Namespace) {
		return false, nil
	}

	ns, err := l.client.CoreV1().Namespaces().Get(ctx, targetNamespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("get namespace %s: %w", targetNamespace, err)
	}
	if strings.TrimSpace(ns.Labels[runNamespaceManagedByLabel]) != runNamespaceManagedByValue {
		return false, nil
	}
	if strings.TrimSpace(ns.Labels[runNamespacePurposeLabel]) != runNamespacePurposeValue {
		return false, nil
	}
	if strings.TrimSpace(ns.Labels[runNamespaceRuntimeModeLabel]) == "" {
		return false, nil
	}
	if err := l.client.CoreV1().Namespaces().Delete(ctx, targetNamespace, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("delete namespace %s: %w", targetNamespace, err)
	}
	return true, nil
}

type accessProfileSpec struct {
	ServiceAccountName string
	RoleName           string
	RoleBindingName    string
	Rules              []rbacv1.PolicyRule
}

// EnsureAccessProfile prepares ServiceAccount + Role + RoleBinding for one runtime access profile.
func (l *Launcher) EnsureAccessProfile(ctx context.Context, namespace string, profile agentdomain.RuntimeAccessProfile) (string, error) {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return "", fmt.Errorf("namespace is required")
	}

	spec := l.resolveAccessProfileSpec(profile)
	if err := l.ensureServiceAccount(ctx, namespace, spec.ServiceAccountName); err != nil {
		return "", fmt.Errorf("ensure serviceaccount %s: %w", spec.ServiceAccountName, err)
	}
	if err := l.ensureRole(ctx, namespace, spec.RoleName, spec.Rules); err != nil {
		return "", fmt.Errorf("ensure role %s: %w", spec.RoleName, err)
	}
	if err := l.ensureRoleBinding(ctx, namespace, spec.RoleBindingName, spec.RoleName, spec.ServiceAccountName); err != nil {
		return "", fmt.Errorf("ensure rolebinding %s: %w", spec.RoleBindingName, err)
	}
	return spec.ServiceAccountName, nil
}

func (l *Launcher) resolveAccessProfileSpec(profile agentdomain.RuntimeAccessProfile) accessProfileSpec {
	switch agentdomain.ParseRuntimeAccessProfile(string(profile)) {
	case agentdomain.RuntimeAccessProfileProductionReadOnly:
		return accessProfileSpec{
			ServiceAccountName: l.cfg.RunReadOnlyServiceAccountName,
			RoleName:           l.cfg.RunReadOnlyRoleName,
			RoleBindingName:    l.cfg.RunReadOnlyRoleBindingName,
			Rules:              productionReadOnlyRoleRules(),
		}
	default:
		return accessProfileSpec{
			ServiceAccountName: l.cfg.RunServiceAccountName,
			RoleName:           l.cfg.RunRoleName,
			RoleBindingName:    l.cfg.RunRoleBindingName,
			Rules:              candidateRoleRules(),
		}
	}
}

// ensureNamespaceObject upserts namespace metadata required for managed runtime namespaces.
func (l *Launcher) ensureNamespaceObject(ctx context.Context, spec NamespaceSpec) (NamespaceEnsureResult, error) {
	namespace := strings.TrimSpace(spec.Namespace)
	leaseExpiresAt := resolveNamespaceLeaseExpiresAt(spec)
	leaseTTL := resolveNamespaceLeaseTTL(spec)
	leaseUpdatedAt := time.Now().UTC()

	labels := map[string]string{
		runNamespaceManagedByLabel:   runNamespaceManagedByValue,
		runNamespacePurposeLabel:     runNamespacePurposeValue,
		runNamespaceRuntimeModeLabel: string(spec.RuntimeMode),
		runNamespaceRunIDLabel:       sanitizeLabel(spec.RunID),
		runNamespaceProjectIDLabel:   sanitizeLabel(spec.ProjectID),
	}
	if spec.IssueNumber > 0 {
		labels[runNamespaceIssueNumberLabel] = strconv.FormatInt(spec.IssueNumber, 10)
	}
	agentKey := sanitizeLabel(spec.AgentKey)
	if agentKey != "" && agentKey != "unknown" {
		labels[runNamespaceAgentKeyLabel] = agentKey
	}
	projectLabel := sanitizeLabel(spec.ProjectID)
	if projectLabel != "unknown" {
		labels[runNamespaceProjectIDLabel] = projectLabel
	}
	annotations := map[string]string{
		runNamespaceCorrelationAnnotKey: spec.CorrelationID,
		runNamespaceLeaseTTLAnnotKey:    leaseTTL.String(),
		runNamespaceLeaseExpAnnotKey:    leaseExpiresAt.Format(time.RFC3339),
		runNamespaceLeaseUpdAnnotKey:    leaseUpdatedAt.Format(time.RFC3339),
	}

	existing, err := l.client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return NamespaceEnsureResult{}, fmt.Errorf("get namespace: %w", err)
		}
		_, createErr := l.client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        namespace,
				Labels:      labels,
				Annotations: annotations,
			},
		}, metav1.CreateOptions{})
		if createErr != nil && !apierrors.IsAlreadyExists(createErr) {
			return NamespaceEnsureResult{}, fmt.Errorf("create namespace: %w", createErr)
		}
		return NamespaceEnsureResult{
			Created:        true,
			Reused:         false,
			LeaseExpiresAt: leaseExpiresAt,
		}, nil
	}
	if existing.DeletionTimestamp != nil {
		return NamespaceEnsureResult{}, fmt.Errorf("namespace %s is terminating", namespace)
	}

	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}
	for key, value := range labels {
		existing.Labels[key] = value
	}
	if existing.Annotations == nil {
		existing.Annotations = map[string]string{}
	}
	for key, value := range annotations {
		existing.Annotations[key] = value
	}

	_, err = l.client.CoreV1().Namespaces().Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return NamespaceEnsureResult{}, fmt.Errorf("update namespace: %w", err)
	}
	return NamespaceEnsureResult{
		Created:        false,
		Reused:         true,
		LeaseExpiresAt: leaseExpiresAt,
	}, nil
}

func resolveNamespaceLeaseTTL(spec NamespaceSpec) time.Duration {
	if spec.LeaseTTL > 0 {
		return spec.LeaseTTL
	}
	return 24 * time.Hour
}

func resolveNamespaceLeaseExpiresAt(spec NamespaceSpec) time.Time {
	if !spec.LeaseExpiresAt.IsZero() {
		return spec.LeaseExpiresAt.UTC()
	}
	return time.Now().UTC().Add(resolveNamespaceLeaseTTL(spec))
}

func parseNamespaceLeaseExpiresAt(annotations map[string]string) (time.Time, bool) {
	if len(annotations) == 0 {
		return time.Time{}, false
	}
	raw := strings.TrimSpace(annotations[runNamespaceLeaseExpAnnotKey])
	if raw == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func parseNamespaceLeaseTTL(annotations map[string]string) time.Duration {
	if len(annotations) == 0 {
		return 0
	}
	raw := strings.TrimSpace(annotations[runNamespaceLeaseTTLAnnotKey])
	if raw == "" {
		return 0
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil || parsed <= 0 {
		return 0
	}
	return parsed
}

func managedRunNamespaceMatchesCleanupScope(namespace string, configuredPrefix string) bool {
	targetNamespace := strings.TrimSpace(namespace)
	if targetNamespace == "" {
		return false
	}
	if configuredPrefix == "" {
		return true
	}
	return strings.HasPrefix(targetNamespace, configuredPrefix) || strings.HasPrefix(targetNamespace, runNamespaceCleanupSlotPrefix)
}

func activePodNames(items []corev1.Pod) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		if item.DeletionTimestamp != nil {
			continue
		}
		switch item.Status.Phase {
		case corev1.PodSucceeded, corev1.PodFailed:
			continue
		}
		names = append(names, strings.TrimSpace(item.Name))
	}
	sort.Strings(names)
	return names
}

func activeJobNames(items []batchv1.Job) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		if item.DeletionTimestamp != nil {
			continue
		}
		if jobIsTerminal(item) {
			continue
		}
		names = append(names, strings.TrimSpace(item.Name))
	}
	sort.Strings(names)
	return names
}

func activeCronJobNames(items []batchv1.CronJob) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		if item.DeletionTimestamp != nil {
			continue
		}
		if len(item.Status.Active) > 0 {
			names = append(names, strings.TrimSpace(item.Name))
			continue
		}
		if item.Spec.Suspend != nil && *item.Spec.Suspend {
			continue
		}
		names = append(names, strings.TrimSpace(item.Name))
	}
	sort.Strings(names)
	return names
}

func activeDaemonSetNames(items []appsv1.DaemonSet) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		if item.DeletionTimestamp != nil {
			continue
		}
		if item.Status.DesiredNumberScheduled == 0 &&
			item.Status.CurrentNumberScheduled == 0 &&
			item.Status.NumberReady == 0 {
			continue
		}
		names = append(names, strings.TrimSpace(item.Name))
	}
	sort.Strings(names)
	return names
}

func activeReplicatedWorkloadNames[T any](items []T) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		name, deletionTimestamp, specReplicas, statusReplicas, readyReplicas, ok := replicatedWorkloadStatus(reflect.ValueOf(item))
		if !ok || deletionTimestamp != nil {
			continue
		}
		desired := int32(1)
		if specReplicas != nil {
			desired = *specReplicas
		}
		if desired <= 0 && statusReplicas == 0 && readyReplicas == 0 {
			continue
		}
		names = append(names, strings.TrimSpace(name))
	}
	sort.Strings(names)
	return names
}

func replicatedWorkloadStatus(value reflect.Value) (string, *metav1.Time, *int32, int32, int32, bool) {
	if !value.IsValid() {
		return "", nil, nil, 0, 0, false
	}
	objectMeta := value.FieldByName("ObjectMeta")
	spec := value.FieldByName("Spec")
	status := value.FieldByName("Status")
	if !objectMeta.IsValid() || !spec.IsValid() || !status.IsValid() {
		return "", nil, nil, 0, 0, false
	}

	nameField := objectMeta.FieldByName("Name")
	deletionTimestampField := objectMeta.FieldByName("DeletionTimestamp")
	specReplicasField := spec.FieldByName("Replicas")
	statusReplicasField := status.FieldByName("Replicas")
	readyReplicasField := status.FieldByName("ReadyReplicas")
	if !nameField.IsValid() || !deletionTimestampField.IsValid() || !specReplicasField.IsValid() || !statusReplicasField.IsValid() || !readyReplicasField.IsValid() {
		return "", nil, nil, 0, 0, false
	}

	var deletionTimestamp *metav1.Time
	if !deletionTimestampField.IsNil() {
		value := deletionTimestampField.Interface().(*metav1.Time)
		deletionTimestamp = value
	}

	var specReplicas *int32
	if !specReplicasField.IsNil() {
		value := int32(specReplicasField.Elem().Int())
		specReplicas = &value
	}

	return nameField.String(), deletionTimestamp, specReplicas, int32(statusReplicasField.Int()), int32(readyReplicasField.Int()), true
}

func jobIsTerminal(item batchv1.Job) bool {
	for _, condition := range item.Status.Conditions {
		if condition.Status != corev1.ConditionTrue {
			continue
		}
		if condition.Type == batchv1.JobComplete || condition.Type == batchv1.JobFailed {
			return true
		}
	}
	return false
}

func (l *Launcher) ensureServiceAccount(ctx context.Context, namespace string, name string) error {
	existing, err := l.client.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get serviceaccount %s: %w", name, err)
		}
		_, createErr := l.client.CoreV1().ServiceAccounts(namespace).Create(ctx, &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					runNamespaceManagedByLabel: runNamespaceManagedByValue,
				},
			},
		}, metav1.CreateOptions{})
		if createErr != nil && !apierrors.IsAlreadyExists(createErr) {
			return fmt.Errorf("create serviceaccount %s: %w", name, createErr)
		}
		return nil
	}

	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}
	existing.Labels[runNamespaceManagedByLabel] = runNamespaceManagedByValue
	_, err = l.client.CoreV1().ServiceAccounts(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update serviceaccount %s: %w", name, err)
	}
	return nil
}

func (l *Launcher) ensureRole(ctx context.Context, namespace string, name string, expectedRules []rbacv1.PolicyRule) error {
	existing, err := l.client.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get role %s: %w", name, err)
		}
		_, createErr := l.client.RbacV1().Roles(namespace).Create(ctx, &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					runNamespaceManagedByLabel: runNamespaceManagedByValue,
				},
			},
			Rules: expectedRules,
		}, metav1.CreateOptions{})
		if createErr != nil && !apierrors.IsAlreadyExists(createErr) {
			return fmt.Errorf("create role %s: %w", name, createErr)
		}
		return nil
	}

	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}
	existing.Labels[runNamespaceManagedByLabel] = runNamespaceManagedByValue
	existing.Rules = expectedRules

	_, err = l.client.RbacV1().Roles(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update role %s: %w", name, err)
	}
	return nil
}

func (l *Launcher) ensureRoleBinding(ctx context.Context, namespace string, name string, roleName string, serviceAccountName string) error {
	expectedSubjects := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      serviceAccountName,
			Namespace: namespace,
		},
	}
	expectedRoleRef := rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     "Role",
		Name:     roleName,
	}

	existing, err := l.client.RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get rolebinding %s: %w", name, err)
		}
		_, createErr := l.client.RbacV1().RoleBindings(namespace).Create(ctx, &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					runNamespaceManagedByLabel: runNamespaceManagedByValue,
				},
			},
			RoleRef:  expectedRoleRef,
			Subjects: expectedSubjects,
		}, metav1.CreateOptions{})
		if createErr != nil && !apierrors.IsAlreadyExists(createErr) {
			return fmt.Errorf("create rolebinding %s: %w", name, createErr)
		}
		return nil
	}

	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}
	existing.Labels[runNamespaceManagedByLabel] = runNamespaceManagedByValue
	existing.RoleRef = expectedRoleRef
	existing.Subjects = expectedSubjects

	_, err = l.client.RbacV1().RoleBindings(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update rolebinding %s: %w", name, err)
	}
	return nil
}

func candidateRoleRules() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{
				"configmaps",
				"endpoints",
				"events",
				"limitranges",
				"persistentvolumeclaims",
				"pods",
				"pods/attach",
				"pods/exec",
				"pods/log",
				"pods/portforward",
				"replicationcontrollers",
				"resourcequotas",
				"serviceaccounts",
				"services",
				"services/proxy",
			},
			Verbs: []string{"*"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		},
		{
			APIGroups: []string{"batch"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		},
		{
			APIGroups: []string{"autoscaling"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		},
		{
			APIGroups: []string{"networking.k8s.io"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		},
		{
			APIGroups: []string{"policy"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		},
	}
}

func productionReadOnlyRoleRules() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{
				"configmaps",
				"endpoints",
				"events",
				"limitranges",
				"persistentvolumeclaims",
				"pods",
				"replicationcontrollers",
				"resourcequotas",
				"serviceaccounts",
				"services",
			},
			Verbs: []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods/log"},
			Verbs:     []string{"get"},
		},
		{
			APIGroups: []string{"events.k8s.io"},
			Resources: []string{"events"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"daemonsets", "deployments", "replicasets", "statefulsets"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"batch"},
			Resources: []string{"cronjobs", "jobs"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"autoscaling"},
			Resources: []string{"horizontalpodautoscalers"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"networking.k8s.io"},
			Resources: []string{"ingresses", "networkpolicies"},
			Verbs:     []string{"get", "list", "watch"},
		},
	}
}

// ensureResourceQuota limits aggregate namespace resource consumption per run namespace.
func (l *Launcher) ensureResourceQuota(ctx context.Context, namespace string) error {
	hard := corev1.ResourceList{
		corev1.ResourcePods: *resource.NewQuantity(l.cfg.RunResourceQuotaPods, resource.DecimalSI),
	}

	name := l.cfg.RunResourceQuotaName

	existing, err := l.client.CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get resourcequota %s: %w", name, err)
		}
		_, createErr := l.client.CoreV1().ResourceQuotas(namespace).Create(ctx, &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					runNamespaceManagedByLabel: runNamespaceManagedByValue,
				},
			},
			Spec: corev1.ResourceQuotaSpec{Hard: hard},
		}, metav1.CreateOptions{})
		if createErr != nil && !apierrors.IsAlreadyExists(createErr) {
			return fmt.Errorf("create resourcequota %s: %w", name, createErr)
		}
		return nil
	}

	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}
	existing.Labels[runNamespaceManagedByLabel] = runNamespaceManagedByValue
	existing.Spec.Hard = hard
	_, err = l.client.CoreV1().ResourceQuotas(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update resourcequota %s: %w", name, err)
	}
	return nil
}

// ensureLimitRange removes managed per-container defaults to avoid cpu/memory constraints.
func (l *Launcher) ensureLimitRange(ctx context.Context, namespace string) error {
	return l.deleteLimitRangeIfExists(ctx, namespace)
}

func (l *Launcher) deleteLimitRangeIfExists(ctx context.Context, namespace string) error {
	name := l.cfg.RunLimitRangeName
	if strings.TrimSpace(name) == "" {
		return nil
	}
	if err := l.client.CoreV1().LimitRanges(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete limitrange %s: %w", name, err)
	}
	return nil
}
