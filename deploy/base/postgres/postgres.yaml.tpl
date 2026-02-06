apiVersion: v1
kind: ConfigMap
metadata:
  name: codex-k8s-postgres-init
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
data:
  01-pgvector.sql: |
    CREATE EXTENSION IF NOT EXISTS vector;
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: postgres
spec:
  selector:
    app.kubernetes.io/name: postgres
  ports:
    - name: pg
      port: 5432
      targetPort: 5432
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: postgres
  template:
    metadata:
      labels:
        app.kubernetes.io/name: postgres
    spec:
      containers:
        - name: postgres
          image: pgvector/pgvector:pg16
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 5432
              name: pg
          env:
            - name: POSTGRES_DB
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-postgres
                  key: CODEXK8S_POSTGRES_DB
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-postgres
                  key: CODEXK8S_POSTGRES_USER
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-postgres
                  key: CODEXK8S_POSTGRES_PASSWORD
          volumeMounts:
            - name: data
              mountPath: /var/lib/postgresql/data
            - name: init-sql
              mountPath: /docker-entrypoint-initdb.d/01-pgvector.sql
              subPath: 01-pgvector.sql
          readinessProbe:
            exec:
              command: ["sh", "-c", "pg_isready -U $POSTGRES_USER -d $POSTGRES_DB"]
            initialDelaySeconds: 10
            periodSeconds: 10
      volumes:
        - name: init-sql
          configMap:
            name: codex-k8s-postgres-init
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 20Gi
