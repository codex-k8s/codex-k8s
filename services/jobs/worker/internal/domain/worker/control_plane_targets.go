package worker

import (
	"fmt"
	"net"
	"strings"

	agentdomain "github.com/codex-k8s/codex-k8s/libs/go/domain/agent"
)

const (
	controlPlaneServiceName = "codex-k8s-control-plane"
	clusterLocalSuffix      = "svc.cluster.local"
	controlPlaneGRPCPort    = "9090"
)

func resolveRunControlPlaneGRPCTarget(runtimeMode agentdomain.RuntimeMode, productionNamespace string, fallbackTarget string) string {
	if runtimeMode != agentdomain.RuntimeModeFullEnv {
		return strings.TrimSpace(fallbackTarget)
	}

	namespace := strings.TrimSpace(productionNamespace)
	if namespace == "" {
		return strings.TrimSpace(fallbackTarget)
	}

	return net.JoinHostPort(
		fmt.Sprintf("%s.%s.%s", controlPlaneServiceName, namespace, clusterLocalSuffix),
		controlPlaneGRPCPort,
	)
}
