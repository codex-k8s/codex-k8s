package joblauncher

import (
	"context"
	"testing"
	"time"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLauncher_EnsureNamespace_PreparesBaselineResourcesAndLeaseMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := fake.NewClientset()
	launcher := NewForClient(Config{Namespace: "codex-k8s-prod"}, client)

	expiresAt := time.Date(2026, 2, 21, 12, 0, 0, 0, time.UTC)
	spec := NamespaceSpec{
		RunID:          "run-1",
		ProjectID:      "project-1",
		IssueNumber:    74,
		AgentKey:       "dev",
		CorrelationID:  "corr-1",
		RuntimeMode:    agentdomain.RuntimeModeFullEnv,
		Namespace:      "codex-issue-p1-i74-r1",
		LeaseTTL:       24 * time.Hour,
		LeaseExpiresAt: expiresAt,
	}

	result, err := launcher.EnsureNamespace(ctx, spec)
	if err != nil {
		t.Fatalf("EnsureNamespace() error = %v", err)
	}
	if !result.Created {
		t.Fatal("expected namespace to be created on first ensure")
	}
	if result.Reused {
		t.Fatal("expected created namespace not to be marked as reused")
	}
	if !result.LeaseExpiresAt.Equal(expiresAt) {
		t.Fatalf("unexpected lease expires at: got %s want %s", result.LeaseExpiresAt.Format(time.RFC3339), expiresAt.Format(time.RFC3339))
	}

	ns, err := client.CoreV1().Namespaces().Get(ctx, spec.Namespace, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("namespace not created: %v", err)
	}
	if got, want := ns.Labels[runNamespaceIssueNumberLabel], "74"; got != want {
		t.Fatalf("unexpected issue-number label: got %q want %q", got, want)
	}
	if got, want := ns.Labels[runNamespaceAgentKeyLabel], "dev"; got != want {
		t.Fatalf("unexpected agent-key label: got %q want %q", got, want)
	}
	if got, want := ns.Annotations[runNamespaceLeaseExpAnnotKey], expiresAt.Format(time.RFC3339); got != want {
		t.Fatalf("unexpected lease expires annotation: got %q want %q", got, want)
	}
	if got, want := ns.Annotations[runNamespaceLeaseTTLAnnotKey], (24 * time.Hour).String(); got != want {
		t.Fatalf("unexpected lease ttl annotation: got %q want %q", got, want)
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

func TestLauncher_FindReusableNamespace_ReturnsLatestActiveLease(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Date(2026, 2, 21, 10, 0, 0, 0, time.UTC)
	project := "project-1"
	projectLabel := sanitizeLabel(project)
	client := fake.NewClientset(
		newLeaseNamespace("ns-old", leaseNamespaceParams{
			projectLabel: projectLabel,
			issueNumber:  "74",
			agentKey:     "dev",
			expiresAt:    now.Add(1 * time.Hour),
		}),
		newLeaseNamespace("ns-new", leaseNamespaceParams{
			projectLabel: projectLabel,
			issueNumber:  "74",
			agentKey:     "dev",
			expiresAt:    now.Add(3 * time.Hour),
		}),
	)
	launcher := NewForClient(Config{Namespace: "codex-k8s-prod"}, client)

	result, ok, err := launcher.FindReusableNamespace(ctx, NamespaceReuseLookup{
		ProjectID:   project,
		IssueNumber: 74,
		AgentKey:    "dev",
		Now:         now,
	})
	if err != nil {
		t.Fatalf("FindReusableNamespace() error = %v", err)
	}
	if !ok {
		t.Fatal("expected reusable namespace to be found")
	}
	if got, want := result.Namespace, "ns-new"; got != want {
		t.Fatalf("unexpected reusable namespace: got %q want %q", got, want)
	}
}

func TestLauncher_CleanupExpiredNamespaces_DeletesOnlyExpired(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Date(2026, 2, 21, 10, 0, 0, 0, time.UTC)
	client := fake.NewClientset(
		newLeaseNamespace("ns-expired", leaseNamespaceParams{
			runID:     "run-expired",
			expiresAt: now.Add(-1 * time.Minute),
		}),
		newLeaseNamespace("ns-active", leaseNamespaceParams{
			runID:     "run-active",
			expiresAt: now.Add(1 * time.Hour),
		}),
	)
	launcher := NewForClient(Config{Namespace: "codex-k8s-prod"}, client)

	cleaned, err := launcher.CleanupExpiredNamespaces(ctx, NamespaceCleanupParams{Now: now, Limit: 10})
	if err != nil {
		t.Fatalf("CleanupExpiredNamespaces() error = %v", err)
	}
	if len(cleaned) != 1 {
		t.Fatalf("expected one expired namespace cleaned, got %d", len(cleaned))
	}
	if got, want := cleaned[0].Namespace, "ns-expired"; got != want {
		t.Fatalf("unexpected cleaned namespace: got %q want %q", got, want)
	}
	if got, want := cleaned[0].RunID, "run-expired"; got != want {
		t.Fatalf("unexpected cleaned run id: got %q want %q", got, want)
	}

	if _, err := client.CoreV1().Namespaces().Get(ctx, "ns-expired", metav1.GetOptions{}); err == nil {
		t.Fatal("expected expired namespace to be deleted")
	}
	if _, err := client.CoreV1().Namespaces().Get(ctx, "ns-active", metav1.GetOptions{}); err != nil {
		t.Fatalf("expected active namespace to remain: %v", err)
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
	if _, err := launcher.EnsureNamespace(ctx, spec); err != nil {
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

type leaseNamespaceParams struct {
	projectLabel string
	issueNumber  string
	agentKey     string
	runID        string
	expiresAt    time.Time
}

func newLeaseNamespace(name string, params leaseNamespaceParams) *corev1.Namespace {
	labels := map[string]string{
		runNamespaceManagedByLabel:   runNamespaceManagedByValue,
		runNamespacePurposeLabel:     runNamespacePurposeValue,
		runNamespaceRuntimeModeLabel: string(agentdomain.RuntimeModeFullEnv),
	}
	if params.projectLabel != "" {
		labels[runNamespaceProjectIDLabel] = params.projectLabel
	}
	if params.issueNumber != "" {
		labels[runNamespaceIssueNumberLabel] = params.issueNumber
	}
	if params.agentKey != "" {
		labels[runNamespaceAgentKeyLabel] = params.agentKey
	}
	if params.runID != "" {
		labels[runNamespaceRunIDLabel] = params.runID
	}

	annotations := map[string]string{}
	if !params.expiresAt.IsZero() {
		annotations[runNamespaceLeaseExpAnnotKey] = params.expiresAt.Format(time.RFC3339)
	}

	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
	}
}
