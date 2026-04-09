apiVersion: batch/v1
kind: Job
metadata:
  name: {{ envOr "KODEX_REPO_SYNC_JOB_NAME" "" }}
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex-repo-sync
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 600
  template:
    metadata:
      labels:
        app.kubernetes.io/name: kodex-repo-sync
    spec:
      restartPolicy: Never
      volumes:
        - name: repo-cache
          persistentVolumeClaim:
            claimName: {{ envOr "KODEX_REPO_CACHE_PVC_NAME" "kodex-repo-cache" }}
      containers:
        - name: sync
          image: {{ envOr "KODEX_REPO_SYNC_IMAGE" "127.0.0.1:5000/kodex/mirror/alpine-git:2.47.2" }}
          imagePullPolicy: IfNotPresent
          env:
            - name: GIT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: kodex-git-token
                  key: token
            - name: KODEX_GITHUB_REPO
              value: '{{ envOr "KODEX_GITHUB_REPO" "" }}'
            - name: KODEX_BUILD_REF
              value: '{{ envOr "KODEX_BUILD_REF" "" }}'
            - name: KODEX_REPOSITORY_ROOT
              value: '{{ envOr "KODEX_REPOSITORY_ROOT" "/repo-cache" }}'
            - name: KODEX_REPO_SYNC_DEST_DIR
              value: '{{ envOr "KODEX_REPO_SYNC_DEST_DIR" "" }}'
          command:
            - sh
            - -ec
            - |
              if [ -z "$KODEX_REPO_SYNC_DEST_DIR" ]; then
                echo "KODEX_REPO_SYNC_DEST_DIR is required"
                exit 1
              fi

              if [ -z "$KODEX_REPOSITORY_ROOT" ]; then
                echo "KODEX_REPOSITORY_ROOT is required"
                exit 1
              fi

              if [ -z "$KODEX_GITHUB_REPO" ]; then
                echo "KODEX_GITHUB_REPO is required"
                exit 1
              fi

              if [ -z "$KODEX_BUILD_REF" ]; then
                echo "KODEX_BUILD_REF is required"
                exit 1
              fi

              case "$KODEX_REPO_SYNC_DEST_DIR" in
                "$KODEX_REPOSITORY_ROOT"|"${KODEX_REPOSITORY_ROOT}"/*) ;;
                *)
                  echo "KODEX_REPO_SYNC_DEST_DIR must be under $KODEX_REPOSITORY_ROOT, got: $KODEX_REPO_SYNC_DEST_DIR"
                  exit 1
                  ;;
              esac

              repo_dir="$KODEX_REPO_SYNC_DEST_DIR"
              repo_url="https://x-access-token:$GIT_TOKEN@github.com/$KODEX_GITHUB_REPO.git"

              if [ -d "$repo_dir/.git" ]; then
                echo "Repository snapshot already present at $repo_dir, updating..."
                cd "$repo_dir"
                git remote set-url origin "$repo_url"
                git fetch --prune --tags origin
              else
                if [ "$repo_dir" = "$KODEX_REPOSITORY_ROOT" ]; then
                  mkdir -p "$repo_dir"
                  find "$repo_dir" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
                else
                  rm -rf "$repo_dir"
                  mkdir -p "$(dirname "$repo_dir")"
                fi
                git clone "$repo_url" "$repo_dir"
                cd "$repo_dir"
              fi

              checkout_ref="$KODEX_BUILD_REF"
              if git rev-parse --verify -q "origin/$KODEX_BUILD_REF^{commit}" >/dev/null 2>&1; then
                checkout_ref="origin/$KODEX_BUILD_REF"
              fi

              git checkout --detach "$checkout_ref"
              git reset --hard
              git clean -fdx
          volumeMounts:
            - name: repo-cache
              mountPath: {{ envOr "KODEX_REPOSITORY_ROOT" "/repo-cache" }}
