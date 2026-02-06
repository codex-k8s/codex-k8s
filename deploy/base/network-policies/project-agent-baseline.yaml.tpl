apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: codexk8s-default-deny-all
  namespace: ${CODEXK8S_TARGET_NAMESPACE}
spec:
  podSelector: {}
  policyTypes:
    - Ingress
    - Egress
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: codexk8s-allow-ingress-from-platform-and-system
  namespace: ${CODEXK8S_TARGET_NAMESPACE}
spec:
  podSelector: {}
  policyTypes:
    - Ingress
  ingress:
    - from:
        - namespaceSelector:
            matchExpressions:
              - key: codexk8s.io/network-zone
                operator: In
                values:
                  - platform
                  - system
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: codexk8s-allow-egress-dns
  namespace: ${CODEXK8S_TARGET_NAMESPACE}
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
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
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: codexk8s-allow-egress-to-platform-mcp
  namespace: ${CODEXK8S_TARGET_NAMESPACE}
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              codexk8s.io/network-zone: platform
          podSelector:
            matchLabels:
              app.kubernetes.io/name: codex-k8s
      ports:
        - protocol: TCP
          port: ${CODEXK8S_PLATFORM_MCP_PORT}
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: codexk8s-allow-egress-web
  namespace: ${CODEXK8S_TARGET_NAMESPACE}
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
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
