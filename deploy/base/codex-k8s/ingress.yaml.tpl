apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: codex-k8s
  namespace: {{ include "namespace.staging" . }}
  labels:
{{ include "labels.platform" . | indent 4 }}
  annotations:
    cert-manager.io/cluster-issuer: codex-k8s-letsencrypt
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - ${CODEXK8S_STAGING_DOMAIN}
      secretName: codex-k8s-staging-tls
  rules:
    - host: ${CODEXK8S_STAGING_DOMAIN}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: oauth2-proxy
                port:
                  number: 4180
