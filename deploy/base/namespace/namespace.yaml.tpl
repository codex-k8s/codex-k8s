apiVersion: v1
kind: Namespace
metadata:
  name: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/part-of: kodex
    app.kubernetes.io/environment: production
