apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: codex-k8s
  namespace: ${STAGING_NAMESPACE}
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
    - host: ${STAGING_DOMAIN}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: codex-k8s
                port:
                  number: 80
