package joblauncher

import (
	"context"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLauncher_Status_ImagePullBackOffIsFailed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := fake.NewClientset(
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: "codex-k8s-run-abc", Namespace: "ns"},
			Status:     batchv1.JobStatus{},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-1",
				Namespace: "ns",
				Labels: map[string]string{
					"job-name": "codex-k8s-run-abc",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name: "run",
						State: corev1.ContainerState{
							Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff"},
						},
					},
				},
			},
		},
	)

	l := NewForClient(Config{Namespace: "ns"}, client)
	state, err := l.Status(ctx, JobRef{Namespace: "ns", Name: "codex-k8s-run-abc"})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if state != JobStateFailed {
		t.Fatalf("expected %q, got %q", JobStateFailed, state)
	}
}

func TestLauncher_Status_CompleteConditionIsSucceeded(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := fake.NewClientset(
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: "job1", Namespace: "ns"},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
				},
			},
		},
	)

	l := NewForClient(Config{Namespace: "ns"}, client)
	state, err := l.Status(ctx, JobRef{Namespace: "ns", Name: "job1"})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if state != JobStateSucceeded {
		t.Fatalf("expected %q, got %q", JobStateSucceeded, state)
	}
}

func TestLauncher_Status_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := fake.NewClientset()

	l := NewForClient(Config{Namespace: "ns"}, client)
	state, err := l.Status(ctx, JobRef{Namespace: "ns", Name: "missing"})
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if state != JobStateNotFound {
		t.Fatalf("expected %q, got %q", JobStateNotFound, state)
	}
}

func TestLauncher_Launch_AIRepairCreatesPod(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := fake.NewClientset()
	l := NewForClient(Config{Namespace: "ns", Image: "busybox:1.36"}, client)

	spec := JobSpec{
		RunID:              "run-ai-repair",
		CorrelationID:      "corr-ai-repair",
		ProjectID:          "project-1",
		Namespace:          "codex-k8s-prod",
		TriggerKind:        "ai_repair",
		ServiceAccountName: "codex-k8s-control-plane",
	}
	ref, err := l.Launch(ctx, spec)
	if err != nil {
		t.Fatalf("Launch returned error: %v", err)
	}

	if ref.Name == "" || ref.Namespace != "codex-k8s-prod" {
		t.Fatalf("unexpected launch ref: %+v", ref)
	}
	if _, err := client.BatchV1().Jobs(ref.Namespace).Get(ctx, ref.Name, metav1.GetOptions{}); err == nil {
		t.Fatalf("expected ai-repair launch without Job, but job %s/%s exists", ref.Namespace, ref.Name)
	}
	pod, err := client.CoreV1().Pods(ref.Namespace).Get(ctx, ref.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("expected ai-repair pod %s/%s, got error: %v", ref.Namespace, ref.Name, err)
	}
	if got, want := pod.Spec.ServiceAccountName, "codex-k8s-control-plane"; got != want {
		t.Fatalf("expected service account %q, got %q", want, got)
	}
	if got, want := len(pod.Spec.Containers), 2; got != want {
		t.Fatalf("expected %d containers, got %d", want, got)
	}
}

func TestLauncher_Status_AIRepairPodRunContainerSucceeded(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ref := JobRef{Namespace: "ns", Name: "codex-k8s-run-ai"}
	client := fake.NewClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ref.Name,
				Namespace: ref.Namespace,
				Labels: map[string]string{
					metadataLabelRunID: "run-ai-repair",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name: runContainerName,
						State: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{ExitCode: 0},
						},
					},
					{
						Name: aiRepairKeepaliveName,
						State: corev1.ContainerState{
							Running: &corev1.ContainerStateRunning{},
						},
					},
				},
			},
		},
	)

	l := NewForClient(Config{Namespace: "ns"}, client)
	state, err := l.Status(ctx, ref)
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if state != JobStateSucceeded {
		t.Fatalf("expected %q, got %q", JobStateSucceeded, state)
	}
}

func TestLauncher_FindRunJobRefByRunID_FallsBackToPod(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ref := JobRef{Namespace: "ns", Name: BuildRunJobName("run-ai-repair")}
	client := fake.NewClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ref.Name,
				Namespace: ref.Namespace,
				Labels: map[string]string{
					metadataLabelRunID:       "run-ai-repair",
					"app.kubernetes.io/name": runWorkloadAppName,
				},
			},
		},
	)
	l := NewForClient(Config{Namespace: "ns"}, client)

	got, found, err := l.FindRunJobRefByRunID(ctx, "run-ai-repair")
	if err != nil {
		t.Fatalf("FindRunJobRefByRunID returned error: %v", err)
	}
	if !found {
		t.Fatal("expected pod-backed ref to be found")
	}
	if got != ref {
		t.Fatalf("unexpected ref: got %+v want %+v", got, ref)
	}
}
