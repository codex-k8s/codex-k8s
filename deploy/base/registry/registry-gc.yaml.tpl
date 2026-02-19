apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}-gc
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-registry-gc
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}-gc
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-registry-gc
rules:
  - apiGroups: ["apps"]
    resources:
      - deployments
      - deployments/scale
    verbs:
      - get
      - list
      - watch
      - patch
      - update
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}-gc
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-registry-gc
subjects:
  - kind: ServiceAccount
    name: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}-gc
    namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}-gc
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}-gc
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-registry-gc
spec:
  # UTC by default: every day at 03:17.
  schedule: "17 3 * * *"
  concurrencyPolicy: Forbid
  suspend: false
  successfulJobsHistoryLimit: 2
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      backoffLimit: 0
      ttlSecondsAfterFinished: 86400
      activeDeadlineSeconds: 7200
      template:
        metadata:
          labels:
            app.kubernetes.io/name: codex-k8s-registry-gc
        spec:
          serviceAccountName: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}-gc
          restartPolicy: Never
          volumes:
            - name: registry-data
              persistentVolumeClaim:
                claimName: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}-data
            - name: tools
              emptyDir: {}
          initContainers:
            - name: copy-kubectl
              image: {{ envOr "CODEXK8S_KUBECTL_IMAGE" "127.0.0.1:5000/codex-k8s/mirror/kubectl:v1.32.2" }}
              imagePullPolicy: IfNotPresent
              command:
                - sh
                - -ec
                - |
                  set -eu
                  for candidate in /usr/local/bin/kubectl /usr/bin/kubectl /bin/kubectl; do
                    if [ -x "$candidate" ]; then
                      cp "$candidate" /tools/kubectl
                      chmod +x /tools/kubectl
                      exit 0
                    fi
                  done
                  echo "kubectl binary not found in image" >&2
                  exit 1
              volumeMounts:
                - name: tools
                  mountPath: /tools
          containers:
            - name: gc
              image: {{ envOr "CODEXK8S_REGISTRY_IMAGE" "registry:2" }}
              imagePullPolicy: IfNotPresent
              command:
                - sh
                - -ec
                - |
                  set -eu

                  REGISTRY_DEPLOYMENT='{{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}'
                  TARGET_NAMESPACE='{{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}'
                  TARGET_REPLICAS='1'
                  ROLLOUT_TIMEOUT='15m'
                  KUBECTL_BIN='/tools/kubectl'

                  if [ ! -x "$KUBECTL_BIN" ]; then
                    echo "kubectl binary is missing at $KUBECTL_BIN" >&2
                    exit 1
                  fi

                  restore_registry() {
                    "$KUBECTL_BIN" -n "$TARGET_NAMESPACE" scale deployment "$REGISTRY_DEPLOYMENT" --replicas="$TARGET_REPLICAS" >/dev/null 2>&1 || true
                  }

                  trap restore_registry EXIT

                  echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] scale registry down"
                  "$KUBECTL_BIN" -n "$TARGET_NAMESPACE" scale deployment "$REGISTRY_DEPLOYMENT" --replicas=0
                  "$KUBECTL_BIN" -n "$TARGET_NAMESPACE" rollout status deployment "$REGISTRY_DEPLOYMENT" --timeout="$ROLLOUT_TIMEOUT"

                  echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] run registry garbage-collect"
                  registry garbage-collect --delete-untagged /etc/docker/registry/config.yml

                  echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] scale registry up"
                  "$KUBECTL_BIN" -n "$TARGET_NAMESPACE" scale deployment "$REGISTRY_DEPLOYMENT" --replicas="$TARGET_REPLICAS"
                  "$KUBECTL_BIN" -n "$TARGET_NAMESPACE" rollout status deployment "$REGISTRY_DEPLOYMENT" --timeout="$ROLLOUT_TIMEOUT"

                  trap - EXIT
              volumeMounts:
                - name: registry-data
                  mountPath: /var/lib/registry
                - name: tools
                  mountPath: /tools
