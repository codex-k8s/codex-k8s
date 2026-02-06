apiVersion: v1
kind: Namespace
metadata:
  name: ${RUNNER_NAMESPACE}
  labels:
    app.kubernetes.io/part-of: codex-k8s
    app.kubernetes.io/component: github-runner
