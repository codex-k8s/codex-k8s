apiVersion: batch/v1
kind: Job
metadata:
  name: {{ envOr "CODEXK8S_REPO_SYNC_JOB_NAME" "" }}
  namespace: {{ envOr "CODEXK8S_PLATFORM_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-repo-sync
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 600
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s-repo-sync
    spec:
      restartPolicy: Never
      volumes:
        - name: repo-cache
          persistentVolumeClaim:
            claimName: {{ envOr "CODEXK8S_REPO_CACHE_PVC_NAME" "codex-k8s-repo-cache" }}
      containers:
        - name: sync
          image: {{ envOr "CODEXK8S_REPO_SYNC_IMAGE" "alpine/git:2.47.2" }}
          imagePullPolicy: IfNotPresent
          env:
            - name: GIT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-git-token
                  key: token
            - name: CODEXK8S_GITHUB_REPO
              value: '{{ envOr "CODEXK8S_GITHUB_REPO" "" }}'
            - name: CODEXK8S_BUILD_REF
              value: '{{ envOr "CODEXK8S_BUILD_REF" "" }}'
            - name: CODEXK8S_REPOSITORY_ROOT
              value: '{{ envOr "CODEXK8S_REPOSITORY_ROOT" "/repo-cache" }}'
            - name: CODEXK8S_REPO_SYNC_DEST_DIR
              value: '{{ envOr "CODEXK8S_REPO_SYNC_DEST_DIR" "" }}'
          command:
            - sh
            - -ec
            - |
              if [ -z "$CODEXK8S_REPO_SYNC_DEST_DIR" ]; then
                echo "CODEXK8S_REPO_SYNC_DEST_DIR is required"
                exit 1
              fi

              if [ -z "$CODEXK8S_REPOSITORY_ROOT" ]; then
                echo "CODEXK8S_REPOSITORY_ROOT is required"
                exit 1
              fi

              case "$CODEXK8S_REPO_SYNC_DEST_DIR" in
                "$CODEXK8S_REPOSITORY_ROOT"/*) ;;
                *)
                  echo "CODEXK8S_REPO_SYNC_DEST_DIR must be under $CODEXK8S_REPOSITORY_ROOT, got: $CODEXK8S_REPO_SYNC_DEST_DIR"
                  exit 1
                  ;;
              esac

              if [ -d "$CODEXK8S_REPO_SYNC_DEST_DIR/.git" ]; then
                echo "Repository snapshot already present at $CODEXK8S_REPO_SYNC_DEST_DIR"
                exit 0
              fi

              rm -rf "$CODEXK8S_REPO_SYNC_DEST_DIR"
              mkdir -p "$(dirname "$CODEXK8S_REPO_SYNC_DEST_DIR")"

              git clone "https://x-access-token:$GIT_TOKEN@github.com/$CODEXK8S_GITHUB_REPO.git" "$CODEXK8S_REPO_SYNC_DEST_DIR"
              cd "$CODEXK8S_REPO_SYNC_DEST_DIR"
              git checkout --detach "$CODEXK8S_BUILD_REF"
          volumeMounts:
            - name: repo-cache
              mountPath: {{ envOr "CODEXK8S_REPOSITORY_ROOT" "/repo-cache" }}
