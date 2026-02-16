apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: codex-k8s
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  annotations:
    cert-manager.io/cluster-issuer: codex-k8s-letsencrypt
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - {{ envOr "CODEXK8S_PRODUCTION_DOMAIN" "" }}
      secretName: codex-k8s-production-tls
  rules:
    - host: {{ envOr "CODEXK8S_PRODUCTION_DOMAIN" "" }}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: oauth2-proxy
                port:
                  number: 4180
