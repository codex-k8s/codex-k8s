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
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s
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
            - name: CODEXK8S_GITHUB_WEBHOOK_SECRET
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_GITHUB_WEBHOOK_SECRET
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
