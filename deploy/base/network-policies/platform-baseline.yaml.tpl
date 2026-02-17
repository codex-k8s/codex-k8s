apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: codex-k8s-postgres-ingress-from-platform
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
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
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
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
    # Keep this rule restrictive by setting CODEXK8S_K8S_API_CIDR to the node IP /32 in production.
    - to:
        - ipBlock:
            cidr: {{ envOr "CODEXK8S_K8S_API_CIDR" "" }}
      ports:
        - protocol: TCP
          port: {{ envOr "CODEXK8S_K8S_API_PORT" "" }}
    - to:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: postgres
      ports:
        - protocol: TCP
          port: 5432
    # Allow api-gateway (and other platform pods) to call control-plane over gRPC/HTTP.
    - to:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: codex-k8s
              app.kubernetes.io/component: control-plane
      ports:
        - protocol: TCP
          port: 9090
        - protocol: TCP
          port: 8081
    # Allow platform pods to query the internal registry (staff: Registry Images).
    - to:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: codex-k8s-registry
      ports:
        - protocol: TCP
          port: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_PORT" "" }}
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
    # Staff UI in production/dev runs as a Vite dev server, and api-gateway reverse-proxies it.
    # Allow egress to the in-cluster web-console service port.
    - to:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: codex-k8s-web-console
      ports:
        - protocol: TCP
          port: 5173
