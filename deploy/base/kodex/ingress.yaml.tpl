{{- $host := envOr "KODEX_PUBLIC_DOMAIN" (envOr "KODEX_PRODUCTION_DOMAIN" "") -}}
{{- $tlsSecret := envOr "KODEX_TLS_SECRET_NAME" "kodex-production-tls" -}}
{{- $certManager := eq (envOr "KODEX_CERT_MANAGER_ANNOTATE" "false") "true" -}}
{{- $isAI := eq (toLower (envOr "KODEX_ENV" "")) "ai" -}}
{{- $sharedAuthURL := envOr "KODEX_SHARED_OAUTH2_PROXY_AUTH_URL" "" -}}
{{- $sharedSigninURL := envOr "KODEX_SHARED_OAUTH2_PROXY_SIGNIN_URL" "" -}}
{{- $useSharedOAuth := and $isAI (ne $sharedAuthURL "") (ne $sharedSigninURL "") -}}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kodex
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
{{- if or $certManager $useSharedOAuth }}
  annotations:
{{- if $certManager }}
    cert-manager.io/cluster-issuer: kodex-letsencrypt
{{- end }}
{{- if $useSharedOAuth }}
    nginx.ingress.kubernetes.io/auth-url: '{{ $sharedAuthURL }}'
    nginx.ingress.kubernetes.io/auth-signin: '{{ $sharedSigninURL }}'
    nginx.ingress.kubernetes.io/auth-response-headers: X-Auth-Request-User,X-Auth-Request-Email,Authorization
{{- end }}
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
{{- if $useSharedOAuth }}
                name: kodex
                port:
                  number: 80
{{- else }}
                name: oauth2-proxy
                port:
                  number: 4180
{{- end }}
---
{{- if not $isAI }}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kodex-telegram-webhook
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
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
          - path: /api/v1/telegram/interactions/webhook
            pathType: Exact
            backend:
              service:
                name: kodex-telegram-interaction-adapter
                port:
                  number: 8080
{{- end }}
