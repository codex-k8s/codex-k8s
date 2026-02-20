package runtimedeploy

import "testing"

func TestDefaultWorkerReplicas(t *testing.T) {
	t.Parallel()

	assertDefaultWorkerReplicas(t, "production", "2", "3")
	assertDefaultWorkerReplicas(t, "prod", "5", "5")
	assertDefaultWorkerReplicas(t, "ai", "1", "1")
	assertDefaultWorkerReplicas(t, "ai", "", "1")
}

func assertDefaultWorkerReplicas(t *testing.T, targetEnv string, platformReplicas string, want string) {
	t.Helper()

	if got := defaultWorkerReplicas(targetEnv, platformReplicas); got != want {
		t.Fatalf("defaultWorkerReplicas(%q, %q) = %q, want %q", targetEnv, platformReplicas, got, want)
	}
}
