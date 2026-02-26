apiVersion: batch/v1
kind: Job
metadata:
  name: {{ envOr "CODEXK8S_REPO_SYNC_JOB_NAME" "" }}
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
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
          image: {{ envOr "CODEXK8S_REPO_SYNC_IMAGE" "127.0.0.1:5000/codex-k8s/mirror/alpine-git:2.47.2" }}
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

              if [ -z "$CODEXK8S_GITHUB_REPO" ]; then
                echo "CODEXK8S_GITHUB_REPO is required"
                exit 1
              fi

              if [ -z "$CODEXK8S_BUILD_REF" ]; then
                echo "CODEXK8S_BUILD_REF is required"
                exit 1
              fi

              case "$CODEXK8S_REPO_SYNC_DEST_DIR" in
                "$CODEXK8S_REPOSITORY_ROOT"|"${CODEXK8S_REPOSITORY_ROOT}"/*) ;;
                *)
                  echo "CODEXK8S_REPO_SYNC_DEST_DIR must be under $CODEXK8S_REPOSITORY_ROOT, got: $CODEXK8S_REPO_SYNC_DEST_DIR"
                  exit 1
                  ;;
              esac

              repo_dir="$CODEXK8S_REPO_SYNC_DEST_DIR"
              repo_url="https://x-access-token:$GIT_TOKEN@github.com/$CODEXK8S_GITHUB_REPO.git"

              if [ -d "$repo_dir/.git" ]; then
                echo "Repository snapshot already present at $repo_dir, updating..."
                cd "$repo_dir"
                git remote set-url origin "$repo_url"
                git fetch --prune --tags origin
              else
                rm -rf "$repo_dir"
                mkdir -p "$(dirname "$repo_dir")"
                git clone "$repo_url" "$repo_dir"
                cd "$repo_dir"
              fi

              checkout_ref="$CODEXK8S_BUILD_REF"
              if git rev-parse --verify -q "origin/$CODEXK8S_BUILD_REF^{commit}" >/dev/null 2>&1; then
                checkout_ref="origin/$CODEXK8S_BUILD_REF"
              fi

              git checkout --detach "$checkout_ref"
              git reset --hard
              git clean -fdx
          volumeMounts:
            - name: repo-cache
              mountPath: {{ envOr "CODEXK8S_REPOSITORY_ROOT" "/repo-cache" }}
