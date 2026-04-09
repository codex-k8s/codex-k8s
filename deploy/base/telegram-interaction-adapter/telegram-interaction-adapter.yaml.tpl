apiVersion: v1
kind: Service
metadata:
  name: kodex-telegram-interaction-adapter
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: telegram-interaction-adapter
spec:
  selector:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: telegram-interaction-adapter
  ports:
    - name: http
      port: 8080
      targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kodex-telegram-interaction-adapter
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: telegram-interaction-adapter
spec:
  replicas: {{ envOr "KODEX_TELEGRAM_INTERACTION_ADAPTER_REPLICAS" "2" }}
  minReadySeconds: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
      maxSurge: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: kodex
      app.kubernetes.io/component: telegram-interaction-adapter
  template:
    metadata:
      labels:
        app.kubernetes.io/name: kodex
        app.kubernetes.io/component: telegram-interaction-adapter
    spec:
      containers:
        - name: telegram-interaction-adapter
          image: {{ envOr "KODEX_TELEGRAM_INTERACTION_ADAPTER_IMAGE" "" }}
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
              name: http
          env:
            - name: KODEX_HTTP_ADDR
              value: ":8080"
            - name: KODEX_ENV
              value: '{{ envOr "KODEX_ENV" "production" }}'
            - name: KODEX_PUBLIC_BASE_URL
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_PUBLIC_BASE_URL
            - name: KODEX_CONTROL_PLANE_GRPC_TARGET
              value: '{{ envOr "KODEX_CONTROL_PLANE_GRPC_TARGET" "kodex-control-plane:9090" }}'
            - name: KODEX_OPENAI_API_KEY
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_OPENAI_API_KEY
                  optional: true
            - name: KODEX_TELEGRAM_BOT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_TELEGRAM_BOT_TOKEN
                  optional: true
            - name: KODEX_TELEGRAM_CHAT_ID
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_TELEGRAM_CHAT_ID
                  optional: true
            - name: KODEX_TELEGRAM_INTERACTION_ADAPTER_RECIPIENT_BINDINGS_JSON
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_TELEGRAM_INTERACTION_ADAPTER_RECIPIENT_BINDINGS_JSON
                  optional: true
            - name: KODEX_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_TELEGRAM_INTERACTION_ADAPTER_BEARER_TOKEN
            - name: KODEX_TELEGRAM_INTERACTION_ADAPTER_WEBHOOK_SECRET
              valueFrom:
                secretKeyRef:
                  name: kodex-runtime
                  key: KODEX_TELEGRAM_INTERACTION_ADAPTER_WEBHOOK_SECRET
            - name: KODEX_TELEGRAM_INTERACTION_ADAPTER_HTTP_TIMEOUT
              value: '{{ envOr "KODEX_TELEGRAM_INTERACTION_ADAPTER_HTTP_TIMEOUT" "10s" }}'
            - name: KODEX_TELEGRAM_INTERACTION_ADAPTER_STT_MODEL
              value: '{{ envOr "KODEX_TELEGRAM_INTERACTION_ADAPTER_STT_MODEL" "gpt-4o-mini-transcribe" }}'
            - name: KODEX_TELEGRAM_INTERACTION_ADAPTER_STT_TIMEOUT
              value: '{{ envOr "KODEX_TELEGRAM_INTERACTION_ADAPTER_STT_TIMEOUT" "30s" }}'
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
