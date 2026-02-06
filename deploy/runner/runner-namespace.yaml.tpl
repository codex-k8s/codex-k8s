apiVersion: v1
kind: Namespace
metadata:
  name: ${CODEXK8S_RUNNER_NAMESPACE}
  labels:
    app.kubernetes.io/part-of: codex-k8s
    app.kubernetes.io/component: github-runner
