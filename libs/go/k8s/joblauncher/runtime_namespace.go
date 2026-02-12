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

	existing, err := l.client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get namespace: %w", err)
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
		return fmt.Errorf("update namespace: %w", err)
	}
	return nil
}

// ensureRunCredentialsSecret stores runtime credentials consumed by run pod env.
func (l *Launcher) ensureRunCredentialsSecret(ctx context.Context, namespace string, spec JobSpec) error {
	name := l.cfg.RunCredentialsSecretName
	if strings.TrimSpace(name) == "" {
		return nil
	}

	secretData := map[string][]byte{
		"CODEXK8S_OPENAI_API_KEY": []byte(strings.TrimSpace(spec.OpenAIAPIKey)),
		"CODEXK8S_GIT_BOT_TOKEN":  []byte(strings.TrimSpace(spec.GitBotToken)),
	}

	existing, err := l.client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get secret %s: %w", name, err)
		}
		_, createErr := l.client.CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					runNamespaceManagedByLabel: runNamespaceManagedByValue,
				},
			},
			Type: corev1.SecretTypeOpaque,
			Data: secretData,
		}, metav1.CreateOptions{})
		if createErr != nil && !apierrors.IsAlreadyExists(createErr) {
			return fmt.Errorf("create secret %s: %w", name, createErr)
		}
		return nil
	}

	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}
	existing.Labels[runNamespaceManagedByLabel] = runNamespaceManagedByValue
	existing.Type = corev1.SecretTypeOpaque
	existing.Data = secretData

	if _, err := l.client.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("update secret %s: %w", name, err)
	}
	return nil
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

// ensureRunRole ensures least-privilege Role for runtime namespace debug/deploy scenarios.
func (l *Launcher) ensureRunRole(ctx context.Context, namespace string) error {
	name := l.cfg.RunRoleName
	expectedRules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"pods", "pods/log", "events", "services", "endpoints", "configmaps"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods/exec"},
			Verbs:     []string{"create"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"deployments", "replicasets", "statefulsets", "daemonsets"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"batch"},
			Resources: []string{"jobs"},
			Verbs:     []string{"get", "list", "watch"},
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
	requestsCPU, err := resource.ParseQuantity(l.cfg.RunResourceRequestsCPU)
	if err != nil {
		return fmt.Errorf("parse requests.cpu quantity %q: %w", l.cfg.RunResourceRequestsCPU, err)
	}
	requestsMemory, err := resource.ParseQuantity(l.cfg.RunResourceRequestsMemory)
	if err != nil {
		return fmt.Errorf("parse requests.memory quantity %q: %w", l.cfg.RunResourceRequestsMemory, err)
	}
	limitsCPU, err := resource.ParseQuantity(l.cfg.RunResourceLimitsCPU)
	if err != nil {
		return fmt.Errorf("parse limits.cpu quantity %q: %w", l.cfg.RunResourceLimitsCPU, err)
	}
	limitsMemory, err := resource.ParseQuantity(l.cfg.RunResourceLimitsMemory)
	if err != nil {
		return fmt.Errorf("parse limits.memory quantity %q: %w", l.cfg.RunResourceLimitsMemory, err)
	}

	hard := corev1.ResourceList{
		corev1.ResourcePods:           *resource.NewQuantity(l.cfg.RunResourceQuotaPods, resource.DecimalSI),
		corev1.ResourceRequestsCPU:    requestsCPU,
		corev1.ResourceRequestsMemory: requestsMemory,
		corev1.ResourceLimitsCPU:      limitsCPU,
		corev1.ResourceLimitsMemory:   limitsMemory,
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

// ensureLimitRange defines default per-container requests/limits within run namespaces.
func (l *Launcher) ensureLimitRange(ctx context.Context, namespace string) error {
	defaultReqCPU, err := resource.ParseQuantity(l.cfg.RunDefaultRequestCPU)
	if err != nil {
		return fmt.Errorf("parse default request cpu quantity %q: %w", l.cfg.RunDefaultRequestCPU, err)
	}
	defaultReqMemory, err := resource.ParseQuantity(l.cfg.RunDefaultRequestMemory)
	if err != nil {
		return fmt.Errorf("parse default request memory quantity %q: %w", l.cfg.RunDefaultRequestMemory, err)
	}
	defaultLimitCPU, err := resource.ParseQuantity(l.cfg.RunDefaultLimitCPU)
	if err != nil {
		return fmt.Errorf("parse default limit cpu quantity %q: %w", l.cfg.RunDefaultLimitCPU, err)
	}
	defaultLimitMemory, err := resource.ParseQuantity(l.cfg.RunDefaultLimitMemory)
	if err != nil {
		return fmt.Errorf("parse default limit memory quantity %q: %w", l.cfg.RunDefaultLimitMemory, err)
	}

	limit := corev1.LimitRangeItem{
		Type: corev1.LimitTypeContainer,
		DefaultRequest: corev1.ResourceList{
			corev1.ResourceCPU:    defaultReqCPU,
			corev1.ResourceMemory: defaultReqMemory,
		},
		Default: corev1.ResourceList{
			corev1.ResourceCPU:    defaultLimitCPU,
			corev1.ResourceMemory: defaultLimitMemory,
		},
	}
	name := l.cfg.RunLimitRangeName

	existing, err := l.client.CoreV1().LimitRanges(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get limitrange %s: %w", name, err)
		}
		_, createErr := l.client.CoreV1().LimitRanges(namespace).Create(ctx, &corev1.LimitRange{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					runNamespaceManagedByLabel: runNamespaceManagedByValue,
				},
			},
			Spec: corev1.LimitRangeSpec{
				Limits: []corev1.LimitRangeItem{limit},
			},
		}, metav1.CreateOptions{})
		if createErr != nil && !apierrors.IsAlreadyExists(createErr) {
			return fmt.Errorf("create limitrange %s: %w", name, createErr)
		}
		return nil
	}

	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}
	existing.Labels[runNamespaceManagedByLabel] = runNamespaceManagedByValue
	existing.Spec.Limits = []corev1.LimitRangeItem{limit}
	_, err = l.client.CoreV1().LimitRanges(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("update limitrange %s: %w", name, err)
	}
	return nil
}
