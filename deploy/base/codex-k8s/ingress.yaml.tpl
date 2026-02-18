{{- $host := envOr "CODEXK8S_PUBLIC_DOMAIN" (envOr "CODEXK8S_PRODUCTION_DOMAIN" "") -}}
{{- $tlsSecret := envOr "CODEXK8S_TLS_SECRET_NAME" "codex-k8s-production-tls" -}}
{{- $certManager := eq (envOr "CODEXK8S_CERT_MANAGER_ANNOTATE" "false") "true" -}}
{{- $isAI := eq (toLower (envOr "CODEXK8S_ENV" "")) "ai" -}}
{{- $sharedAuthURL := envOr "CODEXK8S_SHARED_OAUTH2_PROXY_AUTH_URL" "" -}}
{{- $sharedSigninURL := envOr "CODEXK8S_SHARED_OAUTH2_PROXY_SIGNIN_URL" "" -}}
{{- $useSharedOAuth := and $isAI (ne $sharedAuthURL "") (ne $sharedSigninURL "") -}}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: codex-k8s
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
{{- if or $certManager $useSharedOAuth }}
  annotations:
{{- if $certManager }}
    cert-manager.io/cluster-issuer: codex-k8s-letsencrypt
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
                name: codex-k8s
                port:
                  number: 80
{{- else }}
                name: oauth2-proxy
                port:
                  number: 4180
{{- end }}
