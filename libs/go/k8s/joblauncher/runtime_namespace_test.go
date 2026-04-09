package joblauncher

import (
	"context"
	"reflect"
	"testing"
	"time"

	agentdomain "github.com/codex-k8s/kodex/libs/go/domain/agent"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLauncher_EnsureNamespace_PreparesBaselineResourcesAndLeaseMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := fake.NewClientset()
	launcher := NewForClient(Config{Namespace: "kodex-prod"}, client)

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
	launcher := NewForClient(Config{Namespace: "kodex-prod"}, client)

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

func TestLauncher_ListManagedRunNamespaces_KeepsKnownSlotNamespacesInCleanupScope(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := fake.NewClientset(
		newLeaseNamespace("codex-issue-expired", leaseNamespaceParams{
			runID:     "run-expired",
			expiresAt: time.Date(2026, 2, 21, 9, 0, 0, 0, time.UTC),
		}),
		newLeaseNamespace("kodex-dev-1", leaseNamespaceParams{
			runID:     "run-slot",
			expiresAt: time.Date(2026, 2, 21, 8, 0, 0, 0, time.UTC),
		}),
		newLeaseNamespace("other-prefix-expired", leaseNamespaceParams{
			runID:     "run-active",
			expiresAt: time.Date(2026, 2, 21, 9, 0, 0, 0, time.UTC),
		}),
	)
	launcher := NewForClient(Config{Namespace: "kodex-prod"}, client)

	items, err := launcher.ListManagedRunNamespaces(ctx, ManagedNamespaceListParams{NamespacePrefix: "codex-issue"})
	if err != nil {
		t.Fatalf("ListManagedRunNamespaces() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected two managed namespaces in cleanup scope, got %d", len(items))
	}
	if got, want := []string{items[0].Namespace, items[1].Namespace}, []string{"codex-issue-expired", "kodex-dev-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected managed namespaces: got %v want %v", got, want)
	}
	if got, want := items[0].RunID, "run-expired"; got != want {
		t.Fatalf("unexpected managed run id: got %q want %q", got, want)
	}
	if got, want := items[0].LeaseTTL, 24*time.Hour; got != want {
		t.Fatalf("unexpected managed lease ttl: got %s want %s", got, want)
	}
	if got, want := items[0].LeaseExpiresAt.Format(time.RFC3339), "2026-02-21T09:00:00Z"; got != want {
		t.Fatalf("unexpected lease expiry: got %q want %q", got, want)
	}
}

func TestLauncher_InspectNamespaceWorkloads_DetectsActiveResources(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	namespace := "codex-issue-run"
	var zeroReplicas int32
	client := fake.NewClientset(
		newLeaseNamespace(namespace, leaseNamespaceParams{}),
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "running-pod", Namespace: namespace},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "done-pod", Namespace: namespace},
			Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
		},
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: "active-job", Namespace: namespace},
			Status:     batchv1.JobStatus{Active: 1},
		},
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: "finished-job", Namespace: namespace},
			Status: batchv1.JobStatus{Conditions: []batchv1.JobCondition{{
				Type:   batchv1.JobComplete,
				Status: corev1.ConditionTrue,
			}}},
		},
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{Name: "scheduled-cron", Namespace: namespace},
		},
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{Name: "suspended-cron", Namespace: namespace},
			Spec:       batchv1.CronJobSpec{Suspend: boolPtr(true)},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "live-deploy", Namespace: namespace},
			Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(1)},
		},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{Name: "scaled-down-rs", Namespace: namespace},
			Spec:       appsv1.ReplicaSetSpec{Replicas: &zeroReplicas},
		},
	)
	launcher := NewForClient(Config{Namespace: "kodex-prod"}, client)

	workloads, err := launcher.InspectNamespaceWorkloads(ctx, namespace)
	if err != nil {
		t.Fatalf("InspectNamespaceWorkloads() error = %v", err)
	}
	if got, want := workloads.ActivePods, []string{"running-pod"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("active pods = %v, want %v", got, want)
	}
	if got, want := workloads.ActiveJobs, []string{"active-job"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("active jobs = %v, want %v", got, want)
	}
	if got, want := workloads.ActiveCronJobs, []string{"scheduled-cron"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("active cronjobs = %v, want %v", got, want)
	}
	if got, want := workloads.ActiveDeployments, []string{"live-deploy"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("active deployments = %v, want %v", got, want)
	}
	if workloads.HasActiveWorkloads() == false {
		t.Fatal("expected active workloads to be detected")
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
	launcher := NewForClient(Config{Namespace: "kodex-prod"}, client)

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
	launcher := NewForClient(Config{Namespace: "kodex-prod"}, client)

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

func TestLauncher_EnsureAccessProfile_ProductionReadOnlyForbidsExecPortForwardAndSecrets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := fake.NewClientset()
	launcher := NewForClient(Config{Namespace: "kodex-prod"}, client)

	serviceAccountName, err := launcher.EnsureAccessProfile(ctx, "kodex-prod", agentdomain.RuntimeAccessProfileProductionReadOnly)
	if err != nil {
		t.Fatalf("EnsureAccessProfile() error = %v", err)
	}
	if got, want := serviceAccountName, launcher.cfg.RunReadOnlyServiceAccountName; got != want {
		t.Fatalf("service account = %q, want %q", got, want)
	}

	role, err := client.RbacV1().Roles("kodex-prod").Get(ctx, launcher.cfg.RunReadOnlyRoleName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("load readonly role failed: %v", err)
	}

	for _, rule := range role.Rules {
		for _, verb := range rule.Verbs {
			if verb != "get" && verb != "list" && verb != "watch" {
				t.Fatalf("unexpected verb %q in readonly role", verb)
			}
		}
		for _, resource := range rule.Resources {
			switch resource {
			case "pods/exec", "pods/portforward", "secrets", "secrets/*":
				t.Fatalf("unexpected resource %q in readonly role", resource)
			}
		}
	}
}

type leaseNamespaceParams struct {
	projectLabel string
	issueNumber  string
	agentKey     string
	runID        string
	leaseTTL     time.Duration
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
		leaseTTL := params.leaseTTL
		if leaseTTL <= 0 {
			leaseTTL = 24 * time.Hour
		}
		annotations[runNamespaceLeaseTTLAnnotKey] = leaseTTL.String()
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

func int32Ptr(value int32) *int32 {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}
