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
  replicas: 1
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
            # Disable HMR in-cluster to avoid websocket/proxy issues that can cause periodic full-page reloads.
            - name: VITE_DISABLE_HMR
              value: "true"
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
          resources:
            requests:
              cpu: 50m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
