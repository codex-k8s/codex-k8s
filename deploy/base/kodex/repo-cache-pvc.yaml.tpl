apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: kodex-repo-cache
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: repo-cache
spec:
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      storage: {{ envOr "KODEX_REPO_CACHE_STORAGE_SIZE" "10Gi" }}
