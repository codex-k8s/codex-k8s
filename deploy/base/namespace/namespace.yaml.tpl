apiVersion: v1
kind: Namespace
metadata:
  name: ${STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/part-of: codex-k8s
    app.kubernetes.io/environment: ai-staging
