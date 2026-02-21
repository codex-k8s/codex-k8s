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
              set -eu
              if [ -z "${SOURCE_IMAGE:-}" ] || [ -z "${TARGET_IMAGE:-}" ]; then
                echo "SOURCE_IMAGE and TARGET_IMAGE are required" >&2
                exit 1
              fi

              if digest="$(crane digest --insecure "$TARGET_IMAGE" 2>/dev/null)"; then
                target_no_digest="${TARGET_IMAGE%@*}"
                target_last_segment="${target_no_digest##*/}"
                target_repo="$target_no_digest"
                if [ "${target_last_segment#*:}" != "$target_last_segment" ]; then
                  target_repo="${target_no_digest%:*}"
                fi
                if crane manifest --insecure "${target_repo}@${digest}" >/dev/null 2>&1; then
                  echo "Mirror is healthy: ${TARGET_IMAGE} (${digest})"
                  exit 0
                fi
                echo "Mirror tag exists but digest is stale, repairing: ${TARGET_IMAGE} (${digest})"
              else
                echo "Mirror is missing, syncing: ${TARGET_IMAGE}"
              fi

              crane copy --insecure "$SOURCE_IMAGE" "$TARGET_IMAGE"
