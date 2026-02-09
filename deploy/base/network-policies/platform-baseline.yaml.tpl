apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: codex-k8s-postgres-ingress-from-platform
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: postgres
  policyTypes:
    - Ingress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: codex-k8s
      ports:
        - protocol: TCP
          port: 5432
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: codex-k8s-egress-baseline
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: codex-k8s
  policyTypes:
    - Egress
  egress:
    # NOTE: codex-k8s uses Kubernetes API via in-cluster client-go (kubernetes.default.svc).
    # On some setups (e.g. k3s + certain CNI implementations), egress filtering may be evaluated
    # against the API server endpoint (nodeIP:6443) rather than the Service port (443).
    # Keep this rule restrictive by setting CODEXK8S_K8S_API_CIDR to the node IP /32 in staging.
    - to:
        - ipBlock:
            cidr: ${CODEXK8S_K8S_API_CIDR}
      ports:
        - protocol: TCP
          port: ${CODEXK8S_K8S_API_PORT}
    - to:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: postgres
      ports:
        - protocol: TCP
          port: 5432
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: kube-system
          podSelector:
            matchLabels:
              k8s-app: kube-dns
      ports:
        - protocol: UDP
          port: 53
        - protocol: TCP
          port: 53
    - to:
        - ipBlock:
            cidr: 0.0.0.0/0
      ports:
        - protocol: TCP
          port: 80
        - protocol: TCP
          port: 443
    - to:
        - ipBlock:
            cidr: ::/0
      ports:
        - protocol: TCP
          port: 80
        - protocol: TCP
          port: 443
