package joblauncher

import (
	"context"
	"testing"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLauncher_EnsureNamespace_PreparesBaselineResources(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := fake.NewClientset()
	launcher := NewForClient(Config{Namespace: "codex-k8s-prod"}, client)

	spec := NamespaceSpec{
		RunID:         "run-1",
		ProjectID:     "project-1",
		CorrelationID: "corr-1",
		RuntimeMode:   agentdomain.RuntimeModeFullEnv,
		Namespace:     "codex-issue-p1-i1-r1",
	}
	if err := launcher.EnsureNamespace(ctx, spec); err != nil {
		t.Fatalf("EnsureNamespace() error = %v", err)
	}

	if _, err := client.CoreV1().Namespaces().Get(ctx, spec.Namespace, metav1.GetOptions{}); err != nil {
		t.Fatalf("namespace not created: %v", err)
	}
	if _, err := client.CoreV1().ServiceAccounts(spec.Namespace).Get(ctx, launcher.cfg.RunServiceAccountName, metav1.GetOptions{}); err != nil {
		t.Fatalf("serviceaccount not created: %v", err)
	}
	if _, err := client.RbacV1().Roles(spec.Namespace).Get(ctx, launcher.cfg.RunRoleName, metav1.GetOptions{}); err != nil {
		t.Fatalf("role not created: %v", err)
	}
	if _, err := client.RbacV1().RoleBindings(spec.Namespace).Get(ctx, launcher.cfg.RunRoleBindingName, metav1.GetOptions{}); err != nil {
		t.Fatalf("rolebinding not created: %v", err)
	}
	if _, err := client.CoreV1().ResourceQuotas(spec.Namespace).Get(ctx, launcher.cfg.RunResourceQuotaName, metav1.GetOptions{}); err != nil {
		t.Fatalf("resourcequota not created: %v", err)
	}
	if _, err := client.CoreV1().LimitRanges(spec.Namespace).Get(ctx, launcher.cfg.RunLimitRangeName, metav1.GetOptions{}); err == nil {
		t.Fatalf("limitrange should not be present in managed runtime namespace")
	}
}

func TestLauncher_CleanupNamespace_DeletesManagedNamespace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	namespace := "codex-issue-p1-i1-r1"
	client := fake.NewClientset(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				runNamespaceManagedByLabel:   runNamespaceManagedByValue,
				runNamespacePurposeLabel:     runNamespacePurposeValue,
				runNamespaceRuntimeModeLabel: string(agentdomain.RuntimeModeFullEnv),
			},
		},
	})
	launcher := NewForClient(Config{Namespace: "codex-k8s-prod"}, client)

	err := launcher.CleanupNamespace(ctx, NamespaceSpec{
		RunID:       "run-1",
		RuntimeMode: agentdomain.RuntimeModeFullEnv,
		Namespace:   namespace,
	})
	if err != nil {
		t.Fatalf("CleanupNamespace() error = %v", err)
	}

	if _, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{}); err == nil {
		t.Fatalf("expected namespace %s to be deleted", namespace)
	}
}

func TestLauncher_EnsureNamespace_RunRoleDoesNotGrantSecretsAccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := fake.NewClientset()
	launcher := NewForClient(Config{Namespace: "codex-k8s-prod"}, client)

	spec := NamespaceSpec{
		RunID:         "run-2",
		ProjectID:     "project-2",
		CorrelationID: "corr-2",
		RuntimeMode:   agentdomain.RuntimeModeFullEnv,
		Namespace:     "codex-issue-p2-i2-r2",
	}
	if err := launcher.EnsureNamespace(ctx, spec); err != nil {
		t.Fatalf("EnsureNamespace() error = %v", err)
	}

	role, err := client.RbacV1().Roles(spec.Namespace).Get(ctx, launcher.cfg.RunRoleName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("load role failed: %v", err)
	}

	for _, rule := range role.Rules {
		isCoreGroup := false
		for _, apiGroup := range rule.APIGroups {
			if apiGroup == "" {
				isCoreGroup = true
				break
			}
		}
		if !isCoreGroup {
			continue
		}
		for _, resource := range rule.Resources {
			if resource == "secrets" || resource == "secrets/*" {
				t.Fatalf("unexpected secrets access in role rules: %+v", role.Rules)
			}
		}
	}
}
