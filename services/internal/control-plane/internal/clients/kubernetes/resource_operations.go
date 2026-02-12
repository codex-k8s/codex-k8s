package kubernetes

import mcpdomain "github.com/codex-k8s/codex-k8s/services/internal/control-plane/internal/domain/mcp"

type resourceListOperation string

var resourceListOperationByKind = map[mcpdomain.KubernetesResourceKind]resourceListOperation{
	mcpdomain.KubernetesResourceKindDeployment:            "list deployments",
	mcpdomain.KubernetesResourceKindDaemonSet:             "list daemonsets",
	mcpdomain.KubernetesResourceKindStatefulSet:           "list statefulsets",
	mcpdomain.KubernetesResourceKindReplicaSet:            "list replicasets",
	mcpdomain.KubernetesResourceKindReplicationController: "list replicationcontrollers",
	mcpdomain.KubernetesResourceKindJob:                   "list jobs",
	mcpdomain.KubernetesResourceKindCronJob:               "list cronjobs",
	mcpdomain.KubernetesResourceKindConfigMap:             "list configmaps",
	mcpdomain.KubernetesResourceKindSecret:                "list secrets",
	mcpdomain.KubernetesResourceKindResourceQuota:         "list resourcequotas",
	mcpdomain.KubernetesResourceKindHPA:                   "list hpas",
	mcpdomain.KubernetesResourceKindService:               "list services",
	mcpdomain.KubernetesResourceKindEndpoints:             "list endpoints",
	mcpdomain.KubernetesResourceKindIngress:               "list ingresses",
	mcpdomain.KubernetesResourceKindIngressClass:          "list ingressclasses",
	mcpdomain.KubernetesResourceKindNetworkPolicy:         "list networkpolicies",
	mcpdomain.KubernetesResourceKindPVC:                   "list pvcs",
	mcpdomain.KubernetesResourceKindPV:                    "list pvs",
	mcpdomain.KubernetesResourceKindStorageClass:          "list storageclasses",
}

func resourceListOperationForKind(kind mcpdomain.KubernetesResourceKind) resourceListOperation {
	operation, ok := resourceListOperationByKind[kind]
	if ok {
		return operation
	}
	return resourceListOperation("list resources")
}
