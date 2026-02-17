apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}-data
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-registry
spec:
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      storage: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE" "" }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-registry
spec:
  selector:
    app.kubernetes.io/name: codex-k8s-registry
  ports:
    - name: registry
      port: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_PORT" "" }}
      targetPort: registry
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}
  namespace: {{ envOr "CODEXK8S_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: codex-k8s-registry
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app.kubernetes.io/name: codex-k8s-registry
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s-registry
    spec:
      containers:
        - name: registry
          image: {{ envOr "CODEXK8S_REGISTRY_IMAGE" "registry:2" }}
          imagePullPolicy: IfNotPresent
          ports:
            - name: registry
              containerPort: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_PORT" "" }}
              hostPort: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_PORT" "" }}
              hostIP: 127.0.0.1
          env:
            - name: REGISTRY_HTTP_ADDR
              value: ':{{ envOr "CODEXK8S_INTERNAL_REGISTRY_PORT" "" }}'
            - name: REGISTRY_STORAGE_DELETE_ENABLED
              value: "true"
          readinessProbe:
            httpGet:
              path: /v2/
              port: registry
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /v2/
              port: registry
            initialDelaySeconds: 15
            periodSeconds: 20
          volumeMounts:
            - name: data
              mountPath: /var/lib/registry
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: {{ envOr "CODEXK8S_INTERNAL_REGISTRY_SERVICE" "" }}-data
