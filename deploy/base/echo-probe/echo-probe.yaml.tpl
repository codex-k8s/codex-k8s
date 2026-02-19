apiVersion: v1
kind: Service
metadata:
  name: codex-k8s-echo-probe
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-echo-probe
spec:
  selector:
    app.kubernetes.io/name: codex-k8s-echo-probe
  ports:
    - name: http
      port: 8080
      targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: codex-k8s-echo-probe
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-echo-probe
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: codex-k8s-echo-probe
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s-echo-probe
    spec:
      containers:
        - name: echo-probe
          image: {{ envOr "CODEXK8S_BUSYBOX_IMAGE" "busybox:1.36" }}
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
              name: http
          env:
            - name: CODEXK8S_ECHO_PROBE_TOKEN
              value: '{{ envOr "CODEXK8S_ECHO_PROBE_TOKEN" "" }}'
          command:
            - sh
            - -ec
            - |
              token="${CODEXK8S_ECHO_PROBE_TOKEN:-}"
              if [ -z "$token" ]; then
                echo "CODEXK8S_ECHO_PROBE_TOKEN is empty"
                exit 1
              fi
              mkdir -p /www/.well-known/codex-k8s-probe
              echo -n "$token" > "/www/.well-known/codex-k8s-probe/$token"
              echo -n "ok" > /www/index.html
              exec httpd -f -p 8080 -h /www
          readinessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 2
            periodSeconds: 5
          livenessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 10
            periodSeconds: 20
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: codex-k8s-echo-probe
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-echo-probe
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "false"
spec:
  ingressClassName: nginx
  rules:
    - host: {{ envOr "CODEXK8S_ECHO_PROBE_HOST" "" }}
      http:
        paths:
          - path: /.well-known/codex-k8s-probe/
            pathType: Prefix
            backend:
              service:
                name: codex-k8s-echo-probe
                port:
                  number: 8080
