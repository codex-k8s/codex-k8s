apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: codex-k8s-staging-deployer
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
rules:
  - apiGroups: [""]
    resources: ["secrets", "configmaps", "services", "pods", "pods/log", "persistentvolumeclaims", "serviceaccounts"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "replicasets"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["batch"]
    resources: ["jobs"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["ingresses", "networkpolicies"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles", "rolebindings"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: codex-k8s-staging-deployer
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
subjects:
  - kind: ServiceAccount
    name: ${CODEXK8S_RUNNER_SCALE_SET_NAME}-gha-rs-no-permission
    namespace: ${CODEXK8S_RUNNER_NAMESPACE}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: codex-k8s-staging-deployer
