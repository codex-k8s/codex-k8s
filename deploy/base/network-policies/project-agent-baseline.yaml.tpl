apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kodex-default-deny-all
  namespace: {{ envOr "KODEX_TARGET_NAMESPACE" "" }}
spec:
  podSelector: {}
  policyTypes:
    - Ingress
    - Egress
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kodex-allow-ingress-from-platform-and-system
  namespace: {{ envOr "KODEX_TARGET_NAMESPACE" "" }}
spec:
  podSelector: {}
  policyTypes:
    - Ingress
  ingress:
    - from:
        - namespaceSelector:
            matchExpressions:
              - key: kodex.io/network-zone
                operator: In
                values:
                  - platform
                  - system
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kodex-allow-egress-dns
  namespace: {{ envOr "KODEX_TARGET_NAMESPACE" "" }}
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
  name: kodex-allow-egress-to-platform-mcp
  namespace: {{ envOr "KODEX_TARGET_NAMESPACE" "" }}
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              kodex.io/network-zone: platform
          podSelector:
            matchLabels:
              app.kubernetes.io/name: kodex
      ports:
        - protocol: TCP
          port: {{ envOr "KODEX_PLATFORM_MCP_PORT" "" }}
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kodex-allow-egress-web
  namespace: {{ envOr "KODEX_TARGET_NAMESPACE" "" }}
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
