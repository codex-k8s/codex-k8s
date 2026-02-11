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
          image: ${CODEXK8S_CONTROL_PLANE_IMAGE}
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
              # Postgres Service can be routable slightly before the actual server
              # accepts connections. Keep this step resilient to short transient failures.
              retries=60
              i=1
              while [ "$i" -le "$retries" ]; do
                if /usr/local/bin/goose -dir /migrations up; then
                  exit 0
                fi
                echo "goose up failed (attempt ${i}/${retries}); retry in 2s..." >&2
                i=$((i + 1))
                sleep 2
              done
              echo "goose up failed after ${retries} attempts" >&2
              exit 1
          volumeMounts:
            - name: migrations
              mountPath: /migrations
      volumes:
        - name: migrations
          configMap:
            name: codex-k8s-migrations
