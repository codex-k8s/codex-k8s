{{- $kanikoCacheEnabled := envOr "KODEX_KANIKO_CACHE_ENABLED" "false" -}}
{{- $kanikoCacheRepo := envOr "KODEX_KANIKO_CACHE_REPO" "" -}}
{{- $kanikoCacheTTL := envOr "KODEX_KANIKO_CACHE_TTL" "" -}}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ envOr "KODEX_KANIKO_JOB_NAME" "" }}
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex-kaniko
    app.kubernetes.io/component: {{ envOr "KODEX_KANIKO_COMPONENT" "" }}
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 600
  template:
    metadata:
      labels:
        app.kubernetes.io/name: kodex-kaniko
        app.kubernetes.io/component: {{ envOr "KODEX_KANIKO_COMPONENT" "" }}
    spec:
      # For single-node production we rely on the node loopback registry and therefore need hostNetwork.
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      restartPolicy: Never
      volumes:
        - name: workspace
          emptyDir: {}
      initContainers:
        - name: clone
          image: {{ envOr "KODEX_KANIKO_CLONE_IMAGE" "127.0.0.1:5000/kodex/mirror/alpine-git:2.47.2" }}
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
          command:
            - sh
            - -ec
            - |
              # Use shell runtime variables from container env (set above via secret refs).
              git clone "https://x-access-token:$GIT_TOKEN@github.com/$KODEX_GITHUB_REPO.git" /workspace
              cd /workspace
              checkout_ref="$KODEX_BUILD_REF"
              if git rev-parse --verify -q "origin/$KODEX_BUILD_REF^{commit}" >/dev/null 2>&1; then
                checkout_ref="origin/$KODEX_BUILD_REF"
              fi
              git checkout --detach "$checkout_ref"
          volumeMounts:
            - name: workspace
              mountPath: /workspace
      containers:
        - name: kaniko
          image: {{ envOr "KODEX_KANIKO_EXECUTOR_IMAGE" "127.0.0.1:5000/kodex/mirror/kaniko-executor:v1.23.2-debug" }}
          imagePullPolicy: IfNotPresent
          args:
            - --context={{ envOr "KODEX_KANIKO_CONTEXT" "" }}
            - --dockerfile={{ envOr "KODEX_KANIKO_DOCKERFILE" "" }}
            - --destination={{ envOr "KODEX_KANIKO_DESTINATION_LATEST" "" }}
            - --destination={{ envOr "KODEX_KANIKO_DESTINATION_SHA" "" }}
            - --cache={{ $kanikoCacheEnabled }}
{{ if ne $kanikoCacheEnabled "false" }}
{{ if ne $kanikoCacheRepo "" }}
            - --cache-repo={{ $kanikoCacheRepo }}
{{ end }}
{{ if ne $kanikoCacheTTL "" }}
            - --cache-ttl={{ $kanikoCacheTTL }}
{{ end }}
            - --compressed-caching={{ envOr "KODEX_KANIKO_CACHE_COMPRESSED" "false" }}
            - --cache-copy-layers={{ envOr "KODEX_KANIKO_CACHE_COPY_LAYERS" "true" }}
{{ end }}
            - --snapshot-mode={{ envOr "KODEX_KANIKO_SNAPSHOT_MODE" "redo" }}
            - --single-snapshot={{ envOr "KODEX_KANIKO_SINGLE_SNAPSHOT" "true" }}
            - --use-new-run={{ envOr "KODEX_KANIKO_USE_NEW_RUN" "true" }}
            - --verbosity={{ envOr "KODEX_KANIKO_VERBOSITY" "info" }}
            - --cleanup={{ envOr "KODEX_KANIKO_CLEANUP" "true" }}
            - --insecure
            - --insecure-registry={{ envOr "KODEX_INTERNAL_REGISTRY_HOST" "" }}
            - --skip-tls-verify-registry={{ envOr "KODEX_INTERNAL_REGISTRY_HOST" "" }}
          resources:
            requests:
              cpu: '{{ envOr "KODEX_KANIKO_CPU_REQUEST" "4" }}'
              memory: '{{ envOr "KODEX_KANIKO_MEMORY_REQUEST" "8Gi" }}'
            limits:
              cpu: '{{ envOr "KODEX_KANIKO_CPU_LIMIT" "16" }}'
              memory: '{{ envOr "KODEX_KANIKO_MEMORY_LIMIT" "32Gi" }}'
          volumeMounts:
            - name: workspace
              mountPath: /workspace
