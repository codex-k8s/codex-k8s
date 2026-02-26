{{- $host := envOr "CODEXK8S_PUBLIC_DOMAIN" (envOr "CODEXK8S_PRODUCTION_DOMAIN" "") -}}
apiVersion: v1
kind: Service
metadata:
  name: codex-k8s-web-console
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-web-console
spec:
  selector:
    app.kubernetes.io/name: codex-k8s-web-console
  ports:
    - name: http
      port: 5173
      targetPort: 5173
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: codex-k8s-web-console
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-web-console
spec:
  replicas: {{ envOr "CODEXK8S_PLATFORM_DEPLOYMENT_REPLICAS" "" }}
  selector:
    matchLabels:
      app.kubernetes.io/name: codex-k8s-web-console
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s-web-console
        app.kubernetes.io/component: staff-ui
    spec:
      containers:
        - name: web-console
          image: {{ envOr "CODEXK8S_WEB_CONSOLE_IMAGE" "" }}
          imagePullPolicy: Always
{{ if eq (envOr "CODEXK8S_HOT_RELOAD" "") "true" }}
          command:
            - sh
            - -ec
            - |
              if [ -d /workspace/services/staff/web-console ]; then
                cd /workspace/services/staff/web-console
                if [ ! -d node_modules ]; then
                  npm ci
                fi
                exec npm run dev -- --host 0.0.0.0 --port 5173
              fi
              cd /app
              exec npm run dev -- --host 0.0.0.0 --port 5173
{{ end }}
          ports:
            - containerPort: 5173
              name: http
          env:
            # Vite blocks unknown hosts by default; this keeps production usable behind Ingress.
            - name: VITE_ALLOWED_HOSTS
              value: '{{ $host }}'
            # HMR in production runs behind public Ingress (HTTPS) and must not try to connect to localhost:5173.
            - name: VITE_HMR_HOST
              value: '{{ $host }}'
            - name: VITE_HMR_PROTOCOL
              value: "wss"
            - name: VITE_HMR_CLIENT_PORT
              value: "443"
            # Dedicated path so we can route websocket directly to Vite service (bypassing auth proxy).
            - name: VITE_HMR_PATH
              value: "/__vite_ws"
          readinessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 2
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 10
            periodSeconds: 20
{{ if eq (envOr "CODEXK8S_HOT_RELOAD" "") "true" }}
          volumeMounts:
            - name: repo-cache
              mountPath: /workspace
{{ end }}
{{ if eq (envOr "CODEXK8S_HOT_RELOAD" "") "true" }}
      volumes:
        - name: repo-cache
          persistentVolumeClaim:
            claimName: codex-k8s-repo-cache
{{ end }}
