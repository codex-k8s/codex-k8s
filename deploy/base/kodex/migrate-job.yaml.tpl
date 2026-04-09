apiVersion: batch/v1
kind: Job
metadata:
  name: kodex-migrate
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: migrate
spec:
  backoffLimit: 0
  template:
    metadata:
      labels:
        app.kubernetes.io/name: kodex
        app.kubernetes.io/component: migrate
    spec:
      restartPolicy: Never
      containers:
        - name: migrate
          image: {{ envOr "KODEX_CONTROL_PLANE_IMAGE" "" }}
          imagePullPolicy: Always
          env:
            - name: KODEX_POSTGRES_DB
              valueFrom:
                secretKeyRef:
                  name: kodex-postgres
                  key: KODEX_POSTGRES_DB
            - name: KODEX_POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: kodex-postgres
                  key: KODEX_POSTGRES_USER
            - name: KODEX_POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: kodex-postgres
                  key: KODEX_POSTGRES_PASSWORD
          command:
            - sh
            - -ec
            - |
              export GOOSE_DRIVER=postgres
              # Use shell runtime variables from container env (set above via secret refs).
              export GOOSE_DBSTRING="postgres://$KODEX_POSTGRES_USER:$KODEX_POSTGRES_PASSWORD@postgres:5432/$KODEX_POSTGRES_DB?sslmode=disable"
              # Postgres Service can be routable slightly before the actual server
              # accepts connections. Keep this step resilient to short transient failures,
              # but fail fast after 60s total timeout.
              timeout_seconds=60
              started_at="$(date +%s)"
              i=1
              while true; do
                if /usr/local/bin/goose -dir /migrations up; then
                  exit 0
                fi
                now="$(date +%s)"
                elapsed="$((now - started_at))"
                if [ "$elapsed" -ge "$timeout_seconds" ]; then
                  echo "goose up failed after $elapsed s (timeout $timeout_seconds s)" >&2
                  exit 1
                fi
                echo "goose up failed (attempt $i, elapsed $elapsed s); retry in 2s..." >&2
                i=$((i + 1))
                sleep 2
              done
          volumeMounts:
            - name: migrations
              mountPath: /migrations
      volumes:
        - name: migrations
          configMap:
            name: kodex-migrations
