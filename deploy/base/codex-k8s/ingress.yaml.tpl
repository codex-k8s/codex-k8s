{{- $host := envOr "CODEXK8S_PUBLIC_DOMAIN" (envOr "CODEXK8S_PRODUCTION_DOMAIN" "") -}}
{{- $tlsSecret := envOr "CODEXK8S_TLS_SECRET_NAME" "codex-k8s-production-tls" -}}
{{- $certManager := eq (envOr "CODEXK8S_CERT_MANAGER_ANNOTATE" "false") "true" -}}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: codex-k8s
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
{{- if $certManager }}
  annotations:
    cert-manager.io/cluster-issuer: codex-k8s-letsencrypt
{{- end }}
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - {{ $host }}
      secretName: {{ $tlsSecret }}
  rules:
    - host: {{ $host }}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: oauth2-proxy
                port:
                  number: 4180
