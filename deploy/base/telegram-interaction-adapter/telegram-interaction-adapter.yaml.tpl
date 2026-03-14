apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: codex-k8s-telegram-interaction-adapter-state
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: telegram-interaction-adapter
spec:
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      storage: {{ envOr "CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_STATE_STORAGE_SIZE" "1Gi" }}
---
apiVersion: v1
kind: Service
metadata:
  name: codex-k8s-telegram-interaction-adapter
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: telegram-interaction-adapter
spec:
  selector:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: telegram-interaction-adapter
  ports:
    - name: http
      port: 8080
      targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: codex-k8s-telegram-interaction-adapter
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: telegram-interaction-adapter
spec:
  replicas: {{ envOr "CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_REPLICAS" "1" }}
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app.kubernetes.io/name: codex-k8s
      app.kubernetes.io/component: telegram-interaction-adapter
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s
        app.kubernetes.io/component: telegram-interaction-adapter
    spec:
      containers:
        - name: telegram-interaction-adapter
          image: {{ envOr "CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_IMAGE" "" }}
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
              name: http
          env:
            - name: CODEXK8S_HTTP_ADDR
              value: ":8080"
            - name: CODEXK8S_PUBLIC_BASE_URL
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_PUBLIC_BASE_URL
            - name: CODEXK8S_TELEGRAM_BOT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_TELEGRAM_BOT_TOKEN
                  optional: true
            - name: CODEXK8S_TELEGRAM_CHAT_ID
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_TELEGRAM_CHAT_ID
                  optional: true
            - name: CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_RECIPIENT_BINDINGS_JSON
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_RECIPIENT_BINDINGS_JSON
                  optional: true
            - name: CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN
            - name: CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_WEBHOOK_SECRET
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-runtime
                  key: CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_WEBHOOK_SECRET
            - name: CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_STATE_PATH
              value: "/var/lib/codex-k8s-telegram-interaction-adapter/state.json"
            - name: CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_HTTP_TIMEOUT
              value: '{{ envOr "CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_HTTP_TIMEOUT" "10s" }}'
            - name: CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_CALLBACK_HTTP_TIMEOUT
              value: '{{ envOr "CODEXK8S_TELEGRAM_INTERACTION_ADAPTER_CALLBACK_HTTP_TIMEOUT" "10s" }}'
          volumeMounts:
            - name: state
              mountPath: /var/lib/codex-k8s-telegram-interaction-adapter
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
      volumes:
        - name: state
          persistentVolumeClaim:
            claimName: codex-k8s-telegram-interaction-adapter-state
