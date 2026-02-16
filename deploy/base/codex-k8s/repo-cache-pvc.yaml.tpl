apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: codex-k8s-repo-cache
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: repo-cache
spec:
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      storage: {{ envOr "CODEXK8S_REPO_CACHE_STORAGE_SIZE" "10Gi" }}
