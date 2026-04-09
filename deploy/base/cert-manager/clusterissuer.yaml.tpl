apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: kodex-letsencrypt
spec:
  acme:
    email: '{{ envOr "KODEX_LETSENCRYPT_EMAIL" "" }}'
    server: '{{ envOr "KODEX_LETSENCRYPT_SERVER" "" }}'
    privateKeySecretRef:
      name: kodex-letsencrypt-account-key
    solvers:
      - http01:
          ingress:
            ingressClassName: nginx
