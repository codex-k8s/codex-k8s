apiVersion: v1
kind: Service
metadata:
  name: ${CODEXK8S_INTERNAL_REGISTRY_SERVICE}
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s-registry
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: codex-k8s-registry
  ports:
    - name: registry
      port: ${CODEXK8S_INTERNAL_REGISTRY_PORT}
      targetPort: registry
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: ${CODEXK8S_INTERNAL_REGISTRY_SERVICE}
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s-registry
spec:
  serviceName: ${CODEXK8S_INTERNAL_REGISTRY_SERVICE}
  replicas: 1
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
          image: registry:2
          imagePullPolicy: IfNotPresent
          ports:
            - name: registry
              containerPort: ${CODEXK8S_INTERNAL_REGISTRY_PORT}
          env:
            # No auth by design for MVP: registry is exposed only as in-cluster ClusterIP.
            - name: REGISTRY_HTTP_ADDR
              value: "0.0.0.0:${CODEXK8S_INTERNAL_REGISTRY_PORT}"
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
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: ${CODEXK8S_INTERNAL_REGISTRY_STORAGE_SIZE}
