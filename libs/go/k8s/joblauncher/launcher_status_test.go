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
