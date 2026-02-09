apiVersion: apps/v1
kind: Deployment
metadata:
  name: oauth2-proxy
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: oauth2-proxy
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: oauth2-proxy
  template:
    metadata:
      labels:
        app.kubernetes.io/name: oauth2-proxy
        app.kubernetes.io/component: auth-proxy
    spec:
      containers:
        - name: oauth2-proxy
          image: quay.io/oauth2-proxy/oauth2-proxy:v7.6.0
          imagePullPolicy: IfNotPresent
          args:
            - --provider=github
            - --http-address=0.0.0.0:4180
            - --upstream=http://codex-k8s
            # Ensure upstream sees the original public Host (Ingress host).
            # api-gateway uses it when reverse-proxying the Vite dev server.
            - --pass-host-header=true
            - --redirect-url=https://${CODEXK8S_STAGING_DOMAIN}/oauth2/callback
            - --email-domain=*
            - --cookie-secure=true
            - --cookie-samesite=lax
            # Let GitHub webhooks through (webhook ingress is the primary entrypoint into the platform).
            - --skip-auth-regex=^/api/v1/webhooks/.*
            # Pass identity to upstream via headers (api-gateway will enforce allowlist/RBAC by email).
            - --pass-user-headers=true
            - --set-xauthrequest=true
          env:
            - name: OAUTH2_PROXY_CLIENT_ID
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-oauth2-proxy
                  key: OAUTH2_PROXY_CLIENT_ID
            - name: OAUTH2_PROXY_CLIENT_SECRET
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-oauth2-proxy
                  key: OAUTH2_PROXY_CLIENT_SECRET
            - name: OAUTH2_PROXY_COOKIE_SECRET
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-oauth2-proxy
                  key: OAUTH2_PROXY_COOKIE_SECRET
          ports:
            - name: http
              containerPort: 4180
          readinessProbe:
            httpGet:
              path: /ping
              port: http
            initialDelaySeconds: 2
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /ping
              port: http
            initialDelaySeconds: 10
            periodSeconds: 20
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 500m
              memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: oauth2-proxy
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: oauth2-proxy
spec:
  selector:
    app.kubernetes.io/name: oauth2-proxy
  ports:
    - name: http
      port: 4180
      targetPort: 4180
