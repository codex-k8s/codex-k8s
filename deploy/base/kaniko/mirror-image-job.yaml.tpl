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
            - name: MIRROR_PLATFORM
              value: '{{ envOr "CODEXK8S_IMAGE_MIRROR_PLATFORM" "linux/amd64" }}'
          command:
            - sh
            - -ec
            - |
              set -eu
              if [ -z "${SOURCE_IMAGE:-}" ] || [ -z "${TARGET_IMAGE:-}" ]; then
                echo "SOURCE_IMAGE and TARGET_IMAGE are required" >&2
                exit 1
              fi

              platform="${MIRROR_PLATFORM:-linux/amd64}"
              if crane manifest --insecure --platform "$platform" "$TARGET_IMAGE" >/dev/null 2>&1; then
                if digest="$(crane digest --insecure "$TARGET_IMAGE" 2>/dev/null)"; then
                  echo "Mirror is healthy for platform ${platform}: ${TARGET_IMAGE} (${digest})"
                else
                  echo "Mirror is healthy for platform ${platform}: ${TARGET_IMAGE}"
                fi
                exit 0
              fi

              if digest="$(crane digest --insecure "$TARGET_IMAGE" 2>/dev/null)"; then
                echo "Mirror tag exists but platform manifest is stale, repairing: ${TARGET_IMAGE} (${digest}) platform=${platform}"
              else
                echo "Mirror is missing, syncing: ${TARGET_IMAGE} platform=${platform}"
              fi

              crane copy --insecure --platform "$platform" "$SOURCE_IMAGE" "$TARGET_IMAGE"
