apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: codex-k8s-letsencrypt
spec:
  acme:
    email: '{{ envOr "CODEXK8S_LETSENCRYPT_EMAIL" "" }}'
    server: '{{ envOr "CODEXK8S_LETSENCRYPT_SERVER" "" }}'
    privateKeySecretRef:
      name: codex-k8s-letsencrypt-account-key
    solvers:
      - http01:
          ingress:
            ingressClassName: nginx
