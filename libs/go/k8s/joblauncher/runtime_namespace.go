package joblauncher

import (
	"context"
	"fmt"
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

const (
	runNamespaceManagedByLabel      = metadataLabelManagedBy
	runNamespacePurposeLabel        = metadataLabelNamespacePurpose
	runNamespaceRuntimeModeLabel    = metadataLabelRuntimeMode
	runNamespaceProjectIDLabel      = metadataLabelProjectID
	runNamespaceRunIDLabel          = metadataLabelRunID
	runNamespaceCorrelationAnnotKey = metadataAnnotationCorrelationID

	runNamespaceManagedByValue = "codex-k8s-worker"
	runNamespacePurposeValue   = "run"
)

// EnsureNamespace prepares baseline runtime resources for full-env execution.
func (l *Launcher) EnsureNamespace(ctx context.Context, spec NamespaceSpec) error {
	if spec.RuntimeMode != agentdomain.RuntimeModeFullEnv {
		return nil
	}
	namespace := strings.TrimSpace(spec.Namespace)
	if namespace == "" {
		return fmt.Errorf("runtime namespace is required for full-env run")
	}

	if err := l.ensureNamespaceObject(ctx, spec); err != nil {
		return fmt.Errorf("ensure namespace %s: %w", namespace, err)
	}
	if err := l.ensureRunServiceAccount(ctx, namespace); err != nil {
		return fmt.Errorf("ensure serviceaccount in namespace %s: %w", namespace, err)
	}
	if err := l.ensureRunRole(ctx, namespace); err != nil {
		return fmt.Errorf("ensure role in namespace %s: %w", namespace, err)
	}
	if err := l.ensureRunRoleBinding(ctx, namespace); err != nil {
		return fmt.Errorf("ensure rolebinding in namespace %s: %w", namespace, err)
	}
	if err := l.ensureResourceQuota(ctx, namespace); err != nil {
		return fmt.Errorf("ensure resource quota in namespace %s: %w", namespace, err)
	}
	if err := l.ensureLimitRange(ctx, namespace); err != nil {
		return fmt.Errorf("ensure limit range in namespace %s: %w", namespace, err)
	}
	return nil
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

// ensureNamespaceObject upserts namespace metadata required for managed runtime namespaces.
func (l *Launcher) ensureNamespaceObject(ctx context.Context, spec NamespaceSpec) error {
	namespace := strings.TrimSpace(spec.Namespace)
	labels := map[string]string{
		runNamespaceManagedByLabel:   runNamespaceManagedByValue,
		runNamespacePurposeLabel:     runNamespacePurposeValue,
		runNamespaceRuntimeModeLabel: string(spec.RuntimeMode),
		runNamespaceRunIDLabel:       sanitizeLabel(spec.RunID),
		runNamespaceProjectIDLabel:   sanitizeLabel(spec.ProjectID),
	}
	projectLabel := sanitizeLabel(spec.ProjectID)
	if projectLabel != "unknown" {
		labels[runNamespaceProjectIDLabel] = projectLabel
	}
	annotations := map[string]string{
		runNamespaceCorrelationAnnotKey: spec.CorrelationID,
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		existing, getErr := l.client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
		if getErr != nil {
			if !apierrors.IsNotFound(getErr) {
				return fmt.Errorf("get namespace: %w", getErr)
			}
			_, createErr := l.client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:        namespace,
					Labels:      labels,
					Annotations: annotations,
				},
			}, metav1.CreateOptions{})
			if createErr != nil && !apierrors.IsAlreadyExists(createErr) {
				return fmt.Errorf("create namespace: %w", createErr)
			}
			return nil
		}

		existing = existing.DeepCopy()
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

		_, updateErr := l.client.CoreV1().Namespaces().Update(ctx, existing, metav1.UpdateOptions{})
		return updateErr
	})
	if err != nil {
		return fmt.Errorf("upsert namespace: %w", err)
	}
	return nil
}

// ensureRunServiceAccount ensures ServiceAccount exists for in-namespace run access.
func (l *Launcher) ensureRunServiceAccount(ctx context.Context, namespace string) error {
	name := l.cfg.RunServiceAccountName
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		existing, getErr := l.client.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			if !apierrors.IsNotFound(getErr) {
				return fmt.Errorf("get serviceaccount %s: %w", name, getErr)
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

		existing = existing.DeepCopy()
		if existing.Labels == nil {
			existing.Labels = map[string]string{}
		}
		existing.Labels[runNamespaceManagedByLabel] = runNamespaceManagedByValue
		_, updateErr := l.client.CoreV1().ServiceAccounts(namespace).Update(ctx, existing, metav1.UpdateOptions{})
		return updateErr
	})
	if err != nil {
		return fmt.Errorf("upsert serviceaccount %s: %w", name, err)
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

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		existing, getErr := l.client.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			if !apierrors.IsNotFound(getErr) {
				return fmt.Errorf("get role %s: %w", name, getErr)
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

		existing = existing.DeepCopy()
		if existing.Labels == nil {
			existing.Labels = map[string]string{}
		}
		existing.Labels[runNamespaceManagedByLabel] = runNamespaceManagedByValue
		existing.Rules = expectedRules

		_, updateErr := l.client.RbacV1().Roles(namespace).Update(ctx, existing, metav1.UpdateOptions{})
		return updateErr
	})
	if err != nil {
		return fmt.Errorf("upsert role %s: %w", name, err)
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

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		existing, getErr := l.client.RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			if !apierrors.IsNotFound(getErr) {
				return fmt.Errorf("get rolebinding %s: %w", name, getErr)
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

		existing = existing.DeepCopy()
		if existing.Labels == nil {
			existing.Labels = map[string]string{}
		}
		existing.Labels[runNamespaceManagedByLabel] = runNamespaceManagedByValue
		existing.RoleRef = expectedRoleRef
		existing.Subjects = expectedSubjects

		_, updateErr := l.client.RbacV1().RoleBindings(namespace).Update(ctx, existing, metav1.UpdateOptions{})
		return updateErr
	})
	if err != nil {
		return fmt.Errorf("upsert rolebinding %s: %w", name, err)
	}
	return nil
}

// ensureResourceQuota limits aggregate namespace resource consumption per run namespace.
func (l *Launcher) ensureResourceQuota(ctx context.Context, namespace string) error {
	hard := corev1.ResourceList{
		corev1.ResourcePods: *resource.NewQuantity(l.cfg.RunResourceQuotaPods, resource.DecimalSI),
	}

	name := l.cfg.RunResourceQuotaName

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		existing, getErr := l.client.CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			if !apierrors.IsNotFound(getErr) {
				return fmt.Errorf("get resourcequota %s: %w", name, getErr)
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

		existing = existing.DeepCopy()
		if existing.Labels == nil {
			existing.Labels = map[string]string{}
		}
		existing.Labels[runNamespaceManagedByLabel] = runNamespaceManagedByValue
		existing.Spec.Hard = hard
		_, updateErr := l.client.CoreV1().ResourceQuotas(namespace).Update(ctx, existing, metav1.UpdateOptions{})
		return updateErr
	})
	if err != nil {
		return fmt.Errorf("upsert resourcequota %s: %w", name, err)
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
