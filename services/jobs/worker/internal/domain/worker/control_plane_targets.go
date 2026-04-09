package worker

import (
	"fmt"
	"net"
	"strings"
)

const (
	controlPlaneServiceName = "kodex-control-plane"
	clusterLocalSuffix      = "svc.cluster.local"
	controlPlaneGRPCPort    = "9090"
)

func resolveRunControlPlaneGRPCTarget(productionNamespace string, fallbackTarget string) string {
	namespace := strings.TrimSpace(productionNamespace)
	if namespace == "" {
		return strings.TrimSpace(fallbackTarget)
	}

	return net.JoinHostPort(
		fmt.Sprintf("%s.%s.%s", controlPlaneServiceName, namespace, clusterLocalSuffix),
		controlPlaneGRPCPort,
	)
}
