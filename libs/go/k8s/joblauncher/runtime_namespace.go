package joblauncher

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
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

	runNamespaceManagedByValue = "codex-k8s-worker"
	runNamespacePurposeValue   = "run"
)

// EnsureNamespace prepares baseline runtime resources for full-env execution.
func (l *Launcher) EnsureNamespace(ctx context.Context, spec NamespaceSpec) (NamespaceEnsureResult, error) {
	if spec.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		return NamespaceEnsureResult{}, nil
	}
	namespace := strings.TrimSpace(spec.Namespace)
	if namespace == "" {
		return NamespaceEnsureResult{}, fmt.Errorf("runtime namespace is required for full-env run")
	}

	ensureResult, err := l.ensureNamespaceObject(ctx, spec)
	if err != nil {
		return NamespaceEnsureResult{}, fmt.Errorf("ensure namespace %s: %w", namespace, err)
	}
	if err := l.ensureRunServiceAccount(ctx, namespace); err != nil {
		return NamespaceEnsureResult{}, fmt.Errorf("ensure serviceaccount in namespace %s: %w", namespace, err)
	}
	if err := l.ensureRunRole(ctx, namespace); err != nil {
		return NamespaceEnsureResult{}, fmt.Errorf("ensure role in namespace %s: %w", namespace, err)
	}
	if err := l.ensureRunRoleBinding(ctx, namespace); err != nil {
		return NamespaceEnsureResult{}, fmt.Errorf("ensure rolebinding in namespace %s: %w", namespace, err)
	}
	if err := l.ensureResourceQuota(ctx, namespace); err != nil {
		return NamespaceEnsureResult{}, fmt.Errorf("ensure resource quota in namespace %s: %w", namespace, err)
	}
	if err := l.ensureLimitRange(ctx, namespace); err != nil {
		return NamespaceEnsureResult{}, fmt.Errorf("ensure limit range in namespace %s: %w", namespace, err)
	}
	return ensureResult, nil
}

// CleanupNamespace removes runtime namespace after run completion.
func (l *Launcher) CleanupNamespace(ctx context.Context, spec NamespaceSpec) error {
	if spec.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		return nil
	}

	namespace := strings.TrimSpace(spec.Namespace)
	if namespace == "" {
		return nil
	}
	if namespace == l.cfg.Namespace {
		return nil
	}

	ns, err := l.client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("get namespace %s: %w", namespace, err)
	}
	if strings.TrimSpace(ns.Labels[runNamespaceManagedByLabel]) != runNamespaceManagedByValue {
		return nil
	}
	if strings.TrimSpace(ns.Labels[runNamespacePurposeLabel]) != runNamespacePurposeValue {
		return nil
	}
	if strings.TrimSpace(ns.Labels[runNamespaceRuntimeModeLabel]) != string(agentdomain.RuntimeModeFullEnv) {
		return nil
	}

	if err := l.client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("delete namespace %s: %w", namespace, err)
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

// CleanupExpiredNamespaces removes managed full-env namespaces when lease is expired.
func (l *Launcher) CleanupExpiredNamespaces(ctx context.Context, params NamespaceCleanupParams) ([]NamespaceCleanupResult, error) {
	now := params.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	limit := params.Limit
	if limit <= 0 {
		limit = 200
	}

	items, err := l.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf(
			"%s=%s,%s=%s,%s=%s",
			runNamespaceManagedByLabel,
			runNamespaceManagedByValue,
			runNamespacePurposeLabel,
			runNamespacePurposeValue,
			runNamespaceRuntimeModeLabel,
			string(agentdomain.RuntimeModeFullEnv),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("list managed run namespaces: %w", err)
	}
	sort.Slice(items.Items, func(i, j int) bool {
		return items.Items[i].Name < items.Items[j].Name
	})

	cleaned := make([]NamespaceCleanupResult, 0, limit)
	for _, item := range items.Items {
		if len(cleaned) >= limit {
			break
		}
		if strings.TrimSpace(item.Name) == strings.TrimSpace(l.cfg.Namespace) {
			continue
		}
		if item.DeletionTimestamp != nil {
			continue
		}
		expiresAt, ok := parseNamespaceLeaseExpiresAt(item.Annotations)
		if !ok || expiresAt.After(now) {
			continue
		}

		deleteErr := l.client.CoreV1().Namespaces().Delete(ctx, item.Name, metav1.DeleteOptions{})
		if deleteErr != nil && !apierrors.IsNotFound(deleteErr) {
			return nil, fmt.Errorf("delete expired run namespace %s: %w", item.Name, deleteErr)
		}

		cleaned = append(cleaned, NamespaceCleanupResult{
			Namespace: strings.TrimSpace(item.Name),
			RunID:     strings.TrimSpace(item.Labels[runNamespaceRunIDLabel]),
			ExpiresAt: expiresAt,
		})
	}

	return cleaned, nil
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

// ensureRunServiceAccount ensures ServiceAccount exists for in-namespace run access.
func (l *Launcher) ensureRunServiceAccount(ctx context.Context, namespace string) error {
	name := l.cfg.RunServiceAccountName
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

// ensureRunRole ensures broad full-env namespace permissions except direct secrets access.
func (l *Launcher) ensureRunRole(ctx context.Context, namespace string) error {
	name := l.cfg.RunRoleName
	expectedRules := []rbacv1.PolicyRule{
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

// ensureRunRoleBinding binds runtime ServiceAccount to the managed Role.
func (l *Launcher) ensureRunRoleBinding(ctx context.Context, namespace string) error {
	name := l.cfg.RunRoleBindingName
	expectedSubjects := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      l.cfg.RunServiceAccountName,
			Namespace: namespace,
		},
	}
	expectedRoleRef := rbacv1.RoleRef{
		APIGroup: rbacv1.GroupName,
		Kind:     "Role",
		Name:     l.cfg.RunRoleName,
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
