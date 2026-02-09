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
      containers:
        - name: codex-k8s
          image: ${CODEXK8S_IMAGE}
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
              name: http
          env:
            - name: CODEXK8S_ENV
              value: ai-staging
            - name: CODEXK8S_HTTP_ADDR
              value: ":8080"
            - name: CODEXK8S_VITE_DEV_UPSTREAM
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_VITE_DEV_UPSTREAM
                  optional: true
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
            - name: CODEXK8S_OPENAI_API_KEY
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_OPENAI_API_KEY
            - name: CODEXK8S_CONTEXT7_API_KEY
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_CONTEXT7_API_KEY
                  optional: true
            - name: CODEXK8S_APP_SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_APP_SECRET_KEY
            - name: CODEXK8S_TOKEN_ENCRYPTION_KEY
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_TOKEN_ENCRYPTION_KEY
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
              cpu: 100m
              memory: 256Mi
            limits:
              cpu: 1000m
              memory: 1Gi
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
kind: Role
metadata:
  name: codex-k8s-worker
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: worker
rules:
  - apiGroups: ["batch"]
    resources: ["jobs"]
    verbs: ["get", "list", "watch", "create", "delete", "patch", "update"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: codex-k8s-worker
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: worker
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: codex-k8s-worker
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
          image: ${CODEXK8S_IMAGE}
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
          resources:
            requests:
              cpu: 100m
              memory: 256Mi
            limits:
              cpu: 1000m
              memory: 1Gi
