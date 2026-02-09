apiVersion: batch/v1
kind: Job
metadata:
  name: codex-k8s-migrate
  namespace: ${CODEXK8S_STAGING_NAMESPACE}
  labels:
    app.kubernetes.io/name: codex-k8s
    app.kubernetes.io/component: migrate
spec:
  backoffLimit: 0
  template:
    metadata:
      labels:
        app.kubernetes.io/name: codex-k8s
        app.kubernetes.io/component: migrate
    spec:
      restartPolicy: Never
      containers:
        - name: migrate
          image: ${CODEXK8S_IMAGE}
          imagePullPolicy: Always
          env:
            - name: CODEXK8S_POSTGRES_DB
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-postgres
                  key: CODEXK8S_POSTGRES_DB
            - name: CODEXK8S_POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-postgres
                  key: CODEXK8S_POSTGRES_USER
            - name: CODEXK8S_POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: codex-k8s-postgres
                  key: CODEXK8S_POSTGRES_PASSWORD
          command:
            - sh
            - -ec
            - |
              export GOOSE_DRIVER=postgres
              export GOOSE_DBSTRING="postgres://${CODEXK8S_POSTGRES_USER}:${CODEXK8S_POSTGRES_PASSWORD}@postgres:5432/${CODEXK8S_POSTGRES_DB}?sslmode=disable"
              /usr/local/bin/goose -dir /migrations up
          volumeMounts:
            - name: migrations
              mountPath: /migrations
      volumes:
        - name: migrations
          configMap:
            name: codex-k8s-migrations

