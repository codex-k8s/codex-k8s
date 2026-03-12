package worker

import "context"

// WorkerPresenceChecker lists active worker instances visible in Kubernetes.
type WorkerPresenceChecker interface {
	ListActiveWorkerIDs(ctx context.Context) ([]string, error)
}
