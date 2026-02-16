apiVersion: v1
kind: Service
metadata:
  name: codex-k8s
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
spec:
  selector:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: api-gateway
  ports:
    - name: http
      port: 80
      targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: codex-k8s
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
spec:
  replicas: {{ envOr "CODEXK8S_PLATFORM_DEPLOYMENT_REPLICAS" "" }}
  selector:
    matchLabels:
      app.kubernetes.io/name: codex-k8s
      app.kubernetes.io/component: api-gateway
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s
        app.kubernetes.io/component: api-gateway
    spec:
      initContainers:
        - name: wait-control-plane
          image: {{ envOr "CODEXK8S_BUSYBOX_IMAGE" "busybox:1.36" }}
          imagePullPolicy: IfNotPresent
          command:
            - sh
            - -ec
            - |
              until wget -q -O /dev/null http://codex-k8s-control-plane:8081/health/readyz; do
                echo "waiting for control-plane readiness..."
                sleep 2
              done
      containers:
        - name: codex-k8s
          image: {{ envOr "CODEXK8S_API_GATEWAY_IMAGE" "" }}
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
              name: http
          env:
            - name: CODEXK8S_ENV
              value: '{{ envOr "CODEXK8S_ENV" "" }}'
            - name: CODEXK8S_HOT_RELOAD
              value: '{{ envOr "CODEXK8S_HOT_RELOAD" "" }}'
            - name: CODEXK8S_HTTP_ADDR
              value: ":8080"
            - name: CODEXK8S_CONTROL_PLANE_GRPC_TARGET
              value: '{{ envOr "CODEXK8S_CONTROL_PLANE_GRPC_TARGET" "" }}'
            - name: CODEXK8S_VITE_DEV_UPSTREAM
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_VITE_DEV_UPSTREAM
                  optional: true
            - name: CODEXK8S_GITHUB_WEBHOOK_SECRET
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GITHUB_WEBHOOK_SECRET
            - name: CODEXK8S_PUBLIC_BASE_URL
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_PUBLIC_BASE_URL
            - name: CODEXK8S_COOKIE_SECURE
              value: "true"
            - name: CODEXK8S_GITHUB_OAUTH_CLIENT_ID
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GITHUB_OAUTH_CLIENT_ID
            - name: CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET
            - name: CODEXK8S_JWT_SIGNING_KEY
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_JWT_SIGNING_KEY
            - name: CODEXK8S_JWT_TTL
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_JWT_TTL
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
---
apiVersion: v1
kind: Service
metadata:
  name: codex-k8s-control-plane
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
spec:
  selector:
    app.kubernetes.io/name: codex-k8s
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
  name: codex-k8s-control-plane
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
spec:
  replicas: {{ envOr "CODEXK8S_PLATFORM_DEPLOYMENT_REPLICAS" "" }}
  selector:
    matchLabels:
      app.kubernetes.io/name: codex-k8s
      app.kubernetes.io/component: control-plane
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s
        app.kubernetes.io/component: control-plane
    spec:
      serviceAccountName: codex-k8s-control-plane
      initContainers:
        - name: wait-postgres
          image: {{ envOr "CODEXK8S_BUSYBOX_IMAGE" "busybox:1.36" }}
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
          image: {{ envOr "CODEXK8S_CONTROL_PLANE_IMAGE" "" }}
          imagePullPolicy: Always
          ports:
            - containerPort: 9090
              name: grpc
            - containerPort: 8081
              name: http
          env:
            - name: CODEXK8S_PLATFORM_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: CODEXK8S_ENV
              value: '{{ envOr "CODEXK8S_ENV" "" }}'
            - name: CODEXK8S_HOT_RELOAD
              value: '{{ envOr "CODEXK8S_HOT_RELOAD" "" }}'
            - name: CODEXK8S_SERVICES_CONFIG_PATH
              value: /app/services.yaml
            - name: CODEXK8S_SERVICES_CONFIG_ENV
              value: '{{ envOr "CODEXK8S_SERVICES_CONFIG_ENV" "" }}'
            - name: CODEXK8S_REPOSITORY_ROOT
              value: /repo-cache
            - name: CODEXK8S_PRODUCTION_DOMAIN
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_PRODUCTION_DOMAIN
                  optional: true
            - name: CODEXK8S_AI_DOMAIN
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_AI_DOMAIN
                  optional: true
            - name: CODEXK8S_CONTROL_PLANE_GRPC_ADDR
              value: ":9090"
            - name: CODEXK8S_CONTROL_PLANE_HTTP_ADDR
              value: ":8081"
            - name: CODEXK8S_DB_HOST
              value: postgres
            - name: CODEXK8S_DB_PORT
              value: "5432"
            - name: CODEXK8S_DB_NAME
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-postgres
                  key: CODEXK8S_POSTGRES_DB
            - name: CODEXK8S_DB_USER
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-postgres
                  key: CODEXK8S_POSTGRES_USER
            - name: CODEXK8S_DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-postgres
                  key: CODEXK8S_POSTGRES_PASSWORD
            - name: CODEXK8S_PROJECT_DB_ADMIN_HOST
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_PROJECT_DB_ADMIN_HOST
            - name: CODEXK8S_PROJECT_DB_ADMIN_PORT
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_PROJECT_DB_ADMIN_PORT
            - name: CODEXK8S_PROJECT_DB_ADMIN_USER
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_PROJECT_DB_ADMIN_USER
            - name: CODEXK8S_PROJECT_DB_ADMIN_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_PROJECT_DB_ADMIN_PASSWORD
            - name: CODEXK8S_PROJECT_DB_ADMIN_SSLMODE
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_PROJECT_DB_ADMIN_SSLMODE
                  optional: true
            - name: CODEXK8S_PROJECT_DB_ADMIN_DATABASE
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_PROJECT_DB_ADMIN_DATABASE
                  optional: true
            - name: CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_PROJECT_DB_LIFECYCLE_ALLOWED_ENVS
                  optional: true
            - name: CODEXK8S_TOKEN_ENCRYPTION_KEY
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_TOKEN_ENCRYPTION_KEY
            - name: CODEXK8S_GITHUB_PAT
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GITHUB_PAT
                  optional: true
            - name: CODEXK8S_GITHUB_REPO
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GITHUB_REPO
            - name: CODEXK8S_FIRST_PROJECT_GITHUB_REPO
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_FIRST_PROJECT_GITHUB_REPO
                  optional: true
            - name: CODEXK8S_GIT_BOT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GIT_BOT_TOKEN
                  optional: true
            - name: CODEXK8S_MCP_TOKEN_SIGNING_KEY
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_MCP_TOKEN_SIGNING_KEY
                  optional: true
            - name: CODEXK8S_MCP_TOKEN_TTL
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_MCP_TOKEN_TTL
                  optional: true
            - name: CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_RUN_AGENT_LOGS_RETENTION_DAYS
                  optional: true
            - name: CODEXK8S_CONTROL_PLANE_MCP_BASE_URL
              value: '{{ envOr "CODEXK8S_CONTROL_PLANE_MCP_BASE_URL" "" }}'
            - name: CODEXK8S_LEARNING_MODE_DEFAULT
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_LEARNING_MODE_DEFAULT
                  optional: true
            - name: CODEXK8S_GITHUB_WEBHOOK_SECRET
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GITHUB_WEBHOOK_SECRET
            - name: CODEXK8S_GITHUB_WEBHOOK_URL
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GITHUB_WEBHOOK_URL
                  optional: true
            - name: CODEXK8S_GITHUB_WEBHOOK_EVENTS
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GITHUB_WEBHOOK_EVENTS
                  optional: true
            - name: CODEXK8S_RUN_INTAKE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_INTAKE_LABEL
                  optional: true
            - name: CODEXK8S_RUN_INTAKE_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_INTAKE_REVISE_LABEL
                  optional: true
            - name: CODEXK8S_RUN_VISION_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_VISION_LABEL
                  optional: true
            - name: CODEXK8S_RUN_VISION_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_VISION_REVISE_LABEL
                  optional: true
            - name: CODEXK8S_RUN_PRD_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_PRD_LABEL
                  optional: true
            - name: CODEXK8S_RUN_PRD_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_PRD_REVISE_LABEL
                  optional: true
            - name: CODEXK8S_RUN_ARCH_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_ARCH_LABEL
                  optional: true
            - name: CODEXK8S_RUN_ARCH_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_ARCH_REVISE_LABEL
                  optional: true
            - name: CODEXK8S_RUN_DESIGN_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_DESIGN_LABEL
                  optional: true
            - name: CODEXK8S_RUN_DESIGN_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_DESIGN_REVISE_LABEL
                  optional: true
            - name: CODEXK8S_RUN_PLAN_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_PLAN_LABEL
                  optional: true
            - name: CODEXK8S_RUN_PLAN_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_PLAN_REVISE_LABEL
                  optional: true
            - name: CODEXK8S_RUN_DEV_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_DEV_LABEL
                  optional: true
            - name: CODEXK8S_RUN_DEV_REVISE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_DEV_REVISE_LABEL
                  optional: true
            - name: CODEXK8S_RUN_DOC_AUDIT_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_DOC_AUDIT_LABEL
                  optional: true
            - name: CODEXK8S_RUN_QA_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_QA_LABEL
                  optional: true
            - name: CODEXK8S_RUN_RELEASE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_RELEASE_LABEL
                  optional: true
            - name: CODEXK8S_RUN_POSTDEPLOY_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_POSTDEPLOY_LABEL
                  optional: true
            - name: CODEXK8S_RUN_OPS_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_OPS_LABEL
                  optional: true
            - name: CODEXK8S_RUN_SELF_IMPROVE_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_SELF_IMPROVE_LABEL
                  optional: true
            - name: CODEXK8S_RUN_RETHINK_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_RETHINK_LABEL
                  optional: true
            - name: CODEXK8S_PUBLIC_BASE_URL
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_PUBLIC_BASE_URL
            - name: CODEXK8S_GITHUB_OAUTH_CLIENT_ID
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GITHUB_OAUTH_CLIENT_ID
            - name: CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GITHUB_OAUTH_CLIENT_SECRET
            - name: CODEXK8S_BOOTSTRAP_OWNER_EMAIL
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_BOOTSTRAP_OWNER_EMAIL
            - name: CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_BOOTSTRAP_ALLOWED_EMAILS
                  optional: true
            - name: CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_BOOTSTRAP_PLATFORM_ADMIN_EMAILS
                  optional: true
          volumeMounts:
            - name: repo-cache
              mountPath: /repo-cache
              readOnly: true
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
            claimName: codex-k8s-repo-cache
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: codex-k8s-control-plane
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: control-plane
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: codex-k8s-control-plane-runtime-{{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
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
    resources: ["jobs"]
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
  name: codex-k8s-control-plane-runtime-{{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: control-plane
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: codex-k8s-control-plane-runtime-{{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
subjects:
  - kind: ServiceAccount
    name: codex-k8s-control-plane
    namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: codex-k8s-worker
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: worker
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: codex-k8s-worker-runtime-{{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
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
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: codex-k8s-worker-runtime-{{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: worker
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: codex-k8s-worker-runtime-{{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
subjects:
  - kind: ServiceAccount
    name: codex-k8s-worker
    namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: codex-k8s-worker
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: worker
spec:
  replicas: {{ envOr "CODEXK8S_WORKER_REPLICAS" "" }}
  selector:
    matchLabels:
      app.kubernetes.io/name: codex-k8s
      app.kubernetes.io/component: worker
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s
        app.kubernetes.io/component: worker
    spec:
      serviceAccountName: codex-k8s-worker
      containers:
        - name: worker
          image: {{ envOr "CODEXK8S_WORKER_IMAGE" "" }}
          imagePullPolicy: Always
          env:
            - name: CODEXK8S_ENV
              value: '{{ envOr "CODEXK8S_ENV" "" }}'
            - name: CODEXK8S_HOT_RELOAD
              value: '{{ envOr "CODEXK8S_HOT_RELOAD" "" }}'
            - name: CODEXK8S_DB_HOST
              value: postgres
            - name: CODEXK8S_DB_PORT
              value: "5432"
            - name: CODEXK8S_DB_NAME
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-postgres
                  key: CODEXK8S_POSTGRES_DB
            - name: CODEXK8S_DB_USER
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-postgres
                  key: CODEXK8S_POSTGRES_USER
            - name: CODEXK8S_DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-postgres
                  key: CODEXK8S_POSTGRES_PASSWORD
            - name: CODEXK8S_CONTROL_PLANE_GRPC_TARGET
              value: '{{ envOr "CODEXK8S_CONTROL_PLANE_GRPC_TARGET" "" }}'
            - name: CODEXK8S_CONTROL_PLANE_MCP_BASE_URL
              value: '{{ envOr "CODEXK8S_CONTROL_PLANE_MCP_BASE_URL" "" }}'
            - name: CODEXK8S_OPENAI_API_KEY
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_OPENAI_API_KEY
                  optional: true
            - name: CODEXK8S_OPENAI_AUTH_FILE
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_OPENAI_AUTH_FILE
                  optional: true
            - name: CODEXK8S_CONTEXT7_API_KEY
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_CONTEXT7_API_KEY
                  optional: true
            - name: CODEXK8S_GIT_BOT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GIT_BOT_TOKEN
                  optional: true
            - name: CODEXK8S_GIT_BOT_USERNAME
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GIT_BOT_USERNAME
                  optional: true
            - name: CODEXK8S_GIT_BOT_MAIL
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GIT_BOT_MAIL
                  optional: true
            - name: CODEXK8S_WORKER_POLL_INTERVAL
              value: '{{ envOr "CODEXK8S_WORKER_POLL_INTERVAL" "" }}'
            - name: CODEXK8S_WORKER_CLAIM_LIMIT
              value: '{{ envOr "CODEXK8S_WORKER_CLAIM_LIMIT" "" }}'
            - name: CODEXK8S_WORKER_RUNNING_CHECK_LIMIT
              value: '{{ envOr "CODEXK8S_WORKER_RUNNING_CHECK_LIMIT" "" }}'
            - name: CODEXK8S_WORKER_SLOTS_PER_PROJECT
              value: '{{ envOr "CODEXK8S_WORKER_SLOTS_PER_PROJECT" "" }}'
            - name: CODEXK8S_WORKER_SLOT_LEASE_TTL
              value: '{{ envOr "CODEXK8S_WORKER_SLOT_LEASE_TTL" "" }}'
            - name: CODEXK8S_WORKER_K8S_NAMESPACE
              value: '{{ envOr "CODEXK8S_WORKER_K8S_NAMESPACE" "" }}'
            - name: CODEXK8S_WORKER_JOB_IMAGE
              value: '{{ envOr "CODEXK8S_WORKER_JOB_IMAGE" "" }}'
            - name: CODEXK8S_WORKER_JOB_COMMAND
              value: '{{ envOr "CODEXK8S_WORKER_JOB_COMMAND" "" }}'
            - name: CODEXK8S_WORKER_JOB_TTL_SECONDS
              value: '{{ envOr "CODEXK8S_WORKER_JOB_TTL_SECONDS" "" }}'
            - name: CODEXK8S_WORKER_JOB_BACKOFF_LIMIT
              value: '{{ envOr "CODEXK8S_WORKER_JOB_BACKOFF_LIMIT" "" }}'
            - name: CODEXK8S_WORKER_JOB_ACTIVE_DEADLINE_SECONDS
              value: '{{ envOr "CODEXK8S_WORKER_JOB_ACTIVE_DEADLINE_SECONDS" "" }}'
            - name: CODEXK8S_WORKER_RUN_NAMESPACE_PREFIX
              value: '{{ envOr "CODEXK8S_WORKER_RUN_NAMESPACE_PREFIX" "" }}'
            - name: CODEXK8S_WORKER_RUN_NAMESPACE_CLEANUP
              value: '{{ envOr "CODEXK8S_WORKER_RUN_NAMESPACE_CLEANUP" "" }}'
            - name: CODEXK8S_RUN_DEBUG_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_RUN_DEBUG_LABEL
                  optional: true
            - name: CODEXK8S_STATE_IN_REVIEW_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_STATE_IN_REVIEW_LABEL
                  optional: true
            - name: CODEXK8S_AI_MODEL_GPT_5_3_CODEX_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_AI_MODEL_GPT_5_3_CODEX_LABEL
                  optional: true
            - name: CODEXK8S_AI_MODEL_GPT_5_3_CODEX_SPARK_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_AI_MODEL_GPT_5_3_CODEX_SPARK_LABEL
                  optional: true
            - name: CODEXK8S_AI_MODEL_GPT_5_2_CODEX_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_AI_MODEL_GPT_5_2_CODEX_LABEL
                  optional: true
            - name: CODEXK8S_AI_MODEL_GPT_5_1_CODEX_MAX_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_AI_MODEL_GPT_5_1_CODEX_MAX_LABEL
                  optional: true
            - name: CODEXK8S_AI_MODEL_GPT_5_2_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_AI_MODEL_GPT_5_2_LABEL
                  optional: true
            - name: CODEXK8S_AI_MODEL_GPT_5_1_CODEX_MINI_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_AI_MODEL_GPT_5_1_CODEX_MINI_LABEL
                  optional: true
            - name: CODEXK8S_AI_REASONING_LOW_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_AI_REASONING_LOW_LABEL
                  optional: true
            - name: CODEXK8S_AI_REASONING_MEDIUM_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_AI_REASONING_MEDIUM_LABEL
                  optional: true
            - name: CODEXK8S_AI_REASONING_HIGH_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_AI_REASONING_HIGH_LABEL
                  optional: true
            - name: CODEXK8S_AI_REASONING_EXTRA_HIGH_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: CODEXK8S_AI_REASONING_EXTRA_HIGH_LABEL
                  optional: true
            - name: CODEXK8S_WORKER_RUN_SERVICE_ACCOUNT
              value: '{{ envOr "CODEXK8S_WORKER_RUN_SERVICE_ACCOUNT" "" }}'
            - name: CODEXK8S_WORKER_RUN_ROLE_NAME
              value: '{{ envOr "CODEXK8S_WORKER_RUN_ROLE_NAME" "" }}'
            - name: CODEXK8S_WORKER_RUN_ROLE_BINDING_NAME
              value: '{{ envOr "CODEXK8S_WORKER_RUN_ROLE_BINDING_NAME" "" }}'
            - name: CODEXK8S_WORKER_RUN_RESOURCE_QUOTA_NAME
              value: '{{ envOr "CODEXK8S_WORKER_RUN_RESOURCE_QUOTA_NAME" "" }}'
            - name: CODEXK8S_WORKER_RUN_LIMIT_RANGE_NAME
              value: '{{ envOr "CODEXK8S_WORKER_RUN_LIMIT_RANGE_NAME" "" }}'
            - name: CODEXK8S_WORKER_RUN_CREDENTIALS_SECRET_NAME
              value: '{{ envOr "CODEXK8S_WORKER_RUN_CREDENTIALS_SECRET_NAME" "" }}'
            - name: CODEXK8S_WORKER_RUN_QUOTA_PODS
              value: '{{ envOr "CODEXK8S_WORKER_RUN_QUOTA_PODS" "" }}'
            - name: CODEXK8S_AGENT_DEFAULT_MODEL
              value: '{{ envOr "CODEXK8S_AGENT_DEFAULT_MODEL" "" }}'
            - name: CODEXK8S_AGENT_DEFAULT_REASONING_EFFORT
              value: '{{ envOr "CODEXK8S_AGENT_DEFAULT_REASONING_EFFORT" "" }}'
            - name: CODEXK8S_AGENT_DEFAULT_LOCALE
              value: '{{ envOr "CODEXK8S_AGENT_DEFAULT_LOCALE" "" }}'
            - name: CODEXK8S_AGENT_BASE_BRANCH
              value: '{{ envOr "CODEXK8S_AGENT_BASE_BRANCH" "" }}'
