apiVersion: v1
kind: Service
metadata:
  name: kodex
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
spec:
  selector:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: api-gateway
  ports:
    - name: http
      port: 80
      targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kodex
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
spec:
  replicas: {{ envOr "KODEX_PLATFORM_DEPLOYMENT_REPLICAS" "" }}
  selector:
    matchLabels:
      app.kubernetes.io/name: kodex
      app.kubernetes.io/component: api-gateway
  template:
    metadata:
      labels:
        app.kubernetes.io/name: kodex
        app.kubernetes.io/component: api-gateway
    spec:
      initContainers:
        - name: wait-control-plane
          image: {{ envOr "KODEX_BUSYBOX_IMAGE" "busybox:1.36" }}
          imagePullPolicy: IfNotPresent
          command:
            - sh
            - -ec
            - |
              until wget -q -O /dev/null http://kodex-control-plane:8081/health/readyz; do
                echo "waiting for control-plane readiness..."
                sleep 2
              done
      containers:
        - name: kodex
          image: {{ envOr "KODEX_API_GATEWAY_IMAGE" "" }}
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
              name: http
          env:
            - name: KODEX_ENV
              value: '{{ envOr "KODEX_ENV" "" }}'
            - name: KODEX_HOT_RELOAD
              value: '{{ envOr "KODEX_HOT_RELOAD" "" }}'
            - name: KODEX_HTTP_ADDR
              value: ":8080"
            - name: KODEX_CONTROL_PLANE_GRPC_TARGET
              value: '{{ envOr "KODEX_CONTROL_PLANE_GRPC_TARGET" "kodex-control-plane:9090" }}'
            - name: KODEX_VITE_DEV_UPSTREAM
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_VITE_DEV_UPSTREAM
                  optional: true
            - name: KODEX_GITHUB_WEBHOOK_SECRET
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GITHUB_WEBHOOK_SECRET
            - name: KODEX_PUBLIC_BASE_URL
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_PUBLIC_BASE_URL
            - name: KODEX_COOKIE_SECURE
              value: "true"
            - name: KODEX_GITHUB_OAUTH_CLIENT_ID
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GITHUB_OAUTH_CLIENT_ID
            - name: KODEX_GITHUB_OAUTH_CLIENT_SECRET
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GITHUB_OAUTH_CLIENT_SECRET
            - name: KODEX_JWT_SIGNING_KEY
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_JWT_SIGNING_KEY
            - name: KODEX_JWT_TTL
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_JWT_TTL
{{ if eq (envOr "KODEX_HOT_RELOAD" "") "true" }}
          volumeMounts:
            - name: repo-cache
              mountPath: /workspace
{{ end }}
          readinessProbe:
            httpGet:
              path: /readyz
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          startupProbe:
            httpGet:
              path: /healthz
              port: http
            periodSeconds: 5
            failureThreshold: 60
          livenessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 15
            periodSeconds: 20
{{ if eq (envOr "KODEX_HOT_RELOAD" "") "true" }}
      volumes:
        - name: repo-cache
          persistentVolumeClaim:
            claimName: kodex-repo-cache
{{ end }}
---
apiVersion: v1
kind: Service
metadata:
  name: kodex-control-plane
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
spec:
  selector:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: control-plane
  ports:
    - name: grpc
      port: 9090
      targetPort: 9090
    - name: http
      port: 8081
      targetPort: 8081
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kodex-control-plane
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
spec:
  replicas: {{ envOr "KODEX_PLATFORM_DEPLOYMENT_REPLICAS" "" }}
  selector:
    matchLabels:
      app.kubernetes.io/name: kodex
      app.kubernetes.io/component: control-plane
  template:
    metadata:
      labels:
        app.kubernetes.io/name: kodex
        app.kubernetes.io/component: control-plane
    spec:
      serviceAccountName: kodex-control-plane
      initContainers:
        - name: wait-postgres
          image: {{ envOr "KODEX_BUSYBOX_IMAGE" "busybox:1.36" }}
          imagePullPolicy: IfNotPresent
          command:
            - sh
            - -ec
            - |
              until nc -z postgres 5432; do
                echo "waiting for postgres tcp:5432..."
                sleep 2
              done
      containers:
        - name: control-plane
          image: {{ envOr "KODEX_CONTROL_PLANE_IMAGE" "" }}
          imagePullPolicy: Always
          ports:
            - containerPort: 9090
              name: grpc
            - containerPort: 8081
              name: http
          env:
            - name: KODEX_PLATFORM_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: KODEX_ENV
              value: '{{ envOr "KODEX_ENV" "" }}'
            - name: KODEX_HOT_RELOAD
              value: '{{ envOr "KODEX_HOT_RELOAD" "" }}'
            - name: KODEX_SERVICES_CONFIG_PATH
              value: '{{ envOr "KODEX_SERVICES_CONFIG_PATH" "services.yaml" }}'
            - name: KODEX_SERVICES_CONFIG_ENV
              value: '{{ envOr "KODEX_SERVICES_CONFIG_ENV" "" }}'
            - name: KODEX_REPOSITORY_ROOT
              value: '{{ envOr "KODEX_REPOSITORY_ROOT" "/repo-cache" }}'
            - name: KODEX_KANIKO_CACHE_ENABLED
              value: '{{ envOr "KODEX_KANIKO_CACHE_ENABLED" "false" }}'
            - name: KODEX_KANIKO_CACHE_REPO
              value: '{{ envOr "KODEX_KANIKO_CACHE_REPO" "" }}'
            - name: KODEX_KANIKO_CACHE_TTL
              value: '{{ envOr "KODEX_KANIKO_CACHE_TTL" "" }}'
            - name: KODEX_KANIKO_MAX_PARALLEL
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_KANIKO_MAX_PARALLEL
                  optional: true
            - name: KODEX_IMAGE_MIRROR_MAX_PARALLEL
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_IMAGE_MIRROR_MAX_PARALLEL
                  optional: true
            - name: KODEX_RUNTIME_DEPLOY_WORKERS_PER_POD
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_RUNTIME_DEPLOY_WORKERS_PER_POD
                  optional: true
            - name: KODEX_PRODUCTION_DOMAIN
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_PRODUCTION_DOMAIN
                  optional: true
            - name: KODEX_AI_DOMAIN
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_AI_DOMAIN
                  optional: true
            - name: KODEX_CONTROL_PLANE_GRPC_ADDR
              value: ":9090"
            - name: KODEX_CONTROL_PLANE_HTTP_ADDR
              value: ":8081"
            - name: KODEX_DB_HOST
              value: postgres
            - name: KODEX_DB_PORT
              value: "5432"
            - name: KODEX_DB_NAME
              valueFrom:
                secretKeyRef:
                  name: kodex-postgres
                  key: KODEX_POSTGRES_DB
            - name: KODEX_DB_USER
              valueFrom:
                secretKeyRef:
                  name: kodex-postgres
                  key: KODEX_POSTGRES_USER
            - name: KODEX_DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: kodex-postgres
                  key: KODEX_POSTGRES_PASSWORD
            - name: KODEX_PROJECT_DB_ADMIN_HOST
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_PROJECT_DB_ADMIN_HOST
            - name: KODEX_PROJECT_DB_ADMIN_PORT
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_PROJECT_DB_ADMIN_PORT
            - name: KODEX_PROJECT_DB_ADMIN_USER
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_PROJECT_DB_ADMIN_USER
            - name: KODEX_PROJECT_DB_ADMIN_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_PROJECT_DB_ADMIN_PASSWORD
            - name: KODEX_PROJECT_DB_ADMIN_SSLMODE
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_PROJECT_DB_ADMIN_SSLMODE
                  optional: true
            - name: KODEX_PROJECT_DB_ADMIN_DATABASE
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_PROJECT_DB_ADMIN_DATABASE
                  optional: true
            - name: KODEX_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS
                  optional: true
            - name: KODEX_TOKEN_ENCRYPTION_KEY
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_TOKEN_ENCRYPTION_KEY
            - name: KODEX_GITHUB_PAT
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GITHUB_PAT
                  optional: true
            - name: KODEX_GITHUB_REPO
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GITHUB_REPO
            - name: KODEX_FIRST_PROJECT_GITHUB_REPO
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_FIRST_PROJECT_GITHUB_REPO
                  optional: true
            - name: KODEX_GIT_BOT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GIT_BOT_TOKEN
                  optional: true
            - name: KODEX_MCP_TOKEN_SIGNING_KEY
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_MCP_TOKEN_SIGNING_KEY
                  optional: true
            - name: KODEX_MCP_TOKEN_TTL
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_MCP_TOKEN_TTL
                  optional: true
            - name: KODEX_RUN_AGENT_LOGS_RETENTION_DAYS
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_RUN_AGENT_LOGS_RETENTION_DAYS
                  optional: true
            - name: KODEX_RUN_HEAVY_FIELDS_RETENTION_DAYS
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_RUN_HEAVY_FIELDS_RETENTION_DAYS
                  optional: true
            - name: KODEX_CONTROL_PLANE_MCP_BASE_URL
              value: '{{ envOr "KODEX_CONTROL_PLANE_MCP_BASE_URL" "http://kodex-control-plane:8081/mcp" }}'
            - name: KODEX_LEARNING_MODE_DEFAULT
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_LEARNING_MODE_DEFAULT
                  optional: true
            - name: KODEX_GITHUB_WEBHOOK_SECRET
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GITHUB_WEBHOOK_SECRET
            - name: KODEX_GITHUB_WEBHOOK_URL
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GITHUB_WEBHOOK_URL
                  optional: true
            - name: KODEX_GITHUB_WEBHOOK_EVENTS
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GITHUB_WEBHOOK_EVENTS
                  optional: true
            - name: KODEX_RUN_INTAKE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_INTAKE_LABEL
                  optional: true
            - name: KODEX_RUN_INTAKE_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_INTAKE_REVISE_LABEL
                  optional: true
            - name: KODEX_RUN_VISION_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_VISION_LABEL
                  optional: true
            - name: KODEX_RUN_VISION_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_VISION_REVISE_LABEL
                  optional: true
            - name: KODEX_RUN_PRD_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_PRD_LABEL
                  optional: true
            - name: KODEX_RUN_PRD_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_PRD_REVISE_LABEL
                  optional: true
            - name: KODEX_RUN_ARCH_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_ARCH_LABEL
                  optional: true
            - name: KODEX_RUN_ARCH_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_ARCH_REVISE_LABEL
                  optional: true
            - name: KODEX_RUN_DESIGN_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_DESIGN_LABEL
                  optional: true
            - name: KODEX_RUN_DESIGN_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_DESIGN_REVISE_LABEL
                  optional: true
            - name: KODEX_RUN_PLAN_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_PLAN_LABEL
                  optional: true
            - name: KODEX_RUN_PLAN_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_PLAN_REVISE_LABEL
                  optional: true
            - name: KODEX_RUN_DEV_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_DEV_LABEL
                  optional: true
            - name: KODEX_RUN_DEV_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_DEV_REVISE_LABEL
                  optional: true
            - name: KODEX_RUN_DOC_AUDIT_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_DOC_AUDIT_LABEL
                  optional: true
            - name: KODEX_RUN_AI_REPAIR_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_AI_REPAIR_LABEL
                  optional: true
            - name: KODEX_RUN_QA_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_QA_LABEL
                  optional: true
            - name: KODEX_RUN_RELEASE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_RELEASE_LABEL
                  optional: true
            - name: KODEX_RUN_POSTDEPLOY_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_POSTDEPLOY_LABEL
                  optional: true
            - name: KODEX_RUN_OPS_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_OPS_LABEL
                  optional: true
            - name: KODEX_RUN_SELF_IMPROVE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_SELF_IMPROVE_LABEL
                  optional: true
            - name: KODEX_RUN_RETHINK_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_RETHINK_LABEL
                  optional: true
            - name: KODEX_PUBLIC_BASE_URL
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_PUBLIC_BASE_URL
            - name: KODEX_INTERACTION_CALLBACK_BASE_URL
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_INTERACTION_CALLBACK_BASE_URL
                  optional: true
            - name: KODEX_GITHUB_OAUTH_CLIENT_ID
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GITHUB_OAUTH_CLIENT_ID
            - name: KODEX_GITHUB_OAUTH_CLIENT_SECRET
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GITHUB_OAUTH_CLIENT_SECRET
            - name: KODEX_BOOTSTRAP_OWNER_EMAIL
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_BOOTSTRAP_OWNER_EMAIL
            - name: KODEX_BOOTSTRAP_ALLOWED_EMAILS
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_BOOTSTRAP_ALLOWED_EMAILS
                  optional: true
            - name: KODEX_BOOTSTRAP_PLATFORM_ADMIN_EMAILS
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_BOOTSTRAP_PLATFORM_ADMIN_EMAILS
                  optional: true
          volumeMounts:
{{ if eq (envOr "KODEX_HOT_RELOAD" "") "true" }}
            - name: repo-cache
              mountPath: /workspace
{{ end }}
            - name: repo-cache
              mountPath: /repo-cache
{{ if ne (envOr "KODEX_HOT_RELOAD" "") "true" }}
              readOnly: true
{{ end }}
          readinessProbe:
            httpGet:
              path: /health/readyz
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          startupProbe:
            httpGet:
              path: /health/livez
              port: http
            periodSeconds: 5
            failureThreshold: 60
          livenessProbe:
            httpGet:
              path: /health/livez
              port: http
            initialDelaySeconds: 15
            periodSeconds: 20
      volumes:
        - name: repo-cache
          persistentVolumeClaim:
            claimName: kodex-repo-cache
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kodex-control-plane
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: control-plane
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kodex-control-plane-runtime-{{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: control-plane
rules:
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: [""]
    resources: ["secrets", "configmaps", "services", "endpoints", "persistentvolumeclaims", "pods", "pods/log", "events", "serviceaccounts"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["apps"]
    resources: ["daemonsets", "deployments", "replicasets", "statefulsets"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["batch"]
    resources: ["jobs", "cronjobs"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["ingresses", "networkpolicies"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles", "rolebindings"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles"]
    verbs: ["escalate", "bind"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["clusterroles", "clusterrolebindings"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["clusterroles"]
    verbs: ["escalate", "bind"]
  - apiGroups: ["cert-manager.io"]
    resources: ["clusterissuers"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kodex-control-plane-runtime-{{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: control-plane
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kodex-control-plane-runtime-{{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
subjects:
  - kind: ServiceAccount
    name: kodex-control-plane
    namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kodex-worker
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: worker
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kodex-worker-runtime-{{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: worker
rules:
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list", "watch", "create", "delete", "patch", "update"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles", "rolebindings"]
    verbs: ["get", "list", "watch", "create", "delete", "patch", "update"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles"]
    verbs: ["escalate", "bind"]
  - apiGroups: [""]
    resources: ["serviceaccounts", "resourcequotas", "limitranges", "pods", "pods/log", "events"]
    verbs: ["get", "list", "watch", "create", "delete", "patch", "update"]
  - apiGroups: [""]
    resources: ["configmaps", "endpoints", "services"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["daemonsets", "deployments", "replicasets", "statefulsets"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["pods/exec"]
    verbs: ["create"]
  - apiGroups: ["batch"]
    resources: ["jobs"]
    verbs: ["get", "list", "watch", "create", "delete", "patch", "update"]
  - apiGroups: ["batch"]
    resources: ["cronjobs"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kodex-worker-runtime-{{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: worker
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kodex-worker-runtime-{{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
subjects:
  - kind: ServiceAccount
    name: kodex-worker
    namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kodex-worker
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: worker
spec:
  replicas: {{ envOr "KODEX_WORKER_REPLICAS" "" }}
  selector:
    matchLabels:
      app.kubernetes.io/name: kodex
      app.kubernetes.io/component: worker
  template:
    metadata:
      labels:
        app.kubernetes.io/name: kodex
        app.kubernetes.io/component: worker
    spec:
      serviceAccountName: kodex-worker
      containers:
        - name: worker
          image: {{ envOr "KODEX_WORKER_IMAGE" "" }}
          imagePullPolicy: Always
          ports:
            - name: http
              containerPort: 8082
          env:
            - name: KODEX_WORKER_HTTP_ADDR
              value: '{{ envOr "KODEX_WORKER_HTTP_ADDR" ":8082" }}'
            - name: KODEX_ENV
              value: '{{ envOr "KODEX_ENV" "" }}'
            - name: KODEX_HOT_RELOAD
              value: '{{ envOr "KODEX_HOT_RELOAD" "" }}'
            - name: KODEX_DB_HOST
              value: postgres
            - name: KODEX_DB_PORT
              value: "5432"
            - name: KODEX_DB_NAME
              valueFrom:
                secretKeyRef:
                  name: kodex-postgres
                  key: KODEX_POSTGRES_DB
            - name: KODEX_DB_USER
              valueFrom:
                secretKeyRef:
                  name: kodex-postgres
                  key: KODEX_POSTGRES_USER
            - name: KODEX_DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: kodex-postgres
                  key: KODEX_POSTGRES_PASSWORD
            - name: KODEX_CONTROL_PLANE_GRPC_TARGET
              value: '{{ envOr "KODEX_CONTROL_PLANE_GRPC_TARGET" "kodex-control-plane:9090" }}'
            - name: KODEX_CONTROL_PLANE_MCP_BASE_URL
              value: '{{ envOr "KODEX_CONTROL_PLANE_MCP_BASE_URL" "http://kodex-control-plane:8081/mcp" }}'
            - name: KODEX_TELEGRAM_INTERACTION_ADAPTER_BASE_URL
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_TELEGRAM_INTERACTION_ADAPTER_BASE_URL
                  optional: true
            - name: KODEX_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN
                  optional: true
            - name: KODEX_TELEGRAM_INTERACTION_ADAPTER_TIMEOUT
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_TELEGRAM_INTERACTION_ADAPTER_TIMEOUT
                  optional: true
            - name: KODEX_OPENAI_API_KEY
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_OPENAI_API_KEY
                  optional: true
            - name: KODEX_CONTEXT7_API_KEY
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_CONTEXT7_API_KEY
                  optional: true
            - name: KODEX_GIT_BOT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GIT_BOT_TOKEN
                  optional: true
            - name: KODEX_GIT_BOT_USERNAME
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GIT_BOT_USERNAME
                  optional: true
            - name: KODEX_GIT_BOT_MAIL
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_GIT_BOT_MAIL
                  optional: true
            - name: KODEX_WORKER_POLL_INTERVAL
              value: '{{ envOr "KODEX_WORKER_POLL_INTERVAL" "" }}'
            - name: KODEX_WORKER_HEARTBEAT_INTERVAL
              value: '{{ envOr "KODEX_WORKER_HEARTBEAT_INTERVAL" "" }}'
            - name: KODEX_WORKER_INSTANCE_TTL
              value: '{{ envOr "KODEX_WORKER_INSTANCE_TTL" "" }}'
            - name: KODEX_WORKER_CLAIM_LIMIT
              value: '{{ envOr "KODEX_WORKER_CLAIM_LIMIT" "" }}'
            - name: KODEX_WORKER_RUNNING_CHECK_LIMIT
              value: '{{ envOr "KODEX_WORKER_RUNNING_CHECK_LIMIT" "" }}'
            - name: KODEX_WORKER_STALE_LEASE_SWEEP_LIMIT
              value: '{{ envOr "KODEX_WORKER_STALE_LEASE_SWEEP_LIMIT" "" }}'
            - name: KODEX_WORKER_SLOTS_PER_PROJECT
              value: '{{ envOr "KODEX_WORKER_SLOTS_PER_PROJECT" "" }}'
            - name: KODEX_WORKER_SLOT_LEASE_TTL
              value: '{{ envOr "KODEX_WORKER_SLOT_LEASE_TTL" "" }}'
            - name: KODEX_WORKER_RUN_LEASE_TTL
              value: '{{ envOr "KODEX_WORKER_RUN_LEASE_TTL" "" }}'
            - name: KODEX_WORKER_GITHUB_RATE_LIMIT_SWEEP_LIMIT
              value: '{{ envOr "KODEX_WORKER_GITHUB_RATE_LIMIT_SWEEP_LIMIT" "" }}'
            - name: KODEX_WORKER_MISSION_CONTROL_WARMUP_INTERVAL
              value: '{{ envOr "KODEX_WORKER_MISSION_CONTROL_WARMUP_INTERVAL" "" }}'
            - name: KODEX_WORKER_MISSION_CONTROL_WARMUP_PROJECT_LIMIT
              value: '{{ envOr "KODEX_WORKER_MISSION_CONTROL_WARMUP_PROJECT_LIMIT" "" }}'
            - name: KODEX_WORKER_MISSION_CONTROL_PENDING_COMMAND_LIMIT
              value: '{{ envOr "KODEX_WORKER_MISSION_CONTROL_PENDING_COMMAND_LIMIT" "" }}'
            - name: KODEX_WORKER_MISSION_CONTROL_CLAIM_TTL
              value: '{{ envOr "KODEX_WORKER_MISSION_CONTROL_CLAIM_TTL" "" }}'
            - name: KODEX_WORKER_MISSION_CONTROL_RETRY_MAX_ATTEMPTS
              value: '{{ envOr "KODEX_WORKER_MISSION_CONTROL_RETRY_MAX_ATTEMPTS" "" }}'
            - name: KODEX_WORKER_MISSION_CONTROL_RETRY_BASE_INTERVAL
              value: '{{ envOr "KODEX_WORKER_MISSION_CONTROL_RETRY_BASE_INTERVAL" "" }}'
            - name: KODEX_WORKER_K8S_NAMESPACE
              value: '{{ envOr "KODEX_WORKER_K8S_NAMESPACE" "" }}'
            - name: KODEX_WORKER_POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: KODEX_WORKER_POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: KODEX_PRODUCTION_NAMESPACE
              value: '{{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}'
            - name: KODEX_WORKER_JOB_IMAGE
              value: '{{ envOr "KODEX_WORKER_JOB_IMAGE" "" }}'
            - name: KODEX_WORKER_JOB_IMAGE_FALLBACK
              value: '{{ envOr "KODEX_WORKER_JOB_IMAGE_FALLBACK" "" }}'
            - name: KODEX_WORKER_JOB_IMAGE_CHECK_TIMEOUT
              value: '{{ envOr "KODEX_WORKER_JOB_IMAGE_CHECK_TIMEOUT" "" }}'
            - name: KODEX_WORKER_JOB_COMMAND
              value: '{{ envOr "KODEX_WORKER_JOB_COMMAND" "" }}'
            - name: KODEX_WORKER_JOB_TTL_SECONDS
              value: '{{ envOr "KODEX_WORKER_JOB_TTL_SECONDS" "" }}'
            - name: KODEX_WORKER_JOB_BACKOFF_LIMIT
              value: '{{ envOr "KODEX_WORKER_JOB_BACKOFF_LIMIT" "" }}'
            - name: KODEX_WORKER_JOB_ACTIVE_DEADLINE_SECONDS
              value: '{{ envOr "KODEX_WORKER_JOB_ACTIVE_DEADLINE_SECONDS" "" }}'
            - name: KODEX_WORKER_RUN_NAMESPACE_PREFIX
              value: '{{ envOr "KODEX_WORKER_RUN_NAMESPACE_PREFIX" "" }}'
            - name: KODEX_WORKER_RUN_NAMESPACE_CLEANUP
              value: '{{ envOr "KODEX_WORKER_RUN_NAMESPACE_CLEANUP" "" }}'
            - name: KODEX_WORKER_NAMESPACE_LEASE_SWEEP_LIMIT
              value: '{{ envOr "KODEX_WORKER_NAMESPACE_LEASE_SWEEP_LIMIT" "" }}'
            - name: KODEX_INTERNAL_REGISTRY_HOST
              value: '{{ envOr "KODEX_INTERNAL_REGISTRY_HOST" "" }}'
            - name: KODEX_INTERNAL_REGISTRY_SCHEME
              value: '{{ envOr "KODEX_INTERNAL_REGISTRY_SCHEME" "" }}'
            - name: KODEX_RUN_DEBUG_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_RUN_DEBUG_LABEL
                  optional: true
            - name: KODEX_STATE_IN_REVIEW_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_STATE_IN_REVIEW_LABEL
                  optional: true
            - name: KODEX_AI_MODEL_GPT_5_3_CODEX_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_AI_MODEL_GPT_5_3_CODEX_LABEL
                  optional: true
            - name: KODEX_AI_MODEL_GPT_5_3_CODEX_SPARK_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_AI_MODEL_GPT_5_3_CODEX_SPARK_LABEL
                  optional: true
            - name: KODEX_AI_MODEL_GPT_5_2_CODEX_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_AI_MODEL_GPT_5_2_CODEX_LABEL
                  optional: true
            - name: KODEX_AI_MODEL_GPT_5_1_CODEX_MAX_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_AI_MODEL_GPT_5_1_CODEX_MAX_LABEL
                  optional: true
            - name: KODEX_AI_MODEL_GPT_5_2_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_AI_MODEL_GPT_5_2_LABEL
                  optional: true
            - name: KODEX_AI_MODEL_GPT_5_1_CODEX_MINI_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_AI_MODEL_GPT_5_1_CODEX_MINI_LABEL
                  optional: true
            - name: KODEX_AI_REASONING_LOW_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_AI_REASONING_LOW_LABEL
                  optional: true
            - name: KODEX_AI_REASONING_MEDIUM_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_AI_REASONING_MEDIUM_LABEL
                  optional: true
            - name: KODEX_AI_REASONING_HIGH_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_AI_REASONING_HIGH_LABEL
                  optional: true
            - name: KODEX_AI_REASONING_EXTRA_HIGH_LABEL
              valueFrom:
                configMapKeyRef:
                  name: kodex-label-catalog
                  key: KODEX_AI_REASONING_EXTRA_HIGH_LABEL
                  optional: true
            - name: KODEX_WORKER_RUN_SERVICE_ACCOUNT
              value: '{{ envOr "KODEX_WORKER_RUN_SERVICE_ACCOUNT" "" }}'
            - name: KODEX_WORKER_RUN_ROLE_NAME
              value: '{{ envOr "KODEX_WORKER_RUN_ROLE_NAME" "" }}'
            - name: KODEX_WORKER_RUN_ROLE_BINDING_NAME
              value: '{{ envOr "KODEX_WORKER_RUN_ROLE_BINDING_NAME" "" }}'
            - name: KODEX_WORKER_RUN_RESOURCE_QUOTA_NAME
              value: '{{ envOr "KODEX_WORKER_RUN_RESOURCE_QUOTA_NAME" "" }}'
            - name: KODEX_WORKER_RUN_LIMIT_RANGE_NAME
              value: '{{ envOr "KODEX_WORKER_RUN_LIMIT_RANGE_NAME" "" }}'
            - name: KODEX_WORKER_RUN_CREDENTIALS_SECRET_NAME
              value: '{{ envOr "KODEX_WORKER_RUN_CREDENTIALS_SECRET_NAME" "" }}'
            - name: KODEX_WORKER_RUN_QUOTA_PODS
              value: '{{ envOr "KODEX_WORKER_RUN_QUOTA_PODS" "" }}'
            - name: KODEX_WORKER_AI_REPAIR_NAMESPACE
              value: '{{ envOr "KODEX_WORKER_AI_REPAIR_NAMESPACE" "" }}'
            - name: KODEX_WORKER_AI_REPAIR_SERVICE_ACCOUNT
              value: '{{ envOr "KODEX_WORKER_AI_REPAIR_SERVICE_ACCOUNT" "" }}'
            - name: KODEX_AGENT_DEFAULT_MODEL
              value: '{{ envOr "KODEX_AGENT_DEFAULT_MODEL" "" }}'
            - name: KODEX_AGENT_DEFAULT_REASONING_EFFORT
              value: '{{ envOr "KODEX_AGENT_DEFAULT_REASONING_EFFORT" "" }}'
            - name: KODEX_AGENT_DEFAULT_LOCALE
              value: '{{ envOr "KODEX_AGENT_DEFAULT_LOCALE" "" }}'
            - name: KODEX_AGENT_BASE_BRANCH
              value: '{{ envOr "KODEX_AGENT_BASE_BRANCH" "" }}'
          readinessProbe:
            httpGet:
              path: /health/readyz
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          startupProbe:
            httpGet:
              path: /health/livez
              port: http
            periodSeconds: 5
            failureThreshold: 60
          livenessProbe:
            httpGet:
              path: /health/livez
              port: http
            initialDelaySeconds: 10
            periodSeconds: 20
{{ if eq (envOr "KODEX_HOT_RELOAD" "") "true" }}
          volumeMounts:
            - name: repo-cache
              mountPath: /workspace
{{ end }}
{{ if eq (envOr "KODEX_HOT_RELOAD" "") "true" }}
      volumes:
        - name: repo-cache
          persistentVolumeClaim:
            claimName: kodex-repo-cache
{{ end }}
{{ if eq (envOr "KODEX_ENV" "production") "production" }}
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: kodex-worker-namespace-cleanup
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: worker-namespace-cleanup
spec:
  schedule: '{{ envOr "KODEX_WORKER_NAMESPACE_CLEANUP_CRON_SCHEDULE" "*/15 * * * *" }}'
  timeZone: Etc/UTC
  concurrencyPolicy: Forbid
  suspend: {{ eq (envOr "KODEX_WORKER_NAMESPACE_CLEANUP_CRON_SUSPEND" "false") "true" }}
  startingDeadlineSeconds: {{ envOr "KODEX_WORKER_NAMESPACE_CLEANUP_CRON_STARTING_DEADLINE_SECONDS" "300" }}
  successfulJobsHistoryLimit: {{ envOr "KODEX_WORKER_NAMESPACE_CLEANUP_CRON_SUCCESS_HISTORY_LIMIT" "1" }}
  failedJobsHistoryLimit: {{ envOr "KODEX_WORKER_NAMESPACE_CLEANUP_CRON_FAILED_HISTORY_LIMIT" "3" }}
  jobTemplate:
    spec:
      backoffLimit: 0
      ttlSecondsAfterFinished: 86400
      activeDeadlineSeconds: {{ envOr "KODEX_WORKER_NAMESPACE_CLEANUP_CRON_ACTIVE_DEADLINE_SECONDS" "900" }}
      template:
        metadata:
          labels:
            app.kubernetes.io/name: kodex
            app.kubernetes.io/component: worker-namespace-cleanup
        spec:
          serviceAccountName: kodex-worker
          restartPolicy: Never
          containers:
            - name: namespace-cleanup
              image: {{ envOr "KODEX_WORKER_IMAGE" "" }}
              imagePullPolicy: Always
              env:
                - name: KODEX_WORKER_MODE
                  value: namespace_cleanup_once
                - name: KODEX_WORKER_ID
                  value: kodex-worker-namespace-cleanup
                - name: KODEX_DB_HOST
                  value: postgres
                - name: KODEX_DB_PORT
                  value: "5432"
                - name: KODEX_DB_NAME
                  valueFrom:
                    secretKeyRef:
                      name: kodex-postgres
                      key: KODEX_POSTGRES_DB
                - name: KODEX_DB_USER
                  valueFrom:
                    secretKeyRef:
                      name: kodex-postgres
                      key: KODEX_POSTGRES_USER
                - name: KODEX_DB_PASSWORD
                  valueFrom:
                    secretKeyRef:
                      name: kodex-postgres
                      key: KODEX_POSTGRES_PASSWORD
                - name: KODEX_CONTROL_PLANE_GRPC_TARGET
                  value: '{{ envOr "KODEX_CONTROL_PLANE_GRPC_TARGET" "kodex-control-plane:9090" }}'
                - name: KODEX_WORKER_K8S_NAMESPACE
                  value: '{{ envOr "KODEX_WORKER_K8S_NAMESPACE" "" }}'
                - name: KODEX_PRODUCTION_NAMESPACE
                  value: '{{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}'
                - name: KODEX_WORKER_RUN_NAMESPACE_PREFIX
                  value: '{{ envOr "KODEX_WORKER_RUN_NAMESPACE_PREFIX" "" }}'
                - name: KODEX_WORKER_RUN_NAMESPACE_CLEANUP
                  value: '{{ envOr "KODEX_WORKER_RUN_NAMESPACE_CLEANUP" "" }}'
                - name: KODEX_WORKER_NAMESPACE_LEASE_SWEEP_LIMIT
                  value: '{{ envOr "KODEX_WORKER_NAMESPACE_LEASE_SWEEP_LIMIT" "" }}'
{{ end }}
---
apiVersion: v1
kind: Service
metadata:
  name: kodex-worker
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: worker
spec:
  selector:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: worker
  ports:
    - name: http
      port: 8082
      targetPort: http
