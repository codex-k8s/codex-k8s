package workerpresence

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListActiveWorkerIDs(t *testing.T) {
	t.Parallel()

	now := metav1.NewTime(time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC))
	client := fake.NewClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "worker-b",
				Namespace: "codex-k8s-prod",
				Labels: map[string]string{
					"app.kubernetes.io/name":      "codex-k8s",
					"app.kubernetes.io/component": "worker",
				},
			},
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "worker-a",
				Namespace: "codex-k8s-prod",
				Labels: map[string]string{
					"app.kubernetes.io/name":      "codex-k8s",
					"app.kubernetes.io/component": "worker",
				},
			},
			Status: corev1.PodStatus{Phase: corev1.PodPending},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "worker-terminating",
				Namespace: "codex-k8s-prod",
				Labels: map[string]string{
					"app.kubernetes.io/name":      "codex-k8s",
					"app.kubernetes.io/component": "worker",
				},
				DeletionTimestamp: &now,
			},
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "worker-succeeded",
				Namespace: "codex-k8s-prod",
				Labels: map[string]string{
					"app.kubernetes.io/name":      "codex-k8s",
					"app.kubernetes.io/component": "worker",
				},
			},
			Status: corev1.PodStatus{Phase: corev1.PodSucceeded},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "api-gateway",
				Namespace: "codex-k8s-prod",
				Labels: map[string]string{
					"app.kubernetes.io/name":      "codex-k8s",
					"app.kubernetes.io/component": "api-gateway",
				},
			},
			Status: corev1.PodStatus{Phase: corev1.PodRunning},
		},
	)

	adapter := NewAdapterForClient("codex-k8s-prod", client)

	got, err := adapter.ListActiveWorkerIDs(context.Background())
	if err != nil {
		t.Fatalf("ListActiveWorkerIDs() error = %v", err)
	}

	want := []string{"worker-a", "worker-b"}
	if len(got) != len(want) {
		t.Fatalf("ListActiveWorkerIDs() len=%d want %d; got=%v", len(got), len(want), got)
	}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("ListActiveWorkerIDs()[%d]=%q want %q", idx, got[idx], want[idx])
		}
	}
}
