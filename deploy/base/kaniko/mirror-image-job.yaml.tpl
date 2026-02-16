apiVersion: batch/v1
kind: Job
metadata:
  name: {{ envOr "CODEXK8S_IMAGE_MIRROR_JOB_NAME" "" }}
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-image-mirror
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 300
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s-image-mirror
    spec:
      # For single-node production we mirror into node-local loopback registry.
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      restartPolicy: Never
      containers:
        - name: mirror
          image: {{ envOr "CODEXK8S_IMAGE_MIRROR_TOOL_IMAGE" "gcr.io/go-containerregistry/crane:debug" }}
          imagePullPolicy: IfNotPresent
          env:
            - name: SOURCE_IMAGE
              value: '{{ envOr "CODEXK8S_IMAGE_MIRROR_SOURCE" "" }}'
            - name: TARGET_IMAGE
              value: '{{ envOr "CODEXK8S_IMAGE_MIRROR_TARGET" "" }}'
          command:
            - sh
            - -ec
            - |
              crane digest --insecure "$TARGET_IMAGE" >/dev/null 2>&1 || \
              crane copy --insecure "$SOURCE_IMAGE" "$TARGET_IMAGE"
