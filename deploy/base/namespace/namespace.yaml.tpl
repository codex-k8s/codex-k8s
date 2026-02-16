apiVersion: v1
kind: Namespace
metadata:
  name: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/part-of: codex-k8s
    app.kubernetes.io/environment: production
