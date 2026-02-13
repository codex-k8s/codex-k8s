apiVersion: v1
kind: Service
metadata:
  name: codex-k8s
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
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
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s
spec:
  replicas: 1
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
          image: busybox:1.36
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
          image: ${CODEXK8S_API_GATEWAY_IMAGE}
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
              name: http
          env:
            - name: CODEXK8S_ENV
              value: ai-staging
            - name: CODEXK8S_HTTP_ADDR
              value: ":8080"
            - name: CODEXK8S_CONTROL_PLANE_GRPC_TARGET
              value: "${CODEXK8S_CONTROL_PLANE_GRPC_TARGET}"
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
          livenessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 15
            periodSeconds: 20
          resources:
            requests:
              cpu: ${CODEXK8S_API_GATEWAY_RESOURCES_REQUEST_CPU}
              memory: ${CODEXK8S_API_GATEWAY_RESOURCES_REQUEST_MEMORY}
            limits:
              cpu: ${CODEXK8S_API_GATEWAY_RESOURCES_LIMIT_CPU}
              memory: ${CODEXK8S_API_GATEWAY_RESOURCES_LIMIT_MEMORY}
---
apiVersion: v1
kind: Service
metadata:
  name: codex-k8s-control-plane
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
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
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s
spec:
  replicas: 1
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
      initContainers:
        - name: wait-postgres
          image: busybox:1.36
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
          image: ${CODEXK8S_CONTROL_PLANE_IMAGE}
          imagePullPolicy: Always
          command: ["/usr/local/bin/codex-k8s-control-plane"]
          ports:
            - containerPort: 9090
              name: grpc
            - containerPort: 8081
              name: http
          env:
            - name: CODEXK8S_ENV
              value: ai-staging
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
              value: "${CODEXK8S_CONTROL_PLANE_MCP_BASE_URL}"
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
            - name: CODEXK8S_PUBLIC_BASE_URL
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_PUBLIC_BASE_URL
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
          readinessProbe:
            httpGet:
              path: /health/readyz
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /health/livez
              port: http
            initialDelaySeconds: 15
            periodSeconds: 20
          resources:
            requests:
              cpu: ${CODEXK8S_CONTROL_PLANE_RESOURCES_REQUEST_CPU}
              memory: ${CODEXK8S_CONTROL_PLANE_RESOURCES_REQUEST_MEMORY}
            limits:
              cpu: ${CODEXK8S_CONTROL_PLANE_RESOURCES_LIMIT_CPU}
              memory: ${CODEXK8S_CONTROL_PLANE_RESOURCES_LIMIT_MEMORY}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: codex-k8s-worker
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: worker
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: codex-k8s-worker-runtime
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
  name: codex-k8s-worker-runtime
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: worker
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: codex-k8s-worker-runtime
subjects:
  - kind: ServiceAccount
    name: codex-k8s-worker
    namespace: ${CODEXK8S_STAGING_NAMESPACE}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: codex-k8s-worker
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: worker
spec:
  replicas: ${CODEXK8S_WORKER_REPLICAS}
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
          image: ${CODEXK8S_WORKER_IMAGE}
          imagePullPolicy: Always
          command: ["/usr/local/bin/codex-k8s-worker"]
          env:
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
              value: "${CODEXK8S_CONTROL_PLANE_GRPC_TARGET}"
            - name: CODEXK8S_CONTROL_PLANE_MCP_BASE_URL
              value: "${CODEXK8S_CONTROL_PLANE_MCP_BASE_URL}"
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
              value: "${CODEXK8S_WORKER_POLL_INTERVAL}"
            - name: CODEXK8S_WORKER_CLAIM_LIMIT
              value: "${CODEXK8S_WORKER_CLAIM_LIMIT}"
            - name: CODEXK8S_WORKER_RUNNING_CHECK_LIMIT
              value: "${CODEXK8S_WORKER_RUNNING_CHECK_LIMIT}"
            - name: CODEXK8S_WORKER_SLOTS_PER_PROJECT
              value: "${CODEXK8S_WORKER_SLOTS_PER_PROJECT}"
            - name: CODEXK8S_WORKER_SLOT_LEASE_TTL
              value: "${CODEXK8S_WORKER_SLOT_LEASE_TTL}"
            - name: CODEXK8S_WORKER_K8S_NAMESPACE
              value: "${CODEXK8S_WORKER_K8S_NAMESPACE}"
            - name: CODEXK8S_WORKER_JOB_IMAGE
              value: "${CODEXK8S_WORKER_JOB_IMAGE}"
            - name: CODEXK8S_WORKER_JOB_COMMAND
              value: "${CODEXK8S_WORKER_JOB_COMMAND}"
            - name: CODEXK8S_WORKER_JOB_TTL_SECONDS
              value: "${CODEXK8S_WORKER_JOB_TTL_SECONDS}"
            - name: CODEXK8S_WORKER_JOB_BACKOFF_LIMIT
              value: "${CODEXK8S_WORKER_JOB_BACKOFF_LIMIT}"
            - name: CODEXK8S_WORKER_JOB_ACTIVE_DEADLINE_SECONDS
              value: "${CODEXK8S_WORKER_JOB_ACTIVE_DEADLINE_SECONDS}"
            - name: CODEXK8S_WORKER_RUN_NAMESPACE_PREFIX
              value: "${CODEXK8S_WORKER_RUN_NAMESPACE_PREFIX}"
            - name: CODEXK8S_WORKER_RUN_NAMESPACE_CLEANUP
              value: "${CODEXK8S_WORKER_RUN_NAMESPACE_CLEANUP}"
            - name: CODEXK8S_RUN_DEBUG_LABEL
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_RUN_DEBUG_LABEL
                  optional: true
            - name: CODEXK8S_STATE_IN_REVIEW_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: STATE_IN_REVIEW_LABEL
                  optional: true
            - name: AI_MODEL_GPT_5_3_CODEX_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: AI_MODEL_GPT_5_3_CODEX_LABEL
                  optional: true
            - name: AI_MODEL_GPT_5_3_CODEX_SPARK_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: AI_MODEL_GPT_5_3_CODEX_SPARK_LABEL
                  optional: true
            - name: AI_MODEL_GPT_5_2_CODEX_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: AI_MODEL_GPT_5_2_CODEX_LABEL
                  optional: true
            - name: AI_MODEL_GPT_5_1_CODEX_MAX_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: AI_MODEL_GPT_5_1_CODEX_MAX_LABEL
                  optional: true
            - name: AI_MODEL_GPT_5_2_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: AI_MODEL_GPT_5_2_LABEL
                  optional: true
            - name: AI_MODEL_GPT_5_1_CODEX_MINI_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: AI_MODEL_GPT_5_1_CODEX_MINI_LABEL
                  optional: true
            - name: AI_REASONING_LOW_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: AI_REASONING_LOW_LABEL
                  optional: true
            - name: AI_REASONING_MEDIUM_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: AI_REASONING_MEDIUM_LABEL
                  optional: true
            - name: AI_REASONING_HIGH_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: AI_REASONING_HIGH_LABEL
                  optional: true
            - name: AI_REASONING_EXTRA_HIGH_LABEL
              valueFrom:
                configMapKeyRef:
                  name: codex-k8s-label-catalog
                  key: AI_REASONING_EXTRA_HIGH_LABEL
                  optional: true
            - name: CODEXK8S_WORKER_RUN_SERVICE_ACCOUNT
              value: "${CODEXK8S_WORKER_RUN_SERVICE_ACCOUNT}"
            - name: CODEXK8S_WORKER_RUN_ROLE_NAME
              value: "${CODEXK8S_WORKER_RUN_ROLE_NAME}"
            - name: CODEXK8S_WORKER_RUN_ROLE_BINDING_NAME
              value: "${CODEXK8S_WORKER_RUN_ROLE_BINDING_NAME}"
            - name: CODEXK8S_WORKER_RUN_RESOURCE_QUOTA_NAME
              value: "${CODEXK8S_WORKER_RUN_RESOURCE_QUOTA_NAME}"
            - name: CODEXK8S_WORKER_RUN_LIMIT_RANGE_NAME
              value: "${CODEXK8S_WORKER_RUN_LIMIT_RANGE_NAME}"
            - name: CODEXK8S_WORKER_RUN_CREDENTIALS_SECRET_NAME
              value: "${CODEXK8S_WORKER_RUN_CREDENTIALS_SECRET_NAME}"
            - name: CODEXK8S_WORKER_RUN_QUOTA_PODS
              value: "${CODEXK8S_WORKER_RUN_QUOTA_PODS}"
            - name: CODEXK8S_WORKER_RUN_QUOTA_REQUESTS_CPU
              value: "${CODEXK8S_WORKER_RUN_QUOTA_REQUESTS_CPU}"
            - name: CODEXK8S_WORKER_RUN_QUOTA_REQUESTS_MEMORY
              value: "${CODEXK8S_WORKER_RUN_QUOTA_REQUESTS_MEMORY}"
            - name: CODEXK8S_WORKER_RUN_QUOTA_LIMITS_CPU
              value: "${CODEXK8S_WORKER_RUN_QUOTA_LIMITS_CPU}"
            - name: CODEXK8S_WORKER_RUN_QUOTA_LIMITS_MEMORY
              value: "${CODEXK8S_WORKER_RUN_QUOTA_LIMITS_MEMORY}"
            - name: CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_REQUEST_CPU
              value: "${CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_REQUEST_CPU}"
            - name: CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_REQUEST_MEMORY
              value: "${CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_REQUEST_MEMORY}"
            - name: CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_CPU
              value: "${CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_CPU}"
            - name: CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_MEMORY
              value: "${CODEXK8S_WORKER_RUN_LIMIT_DEFAULT_MEMORY}"
            - name: CODEXK8S_AGENT_DEFAULT_MODEL
              value: "${CODEXK8S_AGENT_DEFAULT_MODEL}"
            - name: CODEXK8S_AGENT_DEFAULT_REASONING_EFFORT
              value: "${CODEXK8S_AGENT_DEFAULT_REASONING_EFFORT}"
            - name: CODEXK8S_AGENT_DEFAULT_LOCALE
              value: "${CODEXK8S_AGENT_DEFAULT_LOCALE}"
            - name: CODEXK8S_AGENT_BASE_BRANCH
              value: "${CODEXK8S_AGENT_BASE_BRANCH}"
          resources:
            requests:
              cpu: ${CODEXK8S_WORKER_RESOURCES_REQUEST_CPU}
              memory: ${CODEXK8S_WORKER_RESOURCES_REQUEST_MEMORY}
            limits:
              cpu: ${CODEXK8S_WORKER_RESOURCES_LIMIT_CPU}
              memory: ${CODEXK8S_WORKER_RESOURCES_LIMIT_MEMORY}
