apiVersion: batch/v1
kind: Job
metadata:
  name: {{ envOr "CODEXK8S_KANIKO_JOB_NAME" "" }}
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-kaniko
    app.kubernetes.io/component: {{ envOr "CODEXK8S_KANIKO_COMPONENT" "" }}
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 600
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s-kaniko
        app.kubernetes.io/component: {{ envOr "CODEXK8S_KANIKO_COMPONENT" "" }}
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
          image: {{ envOr "CODEXK8S_KANIKO_CLONE_IMAGE" "127.0.0.1:5000/codex-k8s/mirror/alpine-git:2.47.2" }}
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
          command:
            - sh
            - -ec
            - |
              # Use shell runtime variables from container env (set above via secret refs).
              git clone "https://x-access-token:$GIT_TOKEN@github.com/$CODEXK8S_GITHUB_REPO.git" /workspace
              cd /workspace
              git checkout --detach "$CODEXK8S_BUILD_REF"
          volumeMounts:
            - name: workspace
              mountPath: /workspace
      containers:
        - name: kaniko
          image: {{ envOr "CODEXK8S_KANIKO_EXECUTOR_IMAGE" "127.0.0.1:5000/codex-k8s/mirror/kaniko-executor:v1.23.2-debug" }}
          imagePullPolicy: IfNotPresent
          args:
            - --context={{ envOr "CODEXK8S_KANIKO_CONTEXT" "" }}
            - --dockerfile={{ envOr "CODEXK8S_KANIKO_DOCKERFILE" "" }}
            - --destination={{ envOr "CODEXK8S_KANIKO_DESTINATION_LATEST" "" }}
            - --destination={{ envOr "CODEXK8S_KANIKO_DESTINATION_SHA" "" }}
            - --cache={{ envOr "CODEXK8S_KANIKO_CACHE_ENABLED" "" }}
            - --cache-repo={{ envOr "CODEXK8S_KANIKO_CACHE_REPO" "" }}
            - --cache-ttl={{ envOr "CODEXK8S_KANIKO_CACHE_TTL" "" }}
            - --compressed-caching={{ envOr "CODEXK8S_KANIKO_CACHE_COMPRESSED" "" }}
            - --cache-copy-layers={{ envOr "CODEXK8S_KANIKO_CACHE_COPY_LAYERS" "true" }}
            - --snapshot-mode={{ envOr "CODEXK8S_KANIKO_SNAPSHOT_MODE" "redo" }}
            - --single-snapshot={{ envOr "CODEXK8S_KANIKO_SINGLE_SNAPSHOT" "true" }}
            - --use-new-run={{ envOr "CODEXK8S_KANIKO_USE_NEW_RUN" "true" }}
            - --verbosity={{ envOr "CODEXK8S_KANIKO_VERBOSITY" "info" }}
            - --cleanup={{ envOr "CODEXK8S_KANIKO_CLEANUP" "true" }}
            - --insecure
            - --insecure-registry={{ envOr "CODEXK8S_INTERNAL_REGISTRY_HOST" "" }}
            - --skip-tls-verify-registry={{ envOr "CODEXK8S_INTERNAL_REGISTRY_HOST" "" }}
          resources:
            requests:
              cpu: '{{ envOr "CODEXK8S_KANIKO_CPU_REQUEST" "4" }}'
              memory: '{{ envOr "CODEXK8S_KANIKO_MEMORY_REQUEST" "8Gi" }}'
            limits:
              cpu: '{{ envOr "CODEXK8S_KANIKO_CPU_LIMIT" "16" }}'
              memory: '{{ envOr "CODEXK8S_KANIKO_MEMORY_LIMIT" "32Gi" }}'
          volumeMounts:
            - name: workspace
              mountPath: /workspace
