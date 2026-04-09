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
			Name:              "kodex-dev-2",
			DeletionTimestamp: &deletionTime,
		},
	}))

	err := client.EnsureNamespace(context.Background(), "kodex-dev-2")
	if err == nil {
		t.Fatal("expected terminating namespace error, got nil")
	}
	if got, want := err.Error(), "namespace kodex-dev-2 is terminating"; got != want {
		t.Fatalf("EnsureNamespace() error = %q, want %q", got, want)
	}
}

func TestGetManagedRunNamespace_ReturnsOnlyWorkerManagedRunNamespaces(t *testing.T) {
	t.Parallel()

	client := NewForClient(&rest.Config{Host: "https://example.invalid"}, fake.NewClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kodex-dev-1",
				Labels: map[string]string{
					runNamespaceManagedByLabel: runNamespaceManagedByValue,
					runNamespacePurposeLabel:   runNamespacePurposeValue,
				},
				Annotations: map[string]string{
					"kodex.works/runtime-fingerprint-hash": "hash-1",
				},
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		},
	))

	state, found, err := client.GetManagedRunNamespace(context.Background(), "kodex-dev-1")
	if err != nil {
		t.Fatalf("GetManagedRunNamespace() error = %v", err)
	}
	if !found {
		t.Fatal("expected managed namespace to be found")
	}
	if got, want := state.Name, "kodex-dev-1"; got != want {
		t.Fatalf("expected namespace %q, got %q", want, got)
	}
	if got, want := state.Annotations["kodex.works/runtime-fingerprint-hash"], "hash-1"; got != want {
		t.Fatalf("expected annotation %q, got %q", want, got)
	}

	_, found, err = client.GetManagedRunNamespace(context.Background(), "default")
	if err != nil {
		t.Fatalf("GetManagedRunNamespace(default) error = %v", err)
	}
	if found {
		t.Fatal("expected unmanaged namespace to be filtered out")
	}
}

func TestUpsertNamespaceAnnotations_MergesAnnotations(t *testing.T) {
	t.Parallel()

	clientset := fake.NewClientset(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kodex-dev-1",
			Annotations: map[string]string{
				"existing": "value",
			},
		},
	})
	client := NewForClient(&rest.Config{Host: "https://example.invalid"}, clientset)

	if err := client.UpsertNamespaceAnnotations(context.Background(), "kodex-dev-1", map[string]string{
		"existing": "updated",
		"new":      "value-2",
	}); err != nil {
		t.Fatalf("UpsertNamespaceAnnotations() error = %v", err)
	}

	ns, err := clientset.CoreV1().Namespaces().Get(context.Background(), "kodex-dev-1", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("reload namespace: %v", err)
	}
	if got, want := ns.Annotations["existing"], "updated"; got != want {
		t.Fatalf("expected merged annotation %q, got %q", want, got)
	}
	if got, want := ns.Annotations["new"], "value-2"; got != want {
		t.Fatalf("expected new annotation %q, got %q", want, got)
	}
}
