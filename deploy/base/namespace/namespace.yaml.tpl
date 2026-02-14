apiVersion: v1
kind: Namespace
metadata:
  name: {{ include "namespace.staging" . }}
  labels:
{{ include "labels.platform" . | indent 4 }}
    app.kubernetes.io/environment: {{ .Environment }}
