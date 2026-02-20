package runtimedeploy

import (
	"testing"
	"time"
)

func TestRuntimeDeployLeaseTiming(t *testing.T) {
	t.Parallel()

	if got := runtimeDeployLeaseRenewInterval(10 * time.Minute); got != 30*time.Second {
		t.Fatalf("runtimeDeployLeaseRenewInterval(10m) = %s, want %s", got, 30*time.Second)
	}
	if got := runtimeDeployLeaseRenewInterval(20 * time.Second); got != 10*time.Second {
		t.Fatalf("runtimeDeployLeaseRenewInterval(20s) = %s, want %s", got, 10*time.Second)
	}
	if got := runtimeDeployLeaseRenewInterval(500 * time.Millisecond); got != time.Second {
		t.Fatalf("runtimeDeployLeaseRenewInterval(500ms) = %s, want %s", got, time.Second)
	}

	if got := runtimeDeployStaleRunningTimeout(30 * time.Second); got != 65*time.Second {
		t.Fatalf("runtimeDeployStaleRunningTimeout(30s) = %s, want %s", got, 65*time.Second)
	}
	if got := runtimeDeployStaleRunningTimeout(10 * time.Second); got != 30*time.Second {
		t.Fatalf("runtimeDeployStaleRunningTimeout(10s) = %s, want %s", got, 30*time.Second)
	}
	if got := runtimeDeployStaleRunningTimeout(2 * time.Minute); got != 2*time.Minute {
		t.Fatalf("runtimeDeployStaleRunningTimeout(2m) = %s, want %s", got, 2*time.Minute)
	}
}
