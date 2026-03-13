package kubernetes

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestEnsureNamespace_FailsWhenNamespaceIsTerminating(t *testing.T) {
	t.Parallel()

	deletionTime := metav1.NewTime(time.Date(2026, 3, 12, 19, 39, 0, 0, time.UTC))
	client := NewForClient(&rest.Config{Host: "https://example.invalid"}, fake.NewClientset(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "codex-k8s-dev-2",
			DeletionTimestamp: &deletionTime,
		},
	}))

	err := client.EnsureNamespace(context.Background(), "codex-k8s-dev-2")
	if err == nil {
		t.Fatal("expected terminating namespace error, got nil")
	}
	if got, want := err.Error(), "namespace codex-k8s-dev-2 is terminating"; got != want {
		t.Fatalf("EnsureNamespace() error = %q, want %q", got, want)
	}
}
