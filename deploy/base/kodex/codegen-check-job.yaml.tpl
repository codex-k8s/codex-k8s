apiVersion: batch/v1
kind: Job
metadata:
  name: {{ envOr "KODEX_CODEGEN_CHECK_JOB_NAME" "kodex-codegen-check" }}
  namespace: {{ envOr "KODEX_PRODUCTION_NAMESPACE" "" }}
  labels:
    app.kubernetes.io/name: kodex
    app.kubernetes.io/component: codegen-check
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: {{ envOr "KODEX_CODEGEN_CHECK_TTL_SECONDS" "3600" }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: kodex
        app.kubernetes.io/component: codegen-check
    spec:
      restartPolicy: Never
      containers:
        - name: codegen-check
          image: {{ envOr "KODEX_CODEGEN_CHECK_IMAGE" "golang:1.24-bookworm" }}
          imagePullPolicy: IfNotPresent
          env:
            - name: GIT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: kodex-git-token
                  key: token
            - name: KODEX_GITHUB_REPO
              value: '{{ envOr "KODEX_GITHUB_REPO" "" }}'
            - name: KODEX_BUILD_REF
              value: '{{ envOr "KODEX_BUILD_REF" "main" }}'
            - name: KODEX_NODE_VERSION
              value: '{{ envOr "KODEX_CODEGEN_CHECK_NODE_VERSION" "22.14.0" }}'
          command:
            - bash
            - -ec
            - |
              set -euo pipefail
              apt-get update
              apt-get install -y --no-install-recommends ca-certificates curl git make xz-utils

              curl -fsSL "https://nodejs.org/dist/v${KODEX_NODE_VERSION}/node-v${KODEX_NODE_VERSION}-linux-x64.tar.xz" -o /tmp/node.tar.xz
              tar -xJf /tmp/node.tar.xz -C /tmp
              cp -R "/tmp/node-v${KODEX_NODE_VERSION}-linux-x64/." /usr/local/

              git clone "https://x-access-token:${GIT_TOKEN}@github.com/${KODEX_GITHUB_REPO}.git" /workspace
              cd /workspace
              checkout_ref="$KODEX_BUILD_REF"
              if git rev-parse --verify -q "origin/$KODEX_BUILD_REF^{commit}" >/dev/null 2>&1; then
                checkout_ref="origin/$KODEX_BUILD_REF"
              fi
              git checkout --detach "$checkout_ref"

              npm --prefix services/staff/web-console ci
              make gen-openapi

              git diff --exit-code -- \
                services/external/api-gateway/internal/transport/http/generated/openapi.gen.go \
                services/external/telegram-interaction-adapter/internal/transport/http/generated/openapi.gen.go \
                services/staff/web-console/src/shared/api/generated
