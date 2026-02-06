apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ${CODEXK8S_INTERNAL_REGISTRY_SERVICE}-data
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s-registry
spec:
  accessModes: ["ReadWriteOnce"]
  resources:
    requests:
      storage: ${CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${CODEXK8S_INTERNAL_REGISTRY_SERVICE}
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
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
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      containers:
        - name: registry
          image: registry:2
          imagePullPolicy: IfNotPresent
          ports:
            - name: registry
              containerPort: ${CODEXK8S_INTERNAL_REGISTRY_PORT}
          env:
            # No auth by design for MVP: staging registry is bound to node loopback only.
            - name: REGISTRY_HTTP_ADDR
              value: "127.0.0.1:${CODEXK8S_INTERNAL_REGISTRY_PORT}"
            - name: REGISTRY_STORAGE_DELETE_ENABLED
              value: "true"
          readinessProbe:
            httpGet:
              host: 127.0.0.1
              path: /v2/
              port: registry
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              host: 127.0.0.1
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
            claimName: ${CODEXK8S_INTERNAL_REGISTRY_SERVICE}-data
