apiVersion: v1
kind: Service
metadata:
  name: codex-k8s-web-console
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
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
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s-web-console
spec:
  replicas: ${CODEXK8S_PLATFORM_DEPLOYMENT_REPLICAS}
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
          image: ${CODEXK8S_WEB_CONSOLE_IMAGE}
          imagePullPolicy: Always
          ports:
            - containerPort: 5173
              name: http
          env:
            # Vite blocks unknown hosts by default; this keeps staging usable behind Ingress.
            - name: VITE_ALLOWED_HOSTS
              value: "${CODEXK8S_STAGING_DOMAIN}"
            # HMR in ai-staging runs behind public Ingress (HTTPS) and must not try to connect to localhost:5173.
            - name: VITE_HMR_HOST
              value: "${CODEXK8S_STAGING_DOMAIN}"
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
